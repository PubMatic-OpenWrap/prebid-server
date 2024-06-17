package config

import (
	"time"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/stats"
)

// Config contains the values read from the config file at boot time
type Config struct {
	Server        Server
	Database      Database
	Cache         Cache
	Timeout       Timeout
	Tracker       Tracker
	PixelView     PixelView
	Features      FeatureToggle
	Log           Log
	Stats         stats.Stats
	VastUnwrapCfg unWrapCfg.VastUnWrapCfg
	GeoDB         GeoDB
}

type Server struct {
	HostName string
	DCName   string //Name of the data center
}

type Database struct {
	Host string
	Port int

	Database string
	User     string
	Pass     string

	IdleConnection, MaxConnection, ConnMaxLifeTime, MaxDbContextTimeout int

	Queries Queries
}

/*
GetParterConfig query to get all partners and related configurations for a given pub,profile,version

Data is ordered by partnerId,keyname and entityId so that version level partner params will override the account level partner parasm in the code logic
*/
type Queries struct {
	GetParterConfig                   string
	DisplayVersionInnerQuery          string
	LiveVersionInnerQuery             string
	GetWrapperSlotMappingsQuery       string
	GetWrapperLiveVersionSlotMappings string
	GetPMSlotToMappings               string
	GetAdunitConfigQuery              string
	GetAdunitConfigForLiveVersion     string
	GetSlotNameHash                   string
	GetPublisherVASTTagsQuery         string
	GetAllDspFscPcntQuery             string
	GetPublisherFeatureMapQuery       string
	GetAnalyticsThrottlingQuery       string
	GetAdpodConfig                    string
	GetProfileTypePlatformQuery       string
	GetAppIntegrationPathQuery        string
	GetAppSubIntegrationPathQuery     string
}

type Cache struct {
	CacheConTimeout int // Connection timeout for cache

	CacheDefaultExpiry         int // in seconds
	VASTTagCacheExpiry         int // in seconds
	ProfileMetaDataCacheExpiry int // in seconds
}

type Timeout struct {
	MaxTimeout int64
	MinTimeout int64
	HBTimeout  int64
}

type Tracker struct {
	Endpoint                  string
	VideoErrorTrackerEndpoint string
}

type PixelView struct {
	OMScript string //js script path for conditional tracker call fire
}

type FeatureToggle struct {
	VASTUnwrapPercent             int
	VASTUnwrapStatsPercent        int
	AnalyticsThrottlingPercentage string
}

type Log struct { //Log Details
	LogPath            string
	LogLevel           int
	MaxLogSize         uint64
	MaxLogFiles        int
	LogRotationTime    time.Duration
	DebugLogUpdateTime time.Duration
	DebugAuthKey       string
}

type GeoDB struct {
	Location string
}
