package customdimensions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

	"github.com/buger/jsonparser"
)

// returns customDimension map if req.ext.prebid.bidderparams.pubmatic have cds
func GetCustomDimensions(bidderParams json.RawMessage) map[string]models.CustomDimension {
	cds := make(map[string]models.CustomDimension, 0)
	if len(bidderParams) == 0 {
		return cds
	}
	if cdsContent, _, _, err := jsonparser.Get([]byte(bidderParams), models.BidderPubMatic, models.CustomDimensions); err == nil {
		if err := json.Unmarshal(cdsContent, &cds); err == nil {
			return cds
		}
	}
	return cds
}

// ConvertCustomDimensionsToString will convert to key-val string
func ConvertCustomDimensionsToString(cds map[string]models.CustomDimension) string {
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
