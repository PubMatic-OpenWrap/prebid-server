package openwrap

import (
	"reflect"
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func Test_getTestModePartnerConfigMap(t *testing.T) {
	type args struct {
		platform       string
		timeout        int64
		displayVersion int
	}
	tests := []struct {
		name string
		args args
		want map[int]map[string]string
	}{
		{
			name: "get_test_mode_partnerConfigMap",
			args: args{
				platform:       "in-app",
				timeout:        200,
				displayVersion: 2,
			},
			want: map[int]map[string]string{
				1: {
					models.PARTNER_ID:          "1",
					models.PREBID_PARTNER_NAME: "pubmatic",
					models.BidderCode:          "pubmatic",
					models.SERVER_SIDE_FLAG:    "1",
					models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
					models.TIMEOUT:             "200",
				},
				-1: {
					models.PLATFORM_KEY:     "in-app",
					models.DisplayVersionID: "2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTestModePartnerConfigMap(tt.args.platform, tt.args.timeout, tt.args.displayVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected= %v, but got= %v", tt.want, got)
			}
		})
	}
}
