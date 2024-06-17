package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/macros"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// requestBuilder is an interface containing parseRequest, makeRequest functions
type requestBuilder interface {
	parseRequest(*openrtb2.BidRequest) error
	makeRequest() ([]*adapters.RequestData, []error)
}

type requestBuilderImpl struct {
	endpoint            string
	endpointTemplate    *template.Template
	requestParams       map[string]bidderparams.BidderParamMapper
	hasMacrosInEndpoint bool
	rawRequest          json.RawMessage
}

// newRequestBuilder returns the request-builder based on requestType argument
func newRequestBuilder(requestType, endpoint string, endpointTemplate *template.Template, requestParams map[string]bidderparams.BidderParamMapper) requestBuilder {
	requestBuilder := requestBuilderImpl{
		endpoint:            endpoint,
		endpointTemplate:    endpointTemplate,
		requestParams:       requestParams,
		hasMacrosInEndpoint: strings.Contains(endpoint, urlMacroPrefix),
	}
	if requestType == multiRequestBuilderType {
		return &multiRequestBuilder{
			requestBuilderImpl: requestBuilder,
		}
	}
	return &singleRequestBuilder{
		requestBuilderImpl: requestBuilder,
	}
}

// getEndpoint returns the endpoint-url, if required replaces macros
func (rb *requestBuilderImpl) getEndpoint(values map[string]any) (string, error) {
	if !rb.hasMacrosInEndpoint {
		return rb.endpoint, nil
	}
	uri, err := macros.ResolveMacros(rb.endpointTemplate, values)
	if err != nil {
		return uri, fmt.Errorf("failed to replace macros in endpoint, err:%s", err.Error())
	}
	uri = strings.ReplaceAll(uri, urlMacroNoValue, "")
	return uri, err
}

func cloneRequest(request json.RawMessage) (map[string]any, error) {
	req := map[string]any{}
	err := jsonutil.Unmarshal(request, &req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// appendRequestData creates new RequestData using request and uri then appends it to requestData passed as argument
func appendRequestData(requestData []*adapters.RequestData, request map[string]any, uri string) ([]*adapters.RequestData, error) {
	rawRequest, err := jsonutil.Marshal(request)
	if err != nil {
		return requestData, err
	}
	requestData = append(requestData, &adapters.RequestData{
		Method: http.MethodPost,
		Uri:    uri,
		Body:   rawRequest,
		Headers: http.Header{
			"Content-Type": {"application/json;charset=utf-8"},
			"Accept":       {"application/json"},
		},
	})
	return requestData, nil
}
