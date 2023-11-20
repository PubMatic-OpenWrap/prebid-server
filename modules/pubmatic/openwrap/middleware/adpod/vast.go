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
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/endpoints/events"
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
	VASTVersionsStr   = []string{"0", "1.0", "2.0", "3.0", "4.0"}
	EmptyVASTResponse = []byte(`<VAST version="2.0"/>`)
	//HeaderOpenWrapStatus Status of OW Request
	HeaderOpenWrapStatus = "X-Ow-Status"
	ERROR_CODE           = "ErrorCode"
	ERROR_STRING         = "Error"
	//ErrorFormat parsing error format
	ErrorFormat = `{"` + ERROR_CODE + `":%v,"` + ERROR_STRING + `":"%s"}`
)

func formVastResponse(response []byte) ([]byte, error) {
	var bidResponse *openrtb2.BidResponse

	err := json.Unmarshal(response, &bidResponse)
	if err != nil {
		return EmptyVASTResponse, errors.New("Failed to unmarshal the bid response")
	}

	vast, err := getVast(bidResponse)
	if err != nil {
		return EmptyVASTResponse, err
	}

	return []byte(vast), nil
}

func getVast(bidResponse *openrtb2.BidResponse) (string, error) {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return "", errors.New("recieved invalid bidResponse")
	}

	bidArray := make([]openrtb2.Bid, 0)
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			if bid.Price > 0 {
				bidArray = append(bidArray, bid)
			}
		}
	}

	creative, _ := getAdPodBidCreativeAndPrice(bidArray)
	if len(creative) == 0 {
		return "", errors.New("error while creating creative")
	}

	return creative, nil
}

// getAdPodBidCreative get commulative adpod bid details
func getAdPodBidCreativeAndPrice(bids []openrtb2.Bid) (string, float64) {
	if len(bids) == 0 {
		return "", 0
	}

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
