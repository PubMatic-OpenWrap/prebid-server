package profilemetadata

func (pmd *profileMetaData) GetAppSubIntegrationPath(appSubIntegrationPath string) (int, bool) {
	pmd.RLock()
	val, ok := pmd.appSubIntegrationPath[appSubIntegrationPath]
	pmd.RUnlock()
	return val, ok
}
