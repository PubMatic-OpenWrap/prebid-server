package modules

import (
	prebidOrtb2blocking "github.com/prebid/prebid-server/modules/prebid/ortb2blocking"
	trafficshapping "github.com/prebid/prebid-server/modules/prebid/traffic_shapping"
)

// builders returns mapping between module name and its builder
// vendor and module names are chosen based on the module directory name
func builders() ModuleBuilders {
	return ModuleBuilders{
		"prebid": {
			"ortb2blocking":    prebidOrtb2blocking.Builder,
			"traffic_shapping": trafficshapping.Builder,
		},
	}
}
