package resolver

import "github.com/prebid/prebid-server/v2/adapters/ortbbidder/util"

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
	return validateFledgeConfigs(value)
}

func validateFledgeConfigs(value any) ([]map[string]any, bool) {
	inputFledgeCfgs, ok := value.([]any)
	if !ok {
		return nil, false
	}

	outputFledgeCfgs := make([]map[string]any, 0, len(inputFledgeCfgs))
	for _, fledgeCfg := range inputFledgeCfgs {
		validFledgeCfg, ok := validateFledgeConfig(fledgeCfg)
		if ok {
			outputFledgeCfgs = append(outputFledgeCfgs, validFledgeCfg)
		}
	}
	return outputFledgeCfgs, len(outputFledgeCfgs) != 0
}

func validateFledgeConfig(fledgeCfg any) (map[string]any, bool) {
	inputFledgeCfg, ok := fledgeCfg.(map[string]any)
	if !ok {
		return nil, false
	}

	outputFledgeCfg := make(map[string]any, len(inputFledgeCfg))
	for key, value := range inputFledgeCfg {
		ok = true
		switch key {
		case ortbFieldImpId, ortbFieldBidder, ortbFieldAdapter:
			value, ok = value.(string)
		case ortbFieldConfig:
			value, ok = value.(map[string]any)
		}
		if ok {
			outputFledgeCfg[key] = value
		}
	}
	return outputFledgeCfg, len(outputFledgeCfg) != 0
}

func (f *fledgeResolver) setValue(adapterBid map[string]any, value any) bool {
	adapterBid[fledgeAuctionConfigKey] = value
	return true
}
