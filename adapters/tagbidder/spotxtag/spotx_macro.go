package spotxtag

import (
	"encoding/json"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters/tagbidder"
	"github.com/PubMatic-OpenWrap/prebid-server/errortypes"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

//SpotxMacro contains openrtb macros for spotx adapter
type SpotxMacro struct {
	*tagbidder.BidderMacro

	/*bidder specific extensions*/
	ext *openrtb_ext.ExtImpSpotX
}

//NewSpotxMacro contains spotx specific parameter parsing
func NewSpotxMacro() tagbidder.IBidderMacro {
	bidder := &SpotxMacro{
		BidderMacro: tagbidder.NewBidderMacro(),
	}
	return bidder
}

//LoadImpression will set current imp
func (tag *SpotxMacro) LoadImpression(imp *openrtb.Imp) error {
	tag.Imp = imp

	//reload ext object
	var bidderExt adapters.ExtImpBidder
	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return &errortypes.BadInput{Message: err.Error()}
	}

	var spotxExt openrtb_ext.ExtImpSpotX
	if err := json.Unmarshal(bidderExt.Bidder, &spotxExt); err != nil {
		return &errortypes.BadInput{Message: err.Error()}
	}

	tag.ext = &spotxExt
	return nil
}

//Custom contains definition for CacheBuster Parameter
func (tag *SpotxMacro) Custom(key string) string {
	//Second Method
	switch key {
	case `channel_id`:
		//do processing
		return tag.ext.ChannelID
	}
	return ""
}

//MacroVideoAPI overriding default behaviour of MacroVideoAPI
func (tag *SpotxMacro) MacroVideoAPI(key string) string {
	return "MP4"
}

/*
Custom Mapper Example
var spotxCustomMapper map[string]func(*SpotxMacro) string

//Second Method of Adding Custom Macro's
func addCustomMacro(key string, cached bool, callback func(*SpotxMacro) string) {
	spotxMapper.AddCustomMacro(key, cached)
	spotxCustomMapper[key] = callback
}

//Second Method
addCustomMacro(`channel_id`, false, channelID)

//Custom contains definition for CacheBuster Parameter
func (tag *SpotxMacro) Custom(key string) string {
	//First Method
	if callback, ok := spotxCustomMapper[key]; ok {
		return callback(tag)
	}
}

func channelID(tag *SpotxMacro) string {
	return tag.ext.ChannelID
}

*/

/*
https://search.spotxchange.com/vast/2.00/85394?VPI=MP4&app[bundle]=[REPLACE_ME]&app[name]=[REPLACE_ME]&app[cat]=[REPLACE_ME]&app[domain]=[REPLACE_ME]&app[privacypolicy]=[REPLACE_ME]&app[storeurl]=[REPLACE_ME]&app[ver]=[REPLACE_ME]&cb=[REPLACE_ME]&device[devicetype]=[REPLACE_ME]&device[ifa]=[REPLACE_ME]&device[make]=[REPLACE_ME]&device[model]=[REPLACE_ME]&device[dnt]=[REPLACE_ME]&player_height=[REPLACE_ME]&player_width=[REPLACE_ME]&ip_addr=[REPLACE_ME]&device[ua]=[REPLACE_ME]]&schain=[REPLACE_ME]

https://search.spotxchange.com/vast/2.00/85394?VPI=MP4&app[bundle]=roku.weatherapp&app[name]=myctvapp&app[cat]=IAB6-8&app[domain]=http%3A%2F%2Fpublishername.com/appname&app[privacypolicy]=1&app[storeurl]=http%3A%2F%2Fchannelstore.roku.com/details/11055/weatherapp&app[ver]=1.2.1&cb=7437276459847&device[devicetype]=7&device[ifa]=236A005B-700F-4889-B9CE-999EAB2B605D&device[make]=Roku&device[model]=Roku&device[dnt]=0&player_height=1080&player_width=1920&ip_addr=165.23.234.23&device[ua]=Roku%2FDVP-7.10%2520(047.10E04062A)]&schain=1.0,1!exchange1.com,1234,1,bid-request-1,publisher,publisher.com,ext_stuff!exchange2.com,abcd,1,bid-request2,intermediary,intermediary.com,other_ext_stuff
*/
