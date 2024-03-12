package stats

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
