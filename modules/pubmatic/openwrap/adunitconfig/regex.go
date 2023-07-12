package adunitconfig

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func getRegexMatch(rctx models.RequestCtx, slotName string) string {
	for expression := range rctx.AdUnitConfig.Config {
		if expression != models.AdunitConfigDefaultKey {
			//Populating and Validating
			re, err := Compile(expression)
			if err != nil {
				// TODO: add debug messages
				// errs = append(errs, err)
				continue
			}

			if re.MatchString(slotName) {
				return expression
			}
		}
	}
	return ""
}
