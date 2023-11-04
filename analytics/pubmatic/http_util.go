package pubmatic

// mhttp - Multi Http Calls
// mhttp like multi curl provides a wrapper interface over net/http client to
// create multiple http calls and fire them in parallel. Each http call is fired in
// a separate go routine and waits for all responses for a given timeout;
// All the response are captured automatically in respective individual HttpCall
// structures for further processing

import (
	"bytes"
	"io/ioutil"

	//"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// Default Sizes
const (
	MAX_HTTP_CLIENTS      = 10240 //Max HTTP live clients; clients are global reused in round-robin fashion
	MAX_HTTP_CONNECTIONS  = 1024  //For each HTTP client these many idle connections are created and resued
	MAX_HTTP_CALLS        = 200   //Only these many multi calls are considered at a time;
	HTTP_CONNECT_TIMEOUT  = 500   //Not used at present;
	HTTP_RESPONSE_TIMEOUT = 2000  //Total response timeout = connect time + request write time + wait time + response read time
)

var (
	// Global http client list; One client is used for every mutli call handle
	clients            [MAX_HTTP_CLIENTS]*http.Client
	maxHttpClients     int32 //No of http clients pre-created
	maxHttpConnections int   //no of idle connections per client
	maxHttpCalls       int   //Max allowed http calls in parallel
	nextClientIndex    int32 //Index denotes next client index in round-robin
)

//func dialTimeout(network, addr string) (net.Conn, error) {
//	logger.Debug("Dialling...")
//	timeout := time.Duration(time.Duration(HTTP_CONNECT_TIMEOUT) * time.Millisecond)
//	return net.DialTimeout(network, addr, timeout)
//}

// Init prepares the global list of http clients which are re-used in round-robin
// it should be called only once during start of the application before making any
// multi-http call
//
// maxClients		: Max Http clients to create (<=MAX_HTTP_CLIENTS)
// maxConnections	: Max idle Connections per client (<=MAX_HTTP_CONNECTIONS)
// maxHttpCalls 	: Max allowed http calls in parallel (<= MAX_HTTP_CALLS)
// respTimeout	: http timeout
func Init(maxClients int32, maxConnections, maxCalls, respTimeout int) {
	maxHttpClients = maxClients
	maxHttpConnections = maxConnections
	maxHttpCalls = maxCalls

	if maxHttpClients > MAX_HTTP_CLIENTS {
		maxHttpClients = MAX_HTTP_CLIENTS
	}

	if maxHttpConnections > MAX_HTTP_CONNECTIONS {
		maxHttpConnections = MAX_HTTP_CONNECTIONS
	}

	if maxHttpCalls > MAX_HTTP_CALLS {
		maxHttpCalls = MAX_HTTP_CALLS
	}

	if respTimeout <= 0 || respTimeout >= HTTP_RESPONSE_TIMEOUT {
		respTimeout = HTTP_RESPONSE_TIMEOUT
	}

	timeout := time.Duration(time.Duration(respTimeout) * time.Millisecond)
	for i := int32(0); i < maxClients; i++ {
		//tr := &http.Transport{MaxIdleConnsPerHost: maxConnections, Dial: dialTimeout, ResponseHeaderTimeout: timeout}
		tr := &http.Transport{DisableKeepAlives: false, MaxIdleConnsPerHost: maxConnections}
		clients[i] = &http.Client{Transport: tr, Timeout: timeout}
	}
	nextClientIndex = -1
}

// Wrapper to hold both http request and response data for a single http call
type HttpCall struct {
	//Request Section
	request *http.Request

	//Response Section
	response *http.Response
	err      error
	respBody string
}

// create and returns a HttpCall object
func NewHttpCall(url string, postdata string) (hc *HttpCall, err error) {
	hc = new(HttpCall)
	hc.response = nil
	hc.respBody = ""
	hc.err = nil
	method := "POST"
	if postdata == "" {
		method = "GET"
	}
	hc.request, hc.err = http.NewRequest(method, url, bytes.NewBuffer([]byte(postdata)))
	//hc.request.Close = true
	return
}

// Appends an http header
func (hc *HttpCall) AddHeader(name, value string) {
	hc.request.Header.Add(name, value)
}

// Add http Cookie
func (hc *HttpCall) AddCookie(name, value string) {
	cookie := http.Cookie{Name: name, Value: value}
	hc.request.AddCookie(&cookie)
}

// API call to get the reponse body in string format
func (hc *HttpCall) GetResponseBody() string {
	return hc.respBody
}

// API call to get the reponse body in string format
func (hc *HttpCall) GetResponseHeader(hname string) string {
	return hc.response.Header.Get(hname)
}

// Get response headers map
func (hc *HttpCall) GetResponseHeaders() *http.Header {
	return &hc.response.Header
}

// MultiHttpContext is required to hold the information about all http calls to run
type MultiHttpContext struct {
	hclist  [MAX_HTTP_CALLS]*HttpCall
	hccount int
	wg      sync.WaitGroup
}

// Create a multi-http-context
func NewMultiHttpContext() *MultiHttpContext {
	mhc := new(MultiHttpContext)
	mhc.hccount = 0
	return mhc
}

// Add a http call to multi-http-context
func (mhc *MultiHttpContext) AddHttpCall(hc *HttpCall) {
	if mhc.hccount < maxHttpCalls {
		mhc.hclist[mhc.hccount] = hc
		mhc.hccount += 1
	}
}

// Start firing parallel http calls that have been added so far
// Current go routine is blocked till it finishes with all http calls
// vrc: valid response count
// erc: error reponse count including timeouts
func (mhc *MultiHttpContext) Execute() (vrc int, erc int) {
	vrc = 0 // Mark valid response count to zero
	erc = 0 // Mark invalid response count to zero
	if mhc.hccount <= 0 {
		return
	}

	mhc.wg.Add(mhc.hccount) //Set waitgroup count
	for i := 0; i < mhc.hccount; i++ {
		go mhc.hclist[i].submit(&mhc.wg)
	}
	mhc.wg.Wait() //Wait for all go routines to finish

	for i := 0; i < mhc.hccount; i++ { // validate each response
		if mhc.hclist[i].err == nil && mhc.hclist[i].respBody != "" {
			vrc += 1
		} else {
			erc += 1
		}
	}
	return vrc, erc
}

// Get all the http calls from multi-http-context
func (mhc *MultiHttpContext) GetRequestsFromMultiHttpContext() [MAX_HTTP_CALLS]*HttpCall {
	return mhc.hclist

}

/////////////////////////////////////////////////////////////////////////////////
///   Internal API calls
/////////////////////////////////////////////////////////////////////////////////

// Internal api to get the next http client for use
func getNextHttpClient() *http.Client {
	index := atomic.AddInt32(&nextClientIndex, 1)
	if index >= maxHttpClients {
		index = index % maxHttpClients
		atomic.StoreInt32(&nextClientIndex, index)
	}
	return clients[index]
}

// Internal API to fire individual http call
func (hc *HttpCall) submit(wg *sync.WaitGroup) {
	defer wg.Done()
	client := getNextHttpClient()
	hc.response, hc.err = client.Do(hc.request)
	//logger.Debug("ADCALL RESPONSE :%v" , hc.response)
	if hc.err != nil {
		hc.respBody = ""
		return
	}
	defer hc.response.Body.Close()
	body, err := ioutil.ReadAll(hc.response.Body)
	hc.respBody = string(body)
	hc.err = err
}

// send function will send the owlogger to analytics endpoint
func send(rCtx *models.RequestCtx, url string, headers http.Header) {
	startTime := time.Now()
	mhc := NewMultiHttpContext()
	hc, _ := NewHttpCall(url, "")

	for k, v := range headers {
		if len(v) != 0 {
			hc.AddHeader(k, v[0])
		}
	}

	if rCtx.KADUSERCookie != nil {
		hc.AddCookie(models.KADUSERCOOKIE, rCtx.KADUSERCookie.Value)
	}

	mhc.AddHttpCall(hc)
	_, erc := mhc.Execute()
	if erc != 0 {
		glog.Errorf("Failed to send the owlogger for pub:[%d], profile:[%d], version:[%d].",
			rCtx.PubID, rCtx.ProfileID, rCtx.VersionID)

		// we will not record at version level in prometheus metric
		rCtx.MetricsEngine.RecordPublisherWrapperLoggerFailure(rCtx.PubIDStr, rCtx.ProfileIDStr, "")
		rCtx.MetricsEngine.RecordSendLoggerDataTime(rCtx.Endpoint, rCtx.ProfileIDStr, time.Since(startTime))
		// TODO: this will increment HB specific metric (ow_pbs_sshb_*), verify labels
	}
}
