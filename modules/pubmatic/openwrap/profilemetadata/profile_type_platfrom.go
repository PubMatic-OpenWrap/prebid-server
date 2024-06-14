package profilemetadata

func (pmd *profileMetaData) GetProfileTypePlatform(profileTypePlatformStr string) (int, bool) {
	pmd.RLock()
	defer pmd.RUnlock()
	val, ok := pmd.profileTypePlatform[profileTypePlatformStr]
	return val, ok
}
