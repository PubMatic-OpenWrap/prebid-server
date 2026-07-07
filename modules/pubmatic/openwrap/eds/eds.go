package eds

import (
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

const bidderParamsEdsKey = "eds"

// ResolveEds extracts device.ext.eds and app.ext.eds from signal first, then the main request.
func ResolveEds(signal, request *openrtb2.BidRequest) models.ResolvedEds {
	return ResolveEdsWithDebug(signal, request, "")
}

func ResolveEdsWithDebug(signal, request *openrtb2.BidRequest, wiid string) models.ResolvedEds {
	if resolved := resolveEdsFromRequestWithDebug(signal, wiid, "signal"); !resolved.IsEmpty() {
		return resolved
	}
	return resolveEdsFromRequestWithDebug(request, wiid, "request")
}

// BuildPubmaticEdsBidderParams places resolved EDS under ext.prebid.bidderparams.{bidder}.eds
// for PubMatic (and PubMatic-core aliases). Signal ext.eds takes priority over the main request.
func BuildPubmaticEdsBidderParams(bidderParams json.RawMessage, signal, request *openrtb2.BidRequest, bidderCodes ...string) (json.RawMessage, models.ResolvedEds, error) {
	return BuildPubmaticEdsBidderParamsWithDebug(bidderParams, signal, request, "", bidderCodes...)
}

// BuildPubmaticEdsBidderParamsWithDebug is the same as BuildPubmaticEdsBidderParams but emits temporary ULP debug logs when wiid is non-empty.
func BuildPubmaticEdsBidderParamsWithDebug(bidderParams json.RawMessage, signal, request *openrtb2.BidRequest, wiid string, bidderCodes ...string) (json.RawMessage, models.ResolvedEds, error) {
	resolved := ResolveEdsWithDebug(signal, request, wiid)
	if resolved.IsEmpty() || len(bidderCodes) == 0 {
		if wiid != "" {
			LogNote(wiid, "build_bidderparams_skip", "resolved empty or no bidder codes")
		}
		return bidderParams, resolved, nil
	}

	if wiid != "" {
		LogResolvedEds(wiid, "build_bidderparams_resolved", resolved)
		LogBidderParamsRaw(wiid, "build_bidderparams_in", bidderParams)
	}

	updated, err := injectIntoBidderParamsWithDebug(bidderParams, resolved, wiid, bidderCodes...)
	if wiid != "" {
		LogMarshalError(wiid, "build_bidderparams_inject", "injectIntoBidderParams", err)
		LogBidderParamsRaw(wiid, "build_bidderparams_out", updated)
	}
	return updated, resolved, err
}

// StripFromRequest removes device.ext.eds and app.ext.eds from the shared bid request
// so other bidders do not receive PubMatic-only EDS.
func StripFromRequest(req *openrtb2.BidRequest) {
	if req == nil {
		return
	}

	if req.Device != nil {
		req.Device.Ext = nilIfEmptyExt(jsonparser.Delete(req.Device.Ext, "eds"))
	}
	if req.App != nil {
		req.App.Ext = nilIfEmptyExt(jsonparser.Delete(req.App.Ext, "eds"))
	}
}

// StripFromDeviceCtx removes cached device.ext.eds so profile/device enrichment does not
// write EDS back onto the shared request after StripFromRequest.
func StripFromDeviceCtx(dvc *models.DeviceCtx) {
	if dvc == nil || dvc.Ext == nil {
		return
	}
	dvc.Ext.DeleteEds()
}

func resolveEdsFromRequest(req *openrtb2.BidRequest) models.ResolvedEds {
	return resolveEdsFromRequestWithDebug(req, "", "request")
}

func resolveEdsFromRequestWithDebug(req *openrtb2.BidRequest, wiid, source string) models.ResolvedEds {
	if req == nil {
		return models.ResolvedEds{}
	}

	resolved := models.ResolvedEds{}
	if req.Device != nil {
		resolved.Device = nestedObject(req.Device.Ext, "eds")
		if wiid != "" {
			LogNestedEdsSubslice(wiid, "resolve_eds_"+source+"_device", req.Device.Ext, resolved.Device)
		}
	}
	if req.App != nil {
		resolved.App = nestedObject(req.App.Ext, "eds")
		if wiid != "" {
			LogNestedEdsSubslice(wiid, "resolve_eds_"+source+"_app", req.App.Ext, resolved.App)
		}
	}
	return resolved
}

func injectIntoBidderParams(bidderParams json.RawMessage, resolved models.ResolvedEds, bidderCodes ...string) (json.RawMessage, error) {
	return injectIntoBidderParamsWithDebug(bidderParams, resolved, "", bidderCodes...)
}

func injectIntoBidderParamsWithDebug(bidderParams json.RawMessage, resolved models.ResolvedEds, wiid string, bidderCodes ...string) (json.RawMessage, error) {
	if wiid != "" {
		LogResolvedEds(wiid, "inject_before_marshal_resolved", resolved)
	}

	edsPayload, err := json.Marshal(resolved)
	if wiid != "" {
		LogMarshalError(wiid, "inject", "json.Marshal(resolved)", err)
		LogRawBytes(wiid, "inject", "edsPayload", edsPayload)
	}
	if err != nil {
		return bidderParams, err
	}

	paramsMap := make(map[string]map[string]json.RawMessage)
	if len(bidderParams) > 0 {
		if err := json.Unmarshal(bidderParams, &paramsMap); err != nil {
			if wiid != "" {
				LogMarshalError(wiid, "inject", "json.Unmarshal(bidderParams)", err)
			}
			return bidderParams, err
		}
	}

	for _, code := range bidderCodes {
		if paramsMap[code] == nil {
			paramsMap[code] = make(map[string]json.RawMessage)
		}
		paramsMap[code][bidderParamsEdsKey] = edsPayload
	}

	if wiid != "" {
		for code, params := range paramsMap {
			for key, val := range params {
				if len(val) == 0 || !json.Valid(val) {
					glogWarningEmptyKey(wiid, code, key, len(val))
				}
			}
		}
		ProbeMarshal(wiid, "inject", "paramsMap", paramsMap)
	}

	out, err := json.Marshal(paramsMap)
	if wiid != "" {
		LogMarshalError(wiid, "inject", "json.Marshal(paramsMap)", err)
	}
	return out, err
}

// nestedObject returns a deep copy of the raw JSON value at key when it is a non-empty object.
func nestedObject(ext []byte, key string) []byte {
	if len(ext) == 0 {
		return nil
	}

	value, dataType, _, err := jsonparser.Get(ext, key)
	if err != nil || dataType != jsonparser.Object || len(value) <= 2 {
		return nil
	}

	copied := make([]byte, len(value))
	copy(copied, value)
	return copied
}

func nilIfEmptyExt(ext []byte) []byte {
	if len(ext) == 0 {
		return nil
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(ext, &keys); err != nil || len(keys) == 0 {
		return nil
	}
	return ext
}
