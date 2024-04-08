package adpod

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/buger/jsonparser"
	"github.com/gofrs/uuid"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/endpoints/events"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/combination"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/impressions"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/response"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/util"
	"github.com/prebid/prebid-server/metrics"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type DynamicAdpod struct {
	AdpodCtx
	MinPodDuration int64                         `json:"-"`
	MaxPodDuration int64                         `json:"-"`
	MaxExtended    int64                         `json:"-"`
	Imp            openrtb2.Imp                  `json:"-"`
	VideoExt       *openrtb_ext.ExtVideoAdPod    `json:"vidext,omitempty"`
	ImpConfigs     []*types.ImpAdPodConfig       `json:"imp,omitempty"`
	AdpodBid       *types.AdPodBid               `json:"-"`
	Error          *openrtb_ext.ExtBidderMessage `json:"ec,omitempty"`
}

func (da *DynamicAdpod) GetPodType() PodType {
	return da.Type
}

func (da *DynamicAdpod) AddImpressions(imp openrtb2.Imp) {
	da.Imps = append(da.Imps, imp)
}

func (sa *DynamicAdpod) GetImpressions() []openrtb2.Imp {
	return sa.Imps
}

func (da *DynamicAdpod) GenerateImpressions() {
	da.getAdPodImpConfigs()

	// Generate Impressions based on configs
	for i := range da.ImpConfigs {
		imp := newImpression(da.Imp, da.ImpConfigs[i])
		da.AddImpressions(imp)
	}
}

func (da *DynamicAdpod) CollectBid(bid openrtb2.Bid, seat string) {
	originalImpId, sequence := util.DecodeImpressionID(bid.ImpID)

	if da.AdpodBid == nil {
		da.AdpodBid = &types.AdPodBid{
			Bids:          make([]*types.Bid, 0),
			OriginalImpID: originalImpId,
			SeatName:      string(openrtb_ext.BidderOWPrebidCTV),
		}
	}

	value, err := util.GetTargeting(openrtb_ext.HbCategoryDurationKey, openrtb_ext.BidderName(seat), bid)
	if err == nil {
		// ignore error
		addTargetingKey(&bid, openrtb_ext.HbCategoryDurationKey, value)
	}

	value, err = util.GetTargeting(openrtb_ext.HbpbConstantKey, openrtb_ext.BidderName(seat), bid)
	if err == nil {
		// ignore error
		addTargetingKey(&bid, openrtb_ext.HbpbConstantKey, value)
	}

	ext := openrtb_ext.ExtBid{}
	if bid.Ext != nil {
		json.Unmarshal(bid.Ext, &ext)
	}

	//get duration of creative
	duration, status := getBidDuration(&bid, da.ReqExt, da.ImpConfigs, da.ImpConfigs[sequence-1].MaxDuration)

	da.AdpodBid.Bids = append(da.AdpodBid.Bids, &types.Bid{
		Bid:               &bid,
		ExtBid:            ext,
		Status:            status,
		Duration:          int(duration),
		DealTierSatisfied: util.GetDealTierSatisfied(&ext),
		Seat:              seat,
	})
}

func (da *DynamicAdpod) PerformAuctionAndExclusion() {
	if da.AdpodBid == nil || len(da.AdpodBid.Bids) == 0 {
		return
	}

	// Check if we need sorting
	// sort.Slice(da.AdpodBid.Bids, func(i, j int) bool { return da.AdpodBid.Bids[i].Price > da.AdpodBid.Bids[j].Price })

	buckets := util.GetDurationWiseBidsBucket(da.AdpodBid.Bids)
	if len(buckets) == 0 {
		da.Error = util.DurationMismatchWarning
		return
	}

	//combination generator
	comb := combination.NewCombination(buckets, uint64(da.MinPodDuration), uint64(da.MaxPodDuration), da.VideoExt.AdPod)

	//adpod generator
	adpodGenerator := response.NewAdPodGenerator(buckets, comb, da.VideoExt.AdPod, da.MetricsEngine)

	adpodBid := adpodGenerator.GetAdPodBids()
	if adpodBid == nil {
		da.Error = util.UnableToGenerateAdPodWarning
		return
	}
	adpodBid.OriginalImpID = da.AdpodBid.OriginalImpID
	adpodBid.SeatName = da.AdpodBid.SeatName

	// Update the original adpodBid
	da.AdpodBid = adpodBid

}

func (da *DynamicAdpod) Validate() []error {
	var valdiationErrs []error

	if da.VideoExt == nil {
		return valdiationErrs
	}

	extErrs := da.VideoExt.Validate()
	if len(extErrs) > 0 {
		valdiationErrs = append(valdiationErrs, extErrs...)
	}

	durationErrs := da.VideoExt.AdPod.ValidateAdPodDurations(da.MinPodDuration, da.MaxPodDuration, da.MaxExtended)
	if len(durationErrs) > 0 {
		valdiationErrs = append(valdiationErrs, durationErrs...)
	}

	return valdiationErrs
}

func (da *DynamicAdpod) GetAdpodSeatBids() []openrtb2.SeatBid {
	// Record Rejected bids
	da.recordRejectedAdPodBids(da.PubId)

	return da.getBidResponseSeatBids()
}

func (da *DynamicAdpod) GetAdpodExtension(blockedVastTagID map[string]map[string][]string) *types.ImpData {
	da.setBidExtParams()

	data := types.ImpData{
		ImpID:           da.Imp.ID,
		Bid:             da.AdpodBid,
		VideoExt:        da.VideoExt,
		Config:          da.ImpConfigs,
		BlockedVASTTags: blockedVastTagID[da.Imp.ID],
		Error:           da.Error,
	}

	return &data
}

/***************************** Dynamic adpod processing method ************************************/

// getAdPodImpsConfigs will return number of impressions configurations within adpod
func (da *DynamicAdpod) getAdPodImpConfigs() {
	// monitor
	start := time.Now()
	selectedAlgorithm := impressions.SelectAlgorithm(da.ReqExt)
	impGen := impressions.NewImpressions(da.MinPodDuration, da.MaxPodDuration, da.ReqExt, da.VideoExt.AdPod, selectedAlgorithm)
	impRanges := impGen.Get()
	labels := metrics.PodLabels{AlgorithmName: impressions.MonitorKey[selectedAlgorithm], NoOfImpressions: new(int)}

	//log number of impressions in stats
	*labels.NoOfImpressions = len(impRanges)
	da.MetricsEngine.RecordPodImpGenTime(labels, start)

	// check if algorithm has generated impressions
	if len(impRanges) == 0 {
		da.Error = &openrtb_ext.ExtBidderMessage{
			Code:    util.UnableToGenerateImpressionsError.Code(),
			Message: util.UnableToGenerateImpressionsError.Message,
		}
		return
	}

	config := make([]*types.ImpAdPodConfig, len(impRanges))
	for i, value := range impRanges {
		config[i] = &types.ImpAdPodConfig{
			ImpID:          util.GetCTVImpressionID(da.Imp.ID, i+1),
			MinDuration:    value[0],
			MaxDuration:    value[1],
			SequenceNumber: int8(i + 1), /* Must be starting with 1 */
		}
	}

	da.ImpConfigs = config
}

// newImpression will clone existing impression object and create video object with ImpAdPodConfig.
func newImpression(imp openrtb2.Imp, config *types.ImpAdPodConfig) openrtb2.Imp {
	video := *imp.Video
	video.MinDuration = config.MinDuration
	video.MaxDuration = config.MaxDuration
	video.Sequence = config.SequenceNumber
	video.MaxExtended = 0
	//TODO: remove video adpod extension if not required

	newImp := imp
	newImp.ID = config.ImpID
	//newImp.BidFloor = 0
	newImp.Video = &video
	return newImp
}

/*
getBidDuration determines the duration of video ad from given bid.
it will try to get the actual ad duration returned by the bidder using prebid.video.duration
if prebid.video.duration not present then uses defaultDuration passed as an argument
if video lengths matching policy is present for request then it will validate and update duration based on policy
*/
func getBidDuration(bid *openrtb2.Bid, reqExt *openrtb_ext.ExtRequestAdPod, config []*types.ImpAdPodConfig, defaultDuration int64) (int64, constant.BidStatus) {

	// C1: Read it from bid.ext.prebid.video.duration field
	duration, err := jsonparser.GetInt(bid.Ext, "prebid", "video", "duration")
	if nil != err || duration <= 0 {
		// incase if duration is not present use impression duration directly as it is
		return defaultDuration, constant.StatusOK
	}

	// C2: Based on video lengths matching policy validate and return duration
	if nil != reqExt && len(reqExt.VideoAdDurationMatching) > 0 {
		return getDurationBasedOnDurationMatchingPolicy(duration, reqExt.VideoAdDurationMatching, config)
	}

	//default return duration which is present in bid.ext.prebid.vide.duration field
	return duration, constant.StatusOK
}

// getDurationBasedOnDurationMatchingPolicy will return duration based on durationmatching policy
func getDurationBasedOnDurationMatchingPolicy(duration int64, policy openrtb_ext.OWVideoAdDurationMatchingPolicy, config []*types.ImpAdPodConfig) (int64, constant.BidStatus) {
	switch policy {
	case openrtb_ext.OWExactVideoAdDurationMatching:
		tmp := util.GetNearestDuration(duration, config)
		if tmp != duration {
			return duration, constant.StatusDurationMismatch
		}
		//its and valid duration return it with StatusOK

	case openrtb_ext.OWRoundupVideoAdDurationMatching:
		tmp := util.GetNearestDuration(duration, config)
		if tmp == -1 {
			return duration, constant.StatusDurationMismatch
		}
		//update duration with nearest one duration
		duration = tmp
		//its and valid duration return it with StatusOK
	}

	return duration, constant.StatusOK
}

/***************************Bid Response Processing************************/

func (da *DynamicAdpod) getBidResponseSeatBids() []openrtb2.SeatBid {
	if da.AdpodBid == nil || len(da.AdpodBid.Bids) == 0 {
		return nil
	}

	bid := da.getAdPodBid(da.AdpodBid)
	if bid == nil {
		return nil
	}

	adpodSeat := openrtb2.SeatBid{
		Seat: da.AdpodBid.SeatName,
	}
	adpodSeat.Bid = append(adpodSeat.Bid, *bid.Bid)

	return []openrtb2.SeatBid{adpodSeat}
}

// getAdPodBid
func (da *DynamicAdpod) getAdPodBid(adpod *types.AdPodBid) *types.Bid {
	bid := types.Bid{
		Bid: &openrtb2.Bid{},
	}

	//TODO: Write single for loop to get all details
	bidID, err := uuid.NewV4()
	if nil == err {
		bid.ID = bidID.String()
	} else {
		bid.ID = adpod.Bids[0].ID
	}

	bid.ImpID = adpod.OriginalImpID
	bid.Price = adpod.Price
	bid.ADomain = adpod.ADomain[:]
	bid.Cat = adpod.Cat[:]
	bid.AdM = *getAdPodBidCreative(da.Imp.Video, adpod, true)
	bid.Ext = getAdPodBidExtension(adpod)
	return &bid
}

// getAdPodBidCreative get commulative adpod bid details
func getAdPodBidCreative(video *openrtb2.Video, adpod *types.AdPodBid, generatedBidID bool) *string {
	doc := etree.NewDocument()
	vast := doc.CreateElement(constant.VASTElement)
	sequenceNumber := 1
	var version float64 = 2.0

	for _, bid := range adpod.Bids {
		var newAd *etree.Element

		if strings.HasPrefix(bid.AdM, constant.HTTPPrefix) {
			newAd = etree.NewElement(constant.VASTAdElement)
			wrapper := newAd.CreateElement(constant.VASTWrapperElement)
			vastAdTagURI := wrapper.CreateElement(constant.VASTAdTagURIElement)
			vastAdTagURI.CreateCharData(bid.AdM)
		} else {
			adDoc := etree.NewDocument()
			if err := adDoc.ReadFromString(bid.AdM); err != nil {
				continue
			}

			if generatedBidID == false {
				// adjust bidid in video event trackers and update
				adjustBidIDInVideoEventTrackers(adDoc, bid.Bid)
				adm, err := adDoc.WriteToString()
				if nil != err {
					util.JLogf("ERROR, %v", err.Error())
				} else {
					bid.AdM = adm
				}
			}

			vastTag := adDoc.SelectElement(constant.VASTElement)

			//Get Actual VAST Version
			bidVASTVersion, _ := strconv.ParseFloat(vastTag.SelectAttrValue(constant.VASTVersionAttribute, constant.VASTDefaultVersionStr), 64)
			version = math.Max(version, bidVASTVersion)

			ads := vastTag.SelectElements(constant.VASTAdElement)
			if len(ads) > 0 {
				newAd = ads[0].Copy()
			}
		}

		if nil != newAd {
			//creative.AdId attribute needs to be updated
			newAd.CreateAttr(constant.VASTSequenceAttribute, fmt.Sprint(sequenceNumber))
			vast.AddChild(newAd)
			sequenceNumber++
		}
	}

	if int(version) > len(constant.VASTVersionsStr) {
		version = constant.VASTMaxVersion
	}

	vast.CreateAttr(constant.VASTVersionAttribute, constant.VASTVersionsStr[int(version)])
	bidAdM, err := doc.WriteToString()
	if err != nil {
		fmt.Printf("ERROR, %v", err.Error())
		return nil
	}
	return &bidAdM
}

func adjustBidIDInVideoEventTrackers(doc *etree.Document, bid *openrtb2.Bid) {
	// adjusment: update bid.id with ctv module generated bid.id
	creatives := events.FindCreatives(doc)
	for _, creative := range creatives {
		trackingEvents := creative.FindElements("TrackingEvents/Tracking")
		if nil != trackingEvents {
			// update bidid= value with ctv generated bid id for this bid
			for _, trackingEvent := range trackingEvents {
				u, e := url.Parse(trackingEvent.Text())
				if nil == e {
					values, e := url.ParseQuery(u.RawQuery)
					// only do replacment if operId=8
					if nil == e && nil != values["bidid"] && nil != values["operId"] && values["operId"][0] == "8" {
						values.Set("bidid", bid.ID)
					} else {
						continue
					}

					//OTT-183: Fix
					if nil != values["operId"] && values["operId"][0] == "8" {
						operID := values.Get("operId")
						values.Del("operId")
						values.Add("_operId", operID) // _ (underscore) will keep it as first key
					}

					u.RawQuery = values.Encode() // encode sorts query params by key. _ must be first (assuing no other query param with _)
					// replace _operId with operId
					u.RawQuery = strings.ReplaceAll(u.RawQuery, "_operId", "operId")
					trackingEvent.SetText(u.String())
				}
			}
		}
	}
}

// getAdPodBidExtension get commulative adpod bid details
func getAdPodBidExtension(adpod *types.AdPodBid) json.RawMessage {
	bidExt := &openrtb_ext.ExtOWBid{
		ExtBid: openrtb_ext.ExtBid{
			Prebid: &openrtb_ext.ExtBidPrebid{
				Type:  openrtb_ext.BidTypeVideo,
				Video: &openrtb_ext.ExtBidPrebidVideo{},
			},
		},
		AdPod: &openrtb_ext.BidAdPodExt{
			RefBids: make([]string, len(adpod.Bids)),
		},
	}

	for i, bid := range adpod.Bids {
		//get unique bid id
		bidID := bid.ID
		if bid.ExtBid.Prebid != nil && bid.ExtBid.Prebid.BidId != "" {
			bidID = bid.ExtBid.Prebid.BidId
		}

		//adding bid id in adpod.refbids
		bidExt.AdPod.RefBids[i] = bidID

		//updating exact duration of adpod creative
		bidExt.Prebid.Video.Duration += int(bid.Duration)

		//setting bid status as winning bid
		bid.Status = constant.StatusWinningBid
	}
	rawExt, _ := json.Marshal(bidExt)
	return rawExt
}

// recordRejectedAdPodBids records the bids lost in ad-pod auction using metricsEngine
func (da *DynamicAdpod) recordRejectedAdPodBids(pubID string) {
	if da.AdpodBid != nil && len(da.AdpodBid.Bids) > 0 {
		for _, bid := range da.AdpodBid.Bids {
			if bid.Status != constant.StatusWinningBid {
				reason := ConvertAPRCToNBRC(bid.Status)
				if reason == nil {
					continue
				}
				rejReason := strconv.FormatInt(int64(*reason), 10)
				da.MetricsEngine.RecordRejectedBids(pubID, bid.Seat, rejReason)
			}
		}
	}

}

// setBidExtParams function sets the prebid.video.duration and adpod.aprc parameters
func (da *DynamicAdpod) setBidExtParams() {
	if da.AdpodBid != nil {
		for _, bid := range da.AdpodBid.Bids {
			//update adm
			//bid.AdM = constant.VASTDefaultTag

			//add duration value
			raw, err := jsonparser.Set(bid.Ext, []byte(strconv.Itoa(int(bid.Duration))), "prebid", "video", "duration")
			if nil == err {
				bid.Ext = raw
			}

			//add bid filter reason value
			raw, err = jsonparser.Set(bid.Ext, []byte(strconv.FormatInt(bid.Status, 10)), "adpod", "aprc")
			if nil == err {
				bid.Ext = raw
			}
		}
	}

}
