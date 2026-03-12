package openwrap

import (
	"encoding/json"
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/parser"
)

var ctaOverlayAllowedSDKVersions = map[string]struct{}{
	"4.9.0":  {},
	"4.9.1":  {},
	"4.10.0": {},
	"4.11.0": {},
}

// getCTAOverlayFromFastXMLHandler returns the ctaoverlay JSON from an already-parsed FastXML handler.
// The caller must create the handler and call Parse(adm) before calling this. Returns json.RawMessage
// so the caller can inject it directly without a second unmarshal. Parser returns first CreativeExtension name=PubMatic (case-insensitive; CDATA-trimmed).
func getCTAOverlayFromFastXMLHandler(h *parser.FastXMLHandler) (json.RawMessage, bool) {
	raw := h.ExtractCTAOverlayFromVAST()
	if raw == "" {
		return nil, false
	}
	var payload struct {
		Ctaoverlay json.RawMessage `json:"ctaoverlay"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || len(payload.Ctaoverlay) == 0 {
		glog.Errorf("[getCTAOverlayFromFastXMLHandler] Error unmarshalling ctaoverlay: %s, raw: %s", err, raw)
		return nil, false
	}
	return payload.Ctaoverlay, true
}

// ExtractCTAOverlayFromVASTFastXML parses adm with the FastXML handler and returns ctaoverlay as json.RawMessage
// for direct injection (e.g. into bid.ext.owsdk.ctaoverlay). It creates the handler, calls Parse(adm), then getCTAOverlayFromFastXMLHandler(h).
func ExtractCTAOverlayFromVASTFastXML(adm string) (json.RawMessage, bool) {
	if adm == "" {
		return nil, false
	}
	h := &parser.FastXMLHandler{}
	if err := h.Parse(adm); err != nil {
		return nil, false
	}
	return getCTAOverlayFromFastXMLHandler(h)
}

// IsVideoBidEligibleForCTAOverlay returns true when we should parse adm and inject bid.ext.owsdk.ctaoverlay.
func IsVideoBidEligibleForCTAOverlay(bidExt *models.BidExt, ctaOverlayRequested bool, displayManagerVer string) bool {
	if !ctaOverlayRequested {
		return false
	}
	if bidExt == nil || bidExt.CreativeType != models.MediaTypeVideo {
		return false
	}
	if _, ok := ctaOverlayAllowedSDKVersions[strings.TrimSpace(displayManagerVer)]; !ok {
		return false
	}
	if bidExt.OWSDK != nil && bidExt.OWSDK[models.CTAOVERLAY] != nil {
		return false
	}
	return true
}
