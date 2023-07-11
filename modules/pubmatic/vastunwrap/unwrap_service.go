package vastunwrap

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	unwrapper "git.pubmatic.com/vastunwrap"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/adapters"
)

func doUnwrap(m VastUnwrapModule, bid *adapters.TypedBid, userAgent string, unwrapURL string, accountID string, bidder string) {

	startTime := time.Now()
	defer func() {
		respTime := time.Since(startTime)
		m.MetricsEngine.RecordRequestTime(accountID, bidder, respTime)
	}()
	wrapperCnt := 0

	headers := http.Header{}

	headers.Add(ContentType, "application/xml; charset=utf-8")
	headers.Add(UserAgent, userAgent)
	headers.Add(UnwrapTimeout, strconv.Itoa(m.Cfg.APPConfig.UnwrapDefaultTimeout))
	httpReq, err := http.NewRequest(http.MethodPost, unwrapURL, strings.NewReader(bid.Bid.AdM))
	if err != nil {
		m.MetricsEngine.RecordRequestStatus(accountID, bidder, "Failure")
		return
	}
	httpReq.Header = headers

	httpResp := httptest.NewRecorder()
	unwrapper.UnwrapRequest(httpResp, httpReq)
	wrap_cnt := httpResp.Header().Get(UnwrapCount)
	respStatus := httpResp.Header().Get(UnwrapStatus)
	if wrap_cnt != "" {
		wrapperCnt, _ = strconv.Atoi(wrap_cnt)
	}
	respBody := httpResp.Body.Bytes()
	if httpResp.Code == http.StatusOK {
		bid.Bid.AdM = string(respBody)
		glog.Infof("\n UnWrap Done for BidId = %s Cnt = %d in %v (ms)", bid.Bid.ID, wrapperCnt, time.Since(startTime).Milliseconds())
		m.MetricsEngine.RecordRequestStatus(accountID, bidder, Success)
		return
	}
	if respStatus == UnwrapStatusTimeout {
		m.MetricsEngine.RecordRequestStatus(accountID, bidder, Timeout)
		glog.Infof("\n UnWrap Response code = %d for BidId = %s ", httpResp.Code, bid.Bid.ID)
		return
	}
	glog.Infof("\n UnWrap Response code = %d for BidId = %s ", httpResp.Code, bid.Bid.ID)
	m.MetricsEngine.RecordRequestStatus(accountID, bidder, Failure)
	return
}
