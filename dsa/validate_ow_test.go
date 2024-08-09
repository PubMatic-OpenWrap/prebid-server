package dsa

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/exchange/entities"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestValidateDSA(t *testing.T) {
	tests := []struct {
		name        string
		giveRequest *openrtb_ext.RequestWrapper
		giveBid     *entities.PbsOrtbBid
		wantError   error
		wantBid     *entities.PbsOrtbBid
	}{
		{
			name: "dsa_present_in_both_req_and_bid",
			giveRequest: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Regs: &openrtb2.Regs{
						Ext: json.RawMessage(`{"dsa": {"dsarequired": 2,"pubrender": 3}}`),
					},
				},
			},
			giveBid: &entities.PbsOrtbBid{
				Bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"dsa":{"adrender": 1}}`),
				},
			},
			wantError: nil,
			wantBid: &entities.PbsOrtbBid{
				Bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"dsa":{"adrender": 1}}`),
				},
			},
		},
		{
			name: "dsa_present_in_req_but_absent_in_bid",
			giveRequest: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Regs: &openrtb2.Regs{
						Ext: json.RawMessage(`{"dsa": {"dsarequired": 0}}`),
					},
				},
			},
			giveBid:   &entities.PbsOrtbBid{},
			wantError: nil,
			wantBid:   &entities.PbsOrtbBid{},
		},
		{
			name: "dsa_present_in_bid_but_absent_in_req",
			giveRequest: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Regs: &openrtb2.Regs{},
				},
			},
			giveBid: &entities.PbsOrtbBid{
				Bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"dsa":{"adrender": 1}}`),
				},
			},
			wantError: nil,
			wantBid: &entities.PbsOrtbBid{
				Bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{}`),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.giveRequest, tt.giveBid)
			if tt.wantError != nil {
				assert.Equal(t, err, tt.wantError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantBid, tt.giveBid, "mismatched bid")
		})
	}
}
