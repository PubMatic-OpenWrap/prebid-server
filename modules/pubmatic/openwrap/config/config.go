package config

import (
	"encoding/json"
	"time"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
)

// Config contains the values read from the config file at boot time
type SSHB struct {
	OpenWrap struct {
		Vastunwrap unWrapCfg.VastUnWrapCfg

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
