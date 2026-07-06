package models

const (
	ExtEDSKey         = "eds"
	PubmaticBidderKey = "pubmatic"
	EDSDeviceKey      = "device"
	EDSAppKey         = "app"

	// EDSBlockedCountriesAll disables EDS params for all countries when used as feature value.
	EDSBlockedCountriesAll = "*"
)

// DeviceEDSTier1BlockedParams are stripped from bidderparams.pubmatic.eds.device for EDS blocked countries.
var DeviceEDSTier1BlockedParams = []string{
	"boottime",
	"diskspace",
	"totaldisk",
	"inputlaunguage",
	"totalmem",
}

// AppEDSTier1BlockedParams are stripped from bidderparams.pubmatic.eds.app for EDS blocked countries.
var AppEDSTier1BlockedParams = []string{
	"install_time",
	"first_launch_time",
}
