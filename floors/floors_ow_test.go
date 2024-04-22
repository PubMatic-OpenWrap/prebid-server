package floors

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
)

func TestRequestHasFloors(t *testing.T) {

	tests := []struct {
		name       string
		bidRequest *openrtb2.BidRequest
		want       bool
	}{
		{
			bidRequest: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
			},
			want: false,
		},
		{
			bidRequest: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp: []openrtb2.Imp{{ID: "1234", BidFloor: 10, BidFloorCur: "USD", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RequestHasFloors(tt.bidRequest); got != tt.want {
				t.Errorf("RequestHasFloors() = %v, want %v", got, tt.want)
			}
		})
	}
}
