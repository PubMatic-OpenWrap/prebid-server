package spotxtag

import "github.com/PubMatic-OpenWrap/prebid-server/adapters/tagbidder"

var spotxMapper tagbidder.Mapper

func init() {
	spotxMapper = tagbidder.GetNewDefaultMapper()
	/*
		//updating parameter caching status
		spotxMapper.SetCache(MacroTest, true)

		//adding custom macros
		spotxMapper.AddCustomMacro(`ad_unit"`,false)
	*/

	//SetBidderMapper(`spotx`, spotxMapper)
}
