package mysql

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
)

type mySqlDB struct {
	conn *sql.DB
	cfg  config.Database
}

var db *mySqlDB
var dbOnce sync.Once

func New(conn *sql.DB, cfg config.Database) *mySqlDB {
	dbOnce.Do(
		func() {
			cfg.Queries = updateQueriesWithMaxQueryExecutionTimeout(cfg)
			db = &mySqlDB{conn: conn, cfg: cfg}
		})
	return db
}

func updateQueriesWithMaxQueryExecutionTimeout(cfg config.Database) config.Queries {
	queries := config.Queries{
		GetParterConfig:                   fmt.Sprintf(cfg.Queries.GetParterConfig, cfg.MaxQueryExecutionTimeout),
		DisplayVersionInnerQuery:          fmt.Sprintf(cfg.Queries.DisplayVersionInnerQuery, cfg.MaxQueryExecutionTimeout),
		LiveVersionInnerQuery:             fmt.Sprintf(cfg.Queries.LiveVersionInnerQuery, cfg.MaxQueryExecutionTimeout),
		GetWrapperSlotMappingsQuery:       fmt.Sprintf(cfg.Queries.GetWrapperSlotMappingsQuery, cfg.MaxQueryExecutionTimeout),
		GetWrapperLiveVersionSlotMappings: fmt.Sprintf(cfg.Queries.GetWrapperLiveVersionSlotMappings, cfg.MaxQueryExecutionTimeout),
		GetPMSlotToMappings:               fmt.Sprintf(cfg.Queries.GetPMSlotToMappings, cfg.MaxQueryExecutionTimeout),
		GetAdunitConfigQuery:              fmt.Sprintf(cfg.Queries.GetAdunitConfigQuery, cfg.MaxQueryExecutionTimeout),
		GetAdunitConfigForLiveVersion:     fmt.Sprintf(cfg.Queries.GetAdunitConfigForLiveVersion, cfg.MaxQueryExecutionTimeout),
		GetSlotNameHash:                   fmt.Sprintf(cfg.Queries.GetSlotNameHash, cfg.MaxQueryExecutionTimeout),
		GetPublisherVASTTagsQuery:         fmt.Sprintf(cfg.Queries.GetPublisherVASTTagsQuery, cfg.MaxQueryExecutionTimeout),
		GetAllDspFscPcntQuery:             fmt.Sprintf(cfg.Queries.GetAllDspFscPcntQuery, cfg.MaxQueryExecutionTimeout),
		GetPublisherFeatureMapQuery:       fmt.Sprintf(cfg.Queries.GetPublisherFeatureMapQuery, cfg.MaxQueryExecutionTimeout),
		GetAnalyticsThrottlingQuery:       fmt.Sprintf(cfg.Queries.GetAnalyticsThrottlingQuery, cfg.MaxQueryExecutionTimeout),
		GetAdpodConfig:                    fmt.Sprintf(cfg.Queries.GetAdpodConfig, cfg.MaxQueryExecutionTimeout),
		GetProfileTypePlatformMapQuery:    fmt.Sprintf(cfg.Queries.GetProfileTypePlatformMapQuery, cfg.MaxQueryExecutionTimeout),
		GetAppIntegrationPathMapQuery:     fmt.Sprintf(cfg.Queries.GetAppIntegrationPathMapQuery, cfg.MaxQueryExecutionTimeout),
		GetAppSubIntegrationPathMapQuery:  fmt.Sprintf(cfg.Queries.GetAppSubIntegrationPathMapQuery, cfg.MaxQueryExecutionTimeout),
	}
	return queries
}
