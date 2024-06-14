package profilemetadata

func (pmd *profileMetaData) GetAppIntegrationPath(appIntegrationPathStr string) (int, bool) {
	pmd.RLock()
	defer pmd.RUnlock()
	val, ok := pmd.appIntegrationPath[appIntegrationPathStr]
	return val, ok
}
