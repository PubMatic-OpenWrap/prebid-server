package vastbidder

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/beevik/etree"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type xmlParser interface {
	Parse([]byte) error
	GetAdvertiser() []string
	GetPricingDetails() (float64, string)
	GetCreativeID() string
	GetDuration() (int, error)
}

type fastXMLParser struct {
	vastTag         *openrtb_ext.ExtImpVASTBidderTag
	reader          *fastxml.XMLReader
	vastVersion     float64
	crID            string
	adElement       *fastxml.Element
	creativeElement *fastxml.Element
}

func (p *fastXMLParser) Parse(vastXML []byte) (err error) {
	p.reader = fastxml.NewXMLReader(nil)

	//parse vast xml
	if err := p.reader.Parse(vastXML); err != nil {
		return err
	}

	//validate VAST tag
	vast := p.reader.FindElement(nil, "VAST")
	if vast == nil {
		return errMissingVASTElement
	}

	//validate vast version
	versionStr := p.reader.GetAttributeValue(vast, "version")
	p.vastVersion, err = parseVASTVersion(versionStr)
	if err != nil {
		return err
	}

	//validate VAST/Ad tag
	p.adElement = p.getAdElement(vast)
	if p.adElement == nil {
		return errMissingAdElement
	}

	//creative is not mandatory
	p.creativeElement = p.reader.FindElement(p.adElement, "Creatives", "Creative")
	return
}

func (p *fastXMLParser) getAdElement(vast *fastxml.Element) *fastxml.Element {
	element := p.reader.FindElement(vast, "Ad")
	if element == nil {
		return nil
	}
	adElement := p.reader.FindElement(element, "Wrapper")
	if adElement == nil {
		adElement = p.reader.FindElement(element, "InLine")
	}
	return adElement
}

func (p *fastXMLParser) GetPricingDetails() (price float64, currency string) {
	var node *fastxml.Element

	if int(p.vastVersion) == 2 {
		node = p.reader.FindElement(p.adElement, "Extensions", "Extension", "Price")
	} else {
		node = p.reader.FindElement(p.adElement, "Pricing")
	}

	if node == nil {
		return 0.0, ""
	}

	priceValue, err := strconv.ParseFloat(strings.TrimSpace(p.reader.GetText(node, true)), 64)
	if nil != err {
		return 0.0, ""
	}

	if currency = p.reader.GetAttributeValue(node, "currency"); currency == "" {
		currency = "USD"
	}

	return priceValue, currency
}

func (p *fastXMLParser) GetAdvertiser() (advertisers []string) {
	switch int(p.vastVersion) {
	case vastVersion2x, vastVersion3x:
		for _, ext := range p.reader.FindElements(p.adElement, "Extensions", "Extension") {
			if p.reader.GetAttributeValue(ext, "type") == "advertiser" {
				ele := p.reader.FindElement(ext, "Advertiser")
				if ele != nil {
					if value := strings.TrimSpace(p.reader.GetText(ele, true)); len(value) > 0 {
						advertisers = append(advertisers, value)
					}
				}
			}
		}

	case vastVersion4x:
		if ele := p.reader.FindElement(p.adElement, "Advertiser"); ele != nil {
			if value := strings.TrimSpace(p.reader.GetText(ele, true)); len(value) > 0 {
				advertisers = append(advertisers, value)
			}
		}

	default:
		glog.V(3).Infof("Handle getAdvertisers for VAST version %v", p.vastVersion)
	}

	return
}

func (p *fastXMLParser) GetCreativeID() string {
	if p.crID == "" && p.creativeElement != nil {
		p.crID = p.reader.GetAttributeValue(p.creativeElement, "id")
	}

	if p.crID == "" {
		p.crID = "cr_" + GetRandomID()
	}

	return p.crID
}

func (p *fastXMLParser) GetDuration() (int, error) {
	if p.creativeElement == nil {
		return 0, errEmptyVideoCreative
	}
	node := p.reader.FindElement(p.creativeElement, "Linear", "Duration")
	if node == nil {
		return 0, errEmptyVideoDuration
	}
	return getCreativeDuration(strings.TrimSpace(p.reader.GetText(node, true)))
}

type responseGenerator struct {
	internalRequest *openrtb2.BidRequest
	externalRequest *adapters.RequestData
	response        *adapters.ResponseData
	vastTag         *openrtb_ext.ExtImpVASTBidderTag
	parser          xmlParser
}

func (rg *responseGenerator) GetBidderResponse() (*adapters.BidderResponse, error) {
	//get price and currency details
	price, currency := rg.parser.GetPricingDetails()
	if price <= 0 {
		price, currency = getStaticPricingDetails(rg.vastTag)
		if price <= 0 {
			return nil, errMissingBidPrice
		}
	}
	if currency == "" {
		currency = "USD"
	}

	//duration prebid expects int value
	dur, err := rg.parser.GetDuration()
	if nil != err {
		//get duration from input bidder vast tag
		dur = getStaticDuration(rg.vastTag)
	}

	//creating openrtb formatted bid object
	bid := &openrtb2.Bid{
		ID:      GetRandomID(),
		ImpID:   rg.internalRequest.Imp[rg.externalRequest.Params.ImpIndex].ID,
		AdM:     string(rg.response.Body),
		Price:   price,
		CrID:    rg.parser.GetCreativeID(),
		ADomain: rg.parser.GetAdvertiser(),
	}

	// bid.ext settting vasttagid and bid type
	bidExt := openrtb_ext.ExtBid{
		Prebid: &openrtb_ext.ExtBidPrebid{
			Video: &openrtb_ext.ExtBidPrebidVideo{
				VASTTagID: rg.vastTag.TagID,
				Duration:  dur,
			},
			Type: openrtb_ext.BidTypeVideo,
		},
	}
	bid.Ext, _ = json.Marshal(bidExt)

	//bidderresponse generation
	bidResponse := &adapters.BidderResponse{
		Bids: []*adapters.TypedBid{
			&adapters.TypedBid{
				Bid:      bid,
				BidType:  bidExt.Prebid.Type,
				BidVideo: bidExt.Prebid.Video,
			},
		},
		Currency: currency,
	}
	return bidResponse, nil
}

type etreeXMLParser struct {
	vastTag         *openrtb_ext.ExtImpVASTBidderTag
	reader          *etree.Document
	vastVersion     float64
	crID            string
	adElement       *etree.Element
	creativeElement *etree.Element
}

func (p *etreeXMLParser) Parse(vastXML []byte) (err error) {
	p.reader = etree.NewDocument()

	//parse vast xml
	if err := p.reader.ReadFromBytes(vastXML); err != nil {
		return err
	}

	//validate VAST tag
	vast := p.reader.Element.FindElement("./VAST")
	if vast == nil {
		return errMissingVASTElement
	}

	//validate vast version
	versionStr := vast.SelectAttrValue("version", "2.0")
	p.vastVersion, err = parseVASTVersion(versionStr)
	if err != nil {
		return err
	}

	//validate VAST/Ad tag
	p.adElement = p.getAdElement(vast)
	if p.adElement == nil {
		return errMissingAdElement
	}

	//creative is not mandatory
	p.creativeElement = p.adElement.FindElement("./Creatives/Creative")
	return
}

func (p *etreeXMLParser) getAdElement(vast *etree.Element) *etree.Element {
	element := vast.SelectElement("Ad")
	if element == nil {
		return nil
	}
	adElement := element.SelectElement("Wrapper")
	if adElement == nil {
		adElement = element.SelectElement("InLine")
	}
	return adElement
}

func (p *etreeXMLParser) GetPricingDetails() (price float64, currency string) {
	var node *etree.Element

	if int(p.vastVersion) == 2 {
		node = p.adElement.FindElement("./Extensions/Extension/Price")
	} else {
		node = p.adElement.SelectElement("Pricing")
	}

	if node == nil {
		return 0.0, ""
	}

	priceValue, err := strconv.ParseFloat(strings.TrimSpace(node.Text()), 64)
	if nil != err {
		return 0.0, ""
	}

	if value := node.SelectAttrValue("currency", "USD"); len(value) > 0 {
		currency = value
	}

	return priceValue, currency
}

func (p *etreeXMLParser) GetAdvertiser() (advertisers []string) {
	switch int(p.vastVersion) {
	case vastVersion2x, vastVersion3x:
		for _, ext := range p.adElement.FindElements("./Extensions/Extension/") {
			if ext.SelectAttrValue("type", "") == "advertiser" {
				ele := ext.SelectElement("Advertiser")
				if ele != nil {
					if value := strings.TrimSpace(ele.Text()); len(value) > 0 {
						advertisers = append(advertisers, value)
					}
				}
			}
		}

	case vastVersion4x:
		if ele := p.adElement.SelectElement("Advertiser"); ele != nil {
			if value := strings.TrimSpace(ele.Text()); len(value) > 0 {
				advertisers = append(advertisers, value)
			}
		}

	default:
		glog.V(3).Infof("Handle getAdvertisers for VAST version %v", p.vastVersion)
	}

	return
}

func (p *etreeXMLParser) GetCreativeID() string {
	if p.crID == "" && p.creativeElement != nil {
		p.crID = p.creativeElement.SelectAttrValue("id", "")
	}

	if p.crID == "" {
		p.crID = "cr_" + GetRandomID()
	}

	return p.crID
}

func (p *etreeXMLParser) GetDuration() (int, error) {
	if p.creativeElement == nil {
		return 0, errEmptyVideoCreative
	}
	node := p.creativeElement.FindElement("./Linear/Duration")
	if node == nil {
		return 0, errEmptyVideoDuration
	}
	return getCreativeDuration(strings.TrimSpace(node.Text()))
}

/*
type fastXMLResponseGenerator struct {
	vastTag     *openrtb_ext.ExtImpVASTBidderTag
	reader      *fastxml.XMLReader
	vastVersion float64
}

func NewFastXMLResponseGenerator(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) *fastXMLResponseGenerator {
	f := &fastXMLResponseGenerator{}
	return f
}

func (f *fastXMLResponseGenerator) GetBidderResponse() (*adapters.BidderResponse, []error) {
	typedBid := &adapters.TypedBid{
		Bid:     &openrtb2.Bid{},
		BidType: openrtb_ext.BidTypeVideo,
		BidVideo: &openrtb_ext.ExtBidPrebidVideo{
			VASTTagID: handler.VASTTag.TagID,
		},
	}

	bidResponse := &adapters.BidderResponse{
		Bids:     []*adapters.TypedBid{typedBid},
		Currency: `USD`, //TODO: Need to check how to get currency value
	}

	return bidResponse, nil
}

func (f *fastXMLResponseGenerator) ParseVASTXML(vastXML []byte) (*openrtb2.Bid, error) {
	reader := fastxml.NewXMLReader(nil)

	//Read Document
	if err := reader.Parse(vastXML); err != nil {
		return nil, err
	}

	//vast xml validation
	//Check VAST Tag
	vast := reader.FindElement(nil, "VAST")
	if vast == nil {
		return nil, errMissingVASTElement
	}

	//Check VAST/Ad Tag
	adElement := f.getAdElement(vast)
	if adElement == nil {
		return nil, errMissingAdElement
	}

	//read vast version
	version := reader.GetAttributeValue(vast, "version")
	if version == "" {
		f.vastVersion = 2.0
	}

	//if bid.price is not set in ParseExtension
	price, currency := f.GetPricingDetails(adElement)
	if price <= 0 {
		price, currency = getStaticPricingDetails(f.vastTag)
		if price <= 0 {
			return nil, errMissingBidPrice
		}
	}
	if len(currency) > 0 {
		bidResponse.Currency = currency
	}

	bid := &openrtb2.Bid{
		ID:      GetRandomID(),
		ImpID:   internalRequest.Imp[externalRequest.Params.ImpIndex].ID,
		AdM:     string(vastXML),
		Price:   price,
		ADomain: f.GetAdvertiser(adElement),
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
		bid.Ext = bidExtBytes
	}

	// typedBid := &adapters.TypedBid{
	// 	Bid:     &openrtb2.Bid{},
	// 	BidType: openrtb_ext.BidTypeVideo,
	// 	BidVideo: &openrtb_ext.ExtBidPrebidVideo{
	// 		VASTTagID: handler.VASTTag.TagID,
	// 	},
	// }

	creatives := reader.FindElements(adElement, "Creatives", "Creative")
	if nil != creatives {
		for _, creative := range creatives {
			// get creative id
			bid.CrID = reader.GetAttributeValue(creative, "id")

			// get duration from vast creative
			dur, err := f.GetDuration(creative)
			if nil != err {
				//TODO: if not present then set duration from outside
				//get duration from input bidder vast tag
				dur = getStaticDuration(f.vastTag)
			}
			if dur > 0 {
				typedBid.BidVideo.Duration = int(dur) // prebid expects int value
			}
		}
	}
	//if bid.CrID is not set in ParseExtension
	if len(bid.CrID) == 0 {
		bid.CrID = "cr_" + GetRandomID()
	}

	// bidResponse := &adapters.BidderResponse{
	// 	Bids:     []*adapters.TypedBid{typedBid},
	// 	Currency: `USD`, //TODO: Need to check how to get currency value
	// }

	return "", nil
}

func (f *fastXMLResponseGenerator) GetDuration(creative *fastxml.Element) (int, error) {
	if creative == nil{
		return 0, errEmptyVideoCreative
	}

	node := f.reader.FindElement(creative, "Linear", "Duration")
	if node == nil{
		return 0, errEmptyVideoDuration
	}

	return getCreativeDuration(strings.TrimSpace(f.reader.GetText(node, true)))
}

func (f *fastXMLResponseGenerator) GetAdvertiser(ad *fastxml.Element) (advertisers []string) {
	switch int(f.vastVersion) {
	case vastVersion2x, vastVersion3x:
		for _, ext := range f.reader.FindElements(ad, "Extensions", "Extension") {
			if f.reader.GetAttributeValue(ext, "type") == "advertiser" {
				ele := f.reader.FindElement(ext, "Advertiser")
				if ele != nil {
					if value := strings.TrimSpace(f.reader.GetText(ele, true)); len(value) > 0 {
						advertisers = append(advertisers, value)
					}
				}
			}
		}

	case vastVersion4x:
		if ele := f.reader.FindElement(ad, "Advertiser"); ele != nil {
			if value := strings.TrimSpace(f.reader.GetText(ele, true)); len(value) > 0 {
				advertisers = append(advertisers, value)
			}
		}

	default:
		glog.V(3).Infof("Handle getAdvertisers for VAST version %v", f.vastVersion)
	}

	return
}

func (f *fastXMLResponseGenerator) GetPricingDetails(ad *fastxml.Element) (float64, string) {
	var currency string
	var node *fastxml.Element

	if int(f.vastVersion) == 2 {
		node = f.reader.FindElement(ad, "Extensions", "Extension", "Price")
	} else {
		node = f.reader.FindElement(ad, "Pricing")
	}

	if node == nil {
		return 0.0, ""
	}

	priceValue, err := strconv.ParseFloat(strings.TrimSpace(f.reader.GetText(node, true)), 64)
	if nil != err {
		return 0.0, ""
	}

	if value := f.reader.GetAttributeValue(node, "currency"); len(value) > 0 {
		currency = value
	}

	return priceValue, currency
}

func (f *fastXMLResponseGenerator) getAdElement(vast *fastxml.Element) *fastxml.Element {
	element := f.reader.FindElement(vast, "Ad")
	if element == nil {
		return nil
	}
	adElement := f.reader.FindElement(element, "Wrapper")
	if adElement == nil {
		adElement = f.reader.FindElement(element, "InLine")
	}
	return adElement
}

/*
fastXMLParser
etreeXMLParser

Parse() (error)
GetDuration() (int, error)
GetAdvertiser() ([]string)
GetPricingDetails() (float64, string)
*/
