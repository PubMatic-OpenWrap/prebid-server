package tracker

import (
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func TestInjectTrackers(t *testing.T) {
	type args struct {
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
	}
	tests := []struct {
		name    string
		args    args
		want    *openrtb2.BidResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InjectTrackers(tt.args.rctx, tt.args.bidResponse)
			if (err != nil) != tt.wantErr {
				t.Errorf("InjectTrackers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InjectTrackers() = %v, want %v", got, tt.want)
			}
		})
	}
}
