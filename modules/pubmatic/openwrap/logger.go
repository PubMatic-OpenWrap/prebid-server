package openwrap

import (
	"fmt"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func getIncomingSlots(imp openrtb2.Imp) []string {
	sizes := map[string]struct{}{}
	if imp.Banner != nil {
		if imp.Banner.W != nil && imp.Banner.H != nil {
			sizes[fmt.Sprintf("%dx%d", *imp.Banner.W, *imp.Banner.H)] = struct{}{}
		}

		for _, format := range imp.Banner.Format {
			sizes[fmt.Sprintf("%dx%d", format.W, format.H)] = struct{}{}
		}
	}

	if imp.Video != nil {
		sizes[fmt.Sprintf("%dx%dv", imp.Video.W, imp.Video.H)] = struct{}{}
	}

	var s []string
	for k := range sizes {
		s = append(s, k)
	}
	return s
}

func getDefaultImpBidCtx(request openrtb2.BidRequest) map[string]models.ImpCtx {
	impBidCtx := make(map[string]models.ImpCtx)
	for _, imp := range request.Imp {
		incomingSlots := getIncomingSlots(imp)

		impBidCtx[imp.ID] = models.ImpCtx{
			IncomingSlots: incomingSlots,
		}
	}
	return impBidCtx
}
