package openwrap

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestAddInAppTargettingKeys(t *testing.T) {
	type args struct {
		targeting        map[string]string
		seat             string
		ecpm             float64
		bid              *openrtb2.Bid
		isWinningBid     bool
		priceGranularity *openrtb_ext.PriceGranularity
	}
	tests := []struct {
		name          string
		args          args
		wantTargeting map[string]string
	}{
		{
			name: "winning bid with no price granularity",
			args: args{
				bid: &openrtb2.Bid{
					ID:     "12345",
					Price:  2.342,
					W:      250,
					H:      300,
					DealID: "1",
				},
				seat:         "pubmatic",
				isWinningBid: true,
				ecpm:         2.34,
				targeting:    map[string]string{},
			},
			wantTargeting: map[string]string{
				"pwtecp_pubmatic": "2.34",
				"pwtplt_pubmatic": "inapp",
				"pwtsz":           "250x300",
				"pwtecp":          "2.34",
				"pwtdid_pubmatic": "1",
				"pwtplt":          "inapp",
				"pwtbst_pubmatic": "1",
				"pwtsid":          "12345",
				"pwtbst":          "1",
				"pwtsid_pubmatic": "12345",
				"pwtsz_pubmatic":  "250x300",
				"pwtpid_pubmatic": "pubmatic",
				"pwtpid":          "pubmatic",
				"pwtdid":          "1",
			},
		},
		{
			name: "winning bid",
			args: args{
				bid: &openrtb2.Bid{
					ID:     "12345",
					Price:  2.342,
					W:      250,
					H:      300,
					DealID: "1",
				},
				seat:             "pubmatic",
				priceGranularity: &priceGranularityAuto,
				isWinningBid:     true,
				ecpm:             2.34,
				targeting:        map[string]string{},
			},
			wantTargeting: map[string]string{
				"pwtecp_pubmatic": "2.34",
				"pwtplt_pubmatic": "inapp",
				"pwtpb_pubmatic":  "2.30",
				"pwtsz":           "250x300",
				"pwtecp":          "2.34",
				"pwtdid_pubmatic": "1",
				"pwtplt":          "inapp",
				"pwtbst_pubmatic": "1",
				"pwtsid":          "12345",
				"pwtbst":          "1",
				"pwtsid_pubmatic": "12345",
				"pwtsz_pubmatic":  "250x300",
				"pwtpid_pubmatic": "pubmatic",
				"pwtpid":          "pubmatic",
				"pwtdid":          "1",
				"pwtpb":           "2.30",
			},
		},
		{
			name: "non winning bid",
			args: args{
				bid: &openrtb2.Bid{
					ID:     "12345",
					Price:  2.342,
					W:      250,
					H:      300,
					DealID: "1",
				},
				seat:             "pubmatic",
				priceGranularity: &priceGranularityAuto,
				isWinningBid:     false,
				ecpm:             2.34,
				targeting:        map[string]string{},
			},
			wantTargeting: map[string]string{
				"pwtecp_pubmatic": "2.34",
				"pwtplt_pubmatic": "inapp",
				"pwtpb_pubmatic":  "2.30",
				"pwtdid_pubmatic": "1",
				"pwtbst_pubmatic": "1",
				"pwtsid_pubmatic": "12345",
				"pwtsz_pubmatic":  "250x300",
				"pwtpid_pubmatic": "pubmatic",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addInAppTargettingKeys(tt.args.targeting, tt.args.seat, tt.args.ecpm, tt.args.bid, tt.args.isWinningBid, tt.args.priceGranularity)
			assert.Equal(t, tt.wantTargeting, tt.args.targeting, tt.name)
		})
	}
}
