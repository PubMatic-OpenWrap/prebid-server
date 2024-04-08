package openrtb2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	uuid "github.com/gofrs/uuid"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"github.com/prebid/openrtb/v19/openrtb2"
	accountService "github.com/prebid/prebid-server/account"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/adpod"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/util"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/exchange"
	"github.com/prebid/prebid-server/gdpr"
	"github.com/prebid/prebid-server/hooks"
	"github.com/prebid/prebid-server/hooks/hookexecution"
	"github.com/prebid/prebid-server/metrics"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/stored_requests"
	"github.com/prebid/prebid-server/usersync"
	"github.com/prebid/prebid-server/util/iputil"
	"github.com/prebid/prebid-server/util/ptrutil"
	"github.com/prebid/prebid-server/util/uuidutil"
)

// CTV Specific Endpoint
type ctvEndpointDeps struct {
	endpointDeps
	request                   *openrtb2.BidRequest
	reqExt                    *openrtb_ext.ExtRequestAdPod
	isAdPodRequest            bool
	impsExtPrebidBidder       map[string]map[string]map[string]interface{}
	impPartnerBlockedTagIDMap map[string]map[string][]string
	ImpPodIdMap               map[string]string
	podCtx                    map[string]adpod.Adpod
	videoImps                 []openrtb2.Imp
	videoSeats                []*openrtb2.SeatBid //stores pure video impression bids

	labels metrics.Labels
}

// NewCTVEndpoint new ctv endpoint object
func NewCTVEndpoint(
	ex exchange.Exchange,
	validator openrtb_ext.BidderParamValidator,
	requestsByID stored_requests.Fetcher,
	videoFetcher stored_requests.Fetcher,
	accounts stored_requests.AccountFetcher,
	//categories stored_requests.CategoryFetcher,
	cfg *config.Configuration,
	met metrics.MetricsEngine,
	pbsAnalytics analytics.PBSAnalyticsModule,
	disabledBidders map[string]string,
	defReqJSON []byte,
	bidderMap map[string]openrtb_ext.BidderName,
	planBuilder hooks.ExecutionPlanBuilder,
	tmaxAdjustments *exchange.TmaxAdjustmentsPreprocessed) (httprouter.Handle, error) {

	if ex == nil || validator == nil || requestsByID == nil || accounts == nil || cfg == nil || met == nil {
		return nil, errors.New("NewCTVEndpoint requires non-nil arguments")
	}
	defRequest := len(defReqJSON) > 0

	ipValidator := iputil.PublicNetworkIPValidator{
		IPv4PrivateNetworks: cfg.RequestValidation.IPv4PrivateNetworksParsed,
		IPv6PrivateNetworks: cfg.RequestValidation.IPv6PrivateNetworksParsed,
	}
	var uuidGenerator uuidutil.UUIDGenerator
	return httprouter.Handle((&ctvEndpointDeps{
		endpointDeps: endpointDeps{
			uuidGenerator,
			ex,
			validator,
			requestsByID,
			videoFetcher,
			accounts,
			cfg,
			met,
			pbsAnalytics,
			disabledBidders,
			defRequest,
			defReqJSON,
			bidderMap,
			nil,
			nil,
			ipValidator,
			nil,
			planBuilder,
			tmaxAdjustments,
		},
	}).CTVAuctionEndpoint), nil
}

func (deps *ctvEndpointDeps) CTVAuctionEndpoint(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	defer util.TimeTrack(time.Now(), "CTVAuctionEndpoint")

	var reqWrapper *openrtb_ext.RequestWrapper
	var request *openrtb2.BidRequest
	var response *openrtb2.BidResponse
	var err error
	var errL []error

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
	deps.labels = metrics.Labels{
		Source:        metrics.DemandUnknown,
		RType:         metrics.ReqTypeVideo,
		PubID:         metrics.PublisherUnknown,
		CookieFlag:    metrics.CookieFlagUnknown,
		RequestStatus: metrics.RequestStatusOK,
	}
	defer func() {
		deps.metricsEngine.RecordRequest(deps.labels)
		recordRejectedBids(deps.labels.PubID, ao.SeatNonBid, deps.metricsEngine)
		deps.metricsEngine.RecordRequestTime(deps.labels, time.Since(start))
		deps.analytics.LogAuctionObject(&ao)
	}()

	hookExecutor := hookexecution.NewHookExecutor(deps.hookExecutionPlanBuilder, hookexecution.EndpointCtv, deps.metricsEngine)

	//Parse ORTB Request and do Standard Validation
	reqWrapper, _, _, _, _, _, errL = deps.parseRequest(r, &deps.labels, hookExecutor)
	if errortypes.ContainsFatalError(errL) && writeError(errL, w, &deps.labels) {
		return
	}
	if reqWrapper.RebuildRequestExt() != nil {
		return
	}
	request = reqWrapper.BidRequest

	//init
	if errs := deps.init(request); len(errs) > 0 {
		errL = append(errL, errs...)
		writeError(errL, w, &deps.labels)
		return
	}

	// set adpod context
	if errs := deps.prepareAdpodCtx(request); len(errs) > 0 {
		errL = append(errL, errs...)
		writeError(errL, w, &deps.labels)
		return
	}

	//Set Default Values
	deps.setDefaultValues()

	//Validate CTV BidRequest
	if err := deps.ValidateAdpodCtx(); err != nil {
		errL = append(errL, err...)
		writeError(errL, w, &deps.labels)
		return
	}

	if deps.isAdPodRequest {
		request = deps.createBidRequest(request)
	}

	//Parsing Cookies and Set Stats
	usersyncs := usersync.ReadCookie(r, usersync.Base64Decoder{}, &deps.cfg.HostCookie)
	usersync.SyncHostCookie(r, usersyncs, &deps.cfg.HostCookie)

	if request.App != nil {
		deps.labels.Source = metrics.DemandApp
		deps.labels.RType = metrics.ReqTypeVideo
		deps.labels.PubID = getAccountID(request.App.Publisher)
	} else { //request.Site != nil
		deps.labels.Source = metrics.DemandWeb
		if !usersyncs.HasAnyLiveSyncs() {
			deps.labels.CookieFlag = metrics.CookieFlagNo
		} else {
			deps.labels.CookieFlag = metrics.CookieFlagYes
		}
		deps.labels.PubID = getAccountID(request.Site.Publisher)
	}
	ctx := r.Context()

	// Look up account now that we have resolved the pubID value
	account, acctIDErrs := accountService.GetAccount(ctx, deps.cfg, deps.accounts, deps.labels.PubID, deps.metricsEngine)
	if len(acctIDErrs) > 0 {
		errL = append(errL, acctIDErrs...)
		writeError(errL, w, &deps.labels)
		return
	}

	//Setting Timeout for Request
	timeout := deps.cfg.AuctionTimeouts.LimitAuctionTimeout(time.Duration(request.TMax) * time.Millisecond)
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, start.Add(timeout))
		defer cancel()
	}

	tcf2Config := gdpr.NewTCF2Config(deps.cfg.GDPR.TCF2, account.GDPR)
	reqWrapper.BidRequest = request
	auctionRequest := exchange.AuctionRequest{
		BidRequestWrapper: &openrtb_ext.RequestWrapper{BidRequest: request},
		Account:           *account,
		UserSyncs:         usersyncs,
		RequestType:       deps.labels.RType,
		StartTime:         start,
		LegacyLabels:      deps.labels,
		PubID:             deps.labels.PubID,
		HookExecutor:      hookExecutor,
		TCF2Config:        tcf2Config,
		TmaxAdjustments:   deps.tmaxAdjustments,
	}

	auctionResponse, err := deps.holdAuction(ctx, auctionRequest)
	defer func() {
		if !auctionRequest.BidderResponseStartTime.IsZero() {
			deps.metricsEngine.RecordOverheadTime(metrics.MakeAuctionResponse, time.Since(auctionRequest.BidderResponseStartTime))
		}
	}()

	ao.RequestWrapper = auctionRequest.BidRequestWrapper
	if err != nil || auctionResponse == nil || auctionResponse.BidResponse == nil {
		deps.labels.RequestStatus = metrics.RequestStatusErr
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Critical error while running the auction: %v", err)
		glog.Errorf("/openrtb2/video Critical error: %v", err)
		ao.Status = http.StatusInternalServerError
		ao.Errors = append(ao.Errors, err)
		return
	}
	ao.SeatNonBid = auctionResponse.GetSeatNonBid()
	err = setSeatNonBidRaw(ao.RequestWrapper, auctionResponse)
	if err != nil {
		glog.Errorf("Error setting seat non-bid: %v", err)
	}
	response = auctionResponse.BidResponse

	if deps.isAdPodRequest {
		//Validate Bid Response
		if err := deps.validateBidResponse(request, response); err != nil {
			errL = append(errL, err)
			writeError(errL, w, &deps.labels)
			return
		}

		//Create Impression Bids
		deps.collectBids(response)

		//Do AdPod Exclusions
		deps.doAdpodAuctionAndExclusion()

		//Create Bid Response
		adPodBidResponse := deps.createAdPodBidResponse(response)
		adPodBidResponse.Ext = deps.getBidResponseExt(response)
		response = adPodBidResponse
	}
	ao.Response = response

	// Response Generation
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	// Fixes #328
	w.Header().Set("Content-Type", "application/json")

	// If an error happens when encoding the response, there isn't much we can do.
	// If we've sent _any_ bytes, then Go would have sent the 200 status code first.
	// That status code can't be un-sent... so the best we can do is log the error.
	if err := enc.Encode(response); err != nil {
		deps.labels.RequestStatus = metrics.RequestStatusNetworkErr
		ao.Errors = append(ao.Errors, fmt.Errorf("/openrtb2/video Failed to send response: %v", err))
	}
}

func (deps *ctvEndpointDeps) holdAuction(ctx context.Context, auctionRequest exchange.AuctionRequest) (*exchange.AuctionResponse, error) {
	defer util.TimeTrack(time.Now(), fmt.Sprintf("Tid:%v CTVHoldAuction", deps.request.ID))

	//Hold OpenRTB Standard Auction
	if len(deps.request.Imp) == 0 {
		//Dummy Response Object
		return &exchange.AuctionResponse{BidResponse: &openrtb2.BidResponse{ID: deps.request.ID}}, nil
	}

	return deps.ex.HoldAuction(ctx, &auctionRequest, nil)
}

/********************* BidRequest Processing *********************/

func (deps *ctvEndpointDeps) init(req *openrtb2.BidRequest) (errs []error) {
	deps.request = req
	deps.ImpPodIdMap = make(map[string]string)

	// Read adpod Object Request extension
	deps.readRequestExtension()

	//validate request extension
	errs = deps.reqExt.Validate()
	return
}

/*
PrepareAdpodCtx will check for adpod param, and create adpod context if they are
available in the request. It will check for both ORTB 2.6 and legacy adpod parmaters.
*/
func (deps *ctvEndpointDeps) prepareAdpodCtx(request *openrtb2.BidRequest) (errs []error) {
	deps.podCtx = make(map[string]adpod.Adpod)

	for _, imp := range request.Imp {
		if imp.Video != nil {
			// check for adpod in the extension
			adpodConfigInExt, err := deps.readVideoAdPodExt(imp)
			if err != nil {
				errs = append(errs, err)
			}

			if adpodConfigInExt != nil {
				// Dynamic Adpod
				deps.createDynamicAdpodCtx(imp, adpodConfigInExt)
			} else if len(imp.Video.PodID) > 0 && imp.Video.PodDur > 0 {
				// Dynamic adpod
				deps.createDynamicAdpodCtx(imp, adpodConfigInExt)
			} else if len(imp.Video.PodID) > 0 {
				// structured adpod
				deps.createStructuredAdpodCtx(imp)
			} else {
				// Pure video impressions
				deps.videoImps = append(deps.videoImps, imp)
			}
		}
	}
	return
}

func (deps *ctvEndpointDeps) createDynamicAdpodCtx(imp openrtb2.Imp, adpodConfigExt *openrtb_ext.ExtVideoAdPod) {
	podId := imp.Video.PodID
	if len(podId) == 0 {
		podId = imp.ID
	}

	deps.ImpPodIdMap[imp.ID] = podId

	//set Pod Duration
	minPodDuration := imp.Video.MinDuration
	maxPodDuration := imp.Video.MaxDuration

	if adpodConfigExt == nil {
		adpodConfigExt = &openrtb_ext.ExtVideoAdPod{
			Offset: ptrutil.ToPtr(0),
			AdPod: &openrtb_ext.VideoAdPod{
				MinAds:                      ptrutil.ToPtr(1),
				MaxAds:                      ptrutil.ToPtr(int(imp.Video.MaxSeq)),
				MinDuration:                 ptrutil.ToPtr(int(imp.Video.MinDuration)),
				MaxDuration:                 ptrutil.ToPtr(int(imp.Video.MaxDuration)),
				AdvertiserExclusionPercent:  ptrutil.ToPtr(100),
				IABCategoryExclusionPercent: ptrutil.ToPtr(100),
			},
		}
		maxPodDuration = imp.Video.PodDur
	}

	deps.podCtx[podId] = &adpod.DynamicAdpod{
		AdpodCtx: adpod.AdpodCtx{
			PubId:         deps.labels.PubID,
			Type:          adpod.PodTypeDynamic,
			ReqExt:        deps.reqExt,
			MetricsEngine: deps.metricsEngine,
		},
		Imp:            imp,
		VideoExt:       adpodConfigExt,
		MinPodDuration: minPodDuration,
		MaxPodDuration: maxPodDuration,
	}
}

func (deps *ctvEndpointDeps) createStructuredAdpodCtx(imp openrtb2.Imp) {
	deps.ImpPodIdMap[imp.ID] = imp.Video.PodID

	podContext, ok := deps.podCtx[imp.Video.PodID]
	if !ok {
		podContext = &adpod.StructuredAdpod{
			AdpodCtx: adpod.AdpodCtx{
				PubId:         deps.labels.PubID,
				Type:          adpod.PodTypeStructured,
				ReqExt:        deps.reqExt,
				MetricsEngine: deps.metricsEngine,
			},
			ImpBidMap:  make(map[string][]*types.Bid),
			WinningBid: make(map[string]types.Bid),
		}
	}

	podContext.AddImpressions(imp)
	deps.podCtx[imp.Video.PodID] = podContext
}

func (deps *ctvEndpointDeps) ValidateAdpodCtx() []error {
	var errs []error
	for _, eachpod := range deps.podCtx {
		err := eachpod.Validate()
		if err != nil {
			errs = append(errs, err...)
		}
	}

	return errs
}

func (deps *ctvEndpointDeps) readVideoAdPodExt(imp openrtb2.Imp) (*openrtb_ext.ExtVideoAdPod, error) {
	var adpodConfigExt openrtb_ext.ExtVideoAdPod

	if imp.Video != nil && len(imp.Video.Ext) > 0 {
		err := json.Unmarshal(imp.Video.Ext, &adpodConfigExt)
		if err != nil {
			return nil, err
		}
	}

	if adpodConfigExt.AdPod == nil && deps.reqExt == nil {
		return nil, nil
	}

	if deps.reqExt != nil {
		if adpodConfigExt.AdPod == nil {
			adpodConfigExt.AdPod = &openrtb_ext.VideoAdPod{}
		}
		//TODO: Here partical values might be filled check if is it okay to fill partial values?
		adpodConfigExt.AdPod.Merge(&deps.reqExt.VideoAdPod)
	}

	//TODO: Check if we require to set these default values
	//Set Default Values
	adpodConfigExt.SetDefaultValue()
	adpodConfigExt.AdPod.SetDefaultAdDurations(imp.Video.MinDuration, imp.Video.MaxDuration)

	return &adpodConfigExt, nil
}

func (deps *ctvEndpointDeps) readRequestExtension() (err []error) {
	if len(deps.request.Ext) > 0 {
		//TODO: use jsonparser library for get adpod and remove that key
		extAdPod, jsonType, _, errL := jsonparser.Get(deps.request.Ext, constant.CTVAdpod)
		if errL != nil {
			//parsing error
			if jsonparser.NotExist != jsonType {
				//assuming key not present
				err = append(err, errL)
				return
			}
		} else {
			deps.reqExt = &openrtb_ext.ExtRequestAdPod{}
			if errL := json.Unmarshal(extAdPod, deps.reqExt); nil != errL {
				err = append(err, errL)
				return
			}

			deps.reqExt.SetDefaultValue()
		}
	}
	return
}

func (deps *ctvEndpointDeps) setIsAdPodRequest() {
	if len(deps.podCtx) > 0 {
		deps.isAdPodRequest = true
	}
}

// setDefaultValues will set adpod and other default values
func (deps *ctvEndpointDeps) setDefaultValues() {

	//set request is adpod request or normal request
	deps.setIsAdPodRequest()

	if deps.isAdPodRequest {
		deps.readImpExtensionsAndTags()
	}
}

// readImpExtensionsAndTags will read the impression extensions
func (deps *ctvEndpointDeps) readImpExtensionsAndTags() (errs []error) {
	deps.impsExtPrebidBidder = make(map[string]map[string]map[string]interface{})
	deps.impPartnerBlockedTagIDMap = make(map[string]map[string][]string) //Initially this will have all tags, eligible tags will be filtered in filterImpsVastTagsByDuration

	for _, imp := range deps.request.Imp {
		bidderExtBytes, _, _, err := jsonparser.Get(imp.Ext, "prebid", "bidder")
		if err != nil {
			errs = append(errs, err)
			continue
		}
		impsExtPrebidBidder := make(map[string]map[string]interface{})

		err = json.Unmarshal(bidderExtBytes, &impsExtPrebidBidder)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		deps.impPartnerBlockedTagIDMap[imp.ID] = make(map[string][]string)

		for partnerName, partnerExt := range impsExtPrebidBidder {
			impVastTags, ok := partnerExt["tags"].([]interface{})
			if !ok {
				continue
			}

			for _, tag := range impVastTags {
				vastTag, ok := tag.(map[string]interface{})
				if !ok {
					continue
				}

				deps.impPartnerBlockedTagIDMap[imp.ID][partnerName] = append(deps.impPartnerBlockedTagIDMap[imp.ID][partnerName], vastTag["tagid"].(string))
			}
		}

		deps.impsExtPrebidBidder[imp.ID] = impsExtPrebidBidder
	}

	return errs
}

/********************* Creating CTV BidRequest *********************/

// createBidRequest will return new bid request with all things copy from bid request except impression objects
func (deps *ctvEndpointDeps) createBidRequest(req *openrtb2.BidRequest) *openrtb2.BidRequest {
	ctvRequest := *req

	for _, adpodCtx := range deps.podCtx {
		adpodCtx.GenerateImpressions()
	}

	//createImpressions
	var imps []openrtb2.Imp
	for _, adpodCtx := range deps.podCtx {
		imps = append(imps, adpodCtx.GetImpressions()...)
	}
	if len(deps.videoImps) > 0 {
		imps = append(imps, deps.videoImps...)
	}
	ctvRequest.Imp = imps

	deps.filterImpsVastTagsByDuration(&ctvRequest)

	//TODO: remove adpod extension if not required to send further
	return &ctvRequest
}

// filterImpsVastTagsByDuration checks if a Vast tag should be called for a generated impression based on the duration of tag and impression
func (deps *ctvEndpointDeps) filterImpsVastTagsByDuration(bidReq *openrtb2.BidRequest) {
	for impCount, imp := range bidReq.Imp {
		index := strings.LastIndex(imp.ID, "_")
		if index == -1 {
			continue
		}

		originalImpID := imp.ID[:index]

		impExtBidder := deps.impsExtPrebidBidder[originalImpID]
		impExtBidderCopy := make(map[string]map[string]interface{})
		for partnerName, partnerExt := range impExtBidder {
			impExtBidderCopy[partnerName] = partnerExt
		}

		for partnerName, partnerExt := range impExtBidderCopy {
			if partnerExt["tags"] != nil {
				impVastTags, ok := partnerExt["tags"].([]interface{})
				if !ok {
					continue
				}

				var compatibleVasts []interface{}
				for _, tag := range impVastTags {
					vastTag, ok := tag.(map[string]interface{})
					if !ok {
						continue
					}

					tagDuration := int(vastTag["dur"].(float64))
					if int(imp.Video.MinDuration) <= tagDuration && tagDuration <= int(imp.Video.MaxDuration) {
						compatibleVasts = append(compatibleVasts, tag)

						deps.impPartnerBlockedTagIDMap[originalImpID][partnerName] = remove(deps.impPartnerBlockedTagIDMap[originalImpID][partnerName], vastTag["tagid"].(string))
						if len(deps.impPartnerBlockedTagIDMap[originalImpID][partnerName]) == 0 {
							delete(deps.impPartnerBlockedTagIDMap[originalImpID], partnerName)
						}
					}
				}

				if len(compatibleVasts) < 1 {
					delete(impExtBidderCopy, partnerName)
				} else {
					impExtBidderCopy[partnerName] = map[string]interface{}{
						"tags": compatibleVasts,
					}
				}
			}
		}

		bidderExtBytes, err := json.Marshal(impExtBidderCopy)
		if err != nil {
			continue
		}

		// if imp.ext exists then set prebid.bidder inside it
		impExt, err := jsonparser.Set(imp.Ext, bidderExtBytes, "prebid", "bidder")
		if err != nil {
			continue
		}

		imp.Ext = impExt
		bidReq.Imp[impCount] = imp
	}
}

func remove(slice []string, item string) []string {
	index := -1
	for i := range slice {
		if slice[i] == item {
			index = i
			break
		}
	}

	if index == -1 {
		return slice
	}

	return append(slice[:index], slice[index+1:]...)
}

/********************* Prebid BidResponse Processing *********************/

// validateBidResponse
func (deps *ctvEndpointDeps) validateBidResponse(req *openrtb2.BidRequest, resp *openrtb2.BidResponse) error {
	//remove bids withoug cat and adomain

	return nil
}

func (deps *ctvEndpointDeps) collectBids(response *openrtb2.BidResponse) {
	var vseat *openrtb2.SeatBid

	for _, seat := range response.SeatBid {
		vseat = nil
		for _, bid := range seat.Bid {
			if len(bid.ID) == 0 {
				bidID, err := uuid.NewV4()
				if err != nil {
					continue
				}
				bid.ID = bidID.String()
			}

			if bid.Price == 0 {
				continue
			}

			originalImpID, _ := util.DecodeImpressionID(bid.ImpID)
			podId, ok := deps.ImpPodIdMap[originalImpID]
			if !ok {
				if vseat == nil {
					vseat = &openrtb2.SeatBid{
						Seat:  seat.Seat,
						Group: seat.Group,
						Ext:   seat.Ext,
					}
					deps.videoSeats = append(deps.videoSeats, vseat)
				}
				vseat.Bid = append(vseat.Bid, bid)
				continue
			}

			adpodCtx, ok := deps.podCtx[podId]
			if !ok {
				continue
			}

			adpodCtx.CollectBid(bid, seat.Seat)
		}
	}
}

func (deps *ctvEndpointDeps) doAdpodAuctionAndExclusion() {
	for _, adpodCtx := range deps.podCtx {
		adpodCtx.PerformAuctionAndExclusion()
	}
}

/********************* Creating CTV BidResponse *********************/

// createAdPodBidResponse
func (deps *ctvEndpointDeps) createAdPodBidResponse(resp *openrtb2.BidResponse) *openrtb2.BidResponse {
	var seatbids []openrtb2.SeatBid

	//append pure video request seats
	for _, seat := range deps.videoSeats {
		seatbids = append(seatbids, *seat)
	}

	for _, adpod := range deps.podCtx {
		seatbids = append(seatbids, adpod.GetAdpodSeatBids()...)
	}

	bidResp := &openrtb2.BidResponse{
		ID:         resp.ID,
		Cur:        resp.Cur,
		CustomData: resp.CustomData,
		SeatBid:    seatbids,
	}
	return bidResp
}

// getBidResponseExt prepare and return the bidresponse extension
func (deps *ctvEndpointDeps) getBidResponseExt(resp *openrtb2.BidResponse) (data json.RawMessage) {
	var err error
	adpodExt := types.BidResponseAdPodExt{
		Response: *resp,
		Config:   make(map[string]*types.ImpData, len(deps.podCtx)),
	}

	for podId, adpodCtx := range deps.podCtx {
		adpodExt.Config[podId] = adpodCtx.GetAdpodExtension(deps.impPartnerBlockedTagIDMap)
	}

	//Remove extension parameter
	adpodExt.Response.Ext = nil

	if resp.Ext == nil {
		bidResponseExt := &types.ExtCTVBidResponse{
			AdPod: &adpodExt,
		}

		data, err = json.Marshal(bidResponseExt)
		if err != nil {
			glog.Errorf("JSON Marshal Error: %v", err.Error())
			return nil
		}
	} else {
		data, err = json.Marshal(adpodExt)
		if err != nil {
			glog.Errorf("JSON Marshal Error: %v", err.Error())
			return nil
		}

		data, err = jsonparser.Set(resp.Ext, data, constant.CTVAdpod)
		if err != nil {
			glog.Errorf("JSONParser Set Error: %v", err.Error())
			return nil
		}
	}
	return data
}
