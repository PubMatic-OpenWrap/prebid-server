package models

const (
	// Device
	GAMDeviceIFA      = "rdid"   // OW-OTT equivalent dev.ifa
	GAMDevicetype     = "idtype" // OW-OTT equivalent dev.devicetype
	GAMDeviceLanguage = "hl"     // OW-OTT equivalent dev.language

	//App
	GAMAppID       = "msid" // OW-OTT equivalent app.id
	GAMAppBundle   = "an"   // OW-OTT equivalent  app.bundle
	GAMAppStoreUrl = "url"  // OW-OTT equivalent app.storeurl

	// imp.video
	GAMVideoMinDuration = "pmnd"     // OW-OTT equivalent imp.vid.minduration
	GAMVideoMaxDuration = "pmxd"     // OW-OTT equivalent imp.vid.maxduration
	GAMVideoLinearity   = "vad_type" // OW-OTT equivalent imp.vid.linearity
	GAMVideoDimensions  = "sz"       // OW-OTT equivalent imp.vid.w && imp.vid.h
	GAMVideoLinear      = "linear"
	GAMVideoNonLinear   = "nonlinear"

	//Adpod
	GAMAdpodMaxAds   = "pmad"            // OW-OTT equivalent imp.vid.ext.adpod.maxads
	GAMAdMinDuration = "min_ad_duration" // OW-OTT equivalent imp.vid.ext.adpod.adminduration
	GAMAdMaxDuration = "max_ad_duration" // OW-OTT equivalent imp.vid.ext.adpod.admaxduration
)
