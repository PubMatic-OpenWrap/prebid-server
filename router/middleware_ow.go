package router

import (
	"net/http"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/middleware"
)

// OperRTB Writer
type AdpodOpenRTBWriter struct {
	W http.ResponseWriter
}

func (aw AdpodOpenRTBWriter) Write(data []byte) (int, error) {
	data = middleware.FormOperRTBResponse(data)
	return aw.W.Write(data)
}

func (aw AdpodOpenRTBWriter) Header() http.Header {
	return aw.W.Header()
}

func (aw AdpodOpenRTBWriter) WriteHeader(statusCode int) {
	aw.W.WriteHeader(statusCode)
}

// VAST writer
type AdpodVastWriter struct {
	W http.ResponseWriter
}

func (aw AdpodVastWriter) Write(data []byte) (int, error) {
	data = middleware.FormVastResponse(data)
	return aw.W.Write(data)
}

func (aw AdpodVastWriter) Header() http.Header {
	aw.W.Header().Set("Content-Type", "application/xml")
	return aw.W.Header()
}

func (aw AdpodVastWriter) WriteHeader(statusCode int) {
	aw.W.WriteHeader(statusCode)
}

// JSON Writer
type AdpodJSONWriter struct {
	W http.ResponseWriter
}

func (aw AdpodJSONWriter) Write(data []byte) (int, error) {
	data = middleware.FormJSONResponse(g_cacheClient, data)
	return aw.W.Write(data)
}

func (aw AdpodJSONWriter) Header() http.Header {
	return aw.W.Header()
}

func (aw AdpodJSONWriter) WriteHeader(statusCode int) {
	aw.W.WriteHeader(statusCode)
}
