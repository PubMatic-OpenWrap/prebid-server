package auction

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/gofrs/uuid"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/endpoints/events"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

const (
	//VAST Constants
	VASTDefaultVersion    = 2.0
	VASTMaxVersion        = 4.0
	VASTDefaultVersionStr = `2.0`
	VASTDefaultTag        = `<VAST version="` + VASTDefaultVersionStr + `"/>`
	VASTElement           = `VAST`
	VASTAdElement         = `Ad`
	VASTWrapperElement    = `Wrapper`
	VASTAdTagURIElement   = `VASTAdTagURI`
	VASTVersionAttribute  = `version`
	VASTSequenceAttribute = `sequence`
	HTTPPrefix            = `http`
)

var (
	VASTVersionsStr = []string{"0", "1.0", "2.0", "3.0", "4.0"}
)

const (
	//StatusOK ...
	StatusOK int64 = 0
	//StatusWinningBid ...
	StatusWinningBid int64 = 1
	//StatusCategoryExclusion ...
	StatusCategoryExclusion int64 = 2
	//StatusDomainExclusion ...
	StatusDomainExclusion int64 = 3
	//StatusDurationMismatch ...
	StatusDurationMismatch int64 = 4
)

func getAdpodSeatBids(adpodBids []*AdPodBid, impCtxMap map[string]models.ImpCtx) *openrtb2.SeatBid {
	var adpodSeatBid *openrtb2.SeatBid

	for _, adpod := range adpodBids {
		if len(adpod.Bids) == 0 {
			continue
		}
		bid := getAdPodBid(adpod, impCtxMap)
		if bid != nil {
			if adpodSeatBid ==  nil{
				adpodSeatBid = &openrtb2.SeatBid{
					Seat: adpod.SeatName,
				}
			}
			adpodSeatBid.Bid = append(adpodSeatBid.Bid, *bid.Bid)
		}
	}

	return adpodSeatBid
}

func getAdPodBid(adpod *AdPodBid, impCtxMap map[string]models.ImpCtx) *Bid {
	bid := Bid{
		Bid: &openrtb2.Bid{},
	}
	impCtx := impCtxMap[adpod.OriginalImpID]

	//TODO: Write single for loop to get all details
	bidID, err := uuid.NewV4()
	if err == nil {
		bid.ID = bidID.String()
	} else {
		bid.ID = adpod.Bids[0].ID
	}

	bid.ImpID = adpod.OriginalImpID
	bid.Price = adpod.Price
	bid.ADomain = adpod.ADomain[:]
	bid.Cat = adpod.Cat[:]
	bid.AdM = getAdPodBidCreative(impCtx.Video, adpod, true) // assuming generateBidId will be now always true
	bid.Ext = getAdPodBidExtension(adpod)
	return &bid
}

// getAdPodBidCreative get commulative adpod bid details
func getAdPodBidCreative(video *openrtb2.Video, adpod *AdPodBid, generatedBidID bool) string {
	doc := etree.NewDocument()
	vast := doc.CreateElement(VASTElement)
	sequenceNumber := 1
	var version float64 = 2.0

	for _, bid := range adpod.Bids {
		var newAd *etree.Element

		if strings.HasPrefix(bid.AdM, HTTPPrefix) {
			newAd = etree.NewElement(VASTAdElement)
			wrapper := newAd.CreateElement(VASTWrapperElement)
			vastAdTagURI := wrapper.CreateElement(VASTAdTagURIElement)
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
				if err == nil {
					bid.AdM = adm
				}
			}

			vastTag := adDoc.SelectElement(VASTElement)

			//Get Actual VAST Version
			bidVASTVersion, _ := strconv.ParseFloat(vastTag.SelectAttrValue(VASTVersionAttribute, VASTDefaultVersionStr), 64)
			version = math.Max(version, bidVASTVersion)

			ads := vastTag.SelectElements(VASTAdElement)
			if len(ads) > 0 {
				newAd = ads[0].Copy()
			}
		}

		if newAd != nil {
			//creative.AdId attribute needs to be updated
			newAd.CreateAttr(VASTSequenceAttribute, fmt.Sprint(sequenceNumber))
			vast.AddChild(newAd)
			sequenceNumber++
		}
	}

	if int(version) > len(VASTVersionsStr) {
		version = VASTMaxVersion
	}

	vast.CreateAttr(VASTVersionAttribute, VASTVersionsStr[int(version)])
	bidAdM, err := doc.WriteToString()
	if err != nil {
		fmt.Printf("ERROR, %v", err.Error())
		return ""
	}
	return bidAdM
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
				if e == nil {
					values, e := url.ParseQuery(u.RawQuery)
					// only do replacment if operId=8
					if nil == e && nil != values["bidid"] && nil != values["operId"] && values["operId"][0] == "8" {
						values.Set("bidid", bid.ID)
					} else {
						continue
					}

					//OTT-183: Fix
					if values["operId"] != nil && values["operId"][0] == "8" {
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
func getAdPodBidExtension(adpod *AdPodBid) json.RawMessage {
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
		bid.Status = StatusWinningBid
	}
	rawExt, _ := json.Marshal(bidExt)
	return rawExt
}
