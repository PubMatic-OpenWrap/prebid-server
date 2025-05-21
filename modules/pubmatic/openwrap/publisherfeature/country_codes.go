package publisherfeature

// updateCountryCodesMapping returns country codes mapping
func (fe *feature) updateCountryCodesMapping() {
	// update country codes mapping
	countryCodesMapping, err := fe.cache.GetCountryCodesMapping()
	if err != nil || countryCodesMapping == nil {
		return
	}
	fe.countryCodesMapping = countryCodesMapping
}

// GetAlpha2Code returns alpha2 code for country
func (fe *feature) GetAlpha2Code(countryCode string) string {
	return fe.countryCodesMapping[countryCode].Alpha2Code
}
