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

func TestUpdateQueriesWithMaxDbContextTimeout(t *testing.T) {
	cfg := config.Database{
		MaxDbContextTimeout: 30,
		Queries: config.Queries{
			GetParterConfig:                   "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			DisplayVersionInnerQuery:          "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			LiveVersionInnerQuery:             "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetWrapperSlotMappingsQuery:       "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetWrapperLiveVersionSlotMappings: "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetPMSlotToMappings:               "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetAdunitConfigQuery:              "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetAdunitConfigForLiveVersion:     "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetSlotNameHash:                   "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetPublisherVASTTagsQuery:         "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetAllDspFscPcntQuery:             "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetPublisherFeatureMapQuery:       "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetAnalyticsThrottlingQuery:       "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetAdpodConfig:                    "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetProfileTypePlatformMapQuery:    "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetAppIntegrationPathMapQuery:     "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
			GetAppSubIntegrationPathMapQuery:  "SELECT /*+ MAX_EXECUTION_TIME(%d) */ data FROM some_table ",
		},
	}

	actualQueries := updateQueriesWithMaxDbContextTimeout(cfg)

	expectedQueries := config.Queries{
		GetParterConfig:                   "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		DisplayVersionInnerQuery:          "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		LiveVersionInnerQuery:             "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetWrapperSlotMappingsQuery:       "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetWrapperLiveVersionSlotMappings: "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetPMSlotToMappings:               "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetAdunitConfigQuery:              "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetAdunitConfigForLiveVersion:     "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetSlotNameHash:                   "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetPublisherVASTTagsQuery:         "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetAllDspFscPcntQuery:             "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetPublisherFeatureMapQuery:       "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetAnalyticsThrottlingQuery:       "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetAdpodConfig:                    "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetProfileTypePlatformMapQuery:    "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetAppIntegrationPathMapQuery:     "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
		GetAppSubIntegrationPathMapQuery:  "SELECT /*+ MAX_EXECUTION_TIME(30) */ data FROM some_table ",
	}

	// Compare the results
	assert.Equal(t, expectedQueries, actualQueries)
}
