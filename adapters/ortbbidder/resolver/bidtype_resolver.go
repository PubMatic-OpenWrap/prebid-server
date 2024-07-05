package resolver

import (
	"regexp"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

var (
	videoRegex = regexp.MustCompile(`<VAST\s+`)
)

// bidTypeResolver determines the bid type based on the following hierarchy:
// 1. It first attempts to retrieve the bid type from the response.seat.bid.mtype location.
// 2. If not found, it then tries to retrieve the bid type using the bidder param location.
// 3. If still not found, it automatically detects the bid type using either the adm or impression.
// The determined bid type is subsequently assigned to adapterresponse.typedbid.bidtype
type bidTypeResolver struct {
	defaultValueResolver
}

func (r *bidTypeResolver) getFromORTBObject(bid map[string]any) (any, bool) {
	mtype, ok := bid[ortbFieldMtype].(float64)
	if !ok || mtype == 0 {
		return nil, false
	}

	if bidType := convertToBidType(openrtb2.MarkupType(mtype)); bidType != openrtb_ext.BidType("") {
		return bidType, true
	}
	return nil, false
}

func (r *bidTypeResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	bidType, ok := value.(string)
	return openrtb_ext.BidType(bidType), ok
}

func (r *bidTypeResolver) autoDetect(request *openrtb2.BidRequest, bid map[string]any) (any, bool) {
	adm, ok := bid[ortbFieldAdM].(string)
	if ok && adm != "" {
		return getMediaTypeFromAdm(adm), true // Adm is present, get media type from adm
	}
	impId, ok := bid[ortbFieldImpId].(string)
	if !ok {
		return nil, false
	}
	// Adm is not present, get media type from imp
	return getMediaTypeFromImp(request.Imp, impId), true
}

func (r *bidTypeResolver) setValue(adapterBid map[string]any, value any) bool {
	adapterBid[bidTypeKey] = value
	return true
}

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

func convertToBidType(mtype openrtb2.MarkupType) openrtb_ext.BidType { // change name
	var bidType openrtb_ext.BidType
	switch mtype {
	case openrtb2.MarkupBanner:
		bidType = openrtb_ext.BidTypeBanner
	case openrtb2.MarkupVideo:
		bidType = openrtb_ext.BidTypeVideo
	case openrtb2.MarkupAudio:
		bidType = openrtb_ext.BidTypeAudio
	case openrtb2.MarkupNative:
		bidType = openrtb_ext.BidTypeNative
	}
	return bidType
}
