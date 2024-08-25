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
			cfg.Queries = updateQueriesWithMaxDbContextTimeout(cfg)
			db = &mySqlDB{conn: conn, cfg: cfg}
		})
	return db
}

func updateQueriesWithMaxDbContextTimeout(cfg config.Database) config.Queries {
	return config.Queries{
		GetParterConfig:                   fmt.Sprintf(cfg.Queries.GetParterConfig, cfg.MaxDbContextTimeout),
		DisplayVersionInnerQuery:          fmt.Sprintf(cfg.Queries.DisplayVersionInnerQuery, cfg.MaxDbContextTimeout),
		LiveVersionInnerQuery:             fmt.Sprintf(cfg.Queries.LiveVersionInnerQuery, cfg.MaxDbContextTimeout),
		GetWrapperSlotMappingsQuery:       fmt.Sprintf(cfg.Queries.GetWrapperSlotMappingsQuery, cfg.MaxDbContextTimeout),
		GetWrapperLiveVersionSlotMappings: fmt.Sprintf(cfg.Queries.GetWrapperLiveVersionSlotMappings, cfg.MaxDbContextTimeout),
		GetPMSlotToMappings:               fmt.Sprintf(cfg.Queries.GetPMSlotToMappings, cfg.MaxDbContextTimeout),
		GetAdunitConfigQuery:              fmt.Sprintf(cfg.Queries.GetAdunitConfigQuery, cfg.MaxDbContextTimeout),
		GetAdunitConfigForLiveVersion:     fmt.Sprintf(cfg.Queries.GetAdunitConfigForLiveVersion, cfg.MaxDbContextTimeout),
		GetSlotNameHash:                   fmt.Sprintf(cfg.Queries.GetSlotNameHash, cfg.MaxDbContextTimeout),
		GetPublisherVASTTagsQuery:         fmt.Sprintf(cfg.Queries.GetPublisherVASTTagsQuery, cfg.MaxDbContextTimeout),
		GetAllDspFscPcntQuery:             fmt.Sprintf(cfg.Queries.GetAllDspFscPcntQuery, cfg.MaxDbContextTimeout),
		GetPublisherFeatureMapQuery:       fmt.Sprintf(cfg.Queries.GetPublisherFeatureMapQuery, cfg.MaxDbContextTimeout),
		GetAnalyticsThrottlingQuery:       fmt.Sprintf(cfg.Queries.GetAnalyticsThrottlingQuery, cfg.MaxDbContextTimeout),
		GetAdpodConfig:                    fmt.Sprintf(cfg.Queries.GetAdpodConfig, cfg.MaxDbContextTimeout),
		GetProfileTypePlatformMapQuery:    fmt.Sprintf(cfg.Queries.GetProfileTypePlatformMapQuery, cfg.MaxDbContextTimeout),
		GetAppIntegrationPathMapQuery:     fmt.Sprintf(cfg.Queries.GetAppIntegrationPathMapQuery, cfg.MaxDbContextTimeout),
		GetAppSubIntegrationPathMapQuery:  fmt.Sprintf(cfg.Queries.GetAppSubIntegrationPathMapQuery, cfg.MaxDbContextTimeout),
	}
}
