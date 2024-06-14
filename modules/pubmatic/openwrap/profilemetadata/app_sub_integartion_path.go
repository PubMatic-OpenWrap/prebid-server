package profilemetadata

func (pmd *profileMetaData) GetAppSubIntegrationPath(appSubIntegrationPathStr string) (int, bool) {
	pmd.RLock()
	defer pmd.RUnlock()
	val, ok := pmd.appSubIntegrationPath[appSubIntegrationPathStr]
	return val, ok
}
