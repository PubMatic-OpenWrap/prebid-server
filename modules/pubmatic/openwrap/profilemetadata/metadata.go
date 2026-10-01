package profilemetadata

import "github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"

type ProfileMetaData interface {
	GetProfileTypePlatform(profileTypePlatform string) (int, bool)
	GetAppIntegrationPath(appIntegrationPath string) (int, bool)
	GetAppSubIntegrationPath(appSubIntegrationPath string) (models.AppSubIntegrationPath, bool)
}
