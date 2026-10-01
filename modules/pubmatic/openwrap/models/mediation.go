package models

// AppSubIntegrationPath is one row from app_sub_integration_path.
// MediationName and MediationType come from the mediation name and type columns.
type AppSubIntegrationPath struct {
	ID            int
	MediationName string
	MediationType string
}

// ImpMediation is imp.ext.prebid.bidder.pubmatic.mediation.
// MediationType is written as waterfall_or_bidding.
type ImpMediation struct {
	Name               string `json:"name,omitempty"`
	WaterfallOrBidding string `json:"waterfall_or_bidding,omitempty"`
}

// ImpMediation builds imp.ext.mediation from the sub-integration path row.
// Name comes from MediationName and waterfall_or_bidding comes from MediationType.
func (p AppSubIntegrationPath) ImpMediation() *ImpMediation {
	if p.MediationName == "" || p.MediationType == "" {
		return nil
	}
	return &ImpMediation{
		Name:               p.MediationName,
		WaterfallOrBidding: p.MediationType,
	}
}
