package middleware

import (
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	pbc "github.com/prebid/prebid-server/prebid_cache_client"
)

const (
	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
	ApplicationXML  = "application/xml"
)

// Custom Adpod Writer
type AdpodWriter struct {
	Response []byte
	Headers  http.Header
	Code     int
}

func (aw *AdpodWriter) Write(data []byte) (int, error) {
	if data == nil {
		return 0, nil
	}

	if aw.Response == nil {
		aw.Response = make([]byte, 0)
	}
	aw.Response = append(aw.Response, data...)
	return len(data), nil
}

func (aw *AdpodWriter) Header() http.Header {
	if aw.Headers == nil {
		aw.Headers = make(http.Header)
	}
	return aw.Headers
}

func (aw *AdpodWriter) WriteHeader(statusCode int) {
	aw.Code = statusCode
}

type adpod struct {
	handle      httprouter.Handle
	cacheClient *pbc.Client
}

func NewAdpodWrapperHandle(handleToWrap httprouter.Handle, cacheClient *pbc.Client) *adpod {
	return &adpod{handle: handleToWrap, cacheClient: cacheClient}
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

func (a *adpod) OpenrtbEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	defer panicHandler(r)

	adpodResponseWriter := &AdpodWriter{}
	a.handle(adpodResponseWriter, r, p)

	finalResponse := formOperRTBResponse(adpodResponseWriter.Response)
	w.Header().Set(ContentType, ApplicationJSON)
	if adpodResponseWriter.Code == 0 {
		adpodResponseWriter.Code = http.StatusOK
	}
	w.WriteHeader(adpodResponseWriter.Code)
	w.Write(finalResponse)

}

func (a *adpod) VastEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	defer panicHandler(r)

	adpodResponseWriter := &AdpodWriter{}
	a.handle(adpodResponseWriter, r, p)

	finalResponse, err := formVastResponse(adpodResponseWriter.Response)
	if err != nil {
		w.Header().Set(HeaderOpenWrapStatus, fmt.Sprintf(ErrorFormat, 4, "No Bid"))
	}
	w.Header().Set(ContentType, ApplicationXML)
	if adpodResponseWriter.Code == 0 {
		adpodResponseWriter.Code = http.StatusOK
	}
	w.WriteHeader(adpodResponseWriter.Code)
	w.Write(finalResponse)
}

func (a *adpod) JsonEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	defer panicHandler(r)

	adpodResponseWriter := &AdpodWriter{}
	a.handle(adpodResponseWriter, r, p)

	responseGenerator := jsonResponse{
		cacheClient: a.cacheClient,
		debug:       r.URL.Query().Get(models.Debug),
	}
	finalResponse := responseGenerator.formJSONResponse(adpodResponseWriter.Response)

	w.Header().Set(ContentType, ApplicationJSON)
	if adpodResponseWriter.Code == 0 {
		adpodResponseWriter.Code = http.StatusOK
	}
	w.WriteHeader(adpodResponseWriter.Code)
	w.Write(finalResponse)
}

// JsonGetEndpoint
func (a *adpod) JsonGetEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	defer panicHandler(r)

	redirectURL, debug, err := getAndValidateRedirectURL(r)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	adpodResponseWriter := &AdpodWriter{}
	a.handle(adpodResponseWriter, r, p)

	responseGenerator := jsonResponse{
		cacheClient: a.cacheClient,
		redirectURL: redirectURL,
		debug:       debug,
	}
	finalResponse := responseGenerator.formJSONResponse(adpodResponseWriter.Response)

	if len(redirectURL) > 0 && debug == "0" {
		http.Redirect(w, r, string(finalResponse), http.StatusFound)
		return
	}

	w.Header().Set(ContentType, ApplicationJSON)
	if adpodResponseWriter.Code == 0 {
		adpodResponseWriter.Code = http.StatusOK
	}
	w.WriteHeader(adpodResponseWriter.Code)
	w.Write(finalResponse)
}
