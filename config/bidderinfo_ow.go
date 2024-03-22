package config

type RequestMode string

const (
	RequestModeSingle RequestMode = "single"
)

// OpenWrap stores the openwrap specific bidder-info configuration
type OpenWrap struct {
	// requestMode specifies whether bidder supports single/multi impressions per HTTP request
	// requestMode="single" means bidder supports only one impression per request
	// if this parameter is not specified then we assumes that bidder supports multi impression
	RequestMode RequestMode `yaml:"requestMode" mapstructure:"requestMode"`
}
