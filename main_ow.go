package main_ow

import (
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/adapters/ortbbidder"
)

const (
	requestParamsDirectory   = "./static/bidder-params"
	resposnseParamsDirectory = "./static/bidder-response-params"
)

// main_ow will perform the openwrap specific initialisation tasks
func main_ow() {
	err := ortbbidder.InitBidderParamsConfig(requestParamsDirectory, resposnseParamsDirectory)
	if err != nil {
		glog.Exitf("Unable to initialise bidder-param mapper for oRTB bidders: %v", err)
	}
}
