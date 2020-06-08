package ctv

import (
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

type config struct {
	IImpressions
	generator []adPodConfig
}

func newImpGenA2(podMinDuration, podMaxDuration int64, p openrtb_ext.VideoAdPod) config {
	generator := make([]adPodConfig, 0)
	// step 1
	generator = append(generator, newImpGenA1(podMinDuration, podMaxDuration, p))
	// step 2 - pod duration = pod max, no of ads = max ads
	generator = append(generator, newImpGenA1(podMaxDuration, podMaxDuration, newVideoAdPod(p, *p.MaxAds, *p.MaxAds)))
	// step 3 - pod duration = pod max, no of ads = min ads
	generator = append(generator, newImpGenA1(podMaxDuration, podMaxDuration, newVideoAdPod(p, *p.MinAds, *p.MinAds)))
	// step 4 - pod duration = pod min, no of ads = max  ads
	generator = append(generator, newImpGenA1(podMinDuration, podMinDuration, newVideoAdPod(p, *p.MaxAds, *p.MaxAds)))
	// step 5 - pod duration = pod min, no of ads = min  ads
	generator = append(generator, newImpGenA1(podMinDuration, podMinDuration, newVideoAdPod(p, *p.MinAds, *p.MinAds)))

	return config{generator: generator}
}

func newVideoAdPod(p openrtb_ext.VideoAdPod, minAds, maxAds int) openrtb_ext.VideoAdPod {
	return openrtb_ext.VideoAdPod{MinDuration: p.MinDuration,
		MaxDuration: p.MaxDuration,
		MinAds:      &minAds,
		MaxAds:      &maxAds}
}

// Get ...
func (c config) Get() [][2]int64 {
	imps := make([][2]int64, 0)
	impsChan := make(chan [][2]int64, len(c.generator))
	for i := 0; i < len(c.generator); i++ {
		go get(c.generator[i], impsChan)
		imps = append(imps, <-impsChan...)
	}
	return imps
}

func get(cfg adPodConfig, c chan [][2]int64) {
	imps := cfg.Get()
	print("Impressions = %v\n", imps)
	c <- imps
}
