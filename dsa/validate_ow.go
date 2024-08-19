package dsa

import "github.com/prebid/prebid-server/v2/openrtb_ext"

// dropDSA determines whether to drop the DSA (Digital Services Act) flag from the bid response.
// It returns false if the 'Required' field has a value of Supported, Required, or RequiredOnlinePlatform otherwise returns true
func dropDSA(reqDSA *openrtb_ext.ExtRegsDSA, bidDSA *openrtb_ext.ExtBidDSA) bool {
	if bidDSA == nil {
		return false
	}
	if reqDSA == nil || reqDSA.Required == nil {
		return true
	}
	switch *reqDSA.Required {
	case Supported, Required, RequiredOnlinePlatform:
		return false
	}
	return true
}
