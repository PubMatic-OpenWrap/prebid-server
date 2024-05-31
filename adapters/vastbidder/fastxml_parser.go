package vastbidder

import (
	"strconv"
	"strings"

	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

type fastXMLParser struct {
	vastTag         *openrtb_ext.ExtImpVASTBidderTag
	reader          *fastxml.XMLReader
	vastVersion     float64
	crID            string
	adElement       *fastxml.Element
	creativeElement *fastxml.Element
}

func newFastXMLParser(vastTag *openrtb_ext.ExtImpVASTBidderTag) *fastXMLParser {
	return &fastXMLParser{
		vastTag: vastTag,
	}
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
	vast1 := p.reader.SelectAttrValue(vast, "version")
	vast1 = vast1
	versionStr := p.reader.SelectAttrValue(vast, "version")
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

	priceValue, err := strconv.ParseFloat(strings.TrimSpace(p.reader.Text(node, true)), 64)
	if nil != err {
		return 0.0, ""
	}

	if currency = p.reader.SelectAttrValue(node, "currency"); currency == "" {
		currency = "USD"
	}

	return priceValue, currency
}

func (p *fastXMLParser) GetAdvertiser() (advertisers []string) {
	switch int(p.vastVersion) {
	case vastVersion2x, vastVersion3x:
		for _, ext := range p.reader.FindElements(p.adElement, "Extensions", "Extension") {
			if p.reader.SelectAttrValue(ext, "type") == "advertiser" {
				ele := p.reader.FindElement(ext, "Advertiser")
				if ele != nil {
					if value := strings.TrimSpace(p.reader.Text(ele, true)); len(value) > 0 {
						advertisers = append(advertisers, value)
					}
				}
			}
		}

	case vastVersion4x:
		if ele := p.reader.FindElement(p.adElement, "Advertiser"); ele != nil {
			if value := strings.TrimSpace(p.reader.Text(ele, true)); len(value) > 0 {
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
		p.crID = p.reader.SelectAttrValue(p.creativeElement, "id")
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
	node := p.reader.FindElement(p.creativeElement, "Linear", "Duration")
	if node == nil {
		return 0, errEmptyVideoDuration
	}
	return parseDuration(strings.TrimSpace(p.reader.Text(node, true)))
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
