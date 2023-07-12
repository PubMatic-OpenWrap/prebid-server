package modules

import (
	prebidOrtb2blocking "github.com/prebid/prebid-server/modules/prebid/ortb2blocking"
	vastunwrap "github.com/prebid/prebid-server/modules/pubmatic/vastunwrap"
	pubmaticOpenwrap "github.com/prebid/prebid-server/modules/pubmatic/openwrap"
)

// builders returns mapping between module name and its builder
// vendor and module names are chosen based on the module directory name
func builders() ModuleBuilders {
	return ModuleBuilders{
		"prebid": {
			"ortb2blocking": prebidOrtb2blocking.Builder,
		},
		"pubmatic": {
			"vastunwrap": vastunwrap.Builder,
			"openwrap": pubmaticOpenwrap.Builder,
		},
	}
}
