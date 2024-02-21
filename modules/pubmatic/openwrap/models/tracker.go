package models

// OWTracker vast video parameters to be injected
type OWTracker struct {
	Tracker       Tracker
	TrackerURL    string
	ErrorURL      string
	Price         float64
	PriceModel    string
	PriceCurrency string
	BidType       string `json:"-"` // video, banner, native
	DspId         int    `json:"-"` // dsp id
}

// Tracker tracker url creation parameters
type Tracker struct {
	PubID             int
	PageURL           string
	Timestamp         int64
	IID               string
	ProfileID         string
	VersionID         string
	SlotID            string
	Adunit            string
	PartnerInfo       Partner
	RewardedInventory int
	SURL              string // contains either req.site.domain or req.app.bundle value
	Platform          int
	// SSAI identifies the name of the SSAI vendor
	// Applicable only in case of incase of video/json endpoint.
	SSAI              string
	AdPodSlot         int
	TestGroup         int
	Origin            string
	FloorSkippedFlag  *int
	FloorModelVersion string
	FloorSource       *int
	FloorType         int
	CustomDimensions  string
	ATTS              *float64
	LoggerData        LoggerData // need this in logger to avoid duplicate computation

	ImpID  string `json:"-"`
	Secure int    `json:"-"`
}

// Partner partner information to be logged in tracker object
type Partner struct {
	PartnerID      string
	BidderCode     string
	KGPV           string
	GrossECPM      float64
	NetECPM        float64
	BidID          string
	OrigBidID      string
	AdSize         string
	AdDuration     int
	Adformat       string
	ServerSide     int
	Advertiser     string
	FloorValue     float64
	FloorRuleValue float64
	DealID         string
}

// LoggerData: this data to be needed in logger
type LoggerData struct {
	KGPSV            string
	FloorProvider    string
	FloorFetchStatus *int
}

// FloorsDetails contains floors info derived from responseExt.Prebid.Floors
type FloorsDetails struct {
	FloorType         int
	FloorModelVersion string
	FloorProvider     string
	Skipfloors        *int
	FloorFetchStatus  *int
	FloorSource       *int
}
