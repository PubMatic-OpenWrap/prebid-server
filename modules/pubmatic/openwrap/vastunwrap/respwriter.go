package vastunwrap

import (
	"bytes"
	"fmt"
	"net/http"
)

type CustomRecorder struct {
	HeaderMap   http.Header
	Body        *bytes.Buffer
	Code        int
	wroteHeader bool
	snapHeader  http.Header // snapshot of HeaderMap at first Write
}

// NewCustomRecorder returns an initialized ResponseRecorder.
func NewCustomRecorder() *CustomRecorder {
	return &CustomRecorder{
		HeaderMap: make(http.Header),
		Body:      new(bytes.Buffer),
		Code:      200,
	}
}

// Header implements http.ResponseWriter. It returns the response
// headers to mutate within a handler. To test the headers that were
// written after a handler completes, use the Result method and see
// the returned Response value's Header.
func (r *CustomRecorder) Header() http.Header {
	m := r.HeaderMap
	if m == nil {
		m = make(http.Header)
		r.HeaderMap = m
	}
	return m
}

// Write implements http.ResponseWriter. The data is written to
// r.Body, if not nil.
func (r *CustomRecorder) Write(data []byte) (int, error) {
	r.writeHeader(data, "")
	if r.Body != nil {
		r.Body.Write(data)
	}
	return len(data), nil
}

func checkWriteHeaderCode(code int) {
	// Issue 22880: require valid WriteHeader status codes.
	// For now we only enforce that it's three digits.
	// In the future we might block things over 599 (600 and above aren't defined
	// at https://httpwg.org/specs/rfc7231.html#status.codes)
	// and we might block under 200 (once we have more mature 1xx support).
	// But for now any three digits.
	//
	// We used to send "HTTP/1.1 000 0" on the wire in responses but there's
	// no equivalent bogus thing we can realistically send in HTTP/2,
	// so we'll consistently panic instead and help people find their bugs
	// early. (We can't return an error from WriteHeader even if we wanted to.)
	if code < 100 || code > 999 {
		panic(fmt.Sprintf("invalid WriteHeader code %v", code))
	}
}

// writeHeader writes a header if it was not written yet and
// detects Content-Type if needed.
//
// bytes or str are the beginning of the response body.
// We pass both to avoid unnecessarily generate garbage
// in r.WriteString which was created for performance reasons.
// Non-nil bytes win.
func (r *CustomRecorder) writeHeader(b []byte, str string) {
	if r.wroteHeader {
		return
	}
	if len(str) > 512 {
		str = str[:512]
	}

	m := r.Header()

	_, hasType := m["Content-Type"]
	hasTE := m.Get("Transfer-Encoding") != ""
	if !hasType && !hasTE {
		if b == nil {
			b = []byte(str)
		}
		m.Set("Content-Type", http.DetectContentType(b))
	}

	r.WriteHeader(200)
}

// WriteHeader implements http.ResponseWriter.
func (r *CustomRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	checkWriteHeaderCode(code)
	r.Code = code
	r.wroteHeader = true
	if r.HeaderMap == nil {
		r.HeaderMap = make(http.Header)
	}
	r.snapHeader = r.HeaderMap.Clone()
}
