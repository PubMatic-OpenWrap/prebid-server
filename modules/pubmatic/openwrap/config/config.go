package config

import (
	"encoding/json"
	"fmt"
	"time"
)

type FeatureToggle struct {
	// InlineHBEnabled feature flag controls phase 3.
	InlineHBEnabled bool
	//MultiCurrency support in SSHB
	MultiCurrency bool
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
}

type Cache struct {
	CacheConTimeout int // Connection timeout for cache

	CacheDefaultExpiry int // in seconds
	VASTTagCacheExpiry int // in seconds
}

// Config contains the values read from the config file at boot time
type SSHB struct {
	OpenWrap struct {
		Server struct { //Server Configuration
			ServerPort string //Listen Port
			DCName     string //Name of the data center
			HostName   string //Name of Server
		}

		Log struct { //Log Details
			LogPath            string
			LogLevel           int
			MaxLogSize         uint64
			MaxLogFiles        int
			LogRotationTime    time.Duration
			DebugLogUpdateTime time.Duration
			DebugAuthKey       string
		}

		Database Database

		Stats struct {
			UseHostName     bool // if true use actual_node_name:actual_pod_name into stats key
			DefaultHostName string
			//UDP parameters
			StatsHost           string
			StatsPort           string
			StatsTickerInterval int //in minutes
			CriticalThreshold   int
			CriticalInterval    int //in minutes
			StandardThreshold   int
			StandardInterval    int //in minutes
			//TCP parameters
			PortTCP                   string
			PublishInterval           int
			PublishThreshold          int
			Retries                   int
			DialTimeout               int
			KeepAliveDuration         int
			MaxIdleConnections        int
			MaxIdleConnectionsPerHost int
		}

		Logger struct {
			Enabled        bool
			Endpoint       string
			PublicEndpoint string
			MaxClients     int32
			MaxConnections int
			MaxCalls       int
			RespTimeout    int
		}

		Cache Cache

		Timeout struct {
			MaxTimeout          int64
			MinTimeout          int64
			PrebidDelta         int64
			HBTimeout           int64
			CacheConTimeout     int64 // Connection timeout for cache
			MaxQueryTimeout     int64 // max_execution time for db query
			MaxDbContextTimeout int64 // context timeout for db query
		}

		Tracker struct {
			Endpoint                  string
			VideoErrorTrackerEndpoint string
		}

		Pixelview struct {
			OMScript string //js script path for conditional tracker call fire
		}

		BidderParamMapping map[string]map[string]*ParameterMapping `json:"bidder_param_mapping"`
	}

	Cache struct {
		Host   string
		Scheme string
		Query  string
	}

	Metrics struct {
		Prometheus struct {
			Enabled                   bool
			UseSeparateServerInstance bool
			Port                      int
			ExposePrebidMetrics       bool
			TimeoutMillisRaw          int
			HBNamespace               string
			HBSubsystem               string
		}
	}

	Features FeatureToggle

	Analytics struct {
		Pubmatic struct {
			Enabled bool
		}
	}
}

func (cfg *SSHB) String() string {
	jsonBytes, err := json.Marshal(cfg)

	if nil != err {
		return err.Error()
	}

	return string(jsonBytes[:])
}

func (cfg *SSHB) Validate() (err error) {
	if cfg.OpenWrap.Server.ServerPort == "" {
		return fmt.Errorf("Listen Port Not Specified")
	}

	if cfg.OpenWrap.Stats.StatsTickerInterval >= cfg.OpenWrap.Stats.CriticalInterval {
		return fmt.Errorf("StatsTickerInterval should be less than CriticalInterval")
	}

	if cfg.Metrics.Prometheus.ExposePrebidMetrics || cfg.Metrics.Prometheus.UseSeparateServerInstance {
		if cfg.Metrics.Prometheus.Port <= 0 {
			return fmt.Errorf("value of port should be non-zero(e.g 8002) when value of ExposePrebidMetrics = true OR UseSeparateServerInstance = true")
		}
	}

	if !cfg.Metrics.Prometheus.ExposePrebidMetrics && !cfg.Metrics.Prometheus.UseSeparateServerInstance {
		if cfg.Metrics.Prometheus.Port > 0 {
			return fmt.Errorf("value of port should be zero when value of ExposePrebidMetrics = false OR UseSeparateServerInstance = false")
		}
	}

	return nil
}

// ParameterMapping holds mapping information for bidder parameter
type ParameterMapping struct {
	BidderParamName string      `json:"bidderParameterName,omitempty"`
	KeyName         string      `json:"keyName,omitempty"`
	Datatype        string      `json:"type,omitempty"`
	Required        bool        `json:"required,omitempty"`
	DefaultValue    interface{} `json:"defaultValue,omitempty"`
}
