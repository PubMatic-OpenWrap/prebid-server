package mobfoxpb

import (
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters/adapterstest"
	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

func TestJsonSamples(t *testing.T) {
	bidder, buildErr := Builder(openrtb_ext.BidderMobfoxpb, config.Adapter{
		Endpoint: "http://example.com/?c=__route__&m=__method__&key=__key__"})
	if buildErr != nil {
		t.Fatalf("Builder returned unexpected error %v", buildErr)
	}
	adapterstest.RunJSONBidderTest(t, "mobfoxpbtest", bidder)
}
