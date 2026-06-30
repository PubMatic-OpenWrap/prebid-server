package models

import "encoding/json"

// ResolvedEds holds flattened PubMatic-only enrichment parameters carried in
// ext.prebid.bidderparams.{pubmatic}.eds until the PubMatic adapter merges them.
type ResolvedEds struct {
	Device json.RawMessage `json:"device,omitempty"`
	App    json.RawMessage `json:"app,omitempty"`
}

func (r ResolvedEds) IsEmpty() bool {
	return len(r.Device) == 0 && len(r.App) == 0
}
