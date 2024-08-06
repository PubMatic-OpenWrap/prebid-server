package mysql

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		conn *sql.DB
		cfg  config.Database
	}
	tests := []struct {
		name  string
		args  args
		setup func() *sql.DB
	}{
		{
			name: "test",
			args: args{
				cfg: config.Database{},
			},
			setup: func() *sql.DB {
				db, _, _ := sqlmock.New()
				return db
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.conn = tt.setup()
			db := New(tt.args.conn, tt.args.cfg)
			assert.NotNil(t, db, "db should not be nil")
		})
	}
}

func TestUpdateQueriesWithMaxQueryExecutionTimeout(t *testing.T) {
	cfg := config.Database{
		MaxQueryExecutionTimeout: 30,
		Queries: config.Queries{
			GetParterConfig:                   "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM partner_config",
			DisplayVersionInnerQuery:          "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM display_version",
			LiveVersionInnerQuery:             "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM live_version",
			GetWrapperSlotMappingsQuery:       "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM wrapper_slot_mappings",
			GetWrapperLiveVersionSlotMappings: "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM wrapper_live_version_slot_mappings",
			GetPMSlotToMappings:               "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM pm_slot_to_mappings",
			GetAdunitConfigQuery:              "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM adunit_config",
			GetAdunitConfigForLiveVersion:     "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM adunit_config_live_version",
			GetSlotNameHash:                   "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM slot_name_hash",
			GetPublisherVASTTagsQuery:         "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM publisher_vast_tags",
			GetAllDspFscPcntQuery:             "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM all_dsp_fsc_pcnt",
			GetPublisherFeatureMapQuery:       "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM publisher_feature_map",
			GetAnalyticsThrottlingQuery:       "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM analytics_throttling",
			GetAdpodConfig:                    "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM adpod_config",
			GetProfileTypePlatformMapQuery:    "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM profile_type_platform_map",
			GetAppIntegrationPathMapQuery:     "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM app_integration_path_map",
			GetAppSubIntegrationPathMapQuery:  "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM app_sub_integration_path_map",
		},
	}

	actualQueries := updateQueriesWithMaxQueryExecutionTimeout(cfg)

	expectedQueries := config.Queries{
		GetParterConfig:                   "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM partner_config",
		DisplayVersionInnerQuery:          "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM display_version",
		LiveVersionInnerQuery:             "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM live_version",
		GetWrapperSlotMappingsQuery:       "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM wrapper_slot_mappings",
		GetWrapperLiveVersionSlotMappings: "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM wrapper_live_version_slot_mappings",
		GetPMSlotToMappings:               "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM pm_slot_to_mappings",
		GetAdunitConfigQuery:              "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM adunit_config",
		GetAdunitConfigForLiveVersion:     "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM adunit_config_live_version",
		GetSlotNameHash:                   "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM slot_name_hash",
		GetPublisherVASTTagsQuery:         "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM publisher_vast_tags",
		GetAllDspFscPcntQuery:             "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM all_dsp_fsc_pcnt",
		GetPublisherFeatureMapQuery:       "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM publisher_feature_map",
		GetAnalyticsThrottlingQuery:       "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM analytics_throttling",
		GetAdpodConfig:                    "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM adpod_config",
		GetProfileTypePlatformMapQuery:    "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM profile_type_platform_map",
		GetAppIntegrationPathMapQuery:     "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM app_integration_path_map",
		GetAppSubIntegrationPathMapQuery:  "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM app_sub_integration_path_map",
	}

	// Compare the results
	assert.Equal(t, expectedQueries, actualQueries)
}
