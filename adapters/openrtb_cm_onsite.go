package adapters

import (
	"encoding/json"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func GetImpressionExtCMOnsite(imp *openrtb2.Imp) (*openrtb_ext.ExtImpCMOnsitePrebid, error) {
	var impExt openrtb_ext.ExtImpCMOnsitePrebid
	if err := json.Unmarshal(imp.Ext, &impExt); err != nil {
		return nil, &errortypes.BadInput{
			Message: "Impression extension not provided or can't be unmarshalled",
		}
	}

	return &impExt, nil

}

func GetRequestExtCMOnsite(prebidExt *openrtb_ext.ExtOWRequest) (*openrtb_ext.ExtRequestPrebidOnsite, error) {
	var requestExtCMOnsite *openrtb_ext.ExtRequestPrebidOnsite

	if prebidExt.Prebid.BidderParams != nil {
		if err := json.Unmarshal(prebidExt.Prebid.BidderParams, &requestExtCMOnsite); err != nil {
			return nil, &errortypes.BadInput{
				Message: "Impression extension not provided or can't be unmarshalled",
			}
		}
	}

	return requestExtCMOnsite, nil
}

func ValidateCMOnsiteRequest(request *openrtb2.BidRequest) (
	*openrtb_ext.ExtSiteCommerce, map[string]interface{}, []error) {
	var siteExt *openrtb_ext.ExtSiteCommerce
	var requestExt *openrtb_ext.ExtOWRequest
	var requestExtCMOnsite *openrtb_ext.ExtRequestPrebidOnsite

	var err error
	var errors []error

	siteExt, err = GetSiteExtComm(request)
	if err != nil {
		errors = append(errors, err)
	} 

	requestExt, err = GetRequestExtComm(request)
	if err != nil {
		errors = append(errors, err)
	}

	requestExtCMOnsite, err = GetRequestExtCMOnsite(requestExt)
	if err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return nil, nil, errors
	}

	return siteExt, requestExtCMOnsite, nil
}



