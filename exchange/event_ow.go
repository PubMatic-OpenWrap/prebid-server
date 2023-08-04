package exchange

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/macros"
	"github.com/prebid/prebid-server/metrics"
)

var (
	creativeId = regexp.MustCompile(`^<Creative.*id\s*=\s*"([0-9]*)".*>$`)
)

const (
	errorVast = `<VAST version=\"3.0\"><Ad><Wrapper>
	<AdSystem>prebid.org wrapper</AdSystem>
	<VASTAdTagURI><![CDATA[" %s "]]></VASTAdTagURI>
	<Creatives></Creatives>
	</Wrapper></Ad></VAST>`
)

const (
	creativesStartTag         = "<Creatives>"
	trackingEventsTagStartTag = "<TrackingEvents>"
	trackingEventsTagEndTag   = "</TrackingEvents>"
	videoClicksStartTag       = "<VideoClicks>"
	videoClicksEndTag         = "</VideoClicks>"
	nonLinearStartTag         = "<NonLinear>"
	nonLinearEndTag           = "</NonLinear>"
	linearEndTag              = "</Linear>"
	nonLinearAdsEndTag        = "</NonLinearAds>"
	wrapperEndTag             = "</Wrapper>"
	wrapperStartTag           = "<Wrapper>"
	inLineEndTag              = "</InLine>"
	adSystemEndTag            = "</AdSystem>"
	creativeEndTag            = "</Creative>"
	companionStartTag         = "<Companion>"
	companionEndTag           = "</Companion>"
	impressionEndTag          = "</Impression>"
	companionAdsEndTag        = "</CompanionAds>"
	adElementEndTag           = "</Ad>"
	errorEndTag               = "</Error>"
)

type Injector interface {
	Build(vastXML, nURL string) string
}

type Events struct {
	defaultURL string
	vastEvents map[string][]config.VASTEvent
}
type TrackerInjector struct {
	replacer macros.Replacer
	events   Events
	me       metrics.MetricsEngine
	provider *macros.MacroProvider
}

func NewTrackerInjector(replacer macros.Replacer, provider *macros.MacroProvider, events Events) Injector {
	return &TrackerInjector{
		replacer: replacer,
		provider: provider,
		events:   events,
	}
}

type Ad struct {
	WrapperInlineEndIndex int
	ImpressionEndIndex    int
	ErrorEndIndex         int
	Creatives             []Creative
}

type Creative struct {
	Linear       *Linear
	NonLinearAds *NonLinearAds
	CompanionAds *CompanionAds
	CreativeID   string
}

type NonLinearAds struct {
	TrackingEvent     int
	NonLinears        []int
	NonLinearAdsIndex int
}

type CompanionAds struct {
	Companion         []int
	CompanionAdsIndex int
}

type Linear struct {
	VideoClick     int
	TrackingEvent  int
	LinearEndIndex int
}

// Preallocate a strings.Builder and a byte slice for the final result
var builderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// pair maintains the index and tag to be injected
type pair struct {
	pos int
	tag string
}

func (builder *TrackerInjector) Build(vastXML string, NURL string) string {

	//TODO
	if vastXML == "" && NURL == "" {
		// log <bidder-name>.requests.badserverresponse metric and details in debug response.ext.errors.BIDDERCODE output.
		return ""
	}
	if vastXML == "" && NURL != "" {
		return fmt.Sprintf(errorVast, NURL)
	}

	ads := parseVastXML([]byte(vastXML))
	//TODO
	//	for _, ad := range ads {
	// add metrics
	// log adapter.<bidder-name>.requests.badserverresponse  if wrapper and inline tag is not present
	//	}
	pairs := builder.buildPairs(ads)
	//sort all events position
	sort.SliceStable(pairs[:], func(i, j int) bool {
		return pairs[i].pos < pairs[j].pos
	})

	// Reuse a preallocated strings.Builder
	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)
	offset := 0
	for i := range pairs {
		if offset != pairs[i].pos {
			buf.WriteString(vastXML[offset:pairs[i].pos])
			offset = pairs[i].pos
		}
		buf.WriteString(pairs[i].tag)
	}
	buf.WriteString(vastXML[offset:])
	return buf.String()
}

func parseVastXML(vastXML []byte) []Ad {

	var (
		vastTags  = make([]Ad, 0, 10)
		currIndex int
	)

	ad := Ad{}
	creative := Creative{}
	trackingEventEndIndex := 0
	videoClick := 0
	nonLinearAds := make([]int, 0, 10)
	companions := make([]int, 0, 10)
	length := len(vastXML)

	for currIndex < length {
		if vastXML[currIndex] == '<' {
			currIndex++
			if currIndex < length && vastXML[currIndex] == '/' {
				iterator := currIndex
				for iterator < length && vastXML[iterator] != '>' {
					iterator++
				}
				if iterator < length {
					tag := vastXML[currIndex+1 : iterator]
					handleTag(string(tag), &ad, &creative, &vastTags, &trackingEventEndIndex, &videoClick, &nonLinearAds, &companions, currIndex-1)
				}
				currIndex = iterator
			} else if currIndex+1 < length && vastXML[currIndex] == 'C' && vastXML[currIndex+1] == 'r' {
				iterator := currIndex
				for iterator < length && vastXML[iterator] != '>' {
					iterator++
				}

				match := creativeId.FindSubmatch(vastXML[currIndex-1 : iterator+1])
				if len(match) > 1 {
					creative.CreativeID = string(match[1])
				}
				currIndex = iterator
			}
		}
		currIndex++
	}
	return vastTags
}

func handleTag(tag string, ad *Ad, creative *Creative, vastTags *[]Ad, trackingEventEndIndex *int, videoClick *int, nonLinearAds *[]int, companions *[]int, index int) {
	switch tag {
	case "Ad":
		*vastTags = append(*vastTags, *ad)
		*ad = Ad{}
	case "Impression":
		ad.ImpressionEndIndex = index
	case "Error":
		ad.ErrorEndIndex = index
	case "Creative":
		ad.Creatives = append(ad.Creatives, *creative)
		*creative = Creative{}
	case "InLine", "Wrapper":
		ad.WrapperInlineEndIndex = index
	case "TrackingEvents":
		*trackingEventEndIndex = index
	case "VideoClicks":
		*videoClick = index
	case "NonLinear":
		*nonLinearAds = append(*nonLinearAds, index)
	case "NonLinearAds":
		creative.NonLinearAds = &NonLinearAds{
			TrackingEvent:     *trackingEventEndIndex,
			NonLinearAdsIndex: index,
			NonLinears:        *nonLinearAds,
		}
	case "Linear":
		creative.Linear = &Linear{
			TrackingEvent:  *trackingEventEndIndex,
			LinearEndIndex: index,
			VideoClick:     *videoClick,
		}
		*videoClick = 0
		*trackingEventEndIndex = 0
	case "Companion":
		*companions = append(*companions, index)
	case "CompanionAds":
		creative.CompanionAds = &CompanionAds{
			CompanionAdsIndex: index,
			Companion:         *companions,
		}
	}
}

func (builder *TrackerInjector) buildPairs(vastTags []Ad) []pair {
	pairs := make([]pair, 0, len(vastTags)*4)
	for _, tag := range vastTags {
		if tag.ImpressionEndIndex != 0 {
			pairs = append(pairs, pair{pos: tag.ImpressionEndIndex + len("</Impression>"), tag: builder.getEvent("", "impression")})
		} else {
			pairs = append(pairs, pair{pos: tag.WrapperInlineEndIndex, tag: builder.getEvent("", "impression")})
		}
		if tag.ErrorEndIndex != 0 {
			pairs = append(pairs, pair{pos: tag.ErrorEndIndex + len("</Error>"), tag: builder.getEvent("", "error")})
		} else {
			pairs = append(pairs, pair{pos: tag.WrapperInlineEndIndex, tag: builder.getEvent("", "error")})
		}

		for _, creative := range tag.Creatives {
			if creative.Linear != nil {
				if creative.Linear.TrackingEvent == 0 {
					pairs = append(pairs, pair{pos: creative.Linear.LinearEndIndex, tag: trackingEventsTagStartTag + builder.getEvent(creative.CreativeID, "tracking") + trackingEventsTagEndTag})
				} else {
					pairs = append(pairs, pair{pos: creative.Linear.TrackingEvent, tag: builder.getEvent(creative.CreativeID, "tracking")})
				}

				if creative.Linear.VideoClick == 0 {
					pairs = append(pairs, pair{pos: creative.Linear.LinearEndIndex, tag: videoClicksStartTag + builder.getEvent(creative.CreativeID, "clicktracking") + videoClicksEndTag})
				} else {
					pairs = append(pairs, pair{pos: creative.Linear.VideoClick, tag: builder.getEvent(creative.CreativeID, "clicktracking")})
				}
			}

			if creative.NonLinearAds != nil {
				if creative.NonLinearAds.TrackingEvent == 0 {
					pairs = append(pairs, pair{pos: creative.NonLinearAds.NonLinearAdsIndex, tag: trackingEventsTagStartTag + builder.getEvent(creative.CreativeID, "tracking") + trackingEventsTagEndTag})
				} else {
					pairs = append(pairs, pair{pos: creative.NonLinearAds.TrackingEvent, tag: builder.getEvent(creative.CreativeID, "tracking")})
				}

				if len(creative.NonLinearAds.NonLinears) == 0 {
					pairs = append(pairs, pair{pos: creative.NonLinearAds.NonLinearAdsIndex, tag: nonLinearStartTag + builder.getEvent(creative.CreativeID, "nonlinearclicktracking") + nonLinearEndTag})
				} else {
					for _, nonLinear := range creative.NonLinearAds.NonLinears {
						pairs = append(pairs, pair{pos: nonLinear, tag: builder.getEvent(creative.CreativeID, "nonlinearclicktracking")})
					}
				}
			}

			if creative.CompanionAds != nil {
				if len(creative.CompanionAds.Companion) == 0 {
					pairs = append(pairs, pair{pos: creative.CompanionAds.CompanionAdsIndex, tag: companionStartTag + builder.getEvent(creative.CreativeID, "companionclickthrough") + companionEndTag})
				} else {
					for _, companion := range creative.CompanionAds.Companion {
						pairs = append(pairs, pair{pos: companion, tag: builder.getEvent(creative.CreativeID, "companionclickthrough")})
					}
				}
			}
		}
	}

	return pairs
}

func (builder *TrackerInjector) getEvent(creativeId string, eventType string) string {
	buf := builderPool.Get().(*strings.Builder)
	buf.Reset()
	defer builderPool.Put(buf)
	switch eventType {

	case "impression":
		for _, event := range builder.events.vastEvents["impression"] {
			for _, url := range event.URLs {
				url, err := builder.replacer.Replace(url, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<Impression><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></Impression>`)
			}
			if !event.ExcludeDefaultURL {
				url, err := builder.replacer.Replace(builder.events.defaultURL, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<Impression><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></Impression>`)
			}
		}

	case "error":
		for _, vastEvent := range builder.events.vastEvents["error"] {
			for _, url := range vastEvent.URLs {
				url, err := builder.replacer.Replace(url, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<Error><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></Error>`)
			}
			if !vastEvent.ExcludeDefaultURL {
				url, err := builder.replacer.Replace(builder.events.defaultURL, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<Error><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></Error>`)
			}
		}
	case "tracking":
		for _, vastEvent := range builder.events.vastEvents["tracking"] {
			builder.provider.PopulateEventMacros(creativeId, string(vastEvent.CreateElement), string(vastEvent.Type))
			for _, url := range vastEvent.URLs {
				url, err := builder.replacer.Replace(url, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<Tracking event="`)
				buf.WriteString(string(vastEvent.Type))
				buf.WriteString(`"><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></Tracking>`)
			}
			if !vastEvent.ExcludeDefaultURL {
				url, err := builder.replacer.Replace(builder.events.defaultURL, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<Tracking event="`)
				buf.WriteString(string(vastEvent.Type))
				buf.WriteString(`"><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></Tracking>`)
			}
		}
	case "nonlinearclicktracking":
		for _, vastEvent := range builder.events.vastEvents["nonlinearclicktracking"] {
			builder.provider.PopulateEventMacros(creativeId, string(vastEvent.CreateElement), string(vastEvent.Type))
			for _, url := range vastEvent.URLs {
				url, err := builder.replacer.Replace(url, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<NonLinearClickTracking><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></NonLinearClickTracking>`)

			}
			if !vastEvent.ExcludeDefaultURL {
				url, err := builder.replacer.Replace(builder.events.defaultURL, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<NonLinearClickTracking><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></NonLinearClickTracking>`)
			}
		}
	case "clicktracking":
		for _, vastEvent := range builder.events.vastEvents["clicktracking"] {
			builder.provider.PopulateEventMacros(creativeId, string(vastEvent.CreateElement), string(vastEvent.Type))
			for _, url := range vastEvent.URLs {
				url, err := builder.replacer.Replace(url, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<ClickTracking><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></ClickTracking>`)
			}
			if !vastEvent.ExcludeDefaultURL {
				url, err := builder.replacer.Replace(builder.events.defaultURL, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<ClickTracking><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></ClickTracking>`)
			}
		}
	case "companionclickthrough":
		for _, vastEvent := range builder.events.vastEvents["companionclickthrough"] {
			builder.provider.PopulateEventMacros(creativeId, string(vastEvent.CreateElement), string(vastEvent.Type))
			for _, url := range vastEvent.URLs {
				url, err := builder.replacer.Replace(url, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<CompanionClickThrough><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></CompanionClickThrough>`)
			}
			if !vastEvent.ExcludeDefaultURL {
				url, err := builder.replacer.Replace(builder.events.defaultURL, builder.provider)
				if err != nil {
					continue
				}
				buf.WriteString(`<CompanionClickThrough><![CDATA[`)
				buf.WriteString(url)
				buf.WriteString(`]]></CompanionClickThrough>`)
			}
		}
	}

	return buf.String()
}
