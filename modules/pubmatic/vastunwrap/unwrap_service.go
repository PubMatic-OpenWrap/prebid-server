package vastunwrap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	unwrapper "git.pubmatic.com/vastunwrap"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/adapters"
)

type UnwrapVast interface {
	Dovastunwrap(r *http.Request, w http.ResponseWriter)
}
type Vast struct {
}

func (v *Vast) Dovastunwrap(r *http.Request, w http.ResponseWriter) {
	unwrapper.UnwrapRequest(w, r)

}
func doUnwrap(bid *adapters.TypedBid, userAgent string, unwrapDefaultTimeout int, unwrapURL string, vast UnwrapVast) {

	startTime := time.Now()
	wrapperCnt := 0

	headers := http.Header{}

	headers.Add(ContentType, "application/xml; charset=utf-8")
	headers.Add(UserAgent, userAgent)
	headers.Add(UnwrapTimeout, strconv.Itoa(unwrapDefaultTimeout))
	httpReq, err := http.NewRequest(http.MethodPost, unwrapURL, strings.NewReader(bid.Bid.AdM))
	if err != nil {
		return
	}
	httpReq.Header = headers

	httpResp := httptest.NewRecorder()
	vast.Dovastunwrap(httpReq, httpResp)
	wrap_cnt := httpResp.Header().Get(UnwrapCount)
	if wrap_cnt != "" {
		wrapperCnt, _ = strconv.Atoi(wrap_cnt)
	}
	respBody := httpResp.Body.Bytes()
	fmt.Printf("\n status code - %v \n ", httpResp.Code)
	fmt.Printf("\n resp ADM- %v \n", string(respBody))
	if httpResp.Code == http.StatusOK {
		bid.Bid.AdM = string(respBody)
		glog.Infof("\n UnWrap Done for BidId = %s Cnt = %d in %v (ms)", bid.Bid.ID, wrapperCnt, time.Since(startTime).Milliseconds())
	}
	// glog.Infof("\n UnWrap Response code = %s for BidId = %s ", httpResp.Code, bid.Bid.ID)
	return
}
