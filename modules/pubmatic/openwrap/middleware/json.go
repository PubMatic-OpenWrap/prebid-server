package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	pbc "github.com/prebid/prebid-server/prebid_cache_client"
)

const (
	slotKeyFormat = "s%d_%s"
)

type jsonBid struct {
	*openrtb2.Bid
	Seat string
}

type adPodBid struct {
	ID        *string             `json:"id,omitempty"`
	Targeting []map[string]string `json:"targeting,omitempty"`
	Error     string              `json:"error,omitempty"`
	Ext       interface{}         `json:"ext,omitempty"`
}

type bidResponseAdpod struct {
	AdPodBids   []*adPodBid `json:"adpods,omitempty"`
	Ext         interface{} `json:"ext,omitempty"`
	RedirectURL *string     `json:"redirect_url,omitempty"`
}

type CacheWrapperStruct struct {
	Adm    string  `json:"adm,omitempty"`
	Price  float64 `json:"price"`
	Width  int64   `json:"width,omitempty"`
	Height int64   `json:"height,omitempty"`
}

func FormJSONResponse(cacheClient *pbc.Client, response []byte) []byte {
	bidResponse := openrtb2.BidResponse{}

	err := json.Unmarshal(response, &bidResponse)
	if err != nil {
		return response
	}

	jsonResponse, err := getJsonResponse(cacheClient, &bidResponse)
	if err != nil {
		return response
	}

	return jsonResponse
}

func getJsonResponse(client *pbc.Client, bidResponse *openrtb2.BidResponse) ([]byte, error) {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return nil, errors.New("recieved invalid bidResponse")
	}

	bidArrayMap := make(map[string][]jsonBid)
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			impId, _ := models.GetImpressionID(bid.ImpID)
			bids, ok := bidArrayMap[impId]
			if !ok {
				bids = make([]jsonBid, 0)
			}
			bids = append(bids, jsonBid{Bid: &bid, Seat: seatBid.Seat})
			bidArrayMap[impId] = bids
		}
	}

	adPodBids := formAdpodBids(client, bidArrayMap)

	adPodResponse := bidResponseAdpod{
		AdPodBids: adPodBids,
		Ext:       bidResponse.Ext,
	}

	data, err := json.Marshal(adPodResponse)
	if err != nil {
		return nil, err
	}

	return data, nil

}

func formAdpodBids(client *pbc.Client, bidsMap map[string][]jsonBid) []*adPodBid {
	var adpodBids []*adPodBid
	for impId, bids := range bidsMap {
		adpodBid := adPodBid{
			ID: &impId,
		}
		sort.Slice(bids, func(i, j int) bool { return bids[i].Price > bids[j].Price })

		cacheIds, err := cacheAllBids(client, bids)
		if err != nil {
			adpodBid.Error = err.Error()
			continue
		}

		targetings := []map[string]string{}
		for i := 0; i < len(bids); i++ {
			slotNo := i + 1
			targeting := createTargetting(bids[i], slotNo, cacheIds[i])
			targetings = append(targetings, targeting)
		}
		adpodBid.Targeting = targetings
		adpodBids = append(adpodBids, &adpodBid)
	}

	return adpodBids
}

func prepareSlotLevelKey(slotNo int, key string) string {
	return fmt.Sprintf(slotKeyFormat, slotNo, key)
}

func createTargetting(bid jsonBid, slotNo int, cacheId string) map[string]string {
	targetingKeyValMap := make(map[string]string)

	targetingKeyValMap[prepareSlotLevelKey(slotNo, models.PWT_PARTNERID)] = bid.Seat
	targetingKeyValMap[prepareSlotLevelKey(slotNo, models.PWT_CACHEID)] = cacheId
	if len(bid.DealID) > 0 {
		targetingKeyValMap[prepareSlotLevelKey(slotNo, models.PWT_DEALID)] = bid.DealID
	} else {
		targetingKeyValMap[prepareSlotLevelKey(slotNo, models.PWT_DEALID)] = models.DealIDNotApplicable
	}

	if len(bid.Ext) > 0 {
		fmt.Println(string(bid.Ext))
		bidExt := models.BidExt{}
		err := json.Unmarshal(bid.Ext, &bidExt)
		if err != nil {
			return targetingKeyValMap
		}

		dealTier := models.DealTierNotApplicable
		// add deal tier from ext
		// if bidExt.Prebid != nil && bidExt.Prebid.DealTierSatisfied {

		// }
		targetingKeyValMap[prepareSlotLevelKey(slotNo, models.PwtDealTier)] = dealTier

		priceBucket, ok := bidExt.Prebid.Targeting[models.PwtPb]
		if ok {
			targetingKeyValMap[prepareSlotLevelKey(slotNo, models.PwtPb)] = priceBucket
		}

		if bidExt.Prebid.Video != nil && bidExt.Prebid.Video.Duration > 0 {
			targetingKeyValMap[prepareSlotLevelKey(slotNo, models.PWT_DURATION)] = strconv.Itoa(bidExt.Prebid.Video.Duration)
		}

		catDur, ok := bidExt.Prebid.Targeting[models.PwtPbCatDur]
		if ok {
			cat, dur := getCatAndDurFromPwtCatDur(catDur)
			if len(cat) > 0 {
				targetingKeyValMap[prepareSlotLevelKey(slotNo, models.PwtCat)] = cat
			}

			if len(dur) > 0 {
				targetingKeyValMap[prepareSlotLevelKey(slotNo, models.PWT_DURATION)] = dur
			}
		}
	}

	return targetingKeyValMap

}

func cacheAllBids(client *pbc.Client, bids []jsonBid) ([]string, error) {
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

func portPrebidCacheable(bid jsonBid, platform string) (pbc.Cacheable, error) {
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

func getCatAndDurFromPwtCatDur(pwtCatDur string) (string, string) {
	arr := strings.Split(pwtCatDur, "_")
	if len(arr) == 2 {
		return "", TrimRightByte(arr[1], 's')
	}
	if len(arr) == 3 {
		return arr[1], TrimRightByte(arr[2], 's')
	}
	return "", ""
}

func TrimRightByte(s string, b byte) string {
	if s[len(s)-1] == b {
		return s[:len(s)-1]
	}
	return s
}
