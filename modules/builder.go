package modules

import (
	prebidOrtb2blocking "github.com/prebid/prebid-server/v2/modules/prebid/ortb2blocking"
	pubmaticOpenwrap "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/vastunwrap"
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
			"openwrap":   pubmaticOpenwrap.Builder,
		},
	}
}
