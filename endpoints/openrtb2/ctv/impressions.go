package ctv

import (
	"errors"

	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

// IImpressions ...
type IImpressions interface {
	Get() [][2]int64
}

// Impressions ...
type Impressions struct {
	IImpressions
	generator IImpressions
	algorithm int
}

// NewImpressions generate object of impression generator
// based on input algorithm type
func NewImpressions(podMinDuration, podMaxDuration int64, vPod openrtb_ext.VideoAdPod, algorithm int) (*Impressions, error) {
	switch algorithm {
	case 1:
		return &Impressions{
			generator: newImpGenA1(podMinDuration, podMaxDuration, vPod),
			algorithm: algorithm,
		}, nil

	case 2:
		return &Impressions{
			generator: newImpGenA2(podMinDuration, podMaxDuration, vPod),
			algorithm: algorithm,
		}, nil
	}
	return nil, errors.New("Invalid algorithm value")
}

// Get ...
func (i *Impressions) Get() [][2]int64 {
	return i.generator.Get()
}
