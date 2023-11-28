package customdimensions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

	"github.com/buger/jsonparser"
)

type CustomDimension struct {
	Value     string `json:"value,omitempty"`
	SendToGAM *bool  `json:"sendtoGAM,omitempty"`
}

// returns customDimension map if req.ext.prebid.bidderparams.pubmatic have cds
func GetCustomDimensionsFromRequestExt(bidderParams json.RawMessage) map[string]CustomDimension {
	if len(bidderParams) == 0 {
		return nil
	}
	if cdsContent, _, _, parseErr := jsonparser.Get([]byte(bidderParams), models.BidderPubMatic, models.CustomDimensions); parseErr == nil {
		reqCustomDimension := make(map[string]CustomDimension, 0)
		if err := json.Unmarshal(cdsContent, &reqCustomDimension); err == nil {
			return reqCustomDimension
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
	cds := make(map[string]CustomDimension, 0)
	extData, err := json.Marshal(ext)
	if err == nil {
		reqExt, err := models.GetRequestExt(extData)
		if err != nil {
			return cds, false
		}
		if cds := GetCustomDimensionsFromRequestExt(reqExt.Prebid.BidderParams); cds != nil {
			return cds, true
		}
	}
	return cds, false
}
