package openrtb_ext

import (
	"fmt"
	"testing"
)

func TestSyncRTBBidders(t *testing.T) {

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SyncRTBBidders(); (err != nil) != tt.wantErr {
				t.Errorf("FetchRTBBidders() error = %v, wantErr %v", err, tt.wantErr)
			}
			fmt.Println(CoreBidderNames())
		})
	}
}
