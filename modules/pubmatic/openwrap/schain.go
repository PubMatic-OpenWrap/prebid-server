package openwrap

import (
	"encoding/json"

	"github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/constant"
	"github.com/prebid/openrtb/v19/openrtb2"
)

// SupplyChainConfig reads profile level supply chain object from database
type SupplyChainConfig struct {
	Validation  string                `json:"validation"`
	SupplyChain *openrtb2.SupplyChain `json:"config"`
}

// func getSChainObj(partnerConfigMap map[int]map[string]string) []byte {
// 	if partnerConfigMap != nil && partnerConfigMap[models.VersionLevelConfigID] != nil {
// 		if partnerConfigMap[models.VersionLevelConfigID][models.SChainDBKey] == "1" {
// 			sChainObjJSON := partnerConfigMap[models.VersionLevelConfigID][models.SChainObjectDBKey]
// 			v, _, _, _ := jsonparser.Get([]byte(sChainObjJSON), "config")
// 			return v
// 		}
// 	}
// 	return nil
// }

func getSChainObj(partnerConfigMap map[int]map[string]string) *openrtb2.SupplyChain {
	if partnerConfigMap != nil && partnerConfigMap[constant.VersionLevelConfigID] != nil {
		if partnerConfigMap[constant.VersionLevelConfigID][constant.SChainDBKey] == "1" {
			sChainObjJSON := partnerConfigMap[constant.VersionLevelConfigID][constant.SChainObjectDBKey]
			sChainConfig := &SupplyChainConfig{}
			if err := json.Unmarshal([]byte(sChainObjJSON), sChainConfig); err != nil {
				//logger.Error(errorcodes.ErrJSONUnmarshalFailed, constant.SChainKey, err.Error(), sChainObjJSON)
				return nil
			}
			if sChainConfig != nil && sChainConfig.SupplyChain != nil {
				return sChainConfig.SupplyChain
			}
		}
	}
	return nil
}

// setSChainInSourceObject sets schain object in source.ext.schain
func setSChainInSourceObject(source *openrtb2.Source, partnerConfigMap map[int]map[string]string) {

	var sChainObj *openrtb2.SupplyChain
	if source.SChain == nil {
		sChainObj = getSChainObj(partnerConfigMap)
	} else {
		sChainObj = source.SChain
		source.SChain = nil
	}

	if sChainObj != nil {
		//Temporary change till all bidder start using openrtb 2.6 source.schain

		sourceExt, err := json.Marshal(sChainObj)
		if err == nil {
			source.Ext = sourceExt
		}

	}
}
