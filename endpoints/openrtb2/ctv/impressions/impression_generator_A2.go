// Package impressions provides various algorithms to get the number of impressions
// along with minimum and maximum duration of each impression.
// It uses Ad pod request for it
package impressions

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/PubMatic-OpenWrap/prebid-server/endpoints/openrtb2/ctv"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

// keyDelim used as separator in forming key of maxExpectedDurationMap
var keyDelim = ","

type config struct {
	IImpressions
	generator []generator
	// maxExpectedDurationMap contains key = min , max duration, value = 0 -no of impressions, 1
	// this map avoids the unwanted repeatations of impressions generated
	//   Example,
	//   Step 1 : {{2, 17}, {15, 15}, {15, 15}, {10, 10}, {10, 10}, {10, 10}}
	//   Step 2 : {{2, 17}, {15, 15}, {15, 15}, {10, 10}, {10, 10}, {10, 10}}
	//   Step 3 : {{25, 25}, {25, 25}, {2, 22}, {5, 5}}
	//   Step 4 : {{10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}, {10, 10}}
	//   Step 5 : {{15, 15}, {15, 15}, {15, 15}, {15, 15}}
	//   Optimized Output : {{2, 17}, {15, 15},{15, 15},{15, 15},{15, 15},{10, 10},{10, 10},{10, 10},{10, 10},{10, 10},{10, 10},{25, 25}, {25, 25},{2, 22}, {5, 5}}
	//   This map will contains : {2, 17} = 1, {15, 15} = 4, {10, 10} = 6, {25, 25} = 2, {2, 22} = 1, {5, 5} =1
	maxExpectedDurationMap map[string][2]*int
}

// newMinMaxAlgorithm constructs instance of MinMaxAlgorithm
// It computes durations for Ad Slot and Ad Pod in multiple of X
// it also considers minimum configurations present in the request
func newMinMaxAlgorithm(podMinDuration, podMaxDuration int64, p openrtb_ext.VideoAdPod) config {
	generator := make([]generator, 0)
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

func initGenerator(podMinDuration, podMaxDuration int64, p openrtb_ext.VideoAdPod, minAds, maxAds int) generator {
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
	var wg sync.WaitGroup // ensures each step generating impressions is finished
	impsChan := make(chan [][2]int64, len(c.generator))
	for i := 0; i < len(c.generator); i++ {
		wg.Add(1)
		go get(c.generator[i], impsChan, func() { wg.Done() })
	}

	// ensure impressions channel is closed
	// when all go routines are executed
	func() {
		defer close(impsChan)
		wg.Wait()
	}()

	c.maxExpectedDurationMap = make(map[string][2]*int, 0)
	for impressions := range impsChan {
		for index, impression := range impressions {
			impKey := getKey(impression)
			optimizeRepeatations(c, impKey, index+1 == len(impressions))
		}
	}

	// for impressions array
	for impKey := range c.maxExpectedDurationMap {
		for i := 1; i <= c.getRepeations(impKey); i++ {
			imps = append(imps, getImpression(impKey))
		}
	}
	return imps
}

// getImpression constructs the impression object with min and max duration
// from input impression key
func getImpression(key string) [2]int64 {
	decodedKey := strings.Split(key, keyDelim)
	minDuration, _ := strconv.Atoi(decodedKey[0])
	maxDuration, _ := strconv.Atoi(decodedKey[1])
	return [2]int64{int64(minDuration), int64(maxDuration)}
}

// optimizeRepeatations avoids unwanted repeatations of impression object. Using following logic
// maxExpectedDurationMap value contains 2 types of storage
//  1. value[0] - represents main storage where final repeataions are stored
//  2. value[1] - locate storage used by each impression object to add more repeatations if required
// impKey - key used to obtained already added repeatations for given impression
// updateMainStorage - if true and if current local storage value  > repeatations then repeations will be
// updated as main storage
func optimizeRepeatations(c *config, impKey string, updateMainStorage bool) {
	// update temporary storage of each impression
	value := c.maxExpectedDurationMap[impKey]
	if nil == value[0] && nil == value[1] {
		value[0] = new(int)
		value[1] = new(int)
		c.maxExpectedDurationMap[impKey] = value
	}

	*value[1]++ // update tempoary storage
	// if val(temporary storage)  > actual store then consider temporary value as actual value
	if updateMainStorage {
		for k := range c.maxExpectedDurationMap {
			val := c.maxExpectedDurationMap[k]
			if *val[1] > *val[0] {
				*val[0] = *val[1]
			}
			// clear temporary storage
			*val[1] = 0
		}
	}

}

// getKey returns the key used for refering values of maxExpectedDurationMap
// key is computed based on input impression object having min and max durations
func getKey(impression [2]int64) string {
	return fmt.Sprintf("%v%v%v", impression[0], keyDelim, impression[1])
}

// getRepeations returns number of repeatations at that time that this algorithm will
// return w.r.t. input impressionKey
func (c config) getRepeations(impressionKey string) int {
	return *c.maxExpectedDurationMap[impressionKey][0]
}

// get is internal function that actually computes the number of impressions
// based on configrations present in c
func get(c generator, ch chan [][2]int64, onExit func()) {
	defer onExit()
	imps := c.Get()
	ctv.Logf("A2 Impressions = %v\n", imps)
	ch <- imps
}

// Algorithm returns MinMaxAlgorithm
func (c config) Algorithm() Algorithm {
	return MinMaxAlgorithm
}
