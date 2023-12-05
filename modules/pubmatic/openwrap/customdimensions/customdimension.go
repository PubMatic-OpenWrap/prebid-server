package customdimensions

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

	"github.com/buger/jsonparser"
)

// returns customDimension map if req.ext.prebid.bidderparams.pubmatic have cds
func GetCustomDimensions(bidderParams json.RawMessage) (map[string]models.CustomDimension, error) {
	cds := make(map[string]models.CustomDimension, 0)
	if len(bidderParams) == 0 {
		return cds, errors.New("empty bidderParams")
	}
	if cdsContent, _, _, err := jsonparser.Get([]byte(bidderParams), models.BidderPubMatic, models.CustomDimensions); err == nil {
		reqCustomDimension := make(map[string]models.CustomDimension, 0)
		if err := json.Unmarshal(cdsContent, &reqCustomDimension); err == nil {
			return reqCustomDimension, nil
		}
	}
	return cds, errors.New("custom dimensions not found")
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
