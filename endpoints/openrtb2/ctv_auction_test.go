package openrtb2

import (
	"testing"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/stretchr/testify/assert"
)

func TestGetAdDuration(t *testing.T) {
	var tests = []struct {
		scenario      string
		adDuration    string // actual ad duration. 0 value will be assumed as no ad duration
		maxAdDuration int    // requested max ad duration
		expect        int
	}{
		{"0sec ad duration", "0", 200, 200},
		{"40sec ad duration", "30", 100, 30},
		{"negative ad duration", "-30", 100, 100},
		{"invalid ad duration", "invalid", 80, 80},
		{"ad duration breaking bid.Ext json", `""quote""`, 50, 50},
	}
	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			bid := openrtb.Bid{
				Ext: []byte(`{"prebid" : {"video" : {"duration" : ` + test.adDuration + `}}}`),
			}
			assert.Equal(t, test.expect, getAdDuration(bid, int64(test.maxAdDuration)))
		})
	}
}
