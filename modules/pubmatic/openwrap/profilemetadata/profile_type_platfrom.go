package profilemetadata

func (pmd *profileMetaData) GetProfileTypePlatform(profileTypePlatform string) (int, bool) {
	pmd.RLock()
	val, ok := pmd.profileTypePlatform[profileTypePlatform]
	pmd.RUnlock()
	return val, ok
}
