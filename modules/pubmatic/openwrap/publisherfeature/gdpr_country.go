package publisherfeature

// IsCountryGDPREnabled returns true if country is gdpr enabled
func (fe *feature) IsCountryGDPREnabled(countryCode string) bool {
	fe.RLock()
	defer fe.RUnlock()
	_, enabled := fe.gdprCountryCodes[countryCode]
	return enabled
}
