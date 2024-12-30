package adpod

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
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

const (
	inLineEnd       = "</InLine>"
	wrapperEnd      = "</Wrapper>"
	extensionsStart = "<Extensions>"
	extensionsEnd   = "</Extensions>"
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
	ErrorFormat        = `{"` + ERROR_CODE + `":%v,"` + ERROR_STRING + `":"%s"}`
	NBRFormatWithError = `{"` + NBR + `":%v,"` + ERROR + `":%s}`
	NBRFormatQuote     = `{"` + NBR + `":%v,"` + ERROR + `":"%v"}`
	NBRFormat          = `{"` + NBR + `":%v}`
)

type vastResp struct {
	rctx models.RequestCtx
}

func newVastResponder(rctx models.RequestCtx) Responder {
	return &vastResp{rctx: rctx}
}

func (vr *vastResp) FormResponse(bidResponse *openrtb2.BidResponse, headers http.Header) (interface{}, http.Header, error) {
	if bidResponse.NBR != nil {
		return EmptyVASTResponse, headers, nil
	}

	vast, nbr, err := vr.getVast(bidResponse)
	if err != nil {
		return nil, headers, err
	}
	if nbr != nil {
		return EmptyVASTResponse, headers, nil
	}

	return vast, headers, nil

}

// func (vr *vastResp) addOwStatusHeader(headers map[string]string, nbr openrtb3.NoBidReason) {
// 	if vr.rctx.Debug {
// 		headers[HeaderOpenWrapStatus] = fmt.Sprintf(NBRFormat, nbr)
// 	}
// }

func (vr *vastResp) getVast(bidResponse *openrtb2.BidResponse) (string, *openrtb3.NoBidReason, error) {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return "", nbr.EmptySeatBid.Ptr(), errors.New("empty bid response")
	}

	var reqExt openrtb_ext.ExtBidResponse
	err := json.Unmarshal(bidResponse.Ext, &reqExt)
	if err != nil {
		nbr := openrtb3.NoBidReason(openrtb3.NoBidGeneralError)
		return "", &nbr, errors.New("No Bid")
	}

	isAdpodResponse := false
	if reqExt.Wrapper != nil && reqExt.Wrapper.IsPodRequest {
		isAdpodResponse = true
	}
	reqExt.Wrapper = nil

	bidArray := make([]openrtb2.Bid, 0)
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			if bid.Price > 0 {
				bidArray = append(bidArray, bid)
			}
		}
	}

	creative, _ := getAdPodBidCreativeAndPriceForVast(bidArray, isAdpodResponse)
	if len(creative) == 0 {
		nbr := openrtb3.NoBidReason(openrtb3.NoBidGeneralError)
		return "", &nbr, errors.New("No Bid")
	}

	if vr.rctx.Debug {
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

func addExtInfo(vastBytes []byte, responseExt json.RawMessage) []byte {

	adm := string(vastBytes)
	owExt := "<Extension type=" + `"OpenWrap"` + "><Ext><![CDATA[" + string(responseExt) + "]]></Ext></Extension>"

	// Check if Extensions Exists
	ci := strings.Index(adm, extensionsEnd)
	if ci != -1 {
		adm = strings.Replace(adm, extensionsEnd, owExt+extensionsEnd, 1)
		return []byte(adm)
	}

	// Check if Wrapper Exists
	wi := strings.Index(adm, wrapperEnd)
	if wi != -1 {
		adm = strings.Replace(adm, wrapperEnd, extensionsStart+owExt+extensionsEnd+wrapperEnd, 1)
		return []byte(adm)

	}

	// Check if Inline Exists
	wi = strings.Index(adm, inLineEnd)
	if wi != -1 {
		adm = strings.Replace(adm, inLineEnd, extensionsStart+owExt+extensionsEnd+inLineEnd, 1)
		return []byte(adm)
	}
	return vastBytes
}

func getAdPodBidCreativeAndPriceForVast(bids []openrtb2.Bid, isAdpodRequest bool) (string, float64) {
	if len(bids) == 0 {
		return "", 0
	}
	if !isAdpodRequest {
		return bids[0].AdM, bids[0].Price
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
