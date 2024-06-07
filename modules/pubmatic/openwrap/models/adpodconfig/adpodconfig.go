package adpodconfig

type AdpodConfig struct {
	Dynamic    []Dynamic
	Structured []Structured
	Hybrid     []Hybrid
}

type Dynamic struct {
	PodDur      int   `json:"poddur,omitempty"`
	Maxseq      int   `json:"maxseq,omitempty"`
	MinDuration int   `json:"minduration,omitempty"`
	MaxDuration int   `json:"maxduration,omitempty"`
	Rqddurs     []int `json:"rqddurs,omitempty"`
}

type Structured struct {
	MinDuration int   `json:"minduration,omitempty"`
	MaxDuration int   `json:"maxduration,omitempty"`
	Rqddurs     []int `json:"rqddurs,omitempty"`
}

type Hybrid struct {
	PodDur      *int  `json:"poddur,omitempty"`
	Maxseq      *int  `json:"maxseq,omitempty"`
	MinDuration int   `json:"minduration,omitempty"`
	MaxDuration int   `json:"maxduration,omitempty"`
	Rqddurs     []int `json:"rqddurs,omitempty"`
}
