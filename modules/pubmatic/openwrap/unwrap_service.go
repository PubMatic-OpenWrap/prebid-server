package openwrap

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"

	vastunwrap "git.pubmatic.com/vastunwrap"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func (m OpenWrap) doUnwrapandUpdateBid(isStatsEnabled bool, bid *adapters.TypedBid, userAgent string, unwrapURL string, accountID string, bidder string) {
	// startTime := time.Now()
	var wrapperCnt int64
	var respStatus string
	if bid == nil || bid.Bid == nil || bid.Bid.AdM == "" {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("AdM:[%s] Error:[%v] stacktrace:[%s]", bid.Bid.AdM, r, string(debug.Stack()))
		}
		// TBDJ
		// respTime := time.Since(startTime)
		// // m.metricEngine.RecordRequestTime(accountID, bidder, respTime)
		// // m.metricEngine.RecordRequestStatus(accountID, bidder, respStatus)
		// // if respStatus == "0" {
		// // 	m.metricEngine.RecordWrapperCount(accountID, bidder, strconv.Itoa(int(wrapperCnt)))
		// // 	m.metricEngine.RecordUnwrapRespTime(accountID, strconv.Itoa(int(wrapperCnt)), respTime)
		// // }
		respStatus = respStatus
		wrapperCnt = wrapperCnt
	}()
	headers := http.Header{}
	headers.Add(models.ContentType, "application/xml; charset=utf-8")
	headers.Add(models.UserAgent, userAgent)
	headers.Add(models.UnwrapTimeout, strconv.Itoa(m.cfg.VastUnwrapCfg.APPConfig.UnwrapDefaultTimeout))
	httpReq, err := http.NewRequest(http.MethodPost, unwrapURL, strings.NewReader(bid.Bid.AdM))
	if err != nil {
		return
	}
	httpReq.Header = headers
	httpResp := NewCustomRecorder()
	vastunwrap.UnwrapRequest(httpResp, httpReq)
	respStatus = httpResp.Header().Get(models.UnwrapStatus)
	wrapperCnt, _ = strconv.ParseInt(httpResp.Header().Get(models.UnwrapCount), 10, 0)
	if !isStatsEnabled && httpResp.Code == http.StatusOK {
		respBody := httpResp.Body.Bytes()
		bid.Bid.AdM = string(respBody)
		return
	}
	glog.Infof("\n UnWrap Response code = %d for BidId = %s ", httpResp.Code, bid.Bid.ID)
	return
}
