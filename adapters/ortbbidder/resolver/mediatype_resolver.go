package resolver

import (
	"regexp"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

var (
	videoRegex *regexp.Regexp
)

func init() {
	videoRegex, _ = regexp.Compile("<VAST\\s+")
}

// mtypeResolver resolves the media type of the type bid
type mtypeResolver struct {
	valueResolver
}

func (r *mtypeResolver) getFromORTBObject(bid map[string]any) (any, bool) {
	mtype, ok := bid[mtypeKey].(float64)
	if !ok || mtype == 0 {
		return nil, false
	}
	return util.GetMediaType(openrtb2.MarkupType(mtype)), true
}

func (r *mtypeResolver) autoDetect(bid map[string]any) (any, bool) {
	adm, ok := bid[admKey].(string)
	if !ok || adm == "" {
		return nil, false
	}
	return getMediaTypeFromAdm(adm), true
}

func (r *mtypeResolver) setValue(adapterBid map[string]any, value any) {
	adapterBid[bidTypeKey] = value
}

// TODO: check impression.Banner/Native/Video for detecting mtype ?
func getMediaTypeFromAdm(adm string) openrtb_ext.BidType {
	if videoRegex.MatchString(adm) {
		return openrtb_ext.BidTypeVideo
	}

	for _, tag := range []string{"native", "link", "assets"} {
		if _, _, _, err := jsonparser.Get([]byte(adm), tag); err == nil {
			return openrtb_ext.BidTypeNative
		}
	}

	return openrtb_ext.BidTypeBanner
}
