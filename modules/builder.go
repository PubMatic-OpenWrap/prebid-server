package modules

import (
	"github.com/PubMatic-OpenWrap/prebid-server/modules/prebid/openwrap"
	prebidOrtb2blocking "github.com/prebid/prebid-server/modules/prebid/ortb2blocking"
)

// builders returns mapping between module name and its builder
// vendor and module names are chosen based on the module directory name
func builders() ModuleBuilders {
	return ModuleBuilders{
		"prebid": {
			"ortb2blocking": prebidOrtb2blocking.Builder,
		},
		"pubmatic": {
			"openwrap": openwrap.Builder,
		},
	}
}
