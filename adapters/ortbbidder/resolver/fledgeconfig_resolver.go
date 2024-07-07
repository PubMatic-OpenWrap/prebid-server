package resolver

import (
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// fledgeResolver retrieves the fledge auction config of the bidresponse using the bidder param location.
// The determined fledge config is subsequently assigned to adapterresponse.FledgeAuctionConfigs
type fledgeResolver struct {
	defaultValueResolver
}

func (f *fledgeResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, bool) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, false
	}
	return validateFledgeConfig(value)
}

func validateFledgeConfig(value any) (any, bool) {
	fledgeCfgBytes, err := jsonutil.Marshal(value)
	if err != nil {
		return nil, false
	}

	var fledgeCfg []*openrtb_ext.FledgeAuctionConfig
	err = jsonutil.UnmarshalValid(fledgeCfgBytes, &fledgeCfg)
	if err != nil {
		return nil, false
	}
	return value, true
}

func (f *fledgeResolver) setValue(adapterBid map[string]any, value any) bool {
	adapterBid[fledgeAuctionConfigKey] = value
	return true
}
