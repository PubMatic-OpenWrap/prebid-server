package tracker

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/beevik/etree"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type trackerInjector interface {
	Parse(vast string) error
	Inject(videoParams []models.OWTracker, skipTracker bool) (string, error)
	UpdateADMWithAdvCat(adDomain string, adCat []string) (string, error)
}

type etreeTrackerInjector struct {
	doc     *etree.Document
	version string
}

func (ti *etreeTrackerInjector) Parse(vast string) error {
	ti.doc = etree.NewDocument()
	if err := ti.doc.ReadFromString(vast); err != nil {
		return err
	}

	//Check VAST Object
	vastTag := ti.doc.Element.FindElement(models.VideoVASTTag)
	if vastTag == nil {
		return errors.New("VAST Tag Not Found")
	}

	//GetVersion
	ti.version = vastTag.SelectAttrValue(models.VideoVASTVersion, models.VideoVASTVersion2_0)
	return nil
}

func (ti *etreeTrackerInjector) Inject(videoParams []models.OWTracker, skipTracker bool) (string, error) {
	adElements := ti.doc.FindElements(models.VASTAdElement)
	for i, adElement := range adElements {
		if i < len(videoParams) {
			adTypeElement := adElement.FindElement(models.AdWrapperElement)
			isWrapper := (nil != adTypeElement)
			if adTypeElement == nil {
				adTypeElement = adElement.FindElement(models.AdInlineElement)
			}

			if adTypeElement == nil {
				return "", errors.New("video creative not in required VAST format")
			}

			if !skipTracker && len(videoParams[i].TrackerURL) > 0 {
				// set tracker URL
				impElement := etree.NewElement(models.ImpressionElement)
				impElement.SetText(videoParams[i].TrackerURL)
				adTypeElement.InsertChild(adTypeElement.SelectElement(models.ImpressionElement), impElement)
			}

			if len(videoParams[i].ErrorURL) > 0 {
				// set error URL
				errorElement := etree.NewElement(models.ErrorElement)
				errorElement.SetText(videoParams[i].ErrorURL)
				adTypeElement.InsertChild(adTypeElement.SelectElement(models.ErrorElement), errorElement)
			}

			if videoParams[i].Price != 0 {
				if (ti.version == models.VideoVASTVersion2_0) || (isWrapper && ti.version == models.VideoVASTVersion3_0) {
					ti.injectPricingNodeInExtension(adTypeElement, videoParams[i].Price, videoParams[i].PriceModel, videoParams[i].PriceCurrency)
				} else {
					ti.injectPricingNodeInVAST(adTypeElement, videoParams[i].Price, videoParams[i].PriceModel, videoParams[i].PriceCurrency)
				}
			}
		}
	}
	return ti.doc.WriteToString()
}

func (ti *etreeTrackerInjector) injectPricingNodeInExtension(adTypeElement *etree.Element, price float64, model string, currency string) {
	extensions := adTypeElement.FindElement(models.VideoTagLookupStart + models.VideoExtensionsTag)
	if nil == extensions {
		extensions = adTypeElement.CreateElement(models.VideoExtensionsTag)
	}

	pricing := extensions.FindElement(models.VideoVAST2ExtensionPriceElement)
	if nil != pricing {
		//Already Present Same Node, So Ignore It
		ti.updatePricingNode(pricing, price, model, currency)
	} else {
		extension := extensions.CreateElement(models.VideoExtensionTag)
		extension.InsertChild(nil, ti.newPricingNode(price, model, currency))
	}
}

func (ti *etreeTrackerInjector) injectPricingNodeInVAST(adTypeElement *etree.Element, price float64, model string, currency string) {
	//Insert into Wrapper Elements
	pricing := adTypeElement.FindElement(models.VideoTagLookupStart + models.VideoPricingTag)
	if nil != pricing {
		//Already Present
		ti.updatePricingNode(pricing, price, model, currency)
	} else {
		adTypeElement.InsertChild(nil, ti.newPricingNode(price, model, currency))
	}
}

func (ti *etreeTrackerInjector) updatePricingNode(node *etree.Element, price float64, model string, currency string) {
	//Update Price
	node.Child = nil
	node.SetText(fmt.Sprintf("%v", price))

	//Update Pricing.Model
	if len(model) == 0 {
		model = models.VideoPricingModelCPM
	}
	attrModel := node.SelectAttr(models.VideoPricingModel)
	if nil == attrModel {
		attrModel = node.CreateAttr(models.VideoPricingModel, model)
	} else {
		attrModel.Value = model
	}

	//Update Pricing.Currency
	currencyStr := models.VideoPricingCurrencyUSD
	if currency != "" {
		currencyStr = currency
	}
	attrCurrency := node.SelectAttr(models.VideoPricingCurrency)
	if nil == attrCurrency {
		attrCurrency = node.CreateAttr(models.VideoPricingCurrency, currencyStr)
	} else {
		attrCurrency.Value = currencyStr
	}
}

func (ti *etreeTrackerInjector) newPricingNode(price float64, model string, currency string) *etree.Element {
	pricing := etree.NewElement(models.VideoPricingTag)
	pricing.SetText(fmt.Sprintf("%v", price))
	if len(model) == 0 {
		model = models.VideoPricingModelCPM
	}
	pricing.CreateAttr(models.VideoPricingModel, model)
	currencyStr := models.VideoPricingCurrencyUSD
	if currency != "" {
		currencyStr = currency
	}
	pricing.CreateAttr(models.VideoPricingCurrency, currencyStr)
	return pricing
}

func (ti *etreeTrackerInjector) UpdateADMWithAdvCat(adDomain string, adCat []string) (string, error) {
	if len(adDomain) == 0 && len(adCat) == 0 {
		return "", errors.New("advertiser domain and category are empty")
	}
	adElements := ti.doc.FindElements(models.VASTAdElement)
	for _, adElement := range adElements {
		adTypeElement := adElement.FindElement(models.AdWrapperElement)
		if adTypeElement == nil {
			adTypeElement = adElement.FindElement(models.AdInlineElement)
		}

		if adTypeElement != nil {
			domain := adTypeElement.FindElement(models.VideoTagLookupStart + models.VideoAdDomainTag)
			if domain != nil && len(adDomain) > 0 {
				//Already Present
				ti.updateDomainNode(domain, adDomain)
			} else if len(adDomain) > 0 {
				adTypeElement.InsertChild(nil, ti.newDomainNode(adDomain))
			}
			Cat := adTypeElement.FindElement(models.VideoTagLookupStart + models.VideoAdCatTag)
			if Cat != nil && len(adCat) > 0 {
				//Already Present
				ti.updateCatNode(Cat, adCat)
			} else if len(adCat) > 0 {
				adTypeElement.InsertChild(nil, ti.newCatNode(adCat))
			}

		}
	}
	return ti.doc.WriteToString()
}

func (ti *etreeTrackerInjector) updateDomainNode(node *etree.Element, domain string) {
	//Update Domain
	node.Child = nil
	node.SetText(fmt.Sprintf("<![CDATA[%v]]>", domain))
}

func (ti *etreeTrackerInjector) newDomainNode(domain string) *etree.Element {
	domainElement := etree.NewElement(models.VideoAdDomainTag)
	domainElement.SetText(fmt.Sprintf("<![CDATA[%v]]>", domain))
	return domainElement
}

func (ti *etreeTrackerInjector) updateCatNode(node *etree.Element, cat []string) {
	//Update Category
	node.Child = nil
	node.SetText(fmt.Sprintf("<![CDATA[%s]]>", strings.Join(cat, ",")))
}

func (ti *etreeTrackerInjector) newCatNode(cat []string) *etree.Element {
	catElement := etree.NewElement(models.VideoAdCatTag)
	catElement.SetText(fmt.Sprintf("<![CDATA[%s]]>", strings.Join(cat, ",")))
	return catElement
}

type fastXMLTrackerInjector struct {
	doc     *fastxml.XMLReader
	xu      *fastxml.XMLUpdater
	vastTag *fastxml.Element
	version string
}

func (ti *fastXMLTrackerInjector) Parse(vast string) error {
	ti.doc = fastxml.NewXMLReader()
	if err := ti.doc.Parse([]byte(vast)); err != nil {
		return err
	}

	//Check VAST Object
	if ti.vastTag = ti.doc.SelectElement(nil, "VAST"); ti.vastTag == nil {
		return errors.New("VAST Tag Not Found")
	}

	//GetVersion
	ti.version = ti.doc.SelectAttrValue(ti.vastTag, models.VideoVASTVersion, models.VideoVASTVersion2_0)

	ti.xu = fastxml.NewXMLUpdater(ti.doc, fastxml.WriteSettings{
		CDATAWrap:          true,
		CompressWhitespace: true})

	return nil
}

func (ti *fastXMLTrackerInjector) Inject(videoParams []models.OWTracker, skipTracker bool) (string, error) {
	adElements := ti.doc.SelectElements(ti.vastTag, "Ad")
	for i, adElement := range adElements {
		if i < len(videoParams) {
			adTypeElement := ti.doc.SelectElement(adElement, "Wrapper")
			isWrapper := (nil != adTypeElement)
			if adTypeElement == nil {
				adTypeElement = ti.doc.SelectElement(adElement, "InLine")
			}

			if adTypeElement == nil {
				return "", errors.New("video creative not in required VAST format")
			}

			if !skipTracker && len(videoParams[i].TrackerURL) > 0 {
				// set tracker URL
				impElement := fastxml.NewElement(models.ImpressionElement)
				impElement.SetText(videoParams[i].TrackerURL, true, fastxml.NoEscaping)
				ti.addElement(adTypeElement, ti.doc.SelectElement(adTypeElement, models.ImpressionElement), impElement)
			}

			if len(videoParams[i].ErrorURL) > 0 {
				// set error URL
				errorElement := fastxml.NewElement(models.ErrorElement)
				errorElement.SetText(videoParams[i].ErrorURL, true, fastxml.NoEscaping)
				ti.addElement(adTypeElement, ti.doc.SelectElement(adTypeElement, models.ErrorElement), errorElement)
			}

			if videoParams[i].Price != 0 {
				if (ti.version == models.VideoVASTVersion2_0) || (isWrapper && ti.version == models.VideoVASTVersion3_0) {
					ti.injectPricingNodeInExtension(adTypeElement, videoParams[i].Price, videoParams[i].PriceModel, videoParams[i].PriceCurrency)
				} else {
					ti.injectPricingNodeInVAST(adTypeElement, videoParams[i].Price, videoParams[i].PriceModel, videoParams[i].PriceCurrency)
				}
			}
		}
	}

	// buf := &bytes.Buffer{}
	// ti.xu.Build(buf)
	// return buf.String(), nil
	return ti.xu.String(), nil
}

func (ti *fastXMLTrackerInjector) injectPricingNodeInExtension(adTypeElement *fastxml.Element, price float64, model string, currency string) {
	extensions := ti.doc.SelectElement(adTypeElement, models.VideoExtensionsTag)
	if extensions == nil {
		extension := fastxml.NewElement(models.VideoExtensionTag)
		extension.AddChild(ti.newPricingNode(price, model, currency))

		extensions := fastxml.NewElement(models.VideoExtensionsTag)
		extensions.AddChild(extension)
		ti.xu.AppendElement(adTypeElement, extensions)
		return
	}

	pricing := ti.doc.SelectElement(extensions, models.VideoExtensionTag, models.VideoPricingTag)
	if nil != pricing {
		//Already Present Same Node, So Ignore It
		ti.updatePricingNode(pricing, price, model, currency)
	} else {
		extension := fastxml.NewElement(models.VideoExtensionTag)
		extension.AddChild(ti.newPricingNode(price, model, currency))
		ti.xu.AppendElement(extensions, extension)
	}
}

func (ti *fastXMLTrackerInjector) injectPricingNodeInVAST(adTypeElement *fastxml.Element, price float64, model string, currency string) {
	//Insert into Wrapper Elements
	pricing := ti.doc.SelectElement(adTypeElement, models.VideoPricingTag)
	if nil != pricing {
		//Already Present
		ti.updatePricingNode(pricing, price, model, currency)
	} else {
		ti.xu.AppendElement(adTypeElement, ti.newPricingNode(price, model, currency))
	}
}

func (ti *fastXMLTrackerInjector) updatePricingNode(node *fastxml.Element, price float64, model string, currency string) {
	ti.xu.UpdateText(node, fmt.Sprintf("%v", price), true, fastxml.NoEscaping)

	//Update Pricing.Model
	if len(model) == 0 {
		model = models.VideoPricingModelCPM
	}

	attrModel := ti.doc.SelectAttr(node, models.VideoPricingModel)
	if nil == attrModel {
		ti.xu.AddAttribute(node, "", models.VideoPricingModel, model)
	} else {
		ti.xu.UpdateAttributeValue(attrModel, model)
	}

	//Update Pricing.Currency
	currencyStr := models.VideoPricingCurrencyUSD
	if currency != "" {
		currencyStr = currency
	}
	attrCurrency := ti.doc.SelectAttr(node, models.VideoPricingCurrency)
	if nil == attrCurrency {
		ti.xu.AddAttribute(node, "", models.VideoPricingCurrency, currencyStr)
	} else {
		ti.xu.UpdateAttributeValue(attrCurrency, currencyStr)
	}
}

func (ti *fastXMLTrackerInjector) newPricingNode(price float64, model string, currency string) *fastxml.XMLElement {
	pricing := fastxml.NewElement(models.VideoPricingTag)
	pricing.SetText(fmt.Sprintf("%v", price), true, fastxml.NoEscaping)

	if len(model) == 0 {
		model = models.VideoPricingModelCPM
	}
	pricing.AddAttribute("", models.VideoPricingModel, model)

	if len(currency) == 0 {
		currency = models.VideoPricingCurrencyUSD
	}
	pricing.AddAttribute("", models.VideoPricingCurrency, currency)
	return pricing
}

func (ti *fastXMLTrackerInjector) addElement(root, base *fastxml.Element, element fastxml.XMLWriter) {
	if nil == base {
		ti.xu.AppendElement(root, element)
	}
	ti.xu.BeforeElement(base, element)
}

func GetTrackerInjector() trackerInjector {
	if openrtb_ext.IsFastXMLEnabled() {
		return &fastXMLTrackerInjector{}
	}
	return &etreeTrackerInjector{}
}

func (ti *fastXMLTrackerInjector) UpdateADMWithAdvCat(adDomain string, adCat []string) (string, error) {
	if len(adDomain) == 0 && len(adCat) == 0 {
		return "", errors.New("advertiser domain and category are empty")
	}
	adElements := ti.doc.SelectElements(ti.vastTag, "Ad")
	for _, adElement := range adElements {
		adTypeElement := ti.doc.SelectElement(adElement, "Wrapper")
		if adTypeElement == nil {
			adTypeElement = ti.doc.SelectElement(adElement, "InLine")
		}
		if adTypeElement != nil {
			domain := ti.doc.SelectElement(adTypeElement, models.VideoTagLookupStart+models.VideoAdDomainTag)
			if domain != nil && len(adDomain) > 0 {
				//Already Present
				ti.updateDomainNode(domain, adDomain)
			} else if len(adDomain) > 0 {
				ti.xu.AppendElement(adTypeElement, ti.newDomainNode(adDomain))
			}
			Cat := ti.doc.SelectElement(adTypeElement, models.VideoTagLookupStart+models.VideoAdCatTag)
			if Cat != nil && len(adCat) > 0 {
				//Already Present
				ti.updateCatNode(Cat, adCat)
			} else if len(adCat) > 0 {
				ti.xu.AppendElement(adTypeElement, ti.newCatNode(adCat))
			}
		}
	}
	return ti.xu.String(), nil

}

func (ti *fastXMLTrackerInjector) newDomainNode(domain string) *fastxml.XMLElement {
	domainElement := fastxml.NewElement(models.VideoAdDomainTag)
	domainElement.SetText(fmt.Sprintf("<![CDATA[%v]]>", domain), true, fastxml.NoEscaping)
	return domainElement
}
func (ti *fastXMLTrackerInjector) updateDomainNode(node *fastxml.Element, domain string) {
	ti.xu.UpdateText(node, fmt.Sprintf("<![CDATA[%v]]>", domain), true, fastxml.NoEscaping)
}

func (ti *fastXMLTrackerInjector) newCatNode(cat []string) *fastxml.XMLElement {
	catElement := fastxml.NewElement(models.VideoAdCatTag)
	catElement.SetText(fmt.Sprintf("<![CDATA[%v]]>", cat), true, fastxml.NoEscaping)
	return catElement
}

func (ti *fastXMLTrackerInjector) updateCatNode(node *fastxml.Element, cat []string) {
	ti.xu.UpdateText(node, fmt.Sprintf("<![CDATA[%v]]>", cat), true, fastxml.NoEscaping)
}
