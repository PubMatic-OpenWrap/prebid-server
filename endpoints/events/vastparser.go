package events

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

var trimRunes = "\t\r\b\n "

func injectTrackersWithCustomXMLParser(vastXML, xmlInput string) (string, bool, error) {
	var outputXML bytes.Buffer
	encoder := xml.NewEncoder(&outputXML)

	trackerInjected := false
	injectTracker := false
	b := strings.NewReader(vastXML)
	p := xml.NewDecoder(b)
	for {
		t, err := p.RawToken()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", false, err
		}
		switch tt := t.(type) {
		case xml.StartElement:
			if tt.Name.Local == "Linear" || tt.Name.Local == "NonLinearAds" {
				injectTracker = true
			} else if tt.Name.Local == "TrackingEvents" {
				injectTracker = false

				encoder.Flush()
				encoder.EncodeToken(tt)
				encoder.Flush()
				outputXML.WriteString(xmlInput)

				trackerInjected = true
				continue
			}
		case xml.EndElement:
			if injectTracker && (tt.Name.Local == "Linear" || tt.Name.Local == "NonLinearAds") {
				injectTracker = false

				encoder.Flush()
				outputXML.WriteString("<TrackingEvents>")
				outputXML.WriteString(xmlInput)
				outputXML.WriteString("</TrackingEvents>")
				encoder.EncodeToken(tt)

				trackerInjected = true
				continue
			}
		case xml.CharData:
			tt2 := strings.Trim(string(tt), trimRunes)
			if len(tt2) != 0 {
				encoder.Flush()
				outputXML.WriteString("<![CDATA[")
				outputXML.WriteString(tt2)
				outputXML.WriteString("]]>")
				continue
			}
		}
		encoder.EncodeToken(t)
	}
	encoder.Flush()
	return outputXML.String(), trackerInjected, nil
}

// package events

// import (
// 	"errors"
// 	"strings"
// )

// /*
// Ways to define XML Tag Elements:
// 	1. Empty: <Linear/>
// 	2. Self closing with Attributes: <Linear version="2"/>
// 	3. Seperate start and end tag and with Attributes: <Linear version="2"></Linear>
// 	4. Seperate start and end tag and without Attributes: <Linear></Linear>

// We don't care about data of these tags since we will inject a valid XML <TrackingEvent> nodes in <TrackingEvents>n beside existing data.

// Final XML would be corrupt only if input XML was already invalid.

// Look for '<Linear' tag, perform above 4 checks and perform below operations.
// 	case 1: Check '/>' and Replace '/>' with '><TRACKER-NODES></Linear>'
// 	case 2: Check ' ' and non Replace first '/>' with '><TRACKER-NODES></Linear>'
// 	case 3: Append '<TRACKER-NODES>' after first '>'
// 	case 4: Append '<TRACKER-NODES>' after first '>'

// Note:
// 	1. Verify usecases around whitespaces between tag name and its closing, attributes, etc. Ex. <a     v="1">, <a   >, <a      />
// 	2. Repeat above steps for all Linear, NonLinearAds tag elements.
// 	3. Use '<TrackingEvents><TRACKER-NODES></TrackingEvents>'if Linear or NonLinearAds tag elements do not have TrackingEvents's tag element defined.
// 	4. How to ignore CDATA. Ex. <Linear in data?
// 	5. Check if VAST has Linear or NonLinearAds tag elements outside Creative. We need to ignore them
// 	6. Check for tags with Perfix. Ex. <LinearABC>
// 	7. Check CDATA recursive
//  8. Handle commentd, etc. Check rawToken() from golang's xml pkg
// */

// const (
// 	TagStartCDATA   = "<![CDATA["
// 	TagEndCDATA     = "]]>"
// 	TagEmptyEnd     = "/>"
// 	TagStartChar    = "<"
// 	TagEndChar      = ">"
// 	TagEndSlashChar = "/"

// 	TagStartLinear       = "<Linear>"
// 	TagEndLinear         = "</Linear>"
// 	TagStartLinearPrefix = "<Linear "  // <Linear version="2">, <Linear    />
// 	TagEndLinearPrefix   = "</Linear " // </Linear   >

// 	TagStartNonLinearAds       = "<NonLinearAds>"
// 	TagEndNonLinearAds         = "</NonLinearAds>"
// 	TagStartNonLinearAdsPrefix = "<NonLinearAds "
// 	TagEndNonLinearAdsPrefix   = "</NonLinearAds "

// 	TagStartTrackingEvents       = "<TrackingEvents>"
// 	TagEndTrackingEvents         = "</TrackingEvents>"
// 	TagStartTrackingEventsPrefix = "<TrackingEvents "
// )

// const (
// 	LenTagStartCDATA   = len("<![CDATA[")
// 	LenTagEndCDATA     = len("]]>")
// 	LenTagEmptyEnd     = len("/>")
// 	LenTagStartChar    = len("<")
// 	LenTagEndChar      = len(">")
// 	LenTagEndSlashChar = len("/")

// 	LenTagStartLinear       = len("<Linear>")
// 	LenTagEndLinear         = len("</Linear>")
// 	LenTagStartLinearPrefix = len("<Linear ")
// 	LenTagEndLinearPrefix   = len("</Linear ")

// 	LenTagStartNonLinearAds       = len("<NonLinearAds>")
// 	LenTagEndNonLinearAds         = len("</NonLinearAds>")
// 	LenTagStartNonLinearAdsPrefix = len("<NonLinearAds ")
// 	LenTagEndNonLinearAdsPrefix   = len("</NonLinearAds ")

// 	LenTagStartTrackingEvents       = len("<TrackingEvents>")
// 	LenTagEndTrackingEvents         = len("</TrackingEvents>")
// 	LenTagStartTrackingEventsPrefix = len("<TrackingEvents ")
// )

// /*

// worse case O(n)^2 ???

// n = no. of charcters in string.

// old:
// 2* O(n) ~ O(n) + log(n) - time (read, build xml tree, write)
// 3 * O(n) ~ O(n) - space (input and output and xml nodes)

// new:

// 2* O(n) ~ O(n) - time (read and write)
// 2 * O(n) ~ O(n) - space (input and output)

// for each <
// 	if < is CDATA tag
// 		ignore xml data till ]]> if found
// 		continue

// 	if <Linear> OR <Linear a="b"  > OR <Linear/> OR <Linear   />
// 		if <Linear/> OR <Linear   />
// 			inject <Linear>trackingXML</Linear>
// 		else
// 			for each <
// 				if < is CDATA tag
// 					ignore xml data till ]]> if found
// 				else if < is TrackingEvents tag
// 					inject <TrackingEvents>trackingXML
// 					break
// 				else if </Linear>
// 					Inject <TrackingEvents>trackingXML</TrackingEvents></Linear>

// */

// func injectTrackers(vastXML, xmlInput string) (string, error) {
// 	// var builder strings.Builder
// 	// Initialize output string
// 	output := ""

// 	currentIndex := 0
// 	for {
// 		// Find the next '<' character
// 		startTagIndex := strings.Index(xmlInput[currentIndex:], TagStartChar)
// 		if startTagIndex == -1 {
// 			output += xmlInput[currentIndex:]
// 			// currentIndex =
// 			break
// 		}
// 		// output += xmlInput[currentIndex:startTagIndex]
// 		// currentIndex = startTagIndex

// 		// We need to ignore CDATA content for false +ve cases. Ex. <![CDATA[<Linear/>]]> inside
// 		// Check if the next characters are <![CDATA[
// 		if xmlInput[currentIndex:startTagIndex+LenTagStartCDATA] == TagStartCDATA {
// 			endTagIndex := strings.Index(xmlInput[startTagIndex+LenTagStartCDATA:], TagEndCDATA)
// 			if endTagIndex == -1 {
// 				// Unable to get closing of CDATA, mostlikely invalid xml, abort!, fallback on old method of tracker XML injection
// 				return "", errors.New("invalid CDATA tag element")
// 			}
// 			output += xmlInput[currentIndex : endTagIndex+LenTagEndCDATA]
// 			currentIndex = endTagIndex + LenTagEndCDATA
// 			continue
// 		}

// 		// currentTagName := ""
// 		// currentTagLen := 0

// 		// if strings.HasPrefix(xmlInput[currentIndex:], TagStartLinear)

// 		// if strings.HasPrefix(xmlInput[currentIndex:], TagStartLinear) ||
// 		// 	strings.HasPrefix(xmlInput[currentIndex:], TagStartLinearPrefix) {

// 		// }

// 		endTagIndex := strings.Index(xmlInput[currentIndex+7:], ">")
// 		if strings.HasPrefix(xmlInput[currentIndex:], "<Linear") {
// 			if endTagIndex == -1 {
// 				return "", errors.New("invalid start tag element")
// 			}
// 			output += "<Linear"
// 			currentIndex += 7

// 			if xmlInput[endTagIndex-1] == '/' { // empty or self closing tag element
// 				output += xmlInput[currentIndex:endTagIndex-1] + "><TRACKER-NODES></Linear>"
// 				currentIndex += len("><TRACKER-NODES></Linear>")
// 			} else { // has seperate end-tag

// 				for {
// 					// Find <TrackingEvents tag by tag to keep O(n)
// 					startTagIndex = strings.Index(xmlInput[currentIndex:], "<")
// 					if startTagIndex == -1 {
// 						output += xmlInput[currentIndex:]
// 						break
// 					}
// 				}

// 				endTagIndex := strings.Index(xmlInput[currentIndex+7:], "</Linear>")
// 				if endTagIndex == -1 {
// 					return "", errors.New("invalid Linear end tag element")
// 				}
// 				// output += xmlInput[startIndex : endTagIndex+1]
// 				// startIndex += len(xmlInput[startIndex : endTagIndex+1])

// 				startTagIndexTKs := strings.Index(xmlInput[currentIndex+7:], "<TrackingEvents")
// 				if startTagIndexTKs == -1 || startTagIndexTKs > endTagIndex {
// 					output += "<TrackingEvents>"
// 					currentIndex += len("<TrackingEvents>")
// 				} else {
// 					endTagIndex := strings.Index(xmlInput[currentIndex+len("<TrackingEvents")+1:], ">")
// 					if xmlInput[endTagIndex-1] == '/' { // empty or self closing tag element
// 						output += xmlInput[currentIndex:endTagIndex-1] + ">"
// 						currentIndex += endTagIndex
// 						startTagIndexTKs = -1 // auto add end tag
// 					} else { // has seperate end-tag
// 						output += xmlInput[currentIndex:endTagIndex-1] + "><TRACKER-NODES></Linear>"
// 						currentIndex += len("><TRACKER-NODES></Linear>")
// 					}
// 				}

// 				output += xmlInput[currentIndex:endTagIndex+1] + "<TRACKER-NODES>"
// 				currentIndex += len(xmlInput[currentIndex:endTagIndex+1] + "<TRACKER-NODES>")

// 				if startTagIndexTKs == -1 {
// 					output += "</TrackingEvents>"
// 					currentIndex += len("</TrackingEvents>")
// 				}
// 			}
// 		} else if strings.HasPrefix(xmlInput[currentIndex:], "<NonLinearAds") {

// 		}
// 	}
// 	return output, nil
// }

// func skipCDATA(xmlInput, output string, currentIndex int) (bool, error) {
// 	// Check if the next characters are <![CDATA[
// 	if xmlInput[currentIndex:currentIndex+LenTagStartCDATA] == TagStartCDATA {
// 		endTagIndex := strings.Index(xmlInput[currentIndex+LenTagStartCDATA:], TagEndCDATA)
// 		if endTagIndex == -1 {
// 			// We need to ignore CDATA content for false +ve cases. Ex. <![CDATA[<Linear/>]]> inside
// 			// Unable to get closing of CDATA, mostlikely invalid xml, abort!, fallback on old method of tracker XML injection
// 			return false, errors.New("invalid CDATA tag element")
// 		}
// 		output += xmlInput[currentIndex : endTagIndex+LenTagEndCDATA]
// 		currentIndex = endTagIndex + LenTagEndCDATA
// 		return true, nil
// 	}
// 	return false, nil
// }

// func injectTrackers2(vastXML, xmlInput string) (string, error) {
// 	var builder strings.Builder

// 	for i := 0; i < len(vastXML); i++ {
// 		if vastXML[i] == '<' && vastXML[i+1] != '/' {
// 			// Ignore everything in <![CDATA[ ... ]]>
// 			if vastXML[i:LenTagStartCDATA] == TagStartCDATA {
// 				endTagIndex := strings.Index(xmlInput[i+LenTagStartCDATA:], TagEndCDATA)
// 				if endTagIndex == -1 {
// 					// Unable to get closing of CDATA, mostlikely invalid xml, abort!, fallback on old method of tracker XML injection
// 					return "", errors.New("invalid CDATA tag element")
// 				}
// 				_, _ = builder.Write([]byte(vastXML[i : endTagIndex+LenTagEndCDATA]))
// 				i = endTagIndex + LenTagEndCDATA // -1? since i+ will skip new <
// 				continue
// 			}

// 			// <Linear>
// 			// <Linear a="b">
// 			// <Linear a="b" >
// 			// <Linear a="b" 	>
// 			// <Linear	>
// 			// <Linear      >
// 			// <Linear/>
// 			// <Linear />
// 			// <Linear	/>

// 			// Process only <Linear> and <NonLinearAds>
// 			currentEndTag := "" // Process only </Linear> and </NonLinearAds>
// 			j := i + 1
// 			for ; j < len(vastXML); j++ {
// 				if vastXML[j] == ' ' || vastXML[j] == '\t' || vastXML[j] == '>' || vastXML[j] == '/' {
// 					// <Linear a="b">, <Linear	>, <Linear >
// 					tagName := vastXML[i:j]
// 					if tagName == "<Linear" {
// 						currentEndTag = TagEndLinear
// 					} else if tagName == "<NonLinearAds" {
// 						currentEndTag = TagEndNonLinearAds
// 					}
// 				}

// 				if currentEndTag == "" {
// 					continue
// 				}

// 			}
// 		}
// 		_ = builder.WriteByte(vastXML[i])
// 	}

// 	return "", nil
// }
// */
