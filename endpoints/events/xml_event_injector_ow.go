package events

import (
	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/beevik/etree"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

type xmlEventInjector interface {
	Name() string
	Inject(vastXML string, eventURLMap map[string]string) (string, error)
}

type etreeEventInjector struct {
	nurlPresent bool
	linearity   adcom1.LinearityMode
}

func newETreeEventInjector(nurlPresent bool, linearity adcom1.LinearityMode) *etreeEventInjector {
	return &etreeEventInjector{
		nurlPresent: nurlPresent,
		linearity:   linearity,
	}
}

func (ev *etreeEventInjector) Name() string {
	return openrtb_ext.XMLParserETree
}

func (ev *etreeEventInjector) Inject(vastXML string, eventURLMap map[string]string) (string, error) {

	// parse VAST
	doc := etree.NewDocument()
	if err := doc.ReadFromString(vastXML); err != nil {
		return vastXML, err
	}

	doc.WriteSettings.CanonicalEndTags = true

	creatives := ev.findCreatives(doc)
	if ev.nurlPresent {
		// create creative object
		creatives = doc.FindElements("VAST/Ad/Wrapper/Creatives")
		creative := doc.CreateElement("Creative")
		creatives[0].AddChild(creative)

		switch ev.linearity {
		case adcom1.LinearityLinear:
			creative.AddChild(doc.CreateElement("Linear"))
		case adcom1.LinearityNonLinear:
			creative.AddChild(doc.CreateElement("NonLinearAds"))
		default: // create both type of creatives
			creative.AddChild(doc.CreateElement("Linear"))
			creative.AddChild(doc.CreateElement("NonLinearAds"))
		}
		creatives = creative.ChildElements() // point to actual cratives
	}

	trackersInjected := false
	for _, creative := range creatives {
		trackingEventsNode := creative.SelectElement("TrackingEvents")
		if nil == trackingEventsNode {
			trackingEventsNode = creative.CreateElement("TrackingEvents")
			creative.AddChild(trackingEventsNode)
		}
		// Inject
		for _, event := range trackingEvents {
			if url, ok := eventURLMap[event]; ok {
				trackingNode := trackingEventsNode.CreateElement("Tracking")
				trackingNode.CreateAttr("event", event)
				trackingNode.SetText(url)
				trackersInjected = true
			}
		}
	}

	if !trackersInjected {
		return vastXML, nil
	}

	out, err := doc.WriteToBytes()
	if err != nil {
		return vastXML, err
	}
	return string(out), nil
}

func (ev *etreeEventInjector) findCreatives(doc *etree.Document) []*etree.Element {
	// Find Creatives of Linear and NonLinear Type
	// Injecting Tracking Events for Companion is not supported here
	creatives := doc.FindElements("VAST/Ad/InLine/Creatives/Creative/Linear")
	creatives = append(creatives, doc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/Linear")...)
	creatives = append(creatives, doc.FindElements("VAST/Ad/InLine/Creatives/Creative/NonLinearAds")...)
	creatives = append(creatives, doc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/NonLinearAds")...)
	return creatives
}

type fastXMLEventInjector struct {
	nurlPresent bool
	linearity   adcom1.LinearityMode
}

func newFastXMLEventInjector(nurlPresent bool, linearity adcom1.LinearityMode) *fastXMLEventInjector {
	return &fastXMLEventInjector{
		nurlPresent: nurlPresent,
		linearity:   linearity,
	}
}

func (ev *fastXMLEventInjector) Name() string {
	return openrtb_ext.XMLParserFastXML
}

func (ev *fastXMLEventInjector) Inject(vastXML string, eventURLMap map[string]string) (string, error) {
	//parse vast xml
	doc := fastxml.NewXMLReader()
	if err := doc.Parse([]byte(vastXML)); err != nil {
		return vastXML, err
	}

	trackersInjected := false
	xu := fastxml.NewXMLUpdater(doc, fastxml.WriteSettings{
		CDATAWrap:    true,
		ExpandInline: true,
	})

	if ev.nurlPresent {
		creative := doc.SelectElement(nil, "VAST", "Ad", "Wrapper", "Creatives")
		if creative != nil {
			cr := fastxml.NewElement("Creative")

			switch ev.linearity {
			case adcom1.LinearityLinear:
				cr.AddChild(fastxml.NewElement("Linear").AddChild(getTrackingEvents(true, eventURLMap)))
			case adcom1.LinearityNonLinear:
				cr.AddChild(fastxml.NewElement("NonLinearAds").AddChild(getTrackingEvents(true, eventURLMap)))
			default:
				cr.AddChild(fastxml.NewElement("Linear").AddChild(getTrackingEvents(true, eventURLMap)))
				cr.AddChild(fastxml.NewElement("NonLinearAds").AddChild(getTrackingEvents(true, eventURLMap)))
			}

			xu.AppendElement(creative, cr)
			trackersInjected = true
		}
	} else {
		// Find creatives
		creatives := doc.SelectElements(nil, "VAST", "Ad", "*", "Creatives", "Creative", "*")

		for _, linearityElement := range creatives {
			name := doc.Name(linearityElement)
			if !(name == "Linear" || name == "NonLinearAds") {
				continue
			}

			createTrackingEvents := false
			parent := doc.SelectElement(linearityElement, "TrackingEvents")
			if parent == nil {
				createTrackingEvents = true
				parent = linearityElement //Linear/NonLinearAds
			}

			xu.AppendElement(parent, getTrackingEvents(createTrackingEvents, eventURLMap))
			trackersInjected = true
		}
	}

	if !trackersInjected {
		return vastXML, nil
	}

	// var buf bytes.Buffer
	// xu.Build(&buf)
	// return buf.String(), nil
	return xu.String(), nil
}

func getTrackingEvents(createTrackingEvents bool, eventURLMap map[string]string) *fastxml.XMLElement {
	te := fastxml.NewElement("")
	if createTrackingEvents {
		te.SetName("TrackingEvents")
	}

	for _, event := range trackingEvents {
		if url, ok := eventURLMap[event]; ok {
			tracking := fastxml.NewElement("Tracking").AddAttribute("", "event", event).SetText(url, true, fastxml.NoEscaping)
			te.AddChild(tracking)
		}
	}
	return te
}

func GetXMLEventInjector(nurlPresent bool, linearity adcom1.LinearityMode) xmlEventInjector {
	if openrtb_ext.IsFastXMLEnabled() {
		return newFastXMLEventInjector(nurlPresent, linearity)
	}
	return newETreeEventInjector(nurlPresent, linearity)
}
