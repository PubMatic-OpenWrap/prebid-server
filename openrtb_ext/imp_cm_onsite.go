package openrtb_ext

type ExtBidCMOnsite struct {
	AdType   int    `json:"adtype,omitempty"`
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
	ZoneMapping map[string]interface{} `json:"mapping,omitempty"`
}

type ExtBidderCMOnsite struct {
	PrebidBidderName string `json:"prebidname,omitempty"`
	BidderCode       string `json:"biddercode,omitempty"`
}
