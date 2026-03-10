package openwrap

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func vastVersionSupportsCreativeExtensions(version string) bool {
	parts := strings.SplitN(strings.TrimSpace(version), ".", 2)
	if len(parts) == 0 {
		return false
	}
	major, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	return err == nil && major >= 3
}

func trimCDATA(b []byte) []byte {
	s := bytes.TrimSpace(b)
	const cdataStart = "<![CDATA["
	const cdataEnd = "]]>"
	if len(s) >= len(cdataStart)+len(cdataEnd) &&
		bytes.EqualFold(s[:len(cdataStart)], []byte(cdataStart)) &&
		bytes.HasSuffix(s, []byte(cdataEnd)) {
		return s[len(cdataStart) : len(s)-len(cdataEnd)]
	}
	return b
}

type creativeExtensionWithID struct {
	Id   string `xml:"id,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
	Data []byte `xml:",innerxml"`
}

type vastInLineOnly struct {
	XMLName xml.Name `xml:"VAST"`
	Version string   `xml:"version,attr,omitempty"`
	Ads     []struct {
		InLine *struct {
			Creatives []struct {
				CreativeExtensions []creativeExtensionWithID `xml:"CreativeExtensions>CreativeExtension"`
			} `xml:"Creatives>Creative"`
		} `xml:"InLine"`
	} `xml:"Ad"`
}

// ExtractCTAOverlayFromVAST parses adm as VAST and returns ctaoverlay for bid.ext.owsdk.ctaoverlay.
// Only InLine; first CreativeExtension id="PubMatic" wins.
func ExtractCTAOverlayFromVAST(adm string) (interface{}, bool) {
	if adm == "" {
		return nil, false
	}

	var doc vastInLineOnly
	if err := xml.Unmarshal([]byte(adm), &doc); err != nil {
		return nil, false
	}
	if !vastVersionSupportsCreativeExtensions(doc.Version) {
		return nil, false
	}
	for _, ad := range doc.Ads {
		if ad.InLine == nil {
			continue
		}
		for _, cr := range ad.InLine.Creatives {
			for _, ext := range cr.CreativeExtensions {
				if ext.Id != "PubMatic" {
					continue
				}
				raw := trimCDATA(ext.Data)
				var payload struct {
					Ctaoverlay json.RawMessage `json:"ctaoverlay"`
				}
				if err := json.Unmarshal(raw, &payload); err != nil {
					glog.Warningf("ctaoverlay: invalid JSON in CreativeExtension id=PubMatic: %v", err)
					continue
				}
				if len(payload.Ctaoverlay) == 0 {
					continue
				}
				var out interface{}
				if err := json.Unmarshal(payload.Ctaoverlay, &out); err != nil {
					glog.Warningf("ctaoverlay: invalid JSON in CreativeExtension id=PubMatic: %v", err)
					continue
				}
				return out, true
			}
		}
	}
	return nil, false
}

var ctaOverlayAllowedSDKVersions = map[string]struct{}{
	"4.9.0":  {},
	"4.9.1":  {},
	"4.10.0": {},
	"4.11.0": {},
}

// IsVideoBidEligibleForCTAOverlay returns true when we should parse adm and inject bid.ext.owsdk.ctaoverlay.
func IsVideoBidEligibleForCTAOverlay(bidExt *models.BidExt, ctaOverlayRequested bool, displayManagerVer string) bool {
	if bidExt == nil || bidExt.CreativeType != models.MediaTypeVideo {
		return false
	}
	if _, ok := ctaOverlayAllowedSDKVersions[strings.TrimSpace(displayManagerVer)]; !ok {
		return false
	}
	if bidExt.OWSDK != nil && bidExt.OWSDK[models.CTAOVERLAY] != nil {
		return false
	}
	return ctaOverlayRequested
}
