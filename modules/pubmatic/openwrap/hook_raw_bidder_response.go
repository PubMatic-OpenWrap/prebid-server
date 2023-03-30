package openwrap

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"golang.org/x/net/context/ctxhttp"
)

type mediaTypes map[string]struct{}

func handleRawBidderResponseHook(

	payload hookstage.RawBidderResponsePayload,
	moduleCtx hookstage.ModuleContext,
) (result hookstage.HookResult[hookstage.RawBidderResponsePayload], err error) {
	bidder := payload.Bidder

	// allowedBids will store all bids that have passed the attribute check
	allowedBids := make([]*adapters.TypedBid, 0)
	for _, bid := range payload.Bids {

		bidMediaTypes := mediaTypesFromBid(bid)
		if _, ok := bidMediaTypes["video"]; ok {
			startTime := time.Now()
			inLineCr, err := vastUnwrapCreative(bid.Bid.AdM)
			respTime := int(time.Since(startTime).Milliseconds())
			if err == nil {
				fmt.Printf("\n Receieved bid ID = %v from = %v type = %v price = %v timeinMs = %v ", bid.Bid.ID, bidder, bidMediaTypes, bid.Bid.Price, respTime)
				bid.Bid.AdM = inLineCr
			} else {
				fmt.Printf("\n vastUnwrapCreative error = %v ", err.Error())
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

func vastUnwrapCreative(in string) (string, error) {
	headers := http.Header{}
	headers.Add("Content-Type", "application/xml; charset=utf-8")
	headers.Add("user-agent", "Mozilla/5.0 (QSP; Roku; AP; 5.4.12.227)")
	headers.Add("unwrap-timeout", "1000")
	httpReq, err := http.NewRequest("POST", "http://localhost:8003/unwrap", strings.NewReader(in))
	if err != nil {
		return in, err
	}
	httpReq.Header = headers
	ctx := context.Background()
	httpResp, err := ctxhttp.Do(ctx, nil, httpReq)
	if err != nil {
		return in, err
	}

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return in, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return in, error(fmt.Errorf("Unexpected status code: %d. Run with request.debug = 1 for more info", httpResp.StatusCode))
	}
	return string(respBody), nil

}
