package ctv

import (
	"errors"
	"fmt"

	"github.com/prebid/openrtb/v19/openrtb2"
)

func FilterNonVideoImpressions(request *openrtb2.BidRequest) error {
	if request != nil && len(request.Imp) > 0 {
		j := 0
		for index, imp := range request.Imp {
			//Validate Native Impressions
			if imp.Video == nil {
				continue
			}

			//Banner Request Not Supported
			imp.Banner = nil

			//Native Request Not Supported
			imp.Native = nil

			if index != j {
				request.Imp[j] = imp
			}
			j++
		}
		request.Imp = request.Imp[:j]
		if len(request.Imp) == 0 {
			return fmt.Errorf("video object is missing for ctv request")
		}
	}
	return nil
}

func ValidateVideoImpressions(request *openrtb2.BidRequest) error {
	if len(request.Imp) == 0 {
		return errors.New("recieved request with no impressions")
	}

	var validImpCount int
	for _, imp := range request.Imp {
		if imp.Video != nil {
			validImpCount++
		}
	}

	if validImpCount == 0 {
		return errors.New("video object is missing for ctv request")
	}

	return nil
}
