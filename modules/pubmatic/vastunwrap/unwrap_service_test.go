package vastunwrap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/adapters"
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
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				data, _ := json.Marshal(tt.expectedBid)
				_, _ = w.Write([]byte(data))
			}))
			defer server.Close()
			server.URL = tt.args.url
			doUnwrap(tt.args.bid, tt.args.userAgent, tt.args.unwrapDefaultTimeout, server.URL)
			if tt.args.bid.Bid.AdM != tt.expectedBid.Bid.AdM {
				t.Errorf("Bid Adm is not updated correctly got %q want %q", tt.args.bid.Bid.AdM, tt.expectedBid.Bid.AdM)
			}
		})
	}
}
