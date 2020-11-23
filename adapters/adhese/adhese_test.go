package adhese

import (
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters/adapterstest"
)

func TestJsonSamples(t *testing.T) {
	adapterstest.RunJSONBidderTest(t, "adhesetest", NewAdheseBidder("https://ads-{{.AccountID}}.adhese.com/json"))
}
