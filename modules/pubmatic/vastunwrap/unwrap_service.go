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

func doUnwrap(bid *adapters.TypedBid, userAgent string, unwrapDefaultTimeout int, unwrapURL string) {

	startTime := time.Now()
	wrapperCnt := 0

	headers := http.Header{}

	headers.Add(ContentType, "application/xml; charset=utf-8")
	headers.Add(UserAgent, userAgent)
	headers.Add(UnwrapTimeout, strconv.Itoa(unwrapDefaultTimeout))
	httpReq, err := http.NewRequest(POST, unwrapURL, strings.NewReader(bid.Bid.AdM))
	if err != nil {
		return
	}
	httpReq.Header = headers

	httpResp := httptest.NewRecorder()
	unwrapper.UnwrapRequest(httpResp, httpReq)

	wrap_cnt := httpResp.Header().Get(UnwrapCount)
	if wrap_cnt != "" {
		wrapperCnt, _ = strconv.Atoi(wrap_cnt)
	}
	respBody := httpResp.Body.Bytes()
	if httpResp.Code == http.StatusOK {
		bid.Bid.AdM = string(respBody)
		glog.Infof("\n UnWrap Done for BidId = %s Cnt = %d in %v (ms)", bid.Bid.ID, wrapperCnt, time.Since(startTime).Milliseconds())
		return
	}
	return
}
