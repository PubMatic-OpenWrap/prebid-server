package dsa

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/exchange/entities"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
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
		{
			name: "dsa_present_in_bid_but_and_empty_dsa_present_in_req",
			giveRequest: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Regs: &openrtb2.Regs{},
				},
			},
			giveBid: &entities.PbsOrtbBid{
				Bid: &openrtb2.Bid{
					Ext: json.RawMessage(`{"dsa": {}}`),
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

func Test_dropDSA(t *testing.T) {
	type args struct {
		reqDSA *openrtb_ext.ExtRegsDSA
		bidDSA *openrtb_ext.ExtBidDSA
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Nil bidDSA",
			args: args{reqDSA: &openrtb_ext.ExtRegsDSA{}, bidDSA: nil},
			want: false,
		},
		{
			name: "Nil reqDSA",
			args: args{reqDSA: nil, bidDSA: &openrtb_ext.ExtBidDSA{}},
			want: true,
		},
		{
			name: "Nil reqDSA.Required",
			args: args{reqDSA: &openrtb_ext.ExtRegsDSA{Required: nil}, bidDSA: &openrtb_ext.ExtBidDSA{}},
			want: true,
		},
		{
			name: "reqDSA.Required Supported",
			args: args{reqDSA: &openrtb_ext.ExtRegsDSA{Required: ptrutil.ToPtr(Supported)}, bidDSA: &openrtb_ext.ExtBidDSA{}},
			want: false,
		},
		{
			name: "reqDSA.Required Required",
			args: args{reqDSA: &openrtb_ext.ExtRegsDSA{Required: ptrutil.ToPtr(Required)}, bidDSA: &openrtb_ext.ExtBidDSA{}},
			want: false,
		},
		{
			name: "reqDSA.Required RequiredOnlinePlatform",
			args: args{reqDSA: &openrtb_ext.ExtRegsDSA{Required: ptrutil.ToPtr(RequiredOnlinePlatform)}, bidDSA: &openrtb_ext.ExtBidDSA{}},
			want: false,
		},
		{
			name: "reqDSA.Required Other Value",
			args: args{reqDSA: &openrtb_ext.ExtRegsDSA{Required: ptrutil.ToPtr[int8](5)}, bidDSA: &openrtb_ext.ExtBidDSA{}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dropDSA(tt.args.reqDSA, tt.args.bidDSA)
			assert.Equal(t, tt.want, got)
		})
	}
}
