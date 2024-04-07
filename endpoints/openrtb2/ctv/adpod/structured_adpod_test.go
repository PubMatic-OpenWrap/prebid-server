package adpod

import "testing"

func TestStructuredAdpodPerformAuctionAndExclusion(t *testing.T) {
	type fields struct {
		AdpodCtx AdpodCtx
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Perform auction when all bids are ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := &StructuredAdpod{
				AdpodCtx: tt.fields.AdpodCtx,
			}
			sa.PerformAuctionAndExclusion()
		})
	}
}
