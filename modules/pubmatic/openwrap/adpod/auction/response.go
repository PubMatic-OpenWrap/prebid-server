package auction

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	pbc "github.com/prebid/prebid-server/prebid_cache_client"
)

type CacheWrapperStruct struct {
	Adm    string  `json:"adm,omitempty"`
	Price  float64 `json:"price"`
	Width  int64   `json:"width,omitempty"`
	Height int64   `json:"height,omitempty"`
}

func GetAndCacheWinningBidsIds(adpodBids []*AdPodBid, endpoint string, impCtx map[string]models.ImpCtx, cacheClient pbc.Client) (map[string][]string, error) {
	var winningBidIds map[string][]string
	if len(adpodBids) == 0 {
		return winningBidIds, nil
	}

	winningBidIds = make(map[string][]string)
	for _, eachAdpodBid := range adpodBids {
		var cobjs []pbc.Cacheable
		for _, bid := range eachAdpodBid.Bids {
			if len(bid.AdM) == 0 {
				continue
			}
			winningBidIds[eachAdpodBid.OriginalImpID] = append(winningBidIds[eachAdpodBid.OriginalImpID], bid.ID)
			if endpoint == models.EndpointJson {
				cobj, err := portPrebidCacheable(bid, "video")
				if err != nil {
					return nil, err
				}
				cobjs = append(cobjs, cobj)
			}
		}

		if endpoint == models.EndpointJson {
			uuid, errs := cacheClient.PutJson(context.Background(), cobjs)
			if len(errs) != 0 {
				return nil, fmt.Errorf("prebid cache failed, error %v", errs)
			}

			bidToCacheIdMap := make(map[string]string)
			for i, bid := range eachAdpodBid.Bids {
				bidToCacheIdMap[bid.ID] = uuid[i]
			}

			eachImpCtx := impCtx[eachAdpodBid.OriginalImpID]
			eachImpCtx.BidCacheIdMap = bidToCacheIdMap
			impCtx[eachAdpodBid.OriginalImpID] = eachImpCtx
		}
	}

	return winningBidIds, nil
}

func portPrebidCacheable(bid *Bid, platform string) (pbc.Cacheable, error) {
	var err error
	var cacheBytes json.RawMessage
	var cacheType pbc.PayloadType

	if platform == "video" {
		cacheType = pbc.TypeXML
		cacheBytes, err = json.Marshal(bid.AdM)
	} else {
		cacheType = pbc.TypeJSON
		cacheBytes, err = json.Marshal(CacheWrapperStruct{
			Adm:    bid.AdM,
			Price:  bid.Price,
			Width:  bid.W,
			Height: bid.H,
		})
	}

	return pbc.Cacheable{
		Type: cacheType,
		Data: cacheBytes,
	}, err
}
