package config

type OpenWrapConfig struct {
	// OpenWrap Configurations
	EnableFastXML       bool                `mapstructure:"enable_fast_xml"`
	TrackerURL          string              `mapstructure:"tracker_url"`
	VendorListScheduler VendorListScheduler `mapstructure:"vendor_list_scheduler"`
	PriceFloorFetcher   PriceFloorFetcher   `mapstructure:"price_floor_fetcher"`
}

type PriceFloorFetcher struct {
	HttpClient HTTPClient `mapstructure:"http_client"`
	CacheSize  int        `mapstructure:"cache_size_mb"`
	Worker     int        `mapstructure:"worker"`
	Capacity   int        `mapstructure:"capacity"`
	MaxRetries int        `mapstructure:"max_retries"`
}

type VendorListScheduler struct {
	Enabled  bool   `mapstructure:"enabled"`
	Interval string `mapstructure:"interval"`
	Timeout  string `mapstructure:"timeout"`
}
