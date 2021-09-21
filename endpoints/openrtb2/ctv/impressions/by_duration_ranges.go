package impressions

import "github.com/prebid/prebid-server/endpoints/openrtb2/ctv/util"

type byDurRangeConfig struct {
	IImpressions
	durationRangeInSec []int
	podMinDuration     int
	podMaxDuration     int
}

func newByDurationRanges(durationRangeInSec []int, podMinDuration, podMaxDuration int) byDurRangeConfig {
	return byDurRangeConfig{
		durationRangeInSec: durationRangeInSec,
		podMinDuration:     podMinDuration,
		podMaxDuration:     podMaxDuration,
	}
}

func (c *byDurRangeConfig) Get() [][2]int64 {
	if len(c.durationRangeInSec) == 0 {
		util.Logf("durationRangeInSec is nil. [%v] algorithm returning not generated impressions", c.Algorithm())
		return make([][2]int64, 0)
	}
	imps := make([][2]int64, 0)
	for _, dur := range c.durationRangeInSec {
		if dur < c.podMinDuration || dur > c.podMaxDuration {
			continue // invalid duration
		}
		imps = append(imps, [2]int64{int64(dur), int64(dur)})
	}
	return imps
}

// Algorithm returns MinMaxAlgorithm
func (c *byDurRangeConfig) Algorithm() Algorithm {
	return ByDurationRanges
}
