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
	"github.com/buger/jsonparser"
	uuid "github.com/gofrs/uuid"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
)

const (
	keyAdPod  = `adpod`
	keyOffset = `offset`
)

type Bid = openrtb.Bid

//AdPodBid combination contains ImpBid
type AdPodBid struct {
	Bids          []*Bid
	OriginalImpID string
	SeatName      string
}

//AdPodBids combination contains ImpBid
type AdPodBids []*AdPodBid

//ImpAdPodConfig configuration for creating ads in adpod
type ImpAdPodConfig struct {
	ImpID          string `json:"id,omitempty"`
	SequenceNumber int8   `json:"seq,omitempty"`
	MinDuration    int64  `json:"minduration,omitempty"`
	MaxDuration    int64  `json:"maxduration,omitempty"`
}

//ImpData example
type ImpData struct {
	//AdPodGenerator
	VideoExt openrtb_ext.VideoExtension
	Config   []*ImpAdPodConfig
	Bid      *AdPodBid
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

/********************* BidRequest Processing *********************/

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
			if len(imp.Video.Ext) > 0 {
				errL := json.Unmarshal(imp.Video.Ext, &deps.impData[index].VideoExt)
				if nil != err {
					err = append(err, errL)
					continue
				}

				imp.Video.Ext = jsonparser.Delete(imp.Video.Ext, keyAdPod)
				imp.Video.Ext = jsonparser.Delete(imp.Video.Ext, keyOffset)
				if string(imp.Video.Ext) == `{}` {
					imp.Video.Ext = nil
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
	if len(deps.request.Ext) > 0 {
		//TODO: use jsonparser library for get adpod and remove that key
		extAdPod, jsonType, _, errL := jsonparser.Get(deps.request.Ext, keyAdPod)
		if nil != err {
			//parsing error
			if jsonparser.NotExist != jsonType {
				//assuming key not present
				err = append(err, errL)
				return
			}
		} else {
			if errL := json.Unmarshal(extAdPod, &deps.reqExt); nil != errL {
				err = append(err, errL)
				return
			}

			//removing key from extensions
			deps.request.Ext = jsonparser.Delete(deps.request.Ext, keyAdPod)
			if string(deps.request.Ext) == `{}` {
				deps.request.Ext = nil
			}
		}
	}
	deps.reqExt.SetDefaultValue()
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

/********************* Creating CTV BidRequest *********************/

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

//getAllAdPodImpsConfigs will return all impression adpod configurations
func (deps *ctvEndpointDeps) getAllAdPodImpsConfigs() {
	for index, imp := range deps.request.Imp {
		if nil == imp.Video {
			continue
		}
		deps.impData[index].Config = getAdPodImpsConfigs(&imp, deps.impData[index].VideoExt.AdPod)
	}
}

//getAdPodImpsConfigs will return number of impressions configurations within adpod
func getAdPodImpsConfigs(imp *openrtb.Imp, adpod *openrtb_ext.VideoAdPod) []*ImpAdPodConfig {
	impRanges := getImpressions(imp.Video.MinDuration, imp.Video.MaxDuration, adpod)

	config := make([]*ImpAdPodConfig, len(impRanges))
	for i, value := range impRanges {
		config[i] = &ImpAdPodConfig{
			ImpID:          fmt.Sprintf("%s_%d", imp.ID, i+1),
			MinDuration:    int64(value[0]),
			MaxDuration:    int64(value[1]),
			SequenceNumber: int8(i + 1), /* Must be starting with 1 */
		}
	}
	return config[:]
}

//getImpressions will create number of impressions based on adpod configurations
func getImpressions(podMinDuration, podMaxDuration int64, adpod *openrtb_ext.VideoAdPod) (imps [][2]int) {
	_, imps = ctv.getImpressions(podMinDuration, podMaxDuration, adpod)
	return imps
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

/********************* Prebid BidResponse Processing *********************/

//validateBidResponse
func (deps *ctvEndpointDeps) validateBidResponse(req *openrtb.BidRequest, resp *openrtb.BidResponse) error {
	//remove bids withoug cat and adomain
	//remove bids without bid.id
	//remove bids with price=0
	return nil
}

//getBids reads bids from bidresponse object
func (deps *ctvEndpointDeps) getBids(resp *openrtb.BidResponse) {
	result := make(map[string]*AdPodBid)

	for _, seat := range resp.SeatBid {
		for _, bid := range seat.Bid {
			originalImpID, _ := decodeImpressionID(bid.ImpID)
			adpodBid, ok := result[originalImpID]
			if !ok {
				adpodBid = &AdPodBid{
					OriginalImpID: originalImpID,
					SeatName:      "pubmatic",
				}
				result[originalImpID] = adpodBid
			}
			adpodBid.Bids = append(adpodBid.Bids, &bid)
		}
	}

	//Sort Bids by Price
	for index, imp := range deps.request.Imp {
		adpodBid, ok := result[imp.ID]
		if ok {
			//sort bids
			sort.Slice(adpodBid.Bids[:], func(i, j int) bool { return adpodBid.Bids[i].Price > adpodBid.Bids[j].Price })
			deps.impData[index].Bid = adpodBid
		}
	}
}

//doAdPodExclusions
func (deps *ctvEndpointDeps) doAdPodExclusions() AdPodBids {
	result := AdPodBids{}
	for index := 0; index < len(deps.request.Imp); index++ {
		bid := deps.impData[index].Bid
		if nil != bid && len(bid.Bids) > 0 {
			adpodGenerator := NewAdPodGenerator(bid, nil, func(x *Bid, y *Bid) bool {
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

/********************* Creating CTV BidResponse *********************/

//createBidResponse
func (deps *ctvEndpointDeps) createBidResponse(resp *openrtb.BidResponse, adpods AdPodBids) *openrtb.BidResponse {
	bidResp := &openrtb.BidResponse{
		ID:         resp.ID,
		Cur:        resp.Cur,
		CustomData: resp.CustomData,
	}
	for _, adpod := range adpods {
		if len(adpod.Bids) == 0 {
			continue
		}
		bid := deps.getAdPodBid(adpod)
		if bid != nil {
			found := false
			for _, seat := range bidResp.SeatBid {
				if seat.Seat == adpod.SeatName {
					seat.Bid = append(seat.Bid, *bid)
					found = true
					break
				}
			}
			if found == false {
				bidResp.SeatBid = append(bidResp.SeatBid, openrtb.SeatBid{
					Seat: adpod.SeatName,
					Bid: []openrtb.Bid{
						*bid,
					},
				})
			}
		}
	}

	//NOTE: this should be called at last
	bidResp.Ext = deps.getBidResponseExt(resp)

	return bidResp
}

//getBidResponseExt will return extension object
func (deps *ctvEndpointDeps) getBidResponseExt(resp *openrtb.BidResponse) json.RawMessage {
	type config struct {
		ImpID string            `json:"impid"`
		Imp   []*ImpAdPodConfig `json:"imp,omitempty"`
	}
	type ext struct {
		Response openrtb.BidResponse `json:"bidresponse,omitempty"`
		Config   []config            `json:"config,omitempty"`
	}

	_ext := ext{
		Response: *resp,
		Config:   make([]config, len(deps.impData)),
	}

	for index, imp := range deps.impData {
		_ext.Config[index].ImpID = deps.request.Imp[index].ID
		_ext.Config[index].Imp = imp.Config[:]
	}

	for i := range resp.SeatBid {
		for j := range resp.SeatBid[i].Bid {
			resp.SeatBid[i].Bid[j].AdM = ""
		}
	}

	//Remove extension parameter
	_ext.Response.Ext = nil

	data, _ := json.Marshal(_ext)
	data, _ = jsonparser.Set(resp.Ext, data, "adpod")

	return data[:]
}

//getAdPodBid
func (deps *ctvEndpointDeps) getAdPodBid(adpod *AdPodBid) *Bid {
	bid := openrtb.Bid{}
	//TODO: Write single for loop to get all details
	bidID, err := uuid.NewV4()
	if nil == err {
		bid.ID = bidID.String()
	} else {
		bid.ID = adpod.Bids[0].ID
	}

	bid.ImpID = adpod.OriginalImpID
	bid.AdM = *getAdPodBidCreative(adpod)
	bid.Price = getAdPodBidPrice(adpod)
	bid.ADomain = getAdPodBidAdvertiserDomain(adpod)
	bid.Cat = getAdPodBidCategories(adpod)
	bid.Ext = getAdPodBidExtension(adpod)
	return &bid
}

//getAdPodBidCreative get commulative adpod bid details
func getAdPodBidCreative(adpod *AdPodBid) *string {
	doc := etree.NewDocument()
	vast := doc.CreateElement("VAST")
	vast.CreateAttr("version", "3.0")
	sequenceNumber := 1
	for _, bid := range adpod.Bids {
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
		fmt.Printf("ERROR, %v", err.Error())
		return nil
	}
	return &bidAdM
}

//getAdPodBidPrice get commulative adpod bid details
func getAdPodBidPrice(adpod *AdPodBid) float64 {
	var price float64 = 0
	for _, ad := range adpod.Bids {
		price = price + ad.Price
	}
	return price
}

//getAdPodBidAdvertiserDomain get commulative adpod bid details
func getAdPodBidAdvertiserDomain(adpod *AdPodBid) []string {
	var domains []string
	for _, ad := range adpod.Bids {
		domains = append(domains, ad.ADomain...)
	}
	//send unique domains only
	return domains[:]
}

//getAdPodBidCategories get commulative adpod bid details
func getAdPodBidCategories(adpod *AdPodBid) []string {
	var category []string
	for _, ad := range adpod.Bids {
		if len(ad.Cat) > 0 {
			category = append(category, ad.Cat...)
		}
	}
	//send unique domains only
	return category[:]
}

//getAdPodBidExtension get commulative adpod bid details
func getAdPodBidExtension(adpod *AdPodBid) json.RawMessage {
	type adpodBidExt struct {
		RefBids []string `json:"refbids,omitempty"`
	}
	type extbid struct {
		/* TODO: this can be moved to openrtb_ext.ExtBid */
		openrtb_ext.ExtBid
		AdPod *adpodBidExt `json:"adpod,omitempty"`
	}
	bidExt := &extbid{
		ExtBid: openrtb_ext.ExtBid{
			Prebid: &openrtb_ext.ExtBidPrebid{
				Type:  openrtb_ext.BidTypeVideo,
				Video: &openrtb_ext.ExtBidPrebidVideo{},
			},
		},
		AdPod: &adpodBidExt{
			RefBids: make([]string, len(adpod.Bids)),
		},
	}

	for i, bid := range adpod.Bids {
		bidExt.AdPod.RefBids[i] = bid.ID
		duration, _ := jsonparser.GetInt(bid.Ext, "prebid", "video", "duration")
		bidExt.Prebid.Video.Duration += int(duration)
	}

	rawExt, _ := json.Marshal(bidExt)
	return rawExt
}

/********************* AdPodGenerator Functions *********************/

//IAdPodGenerator interface for generating AdPod from Ads
type IAdPodGenerator interface {
	GetAdPod() *AdPodBid
}

//Comparator check exclusion conditions
type Comparator func(*Bid, *Bid) bool

//AdPodGenerator AdPodGenerator
type AdPodGenerator struct {
	IAdPodGenerator
	bids   *AdPodBid
	config *openrtb_ext.VideoAdPod
	comp   Comparator
}

//NewAdPodGenerator will generate adpod based on configuration
func NewAdPodGenerator(bids *AdPodBid, config *openrtb_ext.VideoAdPod, comp Comparator) *AdPodGenerator {
	return &AdPodGenerator{
		bids:   bids,
		config: config,
		comp:   comp,
	}
}

//GetAdPod will return Adpod based on configurations
func (o *AdPodGenerator) GetAdPod() *AdPodBid {
	result := &AdPodBid{
		OriginalImpID: o.bids.OriginalImpID,
		SeatName:      o.bids.SeatName,
	}
	count := 3
	for i, bid := range o.bids.Bids {
		if i >= count {
			break
		}
		result.Bids = append(result.Bids, bid)
	}
	return result
}

/********************* Helper Functions *********************/

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

func jsonlog(msg string, obj interface{}) {
	//if glog.V(1) {
	data, _ := json.Marshal(obj)
	glog.Infof("[OPENWRAP] %v:%v", msg, string(data))
	//}
}
