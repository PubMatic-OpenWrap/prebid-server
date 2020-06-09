// Package impressions provides various algorithms to get the number of impressions
// along with minimum and maximum duration of each impression.
// It uses Ad pod request for it
package impressions

import (
	"fmt"
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

func TestGetImpressionsA2(t *testing.T) {
	p := openrtb_ext.VideoAdPod{}
	p.MinDuration = new(int)
	*p.MinDuration = 20
	p.MaxDuration = new(int)
	*p.MaxDuration = 45
	p.MinAds = new(int)
	*p.MinAds = 2
	p.MaxAds = new(int)
	*p.MaxAds = 10

	gen := newImpGenA2(60, 90, p)
	fmt.Println(gen.Get())
	fmt.Println(gen.Algorithm())
}

func BenchmarkGetImpressionsA2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := openrtb_ext.VideoAdPod{}
		p.MinDuration = new(int)
		*p.MinDuration = 20
		p.MaxDuration = new(int)
		*p.MaxDuration = 45
		p.MinAds = new(int)
		*p.MinAds = 2
		p.MaxAds = new(int)
		*p.MaxAds = 10

		newImpGenA2(60, 90, p)
	}
}
