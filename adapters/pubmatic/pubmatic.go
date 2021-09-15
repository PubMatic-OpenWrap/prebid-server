package pubmatic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/pbs"
	"golang.org/x/net/context/ctxhttp"
)

const (
	MAX_IMPRESSIONS_PUBMATIC = 30
	PUBMATIC                 = "[PUBMATIC]"
	buyId                    = "buyid"
	buyIdTargetingKey        = "hb_buyid_pubmatic"
	skAdnetworkKey           = "skadn"
	rewardKey                = "reward"
	ImpExtAdUnitKey          = "dfp_ad_unit_code"
	AdServerGAM              = "gam"
)

type PubmaticAdapter struct {
	http *adapters.HTTPAdapter
	URI  string
}

// used for cookies and such
func (a *PubmaticAdapter) Name() string {
	return "pubmatic"
}

func (a *PubmaticAdapter) SkipNoCookies() bool {
	return false
}

// Below is bidder specific parameters for pubmatic adaptor,
// PublisherId and adSlot are mandatory parameters, others are optional parameters
// Keywords is bid specific parameter,
// WrapExt needs to be sent once per bid request
type pubmaticParams struct {
	PublisherId string            `json:"publisherId"`
	AdSlot      string            `json:"adSlot"`
	WrapExt     json.RawMessage   `json:"wrapper,omitempty"`
	Keywords    map[string]string `json:"keywords,omitempty"`
}

type pubmaticWrapperExt struct {
	ProfileID    int    `json:"profile,omitempty"`
	VersionID    int    `json:"version,omitempty"`
	WrapperImpID string `json:"wiid,omitempty"`
}

type pubmaticBidExtVideo struct {
	Duration *int `json:"duration,omitempty"`
}

type pubmaticBidExt struct {
	BidType           *int                 `json:"BidType,omitempty"`
	VideoCreativeInfo *pubmaticBidExtVideo `json:"video,omitempty"`
}

type ExtImpBidderPubmatic struct {
	adapters.ExtImpBidder
	Data *ExtData `json:"data,omitempty"`
}

type ExtData struct {
	AdServer *ExtAdServer `json:"adserver"`
	PBAdSlot string       `json:"pbadslot"`
}

type ExtAdServer struct {
	Name   string `json:"name"`
	AdSlot string `json:"adslot"`
}

const (
	INVALID_PARAMS    = "Invalid BidParam"
	MISSING_PUBID     = "Missing PubID"
	MISSING_ADSLOT    = "Missing AdSlot"
	INVALID_WRAPEXT   = "Invalid WrapperExt"
	INVALID_ADSIZE    = "Invalid AdSize"
	INVALID_WIDTH     = "Invalid Width"
	INVALID_HEIGHT    = "Invalid Height"
	INVALID_MEDIATYPE = "Invalid MediaType"
	INVALID_ADSLOT    = "Invalid AdSlot"

	dctrKeyName              = "key_val"
	dctrKeywordName          = "dctr"
	pmZoneIDKeyName          = "pmZoneId"
	pmZoneIDRequestParamName = "pmzoneid"

	urlEncodedEqualChar = "%3D"
)

func PrepareLogMessage(tID, pubId, adUnitId, bidID, details string, args ...interface{}) string {
	return fmt.Sprintf("%s ReqID [%s] PubID [%s] AdUnit [%s] BidID [%s] %s \n",
		PUBMATIC, tID, pubId, adUnitId, bidID, details)
}

func (a *PubmaticAdapter) Call(ctx context.Context, req *pbs.PBSRequest, bidder *pbs.PBSBidder) (pbs.PBSBidSlice, error) {
	mediaTypes := []pbs.MediaType{pbs.MEDIA_TYPE_BANNER, pbs.MEDIA_TYPE_VIDEO}
	pbReq, err := adapters.MakeOpenRTBGeneric(req, bidder, a.Name(), mediaTypes)

	if err != nil {
		logf("%s Failed to make ortb request for request id [%s] \n", PUBMATIC, pbReq.ID)
		return nil, err
	}

	var errState []string
	adSlotFlag := false
	pubId := ""
	wrapExt := ""
	if len(bidder.AdUnits) > MAX_IMPRESSIONS_PUBMATIC {
		logf("%s First %d impressions will be considered from request tid %s\n",
			PUBMATIC, MAX_IMPRESSIONS_PUBMATIC, pbReq.ID)
	}

	for i, unit := range bidder.AdUnits {
		var params pubmaticParams
		err := json.Unmarshal(unit.Params, &params)
		if err != nil {
			errState = append(errState, fmt.Sprintf("BidID:%s;Error:%s;param:%s", unit.BidID, INVALID_PARAMS, unit.Params))
			logf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
				fmt.Sprintf("Ignored bid: invalid JSON  [%s] err [%s]", unit.Params, err.Error())))
			continue
		}

		if params.PublisherId == "" {
			errState = append(errState, fmt.Sprintf("BidID:%s;Error:%s;param:%s", unit.BidID, MISSING_PUBID, unit.Params))
			logf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
				fmt.Sprintf("Ignored bid: Publisher Id missing")))
			continue
		}
		pubId = params.PublisherId

		if params.AdSlot == "" {
			errState = append(errState, fmt.Sprintf("BidID:%s;Error:%s;param:%s", unit.BidID, MISSING_ADSLOT, unit.Params))
			logf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
				fmt.Sprintf("Ignored bid: adSlot missing")))
			continue
		}

		// Parse Wrapper Extension i.e. ProfileID and VersionID only once per request
		if wrapExt == "" && len(params.WrapExt) != 0 {
			var wrapExtMap map[string]int
			err := json.Unmarshal([]byte(params.WrapExt), &wrapExtMap)
			if err != nil {
				errState = append(errState, fmt.Sprintf("BidID:%s;Error:%s;param:%s", unit.BidID, INVALID_WRAPEXT, unit.Params))
				logf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
					fmt.Sprintf("Ignored bid: Wrapper Extension Invalid")))
				continue
			}
			wrapExt = string(params.WrapExt)
		}

		adSlotStr := strings.TrimSpace(params.AdSlot)
		adSlot := strings.Split(adSlotStr, "@")
		if len(adSlot) == 2 && adSlot[0] != "" && adSlot[1] != "" {
			// Fixes some segfaults. Since this is legacy code, I'm not looking into it too deeply
			if len(pbReq.Imp) <= i {
				break
			}
			if pbReq.Imp[i].Banner != nil {
				adSize := strings.Split(strings.ToLower(strings.TrimSpace(adSlot[1])), "x")
				if len(adSize) == 2 {
					width, err := strconv.Atoi(strings.TrimSpace(adSize[0]))
					if err != nil {
						errState = append(errState, fmt.Sprintf("BidID:%s;Error:%s;param:%s", unit.BidID, INVALID_WIDTH, unit.Params))
						logf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
							fmt.Sprintf("Ignored bid: invalid adSlot width [%s]", adSize[0])))
						continue
					}

					heightStr := strings.Split(strings.TrimSpace(adSize[1]), ":")
					height, err := strconv.Atoi(strings.TrimSpace(heightStr[0]))
					if err != nil {
						errState = append(errState, fmt.Sprintf("BidID:%s;Error:%s;param:%s", unit.BidID, INVALID_HEIGHT, unit.Params))
						logf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
							fmt.Sprintf("Ignored bid: invalid adSlot height [%s]", heightStr[0])))
						continue
					}

					pbReq.Imp[i].TagID = strings.TrimSpace(adSlot[0])
					pbReq.Imp[i].Banner.W = openrtb2.Int64Ptr(int64(width))
					pbReq.Imp[i].Banner.H = openrtb2.Int64Ptr(int64(height))

					if len(params.Keywords) != 0 {
						kvstr := prepareImpressionExt(params.Keywords)
						pbReq.Imp[i].Ext = json.RawMessage([]byte(kvstr))
					} else {
						pbReq.Imp[i].Ext = nil
					}

					adSlotFlag = true
				} else {
					errState = append(errState, fmt.Sprintf("BidID:%s;Error:%s;param:%s", unit.BidID, INVALID_ADSIZE, unit.Params))
					logf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
						fmt.Sprintf("Ignored bid: invalid adSize [%s]", adSize)))
					continue
				}
			} else {
				errState = append(errState, fmt.Sprintf("BidID:%s;Error:%s;param:%s", unit.BidID, INVALID_MEDIATYPE, unit.Params))
				logf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
					fmt.Sprintf("Ignored bid: invalid Media Type")))
				continue
			}
		} else {
			errState = append(errState, fmt.Sprintf("BidID:%s;Error:%s;param:%s", unit.BidID, INVALID_ADSLOT, unit.Params))
			logf(PrepareLogMessage(pbReq.ID, params.PublisherId, unit.Code, unit.BidID,
				fmt.Sprintf("Ignored bid: invalid adSlot [%s]", params.AdSlot)))
			continue
		}

		if pbReq.Site != nil {
			siteCopy := *pbReq.Site
			siteCopy.Publisher = &openrtb2.Publisher{ID: params.PublisherId, Domain: req.Domain}
			pbReq.Site = &siteCopy
		}
		if pbReq.App != nil {
			appCopy := *pbReq.App
			appCopy.Publisher = &openrtb2.Publisher{ID: params.PublisherId, Domain: req.Domain}
			pbReq.App = &appCopy
		}
	}

	if !(adSlotFlag) {
		return nil, &errortypes.BadInput{
			Message: "Incorrect adSlot / Publisher params, Error list: [" + strings.Join(errState, ",") + "]",
		}
	}

	if wrapExt != "" {
		rawExt := fmt.Sprintf("{\"wrapper\": %s}", wrapExt)
		pbReq.Ext = json.RawMessage(rawExt)
	}

	reqJSON, err := json.Marshal(pbReq)

	debug := &pbs.BidderDebug{
		RequestURI: a.URI,
	}

	if req.IsDebug {
		debug.RequestBody = string(reqJSON)
		bidder.Debug = append(bidder.Debug, debug)
	}

	userId, _, _ := req.Cookie.GetUID(a.Name())
	httpReq, err := http.NewRequest("POST", a.URI, bytes.NewBuffer(reqJSON))
	httpReq.Header.Add("Content-Type", "application/json;charset=utf-8")
	httpReq.Header.Add("Accept", "application/json")
	httpReq.AddCookie(&http.Cookie{
		Name:  "KADUSERCOOKIE",
		Value: userId,
	})

	pbResp, err := ctxhttp.Do(ctx, a.http.Client, httpReq)
	if err != nil {
		return nil, err
	}

	debug.StatusCode = pbResp.StatusCode

	if pbResp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if pbResp.StatusCode == http.StatusBadRequest {
		return nil, &errortypes.BadInput{
			Message: fmt.Sprintf("HTTP status: %d", pbResp.StatusCode),
		}
	}

	if pbResp.StatusCode != http.StatusOK {
		return nil, &errortypes.BadServerResponse{
			Message: fmt.Sprintf("HTTP status: %d", pbResp.StatusCode),
		}
	}

	defer pbResp.Body.Close()
	body, err := ioutil.ReadAll(pbResp.Body)
	if err != nil {
		return nil, err
	}

	if req.IsDebug {
		debug.ResponseBody = string(body)
	}

	var bidResp openrtb2.BidResponse
	err = json.Unmarshal(body, &bidResp)
	if err != nil {
		return nil, &errortypes.BadServerResponse{
			Message: fmt.Sprintf("HTTP status: %d", pbResp.StatusCode),
		}
	}

	bids := make(pbs.PBSBidSlice, 0)

	numBids := 0
	for _, sb := range bidResp.SeatBid {
		for _, bid := range sb.Bid {
			numBids++

			bidID := bidder.LookupBidID(bid.ImpID)
			if bidID == "" {
				return nil, &errortypes.BadServerResponse{
					Message: fmt.Sprintf("Unknown ad unit code '%s'", bid.ImpID),
				}
			}

			pbid := pbs.PBSBid{
				BidID:       bidID,
				AdUnitCode:  bid.ImpID,
				BidderCode:  bidder.BidderCode,
				Price:       bid.Price,
				Adm:         bid.AdM,
				Creative_id: bid.CrID,
				Width:       bid.W,
				Height:      bid.H,
				DealId:      bid.DealID,
			}

			var bidExt pubmaticBidExt
			mediaType := openrtb_ext.BidTypeBanner
			if err := json.Unmarshal(bid.Ext, &bidExt); err == nil {
				mediaType = getBidType(&bidExt)
			}
			pbid.CreativeMediaType = string(mediaType)
			bids = append(bids, &pbid)
			logf("%s Returned Bid for PubID [%s] AdUnit [%s] BidID [%s] Size [%dx%d] Price [%f] \n",
				PUBMATIC, pubId, pbid.AdUnitCode, pbid.BidID, pbid.Width, pbid.Height, pbid.Price)
		}
	}

	return bids, nil
}

func getBidderParam(request *openrtb2.BidRequest, key string) ([]byte, error) {
	var reqExt openrtb_ext.ExtRequest
	if len(request.Ext) <= 0 {
		return nil, nil
	}
	err := json.Unmarshal(request.Ext, &reqExt)
	if err != nil {
		err := fmt.Errorf("%s Error unmarshalling request.ext: %v", PUBMATIC, string(request.Ext))
		return nil, err
	}

	if reqExt.Prebid.BidderParams == nil {
		return nil, nil
	}

	bidderParams, ok := reqExt.Prebid.BidderParams.(map[string]interface{})
	if !ok {
		err := fmt.Errorf("%s Error retrieving request.ext.prebid.ext: %v", PUBMATIC, reqExt.Prebid.BidderParams)
		return nil, err
	}

	iface, ok := bidderParams[key]
	if !ok {
		return nil, nil
	}

	bytes, err := json.Marshal(iface)
	if err != nil {
		err := fmt.Errorf("%s Error retrieving '%s' from request.ext.prebid.ext: %v", PUBMATIC, key, bidderParams)
		return nil, err
	}

	return bytes, nil
}

func getCookiesFromRequest(request *openrtb2.BidRequest) ([]string, error) {
	cbytes, err := getBidderParam(request, "Cookie")
	if err != nil {
		return nil, err
	}

	if cbytes == nil {
		return nil, nil
	}

	var cookies []string
	err = json.Unmarshal(cbytes, &cookies)
	if err != nil {
		err := fmt.Errorf("%s Error unmarshalling retrieving cookies from request.ext.prebid.ext: %v", PUBMATIC, string(cbytes))
		return nil, err
	}

	return cookies, nil
}

func (a *PubmaticAdapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	errs := make([]error, 0, len(request.Imp))

	var err error
	wrapperExt := new(pubmaticWrapperExt)
	pubID := ""
	wrapperExtSet := false

	cookies, err := getCookiesFromRequest(request)
	if err != nil {
		errs = append(errs, err)
	}

	for i := 0; i < len(request.Imp); i++ {
		err = parseImpressionObject(&request.Imp[i], wrapperExt, &pubID, &wrapperExtSet)
		// If the parsing is failed, remove imp and add the error.
		if err != nil {
			errs = append(errs, err)
			request.Imp = append(request.Imp[:i], request.Imp[i+1:]...)
			i--
		}
	}

	//if wrapper ext is set, then add it in request.ext
	if wrapperExtSet {
		//get wrapper impression ID from bidder params
		if wbytes, err := getBidderParam(request, "wiid"); err == nil && len(wbytes) != 0 {
			wrapperExt.WrapperImpID, _ = strconv.Unquote(string(wbytes))
		}
		jsonData, _ := json.Marshal(wrapperExt)
		rawExt := fmt.Sprintf("{\"wrapper\": %s}", string(jsonData))
		request.Ext = json.RawMessage(rawExt)
	}

	if request.Site != nil {
		siteCopy := *request.Site
		if siteCopy.Publisher != nil {
			publisherCopy := *siteCopy.Publisher
			publisherCopy.ID = pubID
			siteCopy.Publisher = &publisherCopy
		} else {
			siteCopy.Publisher = &openrtb2.Publisher{ID: pubID}
		}
		request.Site = &siteCopy
	} else if request.App != nil {
		appCopy := *request.App
		if appCopy.Publisher != nil {
			publisherCopy := *appCopy.Publisher
			publisherCopy.ID = pubID
			appCopy.Publisher = &publisherCopy
		} else {
			appCopy.Publisher = &openrtb2.Publisher{ID: pubID}
		}
		request.App = &appCopy
	}

	// move user.ext.eids to user.eids
	if request.User != nil && request.User.Ext != nil {
		var userExt *openrtb_ext.ExtUser
		if err = json.Unmarshal(request.User.Ext, &userExt); err == nil {
			if userExt != nil && userExt.Eids != nil {
				var eidArr []openrtb2.Eid
				for _, eid := range userExt.Eids {
					//var newEid openrtb2.Eid
					newEid := &openrtb2.Eid{
						ID:     eid.ID,
						Source: eid.Source,
						Ext:    eid.Ext,
					}
					var uidArr []openrtb2.Uid
					for _, uid := range eid.Uids {
						newUID := &openrtb2.Uid{
							ID:    uid.ID,
							AType: uid.Atype,
							Ext:   uid.Ext,
						}
						uidArr = append(uidArr, *newUID)
					}
					newEid.Uids = uidArr
					eidArr = append(eidArr, *newEid)
				}
				request.User.Eids = eidArr
				userExt.Eids = nil
				updatedUserExt, err1 := json.Marshal(userExt)
				if err1 == nil {
					request.User.Ext = updatedUserExt
				}
			}
		}
	}

	//adding hack to support DNT, since hbopenbid does not support lmt
	if request.Device != nil && request.Device.Lmt != nil && *request.Device.Lmt != 0 {
		request.Device.DNT = request.Device.Lmt
	}
	thisURI := a.URI

	// If all the requests are invalid, Call to adaptor is skipped
	if len(request.Imp) == 0 {
		return nil, errs
	}

	reqJSON, err := json.Marshal(request)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")
	headers.Add("Accept", "application/json")
	for _, line := range cookies {
		headers.Add("Cookie", line)
	}

	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     thisURI,
		Body:    reqJSON,
		Headers: headers,
	}}, errs
}

// validateAdslot validate the optional adslot string
// valid formats are 'adslot@WxH', 'adslot' and no adslot
func validateAdSlot(adslot string, imp *openrtb2.Imp) error {
	adSlotStr := strings.TrimSpace(adslot)

	if len(adSlotStr) == 0 {
		return nil
	}

	if !strings.Contains(adSlotStr, "@") {
		imp.TagID = adSlotStr
		return nil
	}

	adSlot := strings.Split(adSlotStr, "@")
	if len(adSlot) == 2 && adSlot[0] != "" && adSlot[1] != "" {
		imp.TagID = strings.TrimSpace(adSlot[0])

		adSize := strings.Split(strings.ToLower(adSlot[1]), "x")
		if len(adSize) != 2 {
			return errors.New(fmt.Sprintf("Invalid size provided in adSlot %v", adSlotStr))
		}

		width, err := strconv.Atoi(strings.TrimSpace(adSize[0]))
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid width provided in adSlot %v", adSlotStr))
		}

		heightStr := strings.Split(adSize[1], ":")
		height, err := strconv.Atoi(strings.TrimSpace(heightStr[0]))
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid height provided in adSlot %v", adSlotStr))
		}

		//In case of video, size could be derived from the player size
		if imp.Banner != nil {
			imp.Banner = assignBannerWidthAndHeight(imp.Banner, int64(width), int64(height))
		}
	} else {
		return errors.New(fmt.Sprintf("Invalid adSlot %v", adSlotStr))
	}

	return nil
}

func assignBannerSize(banner *openrtb2.Banner) (*openrtb2.Banner, error) {
	if banner.W != nil && banner.H != nil {
		return banner, nil
	}

	if len(banner.Format) == 0 {
		return nil, errors.New(fmt.Sprintf("No sizes provided for Banner %v", banner.Format))
	}

	return assignBannerWidthAndHeight(banner, banner.Format[0].W, banner.Format[0].H), nil
}

func assignBannerWidthAndHeight(banner *openrtb2.Banner, w, h int64) *openrtb2.Banner {
	bannerCopy := *banner
	bannerCopy.W = openrtb2.Int64Ptr(w)
	bannerCopy.H = openrtb2.Int64Ptr(h)
	return &bannerCopy
}

// parseImpressionObject parse the imp to get it ready to send to pubmatic
func parseImpressionObject(imp *openrtb2.Imp, wrapExt *pubmaticWrapperExt, pubID *string, wrapperExtSet *bool) error {
	// PubMatic supports banner and video impressions.
	if imp.Banner == nil && imp.Video == nil {
		return fmt.Errorf("Invalid MediaType. PubMatic only supports Banner and Video. Ignoring ImpID=%s", imp.ID)
	}

	if imp.Audio != nil {
		imp.Audio = nil
	}

	var bidderExt ExtImpBidderPubmatic
	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return err
	}

	var pubmaticExt openrtb_ext.ExtImpPubmatic
	if err := json.Unmarshal(bidderExt.Bidder, &pubmaticExt); err != nil {
		return err
	}

	if *pubID == "" {
		*pubID = strings.TrimSpace(pubmaticExt.PublisherId)
	}

	// Parse Wrapper Extension only once per request
	if !*wrapperExtSet && len(pubmaticExt.WrapExt) != 0 {
		err := json.Unmarshal([]byte(pubmaticExt.WrapExt), &wrapExt)
		if err != nil {
			return fmt.Errorf("Error in Wrapper Parameters = %v  for ImpID = %v WrapperExt = %v", err.Error(), imp.ID, string(pubmaticExt.WrapExt))
		}
		*wrapperExtSet = true
	}

	if err := validateAdSlot(strings.TrimSpace(pubmaticExt.AdSlot), imp); err != nil {
		return err
	}

	if imp.Banner != nil {
		bannerCopy, err := assignBannerSize(imp.Banner)
		if err != nil {
			return err
		}
		imp.Banner = bannerCopy
	}

	impExtMap := make(map[string]interface{}, 0)
	if pubmaticExt.Keywords != nil && len(pubmaticExt.Keywords) != 0 {
		addKeywordsToExt(pubmaticExt.Keywords, impExtMap)
	}
	//Give preference to direct values of 'dctr' & 'pmZoneId' params in extension
	if pubmaticExt.Dctr != "" {
		impExtMap[dctrKeyName] = pubmaticExt.Dctr
	}
	if pubmaticExt.PmZoneID != "" {
		impExtMap[pmZoneIDKeyName] = pubmaticExt.PmZoneID
	}

	if bidderExt.Prebid != nil {
		if bidderExt.Prebid.SKAdnetwork != nil {
			impExtMap[skAdnetworkKey] = bidderExt.Prebid.SKAdnetwork
		}
		if bidderExt.Prebid.IsRewardedInventory == 1 {
			impExtMap[rewardKey] = bidderExt.Prebid.IsRewardedInventory
		}
	}

	if bidderExt.Data != nil {
		if bidderExt.Data.AdServer != nil && bidderExt.Data.AdServer.Name == AdServerGAM && bidderExt.Data.AdServer.AdSlot != "" {
			impExtMap[ImpExtAdUnitKey] = bidderExt.Data.AdServer.AdSlot
		} else if bidderExt.Data.PBAdSlot != "" {
			impExtMap[ImpExtAdUnitKey] = bidderExt.Data.PBAdSlot
		}
	}

	imp.Ext = nil
	if len(impExtMap) > 0 {
		impExtBytes, err := json.Marshal(impExtMap)
		if err == nil {
			imp.Ext = impExtBytes
		}
	}

	return nil

}

func addKeywordsToExt(keywords []*openrtb_ext.ExtImpPubmaticKeyVal, extMap map[string]interface{}) {
	for _, keyVal := range keywords {
		if len(keyVal.Values) == 0 {
			logf("No values present for key = %s", keyVal.Key)
			continue
		} else {
			val := strings.Join(keyVal.Values[:], ",")

			key := keyVal.Key
			if strings.EqualFold(key, pmZoneIDRequestParamName) {
				key = pmZoneIDKeyName
			} else if key == dctrKeywordName {
				key = dctrKeyName
				// URL-decode dctr value if it is url-encoded
				if strings.Contains(val, urlEncodedEqualChar) {
					urlDecodedVal, err := url.QueryUnescape(val)
					if err == nil {
						val = urlDecodedVal
					}
				}
			}

			extMap[key] = val
		}
	}
}

func prepareImpressionExt(keywords map[string]string) string {

	eachKv := make([]string, 0, len(keywords))
	for key, val := range keywords {
		if len(val) == 0 {
			logf("No values present for key = %s", key)
			continue
		} else {
			eachKv = append(eachKv, fmt.Sprintf("\"%s\":\"%s\"", key, val))
		}
	}

	kvStr := "{" + strings.Join(eachKv, ",") + "}"
	return kvStr
}

func (a *PubmaticAdapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{fmt.Errorf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode)}
	}

	var bidResp openrtb2.BidResponse
	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{err}
	}

	bidResponse := adapters.NewBidderResponseWithBidsCapacity(5)

	var errs []error
	for _, sb := range bidResp.SeatBid {
		targets := getTargetingKeys(sb.Ext)
		for i := 0; i < len(sb.Bid); i++ {
			bid := sb.Bid[i]
			// Copy SeatBid Ext to Bid.Ext
			bid.Ext = copySBExtToBidExt(sb.Ext, bid.Ext)

			impVideo := &openrtb_ext.ExtBidPrebidVideo{}

			if len(bid.Cat) > 1 {
				bid.Cat = bid.Cat[0:1]
			}

			var bidExt *pubmaticBidExt
			bidType := openrtb_ext.BidTypeBanner
			if err := json.Unmarshal(bid.Ext, &bidExt); err == nil && bidExt != nil {
				if bidExt.VideoCreativeInfo != nil && bidExt.VideoCreativeInfo.Duration != nil {
					impVideo.Duration = *bidExt.VideoCreativeInfo.Duration
				}
				bidType = getBidType(bidExt)
			}

			bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
				Bid:        &bid,
				BidType:    bidType,
				BidVideo:   impVideo,
				BidTargets: targets,
			})
		}
	}
	return bidResponse, errs
}

// getBidType returns the bid type specified in the response bid.ext
func getBidType(bidExt *pubmaticBidExt) openrtb_ext.BidType {
	// setting "banner" as the default bid type
	bidType := openrtb_ext.BidTypeBanner
	if bidExt != nil && bidExt.BidType != nil {
		switch *bidExt.BidType {
		case 0:
			bidType = openrtb_ext.BidTypeBanner
		case 1:
			bidType = openrtb_ext.BidTypeVideo
		case 2:
			bidType = openrtb_ext.BidTypeNative
		default:
			// default value is banner
			bidType = openrtb_ext.BidTypeBanner
		}
	}
	return bidType
}

func logf(msg string, args ...interface{}) {
	if glog.V(2) {
		glog.Infof(msg, args...)
	}
}

func NewPubmaticLegacyAdapter(config *adapters.HTTPAdapterConfig, uri string) *PubmaticAdapter {
	a := adapters.NewHTTPAdapter(config)

	return &PubmaticAdapter{
		http: a,
		URI:  uri,
	}
}

// Builder builds a new instance of the Pubmatic adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	bidder := &PubmaticAdapter{
		URI: config.Endpoint,
	}
	return bidder, nil
}

func getTargetingKeys(bidExt json.RawMessage) map[string]string {
	targets := map[string]string{}
	if bidExt != nil {
		bidExtMap := make(map[string]interface{})
		err := json.Unmarshal(bidExt, &bidExtMap)
		if err == nil && bidExtMap[buyId] != nil {
			targets[buyIdTargetingKey] = string(bidExtMap[buyId].(string))
		}
	}
	return targets
}

func copySBExtToBidExt(sbExt json.RawMessage, bidExt json.RawMessage) json.RawMessage {
	if sbExt != nil {
		sbExtMap := getMapFromJSON(sbExt)
		bidExtMap := make(map[string]interface{})
		if bidExt != nil {
			bidExtMap = getMapFromJSON(bidExt)
		}
		if bidExtMap != nil && sbExtMap != nil {
			if sbExtMap[buyId] != nil && bidExtMap[buyId] == nil {
				bidExtMap[buyId] = sbExtMap[buyId]
			}
		}
		byteAra, _ := json.Marshal(bidExtMap)
		return json.RawMessage(byteAra)
	}
	return bidExt
}

func getMapFromJSON(ext json.RawMessage) map[string]interface{} {
	if ext != nil {
		extMap := make(map[string]interface{})
		err := json.Unmarshal(ext, &extMap)
		if err == nil {
			return extMap
		}
	}
	return nil
}

func head(s []string) string {
	if len(s) == 0 {
		return ""
	}

	return s[0]
}
