package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/PubMatic-OpenWrap/openrtb"
	accountService "github.com/PubMatic-OpenWrap/prebid-server/account"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/analytics"
	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/errortypes"
	"github.com/PubMatic-OpenWrap/prebid-server/prebid_cache_client"
	"github.com/PubMatic-OpenWrap/prebid-server/stored_requests"

	// "github.com/beevik/etree"
	"github.com/PubMatic-OpenWrap/etree"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
)

const (
	AccountParameter   = "a"
	ImpressionCloseTag = "</Impression>"
	ImpressionOpenTag  = "<Impression>"
)

// standard VAST macros
// https://interactiveadvertisingbureau.github.io/vast/vast4macros/vast4-macros-latest.html#macro-spec-adcount
const (
	VASTAdTypeMacro    = "[ADTYPE]"
	VASTAppBundleMacro = "[APPBUNDLE]"
	VASTDomainMacro    = "[DOMAIN]"
	VASTPageURLMacro   = "[PAGEURL]"
	PBSEventIDMacro    = "[EVENT_ID]" // macro for injecting PBS defined  video event tracker id
)

type vtrackEndpoint struct {
	Cfg         *config.Configuration
	Accounts    stored_requests.AccountFetcher
	BidderInfos adapters.BidderInfos
	Cache       prebid_cache_client.Client
}

type BidCacheRequest struct {
	Puts []prebid_cache_client.Cacheable `json:"puts"`
}

type BidCacheResponse struct {
	Responses []CacheObject `json:"responses"`
}

type CacheObject struct {
	UUID string `json:"uuid"`
}

func NewVTrackEndpoint(cfg *config.Configuration, accounts stored_requests.AccountFetcher, cache prebid_cache_client.Client, bidderInfos adapters.BidderInfos) httprouter.Handle {
	vte := &vtrackEndpoint{
		Cfg:         cfg,
		Accounts:    accounts,
		BidderInfos: bidderInfos,
		Cache:       cache,
	}

	return vte.Handle
}

// /vtrack Handler
func (v *vtrackEndpoint) Handle(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// get account id from request parameter
	accountId := getAccountId(r)

	// account id is required
	if accountId == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Account '%s' is required query parameter and can't be empty", AccountParameter)))
		return
	}

	// parse puts request from request body
	req, err := ParseVTrackRequest(r, v.Cfg.MaxRequestSize+1)

	// check if there was any error while parsing puts request
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Invalid request: %s\n", err.Error())))
		return
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(v.Cfg.VTrack.TimeoutMS)*time.Millisecond))
	defer cancel()

	// get account details
	account, errs := accountService.GetAccount(ctx, v.Cfg, v.Accounts, accountId)
	if len(errs) > 0 {
		status, messages := HandleAccountServiceErrors(errs)
		w.WriteHeader(status)

		for _, message := range messages {
			w.Write([]byte(fmt.Sprintf("Invalid request: %s\n", message)))
		}
		return
	}

	// insert impression tracking if account allows events and bidder allows VAST modification
	if v.Cache != nil {
		cachingResponse, errs := v.handleVTrackRequest(ctx, req, account)

		if len(errs) > 0 {
			w.WriteHeader(http.StatusInternalServerError)
			for _, err := range errs {
				w.Write([]byte(fmt.Sprintf("Error(s) updating vast: %s\n", err.Error())))

				return
			}
		}

		d, err := json.Marshal(*cachingResponse)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error serializing pbs cache response: %s\n", err.Error())))

			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(d)

		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("PBS Cache client is not configured"))
}

// GetVastUrlTracking creates a vast url tracking
func GetVastUrlTracking(externalUrl string, bidid string, bidder string, accountId string, timestamp int64) string {

	eventReq := &analytics.EventRequest{
		Type:      analytics.Imp,
		BidID:     bidid,
		AccountID: accountId,
		Bidder:    bidder,
		Timestamp: timestamp,
		Format:    analytics.Blank,
	}

	return EventRequestToUrl(externalUrl, eventReq)
}

// ParseVTrackRequest parses a BidCacheRequest from an HTTP Request
func ParseVTrackRequest(httpRequest *http.Request, maxRequestSize int64) (req *BidCacheRequest, err error) {
	req = &BidCacheRequest{}
	err = nil

	// Pull the request body into a buffer, so we have it for later usage.
	lr := &io.LimitedReader{
		R: httpRequest.Body,
		N: maxRequestSize,
	}

	defer httpRequest.Body.Close()
	requestJson, err := ioutil.ReadAll(lr)
	if err != nil {
		return req, err
	}

	// Check if the request size was too large
	if lr.N <= 0 {
		err = &errortypes.BadInput{Message: fmt.Sprintf("request size exceeded max size of %d bytes", maxRequestSize-1)}
		return req, err
	}

	if len(requestJson) == 0 {
		err = &errortypes.BadInput{Message: "request body is empty"}
		return req, err
	}

	if err := json.Unmarshal(requestJson, req); err != nil {
		return req, err
	}

	for _, bcr := range req.Puts {
		if bcr.BidID == "" {
			err = error(&errortypes.BadInput{Message: fmt.Sprint("'bidid' is required field and can't be empty")})
			return req, err
		}

		if bcr.Bidder == "" {
			err = error(&errortypes.BadInput{Message: fmt.Sprint("'bidder' is required field and can't be empty")})
			return req, err
		}
	}

	return req, nil
}

// handleVTrackRequest handles a VTrack request
func (v *vtrackEndpoint) handleVTrackRequest(ctx context.Context, req *BidCacheRequest, account *config.Account) (*BidCacheResponse, []error) {
	biddersAllowingVastUpdate := getBiddersAllowingVastUpdate(req, &v.BidderInfos, v.Cfg.VTrack.AllowUnknownBidder)
	// cache data
	r, errs := v.cachePutObjects(ctx, req, biddersAllowingVastUpdate, account.ID)

	// handle pbs caching errors
	if len(errs) != 0 {
		glog.Errorf("Error(s) updating vast: %v", errs)
		return nil, errs
	}

	// build response
	response := &BidCacheResponse{
		Responses: []CacheObject{},
	}

	for _, uuid := range r {
		response.Responses = append(response.Responses, CacheObject{
			UUID: uuid,
		})
	}

	return response, nil
}

// cachePutObjects caches BidCacheRequest data
func (v *vtrackEndpoint) cachePutObjects(ctx context.Context, req *BidCacheRequest, biddersAllowingVastUpdate map[string]struct{}, accountId string) ([]string, []error) {
	var cacheables []prebid_cache_client.Cacheable

	for _, c := range req.Puts {

		nc := &prebid_cache_client.Cacheable{
			Type:       c.Type,
			Data:       c.Data,
			TTLSeconds: c.TTLSeconds,
			Key:        c.Key,
		}

		if _, ok := biddersAllowingVastUpdate[c.Bidder]; ok && nc.Data != nil {
			nc.Data = ModifyVastXmlJSON(v.Cfg.ExternalURL, nc.Data, c.BidID, c.Bidder, accountId, c.Timestamp)
		}

		cacheables = append(cacheables, *nc)
	}

	return v.Cache.PutJson(ctx, cacheables)
}

// getBiddersAllowingVastUpdate returns a list of bidders that allow VAST XML modification
func getBiddersAllowingVastUpdate(req *BidCacheRequest, bidderInfos *adapters.BidderInfos, allowUnknownBidder bool) map[string]struct{} {
	bl := map[string]struct{}{}

	for _, bcr := range req.Puts {
		if _, ok := bl[bcr.Bidder]; isAllowVastForBidder(bcr.Bidder, bidderInfos, allowUnknownBidder) && !ok {
			bl[bcr.Bidder] = struct{}{}
		}
	}

	return bl
}

// isAllowVastForBidder checks if a bidder is active and allowed to modify vast xml data
func isAllowVastForBidder(bidder string, bidderInfos *adapters.BidderInfos, allowUnknownBidder bool) bool {
	//if bidder is active and isModifyingVastXmlAllowed is true
	// check if bidder is configured
	if b, ok := (*bidderInfos)[bidder]; bidderInfos != nil && ok {
		// check if bidder is enabled
		return b.Status == adapters.StatusActive && b.ModifyingVastXmlAllowed
	}

	return allowUnknownBidder
}

// getAccountId extracts an account id from an HTTP Request
func getAccountId(httpRequest *http.Request) string {
	return httpRequest.URL.Query().Get(AccountParameter)
}

// ModifyVastXmlString rewrites and returns the string vastXML and a flag indicating if it was modified
func ModifyVastXmlString(externalUrl, vast, bidid, bidder, accountID string, timestamp int64) (string, bool) {
	ci := strings.Index(vast, ImpressionCloseTag)

	// no impression tag - pass it as it is
	if ci == -1 {
		return vast, false
	}

	vastUrlTracking := GetVastUrlTracking(externalUrl, bidid, bidder, accountID, timestamp)
	impressionUrl := "<![CDATA[" + vastUrlTracking + "]]>"
	oi := strings.Index(vast, ImpressionOpenTag)

	if ci-oi == len(ImpressionOpenTag) {
		return strings.Replace(vast, ImpressionOpenTag, ImpressionOpenTag+impressionUrl, 1), true
	}

	return strings.Replace(vast, ImpressionCloseTag, ImpressionCloseTag+ImpressionOpenTag+impressionUrl+ImpressionCloseTag, 1), true
}

// ModifyVastXmlJSON modifies BidCacheRequest element Vast XML data
func ModifyVastXmlJSON(externalUrl string, data json.RawMessage, bidid, bidder, accountId string, timestamp int64) json.RawMessage {
	var vast string
	if err := json.Unmarshal(data, &vast); err != nil {
		// failed to decode json, fall back to string
		vast = string(data)
	}
	vast, ok := ModifyVastXmlString(externalUrl, vast, bidid, bidder, accountId, timestamp)
	if !ok {
		return data
	}
	return json.RawMessage(vast)
}

func InjectVideoEventTrackers(trackerURL, vast string, bid *openrtb.Bid, bidder, accountID string, timestamp int64, req *openrtb.BidRequest) ([]byte, bool) {
	return injectVideoEventTrackers0(trackerURL, vast, bid, bidder, accountID, timestamp, req)
}

func injectVideoEventTrackers1(trackerURL, vast string, bid *openrtb.Bid, bidder, accountID string, timestamp int64, req *openrtb.BidRequest) ([]byte, bool) {
	// collect all creatives from vast
	return []byte(""), false
}

//InjectVideoEventTrackers injects the video tracking events
func injectVideoEventTrackers0(trackerURL, adm string, bid *openrtb.Bid, bidder, accountID string, timestamp int64, req *openrtb.BidRequest) ([]byte, bool) {
	// parse VAST
	doc := etree.NewDocument()
	err := doc.ReadFromString(adm)
	if nil != err {
		glog.Errorf("Error parsing VAST XML. '%v'", err.Error())
		return []byte(adm), false // false indicates events trackers are not injected
	}
	//TODO return if error
	eventURLMap := GetVideoEventTracking(trackerURL, bid, bidder, accountID, timestamp, req, doc)
	trackersInjected := false
	// return if if no tracking URL
	if len(eventURLMap) == 0 {
		return []byte(adm), false
	}

	// Find Creatives of Linear and NonLinear Type
	// Injecting Tracking Events for Companion is not supported here
	creatives := doc.FindElements("VAST/Ad/InLine/Creatives/Creative/Linear")
	creatives = append(creatives, doc.FindElements("VAST/Ad/InLine/Creatives/Creative/NonLinear")...)
	creatives = append(creatives, doc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/Linear")...)
	creatives = append(creatives, doc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/NonLinear")...)
	for _, creative := range creatives {
		trackingEvents := creative.SelectElement("TrackingEvents")
		if nil == trackingEvents {
			trackingEvents = creative.CreateElement("TrackingEvents")
			creative.AddChild(trackingEvents)
		}
		// Inject
		for event, url := range eventURLMap {
			trackingEle := trackingEvents.CreateElement("Tracking")
			trackingEle.CreateAttr("event", event)
			// trackingEle.SetText(fmt.Sprintf("<![CDATA[%s]]>", url))
			//trackingEle.SetText(fmt.Sprintf("%s", url))
			// trackingEle.SetCData(url)
			trackingEle.SetText(fmt.Sprintf("%s", url))
			trackersInjected = true
		}
	}

	out := []byte(adm)
	if trackersInjected {
		out, err = doc.WriteToBytes()
		trackersInjected = trackersInjected && nil == err
		if nil != err {
			glog.Errorf("%v", err.Error())

		}
	}
	return out, trackersInjected
}

// GetVideoEventTracking returns map containing key as event name value as associaed video event tracking URL
// By default PBS will expect [EVENT_ID] macro in trackerURL to inject event information
// [EVENT_ID] will be injected with one of the following values
//    firstQuartile, midpoint, thirdQuartile, complete
// If your company can not use [EVENT_ID] and has its own macro. provide config.TrackerMacros implementation
// and ensure that your macro is part of trackerURL configuration
func GetVideoEventTracking(trackerURL string, bid *openrtb.Bid, bidder string, accountId string, timestamp int64, req *openrtb.BidRequest, doc *etree.Document) map[string]string {
	eventURLMap := make(map[string]string)
	if "" == strings.Trim(trackerURL, " ") {
		return eventURLMap
	}
	for _, event := range []string{"firstQuartile", "midpoint", "thirdQuartile", "complete"} {
		eventURL := trackerURL
		macros := make(map[string]string)
		if nil != config.TrackerMacros {
			macros = config.TrackerMacros(event, req, bidder, bid)
		}
		if nil != macros && len(macros) > 0 {
			//trackerURL = "https://aktrack.pubmatic.com/track?operId=8&p=[PUB_ID]&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=[PARTNER_NAME]&crId=[CREATIVE_IDENTIFIER]&ier=[ERRORCODE]&e=3&adv=[ADERVERTISER_NAME]&orig=[PUB_DOMAIN]&dvc.plt=[PLATFORM]&af=[AD_FORMAT]&iid=[WRAPPER_IMPRESSION_ID]"
			for macro, v := range macros {
				eventURL = replaceMacro(eventURL, macro, v)
			}
		}
		// replace standard macros
		eventURL = replaceMacro(eventURL, VASTAdTypeMacro, "")
		if nil != req && nil != req.App {
			eventURL = replaceMacro(eventURL, VASTAppBundleMacro, req.App.Bundle)
		}
		if nil != req && nil != req.Site {
			eventURL = replaceMacro(eventURL, VASTDomainMacro, req.Site.Domain)
			eventURL = replaceMacro(eventURL, VASTPageURLMacro, req.Site.Page)
		}

		// replace [EVENT_ID] macro with PBS defined event ID
		eventURL = replaceMacro(eventURL, PBSEventIDMacro, event)
		eventURLMap[event] = eventURL
	}
	return eventURLMap
}

func replaceMacro(trackerURL, macro, value string) string {
	if strings.HasPrefix(macro, "[") && strings.HasSuffix(macro, "]") && len(strings.Trim(value, " ")) > 0 {
		// value = url.QueryEscape(value)
		trackerURL = strings.ReplaceAll(trackerURL, macro, value)
	} else {
		glog.Warningf("Invalid macro '%v'. Either empty or missing prefix '[' or suffix ']", macro)
	}
	return trackerURL
}
