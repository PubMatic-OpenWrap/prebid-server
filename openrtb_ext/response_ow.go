package openrtb_ext

// OwLogInfo contains the logger, tracker calls to be sent in response
type OwLogInfo struct {
	Logger  string `json:"logger,omitempty"`
	Tracker string `json:"tracker,omitempty"`
}
