package middleware

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
	pbc "github.com/prebid/prebid-server/v2/prebid_cache_client"
)

type CacheWrapperStruct struct {
	Adm    string  `json:"adm,omitempty"`
	Price  float64 `json:"price"`
	Width  int64   `json:"width,omitempty"`
	Height int64   `json:"height,omitempty"`
}

func cacheAllBids(client *pbc.Client, bids []openrtb2.Bid) ([]string, error) {
	var cobjs []pbc.Cacheable

	for _, bid := range bids {
		if len(bid.AdM) == 0 {
			continue
		}
		cobj, err := portPrebidCacheable(bid, "video")
		if err != nil {
			return nil, err
		}
		cobjs = append(cobjs, cobj)
	}

	uuids, errs := (*client).PutJson(context.Background(), cobjs)
	if len(errs) != 0 {
		return nil, fmt.Errorf("prebid cache failed, error %v", errs)
	}

	return uuids, nil
}

func portPrebidCacheable(bid openrtb2.Bid, platform string) (pbc.Cacheable, error) {
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
