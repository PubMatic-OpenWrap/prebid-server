package middleware

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/beevik/etree"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type AdpodBuilder interface {
	Name() string
	Append(bid *openrtb2.Bid) error
	Build() (string, error)
}

type adpodBuilderETree struct {
	vast           *etree.Element
	version        float64
	sequenceNumber int
}

func newAdpodBuilderETree() *adpodBuilderETree {
	return &adpodBuilderETree{
		vast:           etree.NewElement(VASTElement),
		version:        2.0,
		sequenceNumber: 1,
	}
}

func (ab *adpodBuilderETree) Name() string {
	return openrtb_ext.XMLParserETree
}

func (ab *adpodBuilderETree) Append(bid *openrtb2.Bid) error {
	if bid == nil {
		return fmt.Errorf("invalid bid")
	}

	var adElement *etree.Element
	if strings.HasPrefix(bid.AdM, HTTPPrefix) {
		adElement = etree.NewElement(VASTAdElement)
		wrapper := adElement.CreateElement(VASTWrapperElement)
		vastAdTagURI := wrapper.CreateElement(VASTAdTagURIElement)
		vastAdTagURI.CreateCharData(bid.AdM)
	} else {
		adDoc := etree.NewDocument()
		if err := adDoc.ReadFromString(bid.AdM); err != nil {
			return err
		}

		vastTag := adDoc.SelectElement(VASTElement)
		if vastTag == nil {
			return fmt.Errorf("missing vast element")
		}

		//Get Actual VAST Version
		bidVASTVersion, _ := strconv.ParseFloat(vastTag.SelectAttrValue(VASTVersionAttribute, VASTDefaultVersionStr), 64)
		ab.version = math.Max(ab.version, bidVASTVersion)

		ads := vastTag.SelectElements(VASTAdElement)
		if len(ads) == 0 {
			return fmt.Errorf("missing ad element")
		}

		adElement = ads[0].Copy()
	}

	if adElement == nil {
		return fmt.Errorf("vast creative not found")
	}

	//creative.AdId attribute needs to be updated
	adElement.CreateAttr(VASTSequenceAttribute, fmt.Sprint(ab.sequenceNumber))
	ab.vast.AddChild(adElement)
	ab.sequenceNumber++
	return nil
}

func (ab *adpodBuilderETree) Build() (string, error) {
	if int(ab.version) > len(VASTVersionsStr) {
		ab.version = VASTMaxVersion
	}
	ab.vast.CreateAttr(VASTVersionAttribute, VASTVersionsStr[int(ab.version)])

	doc := etree.NewDocument()
	doc.AddChild(ab.vast)
	return doc.WriteToString()
}

type adpodBuilderFastXML struct {
	vast           *fastxml.XMLElement
	version        float64
	sequenceNumber int
}

func newAdpodBuilderFastXML() *adpodBuilderFastXML {
	return &adpodBuilderFastXML{
		vast:           fastxml.NewElement(VASTElement),
		version:        2.0,
		sequenceNumber: 1,
	}
}

func (ab *adpodBuilderFastXML) Name() string {
	return openrtb_ext.XMLParserFastXML
}

func (ab *adpodBuilderFastXML) Append(bid *openrtb2.Bid) error {
	if bid == nil {
		return fmt.Errorf("invalid bid")
	}

	adElement := fastxml.NewElement(constant.VASTAdElement)
	if strings.HasPrefix(bid.AdM, constant.HTTPPrefix) {
		vastAdTagURI := fastxml.NewElement(constant.VASTAdTagURIElement)
		vastAdTagURI.SetText(bid.AdM, true, fastxml.NoEscaping)
		wrapper := fastxml.NewElement(constant.VASTWrapperElement).AddChild(vastAdTagURI)
		adElement.AddChild(wrapper)
	} else {
		adDoc := fastxml.NewXMLReader()
		if err := adDoc.Parse([]byte(bid.AdM)); err != nil {
			return err
		}

		vastTag := adDoc.SelectElement(nil, constant.VASTElement)
		if vastTag == nil {
			return fmt.Errorf("missing vast element")
		}

		ads := adDoc.SelectElements(vastTag, constant.VASTAdElement)
		if len(ads) == 0 {
			return fmt.Errorf("missing ad element")
		}

		adElement.SetText(adDoc.RawText(ads[0]), false, fastxml.NoEscaping)

		//get VAST version
		if value := adDoc.SelectAttrValue(vastTag, constant.VASTVersionAttribute, constant.VASTDefaultVersionStr); value != "" {
			bidVASTVersion, _ := strconv.ParseFloat(value, 64)
			ab.version = math.Max(ab.version, bidVASTVersion)
		}
	}

	if adElement == nil {
		return fmt.Errorf("vast creative not found")
	}

	//creative.AdId attribute needs to be updated
	adElement.AddAttribute("", VASTSequenceAttribute, fmt.Sprint(ab.sequenceNumber))
	ab.vast.AddChild(adElement)
	ab.sequenceNumber++
	return nil
}

func (ab *adpodBuilderFastXML) Build() (string, error) {
	if int(ab.version) > len(VASTVersionsStr) {
		ab.version = VASTMaxVersion
	}
	ab.vast.AddAttribute("", VASTVersionAttribute, VASTVersionsStr[int(ab.version)])

	// buf := &bytes.Buffer{}
	// ab.vast.Write(buf, nil)
	// return buf.String(), nil
	return ab.vast.String(nil), nil
}

func GetAdPodBuilder() AdpodBuilder {
	if openrtb_ext.IsFastXMLEnabled() {
		return newAdpodBuilderFastXML()
	}
	return newAdpodBuilderETree()
}
