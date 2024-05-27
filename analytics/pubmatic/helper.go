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
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/wakanda"
)

const ParseUrlFormat = "json"

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

func setWakandaObject(rCtx *models.RequestCtx, ao *analytics.AuctionObject, loggerURL string) {
	if rCtx.WakandaDebug.Enabled {
		setWakandaWinningBidFlag(&rCtx.WakandaDebug, ao.Response)
		parseURL, err := url.Parse(loggerURL)
		if err != nil {
			glog.Error("Failed to parse loggerURL while setting wakanda object")
		}
		if parseURL != nil {
			jsonParam := parseURL.Query().Get(ParseUrlFormat)
			rCtx.WakandaDebug.DebugData.Logger = json.RawMessage(jsonParam)
		}
		bytes, err := json.Marshal(ao.Response)
		if err != nil {
			glog.Error("Failed to marshal ao.Response while setting wakanda object")
		}
		rCtx.WakandaDebug.DebugData.HTTPResponseBody = string(bytes)
		rCtx.WakandaDebug.DebugData.OpenRTB = ao.RequestWrapper.BidRequest
		rCtx.WakandaDebug.WriteLogToFiles()
	}
}

// setWakandaWinningBidFlag will set WinningBid flag to true if we are getting any positive bid in response
func setWakandaWinningBidFlag(wakandaDebug *wakanda.Debug, response *openrtb2.BidResponse) {
	if response != nil {
		if len(response.SeatBid) > 0 &&
			len(response.SeatBid[0].Bid) > 0 &&
			response.SeatBid[0].Bid[0].Price > 0 {
			wakandaDebug.DebugData.WinningBid = true
		}
	}
}
