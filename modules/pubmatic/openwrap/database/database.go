package database

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

type Database interface {
	GetAdunitConfig(profileID, displayVersion int) (*adunitconfig.AdUnitConfig, error)
	GetActivePartnerConfigurations(pubID, profileID, displayVersion int) (map[int]map[string]string, error)
	GetPublisherSlotNameHash(pubID int) (map[string]string, error)
	GetWrapperSlotMappings(partnerConfigMap map[int]map[string]string, profileID, displayVersion int) (map[int][]models.SlotMapping, error)
	GetPublisherVASTTags(pubID int) (models.PublisherVASTTags, error)
	GetMappings(slotKey string, slotMap map[string]models.SlotMapping) (map[string]interface{}, error)
	GetFSCThresholdPerDSP() (map[int]int, error)
	GetTBFTrafficForPublishers() (map[int]map[int]int, error)
	GetPublisherFeatureMap() (map[int]int, error)
}
