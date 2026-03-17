package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type FastXMLHandler struct {
	doc     *fastxml.XMLReader
	xu      *fastxml.XMLUpdater
	vastTag *fastxml.Element
	version string
}

func (ti *FastXMLHandler) Parse(vast string) error {
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

func (ti *FastXMLHandler) Inject(videoParams []models.OWTracker, skipTracker bool) (string, error) {
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

func (ti *FastXMLHandler) injectPricingNodeInExtension(adTypeElement *fastxml.Element, price float64, model string, currency string) {
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

func (ti *FastXMLHandler) injectPricingNodeInVAST(adTypeElement *fastxml.Element, price float64, model string, currency string) {
	//Insert into Wrapper Elements
	pricing := ti.doc.SelectElement(adTypeElement, models.VideoPricingTag)
	if nil != pricing {
		//Already Present
		ti.updatePricingNode(pricing, price, model, currency)
	} else {
		ti.xu.AppendElement(adTypeElement, ti.newPricingNode(price, model, currency))
	}
}

func (ti *FastXMLHandler) updatePricingNode(node *fastxml.Element, price float64, model string, currency string) {
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

func (ti *FastXMLHandler) newPricingNode(price float64, model string, currency string) *fastxml.XMLElement {
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

func (ti *FastXMLHandler) addElement(root, base *fastxml.Element, element fastxml.XMLWriter) {
	if nil == base {
		ti.xu.AppendElement(root, element)
	}
	ti.xu.BeforeElement(base, element)
}

func GetTrackerInjector() VASTXMLHandler {
	if openrtb_ext.IsFastXMLEnabled() {
		return &FastXMLHandler{}
	}
	return &etreeHandler{}
}

func (vastXMLHandler *FastXMLHandler) AddAdvertiserTag(advertiser string) (string, error) {
	if vastXMLHandler.doc == nil {
		return "", errors.New("VAST not parsed")
	}
	if len(advertiser) == 0 {
		return "", errors.New("advertiser domain is empty")
	}
	adElements := vastXMLHandler.doc.SelectElements(vastXMLHandler.vastTag, models.VideoAdTag)
	for _, adElement := range adElements {
		adTypeElement := vastXMLHandler.doc.SelectElement(adElement, models.VideoVASTWrapperTag)
		if adTypeElement == nil {
			adTypeElement = vastXMLHandler.doc.SelectElement(adElement, models.VideoVASTInLineTag)
		}
		if adTypeElement != nil {
			domain := vastXMLHandler.doc.SelectElement(adTypeElement, models.VideoAdvertiserTag)
			if domain == nil {
				vastXMLHandler.xu.AppendElement(adTypeElement, vastXMLHandler.newAdvertiserNode(advertiser))
			}
		}
	}
	return vastXMLHandler.xu.String(), nil
}

func (vastXMLHandler *FastXMLHandler) AddCategoryTag(categories []string) (string, error) {
	if vastXMLHandler.doc == nil {
		return "", errors.New("VAST not parsed")
	}
	if len(categories) == 0 {
		return "", errors.New("category is empty")
	}
	adElements := vastXMLHandler.doc.SelectElements(vastXMLHandler.vastTag, models.VideoAdTag)
	for _, adElement := range adElements {
		adTypeElement := vastXMLHandler.doc.SelectElement(adElement, models.VideoVASTWrapperTag)
		if adTypeElement == nil {
			adTypeElement = vastXMLHandler.doc.SelectElement(adElement, models.VideoVASTInLineTag)
		}
		if adTypeElement != nil {
			category := vastXMLHandler.doc.SelectElement(adTypeElement, models.VideoAdCatTag)
			if category == nil {
				vastXMLHandler.xu.AppendElement(adTypeElement, vastXMLHandler.newCategoryNode(categories))
			}
		}
	}
	return vastXMLHandler.xu.String(), nil
}

func (ti *FastXMLHandler) newAdvertiserNode(advertiser string) *fastxml.XMLElement {
	domainElement := fastxml.NewElement(models.VideoAdvertiserTag)
	domainElement.SetText(advertiser, true, fastxml.NoEscaping)
	return domainElement
}
func (ti *FastXMLHandler) newCategoryNode(categories []string) *fastxml.XMLElement {
	catElement := fastxml.NewElement(models.VideoAdCatTag)
	catElement.SetText(strings.Join(categories, ","), true, fastxml.NoEscaping)
	return catElement
}
