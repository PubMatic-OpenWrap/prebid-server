package bidderparams

// BidderParamMapper contains property details like location
type BidderParamMapper struct {
	location []string
}

// GetLocation returns the location of bidderParam
func (bpm *BidderParamMapper) GetLocation() []string {
	return bpm.location
}

// SetLocation sets the location in BidderParamMapper
// Do not modify the location of bidderParam unless you are writing unit test case
func (bpm *BidderParamMapper) SetLocation(location []string) {
	bpm.location = location
}

// config contains mappings requestParams and responseParams
type config struct {
	requestParams  map[string]BidderParamMapper
	responseParams map[string]BidderParamMapper
}

// BidderConfig contains map of bidderName to its requestParams and responseParams
type BidderConfig struct {
	bidderConfigMap map[string]*config
}

// setRequestParams sets the bidder specific requestParams
func (bcfg *BidderConfig) setRequestParams(bidderName string, requestParams map[string]BidderParamMapper) {
	if bcfg == nil {
		return
	}
	if bcfg.bidderConfigMap == nil {
		bcfg.bidderConfigMap = make(map[string]*config)
	}
	if _, found := bcfg.bidderConfigMap[bidderName]; !found {
		bcfg.bidderConfigMap[bidderName] = &config{}
	}
	bcfg.bidderConfigMap[bidderName].requestParams = requestParams
}

// GetRequestParams returns bidder specific requestParams
func (bcfg *BidderConfig) GetRequestParams(bidderName string) (map[string]BidderParamMapper, bool) {
	if bcfg == nil || len(bcfg.bidderConfigMap) == 0 {
		return nil, false
	}
	bidderConfig, _ := bcfg.bidderConfigMap[bidderName]
	if bidderConfig == nil {
		return nil, false
	}
	return bidderConfig.requestParams, true
}