package middleware

import (
	"io"
	"net/http"
	"runtime/debug"

	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
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
	handle httprouter.Handle
}

func NewAdpodWrapperHandle(handleToWrap httprouter.Handle) *adpod {
	return &adpod{handle: handleToWrap}
}

func (a *adpod) OpenrtbEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	defer func() {
		if recover := recover(); recover != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				glog.Error("path:" + r.URL.RequestURI() + " body: " + string(body) + ". stacktrace: \n" + string(debug.Stack()))
				return
			}
			glog.Error("path:" + r.URL.RequestURI() + " body: " + string(body) + ". stacktrace: \n" + string(debug.Stack()))
		}
	}()

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
	defer func() {
		if recover := recover(); recover != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				glog.Error("path:" + r.URL.RequestURI() + " body: " + string(body) + ". stacktrace: \n" + string(debug.Stack()))
				return
			}
			glog.Error("path:" + r.URL.RequestURI() + " body: " + string(body) + ". stacktrace: \n" + string(debug.Stack()))
		}
	}()

	adpodResponseWriter := &AdpodWriter{}
	a.handle(adpodResponseWriter, r, p)

	finalResponse := formVastResponse(adpodResponseWriter.Response)
	w.Header().Set(ContentType, ApplicationXML)
	if adpodResponseWriter.Code == 0 {
		adpodResponseWriter.Code = http.StatusOK
	}
	w.WriteHeader(adpodResponseWriter.Code)
	w.Write(finalResponse)
}

func (a *adpod) JsonEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	defer func() {
		if recover := recover(); recover != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				glog.Error("path:" + r.URL.RequestURI() + " body: " + string(body) + ". stacktrace: \n" + string(debug.Stack()))
				return
			}
			glog.Error("path:" + r.URL.RequestURI() + " body: " + string(body) + ". stacktrace: \n" + string(debug.Stack()))
		}
	}()

	adpodResponseWriter := &AdpodWriter{}
	a.handle(adpodResponseWriter, r, p)

	finalResponse := formJSONResponse(adpodResponseWriter.Response, "", r.URL.Query().Get(models.Debug))
	w.Header().Set(ContentType, ApplicationJSON)
	if adpodResponseWriter.Code == 0 {
		adpodResponseWriter.Code = http.StatusOK
	}
	w.WriteHeader(adpodResponseWriter.Code)
	w.Write(finalResponse)
}

// JsonGetEndpoint
func (a *adpod) JsonGetEndpoint(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	defer func() {
		if recover := recover(); recover != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				glog.Error("path:" + r.URL.RequestURI() + " body: " + string(body) + ". stacktrace: \n" + string(debug.Stack()))
				return
			}
			glog.Error("path:" + r.URL.RequestURI() + " body: " + string(body) + ". stacktrace: \n" + string(debug.Stack()))
		}
	}()

	redirectURL, debug, err := getAndValidateRedirectURL(r)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	adpodResponseWriter := &AdpodWriter{}
	a.handle(adpodResponseWriter, r, p)

	finalResponse := formJSONResponse(adpodResponseWriter.Response, redirectURL, debug)

	if len(redirectURL) > 0 && debug != "1" {
		http.Redirect(w, r, string(finalResponse), http.StatusFound)
	} else {
		w.Header().Set(ContentType, ApplicationJSON)
		if adpodResponseWriter.Code == 0 {
			adpodResponseWriter.Code = http.StatusOK
		}
		w.WriteHeader(adpodResponseWriter.Code)
		w.Write(finalResponse)
	}
}
