package eds

import (
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

const bidderParamsEdsKey = "eds"

// Sources holds bid request objects used to resolve EDS parameters from ext.eds only.
// Signal is used for auction SDK integrations; Request is used for v25 / standard OW SDK.
type Sources struct {
	Signal  *openrtb2.BidRequest
	Request *openrtb2.BidRequest
}

// Resolve flattens ext.eds from signal and/or request into PubMatic-only ext objects.
// When both are present, signal values take priority and request fills missing keys.
func Resolve(src Sources) models.ResolvedEds {
	var resolved models.ResolvedEds

	if src.Signal != nil {
		resolved = resolveFromBidRequest(src.Signal)
	}
	if src.Request != nil {
		requestResolved := resolveFromBidRequest(src.Request)
		if src.Signal != nil {
			resolved = MergeGapFill(resolved, requestResolved)
		} else {
			resolved = requestResolved
		}
	}

	return resolved
}

// MergeGapFill returns base with keys from overlay filled only where base is missing.
func MergeGapFill(base, overlay models.ResolvedEds) models.ResolvedEds {
	if overlay.IsEmpty() {
		return base
	}
	if base.IsEmpty() {
		return overlay
	}

	base.Device = mergeExtJSON(base.Device, overlay.Device, false)
	base.App = mergeExtJSON(base.App, overlay.App, false)
	base.Imp = mergeImpExtMaps(base.Imp, overlay.Imp, false)

	return base
}

// InjectIntoBidderParams stores resolved EDS under ext.prebid.bidderparams.{bidder}.eds.
func InjectIntoBidderParams(bidderParams json.RawMessage, resolved models.ResolvedEds, bidderCodes ...string) (json.RawMessage, error) {
	if resolved.IsEmpty() || len(bidderCodes) == 0 {
		return bidderParams, nil
	}

	edsPayload, err := json.Marshal(resolved)
	if err != nil {
		return bidderParams, err
	}

	paramsMap := make(map[string]map[string]json.RawMessage)
	if len(bidderParams) > 0 {
		if err := json.Unmarshal(bidderParams, &paramsMap); err != nil {
			return bidderParams, err
		}
	}

	for _, code := range bidderCodes {
		if paramsMap[code] == nil {
			paramsMap[code] = make(map[string]json.RawMessage)
		}
		paramsMap[code][bidderParamsEdsKey] = edsPayload
	}

	return json.Marshal(paramsMap)
}

// ExtractFromBidderParams reads EDS stored under ext.prebid.bidderparams.eds on a
// per-bidder request (after PBS exchange filtering).
func ExtractFromBidderParams(bidderParams json.RawMessage) models.ResolvedEds {
	if len(bidderParams) == 0 {
		return models.ResolvedEds{}
	}

	var params map[string]json.RawMessage
	if err := json.Unmarshal(bidderParams, &params); err != nil {
		return models.ResolvedEds{}
	}

	edsRaw, ok := params[bidderParamsEdsKey]
	if !ok || len(edsRaw) == 0 {
		return models.ResolvedEds{}
	}

	var resolved models.ResolvedEds
	if err := json.Unmarshal(edsRaw, &resolved); err != nil {
		return models.ResolvedEds{}
	}
	return resolved
}

// StripFromRequest removes ext.eds wrappers and resolved flat ext keys
// from the shared bid request so other bidders do not receive them.
func StripFromRequest(req *openrtb2.BidRequest, resolved models.ResolvedEds) {
	if req == nil {
		return
	}

	if req.Device != nil {
		req.Device.Ext = stripObjectExt(req.Device.Ext, resolved.Device)
	}
	if req.App != nil {
		req.App.Ext = stripObjectExt(req.App.Ext, resolved.App)
	}
	for i := range req.Imp {
		if resolvedExt, ok := resolved.Imp[req.Imp[i].ID]; ok {
			req.Imp[i].Ext = stripObjectExt(req.Imp[i].Ext, resolvedExt)
		} else {
			req.Imp[i].Ext = jsonparser.Delete(req.Imp[i].Ext, "eds")
		}
	}
}

// ApplyToRequest merges resolved flat ext keys onto the bid request.
func ApplyToRequest(req *openrtb2.BidRequest, resolved models.ResolvedEds) {
	if req == nil || resolved.IsEmpty() {
		return
	}

	if len(resolved.Device) > 0 {
		if req.Device == nil {
			req.Device = &openrtb2.Device{}
		} else {
			deviceCopy := *req.Device
			req.Device = &deviceCopy
		}
		req.Device.Ext = mergeExtJSON(req.Device.Ext, resolved.Device, true)
	}

	if len(resolved.App) > 0 {
		if req.App == nil {
			req.App = &openrtb2.App{}
		} else {
			appCopy := *req.App
			req.App = &appCopy
		}
		req.App.Ext = mergeExtJSON(req.App.Ext, resolved.App, true)
	}

	if len(resolved.Imp) == 0 {
		return
	}

	impNeedsCopy := false
	for i := range req.Imp {
		if resolvedExt, ok := resolved.Imp[req.Imp[i].ID]; ok && len(resolvedExt) > 0 {
			impNeedsCopy = true
			break
		}
	}
	if !impNeedsCopy {
		return
	}

	newImps := make([]openrtb2.Imp, len(req.Imp))
	copy(newImps, req.Imp)
	req.Imp = newImps

	for i := range req.Imp {
		if resolvedExt, ok := resolved.Imp[req.Imp[i].ID]; ok && len(resolvedExt) > 0 {
			req.Imp[i].Ext = mergeExtJSON(req.Imp[i].Ext, resolvedExt, true)
		}
	}
}

func resolveFromBidRequest(req *openrtb2.BidRequest) models.ResolvedEds {
	if req == nil {
		return models.ResolvedEds{}
	}

	resolved := models.ResolvedEds{
		Imp: make(map[string]json.RawMessage),
	}

	if req.Device != nil {
		resolved.Device = flattenEdsObject(req.Device.Ext)
	}
	if req.App != nil {
		resolved.App = flattenEdsObject(req.App.Ext)
	}
	for _, imp := range req.Imp {
		if ext := flattenEdsObject(imp.Ext); len(ext) > 0 {
			resolved.Imp[imp.ID] = ext
		}
	}

	if len(resolved.Imp) == 0 {
		resolved.Imp = nil
	}

	return resolved
}

func flattenEdsObject(ext []byte) json.RawMessage {
	if len(ext) == 0 {
		return nil
	}

	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(ext, &wrapper); err != nil {
		return nil
	}

	edsObj, ok := wrapper["eds"]
	if !ok || len(edsObj) == 0 || string(edsObj) == "null" {
		return nil
	}

	var flat map[string]json.RawMessage
	if err := json.Unmarshal(edsObj, &flat); err != nil || len(flat) == 0 {
		return nil
	}

	out, err := json.Marshal(flat)
	if err != nil {
		return nil
	}
	return out
}

func mergeExtJSON(base, overlay json.RawMessage, overlayWins bool) json.RawMessage {
	if len(overlay) == 0 {
		return base
	}
	if len(base) == 0 {
		return overlay
	}

	var baseMap map[string]json.RawMessage
	if err := json.Unmarshal(base, &baseMap); err != nil || len(baseMap) == 0 {
		return overlay
	}

	var overlayMap map[string]json.RawMessage
	if err := json.Unmarshal(overlay, &overlayMap); err != nil || len(overlayMap) == 0 {
		return base
	}

	for key, val := range overlayMap {
		if !overlayWins {
			if _, exists := baseMap[key]; exists {
				continue
			}
		}
		baseMap[key] = val
	}

	out, err := json.Marshal(baseMap)
	if err != nil {
		return base
	}
	return out
}

func mergeImpExtMaps(base, overlay map[string]json.RawMessage, overlayWins bool) map[string]json.RawMessage {
	if len(overlay) == 0 {
		return base
	}
	if base == nil {
		base = make(map[string]json.RawMessage, len(overlay))
	}
	for impID, overlayExt := range overlay {
		base[impID] = mergeExtJSON(base[impID], overlayExt, overlayWins)
	}
	return base
}

func stripObjectExt(ext []byte, resolvedExt json.RawMessage) []byte {
	if len(ext) == 0 {
		return ext
	}

	ext = jsonparser.Delete(ext, "eds")
	if len(resolvedExt) == 0 {
		return ext
	}

	var resolved map[string]json.RawMessage
	if err := json.Unmarshal(resolvedExt, &resolved); err != nil {
		return ext
	}
	for key := range resolved {
		ext = jsonparser.Delete(ext, key)
	}
	return ext
}
