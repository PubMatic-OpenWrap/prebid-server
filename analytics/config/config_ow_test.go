package config

import (
	"testing"

	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/analytics/filesystem"
)

func TestEnableAnalyticsModule(t *testing.T) {

	modules := enabledAnalytics{}
	file, err := filesystem.NewFileLogger("xyz1.txt")
	if err != nil {
		t.Errorf("NewFileLogger returned error - %v", err.Error())
	}

	type arg struct {
		moduleList analytics.PBSAnalyticsModule
		module     analytics.PBSAnalyticsModule
	}

	type want struct {
		len    int
		errMsg string
	}

	tests := []struct {
		description string
		args        arg
		wants       want
	}{
		{
			description: "add non-nil module to nil module-list",
			args:        arg{moduleList: nil, module: file},
			wants:       want{len: 0, errMsg: "failed to convert moduleList interface from analytics.PBSAnalyticsModule to analytics.enabledAnalytics"},
		},
		{
			description: "add nil module to non-nil module-list",
			args:        arg{moduleList: modules, module: nil},
			wants:       want{len: 0, errMsg: "module to be added is nil"},
		},
		{
			description: "add non-nil module to non-nil module-list",
			args:        arg{moduleList: modules, module: file},
			wants:       want{len: 1, errMsg: ""},
		},
	}

	for _, tt := range tests {
		actual, err := EnableAnalyticsModule(tt.args.module, tt.args.moduleList)

		if err != nil {
			if err.Error() != tt.wants.errMsg {
				t.Errorf("Expected error - [%v], got error - [%v] ,for test-case : [%v]", tt.wants.errMsg, err.Error(), tt.description)
			}
			continue // error message is same i.e. test case passed
		}

		list, ok := actual.(enabledAnalytics)
		if !ok {
			t.Errorf("Failed to convert interface to enabledAnalytics for test case - [%v]", tt.description)
		}

		if len(list) != tt.wants.len {
			t.Errorf("length of enabled modules mismatched, expected - [%d] , got - [%d]", tt.wants.len, len(list))
		}
	}
}
