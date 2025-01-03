package exchange

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"git.pubmatic.com/vastunwrap/unwrap"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/currency"
	"github.com/prebid/prebid-server/v2/errortypes"
	"github.com/prebid/prebid-server/v2/exchange/entities"
	"github.com/prebid/prebid-server/v2/metrics"
	pubmaticstats "github.com/prebid/prebid-server/v2/metrics/pubmatic_stats"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/ortb"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"golang.org/x/net/publicsuffix"
)

const (
	bidCountMetricEnabled = "bidCountMetricEnabled"
	owProfileId           = "owProfileId"
	nodeal                = "nodeal"
	vastVersionUndefined  = "undefined"
)

const (
	VASTTypeWrapperEndTag = "</Wrapper>"
	VASTTypeInLineEndTag  = "</InLine>"
)

var validVastVersions = map[int]bool{
	3: true,
	4: true,
}

// VASTTagType describes the allowed values for VASTTagType
type VASTTagType string

const (
	WrapperVASTTagType VASTTagType = "Wrapper"
	InLineVASTTagType  VASTTagType = "InLine"
	URLVASTTagType     VASTTagType = "URL"
	UnknownVASTTagType VASTTagType = "Unknown"
)

var (
	vastVersionRegex = regexp.MustCompile(`<VAST.+version\s*=[\s\\"']*([\s0-9.]+?)[\\\s"']*>`)
)

// recordAdaptorDuplicateBidIDs finds the bid.id collisions for each bidder and records them with metrics engine
// it returns true if collosion(s) is/are detected in any of the bidder's bids
func recordAdaptorDuplicateBidIDs(metricsEngine metrics.MetricsEngine, adapterBids map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid) bool {
	bidIDCollisionFound := false
	if nil == adapterBids {
		return false
	}
	for bidder, bid := range adapterBids {
		bidIDColisionMap := make(map[string]int, len(adapterBids[bidder].Bids))
		for _, thisBid := range bid.Bids {
			if collisions, ok := bidIDColisionMap[thisBid.Bid.ID]; ok {
				bidIDCollisionFound = true
				bidIDColisionMap[thisBid.Bid.ID]++
				glog.V(3).Infof("Bid.id %v :: %v collision(s) [imp.id = %v] for bidder '%v'", thisBid.Bid.ID, collisions, thisBid.Bid.ImpID, string(bidder))
				metricsEngine.RecordAdapterDuplicateBidID(string(bidder), 1)
			} else {
				bidIDColisionMap[thisBid.Bid.ID] = 1
			}
		}
	}
	return bidIDCollisionFound
}

// normalizeDomain validates, normalizes and returns valid domain or error if failed to validate
// checks if domain starts with http by lowercasing entire domain
// if not it prepends it before domain. This is required for obtaining the url
// using url.parse method. on successfull url parsing, it will replace first occurance of www.
// from the domain
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

// applyAdvertiserBlocking rejects the bids of blocked advertisers mentioned in req.badv
// the rejection is currently only applicable to vast tag bidders. i.e. not for ortb bidders
// it returns seatbids containing valid bids and rejections containing rejected bid.id with reason
func applyAdvertiserBlocking(r *AuctionRequest, seatBids map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid, seatNonBids *openrtb_ext.NonBidCollection) (map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid, []string) {
	bidRequest := r.BidRequestWrapper.BidRequest
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

	if len(nBadvs) == 0 {
		return seatBids, rejections
	}

	for bidderName, seatBid := range seatBids {
		if seatBid.BidderCoreName == openrtb_ext.BidderVASTBidder {
			for bidIndex := len(seatBid.Bids) - 1; bidIndex >= 0; bidIndex-- {
				bid := seatBid.Bids[bidIndex]
				for _, bAdv := range nBadvs {
					aDomains := bid.Bid.ADomain
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
						// Add rejected bid in seatNonBid.
						nonBidParams := entities.GetNonBidParamsFromPbsOrtbBid(bid, seatBid.Seat)
						nonBidParams.NonBidReason = int(ResponseRejectedCreativeAdvertiserBlocking)
						seatNonBids.AddBid(openrtb_ext.NewNonBid(nonBidParams), seatBid.Seat)

						// reject the bid. bid belongs to blocked advertisers list
						seatBid.Bids = append(seatBid.Bids[:bidIndex], seatBid.Bids[bidIndex+1:]...)
						rejections = updateRejections(rejections, bid.Bid.ID, fmt.Sprintf("Bid (From '%s') belongs to blocked advertiser '%s'", bidderName, bAdv))
						break // bid is rejected due to advertiser blocked. No need to check further domains
					}
				}
			}
		}
	}
	return seatBids, rejections
}

func recordBids(ctx context.Context, metricsEngine metrics.MetricsEngine, pubID string, adapterBids map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid) {
	// Temporary code to record bids for publishers
	if metricEnabled, ok := ctx.Value(bidCountMetricEnabled).(bool); metricEnabled && ok {
		if profileID, ok := ctx.Value(owProfileId).(string); ok && profileID != "" {
			for _, seatBid := range adapterBids {
				for _, pbsBid := range seatBid.Bids {
					deal := pbsBid.Bid.DealID
					if deal == "" {
						deal = nodeal
					}
					metricsEngine.RecordBids(pubID, profileID, seatBid.Seat, deal)
					pubmaticstats.IncBidResponseByDealCountInPBS(pubID, profileID, seatBid.Seat, deal)
				}
			}
		}
	}
}

func (e *exchange) RecordFastXMLTestMetrics(ctx *unwrap.UnwrapContext, etreeResp, fastxmlResp *unwrap.UnwrapResponse) {
	e.me.RecordXMLParserResponseTime(metrics.XMLParserLabelFastXML, "unwrap", ctx.FastXMLTestCtx.FastXMLStats.ResponseTime)
	e.me.RecordXMLParserResponseTime(metrics.XMLParserLabelETree, "unwrap", ctx.FastXMLTestCtx.ETreeStats.ResponseTime)
	e.me.RecordXMLParserResponseMismatch("unwrap", (etreeResp != fastxmlResp))
}

func recordVastVersion(metricsEngine metrics.MetricsEngine, adapterBids map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid) {
	for _, seatBid := range adapterBids {
		for _, pbsBid := range seatBid.Bids {
			if pbsBid.BidType != openrtb_ext.BidTypeVideo {
				continue
			}
			if pbsBid.Bid.AdM == "" {
				continue
			}
			vastVersion := vastVersionUndefined
			matches := vastVersionRegex.FindStringSubmatch(pbsBid.Bid.AdM)
			if len(matches) == 2 {
				vastVersion = matches[1]
			}

			metricsEngine.RecordVastVersion(string(seatBid.BidderCoreName), vastVersion)
		}
	}
}

func recordOpenWrapBidResponseMetrics(bidder *bidderAdapter, bidResponse *adapters.BidderResponse) {
	if bidResponse == nil {
		return
	}

	if bidResponse.FastXMLMetrics != nil {
		recordFastXMLMetrics(bidder.me, "vastbidder", bidResponse.FastXMLMetrics)
		if bidResponse.FastXMLMetrics.IsRespMismatch {
			resp, _ := jsonutil.Marshal(bidResponse)
			openrtb_ext.FastXMLLogf("\n[XML_PARSER_TEST] method:[vast_bidder] response:[%s]", resp)
		}
	}

	recordVASTTagType(bidder.me, bidResponse, bidder.BidderName)
}

func recordFastXMLMetrics(metricsEngine metrics.MetricsEngine, method string, vastBidderInfo *openrtb_ext.FastXMLMetrics) {
	metricsEngine.RecordXMLParserResponseTime(metrics.XMLParserLabelFastXML, method, vastBidderInfo.XMLParserTime)
	metricsEngine.RecordXMLParserResponseTime(metrics.XMLParserLabelETree, method, vastBidderInfo.EtreeParserTime)
	metricsEngine.RecordXMLParserResponseMismatch(method, vastBidderInfo.IsRespMismatch)
}

func recordVASTTagType(metricsEngine metrics.MetricsEngine, adapterBids *adapters.BidderResponse, bidder openrtb_ext.BidderName) {
	for _, adapterBid := range adapterBids.Bids {
		if adapterBid.BidType == openrtb_ext.BidTypeVideo {
			vastTagType := UnknownVASTTagType
			if index := strings.LastIndex(adapterBid.Bid.AdM, VASTTypeWrapperEndTag); index != -1 {
				vastTagType = WrapperVASTTagType
			} else if index := strings.LastIndex(adapterBid.Bid.AdM, VASTTypeInLineEndTag); index != -1 {
				vastTagType = InLineVASTTagType
			} else if IsUrl(adapterBid.Bid.AdM) {
				vastTagType = URLVASTTagType
			}
			metricsEngine.RecordVASTTagType(string(bidder), string(vastTagType))
		}
	}
}

func IsUrl(adm string) bool {
	url, err := url.Parse(adm)
	return err == nil && url.Scheme != "" && url.Host != ""
}

// recordPartnerTimeout captures the partnertimeout if any at publisher profile level
func recordPartnerTimeout(ctx context.Context, pubID, aliasBidder string) {
	if metricEnabled, ok := ctx.Value(bidCountMetricEnabled).(bool); metricEnabled && ok {
		if profileID, ok := ctx.Value(owProfileId).(string); ok && profileID != "" {
			pubmaticstats.IncPartnerTimeoutInPBS(pubID, profileID, aliasBidder)
		}
	}
}

// updateSeatNonBidsFloors updates seatnonbid with rejectedBids due to floors
func updateSeatNonBidsFloors(seatNonBids *openrtb_ext.NonBidCollection, rejectedBids []*entities.PbsOrtbSeatBid) {
	for _, pbsRejSeatBid := range rejectedBids {
		for _, pbsRejBid := range pbsRejSeatBid.Bids {
			var rejectionReason = ResponseRejectedBelowFloor
			if pbsRejBid.Bid.DealID != "" {
				rejectionReason = ResponseRejectedBelowDealFloor
			}
			nonBidParams := entities.GetNonBidParamsFromPbsOrtbBid(pbsRejBid, pbsRejSeatBid.Seat)
			nonBidParams.NonBidReason = int(rejectionReason)
			seatNonBids.AddBid(openrtb_ext.NewNonBid(nonBidParams), pbsRejSeatBid.Seat)
		}
	}
}

// GetPriceBucketOW is the externally facing function for computing CPM buckets
func GetPriceBucketOW(cpm float64, config openrtb_ext.PriceGranularity) string {
	bid := openrtb2.Bid{
		Price: cpm,
	}
	newPG := setPriceGranularityOW(&config)
	targetData := targetData{
		priceGranularity:          *newPG,
		mediaTypePriceGranularity: openrtb_ext.MediaTypePriceGranularity{},
	}
	return GetPriceBucket(bid, targetData)
}

func setPriceGranularityOW(pg *openrtb_ext.PriceGranularity) *openrtb_ext.PriceGranularity {
	if pg == nil || len(pg.Ranges) == 0 {
		pg = ptrutil.ToPtr(openrtb_ext.NewPriceGranularityDefault())
		return pg
	}

	if pg.Precision == nil {
		pg.Precision = ptrutil.ToPtr(ortb.DefaultPriceGranularityPrecision)
	}

	var prevMax float64 = 0
	for i, r := range pg.Ranges {
		if pg.Ranges[i].Min != prevMax {
			pg.Ranges[i].Min = prevMax
		}
		prevMax = r.Max
	}

	return pg
}

func applyBidPriceThreshold(seatBids map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid, account config.Account, conversions currency.Conversions) (map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid, []*entities.PbsOrtbSeatBid) {
	rejectedBids := []*entities.PbsOrtbSeatBid{}
	if account.BidPriceThreshold != 0 {
		for bidderName, seatBid := range seatBids {
			if seatBid.Bids == nil {
				continue
			}

			eligibleBids := make([]*entities.PbsOrtbBid, 0, len(seatBid.Bids))
			for _, bid := range seatBid.Bids {
				if bid.Bid == nil {
					continue
				}

				price := bid.Bid.Price
				if seatBid.Currency != "" {
					rate, err := conversions.GetRate(seatBid.Currency, "USD")
					if err != nil {
						glog.Error("currencyconversionfailed applyBidPriceThreshold", bid.Bid.ID, seatBid.Currency, err.Error())
					} else {
						price = rate * bid.Bid.Price
					}
				}

				if price <= account.BidPriceThreshold {
					eligibleBids = append(eligibleBids, bid)
				} else {
					bid.Bid.Price = 0
					rejectedBids = append(rejectedBids, &entities.PbsOrtbSeatBid{
						Seat:      seatBid.Seat,
						Currency:  seatBid.Currency,
						HttpCalls: seatBid.HttpCalls,
						Bids:      []*entities.PbsOrtbBid{bid},
					})
				}
			}
			seatBid.Bids = eligibleBids
			seatBids[bidderName] = seatBid

			if len(seatBid.Bids) == 0 {
				delete(seatBids, bidderName)
			}
		}
		logBidsAbovePriceThreshold(rejectedBids)
		for i := range rejectedBids {
			rejectedBids[i].HttpCalls = nil
		}

	}
	return seatBids, rejectedBids
}

func logBidsAbovePriceThreshold(rejectedBids []*entities.PbsOrtbSeatBid) {
	if len(rejectedBids) == 0 {
		return
	}

	var httpCalls []*openrtb_ext.ExtHttpCall
	for i := range rejectedBids {
		httpCalls = append(httpCalls, rejectedBids[i].HttpCalls...)
	}

	if len(httpCalls) > 0 {
		jsonBytes, err := json.Marshal(struct {
			ExtHttpCall []*openrtb_ext.ExtHttpCall
		}{
			ExtHttpCall: httpCalls,
		})
		glog.Error("owbidrejected due to price threshold:", string(jsonBytes), err)
	}
}

func (e exchange) updateSeatNonBidsPriceThreshold(seatNonBids *openrtb_ext.NonBidCollection, rejectedBids []*entities.PbsOrtbSeatBid) {
	for _, pbsRejSeatBid := range rejectedBids {
		for _, pbsRejBid := range pbsRejSeatBid.Bids {
			nonBidParams := entities.GetNonBidParamsFromPbsOrtbBid(pbsRejBid, pbsRejSeatBid.Seat)
			nonBidParams.NonBidReason = int(ResponseRejectedBidPriceTooHigh)
			seatNonBids.AddBid(openrtb_ext.NewNonBid(nonBidParams), pbsRejSeatBid.Seat)
		}
	}
}

func updateSeatNonBidsInvalidVastVersion(seatNonBids *openrtb_ext.NonBidCollection, seat string, rejectedBids []*entities.PbsOrtbBid) {
	for _, pbsRejBid := range rejectedBids {
		nonBidParams := entities.GetNonBidParamsFromPbsOrtbBid(pbsRejBid, seat)
		nonBidParams.NonBidReason = int(nbr.LossBidLostInVastVersionValidation)
		seatNonBids.AddBid(openrtb_ext.NewNonBid(nonBidParams), seat)
	}
}

func filterBidsByVastVersion(adapterBids map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid, seatNonBid *openrtb_ext.NonBidCollection) []error {
	errs := []error{}
	for _, seatBid := range adapterBids {
		rejectedBid := []*entities.PbsOrtbBid{}
		validBids := make([]*entities.PbsOrtbBid, 0, len(seatBid.Bids))
		for _, pbsBid := range seatBid.Bids {
			if pbsBid.BidType == openrtb_ext.BidTypeVideo && pbsBid.Bid.AdM != "" {
				isValid, vastVersion := validateVastVersion(pbsBid.Bid.AdM)
				if !isValid {
					errs = append(errs, &errortypes.Warning{
						Message:     fmt.Sprintf("%s Bid %s was filtered for Imp %s with Vast Version %s: Incompatible with GAM unwinding requirements", seatBid.Seat, pbsBid.Bid.ID, pbsBid.Bid.ImpID, vastVersion),
						WarningCode: errortypes.InvalidVastVersionWarningCode,
					})
					rejectedBid = append(rejectedBid, pbsBid)
					continue
				}
			}
			validBids = append(validBids, pbsBid)
		}
		updateSeatNonBidsInvalidVastVersion(seatNonBid, seatBid.Seat, rejectedBid)
		seatBid.Bids = validBids
	}
	return errs
}

func validateVastVersion(adM string) (bool, string) {
	matches := vastVersionRegex.FindStringSubmatch(adM)
	if len(matches) != 2 {
		return false, ""
	}
	vastVersionFloat, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return false, matches[1]
	}
	return validVastVersions[int(vastVersionFloat)], matches[1]
}
