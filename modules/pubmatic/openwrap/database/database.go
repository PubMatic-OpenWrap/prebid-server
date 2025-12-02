package database

import (
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adpodconfig"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
)

type Database interface {
	GetAdunitConfig(profileID, displayVersion int) (*adunitconfig.AdUnitConfig, error)
	GetActivePartnerConfigurations(pubID, profileID, displayVersion int) (map[int]map[string]string, error)
	GetPublisherSlotNameHash(pubID int) (map[string]string, error)
	GetWrapperSlotMappings(partnerConfigMap map[int]map[string]string, profileID, displayVersion int) (map[int][]models.SlotMapping, error)
	GetPublisherVASTTags(pubID int) (models.PublisherVASTTags, error)
	GetMappings(slotKey string, slotMap map[string]models.SlotMapping) (map[string]interface{}, error)
	GetFSCThresholdPerDSP() (map[int]int, error)
	GetPublisherFeatureMap() (map[int]map[int]models.FeatureData, error)
	GetAdpodConfig(pubID, profileID, displayVersion int) (*adpodconfig.AdpodConfig, error)
	GetProfileTypePlatforms() (map[string]int, error)
	GetAppIntegrationPaths() (map[string]int, error)
	GetAppSubIntegrationPaths() (map[string]int, error)
	GetGDPRCountryCodes() (models.HashSet, error)
	GetProfileAdUnitMultiFloors() (models.ProfileAdUnitMultiFloors, error)
	GetLatestCountryPartnerFilter() map[string]map[string]struct{}
	GetPerformanceDSPs() (map[int]struct{}, error)
	GetInViewEnabledPublishers() (map[int]struct{}, error)
}
