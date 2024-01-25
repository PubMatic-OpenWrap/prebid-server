package impressions

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// Value use to compute Ad Slot Durations and Pod Durations for internal computation
// Right now this value is set to 5, based on passed data observations
// Observed that typically video impression contains contains minimum and maximum duration in multiples of  5
const (
	multipleOf            = 5
	impressionIDSeparator = `_`
	impressionIDFormat    = `%v` + impressionIDSeparator + `%v`
)

// ImpGenerator ...
type ImpGenerator interface {
	Get() [][2]int64
	// Algorithm() int // returns algorithm used for computing number of impressions
}

func GenerateImpressions(request *openrtb_ext.RequestWrapper, impCtx map[string]models.ImpCtx, pubId string, me metrics.MetricsEngine) ([]*openrtb_ext.ImpWrapper, []error) {
	var imps []*openrtb_ext.ImpWrapper
	var errs []error

	for _, impWrapper := range request.GetImp() {
		eachImpCtx := impCtx[impWrapper.ID]

		impAdpodConfig, err := getAdPodImpConfig(impWrapper.Imp, eachImpCtx.AdpodConfig)
		if impAdpodConfig == nil {
			imps = append(imps, impWrapper)
			if err != nil {
				errs = append(errs, err)
			}
			continue
		}

		me.RecordAdPodGeneratedImpressionsCount(len(impAdpodConfig), pubId)
		eachImpCtx.ImpAdPodCfg = impAdpodConfig
		impCtx[impWrapper.ID] = eachImpCtx

		err = impWrapper.RebuildImpressionExt()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		for i := range impAdpodConfig {
			video := *impWrapper.Video
			video.MinDuration = impAdpodConfig[i].MinDuration
			video.MaxDuration = impAdpodConfig[i].MaxDuration
			video.Sequence = impAdpodConfig[i].SequenceNumber
			video.MaxExtended = 0

			// Remove adpod Extension
			var videoExt map[string]interface{}
			err := json.Unmarshal(video.Ext, &videoExt)
			if err != nil {
				glog.Warningf("error while unmarshalling video extension for impression: %s", impAdpodConfig[i].ImpID)
			}
			delete(videoExt, "adpod")
			delete(videoExt, "offset")
			if len(videoExt) == 0 {
				video.Ext = nil
			} else {
				video.Ext, _ = json.Marshal(videoExt)
			}

			newImp := *impWrapper.Imp
			newImp.ID = impAdpodConfig[i].ImpID
			newImp.Video = &video

			newImpWrapper := &openrtb_ext.ImpWrapper{Imp: &newImp}
			newImpWrapper.GetImpExt()

			imps = append(imps, newImpWrapper)
		}

	}

	return imps, errs

}

func generateImpressionID(impID string, seqNo int) string {
	return fmt.Sprintf(impressionIDFormat, impID, seqNo)
}

// getAdPodImpsConfigs will return number of impressions configurations within adpod
func getAdPodImpConfig(imp *openrtb2.Imp, adpod *models.AdPod) ([]*models.ImpAdPodConfig, error) {
	// This case for non adpod video impression
	if adpod == nil {
		return nil, nil
	}
	selectedAlgorithm := SelectAlgorithm(adpod)
	impGen := NewImpressions(imp.Video.MinDuration, imp.Video.MaxDuration, adpod, selectedAlgorithm)
	impRanges := impGen.Get()

	// labels := metrics.PodLabels{AlgorithmName: impressions.MonitorKey[selectedAlgorithm], NoOfImpressions: new(int)}

	// //log number of impressions in stats
	// *labels.NoOfImpressions = len(impRanges)
	// deps.metricsEngine.RecordPodImpGenTime(labels, start)

	// check if algorithm has generated impressions
	if len(impRanges) == 0 {
		return nil, errors.New("unable to generate impressions for adpod for impression: " + imp.ID)
	}

	config := make([]*models.ImpAdPodConfig, len(impRanges))
	for i, value := range impRanges {
		eachConfig := models.ImpAdPodConfig{
			ImpID:          generateImpressionID(imp.ID, i+1),
			MinDuration:    value[0],
			MaxDuration:    value[1],
			SequenceNumber: int8(i + 1), /* Must be starting with 1 */
		}
		config[i] = &eachConfig
	}
	return config, nil
}

// SelectAlgorithm is factory function which will return valid Algorithm based on adpod parameters
// Return Value:
//   - MinMaxAlgorithm (default)
//   - ByDurationRanges: if reqAdPod extension has VideoAdDuration and VideoAdDurationMatchingPolicy is "exact" algorithm
func SelectAlgorithm(reqAdPod *models.AdPod) int {
	if reqAdPod != nil {
		if len(reqAdPod.VideoAdDuration) > 0 &&
			(models.OWExactVideoAdDurationMatching == reqAdPod.VideoAdDurationMatching || models.OWRoundupVideoAdDurationMatching == reqAdPod.VideoAdDurationMatching) {
			return models.ByDurationRanges
		}
	}
	return models.MinMaxAlgorithm
}

// NewImpressions generate object of impression generator
// based on input algorithm type
// if invalid algorithm type is passed, it returns default algorithm which will compute
// impressions based on minimum ad slot duration
func NewImpressions(podMinDuration, podMaxDuration int64, adpod *models.AdPod, algorithm int) ImpGenerator {
	switch algorithm {
	case models.MaximizeForDuration:
		g := newMaximizeForDuration(podMinDuration, podMaxDuration, adpod)
		return &g

	case models.MinMaxAlgorithm:
		g := newMinMaxAlgorithm(podMinDuration, podMaxDuration, adpod)
		return &g

	case models.ByDurationRanges:
		g := newByDurationRanges(adpod.VideoAdDurationMatching, adpod.VideoAdDuration,
			int(adpod.MaxAds),
			adpod.MinDuration, adpod.MaxDuration)

		return &g
	}

	// return default algorithm with slot durations set to minimum slot duration
	// util.Logf("Selected 'DefaultAlgorithm'")
	defaultGenerator := newConfig(podMinDuration, podMinDuration, adpod)
	return &defaultGenerator
}

// newConfigWithMultipleOf initializes the generator instance
// it internally calls newConfig to obtain the generator instance
// then it computes closed to factor basedon 'multipleOf' parameter value
// and accordingly determines the Pod Min/Max and Slot Min/Max values for internal
// computation only.
func newConfigWithMultipleOf(podMinDuration, podMaxDuration int64, vPod *models.AdPod, multipleOf int) generator {
	config := newConfig(podMinDuration, podMaxDuration, vPod)

	// try to compute slot level min and max duration values in multiple of
	// given number. If computed values are overlapping then prefer requested
	if config.requested.slotMinDuration == config.requested.slotMaxDuration {
		config.internal.slotMinDuration = config.requested.slotMinDuration
		config.internal.slotMaxDuration = config.requested.slotMaxDuration
	} else {
		config.internal.slotMinDuration = getClosestFactorForMinDuration(config.requested.slotMinDuration, int64(multipleOf))
		config.internal.slotMaxDuration = getClosestFactorForMaxDuration(config.requested.slotMaxDuration, int64(multipleOf))
		config.internal.slotDurationComputed = true
		if config.internal.slotMinDuration > config.internal.slotMaxDuration {
			// computed slot min duration > computed slot max duration
			// avoid overlap and prefer requested values
			config.internal.slotMinDuration = config.requested.slotMinDuration
			config.internal.slotMaxDuration = config.requested.slotMaxDuration
			// update marker indicated slot duation values are not computed
			// this required by algorithm in computeTimeForEachAdSlot function
			config.internal.slotDurationComputed = false
		}
	}
	return config
}

// newConfig initializes the generator instance
func newConfig(podMinDuration, podMaxDuration int64, vPod *models.AdPod) generator {
	config := generator{}
	config.totalSlotTime = new(int64)
	// configure requested pod
	config.requested = pod{
		podMinDuration:  podMinDuration,
		podMaxDuration:  podMaxDuration,
		slotMinDuration: int64(vPod.MinDuration),
		slotMaxDuration: int64(vPod.MaxDuration),
		minAds:          int64(vPod.MinAds),
		maxAds:          int64(vPod.MaxAds),
	}

	// configure internal object (FOR INTERNAL USE ONLY)
	// this  is used for internal computation and may contains modified values of
	// slotMinDuration and slotMaxDuration in multiples of multipleOf factor
	// This function will by deault intialize this pod with same values
	// as of requestedPod
	// There is another function newConfigWithMultipleOf, which computes and assigns
	// values to this object
	config.internal = internal{
		slotMinDuration: config.requested.slotMinDuration,
		slotMaxDuration: config.requested.slotMaxDuration,
	}
	return config
}

// Returns closest factor for num, with  respect  input multipleOf
//
//	Example: Closest Factor of 9, in multiples of 5 is '10'
func getClosestFactor(num, multipleOf int64) int64 {
	return int64(math.Round(float64(num)/float64(multipleOf)) * float64(multipleOf))
}

// Returns closestfactor of MinDuration, with  respect to multipleOf
// If computed factor < MinDuration then it will ensure and return
// close factor >=  MinDuration
func getClosestFactorForMinDuration(MinDuration, multipleOf int64) int64 {
	closedMinDuration := getClosestFactor(MinDuration, multipleOf)

	if closedMinDuration == 0 {
		return multipleOf
	}

	if closedMinDuration < MinDuration {
		return closedMinDuration + multipleOf
	}

	return closedMinDuration
}

// Returns closestfactor of maxduration, with  respect to multipleOf
// If computed factor > maxduration then it will ensure and return
// close factor <=  maxduration
func getClosestFactorForMaxDuration(maxduration, multipleOf int64) int64 {
	closedMaxDuration := getClosestFactor(maxduration, multipleOf)
	if closedMaxDuration == maxduration {
		return maxduration
	}

	// set closest maxduration closed to maxduration
	for i := closedMaxDuration; i <= maxduration; {
		if closedMaxDuration < maxduration {
			closedMaxDuration = i + multipleOf
			i = closedMaxDuration
		}
	}

	if closedMaxDuration > maxduration {
		duration := closedMaxDuration - multipleOf
		if duration == 0 {
			// return input value as is instead of zero to avoid NPE
			return maxduration
		}
		return duration
	}

	return closedMaxDuration
}

// Returns Maximum number out off 2 input numbers
func max(num1, num2 int64) int64 {

	if num1 >= num2 {
		return num1
	}

	return num2
}

// Returns true if num is multipleof second argument. False otherwise
func isMultipleOf(num, multipleOf int64) bool {
	return math.Mod(float64(num), float64(multipleOf)) == 0
}
