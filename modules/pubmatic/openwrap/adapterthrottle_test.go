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
		name  string
		args  args
		want  map[string]struct{}
		want1 bool
	}{
		{
			name: "All partner Seever side disabled",
			args: args{
				partnerConfigMap: map[int]map[string]string{
					1: {models.SERVER_SIDE_FLAG: "0", models.BidderCode: "pubm"},
					2: {models.SERVER_SIDE_FLAG: "0", models.BidderCode: "apnx"},
				},
			},
			want:  map[string]struct{}{},
			want1: true,
		},
		{
			name: "One Partner throttled",
			args: args{
				partnerConfigMap: map[int]map[string]string{
					1: {models.SERVER_SIDE_FLAG: "1", models.BidderCode: "pubm", models.THROTTLE: "40"},
					2: {models.SERVER_SIDE_FLAG: "1", models.BidderCode: "apnx", models.THROTTLE: "60"},
				},
			},
			want:  map[string]struct{}{"pubm": {}},
			want1: false,
		},
		{
			name: "All Partner throttled",
			args: args{
				partnerConfigMap: map[int]map[string]string{
					1: {models.SERVER_SIDE_FLAG: "1", models.BidderCode: "pubm", models.THROTTLE: "40"},
					2: {models.SERVER_SIDE_FLAG: "1", models.BidderCode: "apnx", models.THROTTLE: "0"},
				},
			},
			want:  map[string]struct{}{"pubm": {}, "apnx": {}},
			want1: true,
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
			want:  map[string]struct{}{},
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetAdapterThrottleMap(tt.args.partnerConfigMap)
			assert.Equal(t, got, tt.want, tt.name)
			assert.Equal(t, got1, tt.want1, tt.name)
		})
	}
}
