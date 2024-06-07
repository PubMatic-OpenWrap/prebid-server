package ortbbidder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/adapters/ortbbidder/bidderparams"
	"github.com/prebid/prebid-server/v2/macros"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
)

// requestBuilder is a struct used for constructing RequestData object
type requestBuilder struct {
	rawRequest          json.RawMessage
	requestNode         map[string]any
	imps                []any
	endpoint            string
	hasMacrosInEndpoint bool
}

// parseRequest parse the incoming request and populates intermediate fields required for building requestData object
func (reqBuilder *requestBuilder) parseRequest(request *openrtb2.BidRequest) (err error) {
	reqBuilder.rawRequest, err = jsonutil.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request, err:%s", err.Error())
	}
	reqBuilder.requestNode, err = reqBuilder.buildRequestNode()
	if err != nil {
		return err
	}
	var ok bool
	reqBuilder.imps, ok = reqBuilder.requestNode[impKey].([]any)
	if !ok {
		return errImpMissing
	}
	return
}

// buildRequestNode creates request-map by unmarshaling request-bytes
func (reqBuilder *requestBuilder) buildRequestNode() (requestNode map[string]any, err error) {
	err = jsonutil.Unmarshal(reqBuilder.rawRequest, &requestNode)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request, err:%s", err.Error())
	}
	return requestNode, err
}

// buildEndpoint builds the adapter endpoint, if required it replaces the macros present in endpoint
func (reqBuilder *requestBuilder) buildEndpoint(endpointTemplate *template.Template, bidderParams map[string]any) (string, error) {
	if !reqBuilder.hasMacrosInEndpoint {
		return reqBuilder.endpoint, nil
	}
	uri, err := macros.ResolveMacros(endpointTemplate, bidderParams)
	if err != nil {
		return uri, fmt.Errorf("failed to replace macros in endpoint, err:%s", err.Error())
	}
	uri = strings.ReplaceAll(uri, urlMacroNoValue, "")
	return uri, err
}

// requestModeBuilder is an interface containing parseRequest, makeRequest functions
type requestModeBuilder interface {
	parseRequest(*openrtb2.BidRequest) error
	makeRequest(*template.Template, map[string]bidderparams.BidderParamMapper) ([]*adapters.RequestData, []error)
}

// newRequestBuilder returns the request-builder based on requestMode argument
func newRequestBuilder(requestMode, endpoint string) requestModeBuilder {
	requestBuilder := requestBuilder{
		endpoint:            endpoint,
		hasMacrosInEndpoint: strings.Contains(endpoint, urlMacroPrefix),
	}
	if requestMode == requestModeSingle {
		return &singleRequestModeBuilder{&requestBuilder}
	}
	return &multiRequestModeBuilder{&requestBuilder}
}

// struct to build the request for single request mode where single imp is supported in a request
type singleRequestModeBuilder struct {
	*requestBuilder
}

// makeRequest constructs the endpoint URL and maps the bidder-parameters in request to create the RequestData objects.
// it processes a request to generate 'N' RequestData objects, one for each of the 'N' impressions
func (reqBuilder *singleRequestModeBuilder) makeRequest(endpointTemplate *template.Template,
	paramsMapper map[string]bidderparams.BidderParamMapper) ([]*adapters.RequestData, []error) {
	// set "imp" object in request to empty to improve performance while creating deep copy of request
	var err error
	reqBuilder.rawRequest, err = jsonparser.Set(reqBuilder.rawRequest, []byte("[]"), impKey)
	if err != nil {
		return nil, []error{newBadInputError(errImpSetToEmpty.Error())}
	}
	var (
		uri         string
		errs        []error
		requestData []*adapters.RequestData
	)
	for impIndex := range reqBuilder.imps {
		imp, ok := reqBuilder.imps[impIndex].(map[string]any)
		if !ok || imp == nil {
			errs = append(errs, newBadInputError(fmt.Sprintf("invalid imp object found at index:%d", impIndex)))
			continue
		}
		bidderParams := getImpExtBidderParams(imp)
		// build endpoint-url from bidder-params, it must be done before calling setRequestParams, as it removes the imp.ext.bidder parameters.
		uri, err = reqBuilder.buildEndpoint(endpointTemplate, bidderParams)
		if err != nil {
			errs = append(errs, newBadInputError(err.Error()))
			continue
		}
		// override "imp" key in request to ensure request contains single imp
		reqBuilder.requestNode[impKey] = []any{imp}
		// update the request object by mapping bidderParams at expected location.
		updatedRequest := setRequestParams(reqBuilder.requestNode, bidderParams, paramsMapper, []int{0})
		requestData, err = appendRequestData(requestData, reqBuilder.requestNode, uri)
		if err != nil {
			errs = append(errs, newBadInputError(err.Error()))
		}
		// create a deep copy of the request to ensure common fields are not altered.
		// example - if imp2 modifies the original req.bcat field using its bidder-params, imp1 should still be able to use the original req.bcat value.
		if impIndex != 0 && updatedRequest {
			reqBuilder.requestNode, err = reqBuilder.buildRequestNode()
			if err != nil {
				errs = append(errs, newBadInputError(fmt.Sprintf("failed to build request from rawRequest, err:%s", err.Error())))
				return requestData, errs
			}
		}
	}
	return requestData, errs
}

// struct to build the request for multi request mode where single request supports multiple impressions
type multiRequestModeBuilder struct {
	*requestBuilder
}

// makeRequest constructs the endpoint URL and maps the bidder-parameters in request to create the RequestData objects.
// it create single RequestData object for all impressions.
func (reqBuilder *multiRequestModeBuilder) makeRequest(endpointTemplate *template.Template,
	paramsMapper map[string]bidderparams.BidderParamMapper) ([]*adapters.RequestData, []error) {
	var (
		uri           string
		err           error
		errs          []error
		requestData   []*adapters.RequestData
		foundValidImp bool
	)
	// iterate through imps in reverse order to ensure setRequestParams prioritizes
	// the parameters from imp[0].ext.bidder over those from imp[1..N].ext.bidder.
	for impIndex := len(reqBuilder.imps) - 1; impIndex >= 0; impIndex-- {
		imp, ok := reqBuilder.imps[impIndex].(map[string]any)
		if !ok || imp == nil {
			errs = append(errs, newBadInputError(fmt.Sprintf("invalid imp object found at index:%d", impIndex)))
			continue
		}
		bidderParams := getImpExtBidderParams(imp)
		// build endpoint-url only once using first imp's bidder-params
		if impIndex == 0 {
			uri, err = reqBuilder.buildEndpoint(endpointTemplate, bidderParams)
			if err != nil {
				errs = append(errs, newBadInputError(err.Error()))
				return nil, errs
			}
		}
		// update the request object by mapping bidderParams at expected location.
		setRequestParams(reqBuilder.requestNode, bidderParams, paramsMapper, []int{impIndex})
		foundValidImp = true
	}
	// if not single valid imp is found then return error
	if !foundValidImp {
		return nil, errs
	}
	requestData, err = appendRequestData(requestData, reqBuilder.requestNode, uri)
	if err != nil {
		errs = append(errs, newBadInputError(err.Error()))
	}
	return requestData, errs
}

// appendRequestData creates new RequestData using request and uri then appends it to requestData passed as argument
func appendRequestData(requestData []*adapters.RequestData, request map[string]any, uri string) ([]*adapters.RequestData, error) {
	rawRequest, err := jsonutil.Marshal(request)
	if err != nil {
		return requestData, fmt.Errorf("failed to marshal request after setting bidder-params, err:%s", err.Error())
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
