package main_ow

import (
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/adapters/ortbbidder"
)

const paramsDirectory = "./static/bidder-params"

// main_ow will perform the openwrap specific initialisation tasks
func main_ow() {
	err := ortbbidder.InitBiddersConfigMap(paramsDirectory)
	if err != nil {
		glog.Exitf("Unable to initialise bidder-param mapper for oRTB bidders: %v", err)
	}
}
