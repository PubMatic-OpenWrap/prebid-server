package adpod

import (
	"errors"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/exchange"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func addTargetingKey(bid *openrtb2.Bid, key openrtb_ext.TargetingKey, value string) error {
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
