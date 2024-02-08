package openwrap

import (
	"errors"
	"fmt"

	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/geodb"
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
	data := generateEvaluationData(rCtx, bidRequest)
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

func generateEvaluationData(rCtx models.RequestCtx, BidRequest *openrtb2.BidRequest) map[string]interface{} {
	data := map[string]interface{}{}
	data[keycountry] = getCountryFromRequest(rCtx, BidRequest)
	return data
}

func getCountryFromRequest(rCtx models.RequestCtx, bidRequest *openrtb2.BidRequest) string {
	if bidRequest.Device != nil && bidRequest.Device.Geo != nil && bidRequest.Device.Geo.Country != "" {
		return bidRequest.Device.Geo.Country
	}
	if bidRequest.User != nil && bidRequest.User.Geo != nil && bidRequest.User.Geo.Country != "" {
		return bidRequest.User.Geo.Country
	}

	ip := ""
	if bidRequest.Device != nil {
		if bidRequest.Device.IP != "" {
			ip = bidRequest.Device.IP
		} else {
			ip = bidRequest.Device.IPv6
		}
	}

	if ip == "" {
		ip = rCtx.IP
	}

	if ip != "" {
		country, err := getCountryFromIP(rCtx.GeoInfoFetcher, ip)
		if err != nil {
			glog.Errorf("type:[geo_fetch_failed] ip:[%v] error:[%v]", ip, err)
			return ""
		}
		return country
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

func getCountryFromIP(geoInfoFetcher geodb.Geography, ip string) (string, error) {
	if geoInfoFetcher == nil {
		return "", errors.New("geoDB instance is missing")
	}
	geoData, err := geoInfoFetcher.LookUp(ip)
	if err != nil {
		return "", err
	}
	return geoData.AlphaThreeCountryCode, nil
}
