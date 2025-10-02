package ctvjson

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/creativecache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

const (
	slotKeyFormat = "s%d_%s"
)

var (
	redirectTargetingKeys = []string{"pwtpb", "pwtdur", "pwtcid", "pwtpid", "pwtdealtier", "pwtdid", "pwtdt"}
)

type CacheWrapperStruct struct {
	Adm    string  `json:"adm,omitempty"`
	Price  float64 `json:"price"`
	Width  int64   `json:"width,omitempty"`
	Height int64   `json:"height,omitempty"`
}

type adPodBid struct {
	ModifiedURL string                `json:"modifiedurl,omitempty"`
	ID          string                `json:"id,omitempty"`
	NBR         *openrtb3.NoBidReason `json:"nbr,omitempty"`
	Targeting   []map[string]string   `json:"targeting,omitempty"`
	Error       string                `json:"error,omitempty"`
	Ext         interface{}           `json:"ext,omitempty"`
}

func formCTVJSONResponse(rCtx *models.RequestCtx, response *openrtb2.BidResponse, cacheClient creativecache.Client) []*adPodBid {
	impBidMap := make(map[string][]openrtb2.Bid)
	for _, seatBid := range response.SeatBid {
		for _, bid := range seatBid.Bid {
			impBidMap[bid.ImpID] = append(impBidMap[bid.ImpID], bid)
		}
	}

	return formAdpodBids(rCtx, impBidMap, cacheClient)
}

func checkRedirectResponse(rCtx models.RequestCtx) bool {
	if rCtx.Debug {
		return false
	}

	if rCtx.RedirectURL != "" && rCtx.ResponseFormat == models.ResponseFormatRedirect {
		return true
	}

	return false
}

func prepareSlotLevelKey(slotNo int, key string) string {
	return fmt.Sprintf(slotKeyFormat, slotNo, key)
}

func formAdpodBids(rCtx *models.RequestCtx, bidsMap map[string][]openrtb2.Bid, cacheClient creativecache.Client) []*adPodBid {
	var adpodBids []*adPodBid
	for impId, bids := range bidsMap {
		adpodBid := &adPodBid{
			ID: impId,
		}

		sort.Slice(bids, func(i, j int) bool { return bids[i].Price > bids[j].Price })

		cacheIds, err := cacheAllBids(bids, cacheClient)
		if err != nil {
			adpodBid.Error = err.Error()
			adpodBids = append(adpodBids, adpodBid)
			continue
		}

		impCtx, ok := rCtx.ImpBidCtx[impId]
		if !ok {
			continue
		}

		targetings := []map[string]string{}
		for i := range bids {
			bidCtx, ok := impCtx.BidCtx[bids[i].ID]
			if !ok {
				continue
			}

			slotNo := i + 1
			targeting := getTargeting(bidCtx, slotNo, cacheIds[i])
			if len(targeting) > 0 {
				targetings = append(targetings, targeting)
			}
		}

		if len(targetings) > 0 {
			adpodBid.Targeting = targetings
		}

		if len(impCtx.AdserverURL) > 0 {
			adpodBid.ModifiedURL = updateAdServerURL(targetings, impCtx.AdserverURL)
		}

		adpodBids = append(adpodBids, adpodBid)
	}

	return adpodBids
}

func getTargeting(bidCtx models.BidCtx, slotNo int, cacheId string) map[string]string {
	targetingKeyValMap := make(map[string]string)

	if bidCtx.Prebid == nil || bidCtx.Prebid.Targeting == nil {
		return targetingKeyValMap
	}

	bidCtx.Prebid.Targeting[models.PWT_CACHEID] = cacheId
	for key, value := range bidCtx.Prebid.Targeting {
		targetingKeyValMap[prepareSlotLevelKey(slotNo, key)] = value
	}

	return targetingKeyValMap
}

func cacheAllBids(bids []openrtb2.Bid, client creativecache.Client) ([]string, error) {
	var cobjs []creativecache.Cacheable

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

	uuids, errs := client.PutJson(context.Background(), cobjs)
	if len(errs) != 0 {
		return nil, fmt.Errorf("prebid cache failed, error %v", errs)
	}

	return uuids, nil
}

func portPrebidCacheable(bid openrtb2.Bid, platform string) (creativecache.Cacheable, error) {
	var err error
	var cacheBytes json.RawMessage
	var cacheType creativecache.PayloadType

	if platform == "video" {
		cacheType = creativecache.TypeXML
		cacheBytes, err = json.Marshal(bid.AdM)
	} else {
		cacheType = creativecache.TypeJSON
		cacheBytes, err = json.Marshal(CacheWrapperStruct{
			Adm:    bid.AdM,
			Price:  bid.Price,
			Width:  bid.W,
			Height: bid.H,
		})
	}

	return creativecache.Cacheable{
		Type: cacheType,
		Data: cacheBytes,
	}, err
}

func updateAdServerURL(targetings []map[string]string, adServerURL string) string {
	redirectURL, err := url.ParseRequestURI(strings.TrimSpace(adServerURL))
	if err != nil {
		return ""
	}

	if len(targetings) == 0 {
		// This is if there are no valid bids
		return redirectURL.String()
	}

	redirectQuery := redirectURL.Query()
	cursParams, err := url.ParseQuery(strings.TrimSpace(redirectQuery.Get(models.CustParams)))
	if err != nil {
		return ""
	}

	for i, target := range targetings {
		sNo := i + 1
		for _, tk := range redirectTargetingKeys {
			targetingKey := prepareSlotLevelKey(sNo, tk)
			if value, ok := target[targetingKey]; ok {
				cursParams.Set(targetingKey, value)
			}
		}
	}

	redirectQuery.Set(models.CustParams, cursParams.Encode())
	redirectURL.RawQuery = redirectQuery.Encode()

	return redirectURL.String()
}
