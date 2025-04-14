package vastbidder

import (
	"strconv"
	"strings"

	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type fastXMLParser struct {
	vastTag         *openrtb_ext.ExtImpVASTBidderTag
	reader          *fastxml.XMLReader
	vastVersion     float64
	crID            string
	adElement       *fastxml.Element
	creativeElement *fastxml.Element
}

func newFastXMLParser() *fastXMLParser {
	return &fastXMLParser{}
}

func (p *fastXMLParser) Name() string {
	return openrtb_ext.XMLParserFastXML
}

func (p *fastXMLParser) SetVASTTag(vastTag *openrtb_ext.ExtImpVASTBidderTag) {
	p.vastTag = vastTag
}

func (p *fastXMLParser) Parse(vastXML []byte) (err error) {
	p.reader = fastxml.NewXMLReader()

	//parse vast xml
	if err := p.reader.Parse(vastXML); err != nil {
		return err
	}

	//validate VAST tag
	vast := p.reader.SelectElement(nil, "VAST")
	if vast == nil {
		return errMissingVASTElement
	}

	//validate vast version
	versionStr := p.reader.SelectAttrValue(vast, "version", "2.0")
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
	p.creativeElement = p.reader.SelectElement(p.adElement, "Creatives", "Creative")
	return
}

func (p *fastXMLParser) GetPricingDetails() (price float64, currency string) {
	node := p.getPricingNode()

	if node == nil {
		return 0.0, ""
	}

	priceValue, err := strconv.ParseFloat(strings.TrimSpace(p.reader.RawText(node)), 64)
	if nil != err {
		return 0.0, ""
	}

	if currency = p.reader.SelectAttrValue(node, "currency", "USD"); currency == "" {
		currency = "USD"
	}

	return priceValue, currency
}

func (p *fastXMLParser) getPricingNode() *fastxml.Element {
	node := p.reader.SelectElement(p.adElement, "Pricing")
	if node == nil {
		node = p.reader.SelectElement(p.adElement, "Extensions", "Extension", "Pricing")
	}
	if node == nil {
		node = p.reader.SelectElement(p.adElement, "Extensions", "Extension", "Price")
	}

	return node
}

func (p *fastXMLParser) GetAdvertiser() (advertisers []string) {
	switch int(p.vastVersion) {
	case vastVersion2x, vastVersion3x:
		for _, ext := range p.reader.SelectElements(p.adElement, "Extensions", "Extension") {
			if p.reader.SelectAttrValue(ext, "type", "") == "advertiser" {
				ele := p.reader.SelectElement(ext, "Advertiser")
				if ele != nil {
					if value := strings.TrimSpace(p.reader.Text(ele)); len(value) > 0 {
						advertisers = append(advertisers, value)
					}
				}
			}
		}

	case vastVersion4x:
		if ele := p.reader.SelectElement(p.adElement, "Advertiser"); ele != nil {
			if value := strings.TrimSpace(p.reader.Text(ele)); len(value) > 0 {
				advertisers = append(advertisers, value)
			}
		}
	}

	return
}

func (p *fastXMLParser) GetCreativeID() string {
	if p.crID == "" && p.creativeElement != nil {
		p.crID = p.reader.SelectAttrValue(p.creativeElement, "id", "")
	}

	if p.crID == "" {
		p.crID = "cr_" + generateRandomID()
	}
	return p.crID
}

func (p *fastXMLParser) GetDuration() (int, error) {
	if p.creativeElement == nil {
		return 0, errEmptyVideoCreative
	}
	node := p.reader.SelectElement(p.creativeElement, "Linear", "Duration")
	if node == nil {
		return 0, errEmptyVideoDuration
	}
	return parseDuration(strings.TrimSpace(p.reader.RawText(node)))
}

func (p *fastXMLParser) getAdElement(vast *fastxml.Element) *fastxml.Element {
	element := p.reader.SelectElement(vast, "Ad")
	if element == nil {
		return nil
	}
	adElement := p.reader.SelectElement(element, "Wrapper")
	if adElement == nil {
		adElement = p.reader.SelectElement(element, "InLine")
	}
	return adElement
}
