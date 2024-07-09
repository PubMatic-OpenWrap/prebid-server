package openwrap

import (
	"net/http"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// Middleware to set headers for responses
func SetResponseHeaders(serverHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set response headers
		SetSecurityHeaders(w)
		serverHandler.ServeHTTP(w, r)
	})
}

func SetSecurityHeaders(rw http.ResponseWriter) {
	rw.Header().Set("Content-Security-Policy", models.ContentSecurityPolicy)
	rw.Header().Set("X-Content-Type-Options", models.XContentTypeOptions)
	rw.Header().Set("X-XSS-Protection", models.XXSSProtection)
}
