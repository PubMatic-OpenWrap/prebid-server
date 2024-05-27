package openwrap

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/endpoints/legacy/ctv"
)

const (
	VersionLevelConfigID = -1
)

const (
	SChainDBKey       = "sChain"
	SChainObjectDBKey = "sChainObj"
	SChainKey         = "schain"
	//SChainConfigKey   = "config"
)

// SupplyChainConfig reads profile level supply chain object from database
type SupplyChainConfig struct {
	Validation  string                `json:"validation"`
	SupplyChain *openrtb2.SupplyChain `json:"config"`
}

func getSChainObj(partnerConfigMap map[int]map[string]string) *openrtb2.SupplyChain {
	if partnerConfigMap != nil && partnerConfigMap[VersionLevelConfigID] != nil {
		if partnerConfigMap[VersionLevelConfigID][SChainDBKey] == "1" {
			sChainObjJSON := partnerConfigMap[VersionLevelConfigID][SChainObjectDBKey]
			sChainConfig := &SupplyChainConfig{}
			if err := json.Unmarshal([]byte(sChainObjJSON), sChainConfig); err != nil {
				glog.Errorf(ctv.ErrJSONUnmarshalFailed, SChainKey, err.Error(), sChainObjJSON)
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
		var sourceExtMap map[string]any
		if source.Ext == nil {
			source.Ext = []byte(`{}`)
		}
		err := json.Unmarshal(source.Ext, &sourceExtMap)
		if err != nil {
			sourceExtMap = map[string]any{}
		}
		sourceExtMap[SChainKey] = sChainObj
		sourceExtBytes, err := json.Marshal(sourceExtMap)

		if err == nil {
			source.Ext = sourceExtBytes
		}
	}
}
