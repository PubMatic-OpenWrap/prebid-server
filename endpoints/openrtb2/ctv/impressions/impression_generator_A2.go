// Package impressions provides various algorithms to get the number of impressions
// along with minimum and maximum duration of each impression.
// It uses Ad pod request for it
package impressions

import (
	"github.com/PubMatic-OpenWrap/prebid-server/endpoints/openrtb2/ctv"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

type config struct {
	IImpressions
	generator []adPodConfig
	// maxExpectedDurationMap contains key = min , max duration, value = no of impressions
	// this map avoids the unwanted repeatations of impressions generated
	//   Example,
	//   Step 1 : {{2, 17}, {15, 15}, {15, 15}, {10, 10}, {10, 10}, {10, 10}}
	//   Step 2 : {{2, 17}, {15, 15}, {15, 15}, {10, 10}, {10, 10}, {10, 10}}
	//   Step 3 : {{25, 25}, {25, 25}, {2, 22}, {5, 5}}
	//   Step 4 : {{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}}
	//   Step 5 : {{15, 15}, {15, 15}, {15, 15}, {15, 15}}
	//   Optimized Output : {{2, 17}, {15, 15},{15, 15},{15, 15},{15, 15},{10, 10},{10, 10},{10, 10},{10, 10},{10, 10},{10, 10},{25, 25}, {25, 25},{2, 22}, {5, 5}}
	//   This map will contains : {2, 17} = 1, {15, 15} = 4, {10, 10} = 6, {25, 25} = 2, {2, 22} = 1, {5, 5} =1
	maxExpectedDurationMap map[[2]int]int
}

func newImpGenA2(podMinDuration, podMaxDuration int64, p openrtb_ext.VideoAdPod) config {
	generator := make([]adPodConfig, 0)
	// step 1 - same as Algorithm1
	generator = append(generator, initGenerator(podMinDuration, podMaxDuration, p, *p.MinAds, *p.MaxAds))
	// step 2 - pod duration = pod max, no of ads = max ads
	generator = append(generator, initGenerator(podMaxDuration, podMaxDuration, p, *p.MaxAds, *p.MaxAds))
	// step 3 - pod duration = pod max, no of ads = min ads
	generator = append(generator, initGenerator(podMaxDuration, podMaxDuration, p, *p.MinAds, *p.MinAds))
	// step 4 - pod duration = pod min, no of ads = max  ads
	generator = append(generator, initGenerator(podMinDuration, podMinDuration, p, *p.MaxAds, *p.MaxAds))
	// step 5 - pod duration = pod min, no of ads = min  ads
	generator = append(generator, initGenerator(podMinDuration, podMinDuration, p, *p.MinAds, *p.MinAds))

	return config{generator: generator}
}

func initGenerator(podMinDuration, podMaxDuration int64, p openrtb_ext.VideoAdPod, minAds, maxAds int) adPodConfig {
	config := newConfigWithMultipleOf(podMinDuration, podMaxDuration, newVideoAdPod(p, minAds, maxAds), multipleOf)
	return config
}

func newVideoAdPod(p openrtb_ext.VideoAdPod, minAds, maxAds int) openrtb_ext.VideoAdPod {
	return openrtb_ext.VideoAdPod{MinDuration: p.MinDuration,
		MaxDuration: p.MaxDuration,
		MinAds:      &minAds,
		MaxAds:      &maxAds}
}

// Get ...
func (c *config) Get() [][2]int64 {
	imps := make([][2]int64, 0)
	impsChan := make(chan [][2]int64, len(c.generator))
	for i := 0; i < len(c.generator); i++ {
		go get(c.generator[i], impsChan)
		imps = append(imps, <-impsChan...)
	}
	return imps
}

func get(c adPodConfig, ch chan [][2]int64) {
	imps := c.Get()
	ctv.Logf("Impressions = %v\n", imps)
	ch <- imps
}

// Algorithm returns Algorithm2
func (c config) Algorithm() int {
	return Algorithm2
}
