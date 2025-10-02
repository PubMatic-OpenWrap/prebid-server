package ctvlegacy

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// keyDelim used as separator in forming key of maxExpectedDurationMap
var keyDelim = ","

type MinMax struct {
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
	maxExpectedDurationMap map[string][2]int
	requested              pod
}

// newMinMaxAlgorithm constructs instance of MinMaxAlgorithm
// It computes durations for Ad Slot and Ad Pod in multiple of X
// it also considers minimum configurations present in the request
func newMinMaxAlgorithm(podMinDuration, podMaxDuration int64, p *models.AdPod) MinMax {
	generator := make([]generator, 0)
	// step 1 - same as Algorithm1
	generator = append(generator, initGenerator(podMinDuration, podMaxDuration, p, p.MinAds, p.MaxAds))
	// step 2 - pod duration = pod max, no of ads = max ads
	generator = append(generator, initGenerator(podMaxDuration, podMaxDuration, p, p.MaxAds, p.MaxAds))
	// step 3 - pod duration = pod max, no of ads = min ads
	generator = append(generator, initGenerator(podMaxDuration, podMaxDuration, p, p.MinAds, p.MinAds))
	// step 4 - pod duration = pod min, no of ads = max  ads
	generator = append(generator, initGenerator(podMinDuration, podMinDuration, p, p.MaxAds, p.MaxAds))
	// step 5 - pod duration = pod min, no of ads = min  ads
	generator = append(generator, initGenerator(podMinDuration, podMinDuration, p, p.MinAds, p.MinAds))

	return MinMax{generator: generator, requested: generator[0].requested}
}

func initGenerator(podMinDuration, podMaxDuration int64, p *models.AdPod, minAds, maxAds int) generator {
	config := newConfigWithMultipleOf(podMinDuration, podMaxDuration, newVideoAdPod(p, minAds, maxAds), multipleOf)
	return config
}

func newVideoAdPod(p *models.AdPod, minAds, maxAds int) *models.AdPod {

	adpod := models.AdPod{MinDuration: p.MinDuration,
		MaxDuration: p.MaxDuration,
		MinAds:      minAds,
		MaxAds:      maxAds}
	return &adpod
}

// Algorithm returns MinMaxAlgorithm
func (mm *MinMax) Algorithm() int {
	return models.MinMaxAlgorithm
}

// Get ...
func (mm *MinMax) Get() [][2]int64 {
	imps := make([][2]int64, 0)
	wg := new(sync.WaitGroup) // ensures each step generating impressions is finished
	impsChan := make(chan [][2]int64, len(mm.generator))
	for i := 0; i < len(mm.generator); i++ {
		wg.Add(1)
		go get(mm.generator[i], impsChan, wg)
	}

	// ensure impressions channel is closed
	// when all go routines are executed
	func() {
		defer close(impsChan)
		wg.Wait()
	}()

	mm.maxExpectedDurationMap = make(map[string][2]int, 0)
	// util.Logf("Step wise breakup ")
	for impressions := range impsChan {
		for index, impression := range impressions {
			impKey := getKey(impression)
			setMaximumRepeatations(mm, impKey, index+1 == len(impressions))
		}
		// util.Logf("%v", impressions)
	}

	// for impressions array
	indexOffset := 0
	for impKey := range mm.maxExpectedDurationMap {
		totalRepeations := mm.getRepeations(impKey)
		for repeation := 1; repeation <= totalRepeations; repeation++ {
			imps = append(imps, getImpression(impKey))
		}
		// if exact pod duration is provided then do not compute
		// min duration. Instead expect min duration same as max duration
		// It must be set by underneath algorithm
		if mm.requested.podMinDuration != mm.requested.podMaxDuration {
			computeMinDuration(*mm, imps[:], indexOffset, indexOffset+totalRepeations)
		}
		indexOffset += totalRepeations
	}
	return imps
}

// getRepeations returns number of repeatations at that time that this algorithm will
// return w.r.t. input impressionKey
func (mm MinMax) getRepeations(impressionKey string) int {
	return mm.maxExpectedDurationMap[impressionKey][0]
}

// get is internal function that actually computes the number of impressions
// based on configrations present in c
func get(c generator, ch chan [][2]int64, wg *sync.WaitGroup) {
	defer wg.Done()
	imps := c.Get()
	// util.Logf("A2 Impressions = %v\n", imps)
	ch <- imps
}

// getKey returns the key used for refering values of maxExpectedDurationMap
// key is computed based on input impression object having min and max durations
func getKey(impression [2]int64) string {
	return fmt.Sprintf("%v%v%v", impression[models.MinDuration], keyDelim, impression[models.MaxDuration])
}

// setMaximumRepeatations avoids unwanted repeatations of impression object. Using following logic
// maxExpectedDurationMap value contains 2 types of storage
//  1. value[0] - represents current counter where final repeataions are stored
//  2. value[1] - local storage used by each impression object to add more repeatations if required
//
// impKey - key used to obtained already added repeatations for given impression
// updateCurrentCounter - if true and if current local storage value  > repeatations then repeations will be
// updated as current counter
func setMaximumRepeatations(mm *MinMax, impKey string, updateCurrentCounter bool) {
	// update maxCounter of each impression
	value := mm.maxExpectedDurationMap[impKey]
	value[1]++ // increment max counter (contains no of repeatations for given iteration)
	mm.maxExpectedDurationMap[impKey] = value
	// if val(maxCounter)  > actual store then consider temporary value as actual value
	if updateCurrentCounter {
		for k := range mm.maxExpectedDurationMap {
			val := mm.maxExpectedDurationMap[k]
			if val[1] > val[0] {
				val[0] = val[1]
			}
			// clear maxCounter
			val[1] = 0
			mm.maxExpectedDurationMap[k] = val // reassign
		}
	}

}

// getImpression constructs the impression object with min and max duration
// from input impression key
func getImpression(key string) [2]int64 {
	decodedKey := strings.Split(key, keyDelim)
	minDuration, _ := strconv.Atoi(decodedKey[models.MinDuration])
	maxDuration, _ := strconv.Atoi(decodedKey[models.MaxDuration])
	return [2]int64{int64(minDuration), int64(maxDuration)}
}

func computeMinDuration(mm MinMax, impressions [][2]int64, start int, end int) {
	r := mm.requested
	// 5/2 => q = 2 , r = 1 =>  2.5 => 3
	minDuration := int64(math.Round(float64(r.podMinDuration) / float64(r.minAds)))
	for i := start; i < end; i++ {
		impression := &impressions[i]
		// ensure imp duration boundaries
		// if boundaries are not honoured keep min duration which is computed as is
		if minDuration >= r.slotMinDuration && minDuration <= impression[models.MaxDuration] {
			// override previous value
			impression[models.MinDuration] = minDuration
		} else {
			// boundaries are not matching keep min value as is
			// util.Logf("False : minDuration (%v) >= r.slotMinDuration (%v)  &&  minDuration (%v)  <= impression[MaxDuration] (%v)", minDuration, r.slotMinDuration, minDuration, impression[MaxDuration])
			// util.Logf("Hence, setting request level slot minduration (%v) ", r.slotMinDuration)
			impression[models.MinDuration] = r.slotMinDuration
		}
	}
}
