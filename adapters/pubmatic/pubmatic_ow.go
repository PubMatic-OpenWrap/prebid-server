package pubmatic

import (
	"encoding/json"
	"strconv"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

func getTargetingKeys(bidExt json.RawMessage, bidderName string) map[string]string {
	targets := map[string]string{}
	if bidExt != nil {
		bidExtMap := make(map[string]interface{})
		err := json.Unmarshal(bidExt, &bidExtMap)
		if err == nil && bidExtMap[buyId] != nil {
			targets[buyIdTargetingKey+bidderName], _ = bidExtMap[buyId].(string)
		}
	}
	return targets
}

func copySBExtToBidExt(sbExt json.RawMessage, bidExt json.RawMessage) json.RawMessage {
	if sbExt != nil {
		sbExtMap := getMapFromJSON(sbExt)
		bidExtMap := make(map[string]interface{})
		if bidExt != nil {
			bidExtMap = getMapFromJSON(bidExt)
		}
		if bidExtMap != nil && sbExtMap != nil {
			if sbExtMap[buyId] != nil && bidExtMap[buyId] == nil {
				bidExtMap[buyId] = sbExtMap[buyId]
			}
		}
		byteAra, _ := json.Marshal(bidExtMap)
		return json.RawMessage(byteAra)
	}
	return bidExt
}

// prepareMetaObject prepares the Meta structure using Bid Response
func prepareMetaObject(bid openrtb2.Bid, bidExt *pubmaticBidExt, seat string) *openrtb_ext.ExtBidPrebidMeta {

	meta := &openrtb_ext.ExtBidPrebidMeta{
		NetworkID:    bidExt.DspId,
		AdvertiserID: bidExt.AdvertiserID,
		MediaType:    string(getBidType(bidExt)),
	}

	if meta.NetworkID != 0 {
		meta.DemandSource = strconv.Itoa(meta.NetworkID)
	}

	if len(seat) > 0 {
		meta.AdvertiserID, _ = strconv.Atoi(seat)
	}

	meta.AgencyID = meta.AdvertiserID

	if len(bid.Cat) > 0 {
		meta.PrimaryCategoryID = bid.Cat[0]
		meta.SecondaryCategoryIDs = bid.Cat
	}

	// NOTE: We will not recieve below fields from the translator response also not sure on what will be the key names for these in the response,
	// when we needed we can add it back.
	// New fields added, assignee fields name may change
	// Assign meta.BrandId to bidExt.ADomain[0]  //BrandID is of Type int and ADomain values if string type like "mystartab.com"
	// meta.NetworkName = bidExt.NetworkName;
	// meta.AdvertiserName = bidExt.AdvertiserName;
	// meta.AgencyName = bidExt.AgencyName;
	// meta.BrandName = bidExt.BrandName;
	// meta.DChain = bidExt.DChain;

	return meta
}
