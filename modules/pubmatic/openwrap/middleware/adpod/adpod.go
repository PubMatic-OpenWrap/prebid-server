package middleware

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
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
		aw.Response = make([]byte, len(data))
	}
	return copy(aw.Response, data), nil
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

func NewAdpodWrapperHandle(handleToWrap httprouter.Handle, pbsCacheClient *pbc.Client) *adpod {
	return &adpod{handle: handleToWrap, cacheClient: pbsCacheClient}
}

func (a *adpod) OpenrtbEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	adpodResponseWriter := &AdpodWriter{}
	a.handle(adpodResponseWriter, r, p)

	finalResponse := formOperRTBResponse(adpodResponseWriter.Response)
	w.Header().Set(ContentType, ApplicationJSON)
	w.Write(finalResponse)
	w.WriteHeader(adpodResponseWriter.Code)

}

func (a *adpod) VastEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	adpodResponseWriter := &AdpodWriter{}
	a.handle(adpodResponseWriter, r, p)

	finalResponse := formVastResponse(adpodResponseWriter.Response)
	w.Header().Set(ContentType, ApplicationXML)
	w.Write(finalResponse)
	w.WriteHeader(adpodResponseWriter.Code)
}

func (a *adpod) JsonEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	adpodResponseWriter := &AdpodWriter{}
	a.handle(adpodResponseWriter, r, p)

	finalResponse := formJSONResponse(a.cacheClient, adpodResponseWriter.Response, "")
	w.Header().Set(ContentType, ApplicationJSON)
	w.Write(finalResponse)
	w.WriteHeader(adpodResponseWriter.Code)
}

// JsonGetEndpoint
func (a *adpod) JsonGetEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	redirectURL, err := getAndValidateRedirectURL(r)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	adpodResponseWriter := &AdpodWriter{}
	a.handle(adpodResponseWriter, r, p)

	finalResponse := formJSONResponse(a.cacheClient, adpodResponseWriter.Response, redirectURL)

	if len(redirectURL) > 0 {
		http.Redirect(w, r, string(finalResponse), http.StatusFound)
	} else {
		w.Header().Set(ContentType, ApplicationJSON)
		w.Write(finalResponse)
		w.WriteHeader(adpodResponseWriter.Code)
	}
}
