package floors

import (
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func getFalse() *bool {
	b := false
	return &b
}

func getTrue() *bool {
	b := true
	return &b
}

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
func TestShouldEnforceFloors(t *testing.T) {
	type args struct {
		floorExt          *openrtb_ext.PriceFloorRules
		configEnforceRate int
		f                 func(int) int
	}
	tests := []struct {
		name            string
		args            args
		expEnforce      bool
		expReqExtUpdate bool
	}{
		{
			name: "enfocement = true of enforcement object not provided",
			args: args{
				configEnforceRate: 100,
				f: func(n int) int {
					return n - 1
				},
			},
			expEnforce:      true,
			expReqExtUpdate: true,
		},

		{
			name: "No enfocement of floors when enforcePBS is false",
			args: args{
				floorExt: &openrtb_ext.PriceFloorRules{
					Enforcement: &openrtb_ext.PriceFloorEnforcement{
						EnforcePBS: getFalse(),
					},
					Skipped: getFalse(),
				},
				configEnforceRate: 10,
				f: func(n int) int {
					return n
				},
			},
			expEnforce:      false,
			expReqExtUpdate: false,
		},
		{
			name: "No enfocement of floors when enforcePBS is true but enforce rate is low",
			args: args{
				floorExt: &openrtb_ext.PriceFloorRules{
					Enforcement: &openrtb_ext.PriceFloorEnforcement{
						EnforcePBS: getTrue(),
					},
					Skipped: getFalse(),
				},
				configEnforceRate: 10,
				f: func(n int) int {
					return n
				},
			},
			expEnforce:      false,
			expReqExtUpdate: true,
		},
		{
			name: "No enfocement of floors when enforcePBS is true but enforce rate is low in incoming request",
			args: args{
				floorExt: &openrtb_ext.PriceFloorRules{
					Enforcement: &openrtb_ext.PriceFloorEnforcement{
						EnforcePBS:  getTrue(),
						EnforceRate: 10,
					},
					Skipped: getFalse(),
				},
				configEnforceRate: 100,
				f: func(n int) int {
					return n
				},
			},
			expEnforce:      false,
			expReqExtUpdate: true,
		},
		{
			name: "No Enfocement of floors when skipped is true, non zero value of bidfloor in imp",
			args: args{
				floorExt: &openrtb_ext.PriceFloorRules{
					Enforcement: &openrtb_ext.PriceFloorEnforcement{
						EnforcePBS: getTrue(),
					},
					Skipped: getTrue(),
				},
				configEnforceRate: 98,
				f: func(n int) int {
					return n - 5
				},
			},
			expEnforce:      false,
			expReqExtUpdate: false,
		},
		{
			name: "No enfocement of floors when skipped is true, zero value of bidfloor in imp",
			args: args{
				floorExt: &openrtb_ext.PriceFloorRules{
					Enforcement: &openrtb_ext.PriceFloorEnforcement{
						EnforcePBS: getTrue(),
					},
					Skipped: getTrue(),
				},
				configEnforceRate: 98,
				f: func(n int) int {
					return n - 5
				},
			},
			expEnforce:      false,
			expReqExtUpdate: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldEnforce, updateReq := ShouldEnforce(tt.args.floorExt, tt.args.configEnforceRate, tt.args.f)
			if shouldEnforce != tt.expEnforce {
				t.Errorf("shouldEnforce = %v, want %v", shouldEnforce, tt.expEnforce)
			}

			if updateReq != tt.expReqExtUpdate {
				t.Errorf("expReqExtUpdate  %v, want %v", updateReq, tt.expReqExtUpdate)
			}
		})
	}
}
