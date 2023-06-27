package router

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	OpenWrapAuction     = "/pbs/openrtb2/auction"
	OpenWrapV25         = "/openrtb/2.5"
	OpenWrapV25Video    = "/openrtb/2.5/video"
	OpenWrapAmp         = "/openrtb/amp"
	OpenWrapHealthcheck = "/healthcheck"
	OpenWrapCTVOrtb     = "/video/openrtb"
	OpenWrapCTVVast     = "/video/vast"
	OpenWrapCTVJson     = "/video/json"
)

// Support legacy APIs for a grace period.
// not implementing middleware to avoid duplicate processing like read, unmarshal, write, etc.
// handling the temporary middleware stuff in EntryPoint hook.
func (r *Router) registerOpenWrapEndpoints(openrtbEndpoint, ampEndpoint httprouter.Handle) {
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
	r.GET(OpenWrapCTVOrtb, openrtbEndpoint)
	r.GET(OpenWrapCTVVast, openrtbEndpoint)
	r.GET(OpenWrapCTVJson, openrtbEndpoint)
	// POST
	r.POST(OpenWrapCTVOrtb, openrtbEndpoint)
	r.POST(OpenWrapCTVVast, openrtbEndpoint)
	r.POST(OpenWrapCTVJson, openrtbEndpoint)

	// healthcheck used by k8s
	r.GET(OpenWrapHealthcheck, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	})
}
