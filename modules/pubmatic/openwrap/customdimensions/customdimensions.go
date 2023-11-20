package customdimensions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"

	"github.com/buger/jsonparser"
)

type CustomDimension struct {
	Value     string `json:"value"`
	SendToGAM *bool  `json:"sendtoGAM,omitempty"`
}

// returns customDimension map if req.ext.prebid.bidderparams.pubmatic have cds
func GetCustomDimensionsFromRequestExt(prebid *openrtb_ext.ExtRequestPrebid) map[string]CustomDimension {
	if prebid != nil && prebid.BidderParams != nil && len(prebid.BidderParams) > 0 {
		if cdsContent, _, _, parseErr := jsonparser.Get([]byte(prebid.BidderParams), models.BidderPubMatic, models.CustomDimensions); parseErr == nil {
			reqCustomDimension := make(map[string]CustomDimension, 0)
			if err := json.Unmarshal(cdsContent, &reqCustomDimension); err == nil {
				return reqCustomDimension
			}
		}
	}
	return nil
}

// Will parse validated cds and parse/convert to string
func ParseCustomDimensionsToString(cds map[string]CustomDimension) string {
	if len(cds) == 0 {
		return ""
	}

	cdsSlc := []string{}
	for k, v := range cds {
		ele := fmt.Sprintf(`%s=%s`, k, v.Value)
		cdsSlc = append(cdsSlc, ele)
	}

	return strings.Join(cdsSlc, ";")
}

// returns if custom dimensions present in req.ext and map of customDimensiosn with attributes value and sendtoGAM
func IsCustomDimensionsPresent(ext interface{}) (map[string]CustomDimension, bool) {
	CustomDimensionsKeyValues := make(map[string]CustomDimension, 0)

	extData, _ := json.Marshal(ext)
	if wtExt, err := models.GetRequestExt(extData); err == nil {
		if CustomDimensionsKeyValues := GetCustomDimensionsFromRequestExt(&wtExt.Prebid); CustomDimensionsKeyValues != nil {
			return CustomDimensionsKeyValues, true
		}
	}

	return CustomDimensionsKeyValues, false
}
