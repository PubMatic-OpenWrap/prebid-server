package models

// PartnerFeatureRecord represents a country-specific partner feature configuration
type PartnerFeatureRecord struct {
	Country           string
	FeatureValue      string
	Criteria          string
	CriteriaThreshold int
}
