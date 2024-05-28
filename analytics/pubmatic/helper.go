package pubmatic

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/analytics"
	"github.com/prebid/prebid-server/v2/analytics/pubmatic/mhttp"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// PrepareLoggerURL returns the url for OW logger call
func PrepareLoggerURL(wlog *WloggerRecord, loggerURL string, gdprEnabled int) string {
	if wlog == nil {
		return ""
	}
	v := url.Values{}

	jsonString, err := json.Marshal(wlog.record)
	if err != nil {
		return ""
	}

	v.Set(models.WLJSON, string(jsonString))
	v.Set(models.WLPUBID, strconv.Itoa(wlog.PubID))
	if gdprEnabled == 1 {
		v.Set(models.WLGDPR, strconv.Itoa(gdprEnabled))
	}
	queryString := v.Encode()

	finalLoggerURL := loggerURL + "?" + queryString
	return finalLoggerURL
}

// getGdprEnabledFlag returns gdpr flag set in the partner config
func getGdprEnabledFlag(partnerConfigMap map[int]map[string]string) int {
	gdpr := 0
	if val := partnerConfigMap[models.VersionLevelConfigID][models.GDPR_ENABLED]; val != "" {
		gdpr, _ = strconv.Atoi(val)
	}
	return gdpr
}

// send function will send the owlogger to analytics endpoint
func send(rCtx *models.RequestCtx, url string, headers http.Header, mhc mhttp.MultiHttpContextInterface) {
	startTime := time.Now()
	hc, _ := mhttp.NewHttpCall(url, "")

	for k, v := range headers {
		if len(v) != 0 {
			hc.AddHeader(k, v[0])
		}
	}

	if rCtx.KADUSERCookie != nil {
		hc.AddCookie(models.KADUSERCOOKIE, rCtx.KADUSERCookie.Value)
	}

	mhc.AddHttpCall(hc)
	_, erc := mhc.Execute()
	if erc != 0 {
		glog.Errorf("Failed to send the owlogger for pub:[%d], profile:[%d], version:[%d].",
			rCtx.PubID, rCtx.ProfileID, rCtx.VersionID)

		// we will not record at version level in prometheus metric
		rCtx.MetricsEngine.RecordPublisherWrapperLoggerFailure(rCtx.PubIDStr, rCtx.ProfileIDStr, "")
		return
	}
	rCtx.MetricsEngine.RecordSendLoggerDataTime(time.Since(startTime))
	// TODO: this will increment HB specific metric (ow_pbs_sshb_*), verify labels
}

// RestoreBidResponse restores the original bid response for AppLovinMax from the signal data
func RestoreBidResponse(rctx *models.RequestCtx, ao analytics.AuctionObject) error {
	if rctx.Endpoint != models.EndpointAppLovinMax {
		return nil
	}

	if ao.Response.NBR != nil {
		return nil
	}

	signalData := map[string]string{}
	if err := json.Unmarshal(ao.Response.SeatBid[0].Bid[0].Ext, &signalData); err != nil {
		return err
	}

	if val, ok := signalData[models.SignalData]; !ok || val == "" {
		return errors.New("signal data not found in the response")
	}

	orignalResponse := &openrtb2.BidResponse{}
	if err := json.Unmarshal([]byte(signalData[models.SignalData]), orignalResponse); err != nil {
		return err
	}

	*ao.Response = *orignalResponse
	return nil
}

func (wlog *WloggerRecord) logProfileType(partnerConfigMap map[int]map[string]string) {
	if profileType, ok := partnerConfigMap[models.VersionLevelConfigID][models.ProfileTypeKey]; ok {
		wlog.ProfileType, _ = strconv.Atoi(profileType)
	}
}

func (wlog *WloggerRecord) logProfileTypePlatform(partnerConfigMap map[int]map[string]string) {
	if platform, ok := partnerConfigMap[models.VersionLevelConfigID][models.PLATFORM_KEY]; ok {
		wlog.ProfileTypePlatform = models.ProfileTypePlatform[platform]
	}
}

func (wlog *WloggerRecord) logAppPlatform(partnerConfigMap map[int]map[string]string) {
	if appPlatform, ok := partnerConfigMap[models.VersionLevelConfigID][models.AppPlatformKey]; ok {
		wlog.AppPlatform, _ = strconv.Atoi(appPlatform)
	}
}

func (wlog *WloggerRecord) logAppIntegrationPath(partnerConfigMap map[int]map[string]string) {
	if appIntegrationPathStr, ok := partnerConfigMap[models.VersionLevelConfigID][models.IntegrationPathKey]; ok {
		wlog.AppIntegrationPath = models.AppIntegrationPath[appIntegrationPathStr]
	}
}

func (wlog *WloggerRecord) logAppSubIntegrationPath(partnerConfigMap map[int]map[string]string) {
	if appSubIntegrationPathStr, ok := partnerConfigMap[models.VersionLevelConfigID][models.SubIntegrationPathKey]; ok {
		wlog.AppSubIntegrationPath = models.AppSubIntegrationPath[appSubIntegrationPathStr]
	} else if adserver, ok := partnerConfigMap[models.VersionLevelConfigID][models.AdserverKey]; ok {
		wlog.AppSubIntegrationPath = models.AppSubIntegrationPath[adserver]
	}
}
