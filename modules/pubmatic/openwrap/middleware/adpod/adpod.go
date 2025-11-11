package middleware

import (
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/endpoints/legacy/ctv"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	pbc "github.com/prebid/prebid-server/v3/prebid_cache_client"
)

const (
	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
	ApplicationXML  = "application/xml"
	ContentOptions  = "X-Content-Type-Options"
	NoSniff         = "nosniff"
)

type adpod struct {
	handle        httprouter.Handle
	config        *config.Configuration
	cacheClient   *pbc.Client
	metricsEngine metrics.MetricsEngine
}

func NewAdpodWrapperHandle(handleToWrap httprouter.Handle, config *config.Configuration, cc *pbc.Client, me metrics.MetricsEngine) *adpod {
	return &adpod{handle: handleToWrap, config: config, cacheClient: cc, metricsEngine: me}
}

func (a *adpod) OpenrtbEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	adpodResponseWriter := &utils.HTTPResponseBufferWriter{}
	defer a.panicHandler(r)

	if r.Method == http.MethodGet {
		err := enrichRequestBody(r)
		if err != nil {
			a.metricsEngine.RecordBadRequest(models.EndpointORTB, ctv.GetPubIdFromQueryParams(r.URL.Query()), nbr.InvalidVideoRequest.Ptr())
			ext := addErrorInExtension(err.Error(), nil, r.URL.Query().Get(models.Debug))
			errResponse := formErrorBidResponse("", nbr.InvalidVideoRequest.Ptr(), ext)
			w.Header().Set(ContentType, ApplicationJSON)
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errResponse)
			return
		}
	}

	// Invoke prebid auction enpoint
	a.handle(adpodResponseWriter, r, p)

	responseGenerator := ortbResponse{
		debug:              r.URL.Query().Get(models.Debug),
		WrapperLoggerDebug: r.URL.Query().Get(models.WrapperLoggerDebug),
	}
	response, headers, statusCode := responseGenerator.formOperRTBResponse(adpodResponseWriter)

	SetCORSHeaders(w, r)
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(statusCode)
	w.Write(response)
}

func (a *adpod) VastEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if r.Method == http.MethodGet {
		err := enrichRequestBody(r)
		if err != nil {
			a.metricsEngine.RecordBadRequest(models.EndpointVAST, ctv.GetPubIdFromQueryParams(r.URL.Query()), nbr.InvalidVideoRequest.Ptr())
			w.Header().Set(ContentType, ApplicationXML)
			w.Header().Set(HeaderOpenWrapStatus, fmt.Sprintf(NBRFormat, nbr.InvalidVideoRequest))
			w.WriteHeader(http.StatusBadRequest)
			w.Write(EmptyVASTResponse)
			return
		}
	}

	// Invoke prebid auction enpoint
	a.handle(w, r, p)
}

func (a *adpod) JsonEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	adpodResponseWriter := &utils.HTTPResponseBufferWriter{}
	defer a.panicHandler(r)

	if r.Method == http.MethodGet {
		err := enrichRequestBody(r)
		if err != nil {
			a.metricsEngine.RecordBadRequest(models.EndpointJson, ctv.GetPubIdFromQueryParams(r.URL.Query()), nbr.InvalidVideoRequest.Ptr())
			errResponse := formJSONErrorResponse(r, err)
			w.Header().Set(ContentType, ApplicationJSON)
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errResponse)
			return
		}
	}

	// Invoke prebid auction enpoint
	a.handle(adpodResponseWriter, r, p)

	redirectURL := adpodResponseWriter.Header().Get("Location")
	if redirectURL != "" {
		// http.Redirect(w, r, redirectURL, http.StatusFound)
		w.Header().Set("Location", redirectURL)
		w.WriteHeader(http.StatusFound)
		return
	}

	for k, v := range adpodResponseWriter.Header() {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.WriteHeader(adpodResponseWriter.Code)
	w.Write(adpodResponseWriter.Response.Bytes())
}

func (a *adpod) panicHandler(r *http.Request) {
	if recover := recover(); recover != nil {
		a.metricsEngine.RecordPanic(openwrap.GetHostName(), "openwrap-module-middleware")
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
