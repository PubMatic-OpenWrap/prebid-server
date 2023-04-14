package openrtb_ext

// LogInfo contains the logger, tracker calls to be sent in response
type LogInfo struct {
	Logger  string `json:"logger,omitempty"`
	Tracker string `json:"tracker,omitempty"`
}
