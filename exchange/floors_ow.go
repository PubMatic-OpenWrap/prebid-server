package exchange

import (
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// floorsEnabled will return true if floors are enabled in both account and request level
func floorsEnabled(account config.Account, bidRequestWrapper *openrtb_ext.RequestWrapper) (bool, *openrtb_ext.PriceFloorRules) {
	var (
		reqEnabled bool
		floorRules *openrtb_ext.PriceFloorRules
	)

	if requestExt, err := bidRequestWrapper.GetRequestExt(); err == nil {
		if prebidExt := requestExt.GetPrebid(); prebidExt != nil {
			reqEnabled = prebidExt.Floors.GetEnabled()
			floorRules = prebidExt.Floors
		}
	}

	return account.PriceFloors.Enabled && reqEnabled, floorRules
}
