package openrtb_ext

type ExtBidCMOnsite struct {
	ViewUrl  string `json:"vurl,omitempty"`
	ClickUrl string `json:"curl,omitempty"`
}

// ImpExtensionCommerce - Impression Commerce Extension
type ExtImpCMOnsiteParams struct {
	Adtype []int `json:"adtype,omitempty"`
}

// ImpExtensionCommerce - Impression Commerce Extension
type ExtImpCMOnsitePrebid struct {
	ComParams *CMOnsiteImpExtPrebidParams `json:"commerce,omitempty"`
}

type CMOnsiteImpExtPrebidParams struct {
	ExtImpCMOnsiteParams
}

type CMOnsiteInventoryDetails struct {
	AdbulterZoneID int
	Adtype         string
	Width          int
	Height         int
}

type ExtRequestOnsiteParams struct {
	Sequence  int                `json:"seq,omitempty"`
	Targeting []*ExtImpTargeting `json:"targeting,omitempty"`
}

type ExtRequestPrebidOnsite struct {
	ExtRequestOnsiteParams
	ZoneMapping      map[string]interface{} `json:"mapping,omitempty"`
	ReportingKeys    map[string]interface{} `json:"reporting,omitempty"`
	CustomParams     map[string]interface{} `json:"customParams,omitempty"`
	DsConsentApplies interface{}            `json:"ds_consent_applies,omitempty"`
	DsConsentGiven   interface{}            `json:"ds_consent_given,omitempty"`
	UserID           string                 `json:"userId,omitempty"`
	GeoCountry       string                 `json:"geoCountry,omitempty"`
	DeviceType       int                    `json:"deviceType,omitempty"`
}

type ExtBidderCMOnsite struct {
	PrebidBidderName string `json:"prebidname,omitempty"`
	BidderCode       string `json:"biddercode,omitempty"`
}
