package openwrap

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func TestGetAdapterThrottleMap(t *testing.T) {
	original := GetRandomNumberBelow100
	GetRandomNumberBelow100 = func() int {
		return 50
	}
	defer func() {
		GetRandomNumberBelow100 = original
	}()
	type args struct {
		partnerConfigMap map[int]map[string]string
	}
	tests := []struct {
		name                   string
		args                   args
		wantAdapterThrottled   map[string]struct{}
		wantAllAdapterThrolled bool
	}{
		{
			name: "All partner Seever side disabled",
			args: args{
				partnerConfigMap: map[int]map[string]string{
					1: {models.SERVER_SIDE_FLAG: "0", models.BidderCode: "pubm"},
					2: {models.SERVER_SIDE_FLAG: "0", models.BidderCode: "apnx"},
				},
			},
			wantAdapterThrottled:   map[string]struct{}{},
			wantAllAdapterThrolled: true,
		},
		{
			name: "One Partner throttled",
			args: args{
				partnerConfigMap: map[int]map[string]string{
					1: {models.SERVER_SIDE_FLAG: "1", models.BidderCode: "pubm", models.THROTTLE: "40"},
					2: {models.SERVER_SIDE_FLAG: "1", models.BidderCode: "apnx", models.THROTTLE: "60"},
				},
			},
			wantAdapterThrottled:   map[string]struct{}{"pubm": {}},
			wantAllAdapterThrolled: false,
		},
		{
			name: "All Partner throttled",
			args: args{
				partnerConfigMap: map[int]map[string]string{
					1: {models.SERVER_SIDE_FLAG: "1", models.BidderCode: "pubm", models.THROTTLE: "40"},
					2: {models.SERVER_SIDE_FLAG: "1", models.BidderCode: "apnx", models.THROTTLE: "0"},
				},
			},
			wantAdapterThrottled:   map[string]struct{}{"pubm": {}, "apnx": {}},
			wantAllAdapterThrolled: true,
		},
		{
			name: "No Partner throttled",
			args: args{
				partnerConfigMap: map[int]map[string]string{
					1: {models.SERVER_SIDE_FLAG: "1", models.BidderCode: "pubm", models.THROTTLE: "60"},
					2: {models.SERVER_SIDE_FLAG: "1", models.BidderCode: "apnx", models.THROTTLE: "100"},
					3: {models.SERVER_SIDE_FLAG: "1", models.BidderCode: "openx", models.THROTTLE: ""},
				},
			},
			wantAdapterThrottled:   map[string]struct{}{},
			wantAllAdapterThrolled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdapterThrottled, gotAllAdapterThrolled := GetAdapterThrottleMap(tt.args.partnerConfigMap)
			assert.Equal(t, gotAdapterThrottled, tt.wantAdapterThrottled, tt.name)
			assert.Equal(t, gotAllAdapterThrolled, tt.wantAllAdapterThrolled, tt.name)
		})
	}
}
