package build

import (
	"errors"
	"testing"

	"github.com/prebid/prebid-server/v2/analytics"
	"github.com/prebid/prebid-server/v2/analytics/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestEnableAnalyticsModule(t *testing.T) {

	modules := enabledAnalytics{}
	file, err := filesystem.NewFileLogger("xyz1.txt")
	if err != nil {
		t.Errorf("NewFileLogger returned error - %v", err.Error())
	}

	type arg struct {
		moduleList analytics.Runner
		module     analytics.Module
	}

	type want struct {
		len   int
		error error
	}

	tests := []struct {
		description string
		args        arg
		wants       want
	}{
		{
			description: "add non-nil module to nil module-list",
			args:        arg{moduleList: nil, module: file},
			wants:       want{len: 0, error: errors.New("failed to convert moduleList interface from analytics.Module to analytics.enabledAnalytics")},
		},
		{
			description: "add nil module to non-nil module-list",
			args:        arg{moduleList: modules, module: nil},
			wants:       want{len: 0, error: errors.New("module to be added is nil")},
		},
		{
			description: "add non-nil module to non-nil module-list",
			args:        arg{moduleList: modules, module: file},
			wants:       want{len: 1, error: nil},
		},
	}

	for _, tt := range tests {
		actual, err := EnableAnalyticsModule(tt.args.module, tt.args.moduleList)
		assert.Equal(t, err, tt.wants.error)

		if err == nil {
			list, ok := actual.(enabledAnalytics)
			if !ok {
				t.Errorf("Failed to convert interface to enabledAnalytics for test case - [%v]", tt.description)
			}

			if len(list) != tt.wants.len {
				t.Errorf("length of enabled modules mismatched, expected - [%d] , got - [%d]", tt.wants.len, len(list))
			}
		}
	}
}
