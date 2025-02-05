package openwrap

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/endpoints/legacy/ctv"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
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

func setSChainInRequest(requestExt *models.RequestExt, source *openrtb2.Source, partnerConfigMap map[int]map[string]string) {
	setGlobalSChain(source, partnerConfigMap)
	setAllBidderSChain(requestExt, partnerConfigMap)
}

func getSChainObj(partnerConfigMap map[int]map[string]string) *openrtb2.SupplyChain {
	sChainObjJSON := models.GetVersionLevelPropertyFromPartnerConfig(partnerConfigMap, models.SChainObjectDBKey)
	if len(sChainObjJSON) == 0 {
		return nil
	}
	sChainConfig := &SupplyChainConfig{}
	if err := json.Unmarshal([]byte(sChainObjJSON), sChainConfig); err != nil {
		glog.Errorf(ctv.ErrJSONUnmarshalFailed, models.SChainKey, err.Error(), sChainObjJSON)
		return nil
	}
	if sChainConfig.SupplyChain != nil {
		return sChainConfig.SupplyChain
	}
	return nil
}

// setGlobalSChain sets schain object in source.ext.schain
func setGlobalSChain(source *openrtb2.Source, partnerConfigMap map[int]map[string]string) {
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
		sourceExtMap[models.SChainKey] = sChainObj
		sourceExtBytes, err := json.Marshal(sourceExtMap)

		if err == nil {
			source.Ext = sourceExtBytes
		}
	}
}

// setAllBidderSChain sets All Bidder Specific Schain to ext.prebid.schains
func setAllBidderSChain(requestExt *models.RequestExt, partnerConfigMap map[int]map[string]string) {
	if requestExt == nil {
		return
	}

	if requestExt.Prebid.SChains != nil && len(requestExt.Prebid.SChains) > 0 {
		return
	}

	allBidderSChainObjJSON := models.GetVersionLevelPropertyFromPartnerConfig(partnerConfigMap, models.AllBidderSChainObj)
	if len(allBidderSChainObjJSON) == 0 {
		return
	}

	allBidderSChainConfig := []*openrtb_ext.ExtRequestPrebidSChain{}
	if err := json.Unmarshal([]byte(allBidderSChainObjJSON), &allBidderSChainConfig); err != nil {
		glog.Errorf(ctv.ErrJSONUnmarshalFailed, models.AllBidderSChainKey, err.Error(), allBidderSChainObjJSON)
		return
	}
	requestExt.Prebid.SChains = allBidderSChainConfig
}
