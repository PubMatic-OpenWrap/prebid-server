package resolver

import (
	"github.com/prebid/prebid-server/v3/adapters/ortbbidder/util"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/jsonutil"
)

// fledgeResolver retrieves the fledge auction config of the bidresponse using the bidder param location.
// The determined fledge config is subsequently assigned to adapterresponse.FledgeAuctionConfigs
type fledgeResolver struct {
	paramResolver
}

func (f *fledgeResolver) retrieveFromBidderParamLocation(responseNode map[string]any, path string) (any, error) {
	value, found := util.GetValueFromLocation(responseNode, path)
	if !found {
		return nil, NewDefaultValueError("no value sent by bidder at [%s] for [response.ext.fledgeAuctionConfig]", path)
	}
	fledgeCfg, err := validateFledgeConfig(value)
	if err != nil {
		return nil, NewValidationFailedError("invalid value sent by bidder at [%s] for [response.ext.fledgeAuctionConfig]", path)
	}
	return fledgeCfg, nil
}

func validateFledgeConfig(value any) (any, error) {
	fledgeCfgBytes, err := jsonutil.Marshal(value)
	if err != nil {
		return nil, err
	}

	var fledgeCfg []*openrtb_ext.FledgeAuctionConfig
	err = jsonutil.UnmarshalValid(fledgeCfgBytes, &fledgeCfg)
	if err != nil {
		return nil, err
	}

	return fledgeCfg, nil
}

func (f *fledgeResolver) setValue(adapterBid map[string]any, value any) error {
	adapterBid[fledgeAuctionConfigKey] = value
	return nil
}
