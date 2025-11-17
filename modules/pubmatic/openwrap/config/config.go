package config

import (
	"time"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/stats"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/wakanda"
)

// Config contains the values read from the config file at boot time
type Config struct {
	Server           Server
	Database         Database
	Cache            Cache
	Timeout          Timeout
	Tracker          Tracker
	PixelView        PixelView
	Features         FeatureToggle
	Log              Log
	Stats            stats.Stats
	VastUnwrapCfg    unWrapCfg.VastUnWrapCfg
	Wakanda          wakanda.Wakanda
	GeoDB            GeoDB
	BidCache         BidCache
	ResponseOverride ResponseOverride
}

type ResponseOverride struct {
	BidType []string
}

type BidCache struct {
	CacheClient config.HTTPClient    `mapstructure:"http_client_cache" json:"http_client_cache"`
	CacheURL    config.Cache         `mapstructure:"cache" json:"cache"`
	ExtCacheURL config.ExternalCache `mapstructure:"external_cache" json:"external_cache"`
}

type Server struct {
	HostName string
	DCName   string //Name of the data center
	Endpoint string
}

type Database struct {
	Host string
	Port int

	Database string
	User     string
	Pass     string

	IdleConnection, MaxConnection, ConnMaxLifeTime, MaxDbContextTimeout, CountryPartnerFilterMaxDbContextTimeout int

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
	GetProfileTypePlatformMapQuery    string
	GetAppIntegrationPathMapQuery     string
	GetAppSubIntegrationPathMapQuery  string
	GetGDPRCountryCodes               string
	GetBannerSizesQuery               string
	GetProfileAdUnitMultiFloors       string
	GetCountryPartnerFilteringData    string
	GetPerformanceDSPQuery            string
	GetInViewEnabledPublishersQuery   string
}

type Cache struct {
	CacheConTimeout int // Connection timeout for cache

	CacheDefaultExpiry                  int // in seconds
	VASTTagCacheExpiry                  int // in seconds
	ProfileMetaDataCacheExpiry          int // in seconds
	CountryPartnerFilterRefreshInterval time.Duration
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
	VASTUnwrapPercent                     int
	AnalyticsThrottlingPercentage         string
	AllowPartnerLevelThrottlingPercentage int
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
