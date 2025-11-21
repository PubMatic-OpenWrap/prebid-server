package ctvjson

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/prebid/openrtb/v20/adcom1"
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
	slotTargetingKeys     = map[string]struct{}{
		models.PWT_PARTNERID: {},
		models.PWT_DURATION:  {},
		models.PwtDT:         {},
		models.PWT_DEALID:    {},
		models.PwtPb:         {},
		models.PwtCat:        {},
		models.PWT_CACHEID:   {},
	}
)

type bidResponseAdpod struct {
	AdPodBids   []*adPodBid `json:"adpods,omitempty"`
	Ext         interface{} `json:"ext,omitempty"`
	RedirectURL string      `json:"redirect_url,omitempty"`
}

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

// PodPostion represent slot start and end range for pre, mid and post roll.
type PodPosition struct {
	PreRoll  Range `json:"preroll,omitempty"`
	MidRoll  Range `json:"midroll,omitempty"`
	PostRoll Range `json:"postroll,omitempty"`
}

// Range defines start and end range for pod position
type Range struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type impMeta struct {
	impID string
	video *openrtb2.Video
}

// newPodPosition is default slot position config for pre, mid and post roll
func newPodPosition() *PodPosition {
	return &PodPosition{
		PreRoll:  Range{Start: 1, End: 30},
		PostRoll: Range{Start: 31, End: 60},
		MidRoll:  Range{Start: 61, End: 90},
	}
}

func formCTVJSONResponse(rCtx *models.RequestCtx, response *openrtb2.BidResponse, cacheClient creativecache.Client) []*adPodBid {
	impBidMap := make(map[string][]openrtb2.Bid)
	for _, seatBid := range response.SeatBid {
		for _, bid := range seatBid.Bid {
			if len(bid.AdM) == 0 || bid.Price <= 0 {
				continue
			}
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
	impMetas := []impMeta{}
	for _, impCtx := range rCtx.ImpBidCtx {
		if impCtx.Video != nil {
			impMetas = append(impMetas, impMeta{
				impID: impCtx.ImpID,
				video: impCtx.Video,
			})
		}
	}

	sortImps(impMetas)

	podPostion := newPodPosition()
	preRollSlot := podPostion.PreRoll.Start - 1
	midRollSlot := podPostion.MidRoll.Start - 1
	postRollSlot := podPostion.PostRoll.Start - 1

	var adpodBids []*adPodBid
	for i := range impMetas {
		adpodBid := &adPodBid{
			ID: impMetas[i].impID,
		}

		bids, ok := bidsMap[impMetas[i].impID]
		if !ok {
			continue
		}

		sort.Slice(bids, func(i, j int) bool { return bids[i].Price > bids[j].Price })

		cacheIds, err := cacheAllBids(bids, cacheClient)
		if err != nil {
			adpodBid.Error = err.Error()
			adpodBids = append(adpodBids, adpodBid)
			continue
		}

		impCtx, ok := rCtx.ImpBidCtx[impMetas[i].impID]
		if !ok {
			continue
		}

		targetings := []map[string]string{}
		for i := range bids {
			bidCtx, ok := impCtx.BidCtx[bids[i].ID]
			if !ok {
				continue
			}

			slotNo := 0
			videoPosition := getVideoPosition(rCtx, impCtx.Video)
			if videoPosition == adcom1.StartPreRoll {
				preRollSlot = preRollSlot + 1
				slotNo = preRollSlot
			} else if videoPosition == adcom1.StartPostRoll {
				postRollSlot = postRollSlot + 1
				slotNo = postRollSlot
			} else {
				midRollSlot = midRollSlot + 1
				slotNo = midRollSlot
			}

			targeting := getTargeting(bidCtx, slotNo, cacheIds[i])
			if len(targeting) > 0 {
				targetings = append(targetings, targeting)
			}

			if !rCtx.Debug {
				delete(targeting, models.PwtPbCatDur)
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
		if _, ok := slotTargetingKeys[key]; ok {
			targetingKeyValMap[prepareSlotLevelKey(slotNo, key)] = value
			continue
		}
		targetingKeyValMap[key] = value
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

func sortImps(imps []impMeta) {
	sort.Slice(imps, func(i, j int) bool {
		// First, sort by StartDelay category (pre-roll, mid-roll, post-roll)

		videoPositionI := categoriseVideoPosition(imps[i].video.StartDelay)
		videoPositionJ := categoriseVideoPosition(imps[j].video.StartDelay)
		if videoPositionI != videoPositionJ {
			return videoPositionI < videoPositionJ
		}

		// For mid-roll, further sort by StartDelay value
		if videoPositionI == 2 && *imps[i].video.StartDelay != *imps[j].video.StartDelay {
			return *imps[i].video.StartDelay < *imps[j].video.StartDelay
		}

		// Finally, sort by PodID
		return getPodSequencePriority(imps[i].video.PodSeq) < getPodSequencePriority(imps[j].video.PodSeq)
	})
}

// Determines the category of the StartDelay for sorting
// 0: pre-roll, 1: mid-roll, 2: post-roll
func categoriseVideoPosition(delay *adcom1.StartDelay) int {
	if delay == nil {
		return 1 // Treat nil as highest priority (pre-roll)
	}
	switch {
	case *delay == 0:
		return 0 // pre-roll
	case *delay > 0:
		return 2 // mid-roll
	case *delay == -1:
		return 3 // mid-roll
	case *delay == -2:
		return 4 // post-roll
	default:
		return 5 // post-roll
	}
}

func getPodSequencePriority(podSeq adcom1.PodSequence) int {
	switch {
	case podSeq == adcom1.PodSeqFirst:
		return 0
	case podSeq == adcom1.PodSeqAny:
		return 1
	case podSeq == adcom1.PodSeqLast:
		return 2
	default:
		return 2
	}
}

func getVideoPosition(rctx *models.RequestCtx, video *openrtb2.Video) adcom1.StartDelay {
	if !rctx.AdruleFlag || video.StartDelay == nil {
		return adcom1.StartPreRoll
	}

	return video.StartDelay.Val()
}
