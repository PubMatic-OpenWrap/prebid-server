package parser

import (
	"errors"
	"fmt"

	"github.com/beevik/etree"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type VASTXMLHandler interface {
	Parse(vast string) error
	Inject(videoParams []models.OWTracker, skipTracker bool) (string, error)
	AddCategoryTag(adCat []string) (string, error)
	AddAdvertiserTag(adDomain string) (string, error)
}

type etreeHandler struct {
	doc     *etree.Document
	version string
}

func (ti *etreeHandler) Parse(vast string) error {
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

func (ti *etreeHandler) Inject(videoParams []models.OWTracker, skipTracker bool) (string, error) {
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

func (ti *etreeHandler) injectPricingNodeInExtension(adTypeElement *etree.Element, price float64, model string, currency string) {
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

func (ti *etreeHandler) injectPricingNodeInVAST(adTypeElement *etree.Element, price float64, model string, currency string) {
	//Insert into Wrapper Elements
	pricing := adTypeElement.FindElement(models.VideoTagLookupStart + models.VideoPricingTag)
	if nil != pricing {
		//Already Present
		ti.updatePricingNode(pricing, price, model, currency)
	} else {
		adTypeElement.InsertChild(nil, ti.newPricingNode(price, model, currency))
	}
}

func (ti *etreeHandler) updatePricingNode(node *etree.Element, price float64, model string, currency string) {
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

func (ti *etreeHandler) newPricingNode(price float64, model string, currency string) *etree.Element {
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
