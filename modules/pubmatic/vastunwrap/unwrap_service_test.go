package vastunwrap

import (
	"fmt"
	"testing"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/stretchr/testify/assert"
)

func TestDoUnwrap(t *testing.T) {
	type args struct {
		bid                  *adapters.TypedBid
		userAgent            string
		unwrapDefaultTimeout int
		url                  string
	}
	tests := []struct {
		name        string
		args        args
		expectedBid *adapters.TypedBid
	}{
		{
			name: "doUnwrap for adtype video with Empty Bid",
			args: args{
				bid: &adapters.TypedBid{
					Bid: &openrtb2.Bid{},
				},
				userAgent:            "testUA",
				unwrapDefaultTimeout: 1000,
				url:                  UnwrapURL,
			},
		},
		{
			name: "doUnwrap for adtype video with Empty ADM",
			args: args{
				bid: &adapters.TypedBid{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
				userAgent:            "testUA",
				unwrapDefaultTimeout: 1000,
				url:                  UnwrapURL,
			},
			expectedBid: &adapters.TypedBid{
				Bid: &openrtb2.Bid{
					ID:    "Bid-123",
					ImpID: fmt.Sprintf("div-adunit-%d", 123),
					Price: 2.1,
					CrID:  "Cr-234",
					AdM:   "",
					W:     100,
					H:     50,
				},
				BidType: "video",
			},
		},
		{
			name: "doUnwrap for adtype video with invalid URL and timeout",
			args: args{
				bid: &adapters.TypedBid{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   vastXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
				userAgent:            "testUA",
				unwrapDefaultTimeout: 100,
				url:                  "testURL",
			},
			expectedBid: &adapters.TypedBid{
				Bid: &openrtb2.Bid{
					ID:    "Bid-123",
					ImpID: fmt.Sprintf("div-adunit-%d", 123),
					Price: 2.1,
					CrID:  "Cr-234",
					AdM:   vastXMLAdM,
					W:     100,
					H:     50,
				},
				BidType: "video",
			},
		},
		{
			name: "doUnwrap for adtype video",
			args: args{
				bid: &adapters.TypedBid{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   vastXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
				userAgent:            "testUA",
				unwrapDefaultTimeout: 1000,
				url:                  UnwrapURL,
			},
			expectedBid: &adapters.TypedBid{
				Bid: &openrtb2.Bid{
					ID:    "Bid-123",
					ImpID: fmt.Sprintf("div-adunit-%d", 123),
					Price: 2.1,
					CrID:  "Cr-234",
					AdM:   inlineXMLAdM,
					W:     100,
					H:     50,
				},
				BidType: "video",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doUnwrap(tt.args.bid, tt.args.userAgent, tt.args.unwrapDefaultTimeout, tt.args.url)
			if tt.args.bid.Bid.AdM != "" {
				assert.Equal(t, tt.expectedBid.Bid.AdM, tt.args.bid.Bid.AdM, "AdM is not updated correctly after executing RawBidderResponse hook.")
			}

		})
	}
}