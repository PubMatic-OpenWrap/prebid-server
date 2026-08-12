package build

import (
	"fmt"

	"github.com/prebid/prebid-server/v3/analytics"
)

// EnableAnalyticsModule will add the new module into the list of enabled analytics modules
var EnableAnalyticsModule = func(module analytics.Module, moduleList analytics.Runner) (analytics.Runner, error) {
	if module == nil {
		return nil, fmt.Errorf("module to be added is nil")
	}
	enabledModuleList, ok := moduleList.(enabledAnalytics)
	if !ok {
		return nil, fmt.Errorf("failed to convert moduleList interface from analytics.Module to analytics.enabledAnalytics")
	}
	enabledModuleList["pubstack"] = module
	return enabledModuleList, nil
}
