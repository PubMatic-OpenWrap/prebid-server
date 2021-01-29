package exchange

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters/vastbidder"
	"github.com/PubMatic-OpenWrap/prebid-server/stored_requests"
	uuid "github.com/gofrs/uuid"
	"golang.org/x/net/publicsuffix"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/currencies"
	"github.com/PubMatic-OpenWrap/prebid-server/errortypes"
	"github.com/PubMatic-OpenWrap/prebid-server/gdpr"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/PubMatic-OpenWrap/prebid-server/pbsmetrics"
	"github.com/PubMatic-OpenWrap/prebid-server/prebid_cache_client"
	"github.com/golang/glog"
)

type ContextKey string

const DebugContextKey = ContextKey("debugInfo")

type extCacheInstructions struct {
	cacheBids, cacheVAST, returnCreative bool
}

// Exchange runs Auctions. Implementations must be threadsafe, and will be shared across many goroutines.
type Exchange interface {
	// HoldAuction executes an OpenRTB v2.5 Auction.
	HoldAuction(ctx context.Context, r AuctionRequest, debugLog *DebugLog) (*openrtb.BidResponse, error)
}

// IdFetcher can find the user's ID for a specific Bidder.
type IdFetcher interface {
	// GetId returns the ID for the bidder. The boolean will be true if the ID exists, and false otherwise.
	GetId(bidder openrtb_ext.BidderName) (string, bool)
	LiveSyncCount() int
}

type exchange struct {
	adapterMap          map[openrtb_ext.BidderName]adaptedBidder
	me                  pbsmetrics.MetricsEngine
	cache               prebid_cache_client.Client
	cacheTime           time.Duration
	gDPR                gdpr.Permissions
	currencyConverter   *currencies.RateConverter
	UsersyncIfAmbiguous bool
	privacyConfig       config.Privacy
	categoriesFetcher   stored_requests.CategoryFetcher
}

// Container to pass out response ext data from the GetAllBids goroutines back into the main thread
type seatResponseExtra struct {
	ResponseTimeMillis int
	Errors             []openrtb_ext.ExtBidderError
	// httpCalls is the list of debugging info. It should only be populated if the request.test == 1.
	// This will become response.ext.debug.httpcalls.{bidder} on the final Response.
	HttpCalls []*openrtb_ext.ExtHttpCall
}

type bidResponseWrapper struct {
	adapterBids  *pbsOrtbSeatBid
	adapterExtra *seatResponseExtra
	bidder       openrtb_ext.BidderName
}

func NewExchange(client *http.Client, cache prebid_cache_client.Client, cfg *config.Configuration, metricsEngine pbsmetrics.MetricsEngine, infos adapters.BidderInfos, gDPR gdpr.Permissions, currencyConverter *currencies.RateConverter, categoriesFetcher stored_requests.CategoryFetcher) Exchange {
	e := new(exchange)

	e.adapterMap = newAdapterMap(client, cfg, infos, metricsEngine)
	e.cache = cache
	e.cacheTime = time.Duration(cfg.CacheURL.ExpectedTimeMillis) * time.Millisecond
	e.me = metricsEngine
	e.gDPR = gDPR
	e.currencyConverter = currencyConverter
	e.UsersyncIfAmbiguous = cfg.GDPR.UsersyncIfAmbiguous
	e.privacyConfig = config.Privacy{
		CCPA: cfg.CCPA,
		GDPR: cfg.GDPR,
		LMT:  cfg.LMT,
	}
	e.categoriesFetcher = categoriesFetcher
	return e
}

type AuctionRequest struct {
	BidRequest  *openrtb.BidRequest
	Account     config.Account
	UserSyncs   IdFetcher
	RequestType pbsmetrics.RequestType

	// LegacyLabels is included here for temporary compatability with cleanOpenRTBRequests
	// in HoldAuction until we get to factoring it away. Do not use for anything new.
	LegacyLabels pbsmetrics.Labels
}

func (e *exchange) HoldAuction(ctx context.Context, r AuctionRequest, debugLog *DebugLog) (*openrtb.BidResponse, error) {
	var err error
	requestExt, err := extractBidRequestExt(r.BidRequest)
	if err != nil {
		return nil, err
	}

	cacheInstructions := getExtCacheInstructions(requestExt)
	targData := getExtTargetData(requestExt, &cacheInstructions)
	if targData != nil {
		_, targData.cacheHost, targData.cachePath = e.cache.GetExtCacheData()
	}

	debugInfo := getDebugInfo(r.BidRequest, requestExt)
	if debugInfo {
		ctx = e.makeDebugContext(ctx, debugInfo)
	}

	bidAdjustmentFactors := getExtBidAdjustmentFactors(requestExt)

	recordImpMetrics(r.BidRequest, e.me)

	// Make our best guess if GDPR applies
	usersyncIfAmbiguous := e.parseUsersyncIfAmbiguous(r.BidRequest)

	// Slice of BidRequests, each a copy of the original cleaned to only contain bidder data for the named bidder
	blabels := make(map[openrtb_ext.BidderName]*pbsmetrics.AdapterLabels)
	cleanRequests, aliases, privacyLabels, errs := cleanOpenRTBRequests(ctx, r.BidRequest, requestExt, r.UserSyncs, blabels, r.LegacyLabels, e.gDPR, usersyncIfAmbiguous, e.privacyConfig, &r.Account)

	e.me.RecordRequestPrivacy(privacyLabels)

	// List of bidders we have requests for.
	liveAdapters := listBiddersWithRequests(cleanRequests)

	// If we need to cache bids, then it will take some time to call prebid cache.
	// We should reduce the amount of time the bidders have, to compensate.
	auctionCtx, cancel := e.makeAuctionContext(ctx, cacheInstructions.cacheBids)
	defer cancel()

	// Get currency rates conversions for the auction
	conversions := e.currencyConverter.Rates()

	adapterBids, adapterExtra, anyBidsReturned := e.getAllBids(auctionCtx, cleanRequests, aliases, bidAdjustmentFactors, blabels, conversions)

	var auc *auction
	var cacheErrs []error
	if anyBidsReturned {

		adapterBids, rejections := applyAdvertiserBlocking(r.BidRequest, adapterBids, e.adapterMap)
		// add advertiser blocking specific errors
		for _, message := range rejections {
			errs = append(errs, errors.New(message))
		}

		var bidCategory map[string]string
		//If includebrandcategory is present in ext then CE feature is on.
		if requestExt.Prebid.Targeting != nil && requestExt.Prebid.Targeting.IncludeBrandCategory != nil {
			var rejections []string
			bidCategory, adapterBids, rejections, err = applyCategoryMapping(ctx, r.BidRequest, requestExt, adapterBids, e.categoriesFetcher, targData)
			if err != nil {
				return nil, fmt.Errorf("Error in category mapping : %s", err.Error())
			}
			for _, message := range rejections {
				errs = append(errs, errors.New(message))
			}
		}

		if targData != nil {
			// A non-nil auction is only needed if targeting is active. (It is used below this block to extract cache keys)
			auc = newAuction(adapterBids, len(r.BidRequest.Imp), targData.preferDeals)
			auc.setRoundedPrices(targData.priceGranularity)

			if requestExt.Prebid.SupportDeals {
				dealErrs := applyDealSupport(r.BidRequest, auc, bidCategory)
				errs = append(errs, dealErrs...)
			}

			cacheErrs := auc.doCache(ctx, e.cache, targData, r.BidRequest, 60, &r.Account.CacheTTL, bidCategory, debugLog)
			if len(cacheErrs) > 0 {
				errs = append(errs, cacheErrs...)
			}
			targData.setTargeting(auc, r.BidRequest.App != nil, bidCategory)

		}
	}

	bidResponseExt := e.makeExtBidResponse(adapterBids, adapterExtra, r.BidRequest, debugInfo, errs)

	// Ensure caching errors are added in case auc.doCache was called and errors were returned
	if len(cacheErrs) > 0 {
		bidderCacheErrs := errsToBidderErrors(cacheErrs)
		bidResponseExt.Errors[openrtb_ext.PrebidExtKey] = append(bidResponseExt.Errors[openrtb_ext.PrebidExtKey], bidderCacheErrs...)
	}

	if debugLog != nil && debugLog.Enabled {
		if bidRespExtBytes, err := json.Marshal(bidResponseExt); err == nil {
			debugLog.Data.Response = string(bidRespExtBytes)
		} else {
			debugLog.Data.Response = "Unable to marshal response ext for debugging"
			errs = append(errs, err)
		}
		if !anyBidsReturned {
			if rawUUID, err := uuid.NewV4(); err == nil {
				debugLog.CacheKey = rawUUID.String()
			} else {
				errs = append(errs, err)
			}
		}
	}

	// Build the response
	return e.buildBidResponse(ctx, liveAdapters, adapterBids, r.BidRequest, adapterExtra, auc, bidResponseExt, cacheInstructions.returnCreative, errs)
}

func (e *exchange) parseUsersyncIfAmbiguous(bidRequest *openrtb.BidRequest) bool {
	usersyncIfAmbiguous := e.UsersyncIfAmbiguous
	var geo *openrtb.Geo = nil

	if bidRequest.User != nil && bidRequest.User.Geo != nil {
		geo = bidRequest.User.Geo
	} else if bidRequest.Device != nil && bidRequest.Device.Geo != nil {
		geo = bidRequest.Device.Geo
	}
	if geo != nil {
		// If we have a country set, and it is on the list, we assume GDPR applies if not set on the request.
		// Otherwise we assume it does not apply as long as it appears "valid" (is 3 characters long).
		if _, found := e.privacyConfig.GDPR.EEACountriesMap[strings.ToUpper(geo.Country)]; found {
			usersyncIfAmbiguous = false
		} else if len(geo.Country) == 3 {
			// The country field is formatted properly as a three character country code
			usersyncIfAmbiguous = true
		}
	}

	return usersyncIfAmbiguous
}

func recordImpMetrics(bidRequest *openrtb.BidRequest, metricsEngine pbsmetrics.MetricsEngine) {
	for _, impInRequest := range bidRequest.Imp {
		var impLabels pbsmetrics.ImpLabels = pbsmetrics.ImpLabels{
			BannerImps: impInRequest.Banner != nil,
			VideoImps:  impInRequest.Video != nil,
			AudioImps:  impInRequest.Audio != nil,
			NativeImps: impInRequest.Native != nil,
		}
		metricsEngine.RecordImps(impLabels)
	}
}

// applyDealSupport updates targeting keys with deal prefixes if minimum deal tier exceeded
func applyDealSupport(bidRequest *openrtb.BidRequest, auc *auction, bidCategory map[string]string) []error {
	errs := []error{}
	impDealMap := getDealTiers(bidRequest)

	for impID, topBidsPerImp := range auc.winningBidsByBidder {
		impDeal := impDealMap[impID]
		for bidder, topBidPerBidder := range topBidsPerImp {
			if topBidPerBidder.dealPriority > 0 {
				if validateDealTier(impDeal[bidder]) {
					updateHbPbCatDur(topBidPerBidder, impDeal[bidder], bidCategory)
				} else {
					errs = append(errs, fmt.Errorf("dealTier configuration invalid for bidder '%s', imp ID '%s'", string(bidder), impID))
				}
			}
		}
	}

	return errs
}

// getDealTiers creates map of impression to bidder deal tier configuration
func getDealTiers(bidRequest *openrtb.BidRequest) map[string]openrtb_ext.DealTierBidderMap {
	impDealMap := make(map[string]openrtb_ext.DealTierBidderMap)

	for _, imp := range bidRequest.Imp {
		dealTierBidderMap, err := openrtb_ext.ReadDealTiersFromImp(imp)
		if err != nil {
			continue
		}
		impDealMap[imp.ID] = dealTierBidderMap
	}

	return impDealMap
}

func validateDealTier(dealTier openrtb_ext.DealTier) bool {
	return len(dealTier.Prefix) > 0 && dealTier.MinDealTier > 0
}

func updateHbPbCatDur(bid *pbsOrtbBid, dealTier openrtb_ext.DealTier, bidCategory map[string]string) {
	if bid.dealPriority >= dealTier.MinDealTier {
		prefixTier := fmt.Sprintf("%s%d_", dealTier.Prefix, bid.dealPriority)
		bid.dealTierSatisfied = true

		if oldCatDur, ok := bidCategory[bid.bid.ID]; ok {
			oldCatDurSplit := strings.SplitAfterN(oldCatDur, "_", 2)
			oldCatDurSplit[0] = prefixTier

			newCatDur := strings.Join(oldCatDurSplit, "")
			bidCategory[bid.bid.ID] = newCatDur
		}
	}
}

func (e *exchange) makeDebugContext(ctx context.Context, debugInfo bool) (debugCtx context.Context) {
	debugCtx = context.WithValue(ctx, DebugContextKey, debugInfo)
	return
}

func (e *exchange) makeAuctionContext(ctx context.Context, needsCache bool) (auctionCtx context.Context, cancel context.CancelFunc) {
	auctionCtx = ctx
	cancel = func() {}
	if needsCache {
		if deadline, ok := ctx.Deadline(); ok {
			auctionCtx, cancel = context.WithDeadline(ctx, deadline.Add(-e.cacheTime))
		}
	}
	return
}

// This piece sends all the requests to the bidder adapters and gathers the results.
func (e *exchange) getAllBids(ctx context.Context, cleanRequests map[openrtb_ext.BidderName]*openrtb.BidRequest, aliases map[string]string, bidAdjustments map[string]float64, blabels map[openrtb_ext.BidderName]*pbsmetrics.AdapterLabels, conversions currencies.Conversions) (map[openrtb_ext.BidderName]*pbsOrtbSeatBid, map[openrtb_ext.BidderName]*seatResponseExtra, bool) {
	// Set up pointers to the bid results
	adapterBids := make(map[openrtb_ext.BidderName]*pbsOrtbSeatBid, len(cleanRequests))
	adapterExtra := make(map[openrtb_ext.BidderName]*seatResponseExtra, len(cleanRequests))
	chBids := make(chan *bidResponseWrapper, len(cleanRequests))
	bidsFound := false
	bidIDsCollision := false

	for bidderName, req := range cleanRequests {
		// Here we actually call the adapters and collect the bids.
		coreBidder := resolveBidder(string(bidderName), aliases)
		bidderRunner := e.recoverSafely(cleanRequests, func(aName openrtb_ext.BidderName, coreBidder openrtb_ext.BidderName, request *openrtb.BidRequest, bidlabels *pbsmetrics.AdapterLabels, conversions currencies.Conversions) {
			// Passing in aName so a doesn't change out from under the go routine
			if bidlabels.Adapter == "" {
				glog.Errorf("Exchange: bidlables for %s (%s) missing adapter string", aName, coreBidder)
				bidlabels.Adapter = coreBidder
			}
			brw := new(bidResponseWrapper)
			brw.bidder = aName
			// Defer basic metrics to insure we capture them after all the values have been set
			defer func() {
				e.me.RecordAdapterRequest(*bidlabels)
			}()
			start := time.Now()

			adjustmentFactor := 1.0
			if givenAdjustment, ok := bidAdjustments[string(aName)]; ok {
				adjustmentFactor = givenAdjustment
			}
			var reqInfo adapters.ExtraRequestInfo
			reqInfo.PbsEntryPoint = bidlabels.RType
			bids, err := e.adapterMap[coreBidder].requestBid(ctx, request, aName, adjustmentFactor, conversions, &reqInfo)

			// Add in time reporting
			elapsed := time.Since(start)
			brw.adapterBids = bids
			// Structure to record extra tracking data generated during bidding
			ae := new(seatResponseExtra)
			ae.ResponseTimeMillis = int(elapsed / time.Millisecond)
			if bids != nil {
				ae.HttpCalls = bids.httpCalls
			}

			// Timing statistics
			e.me.RecordAdapterTime(*bidlabels, time.Since(start))
			serr := errsToBidderErrors(err)
			bidlabels.AdapterBids = bidsToMetric(brw.adapterBids)
			bidlabels.AdapterErrors = errorsToMetric(err)
			// Append any bid validation errors to the error list
			ae.Errors = serr
			brw.adapterExtra = ae
			if bids != nil {
				for _, bid := range bids.bids {
					var cpm = float64(bid.bid.Price * 1000)
					e.me.RecordAdapterPrice(*bidlabels, cpm)
					e.me.RecordAdapterBidReceived(*bidlabels, bid.bidType, bid.bid.AdM != "")
					if bid.bidType == openrtb_ext.BidTypeVideo && bid.bidVideo != nil && bid.bidVideo.Duration > 0 {
						e.me.RecordAdapterVideoBidDuration(*bidlabels, bid.bidVideo.Duration)
					}
				}
			}
			chBids <- brw
		}, chBids)
		go bidderRunner(bidderName, coreBidder, req, blabels[coreBidder], conversions)
	}
	// Wait for the bidders to do their thing
	for i := 0; i < len(cleanRequests); i++ {
		brw := <-chBids

		//if bidder returned no bids back - remove bidder from further processing
		if brw.adapterBids != nil && len(brw.adapterBids.bids) != 0 {
			adapterBids[brw.bidder] = brw.adapterBids
		}
		//but we need to add all bidders data to adapterExtra to have metrics and other metadata
		adapterExtra[brw.bidder] = brw.adapterExtra

		if !bidsFound && adapterBids[brw.bidder] != nil && len(adapterBids[brw.bidder].bids) > 0 {
			bidsFound = true
			bidIDsCollision = recordAdaptorDuplicateBidIDs(e.me, adapterBids)
		}

	}
	if bidIDsCollision {
		// record this request count this request if bid collision is detected
		e.me.RecordRequestHavingDuplicateBidID()
	}
	return adapterBids, adapterExtra, bidsFound
}

func (e *exchange) recoverSafely(cleanRequests map[openrtb_ext.BidderName]*openrtb.BidRequest, inner func(openrtb_ext.BidderName, openrtb_ext.BidderName, *openrtb.BidRequest, *pbsmetrics.AdapterLabels, currencies.Conversions), chBids chan *bidResponseWrapper) func(openrtb_ext.BidderName, openrtb_ext.BidderName, *openrtb.BidRequest, *pbsmetrics.AdapterLabels, currencies.Conversions) {
	return func(aName openrtb_ext.BidderName, coreBidder openrtb_ext.BidderName, request *openrtb.BidRequest, bidlabels *pbsmetrics.AdapterLabels, conversions currencies.Conversions) {
		defer func() {
			if r := recover(); r != nil {

				allBidders := ""
				sb := strings.Builder{}
				for k := range cleanRequests {
					sb.WriteString(string(k))
					sb.WriteString(",")
				}
				if sb.Len() > 0 {
					allBidders = sb.String()[:sb.Len()-1]
				}

				glog.Errorf("OpenRTB auction recovered panic from Bidder %s: %v. "+
					"Account id: %s, All Bidders: %s, Stack trace is: %v",
					coreBidder, r, bidlabels.PubID, allBidders, string(debug.Stack()))
				e.me.RecordAdapterPanic(*bidlabels)
				// Let the master request know that there is no data here
				brw := new(bidResponseWrapper)
				brw.adapterExtra = new(seatResponseExtra)
				chBids <- brw
			}
		}()
		inner(aName, coreBidder, request, bidlabels, conversions)
	}
}

func bidsToMetric(bids *pbsOrtbSeatBid) pbsmetrics.AdapterBid {
	if bids == nil || len(bids.bids) == 0 {
		return pbsmetrics.AdapterBidNone
	}
	return pbsmetrics.AdapterBidPresent
}

func errorsToMetric(errs []error) map[pbsmetrics.AdapterError]struct{} {
	if len(errs) == 0 {
		return nil
	}
	ret := make(map[pbsmetrics.AdapterError]struct{}, len(errs))
	var s struct{}
	for _, err := range errs {
		switch errortypes.ReadCode(err) {
		case errortypes.TimeoutErrorCode:
			ret[pbsmetrics.AdapterErrorTimeout] = s
		case errortypes.BadInputErrorCode:
			ret[pbsmetrics.AdapterErrorBadInput] = s
		case errortypes.BadServerResponseErrorCode:
			ret[pbsmetrics.AdapterErrorBadServerResponse] = s
		case errortypes.FailedToRequestBidsErrorCode:
			ret[pbsmetrics.AdapterErrorFailedToRequestBids] = s
		default:
			ret[pbsmetrics.AdapterErrorUnknown] = s
		}
	}
	return ret
}

func errsToBidderErrors(errs []error) []openrtb_ext.ExtBidderError {
	serr := make([]openrtb_ext.ExtBidderError, len(errs))
	for i := 0; i < len(errs); i++ {
		serr[i].Code = errortypes.ReadCode(errs[i])
		serr[i].Message = errs[i].Error()
	}
	return serr
}

// This piece takes all the bids supplied by the adapters and crafts an openRTB response to send back to the requester
func (e *exchange) buildBidResponse(ctx context.Context, liveAdapters []openrtb_ext.BidderName, adapterBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid, bidRequest *openrtb.BidRequest, adapterExtra map[openrtb_ext.BidderName]*seatResponseExtra, auc *auction, bidResponseExt *openrtb_ext.ExtBidResponse, returnCreative bool, errList []error) (*openrtb.BidResponse, error) {
	bidResponse := new(openrtb.BidResponse)
	var err error

	bidResponse.ID = bidRequest.ID
	if len(liveAdapters) == 0 {
		// signal "Invalid Request" if no valid bidders.
		bidResponse.NBR = openrtb.NoBidReasonCode.Ptr(openrtb.NoBidReasonCodeInvalidRequest)
	}

	// Create the SeatBids. We use a zero sized slice so that we can append non-zero seat bids, and not include seatBid
	// objects for seatBids without any bids. Preallocate the max possible size to avoid reallocating the array as we go.
	seatBids := make([]openrtb.SeatBid, 0, len(liveAdapters))
	for _, a := range liveAdapters {
		//while processing every single bib, do we need to handle categories here?
		if adapterBids[a] != nil && len(adapterBids[a].bids) > 0 {
			sb := e.makeSeatBid(adapterBids[a], a, adapterExtra, auc, returnCreative)
			seatBids = append(seatBids, *sb)
			bidResponse.Cur = adapterBids[a].currency
		}
	}

	bidResponse.SeatBid = seatBids

	bidResponse.Ext, err = encodeBidResponseExt(bidResponseExt)

	return bidResponse, err
}

func encodeBidResponseExt(bidResponseExt *openrtb_ext.ExtBidResponse) ([]byte, error) {
	buffer := &bytes.Buffer{}
	enc := json.NewEncoder(buffer)

	enc.SetEscapeHTML(false)
	err := enc.Encode(bidResponseExt)

	return buffer.Bytes(), err
}

func applyCategoryMapping(ctx context.Context, bidRequest *openrtb.BidRequest, requestExt *openrtb_ext.ExtRequest, seatBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid, categoriesFetcher stored_requests.CategoryFetcher, targData *targetData) (map[string]string, map[openrtb_ext.BidderName]*pbsOrtbSeatBid, []string, error) {
	res := make(map[string]string)

	type bidDedupe struct {
		bidderName openrtb_ext.BidderName
		bidIndex   int
		bidID      string
		bidPrice   string
	}

	dedupe := make(map[string]bidDedupe)

	impMap := make(map[string]*openrtb.Imp)

	// applyCategoryMapping doesn't get called unless
	// requestExt.Prebid.Targeting != nil && requestExt.Prebid.Targeting.IncludeBrandCategory != nil
	brandCatExt := requestExt.Prebid.Targeting.IncludeBrandCategory

	//If ext.prebid.targeting.includebrandcategory is present in ext then competitive exclusion feature is on.
	var includeBrandCategory = brandCatExt != nil //if not present - category will no be appended
	appendBidderNames := requestExt.Prebid.Targeting.AppendBidderNames

	var primaryAdServer string
	var publisher string
	var err error
	var rejections []string
	var translateCategories = true

	//Maintaining BidRequest Impression Map
	for i := range bidRequest.Imp {
		impMap[bidRequest.Imp[i].ID] = &bidRequest.Imp[i]
	}

	if includeBrandCategory && brandCatExt.WithCategory {
		if brandCatExt.TranslateCategories != nil {
			translateCategories = *brandCatExt.TranslateCategories
		}
		//if translateCategories is set to false, ignore checking primaryAdServer and publisher
		if translateCategories {
			//if ext.prebid.targeting.includebrandcategory present but primaryadserver/publisher not present then error out the request right away.
			primaryAdServer, err = getPrimaryAdServer(brandCatExt.PrimaryAdServer) //1-Freewheel 2-DFP
			if err != nil {
				return res, seatBids, rejections, err
			}
			publisher = brandCatExt.Publisher
		}
	}

	seatBidsToRemove := make([]openrtb_ext.BidderName, 0)

	for bidderName, seatBid := range seatBids {
		bidsToRemove := make([]int, 0)
		for bidInd := range seatBid.bids {
			bid := seatBid.bids[bidInd]
			bidID := bid.bid.ID
			var duration int
			var category string
			var pb string

			if bid.bidVideo != nil {
				duration = bid.bidVideo.Duration
				category = bid.bidVideo.PrimaryCategory
			}
			if brandCatExt.WithCategory && category == "" {
				bidIabCat := bid.bid.Cat
				if len(bidIabCat) != 1 {
					//TODO: add metrics
					//on receiving bids from adapters if no unique IAB category is returned  or if no ad server category is returned discard the bid
					bidsToRemove = append(bidsToRemove, bidInd)
					rejections = updateRejections(rejections, bidID, "Bid did not contain a category")
					continue
				}
				if translateCategories {
					//if unique IAB category is present then translate it to the adserver category based on mapping file
					category, err = categoriesFetcher.FetchCategories(ctx, primaryAdServer, publisher, bidIabCat[0])
					if err != nil || category == "" {
						//TODO: add metrics
						//if mapping required but no mapping file is found then discard the bid
						bidsToRemove = append(bidsToRemove, bidInd)
						reason := fmt.Sprintf("Category mapping file for primary ad server: '%s', publisher: '%s' not found", primaryAdServer, publisher)
						rejections = updateRejections(rejections, bidID, reason)
						continue
					}
				} else {
					//category translation is disabled, continue with IAB category
					category = bidIabCat[0]
				}
			}

			// TODO: consider should we remove bids with zero duration here?

			pb = GetPriceBucket(bid.bid.Price, targData.priceGranularity)

			newDur := duration
			if len(requestExt.Prebid.Targeting.DurationRangeSec) > 0 {
				durationRange := requestExt.Prebid.Targeting.DurationRangeSec
				sort.Ints(durationRange)
				//if the bid is above the range of the listed durations (and outside the buffer), reject the bid
				if duration > durationRange[len(durationRange)-1] {
					bidsToRemove = append(bidsToRemove, bidInd)
					rejections = updateRejections(rejections, bidID, "Bid duration exceeds maximum allowed")
					continue
				}
				for _, dur := range durationRange {
					if duration <= dur {
						newDur = dur
						break
					}
				}
			} else if newDur == 0 {
				if imp, ok := impMap[bid.bid.ImpID]; ok {
					if nil != imp.Video && imp.Video.MaxDuration > 0 {
						newDur = int(imp.Video.MaxDuration)
					}
				}
			}

			var categoryDuration string
			var dupeKey string
			if brandCatExt.WithCategory {
				categoryDuration = fmt.Sprintf("%s_%s_%ds", pb, category, newDur)
				dupeKey = category
			} else {
				categoryDuration = fmt.Sprintf("%s_%ds", pb, newDur)
				dupeKey = categoryDuration
			}

			if appendBidderNames {
				categoryDuration = fmt.Sprintf("%s_%s", categoryDuration, bidderName.String())
			}

			if false == brandCatExt.SkipDedup {
				if dupe, ok := dedupe[dupeKey]; ok {

					dupeBidPrice, err := strconv.ParseFloat(dupe.bidPrice, 64)
					if err != nil {
						dupeBidPrice = 0
					}
					currBidPrice, err := strconv.ParseFloat(pb, 64)
					if err != nil {
						currBidPrice = 0
					}
					if dupeBidPrice == currBidPrice {
						if rand.Intn(100) < 50 {
							dupeBidPrice = -1
						} else {
							currBidPrice = -1
						}
					}

					if dupeBidPrice < currBidPrice {
						if dupe.bidderName == bidderName {
							// An older bid from the current bidder
							bidsToRemove = append(bidsToRemove, dupe.bidIndex)
							rejections = updateRejections(rejections, dupe.bidID, "Bid was deduplicated")
						} else {
							// An older bid from a different seatBid we've already finished with
							oldSeatBid := (seatBids)[dupe.bidderName]
							if len(oldSeatBid.bids) == 1 {
								seatBidsToRemove = append(seatBidsToRemove, dupe.bidderName)
								rejections = updateRejections(rejections, dupe.bidID, "Bid was deduplicated")
							} else {
								oldSeatBid.bids = append(oldSeatBid.bids[:dupe.bidIndex], oldSeatBid.bids[dupe.bidIndex+1:]...)
							}
						}
						delete(res, dupe.bidID)
					} else {
						// Remove this bid
						bidsToRemove = append(bidsToRemove, bidInd)
						rejections = updateRejections(rejections, bidID, "Bid was deduplicated")
						continue
					}
				}
				dedupe[dupeKey] = bidDedupe{bidderName: bidderName, bidIndex: bidInd, bidID: bidID, bidPrice: pb}
			}
			res[bidID] = categoryDuration
		}

		if len(bidsToRemove) > 0 {
			sort.Ints(bidsToRemove)
			if len(bidsToRemove) == len(seatBid.bids) {
				//if all bids are invalid - remove entire seat bid
				seatBidsToRemove = append(seatBidsToRemove, bidderName)
			} else {
				bids := seatBid.bids
				for i := len(bidsToRemove) - 1; i >= 0; i-- {
					remInd := bidsToRemove[i]
					bids = append(bids[:remInd], bids[remInd+1:]...)
				}
				seatBid.bids = bids
			}
		}

	}
	for _, seatBidInd := range seatBidsToRemove {
		seatBids[seatBidInd].bids = nil
	}

	return res, seatBids, rejections, nil
}

func updateRejections(rejections []string, bidID string, reason string) []string {
	message := fmt.Sprintf("bid rejected [bid ID: %s] reason: %s", bidID, reason)
	return append(rejections, message)
}

func getPrimaryAdServer(adServerId int) (string, error) {
	switch adServerId {
	case 1:
		return "freewheel", nil
	case 2:
		return "dfp", nil
	default:
		return "", fmt.Errorf("Primary ad server %d not recognized", adServerId)
	}
}

// Extract all the data from the SeatBids and build the ExtBidResponse
func (e *exchange) makeExtBidResponse(adapterBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid, adapterExtra map[openrtb_ext.BidderName]*seatResponseExtra, req *openrtb.BidRequest, debugInfo bool, errList []error) *openrtb_ext.ExtBidResponse {
	bidResponseExt := &openrtb_ext.ExtBidResponse{
		Errors:               make(map[openrtb_ext.BidderName][]openrtb_ext.ExtBidderError, len(adapterBids)),
		ResponseTimeMillis:   make(map[openrtb_ext.BidderName]int, len(adapterBids)),
		RequestTimeoutMillis: req.TMax,
	}
	if debugInfo {
		bidResponseExt.Debug = &openrtb_ext.ExtResponseDebug{
			HttpCalls:       make(map[openrtb_ext.BidderName][]*openrtb_ext.ExtHttpCall),
			ResolvedRequest: req,
		}
	}

	for bidderName, responseExtra := range adapterExtra {

		if debugInfo {
			bidResponseExt.Debug.HttpCalls[bidderName] = responseExtra.HttpCalls
		}
		// Only make an entry for bidder errors if the bidder reported any.
		if len(responseExtra.Errors) > 0 {
			bidResponseExt.Errors[bidderName] = responseExtra.Errors
		}
		if len(errList) > 0 {
			bidResponseExt.Errors[openrtb_ext.PrebidExtKey] = errsToBidderErrors(errList)
		}
		bidResponseExt.ResponseTimeMillis[bidderName] = responseExtra.ResponseTimeMillis
		// Defering the filling of bidResponseExt.Usersync[bidderName] until later

	}
	return bidResponseExt
}

// Return an openrtb seatBid for a bidder
// BuildBidResponse is responsible for ensuring nil bid seatbids are not included
func (e *exchange) makeSeatBid(adapterBid *pbsOrtbSeatBid, adapter openrtb_ext.BidderName, adapterExtra map[openrtb_ext.BidderName]*seatResponseExtra, auc *auction, returnCreative bool) *openrtb.SeatBid {
	seatBid := new(openrtb.SeatBid)
	seatBid.Seat = adapter.String()
	// Prebid cannot support roadblocking
	seatBid.Group = 0

	if len(adapterBid.ext) > 0 {
		sbExt := ExtSeatBid{
			Bidder: adapterBid.ext,
		}

		ext, err := json.Marshal(sbExt)
		if err != nil {
			extError := openrtb_ext.ExtBidderError{
				Code:    errortypes.ReadCode(err),
				Message: fmt.Sprintf("Error writing SeatBid.Ext: %s", err.Error()),
			}
			adapterExtra[adapter].Errors = append(adapterExtra[adapter].Errors, extError)
		}
		seatBid.Ext = ext
	}

	var errList []error
	seatBid.Bid, errList = e.makeBid(adapterBid.bids, auc, returnCreative)
	if len(errList) > 0 {
		adapterExtra[adapter].Errors = append(adapterExtra[adapter].Errors, errsToBidderErrors(errList)...)
	}

	return seatBid
}

// Create the Bid array inside of SeatBid
func (e *exchange) makeBid(Bids []*pbsOrtbBid, auc *auction, returnCreative bool) ([]openrtb.Bid, []error) {
	bids := make([]openrtb.Bid, 0, len(Bids))
	errList := make([]error, 0, 1)
	for _, thisBid := range Bids {
		bidExt := &openrtb_ext.ExtBid{
			Bidder: thisBid.bid.Ext,
			Prebid: &openrtb_ext.ExtBidPrebid{
				Targeting:         thisBid.bidTargets,
				Type:              thisBid.bidType,
				Video:             thisBid.bidVideo,
				DealPriority:      thisBid.dealPriority,
				DealTierSatisfied: thisBid.dealTierSatisfied,
			},
		}
		if cacheInfo, found := e.getBidCacheInfo(thisBid, auc); found {
			bidExt.Prebid.Cache = &openrtb_ext.ExtBidPrebidCache{
				Bids: &cacheInfo,
			}
		}
		ext, err := json.Marshal(bidExt)
		if err != nil {
			errList = append(errList, err)
		} else {
			bids = append(bids, *thisBid.bid)
			bids[len(bids)-1].Ext = ext
			if !returnCreative {
				bids[len(bids)-1].AdM = ""
			}
		}
	}
	return bids, errList
}

// If bid got cached inside `(a *auction) doCache(ctx context.Context, cache prebid_cache_client.Client, targData *targetData, bidRequest *openrtb.BidRequest, ttlBuffer int64, defaultTTLs *config.DefaultTTLs, bidCategory map[string]string)`,
// a UUID should be found inside `a.cacheIds` or `a.vastCacheIds`. This function returns the UUID along with the internal cache URL
func (e *exchange) getBidCacheInfo(bid *pbsOrtbBid, auction *auction) (cacheInfo openrtb_ext.ExtBidPrebidCacheBids, found bool) {
	uuid, found := findCacheID(bid, auction)

	if found {
		cacheInfo.CacheId = uuid
		cacheInfo.Url = buildCacheURL(e.cache, uuid)
	}

	return
}

func findCacheID(bid *pbsOrtbBid, auction *auction) (string, bool) {
	if bid != nil && bid.bid != nil && auction != nil {
		if id, found := auction.cacheIds[bid.bid]; found {
			return id, true
		}

		if id, found := auction.vastCacheIds[bid.bid]; found {
			return id, true
		}
	}

	return "", false
}

func buildCacheURL(cache prebid_cache_client.Client, uuid string) string {
	scheme, host, path := cache.GetExtCacheData()

	if host == "" || path == "" {
		return ""
	}

	query := url.Values{"uuid": []string{uuid}}
	cacheURL := url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     path,
		RawQuery: query.Encode(),
	}
	cacheURL.Query()

	// URLs without a scheme will begin with //, in which case we
	// want to trim it off to keep compatbile with current behavior.
	return strings.TrimPrefix(cacheURL.String(), "//")
}

func listBiddersWithRequests(cleanRequests map[openrtb_ext.BidderName]*openrtb.BidRequest) []openrtb_ext.BidderName {
	liveAdapters := make([]openrtb_ext.BidderName, len(cleanRequests))
	i := 0
	for a := range cleanRequests {
		liveAdapters[i] = a
		i++
	}
	// Randomize the list of adapters to make the auction more fair
	randomizeList(liveAdapters)

	return liveAdapters
}

// recordAdaptorDuplicateBidIDs finds the bid.id collisions for each bidder and records them with metrics engine
// it returns true if collosion(s) is/are detected in any of the bidder's bids
func recordAdaptorDuplicateBidIDs(metricsEngine pbsmetrics.MetricsEngine, adapterBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid) bool {
	bidIDCollisionFound := false
	if nil == adapterBids {
		return false
	}
	for bidder, bid := range adapterBids {
		bidIDColisionMap := make(map[string]int, len(adapterBids[bidder].bids))
		for _, thisBid := range bid.bids {
			if collisions, ok := bidIDColisionMap[thisBid.bid.ID]; ok {
				bidIDCollisionFound = true
				bidIDColisionMap[thisBid.bid.ID]++
				glog.Warningf("Bid.id %v :: %v collision(s) [imp.id = %v] for bidder '%v'", thisBid.bid.ID, collisions, thisBid.bid.ImpID, string(bidder))
				metricsEngine.RecordAdapterDuplicateBidID(string(bidder), 1)
			} else {
				bidIDColisionMap[thisBid.bid.ID] = 1
			}
		}
	}
	return bidIDCollisionFound
}

//adjustDomain returns tld+1 of the given domain
//checks if domain starts with http by lowercasing entire domain
//if not it prepends it before domain. This is required for obtaining the url
//using url.parse method. if it able to parse url successfully it will then
//find tld+1 of the given domain
func adjustDomain(domain string) (string, error) {
	if strings.Index(domain, "http") == -1 {
		domain = fmt.Sprintf("http://%s", strings.ToLower(domain))
	}
	url, err := url.Parse(strings.Trim(domain, " "))
	if nil == err && url.Host != "" {
		return publicsuffix.EffectiveTLDPlusOne(url.Host)
	}
	return "", err
}

//applyAdvertiserBlocking rejects the bids of blocked advertisers mentioned in req.badv
//the rejection is currently only applicable to vast tag bidders. i.e. not for ortb bidders
//it returns seatbids containing valid bids and rejections containing rejected bid.id with reason
func applyAdvertiserBlocking(bidRequest *openrtb.BidRequest, seatBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid, adapterMap map[openrtb_ext.BidderName]adaptedBidder) (map[openrtb_ext.BidderName]*pbsOrtbSeatBid, []string) {
	rejections := []string{}
	for bidderName, seatBid := range seatBids {
		adptedBidder := adapterMap[bidderName]
		bidder, isBidder := adptedBidder.(*bidderAdapter) // should be non-legacy bidder
		if isBidder {
			// apply advertiser blocking only if bidder is tagbidder
			_, isTagBidder := bidder.Bidder.(*vastbidder.TagBidder)
			if isTagBidder && len(bidRequest.BAdv) > 0 {
				for bidIndex := len(seatBid.bids) - 1; bidIndex >= 0; bidIndex-- {
					bid := seatBid.bids[bidIndex]
					for _, bAdv := range bidRequest.BAdv {
						bAdvTLD, err := adjustDomain(bAdv) // compute once
						bidRejected := false
						if nil == err {
							aDomains := bid.bid.ADomain
							if nil == aDomains {
								aDomains = []string{""} // provision to enable rejecting of bids when req.badv is set
							}
							for _, adomain := range aDomains {
								if aDomainTLD, err := adjustDomain(adomain); nil == err {
									// reject bids if adomain is requested to block or BAdv is set but aDomain is empty
									if aDomainTLD == bAdvTLD || (len(bAdvTLD) > 0 && len(aDomainTLD) == 0) {
										// reject the bid. bid belongs to blocked advertisers list
										seatBid.bids = append(seatBid.bids[:bidIndex], seatBid.bids[bidIndex+1:]...)
										rejections = updateRejections(rejections, bid.bid.ID, fmt.Sprintf("Bid belongs to blocked advertiser '%s'", bAdv))
										bidRejected = true
										break // bid is rejected due to advertiser blocked. No need to check further domains
									}
								}
							}
						}
						if bidRejected {
							break // bid rejected. skip looping over further bAdv values
						}
					}
				}
			}
		}
	}
	return seatBids, rejections
}
