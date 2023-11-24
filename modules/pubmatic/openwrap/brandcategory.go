package openwrap

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
)

/*
setIncludeBrandCategory sets PBS's  bidrequest.ext.prebid.Targeting object
 1. If pReqExt.supportDeals  = true then sets IncludeBrandCategory of targeting as follows
    WithCategory        = false
    TranslateCategories = false
*/
func setIncludeBrandCategory(wrapperExt *models.RequestExtWrapper, prebidExt *openrtb_ext.ExtRequestPrebid, partnerConfigMap map[int]map[string]string, IsCTVAPIRequest bool) {

	if IsCTVAPIRequest {
		includeBrandCategory := &openrtb_ext.ExtIncludeBrandCategory{
			SkipDedup:           true,
			TranslateCategories: ptrutil.ToPtr(false),
		}

		if wrapperExt != nil && wrapperExt.IncludeBrandCategory != nil &&
			(models.IncludeIABBranchCategory == *wrapperExt.IncludeBrandCategory ||
				models.IncludeAdServerBrandCategory == *wrapperExt.IncludeBrandCategory) {

			includeBrandCategory.WithCategory = true

			if models.IncludeAdServerBrandCategory == *wrapperExt.IncludeBrandCategory {
				adserver := models.GetVersionLevelPropertyFromPartnerConfig(partnerConfigMap, models.AdserverKey)
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
		prebidExt.Targeting.IncludeBrandCategory = includeBrandCategory
	}

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
