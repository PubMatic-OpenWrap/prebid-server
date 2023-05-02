package openwrap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	unwrapper "git.pubmatic.com/vastunwrap"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

	"github.com/prebid/prebid-server/hooks/hookstage"
)

type mediaTypes map[string]struct{}

func handleRawBidderResponseHook(
	payload hookstage.RawBidderResponsePayload,
	moduleCtx hookstage.ModuleContext,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {

	rCtx, ok := moduleCtx[RequestContext].(models.RequestCtx)
	if !ok {
		result.DebugMessages = append(result.DebugMessages, "error: request-ctx not found in handleBeforeValidationHook()")
		return result, nil
	}

	// allowedBids will store all bids that have passed the attribute check
	allowedBids := make([]*adapters.TypedBid, 0)
	responseChannel := make(chan *models.UnwrapReq, len(payload.Bids))
	for _, bid := range payload.Bids {
		bidMediaTypes := mediaTypesFromBid(bid)
		if _, ok := bidMediaTypes["video"]; ok {
			go vastUnwrapCreative(bid.Bid.AdM, rCtx.UA, bid.Bid.ID, responseChannel)
		}
	}

	unwrapCrMap := make(map[string]string, len(payload.Bids))
	for i := 0; i < len(payload.Bids); i++ {
		unwrapInfo := <-responseChannel
		if unwrapInfo.Err == nil {
			unwrapCrMap[unwrapInfo.BidId] = unwrapInfo.Adm
		}
	}

	for _, bid := range payload.Bids {
		bidMediaTypes := mediaTypesFromBid(bid)
		if _, ok := bidMediaTypes["video"]; ok {
			if adm, isPresent := unwrapCrMap[bid.Bid.ID]; isPresent {
				bid.Bid.AdM = adm
			}
		}
		allowedBids = append(allowedBids, bid)
	}

	changeSet := hookstage.ChangeSet[hookstage.RawBidderResponsePayload]{}
	if len(payload.Bids) != len(allowedBids) {
		changeSet.RawBidderResponse().Bids().Update(allowedBids)
		result.ChangeSet = changeSet
	}

	return result, err
}

func mediaTypesFromBid(bid *adapters.TypedBid) mediaTypes {
	return mediaTypes{string(bid.BidType): struct{}{}}
}

func vastUnwrapCreative(in string, ua, bidid string, respChan chan<- *models.UnwrapReq) {
	startTime := time.Now()
	wrapperCnt := 0
	headers := http.Header{}

	headers.Add("Content-Type", "application/xml; charset=utf-8")
	//headers.Add("user-agent", "Mozilla/5.0 (QSP; Roku; AP; 5.4.12.227)")
	headers.Add("user-agent", ua)
	headers.Add("unwrap-timeout", "1000")
	httpReq, err := http.NewRequest("POST", "http://localhost:8003/unwrap", strings.NewReader(in))
	if err != nil {
		respChan <- &models.UnwrapReq{Err: err}
	}
	httpReq.Header = headers

	httpResp := httptest.NewRecorder()
	unwrapper.UnwrapRequest(httpResp, httpReq)

	wrap_cnt := httpResp.Header().Get("unwrap-count")
	if wrap_cnt != "" {
		wrapperCnt, _ = strconv.Atoi(wrap_cnt)
	}

	respBody := httpResp.Body.Bytes()
	if httpResp.Code != http.StatusOK {
		respChan <- &models.UnwrapReq{Err: error(fmt.Errorf("Unexpected status code: %d. Run with request.debug = 1 for more info", httpResp.Code))}
	}

	respTime := int(time.Since(startTime).Milliseconds())

	respChan <- &models.UnwrapReq{Adm: string(respBody), BidId: bidid, UnwrapCnt: wrapperCnt, RespTime: respTime, Err: err}
	fmt.Printf("\n UnWrap Done for BidId = %v Cnt = %v in %v (ms)", bidid, wrapperCnt, respTime)
}
