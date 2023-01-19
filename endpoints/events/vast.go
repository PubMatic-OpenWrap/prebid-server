package events

import "encoding/xml"

type VAST struct {
	Ads []Ad `xml:"Ad"`

	Extra      []Node     `xml:",any"`
	Attributes []xml.Attr `xml:",any,attr"`
}

type Ad struct {
	InLine  *InLine  `xml:",omitempty"`
	Wrapper *Wrapper `xml:",omitempty"`

	Extra      []Node     `xml:",any"`
	Attributes []xml.Attr `xml:",any,attr"`
}

type InLine struct {
	Creatives []Creative `xml:"Creatives>Creative"`

	Extra      []Node     `xml:",any"`
	Attributes []xml.Attr `xml:",any,attr"`
}

type Wrapper struct {
	Creatives []CreativeWrapper `xml:"Creatives>Creative"`

	Extra      []Node     `xml:",any"`
	Attributes []xml.Attr `xml:",any,attr"`
}

type Creative struct {
	Linear       *Linear       `xml:",omitempty"`
	NonLinearAds *NonLinearAds `xml:",omitempty"`

	Extra      []Node     `xml:",any"`
	Attributes []xml.Attr `xml:",any,attr"`
}

type WrapperCreative struct {
	TrackingEvents []Tracking `xml:"TrackingEvents>Tracking,omitempty"`

	Extra      []Node     `xml:",any"`
	Attributes []xml.Attr `xml:",any,attr"`
}

type Linear WrapperCreative
type NonLinearAds WrapperCreative
type LinearWrapper WrapperCreative
type NonLinearAdsWrapper WrapperCreative

type CreativeWrapper struct {
	Linear       *LinearWrapper       `xml:",omitempty"`
	NonLinearAds *NonLinearAdsWrapper `xml:"NonLinearAds,omitempty"`

	Extra      []Node     `xml:",any"`
	Attributes []xml.Attr `xml:",any,attr"`
}

type Tracking struct {
	Event string `xml:"event,attr"`
	URI   string `xml:",cdata"`

	Extra      []Node     `xml:",any"`
	Attributes []xml.Attr `xml:",any,attr"`
}

type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content []byte     `xml:",cdata"`
	// Nodes   []Node     `xml:",any"`
}
