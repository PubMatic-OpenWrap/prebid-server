// Package impressions provides various algorithms to get the number of impressions
// along with minimum and maximum duration of each impression.
// It uses Ad pod request for it
package impressions

import (
	"errors"

	"github.com/PubMatic-OpenWrap/prebid-server/endpoints/openrtb2/ctv"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

const (
	// Algorithm1 tends towards Ad Pod Maximum Duration, Ad Slot Maximum Duration
	// and Maximum number of Ads. Accordingly it computes the number of impressions
	Algorithm1 = iota
	// Algorithm2 computes number of impressions using following  steps
	//  1. Passes input configuration as it is (Equivalent of Algorithm1)
	//	2. Ad Pod Duration = Ad Pod Max Duration, Number of Ads = max ads
	//	3. Ad Pod Duration = Ad Pod Max Duration, Number of Ads = min ads
	//	4. Ad Pod Duration = Ad Pod Min Duration, Number of Ads = max ads
	//	5. Ad Pod Duration = Ad Pod Min Duration, Number of Ads = min ads
	Algorithm2
)

// Value use to compute Ad Slot Durations and Pod Durations for internal computation
// Right now this value is set to 5, based on passed data observations
// Observed that typically video impression contains contains minimum and maximum duration in multiples of  5
var multipleOf = int64(5)

// IImpressions ...
type IImpressions interface {
	Get() [][2]int64
	Algorithm() int // returns algorithm used for computing number of impressions
}

// NewImpressions generate object of impression generator
// based on input algorithm type
func NewImpressions(podMinDuration, podMaxDuration int64, vPod openrtb_ext.VideoAdPod, algorithm int) (IImpressions, error) {
	switch algorithm {
	case Algorithm1:
		ctv.Logf("Selected 'Algorithm1' ")
		g := newImpGenA1(podMinDuration, podMaxDuration, vPod)
		return &g, nil

	case Algorithm2:
		ctv.Logf("Selected 'Algorithm2' ")
		g := newImpGenA2(podMinDuration, podMaxDuration, vPod)
		return &g, nil
	}
	return nil, errors.New("Invalid algorithm value")
}
