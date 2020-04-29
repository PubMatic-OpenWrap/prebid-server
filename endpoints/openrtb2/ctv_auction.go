package openrtb2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PubMatic-OpenWrap/etree"
	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/analytics"
	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/exchange"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/PubMatic-OpenWrap/prebid-server/pbsmetrics"
	"github.com/PubMatic-OpenWrap/prebid-server/stored_requests"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
	uuid "github.com/gofrs/uuid"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
)

//ImpBid type of data to be present for combinations
type ImpBid struct {
	*openrtb.Bid
	OriginalImpID  string
	SequenceNumber int
	SeatName       string
}

//AdPodBid combination contains ImpBid
type AdPodBid []*ImpBid

//AdPodBids combination contains ImpBid
type AdPodBids []AdPodBid

//BidsMap map of impression with adpod details
type BidsMap map[string]AdPodBid

//ImpAdPodConfig configuration for creating ads in adpod
type ImpAdPodConfig struct {
	ImpID          string
	SequenceNumber int8
	MinDuration    int64
	MaxDuration    int64
}

//ImpData example
type ImpData struct {
	VideoExt openrtb_ext.VideoExtension
	Config   []*ImpAdPodConfig
	Bids     AdPodBid
	//AdPodGenerator
}

//CTV Specific Endpoint
type ctvEndpointDeps struct {
	endpointDeps
	request *openrtb.BidRequest
	reqExt  openrtb_ext.ReqAdPodExt
	impData []*ImpData
}

//NewCTVEndpoint new ctv endpoint object
func NewCTVEndpoint(
	ex exchange.Exchange,
	validator openrtb_ext.BidderParamValidator,
	requestsByID stored_requests.Fetcher,
	videoFetcher stored_requests.Fetcher,
	categories stored_requests.CategoryFetcher,
	cfg *config.Configuration,
	met pbsmetrics.MetricsEngine,
	pbsAnalytics analytics.PBSAnalyticsModule,
	disabledBidders map[string]string,
	defReqJSON []byte,
	bidderMap map[string]openrtb_ext.BidderName) (httprouter.Handle, error) {

	if ex == nil || validator == nil || requestsByID == nil || cfg == nil || met == nil {
		return nil, errors.New("NewCTVEndpoint requires non-nil arguments.")
	}
	defRequest := defReqJSON != nil && len(defReqJSON) > 0

	return httprouter.Handle((&ctvEndpointDeps{
		endpointDeps: endpointDeps{
			ex,
			validator,
			requestsByID,
			videoFetcher,
			categories,
			cfg,
			met,
			pbsAnalytics,
			disabledBidders,
			defRequest,
			defReqJSON,
			bidderMap,
		},
	}).CTVAuctionEndpoint), nil
}

func (deps *ctvEndpointDeps) CTVAuctionEndpoint(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	ao := analytics.AuctionObject{
		Status: http.StatusOK,
		Errors: make([]error, 0),
	}

	// Prebid Server interprets request.tmax to be the maximum amount of time that a caller is willing
	// to wait for bids. However, tmax may be defined in the Stored Request data.
	//
	// If so, then the trip to the backend might use a significant amount of this time.
	// We can respect timeouts more accurately if we note the *real* start time, and use it
	// to compute the auction timeout.
	start := time.Now()
	//Prebid Stats
	labels := pbsmetrics.Labels{
		Source:        pbsmetrics.DemandUnknown,
		RType:         pbsmetrics.ReqTypeVideo,
		PubID:         pbsmetrics.PublisherUnknown,
		Browser:       getBrowserName(r),
		CookieFlag:    pbsmetrics.CookieFlagUnknown,
		RequestStatus: pbsmetrics.RequestStatusOK,
	}
	defer func() {
		deps.metricsEngine.RecordRequest(labels)
		deps.metricsEngine.RecordRequestTime(labels, time.Since(start))
		deps.analytics.LogAuctionObject(&ao)
	}()

	//Parse ORTB Request and do Standard Validation
	req, errL := deps.parseRequest(r)
	if fatalError(errL) && writeError(errL, w, &labels) {
		return
	}

	jsonlog("Original BidRequest", req) //TODO: REMOVE LOG

	//init
	deps.init(req)

	//Set Default Values
	deps.setDefaultValues()

	//Validate CTV BidRequest
	if err := deps.validateBidRequest(); err != nil {
		errL = append(errL, err...)
		writeError(errL, w, &labels)
		return
	}
	jsonlog("Request Extension", deps.reqExt)
	jsonlog("ImpData", deps.impData)

	//Create New BidRequest
	ctvReq := deps.createBidRequest(req)
	jsonlog("CTV BidRequest", ctvReq) //TODO: REMOVE LOG

	ctx := context.Background()

	//Setting Timeout for Request
	timeout := deps.cfg.AuctionTimeouts.LimitAuctionTimeout(time.Duration(ctvReq.TMax) * time.Millisecond)
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, start.Add(timeout))
		defer cancel()
	}

	//Parsing Cookies and Set Stats
	usersyncs := usersync.ParsePBSCookieFromRequest(r, &(deps.cfg.HostCookie))
	if ctvReq.App != nil {
		labels.Source = pbsmetrics.DemandApp
		labels.RType = pbsmetrics.ReqTypeVideo
		labels.PubID = effectivePubID(ctvReq.App.Publisher)
	} else { //ctvReq.Site != nil
		labels.Source = pbsmetrics.DemandWeb
		if usersyncs.LiveSyncCount() == 0 {
			labels.CookieFlag = pbsmetrics.CookieFlagNo
		} else {
			labels.CookieFlag = pbsmetrics.CookieFlagYes
		}
		labels.PubID = effectivePubID(ctvReq.Site.Publisher)
	}

	//Validate Accounts
	if err := validateAccount(deps.cfg, labels.PubID); err != nil {
		errL = append(errL, err)
		writeError(errL, w, &labels)
		return
	}

	//Hold OpenRTB Standard Auction
	response, err := deps.ex.HoldAuction(ctx, ctvReq, usersyncs, labels, &deps.categories)
	ao.Request = ctvReq
	ao.Response = response
	if err != nil {
		labels.RequestStatus = pbsmetrics.RequestStatusErr
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Critical error while running the auction: %v", err)
		glog.Errorf("/openrtb2/video Critical error: %v", err)
		ao.Status = http.StatusInternalServerError
		ao.Errors = append(ao.Errors, err)
		return
	}
	jsonlog("BidResponse", response) //TODO: REMOVE LOG

	//Validate Bid Response
	if err := deps.validateBidResponse(ctvReq, response); err != nil {
		errL = append(errL, err)
		writeError(errL, w, &labels)
		return
	}

	//Create Impression Bids
	deps.getBids(response)

	//Do AdPod Exclusions
	bids := deps.doAdPodExclusions()

	//Create Bid Response
	ctvResp := deps.createBidResponse(response, bids)
	jsonlog("CTV BidResponse", ctvResp) //TODO: REMOVE LOG

	// Response Generation
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	// Fixes #328
	w.Header().Set("Content-Type", "application/json")

	// If an error happens when encoding the response, there isn't much we can do.
	// If we've sent _any_ bytes, then Go would have sent the 200 status code first.
	// That status code can't be un-sent... so the best we can do is log the error.
	if err := enc.Encode(ctvResp); err != nil {
		labels.RequestStatus = pbsmetrics.RequestStatusNetworkErr
		ao.Errors = append(ao.Errors, fmt.Errorf("/openrtb2/video Failed to send response: %v", err))
	}
}

func (deps *ctvEndpointDeps) init(req *openrtb.BidRequest) {
	deps.request = req
	deps.impData = make([]*ImpData, len(req.Imp))
	for i := range req.Imp {
		deps.impData[i] = &ImpData{}
	}
}

func (deps *ctvEndpointDeps) readVideoExtensions() (err []error) {
	for index, imp := range deps.request.Imp {
		if nil != imp.Video {
			if nil != imp.Video.Ext {
				errL := json.Unmarshal(imp.Video.Ext, &deps.impData[index].VideoExt)
				if nil != err {
					err = append(err, errL)
					continue
				}
			}

			pod := deps.impData[index].VideoExt.AdPod
			if nil == pod {
				pod = &openrtb_ext.VideoAdPod{}
				deps.impData[index].VideoExt.AdPod = pod
			}

			//Use Request Level Parameters
			pod.Merge(&deps.reqExt.VideoAdPod)

			//Set Default Values
			deps.impData[index].VideoExt.SetDefaultValue()
			pod.SetDefaultAdDurations(imp.Video.MinDuration, imp.Video.MaxDuration)
		}
	}
	return err
}

func (deps *ctvEndpointDeps) readRequestExtension() (err []error) {
	if nil != deps.request.Ext {
		errL := json.Unmarshal(deps.request.Ext, &deps.reqExt)
		if nil != err {
			err = append(err, errL)
			return
		}
		deps.reqExt.SetDefaultValue()
	}
	return
}

func (deps *ctvEndpointDeps) readExtensions() (err []error) {
	if errL := deps.readRequestExtension(); nil != errL {
		err = append(err, errL...)
	}

	if errL := deps.readVideoExtensions(); nil != errL {
		err = append(err, errL...)
	}
	return err
}

//setDefaultValues will set adpod and other default values
func (deps *ctvEndpointDeps) setDefaultValues() {
	//read and set extension values
	deps.readExtensions()

	//TODO: remove req.ext.adpod and req.imp.video.ext.adpod and offset parameter
}

//validateBidRequest will validate AdPod specific mandatory Parameters and returns error
func (deps *ctvEndpointDeps) validateBidRequest() (err []error) {
	//validating video extension adpod configurations
	err = deps.reqExt.Validate()

	for index := range deps.request.Imp {
		if errL := deps.impData[index].VideoExt.Validate(); nil != errL {
			err = append(err, errL...)
		}
		//TODO: Validate Invalid Configurations based on imp.video.minduration and imp.video.maxduration
	}
	return
}

//getAdPodImpsConfigs will return number of impressions configurations within adpod
func getAdPodImpsConfigs(imp *openrtb.Imp, adpod *openrtb_ext.VideoAdPod) []*ImpAdPodConfig {
	impCount := *adpod.MaxAds
	if impCount > 5 { //TODO: REMOVE HARDCODING
		impCount = 5
	}

	config := make([]*ImpAdPodConfig, impCount)
	for i := 0; i < impCount; i++ {
		config[i] = &ImpAdPodConfig{
			ImpID:          fmt.Sprintf("%s_%d", imp.ID, i+1),
			MinDuration:    imp.Video.MinDuration,
			MaxDuration:    imp.Video.MaxDuration,
			SequenceNumber: int8(i + 1), /* Must be starting with 1 */
		}
	}
	return config[:]
}

//getAllAdPodImpsConfigs will return all impression adpod configurations
func (deps *ctvEndpointDeps) getAllAdPodImpsConfigs() {
	for index, imp := range deps.request.Imp {
		if nil == imp.Video {
			continue
		}
		deps.impData[index].Config = getAdPodImpsConfigs(&imp, deps.impData[index].VideoExt.AdPod)
	}
}

//getBids reads bids from bidresponse object
func (deps *ctvEndpointDeps) getBids(resp *openrtb.BidResponse) {
	result := make(map[string]AdPodBid)

	for _, seat := range resp.SeatBid {
		for _, bid := range seat.Bid {
			originalImpID, sequence := decodeImpressionID(bid.ImpID)

			result[originalImpID] = append(result[originalImpID], &ImpBid{
				Bid:            &bid,
				SeatName:       seat.Seat,
				SequenceNumber: sequence,
				OriginalImpID:  originalImpID,
			})
		}
	}

	//Sort Bids by Price
	for index, imp := range deps.request.Imp {
		bids, ok := result[imp.ID]
		if ok {
			//sort bids
			sort.Slice(bids[:], func(i, j int) bool { return bids[i].Price > bids[j].Price })
			deps.impData[index].Bids = bids[:]
		}
	}
}

//createBidRequest will return new bid request with all things copy from bid request except impression objects
func (deps *ctvEndpointDeps) createBidRequest(req *openrtb.BidRequest) *openrtb.BidRequest {
	ctvRequest := *req

	//get configurations for all impressions
	deps.getAllAdPodImpsConfigs()

	//createImpressions
	ctvRequest.Imp = deps.createImpressions()

	//TODO: remove adpod extension if not required to send further
	return &ctvRequest
}

//createImpressions will create multiple impressions based on adpod configurations
func (deps *ctvEndpointDeps) createImpressions() []openrtb.Imp {
	impCount := 0
	for _, imp := range deps.impData {
		impCount = impCount + len(imp.Config)
	}

	count := 0
	imps := make([]openrtb.Imp, impCount)
	for index, imp := range deps.request.Imp {
		adPodConfig := deps.impData[index].Config
		for _, config := range adPodConfig {
			imps[count] = *(newImpression(&imp, config))
			count++
		}
	}

	return imps[:]
}

//newImpression will clone existing impression object and create video object with ImpAdPodConfig.
func newImpression(imp *openrtb.Imp, config *ImpAdPodConfig) *openrtb.Imp {
	video := *imp.Video
	video.MinDuration = config.MinDuration
	video.MaxDuration = config.MaxDuration
	video.Sequence = config.SequenceNumber
	video.MaxExtended = 0
	//TODO: remove video adpod extension if not required

	newImp := *imp
	newImp.ID = config.ImpID
	//newImp.BidFloor = 0
	newImp.Video = &video
	return &newImp
}

//validateBidResponse
func (deps *ctvEndpointDeps) validateBidResponse(req *openrtb.BidRequest, resp *openrtb.BidResponse) error {
	//remove bids withoug cat and adomain
	//remove bids without bid.id
	//remove bids with price=0
	return nil
}

//doAdPodExclusions
func (deps *ctvEndpointDeps) doAdPodExclusions() AdPodBids {
	result := AdPodBids{}
	for index := 0; index < len(deps.request.Imp); index++ {
		bids := deps.impData[index].Bids
		if len(bids) > 0 {
			adpodGenerator := NewAdPodGenerator(bids[:], nil, func(x *ImpBid, y *ImpBid) bool {
				return true
			})
			adpod := adpodGenerator.GetAdPod()
			if adpod != nil {
				result = append(result, adpod)
			}
		}
	}
	return result
}

//createBidResponse
func (deps *ctvEndpointDeps) createBidResponse(resp *openrtb.BidResponse, adpods AdPodBids) *openrtb.BidResponse {
	bidResp := &openrtb.BidResponse{
		ID:  resp.ID,
		Ext: resp.Ext,
	}
	for _, adpod := range adpods {
		if len(adpod) == 0 {
			continue
		}
		bid := deps.getAdPodBid(adpod)
		if bid != nil {
			found := false
			for _, seat := range bidResp.SeatBid {
				if seat.Seat == adpod[0].SeatName {
					seat.Bid = append(seat.Bid, *bid)
					found = true
					break
				}
			}
			if found == false {
				bidResp.SeatBid = append(bidResp.SeatBid, openrtb.SeatBid{
					Seat: adpod[0].SeatName,
					Bid: []openrtb.Bid{
						*bid,
					},
				})
			}
		}
	}
	return bidResp
}

//getAdPodBid
func (deps *ctvEndpointDeps) getAdPodBid(adpod AdPodBid) *openrtb.Bid {
	bid := openrtb.Bid{}
	//TODO: Write single for loop to get all details
	bidID, err := uuid.NewV4()
	if nil == err {
		bid.ID = bidID.String()
	} else {
		bid.ID = adpod[0].ID
	}

	bid.ImpID = adpod[0].OriginalImpID
	bid.AdM = *getAdPodBidCreative(adpod)
	bid.Price = getAdPodBidPrice(adpod)
	bid.ADomain = getAdPodBidAdvertiserDomain(adpod)
	bid.Cat = getAdPodBidCategories(adpod)
	bid.Ext = getAdPodBidExtension(adpod)
	return &bid
}

//getAdPodBidCreative get commulative adpod bid details
func getAdPodBidCreative(adpod AdPodBid) *string {
	doc := etree.NewDocument()
	vast := doc.CreateElement("VAST")
	vast.CreateAttr("version", "3.0")
	sequenceNumber := 1
	for _, bid := range adpod {
		adDoc := etree.NewDocument()
		if err := adDoc.ReadFromString(bid.AdM); err != nil {
			continue
		}

		vastTag := adDoc.SelectElement("VAST")
		for _, ad := range vastTag.SelectElements("Ad") {
			newAd := ad.Copy()
			//newAd.CreateAttr("id", bid.OriginalImpID)
			//creative.AdId attribute needs to be updated
			newAd.CreateAttr("sequence", fmt.Sprint(sequenceNumber))
			vast.AddChild(newAd)
			sequenceNumber++
		}
	}
	bidAdM, err := doc.WriteToString()
	if nil != err {
		fmt.Printf("VIRAL ERROR, %v", err.Error())
		return &bidAdM
	}
	return &bidAdM
}

//getAdPodBidPrice get commulative adpod bid details
func getAdPodBidPrice(adpod AdPodBid) float64 {
	var price float64 = 0
	for _, ad := range adpod {
		price = price + ad.Price
	}
	return price
}

//getAdPodBidAdvertiserDomain get commulative adpod bid details
func getAdPodBidAdvertiserDomain(adpod AdPodBid) []string {
	var domains []string
	for _, ad := range adpod {
		domains = append(domains, ad.ADomain...)
	}
	//send unique domains only
	return domains[:]
}

//getAdPodBidCategories get commulative adpod bid details
func getAdPodBidCategories(adpod AdPodBid) []string {
	var category []string
	for _, ad := range adpod {
		if len(ad.Cat) > 0 {
			category = append(category, ad.Cat...)
		}
	}
	//send unique domains only
	return category[:]
}

//getAdPodBidExtension get commulative adpod bid details
func getAdPodBidExtension(adpod AdPodBid) json.RawMessage {
	return adpod[0].Ext
}

func decodeImpressionID(id string) (string, int) {
	values := strings.Split(id, "_")
	if len(values) == 1 {
		return values[0], 1
	}
	sequence, err := strconv.Atoi(values[1])
	if err != nil {
		sequence = 1
	}
	return values[0], sequence
}

//IAdPodGenerator interface for generating AdPod from Ads
type IAdPodGenerator interface {
	GetAdPod() AdPodBid
}

//Comparator check exclusion conditions
type Comparator func(*ImpBid, *ImpBid) bool

//AdPodGenerator AdPodGenerator
type AdPodGenerator struct {
	IAdPodGenerator
	bids   AdPodBid
	config *openrtb_ext.VideoAdPod
	comp   Comparator
}

//NewAdPodGenerator will generate adpod based on configuration
func NewAdPodGenerator(bids AdPodBid, config *openrtb_ext.VideoAdPod, comp Comparator) *AdPodGenerator {
	return &AdPodGenerator{
		bids:   bids,
		config: config,
		comp:   comp,
	}
}

//GetAdPod will return Adpod based on configurations
func (o *AdPodGenerator) GetAdPod() AdPodBid {
	var result AdPodBid
	count := 3
	for i, bid := range o.bids {
		if i >= count {
			break
		}
		result = append(result, bid)
	}
	return result[:]
}

func jsonlog(msg string, obj interface{}) {
	data, _ := json.Marshal(obj)
	glog.Infof("[OPENWRAP] %v:%v", msg, string(data))
}
