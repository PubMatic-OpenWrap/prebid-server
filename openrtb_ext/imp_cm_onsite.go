package openrtb_ext


// ImpExtensionCommerce - Impression Commerce Extension
type ExtImpCMOnsiteParams struct {
	SlotsRequested int                `json:"slots_requested,omitempty"`
}

// ImpExtensionCommerce - Impression Commerce Extension
type ExtImpCMOnsitePrebid struct {
	ComParams *CMOnsiteImpExtPrebidParams `json:"commerce,omitempty"`
}

type CMOnsiteImpExtPrebidParams struct {
	ExtImpCMOnsiteParams
}

type ExtRequestOnsiteParams struct {
	Sequence int               		   `json:"seq,omitempty"`
	Targeting      []*ExtImpTargeting `json:"targeting,omitempty"`
}

type ExtRequestPrebidOnsite struct {
	ExtRequestOnsiteParams
	ZoneMapping     map[string]string    `json:"mapping,omitempty"`
}

type ExtBidderCMOnsite struct {
	PrebidBidderName string              `json:"prebidname,omitempty"`
	BidderCode       string              `json:"biddercode,omitempty"`
}
