package openrtb_ext

import "testing"

func TestFetchRTBBidders(t *testing.T) {

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := FetchRTBBidders(); (err != nil) != tt.wantErr {
				t.Errorf("FetchRTBBidders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
