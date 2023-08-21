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
	OpenWrapAdpodOrtb   = "/video/openrtb"
	OpenWrapAdpodVast   = "/video/vast"
	OpenWrapAdpodJson   = "/video/json"
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
	r.GET(OpenWrapAdpodOrtb, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		customWriter := AdpodOpenRTBWriter{W: w}
		openrtbEndpoint(customWriter, r, p)
	})
	r.GET(OpenWrapAdpodVast, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		customWriter := AdpodVastWriter{W: w}
		openrtbEndpoint(customWriter, r, p)
		customWriter.Header().Set("Content-Type", "application/xml")
	})
	r.GET(OpenWrapAdpodJson, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		redirectURL, err := GetAndValidateRedirectURL(r)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		customWriter := AdpodJSONWriter{W: w, RedirectURL: redirectURL}
		openrtbEndpoint(customWriter, r, p)
	})

	// POST
	r.POST(OpenWrapAdpodOrtb, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		customWriter := AdpodOpenRTBWriter{W: w}
		openrtbEndpoint(customWriter, r, p)
	})
	r.POST(OpenWrapAdpodVast, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		customWriter := AdpodVastWriter{W: w}
		openrtbEndpoint(customWriter, r, p)
		customWriter.Header().Set("Content-Type", "application/xml")
	})
	r.POST(OpenWrapAdpodJson, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		customWriter := AdpodJSONWriter{W: w}
		openrtbEndpoint(customWriter, r, p)
	})

	// healthcheck used by k8s
	r.GET(OpenWrapHealthcheck, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	})
}
