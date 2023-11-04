package pubmatic

import (
	"net/url"
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestPrepareLoggerURL(t *testing.T) {
	type args struct {
		wlog        *WloggerRecord
		loggerURL   string
		gdprEnabled int
	}
	tests := []struct {
		name     string
		args     args
		owlogger string
	}{
		{
			name: "nil_wlog",
			args: args{
				wlog:        nil,
				loggerURL:   "http://t.pubmatic.com/wl",
				gdprEnabled: 1,
			},
			owlogger: "",
		},
		{
			name: "gdprEnabled=1",
			args: args{
				wlog: &WloggerRecord{
					record: record{
						PubID:     10,
						ProfileID: "1",
						VersionID: "0",
					},
				},
				loggerURL:   "http://t.pubmatic.com/wl",
				gdprEnabled: 1,
			},
			owlogger: `http://t.pubmatic.com/wl?gdEn=1&json={"pubid":10,"pid":"1","pdvid":"0","dvc":{},"ft":0}&pubid=10`,
		},
		{
			name: "gdprEnabled=0",
			args: args{
				wlog: &WloggerRecord{
					record: record{
						PubID:     10,
						ProfileID: "1",
						VersionID: "0",
					},
				},
				loggerURL:   "http://t.pubmatic.com/wl",
				gdprEnabled: 0,
			},
			owlogger: `http://t.pubmatic.com/wl?json={"pubid":10,"pid":"1","pdvid":"0","dvc":{},"ft":0}&pubid=10`,
		},
		{
			name: "private endpoint",
			args: args{
				wlog: &WloggerRecord{
					record: record{
						PubID:     5,
						ProfileID: "5",
						VersionID: "1",
					},
				},
				loggerURL:   "http://10.172.141.11/wl",
				gdprEnabled: 0,
			},
			owlogger: `http://10.172.141.11/wl?json={"pubid":5,"pid":"5","pdvid":"1","dvc":{},"ft":0}&pubid=5`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owlogger := PrepareLoggerURL(tt.args.wlog, tt.args.loggerURL, tt.args.gdprEnabled)
			decodedOwlogger, _ := url.QueryUnescape(owlogger)
			assert.Equal(t, decodedOwlogger, tt.owlogger, tt.name)
		})
	}
}

func TestGetGdprEnabledFlag(t *testing.T) {
	tests := []struct {
		name          string
		partnerConfig map[int]map[string]string
		gdprFlag      int
	}{
		{
			name:          "Empty partnerConfig",
			partnerConfig: make(map[int]map[string]string),
			gdprFlag:      0,
		},
		{
			name: "partnerConfig without versionlevel cfg",
			partnerConfig: map[int]map[string]string{
				2: {models.GDPR_ENABLED: "1"},
			},
			gdprFlag: 0,
		},
		{
			name: "partnerConfig without GDPR_ENABLED",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {"any": "1"},
			},
			gdprFlag: 0,
		},
		{
			name: "partnerConfig with invalid GDPR_ENABLED",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {models.GDPR_ENABLED: "non-int"},
			},
			gdprFlag: 0,
		},
		{
			name: "partnerConfig with GDPR_ENABLED=1",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {models.GDPR_ENABLED: "1"},
			},
			gdprFlag: 1,
		},
		{
			name: "partnerConfig with GDPR_ENABLED=2",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {models.GDPR_ENABLED: "2"},
			},
			gdprFlag: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdprFlag := getGdprEnabledFlag(tt.partnerConfig)
			assert.Equal(t, tt.gdprFlag, gdprFlag, tt.name)
		})
	}
}
