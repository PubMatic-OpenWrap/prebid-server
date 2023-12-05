package middleware

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/endpoints/events"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/utils"
	"github.com/rs/vast"
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
	NBR                  = "nbr"
	ERROR                = "error"
	//ErrorFormat parsing error format
	ErrorFormat    = `{"` + ERROR_CODE + `":%v,"` + ERROR_STRING + `":"%s"}`
	NBRFormat      = `{"` + NBR + `":%v,"` + ERROR + `":%s}`
	NBRFormatQuote = `{"` + NBR + `":%v,"` + ERROR + `":"%v"}`
)

type vastResponse struct {
	debug              string
	WrapperLoggerDebug string
}

func (vr *vastResponse) formVastResponse(aw *utils.CustomWriter) ([]byte, map[string]string, int) {
	var statusCode = http.StatusOK
	var headers = map[string]string{
		ContentType:    ApplicationXML,
		ContentOptions: NoSniff,
	}

	response, err := io.ReadAll(aw.Response)
	if err != nil {
		statusCode = http.StatusInternalServerError
		headers[HeaderOpenWrapStatus] = fmt.Sprintf(NBRFormat, nbr.InternalError, err.Error())
		return EmptyVASTResponse, headers, statusCode
	}

	var bidResponse *openrtb2.BidResponse
	err = json.Unmarshal(response, &bidResponse)
	if err != nil {
		statusCode = http.StatusInternalServerError
		headers[HeaderOpenWrapStatus] = fmt.Sprintf(NBRFormat, nbr.InternalError, err.Error())
		return EmptyVASTResponse, headers, statusCode
	}

	if bidResponse.NBR != nil {
		statusCode = http.StatusBadRequest
		data, _, _, _ := jsonparser.Get(bidResponse.Ext, errorLocation...)
		headers[HeaderOpenWrapStatus] = fmt.Sprintf(NBRFormat, *bidResponse.NBR, strconv.Quote(string(data)))
		return EmptyVASTResponse, headers, statusCode
	}

	vast, nbr, err := vr.getVast(bidResponse)
	if nbr != nil {
		headers[HeaderOpenWrapStatus] = fmt.Sprintf(NBRFormat, *nbr, err.Error())
		return EmptyVASTResponse, headers, statusCode
	}

	return []byte(vast), headers, statusCode
}

func (vr *vastResponse) getVast(bidResponse *openrtb2.BidResponse) (string, *openrtb3.NoBidReason, error) {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return "", GetNoBidReasonCode(nbr.EmptySeatBid), errors.New("empty bid response")
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
		return "", GetNoBidReasonCode(nbr.InternalError), errors.New("empty creative")
	}

	if vr.debug == "1" || vr.WrapperLoggerDebug == "1" {
		creative = string(addExtInfo([]byte(creative), bidResponse.Ext))
	}

	return creative, nil, nil
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
		glog.Error("Error while creating vast:", err)
		return "", price
	}
	return bidAdM, price
}

func adjustBidIDInVideoEventTrackers(doc *etree.Document, bid *openrtb2.Bid) {
	// adjusment: update bid.id with ctv module generated bid.id
	creatives := events.FindCreatives(doc)
	for _, creative := range creatives {
		trackingEvents := creative.FindElements("TrackingEvents/Tracking")
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

func addExtInfo(vastBytes []byte, responseExt json.RawMessage) []byte {
	var v vast.VAST
	if err := xml.Unmarshal(vastBytes, &v); err != nil {
		return vastBytes
	}

	if len(v.Ads) == 0 {
		return vastBytes
	}

	owExtBytes := append([]byte("<Ext>"), append(responseExt, []byte("</Ext>")...)...)

	owExt := vast.Extension{
		Type: "OpenWrap",
		Data: owExtBytes,
	}

	ad := v.Ads[0]
	if ad.InLine != nil {
		if ad.InLine.Extensions == nil {
			ad.InLine.Extensions = &([]vast.Extension{})
		}
		*ad.InLine.Extensions = append(*ad.InLine.Extensions, owExt)
	} else if ad.Wrapper != nil {
		if ad.Wrapper.Extensions == nil {
			ad.Wrapper.Extensions = []vast.Extension{}
		}
		ad.Wrapper.Extensions = append(ad.Wrapper.Extensions, owExt)
	}

	newVASTBytes, err := xml.Marshal(v)
	if err != nil {
		return vastBytes
	}

	return newVASTBytes
}
