package events

import (
	"bytes"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/beevik/etree"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

var (
	errEventURLNotConfigured = errors.New("event urls not configured")
	tmpWSRemoverRegex        = regexp.MustCompile(`>\s+<`)
)

// InjectVideoEventTrackers injects the video tracking events
// Returns VAST xml contains as first argument. Second argument indicates whether the trackers are injected and last argument indicates if there is any error in injecting the trackers
func InjectVideoEventTrackers(
	bidRequest *openrtb2.BidRequest,
	bid *openrtb2.Bid,
	vastXML, trackerURL, prebidGenBidId, requestingBidder, bidderCoreName string,
	timestamp int64, fastXMLExperiment bool) (response string, metrics *openrtb_ext.FastXMLMetrics, err error) {

	//Maintaining BidRequest Impression Map (Copied from exchange.go#applyCategoryMapping)
	//TODO: It should be optimized by forming once and reusing
	var imp *openrtb2.Imp
	for _, impr := range bidRequest.Imp {
		if bid.ImpID == impr.ID && impr.Video != nil {
			imp = &impr
			break
		}
	}
	if imp == nil {
		return vastXML, nil, nil
	}

	eventURLMap := GetVideoEventTracking(bidRequest, imp, bid, trackerURL, prebidGenBidId, requestingBidder, bidderCoreName, timestamp)
	if len(eventURLMap) == 0 {
		return vastXML, nil, errEventURLNotConfigured
	}

	adm := strings.TrimSpace(bid.AdM)
	nurlPresent := (adm == "" || strings.HasPrefix(adm, "http"))

	_startTime := time.Now()
	response, err = injectVideoEventsETree(vastXML, eventURLMap, nurlPresent, imp.Video.Linearity)
	etreeParserTime := time.Since(_startTime)

	if fastXMLExperiment && err == nil {
		etreeXMLResponse := response

		_startTime = time.Now()
		fastXMLResponse, _ := injectVideoEventsFastXML(vastXML, eventURLMap, nurlPresent, imp.Video.Linearity)
		fastXMLParserTime := time.Since(_startTime)

		//temporary
		if fastXMLResponse != vastXML {
			fastXMLResponse, etreeXMLResponse = tmpFastXMLProcessing(fastXMLResponse, response)
		}

		isResponseMismatch := (etreeXMLResponse != fastXMLResponse)

		if isResponseMismatch {
			openrtb_ext.FastXMLLogf("\n[XML_PARSER_TEST] method:[vcr] creative:[%s]", base64.StdEncoding.EncodeToString([]byte(vastXML)))
		}

		metrics = &openrtb_ext.FastXMLMetrics{
			XMLParserTime:   fastXMLParserTime,
			EtreeParserTime: etreeParserTime,
			IsRespMismatch:  isResponseMismatch,
		}
	}

	return response, metrics, err
}

func injectVideoEventsETree(vastXML string, eventURLMap map[string]string, nurlPresent bool, linearity adcom1.LinearityMode) (string, error) {

	// parse VAST
	doc := etree.NewDocument()
	if err := doc.ReadFromString(vastXML); err != nil {
		return vastXML, err
	}

	doc.WriteSettings.CanonicalEndTags = true

	creatives := FindCreatives(doc)
	if nurlPresent {
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

func injectVideoEventsFastXML(vastXML string, eventURLMap map[string]string, nurlPresent bool, linearity adcom1.LinearityMode) (string, error) {

	//parse vast xml
	doc := fastxml.NewXMLReader(nil)
	if err := doc.Parse([]byte(vastXML)); err != nil {
		return vastXML, err
	}

	trackersInjected := false
	xu := fastxml.NewXMLUpdater(doc, fastxml.WriteSettings{
		CDATAWrap:    true,
		ExpandInline: true,
	})

	if nurlPresent {
		creative := doc.SelectElement(nil, "VAST", "Ad", "Wrapper", "Creatives")
		if creative != nil {
			cr := fastxml.CreateElement("Creative")

			switch linearity {
			case adcom1.LinearityLinear:
				cr.AddChild(fastxml.CreateElement("Linear").AddChild(getTrackingEvents(true, eventURLMap)))
			case adcom1.LinearityNonLinear:
				cr.AddChild(fastxml.CreateElement("NonLinearAds").AddChild(getTrackingEvents(true, eventURLMap)))
			default:
				cr.AddChild(fastxml.CreateElement("Linear").AddChild(getTrackingEvents(true, eventURLMap)))
				cr.AddChild(fastxml.CreateElement("NonLinearAds").AddChild(getTrackingEvents(true, eventURLMap)))
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

	//Add CDATA and Expand Inline Nodes
	xu.ApplyXMLSettingsOperations()

	var buf bytes.Buffer
	xu.Build(&buf)
	return buf.String(), nil
}

func getTrackingEvents(createTrackingEvents bool, eventURLMap map[string]string) *fastxml.XMLElement {
	te := fastxml.CreateElement("")
	if createTrackingEvents {
		te.SetName("TrackingEvents")
	}

	for _, event := range trackingEvents {
		if url, ok := eventURLMap[event]; ok {
			tracking := fastxml.CreateElement("Tracking").AddAttribute("", "event", event).SetText(url, true, fastxml.NoEscaping)
			te.AddChild(tracking)
		}
	}
	return te
}

func FindCreatives(doc *etree.Document) []*etree.Element {
	// Find Creatives of Linear and NonLinear Type
	// Injecting Tracking Events for Companion is not supported here
	creatives := doc.FindElements("VAST/Ad/InLine/Creatives/Creative/Linear")
	creatives = append(creatives, doc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/Linear")...)
	creatives = append(creatives, doc.FindElements("VAST/Ad/InLine/Creatives/Creative/NonLinearAds")...)
	creatives = append(creatives, doc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/NonLinearAds")...)
	return creatives
}

func tmpFastXMLProcessing(fastXML, etreeXML string) (string, string) {
	//replace only if trackers are injected
	fastXML = strings.TrimSpace(fastXML)                        //step1: remove heading and trailing whitespaces
	fastXML = tmpWSRemoverRegex.ReplaceAllString(fastXML, "><") //step2: remove inbetween whitespaces
	fastXML = strings.ReplaceAll(fastXML, " ><", "><")          //step3: remove attribute endtag whitespace (this should be always before step2)
	fastXML = strings.ReplaceAll(fastXML, "'", "\"")            //step4: convert single quote to double quote

	etreeXML = tmpWSRemoverRegex.ReplaceAllString(etreeXML, "><") //step2: remove inbetween whitespaces
	etreeXML = strings.ReplaceAll(etreeXML, " ><", "><")          //step3: remove attribute endtag whitespace (this should be always before step2)
	etreeXML = strings.ReplaceAll(etreeXML, "'", "\"")
	return fastXML, etreeXML
}
