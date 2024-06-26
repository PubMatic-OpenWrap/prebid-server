package profilemetadata

func (pmd *profileMetaData) GetAppIntegrationPath(appIntegrationPath string) (int, bool) {
	pmd.RLock()
	val, ok := pmd.appIntegrationPath[appIntegrationPath]
	pmd.RUnlock()
	return val, ok
}
