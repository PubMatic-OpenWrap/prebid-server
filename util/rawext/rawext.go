package rawext

import (
	"bytes"
	"encoding/json"

	"github.com/prebid/openrtb/v20/openrtb2"
)

// CloneRawMessage returns an owned copy of b, or nil when empty.
func CloneRawMessage(b json.RawMessage) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	copied := make(json.RawMessage, len(b))
	copy(copied, b)
	return copied
}

// CloneBidderParamsMap returns a deep copy of bidder params so map values are not subslices of request.ext.
func CloneBidderParamsMap(in map[string]map[string]json.RawMessage) map[string]map[string]json.RawMessage {
	if in == nil {
		return nil
	}
	out := make(map[string]map[string]json.RawMessage, len(in))
	for bidder, params := range in {
		cloned := make(map[string]json.RawMessage, len(params))
		for k, v := range params {
			cloned[k] = CloneRawMessage(v)
		}
		out[bidder] = cloned
	}
	return out
}

// RepairExt returns an owned copy when b is valid JSON, or nil when empty or unrecoverable.
func RepairExt(b json.RawMessage) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	if json.Valid(b) {
		return CloneRawMessage(b)
	}
	compacted := bytes.NewBuffer(make([]byte, 0, len(b)))
	if err := json.Compact(compacted, b); err == nil && json.Valid(compacted.Bytes()) {
		return CloneRawMessage(compacted.Bytes())
	}
	return nil
}

// NormalizeBidRequestExts replaces ext fields with owned copies of valid JSON before marshaling.
func NormalizeBidRequestExts(req *openrtb2.BidRequest) {
	if req == nil {
		return
	}
	req.Ext = RepairExt(req.Ext)
	if req.App != nil {
		req.App.Ext = RepairExt(req.App.Ext)
	}
	if req.Device != nil {
		req.Device.Ext = RepairExt(req.Device.Ext)
	}
	if req.User != nil {
		req.User.Ext = RepairExt(req.User.Ext)
	}
	if req.Source != nil {
		req.Source.Ext = RepairExt(req.Source.Ext)
	}
	if req.Regs != nil {
		req.Regs.Ext = RepairExt(req.Regs.Ext)
	}
	for i := range req.Imp {
		req.Imp[i].Ext = RepairExt(req.Imp[i].Ext)
	}
}
