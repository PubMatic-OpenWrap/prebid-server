package adpod

import (
	"errors"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func AddTargetingKey(bid *openrtb2.Bid, key openrtb_ext.TargetingKey, value string) error {
	if nil == bid {
		return errors.New("Invalid bid")
	}

	raw, err := jsonparser.Set(bid.Ext, []byte(strconv.Quote(value)), "prebid", "targeting", string(key))
	if nil == err {
		bid.Ext = raw
	}
	return err
}

// ConvertAPRCToNBRC converts the aprc to NonBidStatusCode
func ConvertAPRCToNBRC(bidStatus int64) *openrtb3.NoBidReason {
	var nbrCode openrtb3.NoBidReason

	switch bidStatus {
	case constant.StatusOK:
		nbrCode = nbr.LossBidLostToHigherBid
	case constant.StatusCategoryExclusion:
		nbrCode = exchange.ResponseRejectedCreativeCategoryExclusions
	case constant.StatusDomainExclusion:
		nbrCode = exchange.ResponseRejectedCreativeAdvertiserExclusions
	case constant.StatusDurationMismatch:
		nbrCode = exchange.ResponseRejectedInvalidCreative

	default:
		return nil
	}
	return &nbrCode
}

// ConvertNBRCTOAPRC converts NonBidStatusCode to aprc
func ConvertNBRCTOAPRC(noBidReason *openrtb3.NoBidReason) *int64 {
	var aprc int64

	switch *noBidReason {
	case nbr.LossBidLostToHigherBid:
		aprc = constant.StatusOK
	case exchange.ResponseRejectedCreativeCategoryExclusions:
		aprc = constant.StatusCategoryExclusion
	case exchange.ResponseRejectedCreativeAdvertiserExclusions:
		aprc = constant.StatusDomainExclusion
	case exchange.ResponseRejectedInvalidCreative:
		aprc = constant.StatusDurationMismatch
	default:
		return nil
	}
	return &aprc
}

func GetPodType(imp openrtb2.Imp, extAdpod openrtb_ext.ExtVideoAdPod) PodType {
	if extAdpod.AdPod != nil {
		return Dynamic
	}

	if len(imp.Video.PodID) > 0 && imp.Video.PodDur > 0 {
		return Dynamic
	}

	if len(imp.Video.PodID) > 0 {
		return Structured
	}

	return NotAdpod
}

func ConvertToV25VideoRequest(request *openrtb2.BidRequest) {
	for i := range request.Imp {
		imp := request.Imp[i]

		if imp.Video == nil {
			continue
		}

		// Remove 2.6 Adpod parameters
		imp.Video.PodID = ""
		imp.Video.PodDur = 0
		imp.Video.MaxSeq = 0
	}
}

func createMapFromSlice(slice []string) map[string]bool {
	resultMap := make(map[string]bool)
	for _, item := range slice {
		resultMap[item] = true
	}
	return resultMap
}

func getExclusionConfigs(podId string, adpodExt *openrtb_ext.ExtRequestAdPod) Exclusion {
	var exclusion Exclusion

	if adpodExt != nil && adpodExt.Exclusion != nil {
		var iabCategory, advertiserDomain bool
		for i := range adpodExt.Exclusion.IABCategory {
			if adpodExt.Exclusion.IABCategory[i] == podId {
				iabCategory = true
				break
			}
		}

		for i := range adpodExt.Exclusion.AdvertiserDomain {
			if adpodExt.Exclusion.AdvertiserDomain[i] == podId {
				advertiserDomain = true
				break
			}
		}

		exclusion.IABCategoryExclusion = iabCategory
		exclusion.AdvertiserDomainExclusion = advertiserDomain
	}

	return exclusion
}
