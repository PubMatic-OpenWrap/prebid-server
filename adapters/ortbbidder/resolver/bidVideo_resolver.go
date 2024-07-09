package resolver

import (
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// bidVideoResolver determines the duration of the bid by retrieving the video field using the bidder param location.
// The determined video field is subsequently assigned to adapterresponse.typedbid.bidvideo
type bidVideoResolver struct {
	defaultValueResolver
}

func (b *bidVideoResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateBidVideo(value)
}

func validateBidVideo(value any) (any, bool) {
	bidVideoBytes, err := jsonutil.Marshal(value)
	if err != nil {
		return nil, false
	}

	var bidVideo openrtb_ext.ExtBidPrebidVideo
	err = jsonutil.UnmarshalValid(bidVideoBytes, &bidVideo)
	if err != nil {
		return nil, false
	}

	var bidVideoMap map[string]any
	err = jsonutil.UnmarshalValid(bidVideoBytes, &bidVideoMap)
	if err != nil {
		return nil, false
	}
	return bidVideoMap, true
}

func (b *bidVideoResolver) setValue(adapterBid map[string]any, value any) bool {
	adapterBid[bidVideoKey] = value
	return true
}

// bidVideoDurationResolver determines the duration of the bid based on the following hierarchy:
// 1. It first attempts to retrieve the bid type from the response.seat.bid.dur location.
// 2. If not found, it then tries to retrieve the duration using the bidder param location.
// The determined bid duration is subsequently assigned to adapterresponse.typedbid.bidvideo.dur
type bidVideoDurationResolver struct {
	defaultValueResolver
}

func (b *bidVideoDurationResolver) getFromORTBObject(bid map[string]any) (any, bool) {
	dur, ok := bid[ortbFieldDuration].(float64)
	if !ok || dur == 0 {
		return nil, false
	}
	return int64(dur), true
}

func (b *bidVideoDurationResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	dur, ok := value.(float64)
	return int64(dur), ok
}

func (b *bidVideoDurationResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidVideo(adapterBid, bidVideoDurationKey, value)
}

// bidVideoPrimaryCategoryResolver determines the primary-category of the bid based on the following hierarchy:
// 1. It first attempts to retrieve the bid category from the response.seat.bid.cat[0] location.
// 2. If not found, it then tries to retrieve the duration using the bidder param location.
// The determined category is subsequently assigned to adapterresponse.typedbid.bidvideo.primary_category
type bidVideoPrimaryCategoryResolver struct {
	defaultValueResolver
}

func (b *bidVideoPrimaryCategoryResolver) getFromORTBObject(bid map[string]any) (any, bool) {
	cat, _ := bid[ortbFieldCategory].([]any)
	if len(cat) == 0 {
		return nil, false
	}
	typedCat, _ := cat[0].(string)
	if len(typedCat) == 0 {
		return nil, false
	}
	return typedCat, true
}

func (b *bidVideoPrimaryCategoryResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	cat, ok := value.(string)
	return cat, ok
}

func (b *bidVideoPrimaryCategoryResolver) setValue(adapterBid map[string]any, value any) bool {
	return setKeyValueInBidVideo(adapterBid, bidVideoPrimaryCategoryKey, value)
}

func setKeyValueInBidVideo(adapterBid map[string]any, key string, value any) bool {
	video, found := adapterBid[bidVideoKey]
	if !found {
		video = map[string]any{}
		adapterBid[bidVideoKey] = video
	}
	videoTyped, ok := video.(map[string]any)
	if !ok || videoTyped == nil {
		return false
	}
	videoTyped[key] = value
	return true
}
