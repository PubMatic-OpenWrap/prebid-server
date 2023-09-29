package vastbidder

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/xmlparser"
)

var vastXPath *xmlparser.XPath

func init() {
	//vastXPath = xmlparser.GetXPath(nil)
}

func (handler *VASTTagResponseHandler) getBidResponse(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	var errs []error

	reader := xmlparser.NewXMLReader(vastXPath)

	//Read Document
	if err := reader.Parse(response.Body); err != nil {
		errs = append(errs, err)
		return nil, errs[:]
	}

	//Check VAST Tag
	vast := reader.FindElement(nil, "VAST")
	if vast == nil {
		errs = append(errs, errors.New("VAST Tag Not Found"))
		return nil, errs[:]
	}

	//Check VAST/Ad Tag
	adElement := func() *xmlparser.Element {
		ad := reader.FindElement(vast, "Ad")
		if ad == nil {
			return nil
		}
		element := reader.FindElement(ad, "Wrapper")
		if element == nil {
			element = reader.FindElement(ad, "InLine")
		}
		return element
	}()
	if adElement == nil {
		errs = append(errs, errors.New("VAST/Ad Tag Not Found"))
		return nil, errs[:]
	}

	typedBid := &adapters.TypedBid{
		Bid:     &openrtb2.Bid{},
		BidType: openrtb_ext.BidTypeVideo,
		BidVideo: &openrtb_ext.ExtBidPrebidVideo{
			VASTTagID: handler.VASTTag.TagID,
		},
	}

	creatives := reader.FindElements(adElement, "Creatives", "Creative")
	if nil != creatives {
		for _, creative := range creatives {
			// get creative id
			typedBid.Bid.CrID = reader.GetAttribute(creative, "id")

			// get duration from vast creative
			dur, err := getDuration1(reader, creative)
			if nil != err {
				// get duration from input bidder vast tag
				dur = getStaticDuration(handler.VASTTag)
			}
			if dur > 0 {
				typedBid.BidVideo.Duration = int(dur) // prebid expects int value
			}
		}
	}

	bidResponse := &adapters.BidderResponse{
		Bids:     []*adapters.TypedBid{typedBid},
		Currency: `USD`, //TODO: Need to check how to get currency value
	}

	//GetVersion
	version := reader.GetAttribute(vast, "version")
	if version == "" {
		version = "2.0"
	}

	//if bid.price is not set in ParseExtension
	if typedBid.Bid.Price <= 0 {
		price, currency := getPricingDetails1(reader, adElement, version)
		if price <= 0 {
			price, currency = getStaticPricingDetails(handler.VASTTag)
			if price <= 0 {
				errs = append(errs, &errortypes.NoBidPrice{Message: "Bid Price Not Present"})
				return nil, errs[:]
			}
		}
		typedBid.Bid.Price = price
		if len(currency) > 0 {
			bidResponse.Currency = currency
		}
	}

	typedBid.Bid.ADomain = getAdvertisers1(reader, adElement, version)

	//if bid.id is not set in ParseExtension
	if len(typedBid.Bid.ID) == 0 {
		typedBid.Bid.ID = GetRandomID()
	}

	//if bid.impid is not set in ParseExtension
	if len(typedBid.Bid.ImpID) == 0 {
		typedBid.Bid.ImpID = internalRequest.Imp[externalRequest.Params.ImpIndex].ID
	}

	//if bid.adm is not set in ParseExtension
	if len(typedBid.Bid.AdM) == 0 {
		typedBid.Bid.AdM = string(response.Body)
	}

	//if bid.CrID is not set in ParseExtension
	if len(typedBid.Bid.CrID) == 0 {
		typedBid.Bid.CrID = "cr_" + GetRandomID()
	}

	// set vastTagId in bid.Ext
	bidExt := openrtb_ext.ExtBid{
		Prebid: &openrtb_ext.ExtBidPrebid{
			Video: typedBid.BidVideo,
			Type:  typedBid.BidType,
		},
	}

	bidExtBytes, err := json.Marshal(bidExt)
	if err == nil {
		typedBid.Bid.Ext = bidExtBytes
	}

	return bidResponse, nil
}

func getDuration1(reader *xmlparser.XMLReader, creative *xmlparser.Element) (int, error) {
	if nil == creative {
		return 0, errors.New("Invalid Creative")
	}

	node := reader.FindElement(creative, "Linear", "Duration")
	if nil == node {
		return 0, errors.New("Invalid Duration")
	}

	duration := strings.TrimSpace(reader.GetText(node, true))

	// check if milliseconds is provided
	match := durationRegExp.FindStringSubmatch(duration)
	if nil == match {
		return 0, errors.New("Invalid Duration")
	}
	repl := "${1}h${2}m${3}s"
	ms := match[5]
	if "" != ms {
		repl += "${5}ms"
	}
	duration = durationRegExp.ReplaceAllString(duration, repl)
	dur, err := time.ParseDuration(duration)
	if err != nil {
		return 0, err
	}
	return int(dur.Seconds()), nil
}

func getAdvertisers1(reader *xmlparser.XMLReader, ad *xmlparser.Element, vastVer string) []string {
	version, err := strconv.ParseFloat(vastVer, 64)
	if err != nil {
		version = 2.0
	}

	advertisers := make([]string, 0)

	switch int(version) {
	case 2, 3:
		for _, ext := range reader.FindElements(ad, "Extensions", "Extension") {
			if reader.GetAttribute(ext, "type") == "advertiser" {
				ele := reader.FindElement(ext, "Advertiser")
				if ele != nil {
					if value := strings.TrimSpace(reader.GetText(ele, true)); len(value) > 0 {
						advertisers = append(advertisers, value)
					}
				}
			}
		}
	case 4:
		if ele := reader.FindElement(ad, "Advertiser"); ele != nil {
			if value := strings.TrimSpace(reader.GetText(ele, true)); len(value) > 0 {
				advertisers = append(advertisers, value)
			}
		}
	default:
		glog.V(3).Infof("Handle getAdvertisers for VAST version %d", int(version))
	}

	if len(advertisers) == 0 {
		return nil
	}
	return advertisers
}

func getPricingDetails1(reader *xmlparser.XMLReader, ad *xmlparser.Element, version string) (float64, string) {
	var currency string
	var node *xmlparser.Element

	if version == `2.0` {
		node = reader.FindElement(ad, "Extensions", "Extension", "Price")
	} else {
		node = reader.FindElement(ad, "Pricing")
	}

	if node == nil {
		return 0.0, currency
	}

	priceValue, err := strconv.ParseFloat(strings.TrimSpace(reader.GetText(node, true)), 64)
	if nil != err {
		return 0.0, currency
	}

	if value := reader.GetAttribute(node, "currency"); len(value) > 0 {
		currency = value
	}

	return priceValue, currency
}
