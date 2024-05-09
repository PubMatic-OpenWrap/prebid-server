package openwrap

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/errortypes"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestGetNonBRCodeFromBidRespExt(t *testing.T) {
	type args struct {
		bidder         string
		bidResponseExt openrtb_ext.ExtBidResponse
	}
	tests := []struct {
		name string
		args args
		nbr  *openrtb3.NoBidReason
	}{
		{
			name: "bidResponseExt.Errors_is_empty",
			args: args{
				bidder: "pubmatic",
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Errors: nil,
				},
			},
			nbr: openrtb3.NoBidUnknownError.Ptr(),
		},
		{
			name: "invalid_partner_err",
			args: args{
				bidder: "pubmatic",
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Errors: map[openrtb_ext.BidderName][]openrtb_ext.ExtBidderMessage{
						"pubmatic": {
							{
								Code: 0,
							},
						},
					},
				},
			},
			nbr: exchange.ErrorGeneral.Ptr(),
		},
		{
			name: "unknown_partner_err",
			args: args{
				bidder: "pubmatic",
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Errors: map[openrtb_ext.BidderName][]openrtb_ext.ExtBidderMessage{
						"pubmatic": {
							{
								Code: errortypes.UnknownErrorCode,
							},
						},
					},
				},
			},
			nbr: exchange.ErrorGeneral.Ptr(),
		},
		{
			name: "partner_timeout_err",
			args: args{
				bidder: "pubmatic",
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Errors: map[openrtb_ext.BidderName][]openrtb_ext.ExtBidderMessage{
						"pubmatic": {
							{
								Code: errortypes.TimeoutErrorCode,
							},
						},
					},
				},
			},
			nbr: exchange.ErrorTimeout.Ptr(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nbr := getNonBRCodeFromBidRespExt(tt.args.bidder, tt.args.bidResponseExt)
			assert.Equal(t, tt.nbr, nbr, tt.name)
		})
	}
}
