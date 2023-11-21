package openwrap

import (
	"testing"

	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
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
		nbr  *openrtb3.NonBidStatusCode
	}{
		{
			name: "bidResponseExt.Errors_is_empty",
			args: args{
				bidder: "pubmatic",
				bidResponseExt: openrtb_ext.ExtBidResponse{
					Errors: nil,
				},
			},
			nbr: models.GetNonBidStatusCodePtr(openrtb3.NoBidGeneral),
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
			nbr: models.GetNonBidStatusCodePtr(openrtb3.NoBidGeneralError),
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
			nbr: models.GetNonBidStatusCodePtr(openrtb3.NoBidGeneralError),
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
			nbr: models.GetNonBidStatusCodePtr(openrtb3.NoBidTimeoutError),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nbr := getNonBRCodeFromBidRespExt(tt.args.bidder, tt.args.bidResponseExt)
			assert.Equal(t, tt.nbr, nbr, tt.name)
		})
	}
}
