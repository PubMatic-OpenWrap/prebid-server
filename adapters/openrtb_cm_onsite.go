package adapters

import (
	"encoding/json"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

const (
	AccountID_KeyOnsite = "Bidder_AccountID"
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
	var mapExt map[string]interface{}

	if prebidExt.Prebid.BidderParams != nil {
		if err := json.Unmarshal(prebidExt.Prebid.BidderParams, &mapExt); err != nil {
			return nil, &errortypes.BadInput{
				Message: "Impression extension not provided or can't be unmarshalled",
			}
		}

		if ext, ok := mapExt["ext"]; ok {
			extBytes, err := json.Marshal(ext)
			if err != nil {
				return nil, &errortypes.BadInput{
					Message: "Error marshalling impression extension",
				}
			}

			if err := json.Unmarshal(extBytes, &requestExtCMOnsite); err != nil {
				return nil, &errortypes.BadInput{
					Message: "Error unmarshalling impression extension to ExtRequestPrebidOnsite",
				}
			}
		} else {
			return nil, &errortypes.BadInput{
				Message: "Impression extension not provided",
			}
		}
	}

	return requestExtCMOnsite, nil
}

func GetInventoryAndAccountDetailsCMOnsite(requestExtCMOnsite *openrtb_ext.ExtRequestPrebidOnsite) (
	map[string]openrtb_ext.CMOnsiteInventoryDetails, string, []error) {
	if requestExtCMOnsite == nil {
		return nil, "", nil
	}

	var errors []error
	inventoryDetails := make(map[string]openrtb_ext.CMOnsiteInventoryDetails)
	var accountId string

	if requestExtCMOnsite.ZoneMapping == nil {
		errors = append(errors, &errortypes.BadInput{
			Message: "ZoneMapping not provided",
		})
	}
	for key, value := range requestExtCMOnsite.ZoneMapping {
		switch val := value.(type) {
		case map[string]interface{}:
			// Attempt to unmarshal into CMOnsiteInventoryDetails
			var details openrtb_ext.CMOnsiteInventoryDetails
			detailBytes, err := json.Marshal(val)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			if err := json.Unmarshal(detailBytes, &details); err == nil {
				inventoryDetails[key] = details
			}
		case string:
			if key == AccountID_KeyOnsite {
				accountId = val
			}
		}
	}
	return inventoryDetails, accountId, errors
}

func ValidateCMOnsiteRequest(request *openrtb2.BidRequest) (
	*openrtb_ext.ExtSiteCommerce, *openrtb_ext.ExtRequestPrebidOnsite, []error) {
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
	} else {

	}

	if len(errors) > 0 {
		return nil, nil, errors
	}

	return siteExt, requestExtCMOnsite, nil
}
