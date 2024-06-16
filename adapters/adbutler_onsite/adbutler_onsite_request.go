package adbutler_onsite

import (
	"github.com/PubMatic-OpenWrap/prebid-server/errortypes"
	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/adapters"
)

func (a *AdButlerOnsiteAdapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	siteExt, requestExt, errors := adapters.ValidateCMOnsiteRequest(request)
	if len(errors) > 0 {
		return nil, errors
	}

	for _, imp := range request.Imp {
		// Parse each imp element here
		impExt := adapters.GetImpressionExtCMOnsite(imp)
		if impExt == nil {
			return nil, []error{&errortypes.BadInput{
				Message: "Missing required ext fields",
			}}
		}
	
	}

	if siteExt == nil || requestExt == nil {
		return nil, []error{&errortypes.BadInput{
			Message: "Missing required ext fields",
		}}
	}

	return nil,nil
}

