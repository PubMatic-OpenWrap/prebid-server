package main_ow

import (
	"github.com/PubMatic-OpenWrap/prebid-server/v2/openrtb_ext"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder"
	"github.com/prebid/prebid-server/v2/config"
)

const (
	requestParamsDirectory  = "./static/bidder-params"
	responseParamsDirectory = "./static/bidder-response-params"
)

// main_ow will perform the openwrap specific initialisation tasks
func main_ow(cfg *config.Configuration) {
	err := ortbbidder.InitBidderParamsConfig(requestParamsDirectory, responseParamsDirectory)
	if err != nil {
		glog.Exitf("Unable to initialise bidder-param mapper for oRTB bidders: %v", err)
	}
	openrtb_ext.SetFastXMLEnablingPercentage(cfg.FastXMLEnabledPercentage)
}
