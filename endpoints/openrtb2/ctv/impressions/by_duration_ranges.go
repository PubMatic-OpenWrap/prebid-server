package impressions

import "github.com/prebid/prebid-server/endpoints/openrtb2/ctv/util"

// byDurRangeConfig struct will be used for creating impressions object based on list of duration ranges
type byDurRangeConfig struct {
	IImpressions         //IImpressions interface
	durations      []int //durations list of durations in seconds used for creating impressions object
	maxAds         int   //maxAds is number of max impressions can be created
	podMaxDuration int   //podMaxDuration, element in durations must be greater than podMaxDuration(boundry check)
}

// newByDurationRanges will create new object ob byDurRangeConfig for creating impressions for adpod request
func newByDurationRanges(durations []int, maxAds, podMaxDuration int) byDurRangeConfig {
	return byDurRangeConfig{
		durations:      durations,
		maxAds:         maxAds,
		podMaxDuration: podMaxDuration,
	}
}

// Get function returns lists of min,max duration ranges ganerated based on durations
// it will return valid durations, duration must be within podMinDuration and podMaxDuration range
// if len(durations) < maxAds then clone valid durations from starting till we reach maxAds length
func (c *byDurRangeConfig) Get() [][2]int64 {
	if len(c.durations) == 0 {
		util.Logf("durations is nil. [%v] algorithm returning not generated impressions", c.Algorithm())
		return make([][2]int64, 0)
	}
	imps := make([][2]int64, 0)
	for _, dur := range c.durations {
		if dur > c.podMaxDuration {
			continue // invalid duration
		}
		imps = append(imps, [2]int64{int64(dur), int64(dur)})
	}

	//adding extra impressions incase of total impressions generated are less than pod max ads.
	for i := 0; len(imps) < c.maxAds; i++ {
		imps = append(imps, [2]int64{imps[i][0], imps[i][1]})
	}

	return imps
}

// Algorithm returns MinMaxAlgorithm
func (c *byDurRangeConfig) Algorithm() Algorithm {
	return ByDurationRanges
}
