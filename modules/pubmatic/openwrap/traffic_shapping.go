package openwrap

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) getFilteredBidders(rCtx models.RequestCtx, bidRequest *openrtb2.BidRequest) (map[string]struct{}, bool) {
	filteredBidders := map[string]struct{}{}
	data := m.generateEvaluationData(rCtx, bidRequest)
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

func (m OpenWrap) generateEvaluationData(rCtx models.RequestCtx, bidRequest *openrtb2.BidRequest) string {
	builder := &strings.Builder{}
	builder.WriteString("{")
	country := m.getCountryFromRequest(rCtx)
	builder.WriteString(fmt.Sprintf(`"country":"%s"`, country))
	builder.WriteString("}")
	return builder.String()
}

func (m OpenWrap) getCountryFromRequest(rctx models.RequestCtx) string {
	if len(rctx.DeviceCtx.Country) > 0 {
		return rctx.DeviceCtx.Country
	}

	if rctx.DeviceCtx.IP != "" {
		_, alpha3CountryCode := m.getCountryCodes(rctx.DeviceCtx.IP)
		return alpha3CountryCode
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

func (m OpenWrap) getCountryCodes(ip string) (string, string) {
	if m.geoInfoFetcher == nil {
		return "", ""
	}

	geoData, err := m.geoInfoFetcher.LookUp(ip)
	if geoData == nil || err != nil {
		glog.Errorf("[geolookup] ip:[%s] error:[%s]", ip, err.Error())
		return "", ""
	}
	return geoData.ISOCountryCode, geoData.AlphaThreeCountryCode
}
