package eds

import (
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"
)

const bidderParamsEdsKey = "eds"

// ResolveEds extracts *.ext.eds (device/app/user/imp) from signal first, then the main request.
// When EDS comes from signal, imp EDS keys are remapped onto request imp IDs by index so the
// PubMatic adapter can match session_count onto the auction request imps (signal/request IDs often differ).
func ResolveEds(signal, request *openrtb2.BidRequest) models.ResolvedEds {
	if resolved := resolveEdsFromRequest(signal); !resolved.IsEmpty() {
		return remapSignalImpEdsToRequestIDs(resolved, signal, request)
	}
	return resolveEdsFromRequest(request)
}

// BuildPubmaticEdsBidderParams places resolved EDS under ext.prebid.bidderparams.{bidder}.eds
// for PubMatic (and PubMatic-core aliases). Signal ext.eds takes priority over the main request.
func BuildPubmaticEdsBidderParams(bidderParams json.RawMessage, signal, request *openrtb2.BidRequest, bidderCodes ...string) (json.RawMessage, models.ResolvedEds, error) {
	resolved := ResolveEds(signal, request)
	if resolved.IsEmpty() || len(bidderCodes) == 0 {
		return bidderParams, resolved, nil
	}

	updated, err := injectIntoBidderParams(bidderParams, resolved, bidderCodes...)
	return updated, resolved, err
}

// remapSignalImpEdsToRequestIDs re-keys resolved.Imp from signal imp IDs to request imp IDs by position.
func remapSignalImpEdsToRequestIDs(resolved models.ResolvedEds, signal, request *openrtb2.BidRequest) models.ResolvedEds {
	if len(resolved.Imp) == 0 || signal == nil || request == nil || len(signal.Imp) == 0 || len(request.Imp) == 0 {
		return resolved
	}

	remapped := make(map[string]json.RawMessage, len(resolved.Imp))
	n := len(signal.Imp)
	if len(request.Imp) < n {
		n = len(request.Imp)
	}
	for i := 0; i < n; i++ {
		edsObj, ok := resolved.Imp[signal.Imp[i].ID]
		if !ok || len(edsObj) == 0 {
			continue
		}
		remapped[request.Imp[i].ID] = edsObj
	}
	resolved.Imp = remapped
	return resolved
}

// StripFromRequest removes *.ext.eds from the shared bid request so other bidders
// do not receive PubMatic-only EDS (device/app/user/imp).
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
	if req.User != nil {
		req.User.Ext = nilIfEmptyExt(jsonparser.Delete(req.User.Ext, "eds"))
	}
	for i := range req.Imp {
		req.Imp[i].Ext = nilIfEmptyExt(jsonparser.Delete(req.Imp[i].Ext, "eds"))
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
	if req == nil {
		return models.ResolvedEds{}
	}

	resolved := models.ResolvedEds{}
	if req.Device != nil {
		resolved.Device = nestedObject(req.Device.Ext, "eds")
	}
	if req.App != nil {
		resolved.App = nestedObject(req.App.Ext, "eds")
	}
	if req.User != nil {
		resolved.User = nestedObject(req.User.Ext, "eds")
	}
	for _, imp := range req.Imp {
		if edsObj := nestedObject(imp.Ext, "eds"); len(edsObj) > 0 {
			if resolved.Imp == nil {
				resolved.Imp = make(map[string]json.RawMessage, len(req.Imp))
			}
			resolved.Imp[imp.ID] = edsObj
		}
	}
	return resolved
}

func injectIntoBidderParams(bidderParams json.RawMessage, resolved models.ResolvedEds, bidderCodes ...string) (json.RawMessage, error) {
	edsPayload, err := buildEdsPayload(resolved)
	if err != nil {
		return bidderParams, err
	}

	result := cloneJSONBytes(bidderParams)
	if len(result) == 0 {
		result = []byte(`{}`)
	} else if _, _, _, err := jsonparser.Get(result); err != nil {
		return bidderParams, err
	}

	edsPayload = cloneJSONBytes(edsPayload)
	for _, code := range bidderCodes {
		result, err = jsonparser.Set(result, edsPayload, code, bidderParamsEdsKey)
		if err != nil {
			return bidderParams, err
		}
	}

	return result, nil
}

func buildEdsPayload(resolved models.ResolvedEds) (json.RawMessage, error) {
	payload := []byte(`{}`)
	var err error

	if len(resolved.Device) > 0 {
		payload, err = jsonparser.Set(payload, cloneJSONBytes(resolved.Device), models.EDSDeviceKey)
		if err != nil {
			return nil, err
		}
	}
	if len(resolved.App) > 0 {
		payload, err = jsonparser.Set(payload, cloneJSONBytes(resolved.App), models.EDSAppKey)
		if err != nil {
			return nil, err
		}
	}
	if len(resolved.User) > 0 {
		payload, err = jsonparser.Set(payload, cloneJSONBytes(resolved.User), models.EDSUserKey)
		if err != nil {
			return nil, err
		}
	}
	if len(resolved.Imp) > 0 {
		impPayload := []byte(`{}`)
		for impID, edsObj := range resolved.Imp {
			impPayload, err = jsonparser.Set(impPayload, cloneJSONBytes(edsObj), impID)
			if err != nil {
				return nil, err
			}
		}
		payload, err = jsonparser.Set(payload, impPayload, models.EDSImpKey)
		if err != nil {
			return nil, err
		}
	}

	return payload, nil
}

func cloneJSONBytes(b json.RawMessage) json.RawMessage {
	if len(b) == 0 {
		return b
	}
	copied := make([]byte, len(b))
	copy(copied, b)
	return copied
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

	return cloneJSONBytes(value)
}

func nilIfEmptyExt(ext []byte) []byte {
	if len(ext) == 0 {
		return nil
	}

	_, dataType, _, err := jsonparser.Get(ext)
	if err != nil || dataType != jsonparser.Object {
		return nil
	}

	hasKeys := false
	if err := jsonparser.ObjectEach(ext, func(_ []byte, _ []byte, _ jsonparser.ValueType, _ int) error {
		hasKeys = true
		return nil
	}); err != nil || !hasKeys {
		return nil
	}

	return ext
}
