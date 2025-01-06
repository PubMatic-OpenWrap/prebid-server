package router

import (
	"net/http"

	unwrap "git.pubmatic.com/vastunwrap/unwrap"
	"github.com/PubMatic-OpenWrap/prebid-server/v2/exchange"
	"github.com/julienschmidt/httprouter"
	middleware "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/middleware/adpod"
)

const (
	OpenWrapAuction     = "/pbs/openrtb2/auction"
	OpenWrapV25         = "/openrtb/2.5"
	OpenWrapV25Video    = "/openrtb/2.5/video"
	OpenWrapAmp         = "/openrtb/amp"
	OpenWrapHealthcheck = "/healthcheck"
	OpenWrapAdpodOrtb   = "/video/openrtb"
	OpenWrapAdpodVast   = "/video/vast"
	OpenWrapAdpodJson   = "/video/json"
)

// Support legacy APIs for a grace period.
// not implementing middleware to avoid duplicate processing like read, unmarshal, write, etc.
// handling the temporary middleware stuff in EntryPoint hook.
func (r *Router) registerOpenWrapEndpoints(openrtbEndpoint, ampEndpoint httprouter.Handle) {
	adpod := middleware.NewAdpodWrapperHandle(openrtbEndpoint, g_cfg, g_cacheClient, r.MetricsEngine)

	//OpenWrap hybrid
	r.POST(OpenWrapAuction, openrtbEndpoint)

	// OpenWrap 2.5 in-app, etc
	r.POST(OpenWrapV25, openrtbEndpoint)

	// OpenWrap 2.5 video
	r.GET(OpenWrapV25Video, openrtbEndpoint)

	// OpenWrap AMP
	r.POST(OpenWrapAmp, ampEndpoint)

	// CTV/OTT
	//GET
	r.GET(OpenWrapAdpodOrtb, adpod.OpenrtbEndpoint)
	r.GET(OpenWrapAdpodVast, adpod.VastEndpoint)
	r.GET(OpenWrapAdpodJson, adpod.JsonGetEndpoint)

	// POST
	r.POST(OpenWrapAdpodOrtb, adpod.OpenrtbEndpoint)
	r.POST(OpenWrapAdpodVast, adpod.VastEndpoint)
	r.POST(OpenWrapAdpodJson, adpod.JsonEndpoint)

	// healthcheck used by k8s
	r.GET(OpenWrapHealthcheck, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	})
	initFastXMLTest()
}

func initFastXMLTest() {
	if g_metrics != nil {
		unwrap.InitRecordFastXMLTestMetrics(func(ctx *unwrap.UnwrapContext, etreeResp, fastxmlResp *unwrap.UnwrapResponse) {
			exchange.RecordFastXMLTestMetrics(g_metrics, ctx, etreeResp, fastxmlResp)
		})
	}
}
