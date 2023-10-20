package openwrap

import (
	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func getSChainObj(partnerConfigMap map[int]map[string]string) []byte {
	if partnerConfigMap != nil && partnerConfigMap[models.VersionLevelConfigID] != nil {
		if partnerConfigMap[models.VersionLevelConfigID][models.SChainDBKey] == "1" {
			sChainObjJSON := partnerConfigMap[models.VersionLevelConfigID][models.SChainObjectDBKey]
			v, _, _, _ := jsonparser.Get([]byte(sChainObjJSON), "config")
			return v
		}
	}
	return nil
}

// setSchainInSourceObject sets schain object in source.ext.schain
func setSchainInSourceObject(source *openrtb2.Source, schain []byte) {
	if source.Ext == nil {
		source.Ext = []byte("{}")
	}

	sourceExt, err := jsonparser.Set(source.Ext, schain, models.SChainKey)
	if err == nil {
		source.Ext = sourceExt
	}
}
