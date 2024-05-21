package wakanda

import (
	"encoding/json"

	"net/http"
	"net/http/httptest"

	"github.com/prebid/openrtb/v20/openrtb2"
)

// DebugData this data will be logged into files
type DebugData struct {
	HTTPRequest        *http.Request
	HTTPRequestBody    json.RawMessage
	HTTPResponse       http.ResponseWriter
	HTTPResponseBody   string
	OpenRTB            *openrtb2.BidRequest
	PrebidHTTPRequest  *http.Request
	PrebidRequestBody  json.RawMessage
	PrebidHTTPResponse *httptest.ResponseRecorder
	Logger             json.RawMessage
	WinningBid         bool
	// PartnerMapping
	// AdUnitConfig
	// Logger
}

type request struct {
	Method      string          `json:"Method,omitempty"`
	Protocol    string          `json:"Protocol,omitempty"`
	Host        string          `json:"Host,omitempty"`
	Path        string          `json:"Path,omitempty"`
	QueryString string          `json:"QueryString,omitempty"`
	Headers     http.Header     `json:"Headers,omitempty"`
	PostJSON    json.RawMessage `json:"PostJSON,omitempty"`
	PostBody    string          `json:"PostBody,omitempty"`
}

func (r *request) set(request *http.Request, postData json.RawMessage) {
	if request != nil {
		r.Method = request.Method
		r.Protocol = request.Proto
		r.Host = request.Host
		r.Headers = request.Header
		r.QueryString = request.URL.RawQuery
		r.Path = request.URL.Path
	}
	if len(postData) > 0 {
		if r.Headers.Get(ContentType) == ContentTypeApplicationJSON {
			r.PostJSON = postData
		} else {
			r.PostBody = string(postData)
		}
	}
}

type response struct {
	StatusCode int             `json:"StatusCode,omitempty"`
	Headers    http.Header     `json:"Headers,omitempty"`
	Body       json.RawMessage `json:"Body,omitempty"`
	Response   string          `json:"Response,omitempty"`
}

func (r *response) setResponseRecorder(resp *httptest.ResponseRecorder) {
	if nil != resp {
		r.StatusCode = resp.Code
		r.Headers = resp.HeaderMap
		if r.Headers.Get(ContentType) == ContentTypeApplicationJSON {
			r.Body = resp.Body.Bytes()[:]
		} else {
			r.Response = resp.Body.String()
		}
	}
}

func (r *response) setResponseWriter(resp http.ResponseWriter, body string) {
	if nil != resp {
		r.Headers = resp.Header()
	}
	if r.Headers.Get(ContentType) == ContentTypeApplicationJSON {
		r.Body = json.RawMessage(body)
	} else {
		r.Response = body
	}
}

type logRecord struct {
	HTTPReq        request              `json:"HTTPReq,omitempty"`
	HTTPResp       response             `json:"HTTPResp,omitempty"`
	OpenRTBRequest *openrtb2.BidRequest `json:"OpenRTBRequest,omitempty"`
	PrebidReq      request              `json:"PrebidReq,omitempty"`
	PrebidResp     response             `json:"PrebidResp,omitempty"`
	Logger         json.RawMessage      `json:"Logger,omitempty"`
	WinningBid     bool                 `json:"WinningBid,omitempty"`
}

// NewLogRecord returns logRecord from wakanda DebugData
func NewLogRecord(wD *DebugData) *logRecord {
	record := &logRecord{
		OpenRTBRequest: wD.OpenRTB,
		WinningBid:     wD.WinningBid,
	}
	record.HTTPReq.set(wD.HTTPRequest, wD.HTTPRequestBody)
	record.HTTPResp.setResponseWriter(wD.HTTPResponse, wD.HTTPResponseBody)
	record.PrebidReq.set(wD.PrebidHTTPRequest, wD.PrebidRequestBody)
	record.PrebidResp.setResponseRecorder(wD.PrebidHTTPResponse)
	record.Logger = wD.Logger

	return record
}
