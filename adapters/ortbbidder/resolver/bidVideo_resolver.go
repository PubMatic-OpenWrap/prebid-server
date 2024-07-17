package resolver

import (
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// bidVideoResolver determines the duration of the bid by retrieving the video field using the bidder param location.
// The determined video field is subsequently assigned to adapterresponse.typedbid.bidvideo
type bidVideoResolver struct {
	paramResolver
}

func (b *bidVideoResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, nil
	}
	video, err := validateBidVideo(value)
	if err != nil {
		return nil, util.NewWarning("failed to map response-param:[bidVideo] method:[response_param_location] value:[%v]", value)
	}
	return video, nil
}

func validateBidVideo(value any) (any, error) {
	bidVideoBytes, err := jsonutil.Marshal(value)
	if err != nil {
		return nil, err
	}

	var bidVideo openrtb_ext.ExtBidPrebidVideo
	err = jsonutil.UnmarshalValid(bidVideoBytes, &bidVideo)
	if err != nil {
		return nil, err
	}

	var bidVideoMap map[string]any
	err = jsonutil.UnmarshalValid(bidVideoBytes, &bidVideoMap)
	if err != nil {
		return nil, err
	}
	return bidVideoMap, nil
}

func (b *bidVideoResolver) setValue(adapterBid map[string]any, value any) error {
	adapterBid[bidVideoKey] = value
	return nil
}

// bidVideoDurationResolver determines the duration of the bid based on the following hierarchy:
// 1. It first attempts to retrieve the bid type from the response.seat.bid.dur location.
// 2. If not found, it then tries to retrieve the duration using the bidder param location.
// The determined bid duration is subsequently assigned to adapterresponse.typedbid.bidvideo.dur
type bidVideoDurationResolver struct {
	paramResolver
}

func (b *bidVideoDurationResolver) getFromORTBObject(bid map[string]any) (any, error) {
	value, ok := bid[ortbFieldDuration]
	if !ok || value == 0 {
		return nil, nil
	}
	duration, ok := validateNumber[int64](value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidVideoDuration] method:[standard_oRTB_param] value:[%v]", value)
	}
	return duration, nil
}

func (b *bidVideoDurationResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, nil
	}
	duration, ok := validateNumber[int64](value)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidVideoDuration] method:[response_param_location] value:[%v]", value)
	}
	return duration, nil
}

func (b *bidVideoDurationResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidVideo(adapterBid, bidVideoDurationKey, value)
}

// bidVideoPrimaryCategoryResolver determines the primary-category of the bid based on the following hierarchy:
// 1. It first attempts to retrieve the bid category from the response.seat.bid.cat[0] location.
// 2. If not found, it then tries to retrieve the primary category using the bidder param location.
// The determined category is subsequently assigned to adapterresponse.typedbid.bidvideo.primary_category
type bidVideoPrimaryCategoryResolver struct {
	paramResolver
}

func (b *bidVideoPrimaryCategoryResolver) getFromORTBObject(bid map[string]any) (any, error) {
	value, found := bid[ortbFieldCategory]
	if !found {
		return nil, nil
	}

	categories, ok := value.([]any)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidVideoPrimaryCategory] method:[standard_oRTB_param] value:[%v]", value)
	}

	if len(categories) == 0 {
		return nil, nil
	}

	category, _ := categories[0].(string)
	if len(category) == 0 {
		return nil, util.NewWarning("failed to map response-param:[bidVideoPrimaryCategory] method:[standard_oRTB_param] value:[%v]", value)
	}

	return category, nil
}

func (b *bidVideoPrimaryCategoryResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, nil
	}
	category, ok := value.(string)
	if !ok {
		return nil, util.NewWarning("failed to map response-param:[bidVideoPrimaryCategory] method:[response_param_location] value:[%v]", value)
	}
	return category, nil
}

func (b *bidVideoPrimaryCategoryResolver) setValue(adapterBid map[string]any, value any) error {
	return setKeyValueInBidVideo(adapterBid, bidVideoPrimaryCategoryKey, value)
}

func setKeyValueInBidVideo(adapterBid map[string]any, key string, value any) error {
	video, found := adapterBid[bidVideoKey]
	if !found {
		video = map[string]any{}
		adapterBid[bidVideoKey] = video
	}
	videoTyped, ok := video.(map[string]any)
	if !ok || videoTyped == nil {
		return util.NewWarning("failed to set key:[%s] in BidVideo, value:[%v] error:[incorrect data type]", key, value)
	}
	videoTyped[key] = value
	return nil
}
