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
	if resolved := resolveEdsFromRequest(signal); !resolved.IsEmpty() {
		return resolved
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

// StripFromRequest removes ext.eds wrappers and resolved flat ext keys
// from the shared bid request so other bidders do not receive them.
func StripFromRequest(req *openrtb2.BidRequest, resolved models.ResolvedEds) {
	if req == nil {
		return
	}

	if req.Device != nil {
		req.Device.Ext = stripObjectExt(req.Device.Ext, stripKeysForObject(req.Device.Ext, resolved.Device))
	}
	if req.App != nil {
		req.App.Ext = stripObjectExt(req.App.Ext, stripKeysForObject(req.App.Ext, resolved.App))
	}
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
	return resolved
}

func injectIntoBidderParams(bidderParams json.RawMessage, resolved models.ResolvedEds, bidderCodes ...string) (json.RawMessage, error) {
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

func stripKeysForObject(ext []byte, resolvedExt json.RawMessage) json.RawMessage {
	if fromEds := nestedObject(ext, "eds"); len(fromEds) > 0 {
		return mergeExtObjects(resolvedExt, fromEds, true)
	}
	return resolvedExt
}

func stripObjectExt(ext []byte, resolvedExt json.RawMessage) []byte {
	if len(ext) == 0 {
		return ext
	}

	ext = jsonparser.Delete(ext, "eds")
	return deleteExtKeysFromObject(ext, resolvedExt)
}
