package openwrap

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"golang.org/x/net/context/ctxhttp"
)

type mediaTypes map[string]struct{}

type unwrapReq struct {
	adm       string
	bidId     string
	unwrapCnt int
	err       error
	respTime  int
}

func handleRawBidderResponseHook(

	payload hookstage.RawBidderResponsePayload,
	moduleCtx hookstage.ModuleContext,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {
	//bidder := payload.Bidder

	// allowedBids will store all bids that have passed the attribute check
	allowedBids := make([]*adapters.TypedBid, 0)

	responseChannel := make(chan *unwrapReq, len(payload.Bids))
	for _, bid := range payload.Bids {

		bidMediaTypes := mediaTypesFromBid(bid)
		if _, ok := bidMediaTypes["video"]; ok {
			go vastUnwrapCreative(bid.Bid.AdM, bid.Bid.ID, responseChannel)
		}
	}

	unwrapCrMap := make(map[string]string, len(payload.Bids))
	for i := 0; i < len(payload.Bids); i++ {
		unwrapInfo := <-responseChannel
		if unwrapInfo.err == nil {
			unwrapCrMap[unwrapInfo.bidId] = unwrapInfo.adm
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

func vastUnwrapCreative(in string, bidid string, respChan chan<- *unwrapReq) {
	startTime := time.Now()
	wrapperCnt := 0
	headers := http.Header{}
	headers.Add("Content-Type", "application/xml; charset=utf-8")
	headers.Add("user-agent", "Mozilla/5.0 (QSP; Roku; AP; 5.4.12.227)")
	headers.Add("unwrap-timeout", "1000")
	httpReq, err := http.NewRequest("POST", "http://localhost:8003/unwrap", strings.NewReader(in))
	if err != nil {
		respChan <- &unwrapReq{err: err}
	}
	httpReq.Header = headers
	ctx := context.Background()
	httpResp, err := ctxhttp.Do(ctx, nil, httpReq)
	if err != nil {
		respChan <- &unwrapReq{err: err}
	}

	wrap_cnt := httpResp.Header.Get("unwrap-count")
	if wrap_cnt != "" {
		wrapperCnt, _ = strconv.Atoi(wrap_cnt)
	}

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		respChan <- &unwrapReq{err: err}
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respChan <- &unwrapReq{err: error(fmt.Errorf("Unexpected status code: %d. Run with request.debug = 1 for more info", httpResp.StatusCode))}
	}

	respTime := int(time.Since(startTime).Milliseconds())

	respChan <- &unwrapReq{adm: string(respBody), bidId: bidid, unwrapCnt: wrapperCnt, respTime: respTime, err: err}
	fmt.Printf("\n UnWrap Done for BidId = %v Cnt = %v in %v (ms)", bidid, wrapperCnt, respTime)
}
