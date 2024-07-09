package utils

import "net/http"

const (
	ContentSecurityPolicy = "frame-ancestors 'self' https://pubmatic.com  https://*.pubmatic.com"
	XContentTypeOptions   = "nosniff"
	XXSSProtection        = "1; mode=block"
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
	rw.Header().Set("Content-Security-Policy", ContentSecurityPolicy)
	rw.Header().Set("X-Content-Type-Options", XContentTypeOptions)
	rw.Header().Set("X-XSS-Protection", XXSSProtection)
}
