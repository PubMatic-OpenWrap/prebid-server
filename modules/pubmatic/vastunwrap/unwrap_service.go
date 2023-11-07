package vastunwrap

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/adapters"
)

func (m VastUnwrapModule) doUnwrapandUpdateBid(bid *adapters.TypedBid, userAgent string, unwrapURL string, accountID string, bidder string, VastUnwrapStatsEnabled bool) {
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
		m.MetricsEngine.RecordRequestTime(bidder, respTime)
		m.MetricsEngine.RecordRequestStatus(bidder, respStatus)
		if respStatus == "0" {
			m.MetricsEngine.RecordWrapperCount(bidder, strconv.Itoa(int(wrapperCnt)))
		}
	}()
	headers := http.Header{}
	headers.Add(ContentType, "application/xml; charset=utf-8")
	headers.Add(UserAgent, userAgent)
	headers.Add(UnwrapTimeout, strconv.Itoa(m.Cfg.APPConfig.UnwrapDefaultTimeout))
	httpReq, err := http.NewRequest(http.MethodPost, unwrapURL, strings.NewReader(bid.Bid.AdM))
	if err != nil {
		return
	}
	httpReq.Header = headers
	httpResp := NewCustomRecorder()
	m.unwrapRequest(httpResp, httpReq)
	respStatus = httpResp.Header().Get(UnwrapStatus)
	wrapperCnt, _ = strconv.ParseInt(httpResp.Header().Get(UnwrapCount), 10, 0)
	if !VastUnwrapStatsEnabled {
		respBody := httpResp.Body.Bytes()
		if httpResp.Code == http.StatusOK {
			bid.Bid.AdM = string(respBody)
			return
		}
	}

	glog.Infof("\n UnWrap Response code = %d for BidId = %s ", httpResp.Code, bid.Bid.ID)
	return
}
