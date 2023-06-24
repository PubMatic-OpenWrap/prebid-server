package config

import "time"

// Config contains the values read from the config file at boot time
type Config struct {
	Server    Server
	Database  Database
	Cache     Cache
	Timeout   Timeout
	Tracker   Tracker
	PixelView PixelView
	Features  FeatureToggle
	Log       Log
	Stats     Stats
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
	GetAllFscDisabledPublishersQuery  string
	GetAllDspFscPcntQuery             string
}

type Cache struct {
	CacheConTimeout int // Connection timeout for cache

	CacheDefaultExpiry int // in seconds
	VASTTagCacheExpiry int // in seconds
}

type Timeout struct {
	MaxTimeout int64
	MinTimeout int64
}

type Tracker struct {
	Endpoint                  string
	VideoErrorTrackerEndpoint string
}

type PixelView struct {
	OMScript string //js script path for conditional tracker call fire
}

type FeatureToggle struct {
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

type Stats struct {
	UseHostName               bool   // if true use actual_node_name:actual_pod_name into stats key
	DefaultHostName           string // combination of node:pod, default value is N:P
	Endpoint                  string // stat-server's endpoint
	PublishInterval           int    // interval (in minutes) to publish stats to server
	PublishThreshold          int    // publish stats if number of stat-records present in map is higher than this threshold
	Retries                   int    // max retries to publish stats to server
	DialTimeout               int    // http connection dial-timeout (in seconds)
	KeepAliveDuration         int    // http connection keep-alive-duration (in minutes)
	MaxIdleConnections        int    // maximum idle connections across all hosts
	MaxIdleConnectionsPerHost int    // maximum idle connections per host
	ResponseHeaderTimeout     int    // amount of time (in seconds) to wait for server's response header
	MaxChannelLength          int    // max number of allowed stat keys
	PoolMaxWorkers            int    // max number of workers that will actually send the data to stats-server
	PoolMaxCapacity           int    // number of tasks that can be hold by the pool
}
