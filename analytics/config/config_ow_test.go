package config

import (
	"testing"

	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/analytics/filesystem"
)

func TestAddAnalyticModules(t *testing.T) {

	modules := enabledAnalytics{}
	file, err := filesystem.NewFileLogger("xyz1.txt")
	if err != nil {
		t.Errorf("NewFileLogger returned error - %v", err.Error())
	}

	tests := []struct {
		description string
		modules     analytics.PBSAnalyticsModule
		module      analytics.PBSAnalyticsModule
		len         int
		expectErr   bool
	}{
		{
			description: "add non-nil module to nil module-list",
			modules:     nil,
			module:      file,
			len:         0,
			expectErr:   true,
		},
		{
			description: "add nil module to non-nil module-list",
			modules:     modules,
			module:      nil,
			len:         0,
			expectErr:   true,
		},
		{
			description: "add non-nil module to non-nil module-list",
			modules:     modules,
			module:      file,
			len:         1,
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		actual, err := AddAnalyticsModule(tt.modules, tt.module)
		if err != nil && tt.expectErr {
			continue
		}
		if err == nil && tt.expectErr {
			t.Errorf("Expecting an error but not received any for test-case : [%v]", tt.description)
		}
		list, ok := actual.(enabledAnalytics)
		if !ok {
			t.Errorf("Failed to convert interface to enabledAnalytics for test case - [%v]", tt.description)
		}

		if len(list) != tt.len {
			t.Errorf("Expected len=%d , got=%d", tt.len, len(list))
		}
	}
}
