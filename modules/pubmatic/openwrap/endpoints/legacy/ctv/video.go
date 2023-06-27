package ctv

import (
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
