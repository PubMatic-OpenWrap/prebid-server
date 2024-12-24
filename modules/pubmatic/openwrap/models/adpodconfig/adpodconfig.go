package adpodconfig

type AdpodConfig struct {
	Dynamic    []Dynamic
	Structured []Structured
	Hybrid     []Hybrid
}

type Dynamic struct {
	PodDur      int64   `json:"poddur,omitempty"`
	MaxSeq      int64   `json:"maxseq,omitempty"`
	MinDuration int64   `json:"minduration,omitempty"`
	MaxDuration int64   `json:"maxduration,omitempty"`
	RqdDurs     []int64 `json:"rqddurs,omitempty"`
}

type Structured struct {
	MinDuration int64   `json:"minduration,omitempty"`
	MaxDuration int64   `json:"maxduration,omitempty"`
	RqdDurs     []int64 `json:"rqddurs,omitempty"`
}

type Hybrid struct {
	PodDur      *int64  `json:"poddur,omitempty"`
	MaxSeq      *int64  `json:"maxseq,omitempty"`
	MinDuration int64   `json:"minduration,omitempty"`
	MaxDuration int64   `json:"maxduration,omitempty"`
	RqdDurs     []int64 `json:"rqddurs,omitempty"`
}
