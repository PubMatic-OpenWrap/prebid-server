package config

import (
	"fmt"

	"github.com/prebid/prebid-server/analytics"
)

// contextKey will be used to pass the object through request.Context
type contextKey string

const (
	CtxWLRecord    contextKey = "wrapperLoggerRecord"
	CtxRejectedBid contextKey = "rejectedBids"
)

func AddAnalyticsModule(moduleList, module analytics.PBSAnalyticsModule) (analytics.PBSAnalyticsModule, error) {
	if module == nil {
		return nil, fmt.Errorf("module to be added is nil")
	}
	enabledModuleList, ok := moduleList.(enabledAnalytics)
	if !ok {
		return nil, fmt.Errorf("failed to convert moduleList interface from analytics.PBSAnalyticsModule to analytics.enabledAnalytics")
	}
	enabledModuleList = append(enabledModuleList, module)
	return enabledModuleList, nil
}
