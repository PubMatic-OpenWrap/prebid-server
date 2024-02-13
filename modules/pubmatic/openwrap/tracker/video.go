package tracker

import (
	"errors"
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// Inject Trackers in Video Creative
func injectVideoCreativeTrackers(bid openrtb2.Bid, videoParams []models.OWTracker) (string, error) {
	if bid.AdM == "" || len(videoParams) == 0 {
		return "", errors.New("bid is nil or tracker data is missing")
	}

	originalCreativeStr := bid.AdM
	if strings.HasPrefix(originalCreativeStr, models.HTTPProtocol) {
		originalCreativeStr = strings.Replace(models.VastWrapper, models.PartnerURLPlaceholder, originalCreativeStr, -1)
		originalCreativeStr = strings.Replace(originalCreativeStr, models.TrackerPlaceholder, videoParams[0].TrackerURL, -1)
		originalCreativeStr = strings.Replace(originalCreativeStr, models.ErrorPlaceholder, videoParams[0].ErrorURL, -1)
		bid.AdM = originalCreativeStr
	} else {
		originalCreativeStr = strings.TrimSpace(originalCreativeStr)
		doc := etree.NewDocument()
		if err := doc.ReadFromString(originalCreativeStr); err != nil {
			return bid.AdM, errors.New("invalid creative format")
		}

		//Check VAST Object
		vast := doc.Element.FindElement(models.VideoVASTTag)
		if vast == nil {
			return bid.AdM, errors.New("VAST Tag Not Found")
		}

		//GetVersion
		version := vast.SelectAttrValue(models.VideoVASTVersion, models.VideoVASTVersion2_0)

		adElements := doc.FindElements(models.VASTAdElement)
		for i, adElement := range adElements {
			if i < len(videoParams) {
				element := adElement.FindElement(models.AdWrapperElement)
				isWrapper := (nil != element)

				if nil == element {
					element = adElement.FindElement(models.AdInlineElement)
				}

				if nil == element {
					return bid.AdM, errors.New("video creative not in required VAST format")
				}

				if len(videoParams[i].TrackerURL) > 0 {
					// set tracker URL
					newElement := etree.NewElement(models.ImpressionElement)
					newElement.SetText(videoParams[i].TrackerURL)
					element.InsertChild(element.SelectElement(models.ImpressionElement), newElement)
				}

				if len(videoParams[i].ErrorURL) > 0 {
					// set error URL
					newElement := etree.NewElement(models.ErrorElement)
					newElement.SetText(videoParams[i].ErrorURL)
					element.InsertChild(element.SelectElement(models.ErrorElement), newElement)
				}

				if false == isWrapper && videoParams[i].Price != 0 {
					if models.VideoVASTVersion2_0 == version {
						injectPricingNodeVAST20(element, videoParams[i].Price, videoParams[i].PriceModel, videoParams[i].PriceCurrency)
					} else {
						injectPricingNodeVAST3x(element, videoParams[i].Price, videoParams[i].PriceModel, videoParams[i].PriceCurrency)
					}
				}
			}
		}

		updatedVastStr, err := doc.WriteToString()
		if err != nil {
			return bid.AdM, err
		}
		return updatedVastStr, nil
	}
	return bid.AdM, nil
}

func injectPricingNodeVAST20(parent *etree.Element, price float64, model string, currency string) {
	extensions := parent.FindElement(models.VideoTagLookupStart + models.VideoExtensionsTag)
	if nil == extensions {
		extensions = parent.CreateElement(models.VideoExtensionsTag)
	}

	pricing := extensions.FindElement(models.VideoVAST2ExtensionPriceElement)
	if nil != pricing {
		//Already Present Same Node, So Ignore It
		updatePricingNode(pricing, price, model, currency)
	} else {
		extension := extensions.CreateElement(models.VideoExtensionTag)
		extension.InsertChild(nil, newPricingNode(price, model, currency))
	}
}

func injectPricingNodeVAST3x(parent *etree.Element, price float64, model string, currency string) {
	//Insert into Wrapper Elements
	pricing := parent.FindElement(models.VideoTagLookupStart + models.VideoPricingTag)
	if nil != pricing {
		//Already Present
		updatePricingNode(pricing, price, model, currency)
	} else {
		parent.InsertChild(nil, newPricingNode(price, model, currency))
	}
}

func updatePricingNode(node *etree.Element, price float64, model string, currency string) {
	//Update Price

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

func newPricingNode(price float64, model string, currency string) *etree.Element {
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
