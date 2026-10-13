package openwrap

import (
	"bytes"
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/endpoints/legacy/ctv"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v4/openrtb_ext"
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

const (
	schainASIAppLovin = "applovin.com"
	schainASIGoogle   = "google.com"
)

func schainNodeASI(endpoint string) string {
	switch endpoint {
	case models.EndpointAppLovinMax:
		return schainASIAppLovin
	case models.EndpointGoogleSDK:
		return schainASIGoogle
	default:
		return ""
	}
}

func (m OpenWrap) updateSchain(rctx *models.RequestCtx, request *openrtb2.BidRequest) {
	if rctx.Endpoint != models.EndpointAppLovinMax && rctx.Endpoint != models.EndpointGoogleSDK {
		return
	}

	if request == nil || request.Source == nil {
		return
	}

	// Get the ASI for the endpoint
	asi := schainNodeASI(rctx.Endpoint)
	if asi == "" {
		return
	}

	// Remove the node from the schain object and source.ext.schain
	if removeSchainNode(request.Source, asi) {
		glog.V(models.LogLevelDebug).Infof("Removed %s node from schain object from request", asi)
		m.metricEngine.RecordRequestWithSchainNodeRemoved(rctx.Endpoint)
	}
}

// removeSchainNode removes nodes whose ASI matches asi from Source.SChain and source.ext.schain.
func removeSchainNode(src *openrtb2.Source, asi string) (removed bool) {
	if src == nil || asi == "" {
		return false
	}

	if isRemoved := removeNode(src.SChain, asi); isRemoved {
		removed = true
	}
	return removeNodeFromSourceExt(src, asi) || removed
}

// removeNode removes supply-chain nodes whose ASI matches asi.
func removeNode(schain *openrtb2.SupplyChain, asi string) (removed bool) {
	if schain == nil || len(schain.Nodes) == 0 || asi == "" {
		return false
	}

	filtered := schain.Nodes[:0]
	for _, n := range schain.Nodes {
		if n.ASI == asi {
			removed = true
			continue
		}
		filtered = append(filtered, n)
	}
	if removed {
		schain.Nodes = filtered
	}
	return removed
}

// removeNodeFromSourceExt removes matching nodes from source.ext.schain when that object exists.
func removeNodeFromSourceExt(src *openrtb2.Source, asi string) (removed bool) {
	if len(src.Ext) == 0 || asi == "" {
		return false
	}

	schainRaw, _, _, err := jsonparser.Get(src.Ext, "schain")
	if err != nil {
		return false
	}

	// Avoid a full unmarshal when the node is not present.
	if !bytes.Contains(schainRaw, []byte(asi)) {
		return false
	}

	var schain openrtb2.SupplyChain
	if err := json.Unmarshal(schainRaw, &schain); err != nil {
		return false
	}

	isRemoved := removeNode(&schain, asi)
	if !isRemoved {
		return false
	}

	updated, err := json.Marshal(&schain)
	if err != nil {
		return false
	}

	src.Ext, err = jsonparser.Set(src.Ext, updated, "schain")
	if err != nil {
		return false
	}
	return true
}
