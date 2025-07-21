package vastbidder

import (
	"strconv"

	"github.com/beevik/etree"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type etreeXMLParser struct {
	vastTag         *openrtb_ext.ExtImpVASTBidderTag
	reader          *etree.Document
	vastVersion     float64
	crID            string
	adElement       *etree.Element
	creativeElement *etree.Element
}

func newETreeXMLParser() *etreeXMLParser {
	return &etreeXMLParser{}
}

func (p *etreeXMLParser) Name() string {
	return openrtb_ext.XMLParserETree
}

func (p *etreeXMLParser) SetVASTTag(vastTag *openrtb_ext.ExtImpVASTBidderTag) {
	p.vastTag = vastTag
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

func (p *etreeXMLParser) GetPricingDetails() (price float64, currency string) {
	node := p.getPricingNode()

	if node == nil {
		return 0.0, ""
	}

	priceValue, err := strconv.ParseFloat(node.TrimmedText(), 64)
	if nil != err {
		return 0.0, ""
	}

	if currency = node.SelectAttrValue("currency", "USD"); currency == "" {
		currency = "USD"
	}

	return priceValue, currency
}

func (p *etreeXMLParser) getPricingNode() *etree.Element {
	node := p.adElement.SelectElement("Pricing")
	if node == nil {
		node = p.adElement.FindElement("./Extensions/Extension/Pricing")
	}
	if node == nil {
		node = p.adElement.FindElement("./Extensions/Extension/Price")
	}

	return node
}

func (p *etreeXMLParser) GetAdvertiser() (advertisers []string) {
	switch int(p.vastVersion) {
	case vastVersion2x, vastVersion3x:
		for _, ext := range p.adElement.FindElements("./Extensions/Extension/") {
			if ext.SelectAttrValue("type", "") == "advertiser" {
				ele := ext.SelectElement("Advertiser")
				if ele != nil {
					if value := ele.TrimmedText(); len(value) > 0 {
						advertisers = append(advertisers, value)
					}
				}
			}
		}

	case vastVersion4x:
		if ele := p.adElement.SelectElement("Advertiser"); ele != nil {
			if value := ele.TrimmedText(); len(value) > 0 {
				advertisers = append(advertisers, value)
			}
		}
	}

	return
}

func (p *etreeXMLParser) GetCreativeID() string {
	if p.crID == "" && p.creativeElement != nil {
		p.crID = p.creativeElement.SelectAttrValue("id", "")
	}

	if p.crID == "" {
		p.crID = "cr_" + generateRandomID()
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
	return parseDuration(node.TrimmedText())
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
