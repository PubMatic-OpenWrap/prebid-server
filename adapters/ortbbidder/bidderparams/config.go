package bidderparams

// BidderParamMapper contains property details like location
type BidderParamMapper struct {
	Location string // do not update this parameter for each request, its being shared across all requests
}

// Config contains mappings RequestParams and ResponseParams
type Config struct {
	RequestParams  map[string]BidderParamMapper
	ResponseParams map[string]BidderParamMapper
}

// BidderConfig contains map of bidderName to its RequestParams and ResponseParams
type BidderConfig struct {
	BidderConfigMap map[string]*Config
}

// NewBidderConfig initializes and returns the object of BidderConfig
func NewBidderConfig() *BidderConfig {
	return &BidderConfig{
		BidderConfigMap: make(map[string]*Config),
	}
}

// GetRequestParams returns bidder specific ResponseParams
func (bcfg *BidderConfig) GetRequestParams(bidderName string) map[string]BidderParamMapper {
	if len(bcfg.BidderConfigMap) == 0 {
		return nil
	}
	bidderConfig := bcfg.BidderConfigMap[bidderName]
	if bidderConfig == nil {
		return nil
	}
	return bidderConfig.RequestParams
}

// GetResponseParams returns bidder specific ResponseParams
func (bcfg *BidderConfig) GetResponseParams(bidderName string) map[string]BidderParamMapper {
	if len(bcfg.BidderConfigMap) == 0 {
		return nil
	}
	bidderConfig := bcfg.BidderConfigMap[bidderName]
	if bidderConfig == nil {
		return map[string]BidderParamMapper{}
	}
	return bidderConfig.ResponseParams
}
