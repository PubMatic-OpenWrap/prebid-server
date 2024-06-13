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
	videoRegex, _ = regexp.Compile(`<VAST\s+`)
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
func (r *mtypeResolver) autoDetect(request *openrtb2.BidRequest, bid map[string]any) (any, bool) {
	adm, ok := bid[admKey].(string)
	if !ok || adm == "" {
		impId, ok := bid[impIdKey].(string)
		if !ok {
			return nil, false
		}
		// Adm is not present, get media type from imp
		return getMediaTypeFromImp(request.Imp, impId), true
	}
	// Adm is present, get media type from adm
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

func getMediaTypeFromImp(imps []openrtb2.Imp, impID string) openrtb_ext.BidType {
	for _, imp := range imps {
		if imp.ID != impID {
			continue
		}
		return getMediaTypes(imp)
	}

	return openrtb_ext.BidType("")
}

func getMediaTypes(imp openrtb2.Imp) openrtb_ext.BidType {
	var (
		multiFormatCount int
		mediaType        openrtb_ext.BidType
	)

	if imp.Banner != nil {
		multiFormatCount++
		mediaType = openrtb_ext.BidTypeBanner
	}
	if imp.Video != nil {
		multiFormatCount++
		mediaType = openrtb_ext.BidTypeVideo
	}
	if imp.Native != nil {
		multiFormatCount++
		mediaType = openrtb_ext.BidTypeNative
	}
	// imp has multiple format, set mediaType to empty
	if multiFormatCount > 1 {
		return openrtb_ext.BidType("")
	}

	return mediaType
}
