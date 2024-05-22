package openwrap

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) getFilteredBidders(rCtx models.RequestCtx, bidRequest *openrtb2.BidRequest) (map[string]struct{}, bool) {
	filteredBidders := map[string]struct{}{}
	data := generateEvaluationData(rCtx, bidRequest, m.geoInfoFetcher)
	allPartnersFilteredFlag := true
	for _, partnerConfig := range rCtx.PartnerConfigMap {
		if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
			continue
		}

		if _, ok := rCtx.AdapterThrottleMap[partnerConfig[models.BidderCode]]; ok {
			continue
		}

		biddingCondition, ok := partnerConfig[models.BidderFilters]
		if ok && !evaluateBiddingCondition(data, biddingCondition) {
			filteredBidders[partnerConfig[models.BidderCode]] = struct{}{}
			continue
		}
		allPartnersFilteredFlag = false
	}

	return filteredBidders, allPartnersFilteredFlag
}

func generateEvaluationData(rCtx models.RequestCtx, bidRequest *openrtb2.BidRequest, gif geodb.Geography) string {
	builder := &strings.Builder{}
	builder.WriteString("{")
	country := getCountryFromRequest(rCtx, gif, bidRequest)
	builder.WriteString(fmt.Sprintf(`"country":"%s"`, country))
	builder.WriteString("}")
	return builder.String()
}

func getCountryFromRequest(rctx models.RequestCtx, gif geodb.Geography, bidRequest *openrtb2.BidRequest) string {
	if len(rctx.Country) > 0 {
		return rctx.Country
	}

	if rctx.IP != "" {
		country, err := getCountryFromIP(gif, rctx.IP)
		if err == nil {
			return country
		}
	}
	return ""
}

func evaluateBiddingCondition(data, rules string) bool {
	var result bytes.Buffer
	err := jsonlogic.Apply(strings.NewReader(rules), strings.NewReader(data), &result)
	if err != nil {
		glog.Errorf("Error evaluating bidding condition for rules: %v | data: %v | Error: %v", rules, data, err)
		return false
	}
	return strings.TrimSpace(result.String()) == "true"
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
