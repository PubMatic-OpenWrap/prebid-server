package main_ow

import (
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder"
)

const paramsDirectory = "./static/bidder-params"

// main_ow will perform the openwrap specific initialisation tasks
func main_ow() {
	err := ortbbidder.InitBidderParamsConfig(paramsDirectory)
	if err != nil {
		glog.Exitf("Unable to initialise bidder-param mapper for oRTB bidders: %v", err)
	}
}
