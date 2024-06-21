package openrtb_ext

type ExtBidCMOnsite struct {
	AdType   int    `json:"adtype,omitempty"`
	ViewUrl  string `json:"vurl,omitempty"`
	ClickUrl string `json:"curl,omitempty"`
}
