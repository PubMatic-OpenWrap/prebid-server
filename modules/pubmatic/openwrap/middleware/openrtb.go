package middleware

import (
	"encoding/json"
	"errors"
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

func FormOperRTBResponse(response []byte) []byte {
	bidResponse := openrtb2.BidResponse{}

	err := json.Unmarshal(response, &bidResponse)
	if err != nil {
		return response
	}

	mergedBidResponse, err := mergeSeatBids(&bidResponse)
	if err != nil {
		return response
	}

	data, err := json.Marshal(mergedBidResponse)
	if err != nil {
		return response
	}

	return data
}

func mergeSeatBids(bidResponse *openrtb2.BidResponse) (*openrtb2.BidResponse, error) {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return nil, errors.New("recieved invalid bidResponse")
	}

	bidArrayMap := make(map[string][]*openrtb2.Bid)
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			impId, _ := models.GetImpressionID(bid.ImpID)
			bids, ok := bidArrayMap[impId]
			if !ok {
				bids = make([]*openrtb2.Bid, 0)
			}
			bids = append(bids, &bid)
			bidArrayMap[impId] = bids
		}
	}

	bidResponse.SeatBid = getPrebidCTVSeatBid(bidArrayMap)

	return bidResponse, nil
}

func getPrebidCTVSeatBid(bidsMap map[string][]*openrtb2.Bid) []openrtb2.SeatBid {
	seatBids := []openrtb2.SeatBid{}

	for impId, bids := range bidsMap {
		bid := openrtb2.Bid{}
		bidID, err := uuid.NewV4()
		if err == nil {
			bid.ID = bidID.String()
		} else {
			bid.ID = bids[0].ID
		}
		creative, price := getAdPodBidCreativeAndPrice(bids, true)
		bid.AdM = creative
		bid.Price = price
		bid.Cat = bids[0].Cat
		bid.ADomain = bids[0].ADomain
		bid.ImpID = impId

		seatBid := openrtb2.SeatBid{}
		seatBid.Seat = "prebid_ctv"
		seatBid.Bid = append(seatBid.Bid, bid)

		seatBids = append(seatBids, seatBid)
	}

	return seatBids
}

// getAdPodBidCreative get commulative adpod bid details
func getAdPodBidCreativeAndPrice(bids []*openrtb2.Bid, generatedBidID bool) (string, float64) {
	var price float64
	doc := etree.NewDocument()
	vast := doc.CreateElement(VASTElement)
	sequenceNumber := 1
	var version float64 = 2.0

	for _, bid := range bids {
		price = price + bid.Price
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
				adjustBidIDInVideoEventTrackers(adDoc, bid)
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
		return "", price
	}
	return bidAdM, price
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
