package models

// AppSubIntegrationPath is one row from app_sub_integration_path.
// MediationName and MediationType come from the mediation name and type columns.
type AppSubIntegrationPath struct {
	ID            int
	MediationName string
	MediationPath string
}

// ImpMediation is imp.ext.prebid.bidder.pubmatic.mediation.
type ImpMediation struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

// ImpMediation builds imp.ext.mediation from the sub-integration path row.
// Name comes from MediationName and path comes from MediationPath.
func (p AppSubIntegrationPath) ImpMediation() *ImpMediation {
	if p.MediationName == "" || p.MediationPath == "" {
		return nil
	}
	return &ImpMediation{
		Name: p.MediationName,
		Path: p.MediationPath,
	}
}
