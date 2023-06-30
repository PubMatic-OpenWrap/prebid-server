package adunitconfig

import (
	"encoding/json"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// AdUnitConfig type definition for Ad Unit config parsed from stored config JSON
type AdUnitConfig struct {
	ConfigPattern string               `json:"configPattern,omitempty"`
	Regex         bool                 `json:"regex,omitempty"`
	Config        map[string]*AdConfig //`json:"config"`
	// TODO add seperate default field
	// Default map[string]*AdConfig `json:"default"`
}
type Content struct {
	Mappings  map[string]openrtb_ext.TransparencyRule `json:"mappings,omitempty"`
	Dimension []string                                `json:"dimension,omitempty"`
}
type Transparency struct {
	Content Content `json:"content,omitempty"`
}

type BannerConfig struct {
	openrtb2.Banner
	ClientConfig json.RawMessage `json:"clientconfig,omitempty"`
}

type Banner struct {
	Enabled *bool         `json:"enabled,omitempty"`
	Config  *BannerConfig `json:"config,omitempty"`
}

type VideoConfig struct {
	openrtb2.Video
	ConnectionType []int           `json:"connectiontype,omitempty"`
	ClientConfig   json.RawMessage `json:"clientconfig,omitempty"`
}

type Native struct {
	Enabled *bool         `json:"enabled,omitempty"`
	Config  *NativeConfig `json:"config,omitempty"`
}

type NativeConfig struct {
	openrtb2.Native
	ClientConfig json.RawMessage `json:"clientconfig,omitempty"`
}

type Video struct {
	Enabled *bool        `json:"enabled,omitempty"`
	Config  *VideoConfig `json:"config,omitempty"`
}

// Struct for UniversalPixel
type UniversalPixel struct {
	Id        int      `json:"id,omitempty"`
	Pixel     string   `json:"pixel,omitempty"`
	PixelType string   `json:"pixeltype,omitempty"`
	Pos       string   `json:"pos,omitempty"`
	MediaType string   `json:"mediatype,omitempty"`
	Partners  []string `json:"partners,omitempty"`
}

type AdConfig struct {
	BidFloor    *float64                     `json:"bidfloor,omitempty"`
	BidFloorCur *string                      `json:"bidfloorcur,omitempty"`
	Floors      *openrtb_ext.PriceFloorRules `json:"floors,omitempty"`

	Exp            *int             `json:"exp,omitempty"`
	Banner         *Banner          `json:"banner,omitempty"`
	Native         *Native          `json:"native,omitempty"`
	Video          *Video           `json:"video,omitempty"`
	App            *openrtb2.App    `json:"app,omitempty"`
	Device         *openrtb2.Device `json:"device,omitempty"`
	Transparency   *Transparency    `json:"transparency,omitempty"`
	Regex          *bool            `json:"regex,omitempty"`
	UniversalPixel []UniversalPixel `json:"universalpixel,omitempty"`
}
