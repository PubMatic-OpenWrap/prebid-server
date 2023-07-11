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
}

type Server struct {
	HostName string
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
