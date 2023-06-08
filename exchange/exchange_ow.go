package exchange

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/exchange/entities"
	"github.com/prebid/prebid-server/metrics"
	pubmaticstats "github.com/prebid/prebid-server/metrics/pubmatic_stats"
	"github.com/prebid/prebid-server/openrtb_ext"
	"golang.org/x/net/publicsuffix"
)

const (
	bidCountMetricEnabled = "bidCountMetricEnabled"
	owProfileId           = "owProfileId"
	nodeal                = "nodeal"
	vastVersionUndefined  = "undefined"
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
				glog.Warningf("Bid.id %v :: %v collision(s) [imp.id = %v] for bidder '%v'", thisBid.Bid.ID, collisions, thisBid.Bid.ImpID, string(bidder))
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
func applyAdvertiserBlocking(r *AuctionRequest, seatBids map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid) (map[openrtb_ext.BidderName]*entities.PbsOrtbSeatBid, []string) {
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
						// Add rejectedBid for analytics logging.
						if r.LoggableObject != nil {
							r.LoggableObject.RejectedBids = append(r.LoggableObject.RejectedBids, analytics.RejectedBid{
								RejectionReason: openrtb3.LossBidAdvertiserBlocking,
								Bid:             bid,
								Seat:            seatBid.Seat,
							})
						}
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

func UpdateRejectedBidExt(loggableObject *analytics.LoggableAuctionObject) {
	for _, rejectedBid := range loggableObject.RejectedBids {
		pbsOrtbBid := rejectedBid.Bid

		if loggableObject != nil && pbsOrtbBid != nil && pbsOrtbBid.Bid != nil {

			bidExtPrebid := &openrtb_ext.ExtBidPrebid{
				DealPriority:      pbsOrtbBid.DealPriority,
				DealTierSatisfied: pbsOrtbBid.DealTierSatisfied, //NOT_SET
				Events:            pbsOrtbBid.BidEvents,         //NOT_REQ
				Targeting:         pbsOrtbBid.BidTargets,        //NOT_REQ
				Type:              pbsOrtbBid.BidType,
				Meta:              pbsOrtbBid.BidMeta,
				Video:             pbsOrtbBid.BidVideo,
				BidId:             pbsOrtbBid.GeneratedBidID, //NOT_SET
				Floors:            pbsOrtbBid.BidFloors,
			}

			rejBid := pbsOrtbBid.Bid

			bidExtJSON, err := makeBidExtJSON(rejBid.Ext, bidExtPrebid, nil, pbsOrtbBid.Bid.ImpID,
				pbsOrtbBid.OriginalBidCPM, pbsOrtbBid.OriginalBidCur, pbsOrtbBid.OriginalBidCPMUSD)

			if err != nil {
				glog.Warningf("For bid-id:[%v], bidder:[%v], makeBidExtJSON returned error - [%v]", rejBid.ID, rejectedBid.Seat, err)
				return
			}
			rejBid.Ext = bidExtJSON
		}
	}
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

// recordPartnerTimeout captures the partnertimeout if any at publisher profile level
func recordPartnerTimeout(ctx context.Context, pubID, aliasBidder string) {
	if metricEnabled, ok := ctx.Value(bidCountMetricEnabled).(bool); metricEnabled && ok {
		if profileID, ok := ctx.Value(owProfileId).(string); ok && profileID != "" {
			pubmaticstats.IncPartnerTimeoutInPBS(pubID, profileID, aliasBidder)
		}
	}
}
