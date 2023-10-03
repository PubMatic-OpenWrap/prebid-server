package openwrap

import (
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestGetAdapterThrottleMap(t *testing.T) {
	type args struct {
		partnerConfigMap map[int]map[string]string
	}
	type want struct {
		adapterThrottleMap       map[string]struct{}
		allPartnersThrottledFlag bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "All_prtners_throttled",
			args: args{
				partnerConfigMap: map[int]map[string]string{
					0: {
						models.THROTTLE:            "0",
						models.PREBID_PARTNER_NAME: "pubmatic",
						models.BidderCode:          "pubmatic",
						models.SERVER_SIDE_FLAG:    "0",
					},
					1: {
						models.THROTTLE:            "0",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "",
					},
				},
			},
			want: want{
				adapterThrottleMap:       map[string]struct{}{},
				allPartnersThrottledFlag: true,
			},
		},
		{
			name: "one_prtner_throttled_out_of_two",
			args: args{
				partnerConfigMap: map[int]map[string]string{
					0: {
						models.THROTTLE:            "0",
						models.PREBID_PARTNER_NAME: "pubmatic",
						models.BidderCode:          "pubmatic",
						models.SERVER_SIDE_FLAG:    "1",
					},
					1: {
						models.THROTTLE:            "100",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
					},
				},
			},
			want: want{
				adapterThrottleMap: map[string]struct{}{
					"pubmatic": {},
				},
				allPartnersThrottledFlag: false,
			},
		},
		{
			name: "no_prtner_throttled_out_of_two",
			args: args{
				partnerConfigMap: map[int]map[string]string{
					0: {
						models.THROTTLE:            "100",
						models.PREBID_PARTNER_NAME: "pubmatic",
						models.BidderCode:          "pubmatic",
						models.SERVER_SIDE_FLAG:    "1",
					},
					1: {
						models.THROTTLE:            "100",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
					},
				},
			},
			want: want{
				adapterThrottleMap:       map[string]struct{}{},
				allPartnersThrottledFlag: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapterThrottleMap, allPartnersThrottledFlag := GetAdapterThrottleMap(tt.args.partnerConfigMap)
			assert.Equal(t, tt.want.adapterThrottleMap, adapterThrottleMap)
			if allPartnersThrottledFlag != tt.want.allPartnersThrottledFlag {
				t.Errorf("GetAdapterThrottleMap() got1 = %v, want %v", allPartnersThrottledFlag, tt.want.allPartnersThrottledFlag)
			}
		})
	}
}

func TestThrottleAdapter(t *testing.T) {
	type args struct {
		partnerConfig map[string]string
		val           int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "partner_throtlle_is_100",
			args: args{
				partnerConfig: map[string]string{
					models.THROTTLE: "100",
				},
			},
			want: false,
		},
		{
			name: "partner_throtlle_is_empty",
			args: args{
				partnerConfig: map[string]string{
					models.THROTTLE: "",
				},
			},
			want: false,
		},
		{
			name: "partner_throtlle_is_0",
			args: args{
				partnerConfig: map[string]string{
					models.THROTTLE: "0",
				},
			},
			want: true,
		},
		{
			name: "partner_throtlle_is_greater_than_0_and_less_than_100_and_random_number_generated_is_10",
			args: args{
				partnerConfig: map[string]string{
					models.THROTTLE: "70",
				},
				val: 10,
			},
			want: true,
		},
		{
			name: "partner_throtlle_is_greater_than_0_and_less_than_100_and_random_number_generated_is_30",
			args: args{
				partnerConfig: map[string]string{
					models.THROTTLE: "70",
				},
				val: 30,
			},
			want: false,
		},
		{
			name: "partner_throtlle_is_greater_than_0_and_less_than_100_and_random_number_generated_is_50",
			args: args{
				partnerConfig: map[string]string{
					models.THROTTLE: "70",
				},
				val: 50,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetRandomNumberBelow100 = func() int {
				return tt.args.val
			}
			if got := ThrottleAdapter(tt.args.partnerConfig); got != tt.want {
				t.Errorf("ThrottleAdapter() = %v, want %v", got, tt.want)
			}
		})
	}
}
