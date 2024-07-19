package events

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/beevik/etree"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
)

var cTrackingEvents = []string{"firstQuartile", "midpoint", "thirdQuartile", "complete", "start"}

// InjectVideoEventTrackers injects the video tracking events
// Returns VAST xml contains as first argument. Second argument indicates whether the trackers are injected and last argument indicates if there is any error in injecting the trackers
func InjectVideoEventTrackers(trackerURL, vastXML string, bid *openrtb2.Bid, prebidGenBidId, requestingBidder, bidderCoreName, accountID string, timestamp int64, bidRequest *openrtb2.BidRequest) ([]byte, bool, error) {

	//Maintaining BidRequest Impression Map (Copied from exchange.go#applyCategoryMapping)
	//TODO: It should be optimized by forming once and reusing
	impMap := make(map[string]*openrtb2.Imp)
	for i := range bidRequest.Imp {
		impMap[bidRequest.Imp[i].ID] = &bidRequest.Imp[i]
	}

	eventURLMap := GetVideoEventTracking(trackerURL, bid, prebidGenBidId, requestingBidder, bidderCoreName, accountID, timestamp, bidRequest, impMap)
	trackersInjected := false
	// return if if no tracking URL
	if len(eventURLMap) == 0 {
		return []byte(vastXML), false, errors.New("Event URLs are not found")
	}

	// parse VAST
	doc := etree.NewDocument()
	err := doc.ReadFromString(vastXML)
	if nil != err {
		err = fmt.Errorf("account:[%s] bidder:[%s] err:[vast_xml_parsing_failed:%s] vast:[%s] ", accountID, requestingBidder, err.Error(), vastXML)
		glog.Error(err.Error())
		return []byte(vastXML), false, err // false indicates events trackers are not injected
	}

	creatives := FindCreatives(doc)

	if adm := strings.TrimSpace(bid.AdM); adm == "" || strings.HasPrefix(adm, "http") {
		// determine which creative type to be created based on linearity
		if imp, ok := impMap[bid.ImpID]; ok && nil != imp.Video {
			// create creative object
			creatives = doc.FindElements("VAST/Ad/Wrapper/Creatives")
			// var creative *etree.Element
			// if len(creatives) > 0 {
			// 	creative = creatives[0] // consider only first creative
			// } else {
			creative := doc.CreateElement("Creative")
			creatives[0].AddChild(creative)

			// }

			switch imp.Video.Linearity {
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
	}
	for _, creative := range creatives {
		trackingEvents := creative.SelectElement("TrackingEvents")
		if nil == trackingEvents {
			trackingEvents = creative.CreateElement("TrackingEvents")
			creative.AddChild(trackingEvents)
		}
		// Inject
		for _, event := range cTrackingEvents {
			if url, ok := eventURLMap[event]; ok {
				trackingEle := trackingEvents.CreateElement("Tracking")
				trackingEle.CreateAttr("event", event)
				trackingEle.SetText(url)
				trackersInjected = true
			}
		}
	}

	out := []byte(vastXML)
	var wErr error
	if trackersInjected {
		out, wErr = doc.WriteToBytes()
		trackersInjected = trackersInjected && nil == wErr
		if nil != wErr {
			glog.Errorf("%v", wErr.Error())
		}
	}
	return out, trackersInjected, wErr
}

// FindCreatives finds Linear, NonLinearAds fro InLine and Wrapper Type of creatives
// from input doc - VAST Document
// NOTE: This function is temporarily seperated to reuse in ctv_auction.go. Because, in case of ctv
// we generate bid.id
func FindCreatives(doc *etree.Document) []*etree.Element {
	// Find Creatives of Linear and NonLinear Type
	// Injecting Tracking Events for Companion is not supported here
	creatives := doc.FindElements("VAST/Ad/InLine/Creatives/Creative/Linear")
	creatives = append(creatives, doc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/Linear")...)
	creatives = append(creatives, doc.FindElements("VAST/Ad/InLine/Creatives/Creative/NonLinearAds")...)
	creatives = append(creatives, doc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/NonLinearAds")...)
	return creatives
}

func injectVideoEventsETree(vastXML string, eventURLMap map[string]string, nurl bool, linearity adcom1.LinearityMode) (string, error) {
	var trackersInjected bool

	// parse VAST
	doc := etree.NewDocument()
	if err := doc.ReadFromString(vastXML); err != nil {
		return vastXML, err
	}

	creatives := FindCreatives(doc)
	if nurl {
		// create creative object
		creatives = doc.FindElements("VAST/Ad/Wrapper/Creatives")
		creative := doc.CreateElement("Creative")
		creatives[0].AddChild(creative)

		switch linearity {
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

	for _, creative := range creatives {
		trackingEvents := creative.SelectElement("TrackingEvents")
		if nil == trackingEvents {
			trackingEvents = creative.CreateElement("TrackingEvents")
			creative.AddChild(trackingEvents)
		}
		// Inject
		for _, event := range cTrackingEvents {
			if url, ok := eventURLMap[event]; ok {
				trackingEle := trackingEvents.CreateElement("Tracking")
				trackingEle.CreateAttr("event", event)
				trackingEle.SetText(url)
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
	return string(out), err
}

func injectVideoEventsFastXML(vastXML string, eventURLMap map[string]string, nurl bool, linearity adcom1.LinearityMode) (string, error) {

	//parse vast xml
	doc := fastxml.NewXMLReader(nil)
	if err := doc.Parse([]byte(vastXML)); err != nil {
		return vastXML, err
	}

	xu := fastxml.NewXMLUpdater([]byte(vastXML))

	// Find creatives
	creatives := doc.SelectElements(nil, "VAST", "Ad", "InLine", "Creatives", "Creative")
	creatives = append(creatives, doc.SelectElements(nil, "VAST", "Ad", "Wrapper", "Creatives", "Creative")...)

	for _, creative := range creatives {
		childs := doc.Childrens(creative)

		found := false
		for _, linearityElement := range childs {
			name := doc.Name(linearityElement)
			if !(name == "Linear" || name == "NonLinearAds") {
				continue
			}
			found = true

			createTrackingEvents := false
			parent := doc.SelectElement(linearityElement, "TrackingEvents")
			if parent == nil {
				createTrackingEvents = true
				parent = linearityElement //Linear/NonLinearAds
			}

			xu.AppendElement(parent, getTrackingEvents(createTrackingEvents, eventURLMap))
		}

		if !found || (len(childs) == 0 && nurl) {
			switch linearity {
			case adcom1.LinearityLinear:
				xu.AppendElement(creative, fastxml.CreateElement("Linear").AddChild(getTrackingEvents(true, eventURLMap)))

			case adcom1.LinearityNonLinear:
				xu.AppendElement(creative, fastxml.CreateElement("NonLinearAds").AddChild(getTrackingEvents(true, eventURLMap)))

			default:
				xu.AppendElement(creative, fastxml.CreateElement("Linear").AddChild(getTrackingEvents(true, eventURLMap)))
				xu.AppendElement(creative, fastxml.CreateElement("NonLinearAds").AddChild(getTrackingEvents(true, eventURLMap)))
			}
			continue
		}
	}

	// wrap cdata
	doc.Iterate(func(element *fastxml.Element) {
		if doc.IsLeaf(element) && !doc.IsCDATA(element) {
			if text := doc.RawText(element); len(text) > 0 {
				xu.UpdateText(element, text, true, fastxml.XMLUnescapeMode)
			}
		}
	})

	var buf bytes.Buffer
	xu.Build(&buf)
	return buf.String(), nil
}

func getTrackingEvents(createTrackingEvents bool, eventURLMap map[string]string) *fastxml.XMLElement {
	te := fastxml.CreateElement("")
	if createTrackingEvents {
		te.SetName("TrackingEvents")
	}

	for _, event := range cTrackingEvents {
		if url, ok := eventURLMap[event]; ok {
			tracking := fastxml.CreateElement("Tracking").AddAttribute("", "event", event).SetText(url, true)
			te.AddChild(tracking)
		}
	}
	return te
}
