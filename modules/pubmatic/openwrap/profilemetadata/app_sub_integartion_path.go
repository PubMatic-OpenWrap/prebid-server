package profilemetadata

import "github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"

func (pmd *profileMetaData) GetAppSubIntegrationPath(appSubIntegrationPath string) (models.AppSubIntegrationPath, bool) {
	pmd.RLock()
	val, ok := pmd.appSubIntegrationPath[appSubIntegrationPath]
	pmd.RUnlock()
	return val, ok
}
