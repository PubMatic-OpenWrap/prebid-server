package metrics

// Labels defines the labels that can be attached to the metrics.
type Labels struct {
	RType         RequestType
	RequestStatus RequestStatus
}

// RequestType : Request type enumeration
type RequestType string

// RequestStatus : The request return status
type RequestStatus string

// LurlStatusLabels defines labels applicable for LURL sent
type LurlStatusLabels struct {
	PublisherID string
	Partner     string
	Status      string
}

// LurlBatchStatusLabels defines labels applicable for LURL batche sent
type LurlBatchStatusLabels struct {
	Status string
}
