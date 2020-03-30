package privacy

import (
	"github.com/prebid/prebid-server/privacy/ccpa"
	"github.com/prebid/prebid-server/privacy/gdpr"
	"github.com/prebid/prebid-server/privacy/lmt"
)

// Policies represents the privacy regulations for an OpenRTB bid request.
type Policies struct {
	CCPA ccpa.Policy
	GDPR gdpr.Policy
	LMT  lmt.Policy
}

// ReadPoliciesFromConsent inspects the consent string kind and sets the corresponding values in a new Policies object.
func ReadPoliciesFromConsent(consent string) (Policies, bool) {
	if len(consent) == 0 {
		return Policies{}, false
	}

	if err := gdpr.ValidateConsent(consent); err == nil {
		return Policies{
			GDPR: gdpr.Policy{
				Consent: consent,
			},
		}, true
	}

	if err := ccpa.ValidateConsent(consent); err == nil {
		return Policies{
			CCPA: ccpa.Policy{
				Value: consent,
			},
		}, true
	}

	return Policies{}, false
}
