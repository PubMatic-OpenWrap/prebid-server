package resolver

import (
	"regexp"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
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
	paramResolver
}

func (r *bidTypeResolver) getFromORTBObject(bid map[string]any) (any, error) {
	value, ok := bid[ortbFieldMtype]
	if !ok {
		return nil, NewDefaultValueError("no value sent by bidder at [bid.mtype] for [bid.ext.prebid.type]")
	}

	mtype, ok := validateNumber[float64](value)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [bid.mtype] for [bid.ext.prebid.type]")
	}

	if mtype == 0 {
		return nil, NewDefaultValueError("default value sent by bidder at [bid.mtype] for [bid.ext.prebid.type]")
	}

	if bidType := convertToBidType(openrtb2.MarkupType(mtype)); bidType != openrtb_ext.BidType("") {
		return bidType, nil
	}
	return nil, NewValidationFailedError("invalid value sent by bidder at [bid.mtype] for [bid.ext.prebid.type]")
}

func (r *bidTypeResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [bid.ext.prebid.type]", path)
	}
	bidType, ok := value.(string)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [bid.ext.prebid.type]", path)
	}
	return openrtb_ext.BidType(bidType), nil
}

func (r *bidTypeResolver) autoDetect(request *openrtb2.BidRequest, bid map[string]any) (any, error) {
	adm, ok := bid[ortbFieldAdM].(string)
	if ok && adm != "" {
		return getMediaTypeFromAdm(adm), nil // Adm is present, get media type from adm
	}
	impId, ok := bid[ortbFieldImpId].(string)
	if !ok {
		return nil, NewValidationFailedError("invalid value sent by bidder at [bid.impid] for [bid.ext.prebid.type]")
	}
	// Adm is not present, get media type from imp
	return getMediaTypeFromImp(request.Imp, impId)
}

func (r *bidTypeResolver) setValue(adapterBid map[string]any, value any) error {
	adapterBid[bidTypeKey] = value
	return nil
}

func getMediaTypeFromAdm(adm string) openrtb_ext.BidType {
	if videoRegex.MatchString(adm) {
		return openrtb_ext.BidTypeVideo
	}

	if openrtb_ext.IsNative(adm) {
		return openrtb_ext.BidTypeNative
	}

	return openrtb_ext.BidTypeBanner
}

func getMediaTypeFromImp(imps []openrtb2.Imp, impID string) (openrtb_ext.BidType, error) {
	for _, imp := range imps {
		if imp.ID != impID {
			continue
		}
		return getMediaTypes(imp)
	}

	return openrtb_ext.BidType(""), NewValidationFailedError("invalid value sent by bidder at [bid.impid], unable to auto detect value for [bid.ext.prebid.type]")
}

func getMediaTypes(imp openrtb2.Imp) (openrtb_ext.BidType, error) {
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
		return openrtb_ext.BidType(""), NewValidationFailedError("found multi-format request, unable to auto detect value for [bid.ext.prebid.type]")
	}

	return mediaType, nil
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
