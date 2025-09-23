package ctvendpointmanager

import (
	"errors"
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

/*
setIncludeBrandCategory sets PBS's  bidrequest.ext.prebid.Targeting object
 1. If pReqExt.supportDeals  = true then sets IncludeBrandCategory of targeting as follows
    WithCategory        = false
    TranslateCategories = false
*/
func setIncludeBrandCategory(rCtx models.RequestCtx) {
	includeBrandCategory := &openrtb_ext.ExtIncludeBrandCategory{
		SkipDedup:           true,
		TranslateCategories: ptrutil.ToPtr(false),
	}

	if rCtx.NewReqExt.Wrapper != nil && rCtx.NewReqExt.Wrapper.IncludeBrandCategory != nil &&
		(models.IncludeIABBranchCategory == *rCtx.NewReqExt.Wrapper.IncludeBrandCategory ||
			models.IncludeAdServerBrandCategory == *rCtx.NewReqExt.Wrapper.IncludeBrandCategory) {

		includeBrandCategory.WithCategory = true

		if models.IncludeAdServerBrandCategory == *rCtx.NewReqExt.Wrapper.IncludeBrandCategory {
			adserver := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.AdserverKey)
			prebidAdServer := getPrebidPrimaryAdServer(adserver)
			if prebidAdServer > 0 {
				includeBrandCategory.PrimaryAdServer = prebidAdServer
				includeBrandCategory.Publisher = getPrebidPublisher(adserver)
				*includeBrandCategory.TranslateCategories = true
			} else {
				includeBrandCategory.WithCategory = false
			}
		}
	}
	rCtx.NewReqExt.Prebid.Targeting.IncludeBrandCategory = includeBrandCategory
}

func getPrebidPrimaryAdServer(adserver string) int {
	//TODO: Make it map[OWPrimaryAdServer]PrebidPrimaryAdServer
	//1-Freewheel 2-DFP
	if models.OWPrimaryAdServerDFP == adserver {
		return models.PrebidPrimaryAdServerDFPID
	}
	return 0
}

func getPrebidPublisher(adserver string) string {
	//TODO: Make it map[OWPrimaryAdServer]PrebidPrimaryAdServer
	if models.OWPrimaryAdServerDFP == adserver {
		return models.PrebidPrimaryAdServerDFP
	}
	return ""
}

func validateVideoImpressions(request *openrtb2.BidRequest) error {
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
		return errors.New("video object is missing in the request")
	}

	return nil
}

// isValidSchain validated the schain object
func isValidSchain(schain *openrtb2.SupplyChain) error {
	if schain.Ver != openrtb_ext.SChainVersion1 {
		return fmt.Errorf("invalid schain version, version should be %s", openrtb_ext.SChainVersion1)
	}

	if (int(schain.Complete) != openrtb_ext.SChainCompleteYes) && (schain.Complete != openrtb_ext.SChainCompleteNo) {
		return errors.New("invalid schain.complete value should be 0 or 1")
	}

	if len(schain.Nodes) == 0 {
		return errors.New("invalid schain node fields, Node can't be empty")
	}

	for _, schainNode := range schain.Nodes {
		if schainNode.ASI == "" {
			return errors.New("invalid schain node fields, ASI can't be empty")
		}

		if schainNode.SID == "" {
			return errors.New("invalid schain node fields, SID can't be empty")
		}

		if len([]rune(schainNode.SID)) > openrtb_ext.SIDLength {
			return errors.New("invalid schain node fields, sid can have maximum 64 characters")
		}

		// for schain version 1.0 hp must be 1
		if schainNode.HP == nil || *schainNode.HP != openrtb_ext.HPOne {
			return errors.New("invalid schain node fields, HP must be one")
		}
	}
	return nil
}

func filterNonVideoImpressions(request *openrtb2.BidRequest) error {
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
