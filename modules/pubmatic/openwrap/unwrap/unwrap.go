package unwrap

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	vastunwrap "git.pubmatic.com/vastunwrap"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/adapters"
	metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type Unwrap struct {
	endpoint      string
	defaultTime   int
	metricEngine  metrics.MetricsEngine
	unwrapRequest http.HandlerFunc
}

type VastUnwrapService interface {
	Unwrap(accountID string, bidder string, bid *adapters.TypedBid, userAgent string, ip string, isStatsEnabled bool)
}

func NewUnwrap(Endpoint string, DefaultTime int, handler http.HandlerFunc, MetricEngine metrics.MetricsEngine) Unwrap {
	uw := Unwrap{
		endpoint:      Endpoint,
		defaultTime:   DefaultTime,
		unwrapRequest: vastunwrap.UnwrapRequest,
		metricEngine:  MetricEngine,
	}

	if handler != nil {
		uw.unwrapRequest = handler
	}
	return uw

}

func (uw Unwrap) Unwrap(bid *adapters.TypedBid, accountID, bidder, userAgent, ip string) (unwrapStatus string) {
	startTime := time.Now()
	var wrapperCnt int64
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("AdM:[%s] Error:[%v] stacktrace:[%s]", bid.Bid.AdM, r, string(debug.Stack()))
		}
		respTime := time.Since(startTime)
		uw.metricEngine.RecordUnwrapRequestTime(accountID, bidder, respTime)
		uw.metricEngine.RecordUnwrapRequestStatus(accountID, bidder, unwrapStatus)
		if unwrapStatus == "0" {
			uw.metricEngine.RecordUnwrapWrapperCount(accountID, bidder, strconv.Itoa(int(wrapperCnt)))
			uw.metricEngine.RecordUnwrapRespTime(accountID, strconv.Itoa(int(wrapperCnt)), respTime)
		}
	}()

	unwrapURL := uw.endpoint + "?" + models.PubID + "=" + accountID + "&" + models.ImpressionID + "=" + bid.Bid.ImpID
	httpReq, err := http.NewRequest(http.MethodPost, unwrapURL, strings.NewReader(bid.Bid.AdM))
	if err != nil {
		return
	}
	headers := http.Header{}
	headers.Add(models.ContentType, "application/xml; charset=utf-8")
	headers.Add(models.UserAgent, userAgent)
	headers.Add(models.XUserAgent, userAgent)
	headers.Add(models.XUserIP, ip)
	headers.Add(models.CreativeID, bid.Bid.ID)
	headers.Add(models.UnwrapTimeout, strconv.Itoa(uw.defaultTime))

	httpReq.Header = headers
	httpResp := NewCustomRecorder()
	uw.unwrapRequest(httpResp, httpReq)
	unwrapStatus = httpResp.Header().Get(models.UnwrapStatus)
	wrapperCnt, _ = strconv.ParseInt(httpResp.Header().Get(models.UnwrapCount), 10, 0)
	if httpResp.Code == http.StatusOK && unwrapStatus == models.UnwrapSucessStatus {
		respBody := httpResp.Body.Bytes()
		bid.Bid.AdM = string(respBody)
	}

	glog.V(models.LogLevelDebug).Infof("[VAST_UNWRAPPER] pubid:[%v] bidder:[%v] impid:[%v] bidid:[%v] status_code:[%v] wrapper_cnt:[%v] httpRespCode= [%v]",
		accountID, bidder, bid.Bid.ImpID, bid.Bid.ID, unwrapStatus, wrapperCnt, httpResp.Code)
	return unwrapStatus
}
