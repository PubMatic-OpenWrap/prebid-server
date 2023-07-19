package vastunwrap

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	unwrapper "git.pubmatic.com/vastunwrap"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/adapters"
)

func doUnwrap(m VastUnwrapModule, bid *adapters.TypedBid, userAgent string, unwrapURL string, accountID string, bidder string) {
	startTime := time.Now()
	var respStatus string
	if bid == nil || bid.Bid == nil || bid.Bid.AdM == "" {
		return
	}
	defer func() {
		// respTime := time.Since(startTime)
		// m.MetricsEngine.RecordRequestTime(accountID, bidder, respTime)
		if r := recover(); r != nil {
			glog.Error("AdM:" + bid.Bid.AdM + ". stacktrace:" + string(debug.Stack()))
		}
		m.MetricsEngine.RecordRequestStatus(accountID, bidder, respStatus)
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
	unwrapper.UnwrapRequest(httpResp, httpReq)
	respStatus = httpResp.Header().Get(UnwrapStatus)
	wrapperCnt, _ := strconv.ParseInt(httpResp.Header().Get(UnwrapCount), 10, 0)
	respBody := httpResp.Body.Bytes()
	if httpResp.Code == http.StatusOK {
		bid.Bid.AdM = string(respBody)
		glog.Infof("\n UnWrap Done for BidId = %s Cnt = %d in %v (ms)", bid.Bid.ID, wrapperCnt, time.Since(startTime).Milliseconds())
		return
	}
	glog.Infof("\n UnWrap Response code = %d for BidId = %s ", httpResp.Code, bid.Bid.ID)
	return
}
