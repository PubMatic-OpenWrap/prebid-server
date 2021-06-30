package exchange

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/stored_requests"

	"github.com/gofrs/uuid"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/gdpr"
	"github.com/prebid/prebid-server/metrics"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/prebid_cache_client"
	"golang.org/x/net/publicsuffix"
)

type ContextKey string

const DebugContextKey = ContextKey("debugInfo")

type extCacheInstructions struct {
	cacheBids, cacheVAST, returnCreative bool
}

// Exchange runs Auctions. Implementations must be threadsafe, and will be shared across many goroutines.
type Exchange interface {
	// HoldAuction executes an OpenRTB v2.5 Auction.
	HoldAuction(ctx context.Context, r AuctionRequest, debugLog *DebugLog) (*openrtb2.BidResponse, error)
}

// IdFetcher can find the user's ID for a specific Bidder.
type IdFetcher interface {
	// GetId returns the ID for the bidder. The boolean will be true if the ID exists, and false otherwise.
	GetId(bidder openrtb_ext.BidderName) (string, bool)
	LiveSyncCount() int
}

type exchange struct {
	adapterMap          map[openrtb_ext.BidderName]adaptedBidder
	bidderInfo          config.BidderInfos
	me                  metrics.MetricsEngine
	cache               prebid_cache_client.Client
	cacheTime           time.Duration
	gDPR                gdpr.Permissions
	currencyConverter   *currency.RateConverter
	externalURL         string
	UsersyncIfAmbiguous bool
	privacyConfig       config.Privacy
	categoriesFetcher   stored_requests.CategoryFetcher
	bidIDGenerator      BidIDGenerator
	trakerURL           string
}

// Container to pass out response ext data from the GetAllBids goroutines back into the main thread
type seatResponseExtra struct {
	ResponseTimeMillis int
	Errors             []openrtb_ext.ExtBidderMessage
	Warnings           []openrtb_ext.ExtBidderMessage
	// httpCalls is the list of debugging info. It should only be populated if the request.test == 1.
	// This will become response.ext.debug.httpcalls.{bidder} on the final Response.
	HttpCalls []*openrtb_ext.ExtHttpCall
}

type bidResponseWrapper struct {
	adapterBids  *pbsOrtbSeatBid
	adapterExtra *seatResponseExtra
	bidder       openrtb_ext.BidderName
}

type BidIDGenerator interface {
	New() (string, error)
	Enabled() bool
}

type bidIDGenerator struct {
	enabled bool
}

func (big *bidIDGenerator) Enabled() bool {
	return big.enabled
}

func (big *bidIDGenerator) New() (string, error) {
	rawUuid, err := uuid.NewV4()
	return rawUuid.String(), err
}

func NewExchange(adapters map[openrtb_ext.BidderName]adaptedBidder, cache prebid_cache_client.Client, cfg *config.Configuration, metricsEngine metrics.MetricsEngine, infos config.BidderInfos, gDPR gdpr.Permissions, currencyConverter *currency.RateConverter, categoriesFetcher stored_requests.CategoryFetcher) Exchange {
	return &exchange{
		adapterMap:          adapters,
		bidderInfo:          infos,
		cache:               cache,
		cacheTime:           time.Duration(cfg.CacheURL.ExpectedTimeMillis) * time.Millisecond,
		categoriesFetcher:   categoriesFetcher,
		currencyConverter:   currencyConverter,
		externalURL:         cfg.ExternalURL,
		gDPR:                gDPR,
		me:                  metricsEngine,
		UsersyncIfAmbiguous: cfg.GDPR.UsersyncIfAmbiguous,
		privacyConfig: config.Privacy{
			CCPA: cfg.CCPA,
			GDPR: cfg.GDPR,
			LMT:  cfg.LMT,
		},
		bidIDGenerator: &bidIDGenerator{cfg.GenerateBidID},
		trakerURL:      cfg.TrackerURL,
	}
}

// AuctionRequest holds the bid request for the auction
// and all other information needed to process that request
type AuctionRequest struct {
	BidRequest                 *openrtb2.BidRequest
	Account                    config.Account
	UserSyncs                  IdFetcher
	RequestType                metrics.RequestType
	StartTime                  time.Time
	Warnings                   []error
	GlobalPrivacyControlHeader string

	// LegacyLabels is included here for temporary compatability with cleanOpenRTBRequests
	// in HoldAuction until we get to factoring it away. Do not use for anything new.
	LegacyLabels metrics.Labels
}

// BidderRequest holds the bidder specific request and all other
// information needed to process that bidder request.
type BidderRequest struct {
	BidRequest     *openrtb2.BidRequest
	BidderName     openrtb_ext.BidderName
	BidderCoreName openrtb_ext.BidderName
	BidderLabels   metrics.AdapterLabels
}

func (e *exchange) HoldAuction(ctx context.Context, r AuctionRequest, debugLog *DebugLog) (*openrtb2.BidResponse, error) {
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

	if debugLog == nil {
		debugLog = &DebugLog{Enabled: false, DebugEnabledOrOverridden: false}
	}

	requestDebugInfo := getDebugInfo(r.BidRequest, requestExt)

	debugInfo := debugLog.DebugEnabledOrOverridden || (requestDebugInfo && r.Account.DebugAllow)
	debugLog.Enabled = debugLog.DebugEnabledOrOverridden || r.Account.DebugAllow

	if debugInfo {
		ctx = e.makeDebugContext(ctx, debugInfo)
	}

	bidAdjustmentFactors := getExtBidAdjustmentFactors(requestExt)

	recordImpMetrics(r.BidRequest, e.me)

	// Make our best guess if GDPR applies
	usersyncIfAmbiguous := e.parseUsersyncIfAmbiguous(r.BidRequest)

	// Slice of BidRequests, each a copy of the original cleaned to only contain bidder data for the named bidder
	bidderRequests, privacyLabels, errs := cleanOpenRTBRequests(ctx, r, requestExt, e.gDPR, e.me, usersyncIfAmbiguous, e.privacyConfig, &r.Account)

	e.me.RecordRequestPrivacy(privacyLabels)

	// List of bidders we have requests for.
	liveAdapters := listBiddersWithRequests(bidderRequests)

	// If we need to cache bids, then it will take some time to call prebid cache.
	// We should reduce the amount of time the bidders have, to compensate.
	auctionCtx, cancel := e.makeAuctionContext(ctx, cacheInstructions.cacheBids)
	defer cancel()

	// Get currency rates conversions for the auction
	conversions := e.getAuctionCurrencyRates(requestExt.Prebid.CurrencyConversions)

	adapterBids, adapterExtra, anyBidsReturned := e.getAllBids(auctionCtx, bidderRequests, bidAdjustmentFactors, conversions, r.Account.DebugAllow, r.GlobalPrivacyControlHeader, debugLog.DebugOverride)

	var auc *auction
	var cacheErrs []error
	var bidResponseExt *openrtb_ext.ExtBidResponse
	if anyBidsReturned {

		adapterBids, rejections := applyAdvertiserBlocking(r.BidRequest, adapterBids)
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

		if e.bidIDGenerator.Enabled() {
			for _, seatBid := range adapterBids {
				for _, pbsBid := range seatBid.bids {
					pbsBid.generatedBidID, err = e.bidIDGenerator.New()
					if err != nil {
						errs = append(errs, errors.New("Error generating bid.ext.prebid.bidid"))
					}
				}
			}
		}

		evTracking := getEventTracking(&requestExt.Prebid, r.StartTime, &r.Account, e.bidderInfo, e.externalURL)
		adapterBids = evTracking.modifyBidsForEvents(adapterBids, r.BidRequest, e.trakerURL)

		if targData != nil {
			// A non-nil auction is only needed if targeting is active. (It is used below this block to extract cache keys)
			auc = newAuction(adapterBids, len(r.BidRequest.Imp), targData.preferDeals)
			auc.setRoundedPrices(targData.priceGranularity)

			if requestExt.Prebid.SupportDeals {
				dealErrs := applyDealSupport(r.BidRequest, auc, bidCategory)
				errs = append(errs, dealErrs...)
			}

			bidResponseExt = e.makeExtBidResponse(adapterBids, adapterExtra, r, debugInfo, errs)
			if debugLog.DebugEnabledOrOverridden {
				if bidRespExtBytes, err := json.Marshal(bidResponseExt); err == nil {
					debugLog.Data.Response = string(bidRespExtBytes)
				} else {
					debugLog.Data.Response = "Unable to marshal response ext for debugging"
					errs = append(errs, err)
				}
			}

			cacheErrs = auc.doCache(ctx, e.cache, targData, evTracking, r.BidRequest, 60, &r.Account.CacheTTL, bidCategory, debugLog)
			if len(cacheErrs) > 0 {
				errs = append(errs, cacheErrs...)
			}

			targData.setTargeting(auc, r.BidRequest.App != nil, bidCategory)

		}
		bidResponseExt = e.makeExtBidResponse(adapterBids, adapterExtra, r, debugInfo, errs)
	} else {
		bidResponseExt = e.makeExtBidResponse(adapterBids, adapterExtra, r, debugInfo, errs)

		if debugLog.DebugEnabledOrOverridden {

			if bidRespExtBytes, err := json.Marshal(bidResponseExt); err == nil {
				debugLog.Data.Response = string(bidRespExtBytes)
			} else {
				debugLog.Data.Response = "Unable to marshal response ext for debugging"
				errs = append(errs, err)
			}
		}
	}

	if !r.Account.DebugAllow && requestDebugInfo && !debugLog.DebugOverride {
		accountDebugDisabledWarning := openrtb_ext.ExtBidderMessage{
			Code:    errortypes.AccountLevelDebugDisabledWarningCode,
			Message: "debug turned off for account",
		}
		bidResponseExt.Warnings[openrtb_ext.BidderReservedGeneral] = append(bidResponseExt.Warnings[openrtb_ext.BidderReservedGeneral], accountDebugDisabledWarning)
	}

	for _, warning := range r.Warnings {
		generalWarning := openrtb_ext.ExtBidderMessage{
			Code:    errortypes.ReadCode(warning),
			Message: warning.Error(),
		}
		bidResponseExt.Warnings[openrtb_ext.BidderReservedGeneral] = append(bidResponseExt.Warnings[openrtb_ext.BidderReservedGeneral], generalWarning)
	}

	// Build the response
	return e.buildBidResponse(ctx, liveAdapters, adapterBids, r.BidRequest, adapterExtra, auc, bidResponseExt, cacheInstructions.returnCreative, errs)
}

func (e *exchange) parseUsersyncIfAmbiguous(bidRequest *openrtb2.BidRequest) bool {
	usersyncIfAmbiguous := e.UsersyncIfAmbiguous
	var geo *openrtb2.Geo = nil

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

func recordImpMetrics(bidRequest *openrtb2.BidRequest, metricsEngine metrics.MetricsEngine) {
	for _, impInRequest := range bidRequest.Imp {
		var impLabels metrics.ImpLabels = metrics.ImpLabels{
			BannerImps: impInRequest.Banner != nil,
			VideoImps:  impInRequest.Video != nil,
			AudioImps:  impInRequest.Audio != nil,
			NativeImps: impInRequest.Native != nil,
		}
		metricsEngine.RecordImps(impLabels)
	}
}

// applyDealSupport updates targeting keys with deal prefixes if minimum deal tier exceeded
func applyDealSupport(bidRequest *openrtb2.BidRequest, auc *auction, bidCategory map[string]string) []error {
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
func getDealTiers(bidRequest *openrtb2.BidRequest) map[string]openrtb_ext.DealTierBidderMap {
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
func (e *exchange) getAllBids(
	ctx context.Context,
	bidderRequests []BidderRequest,
	bidAdjustments map[string]float64,
	conversions currency.Conversions,
	accountDebugAllowed bool,
	globalPrivacyControlHeader string,
	headerDebugAllowed bool) (
	map[openrtb_ext.BidderName]*pbsOrtbSeatBid,
	map[openrtb_ext.BidderName]*seatResponseExtra, bool) {
	// Set up pointers to the bid results
	adapterBids := make(map[openrtb_ext.BidderName]*pbsOrtbSeatBid, len(bidderRequests))
	adapterExtra := make(map[openrtb_ext.BidderName]*seatResponseExtra, len(bidderRequests))
	chBids := make(chan *bidResponseWrapper, len(bidderRequests))
	bidsFound := false
	bidIDsCollision := false

	for _, bidder := range bidderRequests {
		// Here we actually call the adapters and collect the bids.
		bidderRunner := e.recoverSafely(bidderRequests, func(bidderRequest BidderRequest, conversions currency.Conversions) {
			// Passing in aName so a doesn't change out from under the go routine
			if bidderRequest.BidderLabels.Adapter == "" {
				glog.Errorf("Exchange: bidlables for %s (%s) missing adapter string", bidderRequest.BidderName, bidderRequest.BidderCoreName)
				bidderRequest.BidderLabels.Adapter = bidderRequest.BidderCoreName
			}
			brw := new(bidResponseWrapper)
			brw.bidder = bidderRequest.BidderName
			// Defer basic metrics to insure we capture them after all the values have been set
			defer func() {
				e.me.RecordAdapterRequest(bidderRequest.BidderLabels)
			}()
			start := time.Now()

			adjustmentFactor := 1.0
			if givenAdjustment, ok := bidAdjustments[string(bidderRequest.BidderName)]; ok {
				adjustmentFactor = givenAdjustment
			}
			reqInfo := adapters.NewExtraRequestInfo(conversions)
			reqInfo.PbsEntryPoint = bidderRequest.BidderLabels.RType
			reqInfo.GlobalPrivacyControlHeader = globalPrivacyControlHeader
			bids, err := e.adapterMap[bidderRequest.BidderCoreName].requestBid(ctx, bidderRequest.BidRequest, bidderRequest.BidderName, adjustmentFactor, conversions, &reqInfo, accountDebugAllowed, headerDebugAllowed)

			// Add in time reporting
			elapsed := time.Since(start)
			brw.adapterBids = bids
			// Structure to record extra tracking data generated during bidding
			ae := new(seatResponseExtra)
			ae.ResponseTimeMillis = int(elapsed / time.Millisecond)

			if bids != nil {
				// Setting bidderCoreName in SeatBid
				bids.bidderCoreName = bidderRequest.BidderCoreName

				ae.HttpCalls = bids.httpCalls

				// Setting bidderCoreName in SeatBid
				bids.bidderCoreName = bidderRequest.BidderCoreName
			}

			// Timing statistics
			e.me.RecordAdapterTime(bidderRequest.BidderLabels, time.Since(start))
			bidderRequest.BidderLabels.AdapterBids = bidsToMetric(brw.adapterBids)
			bidderRequest.BidderLabels.AdapterErrors = errorsToMetric(err)
			// Append any bid validation errors to the error list
			ae.Errors = errsToBidderErrors(err)
			ae.Warnings = errsToBidderWarnings(err)
			brw.adapterExtra = ae
			if bids != nil {
				for _, bid := range bids.bids {
					var cpm = float64(bid.bid.Price * 1000)
					e.me.RecordAdapterPrice(bidderRequest.BidderLabels, cpm)
					e.me.RecordAdapterBidReceived(bidderRequest.BidderLabels, bid.bidType, bid.bid.AdM != "")
					if bid.bidType == openrtb_ext.BidTypeVideo && bid.bidVideo != nil && bid.bidVideo.Duration > 0 {
						e.me.RecordAdapterVideoBidDuration(bidderRequest.BidderLabels, bid.bidVideo.Duration)
					}
				}
			}
			chBids <- brw
		}, chBids)
		go bidderRunner(bidder, conversions)
	}
	// Wait for the bidders to do their thing
	for i := 0; i < len(bidderRequests); i++ {
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

func (e *exchange) recoverSafely(bidderRequests []BidderRequest,
	inner func(BidderRequest, currency.Conversions),
	chBids chan *bidResponseWrapper) func(BidderRequest, currency.Conversions) {
	return func(bidderRequest BidderRequest, conversions currency.Conversions) {
		defer func() {
			if r := recover(); r != nil {

				allBidders := ""
				sb := strings.Builder{}
				for _, bidder := range bidderRequests {
					sb.WriteString(bidder.BidderName.String())
					sb.WriteString(",")
				}
				if sb.Len() > 0 {
					allBidders = sb.String()[:sb.Len()-1]
				}

				bidderRequestStr := ""
				if nil != bidderRequest.BidRequest {
					value, err := json.Marshal(bidderRequest.BidRequest)
					if nil == err {
						bidderRequestStr = string(value)
					} else {
						bidderRequestStr = err.Error()
					}
				}

				glog.Errorf("OpenRTB auction recovered panic from Bidder %s: %v. "+
					"Account id: %s, All Bidders: %s, BidRequest: %s, Stack trace is: %v",
					bidderRequest.BidderCoreName, r, bidderRequest.BidderLabels.PubID, allBidders, bidderRequestStr, string(debug.Stack()))
				e.me.RecordAdapterPanic(bidderRequest.BidderLabels)
				// Let the master request know that there is no data here
				brw := new(bidResponseWrapper)
				brw.adapterExtra = new(seatResponseExtra)
				chBids <- brw
			}
		}()
		inner(bidderRequest, conversions)
	}
}

func bidsToMetric(bids *pbsOrtbSeatBid) metrics.AdapterBid {
	if bids == nil || len(bids.bids) == 0 {
		return metrics.AdapterBidNone
	}
	return metrics.AdapterBidPresent
}

func errorsToMetric(errs []error) map[metrics.AdapterError]struct{} {
	if len(errs) == 0 {
		return nil
	}
	ret := make(map[metrics.AdapterError]struct{}, len(errs))
	var s struct{}
	for _, err := range errs {
		switch errortypes.ReadCode(err) {
		case errortypes.TimeoutErrorCode:
			ret[metrics.AdapterErrorTimeout] = s
		case errortypes.BadInputErrorCode:
			ret[metrics.AdapterErrorBadInput] = s
		case errortypes.BadServerResponseErrorCode:
			ret[metrics.AdapterErrorBadServerResponse] = s
		case errortypes.FailedToRequestBidsErrorCode:
			ret[metrics.AdapterErrorFailedToRequestBids] = s
		default:
			ret[metrics.AdapterErrorUnknown] = s
		}
	}
	return ret
}

func errsToBidderErrors(errs []error) []openrtb_ext.ExtBidderMessage {
	sErr := make([]openrtb_ext.ExtBidderMessage, 0)
	for _, err := range errortypes.FatalOnly(errs) {
		newErr := openrtb_ext.ExtBidderMessage{
			Code:    errortypes.ReadCode(err),
			Message: err.Error(),
		}
		sErr = append(sErr, newErr)
	}

	return sErr
}

func errsToBidderWarnings(errs []error) []openrtb_ext.ExtBidderMessage {
	sWarn := make([]openrtb_ext.ExtBidderMessage, 0)
	for _, warn := range errortypes.WarningOnly(errs) {
		newErr := openrtb_ext.ExtBidderMessage{
			Code:    errortypes.ReadCode(warn),
			Message: warn.Error(),
		}
		sWarn = append(sWarn, newErr)
	}
	return sWarn
}

// This piece takes all the bids supplied by the adapters and crafts an openRTB response to send back to the requester
func (e *exchange) buildBidResponse(ctx context.Context, liveAdapters []openrtb_ext.BidderName, adapterBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid, bidRequest *openrtb2.BidRequest, adapterExtra map[openrtb_ext.BidderName]*seatResponseExtra, auc *auction, bidResponseExt *openrtb_ext.ExtBidResponse, returnCreative bool, errList []error) (*openrtb2.BidResponse, error) {
	bidResponse := new(openrtb2.BidResponse)
	var err error

	bidResponse.ID = bidRequest.ID
	if len(liveAdapters) == 0 {
		// signal "Invalid Request" if no valid bidders.
		bidResponse.NBR = openrtb2.NoBidReasonCode.Ptr(openrtb2.NoBidReasonCodeInvalidRequest)
	}

	// Create the SeatBids. We use a zero sized slice so that we can append non-zero seat bids, and not include seatBid
	// objects for seatBids without any bids. Preallocate the max possible size to avoid reallocating the array as we go.
	seatBids := make([]openrtb2.SeatBid, 0, len(liveAdapters))
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

func applyCategoryMapping(ctx context.Context, bidRequest *openrtb2.BidRequest, requestExt *openrtb_ext.ExtRequest, seatBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid, categoriesFetcher stored_requests.CategoryFetcher, targData *targetData) (map[string]string, map[openrtb_ext.BidderName]*pbsOrtbSeatBid, []string, error) {
	res := make(map[string]string)

	type bidDedupe struct {
		bidderName openrtb_ext.BidderName
		bidIndex   int
		bidID      string
		bidPrice   string
	}

	dedupe := make(map[string]bidDedupe)

	impMap := make(map[string]*openrtb2.Imp)

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

			if !brandCatExt.SkipDedup {
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
func (e *exchange) makeExtBidResponse(adapterBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid, adapterExtra map[openrtb_ext.BidderName]*seatResponseExtra, r AuctionRequest, debugInfo bool, errList []error) *openrtb_ext.ExtBidResponse {
	req := r.BidRequest
	bidResponseExt := &openrtb_ext.ExtBidResponse{
		Errors:               make(map[openrtb_ext.BidderName][]openrtb_ext.ExtBidderMessage, len(adapterBids)),
		Warnings:             make(map[openrtb_ext.BidderName][]openrtb_ext.ExtBidderMessage, len(adapterBids)),
		ResponseTimeMillis:   make(map[openrtb_ext.BidderName]int, len(adapterBids)),
		RequestTimeoutMillis: req.TMax,
	}
	if debugInfo {
		bidResponseExt.Debug = &openrtb_ext.ExtResponseDebug{
			HttpCalls:       make(map[openrtb_ext.BidderName][]*openrtb_ext.ExtHttpCall),
			ResolvedRequest: req,
		}
	}
	if !r.StartTime.IsZero() {
		// auctiontimestamp is the only response.ext.prebid attribute we may emit
		bidResponseExt.Prebid = &openrtb_ext.ExtResponsePrebid{
			AuctionTimestamp: r.StartTime.UnixNano() / 1e+6,
		}
	}

	for bidderName, responseExtra := range adapterExtra {

		if debugInfo && len(responseExtra.HttpCalls) > 0 {
			bidResponseExt.Debug.HttpCalls[bidderName] = responseExtra.HttpCalls
		}
		if len(responseExtra.Warnings) > 0 {
			bidResponseExt.Warnings[bidderName] = responseExtra.Warnings
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
func (e *exchange) makeSeatBid(adapterBid *pbsOrtbSeatBid, adapter openrtb_ext.BidderName, adapterExtra map[openrtb_ext.BidderName]*seatResponseExtra, auc *auction, returnCreative bool) *openrtb2.SeatBid {
	seatBid := &openrtb2.SeatBid{
		Seat:  adapter.String(),
		Group: 0, // Prebid cannot support roadblocking
	}

	var errList []error
	seatBid.Bid, errList = e.makeBid(adapterBid.bids, auc, returnCreative)
	if len(errList) > 0 {
		adapterExtra[adapter].Errors = append(adapterExtra[adapter].Errors, errsToBidderErrors(errList)...)
	}

	return seatBid
}

func (e *exchange) makeBid(bids []*pbsOrtbBid, auc *auction, returnCreative bool) ([]openrtb2.Bid, []error) {
	result := make([]openrtb2.Bid, 0, len(bids))
	errs := make([]error, 0, 1)

	for _, bid := range bids {
		bidExtPrebid := &openrtb_ext.ExtBidPrebid{
			DealPriority:      bid.dealPriority,
			DealTierSatisfied: bid.dealTierSatisfied,
			Events:            bid.bidEvents,
			Targeting:         bid.bidTargets,
			Type:              bid.bidType,
			Video:             bid.bidVideo,
			BidId:             bid.generatedBidID,
		}

		if cacheInfo, found := e.getBidCacheInfo(bid, auc); found {
			bidExtPrebid.Cache = &openrtb_ext.ExtBidPrebidCache{
				Bids: &cacheInfo,
			}
		}

		if bidExtJSON, err := makeBidExtJSON(bid.bid.Ext, bidExtPrebid); err != nil {
			errs = append(errs, err)
		} else {
			result = append(result, *bid.bid)
			resultBid := &result[len(result)-1]
			resultBid.Ext = bidExtJSON
			if !returnCreative {
				resultBid.AdM = ""
			}
		}
	}
	return result, errs
}

func makeBidExtJSON(ext json.RawMessage, prebid *openrtb_ext.ExtBidPrebid) (json.RawMessage, error) {
	// no existing bid.ext. generate a bid.ext with just our prebid section populated.
	if len(ext) == 0 {
		bidExt := &openrtb_ext.ExtBid{Prebid: prebid}
		return json.Marshal(bidExt)
	}

	// update existing bid.ext with our prebid section. if bid.ext.prebid already exists, it will be overwritten.
	var extMap map[string]interface{}
	if err := json.Unmarshal(ext, &extMap); err != nil {
		return nil, err
	}
	extMap[openrtb_ext.PrebidExtKey] = prebid
	return json.Marshal(extMap)
}

// If bid got cached inside `(a *auction) doCache(ctx context.Context, cache prebid_cache_client.Client, targData *targetData, bidRequest *openrtb2.BidRequest, ttlBuffer int64, defaultTTLs *config.DefaultTTLs, bidCategory map[string]string)`,
// a UUID should be found inside `a.cacheIds` or `a.vastCacheIds`. This function returns the UUID along with the internal cache URL
func (e *exchange) getBidCacheInfo(bid *pbsOrtbBid, auction *auction) (cacheInfo openrtb_ext.ExtBidPrebidCacheBids, found bool) {
	uuid, found := findCacheID(bid, auction)

	if found {
		cacheInfo.CacheId = uuid
		cacheInfo.Url = buildCacheURL(e.cache, uuid)
	}

	return
}

func (e *exchange) getAuctionCurrencyRates(requestRates *openrtb_ext.ExtRequestCurrency) currency.Conversions {
	if requestRates == nil {
		// No bidRequest.ext.currency field was found, use PBS rates as usual
		return e.currencyConverter.Rates()
	}

	// If bidRequest.ext.currency.usepbsrates is nil, we understand its value as true. It will be false
	// only if it's explicitly set to false
	usePbsRates := requestRates.UsePBSRates == nil || *requestRates.UsePBSRates

	if !usePbsRates {
		// At this point, we can safely assume the ConversionRates map is not empty because
		// validateCustomRates(bidReqCurrencyRates *openrtb_ext.ExtRequestCurrency) would have
		// thrown an error under such conditions.
		return currency.NewRates(time.Time{}, requestRates.ConversionRates)
	}

	// Both PBS and custom rates can be used, check if ConversionRates is not empty
	if len(requestRates.ConversionRates) == 0 {
		// Custom rates map is empty, use PBS rates only
		return e.currencyConverter.Rates()
	}

	// Return an AggregateConversions object that includes both custom and PBS currency rates but will
	// prioritize custom rates over PBS rates whenever a currency rate is found in both
	return currency.NewAggregateConversions(currency.NewRates(time.Time{}, requestRates.ConversionRates), e.currencyConverter.Rates())
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

func listBiddersWithRequests(bidderRequests []BidderRequest) []openrtb_ext.BidderName {
	liveAdapters := make([]openrtb_ext.BidderName, len(bidderRequests))
	i := 0
	for _, bidderRequest := range bidderRequests {
		liveAdapters[i] = bidderRequest.BidderName
		i++
	}
	// Randomize the list of adapters to make the auction more fair
	randomizeList(liveAdapters)

	return liveAdapters
}

// recordAdaptorDuplicateBidIDs finds the bid.id collisions for each bidder and records them with metrics engine
// it returns true if collosion(s) is/are detected in any of the bidder's bids
func recordAdaptorDuplicateBidIDs(metricsEngine metrics.MetricsEngine, adapterBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid) bool {
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

//normalizeDomain validates, normalizes and returns valid domain or error if failed to validate
//checks if domain starts with http by lowercasing entire domain
//if not it prepends it before domain. This is required for obtaining the url
//using url.parse method. on successfull url parsing, it will replace first occurance of www.
//from the domain
func normalizeDomain(domain string) (string, error) {
	domain = strings.Trim(strings.ToLower(domain), " ")
	// not checking if it belongs to icann
	suffix, _ := publicsuffix.PublicSuffix(domain)
	if domain != "" && suffix == domain { // input is publicsuffix
		return "", errors.New("domain [" + domain + "] is public suffix")
	}
	if !strings.HasPrefix(domain, "http") {
		domain = fmt.Sprintf("http://%s", domain)
	}
	url, err := url.Parse(domain)
	if nil == err && url.Host != "" {
		return strings.Replace(url.Host, "www.", "", 1), nil
	}
	return "", err
}

//applyAdvertiserBlocking rejects the bids of blocked advertisers mentioned in req.badv
//the rejection is currently only applicable to vast tag bidders. i.e. not for ortb bidders
//it returns seatbids containing valid bids and rejections containing rejected bid.id with reason
func applyAdvertiserBlocking(bidRequest *openrtb2.BidRequest, seatBids map[openrtb_ext.BidderName]*pbsOrtbSeatBid) (map[openrtb_ext.BidderName]*pbsOrtbSeatBid, []string) {
	rejections := []string{}
	nBadvs := []string{}
	if nil != bidRequest.BAdv {
		for _, domain := range bidRequest.BAdv {
			nDomain, err := normalizeDomain(domain)
			if nil == err && nDomain != "" { // skip empty and domains with errors
				nBadvs = append(nBadvs, nDomain)
			}
		}
	}

	for bidderName, seatBid := range seatBids {
		if seatBid.bidderCoreName == openrtb_ext.BidderVASTBidder && len(nBadvs) > 0 {
			for bidIndex := len(seatBid.bids) - 1; bidIndex >= 0; bidIndex-- {
				bid := seatBid.bids[bidIndex]
				for _, bAdv := range nBadvs {
					aDomains := bid.bid.ADomain
					rejectBid := false
					if nil == aDomains {
						// provision to enable rejecting of bids when req.badv is set
						rejectBid = true
					} else {
						for _, d := range aDomains {
							if aDomain, err := normalizeDomain(d); nil == err {
								// compare and reject bid if
								// 1. aDomain == bAdv
								// 2. .bAdv is suffix of aDomain
								// 3. aDomain not present but request has list of block advertisers
								if aDomain == bAdv || strings.HasSuffix(aDomain, "."+bAdv) || (len(aDomain) == 0 && len(bAdv) > 0) {
									// aDomain must be subdomain of bAdv
									rejectBid = true
									break
								}
							}
						}
					}
					if rejectBid {
						// reject the bid. bid belongs to blocked advertisers list
						seatBid.bids = append(seatBid.bids[:bidIndex], seatBid.bids[bidIndex+1:]...)
						rejections = updateRejections(rejections, bid.bid.ID, fmt.Sprintf("Bid (From '%s') belongs to blocked advertiser '%s'", bidderName, bAdv))
						break // bid is rejected due to advertiser blocked. No need to check further domains
					}
				}
			}
		}
	}
	return seatBids, rejections
}
