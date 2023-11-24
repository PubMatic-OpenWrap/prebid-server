package middleware

import (
	"io"
	"net/http"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/utils"
	pbc "github.com/prebid/prebid-server/prebid_cache_client"
)

const (
	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
	ApplicationXML  = "application/xml"
)

type adpod struct {
	handle      httprouter.Handle
	cacheClient *pbc.Client
}

func NewAdpodWrapperHandle(handleToWrap httprouter.Handle, cacheClient *pbc.Client) *adpod {
	return &adpod{handle: handleToWrap, cacheClient: cacheClient}
}

func (a *adpod) OpenrtbEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	adpodResponseWriter := &utils.CustomWriter{}
	defer panicHandler(r)

	// Invoke prebid auction enpoint
	a.handle(adpodResponseWriter, r, p)

	response, headers, statusCode := formOperRTBResponse(adpodResponseWriter)

	SetCORSHeaders(w, r)
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(statusCode)
	w.Write(response)

}

func (a *adpod) VastEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	adpodResponseWriter := &utils.CustomWriter{}
	defer panicHandler(r)

	// Invoke prebid auction enpoint
	a.handle(adpodResponseWriter, r, p)

	responseGenerator := vastResponse{
		debug:              r.URL.Query().Get(models.Debug),
		WrapperLoggerDebug: r.URL.Query().Get(models.WrapperLoggerDebug),
	}
	response, headers, statusCode := responseGenerator.formVastResponse(adpodResponseWriter)

	SetCORSHeaders(w, r)
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(statusCode)
	w.Write(response)
}

func (a *adpod) JsonEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	adpodResponseWriter := &utils.CustomWriter{}
	defer panicHandler(r)

	// Invoke prebid auction enpoint
	a.handle(adpodResponseWriter, r, p)

	responseGenerator := jsonResponse{
		cacheClient: a.cacheClient,
		debug:       r.URL.Query().Get(models.Debug),
	}
	response, headers, statusCode := responseGenerator.formJSONResponse(adpodResponseWriter)

	SetCORSHeaders(w, r)
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(statusCode)
	w.Write(response)
}

// JsonGetEndpoint
func (a *adpod) JsonGetEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	adpodResponseWriter := &utils.CustomWriter{}
	defer panicHandler(r)

	redirectURL, debug, err := getAndValidateRedirectURL(r)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// Invoke prebid auction enpoint
	a.handle(adpodResponseWriter, r, p)

	responseGenerator := jsonResponse{
		cacheClient: a.cacheClient,
		redirectURL: redirectURL,
		debug:       debug,
	}
	response, headers, statusCode := responseGenerator.formJSONResponse(adpodResponseWriter)

	if len(redirectURL) > 0 && debug == "0" {
		http.Redirect(w, r, string(response), http.StatusFound)
		return
	}

	SetCORSHeaders(w, r)
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(statusCode)
	w.Write(response)
}

func panicHandler(r *http.Request) {
	if recover := recover(); recover != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			glog.Error("path:" + r.URL.RequestURI() + " body: " + string(body) + ". stacktrace: \n" + string(debug.Stack()))
			return
		}
		glog.Error("path:" + r.URL.RequestURI() + " body: " + string(body) + ". stacktrace: \n" + string(debug.Stack()))
	}
}

// SetCORSHeaders sets CORS headers in response
func SetCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if len(origin) == 0 {
		origin = "*"
	} else {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
}
