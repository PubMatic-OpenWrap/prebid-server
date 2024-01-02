package openwrap

import (
	"fmt"

	"git.pubmatic.com/PubMatic/go-netacuity-client"
	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

const (
	keycountry = "country"
)

func getFilteredBidders(rCtx models.RequestCtx, bidRequest *openrtb2.BidRequest, c cache.Cache) (map[string]struct{}, bool) {
	filteredBidders := map[string]struct{}{}
	key := fmt.Sprintf("bidderfilter_%d_%d_%d", rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)
	bf, ok := c.Get(key)
	if !ok {
		return filteredBidders, false
	}
	bidderFilter, ok := bf.(map[string]interface{})
	if !ok {
		return filteredBidders, false
	}
	data := generateEvaluationData(bidRequest, rCtx)
	allPartnersFilteredFlag := true
	for _, partnerConfig := range rCtx.PartnerConfigMap {
		if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
			continue
		}

		if _, ok := rCtx.AdapterThrottleMap[partnerConfig[models.BidderCode]]; ok {
			continue
		}

		biddingCondition, ok := bidderFilter[partnerConfig[models.BidderCode]]
		if ok && !evaluateBiddingCondition(data, biddingCondition) {
			filteredBidders[partnerConfig[models.BidderCode]] = struct{}{}
			continue
		}
		allPartnersFilteredFlag = false
	}

	return filteredBidders, allPartnersFilteredFlag
}

func generateEvaluationData(BidRequest *openrtb2.BidRequest, rCtx models.RequestCtx) map[string]interface{} {
	data := map[string]interface{}{}
	data[keycountry] = getCountryFromRequest(BidRequest, rCtx)
	return data
}

func getCountryFromRequest(bidRequest *openrtb2.BidRequest, rCtx models.RequestCtx) string {
	if bidRequest.Device != nil && bidRequest.Device.Geo != nil && bidRequest.Device.Geo.Country != "" {
		return bidRequest.Device.Geo.Country
	}

	if bidRequest.User != nil && bidRequest.User.Geo != nil && bidRequest.User.Geo.Country != "" {
		return bidRequest.User.Geo.Country
	}

	if rCtx.IP != "" {
		if country, err := getCountryFromIP(rCtx.IP); err != nil {
			return country
		}
	}

	if bidRequest.Device.IP != "" {
		if country, err := getCountryFromIP(bidRequest.Device.IP); err != nil {
			return country
		}
	}

	return ""
}

func evaluateBiddingCondition(data, rules interface{}) bool {
	output, err := jsonlogic.ApplyInterface(rules, data)
	if err != nil {
		glog.Errorf("Error evaluating bidding condition for rules: %v | data: %v | Error: %v", rules, data, err)
		return false
	}
	return output == true
}

func getCountryFromIP(ip string) (string, error) {
	geoData, err := netacuity.LookUp(ip)
	if err != nil {
		return "", err
	}
	return geoData.CountryCode, nil
}
