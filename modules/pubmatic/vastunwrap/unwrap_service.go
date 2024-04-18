package vastunwrap

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/adapters"
)

func (m VastUnwrapModule) doUnwrapandUpdateBid(isStatsEnabled bool, bid *adapters.TypedBid, userAgent string, ip string, unwrapURL string, accountID string, bidder string) {
	startTime := time.Now()
	var wrapperCnt int64
	var respStatus string
	if bid == nil || bid.Bid == nil || bid.Bid.AdM == "" {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("AdM:[%s] Error:[%v] stacktrace:[%s]", bid.Bid.AdM, r, string(debug.Stack()))
		}
		respTime := time.Since(startTime)
		m.MetricsEngine.RecordRequestTime(accountID, bidder, respTime)
		m.MetricsEngine.RecordRequestStatus(accountID, bidder, respStatus)
		if respStatus == "0" {
			m.MetricsEngine.RecordWrapperCount(accountID, bidder, strconv.Itoa(int(wrapperCnt)))
			m.MetricsEngine.RecordUnwrapRespTime(accountID, strconv.Itoa(int(wrapperCnt)), respTime)
		}
	}()
	headers := http.Header{}
	headers.Add(ContentType, "application/xml; charset=utf-8")
	headers.Add(UserAgent, userAgent)
	headers.Add(XUserAgent, userAgent)
	headers.Add(XUserIP, ip)
	headers.Add(CreativeID, bid.Bid.ID)
	headers.Add(UnwrapTimeout, strconv.Itoa(m.Cfg.APPConfig.UnwrapDefaultTimeout))

	unwrapURL = unwrapURL + "?" + PubID + "=" + accountID + "&" + ImpressionID + "=" + bid.Bid.ImpID
	httpReq, err := http.NewRequest(http.MethodPost, unwrapURL, strings.NewReader(bid.Bid.AdM))
	if err != nil {
		return
	}
	httpReq.Header = headers
	httpResp := NewCustomRecorder()
	m.unwrapRequest(httpResp, httpReq)
	respStatus = httpResp.Header().Get(UnwrapStatus)
	wrapperCnt, _ = strconv.ParseInt(httpResp.Header().Get(UnwrapCount), 10, 0)
	if !isStatsEnabled && httpResp.Code == http.StatusOK {
		respBody := httpResp.Body.Bytes()
		bid.Bid.AdM = string(respBody)
	}

	glog.V(3).Infof("[VAST_UNWRAPPER] pubid:[%v] bidder:[%v] impid:[%v] bidid:[%v] status_code:[%v] httpRespCode= [%v] statsEnabled:[%v]",
		accountID, bidder, bid.Bid.ImpID, bid.Bid.ID, respStatus, httpResp.Code, isStatsEnabled)

}
