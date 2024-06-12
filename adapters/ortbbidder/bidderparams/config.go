package bidderparams

// BidderParamMapper contains property details like location
type BidderParamMapper struct {
	Location string // do not update this parameter for each request, its being shared across all requests
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

// NewBidderConfig initializes and returns the object of BidderConfig
func NewBidderConfig() *BidderConfig {
	return &BidderConfig{
		bidderConfigMap: make(map[string]*config),
	}
}

// SetRequestParams sets the bidder specific requestParams
func (bcfg *BidderConfig) SetRequestParams(bidderName string, requestParams map[string]BidderParamMapper) {
	if _, found := bcfg.bidderConfigMap[bidderName]; !found {
		bcfg.bidderConfigMap[bidderName] = &config{}
	}
	bcfg.bidderConfigMap[bidderName].requestParams = requestParams
}

// SetResponseParams sets the bidder specific responseParams
func (bcfg *BidderConfig) SetResponseParams(bidderName string, responseParams map[string]BidderParamMapper) {
	if _, found := bcfg.bidderConfigMap[bidderName]; !found {
		bcfg.bidderConfigMap[bidderName] = &config{}
	}
	bcfg.bidderConfigMap[bidderName].responseParams = responseParams
}

// GetRequestParams returns bidder specific responseParams
func (bcfg *BidderConfig) GetRequestParams(bidderName string) map[string]BidderParamMapper {
	if len(bcfg.bidderConfigMap) == 0 {
		return nil
	}
	bidderConfig := bcfg.bidderConfigMap[bidderName]
	if bidderConfig == nil {
		return nil
	}
	return bidderConfig.requestParams
}

// GetResponseParams returns bidder specific responseParams
func (bcfg *BidderConfig) GetResponseParams(bidderName string) map[string]BidderParamMapper {
	if len(bcfg.bidderConfigMap) == 0 {
		return nil
	}
	bidderConfig := bcfg.bidderConfigMap[bidderName]
	if bidderConfig == nil {
		return nil
	}
	return bidderConfig.responseParams
}
