package profilemetadata

type ProfileMetaData interface {
	GetProfileTypePlatform(profileTypePlatform string) (int, bool)
	GetAppIntegrationPath(appIntegrationPath string) (int, bool)
	GetAppSubIntegrationPath(appSubIntegrationPath string) (int, bool)
}
