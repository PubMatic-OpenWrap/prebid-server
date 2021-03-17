package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/PubMatic-OpenWrap/etree"
	// "github.com/beevik/etree"
	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/prebid_cache_client"
	"github.com/PubMatic-OpenWrap/prebid-server/stored_requests"
	"github.com/stretchr/testify/assert"
)

const (
	maxSize = 1024 * 256

	vastXmlWithImpressionWithContent    = "<VAST version=\"3.0\"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[adm2]]></VASTAdTagURI><Impression>content</Impression><Creatives></Creatives></Wrapper></Ad></VAST>"
	vastXmlWithImpressionWithoutContent = "<VAST version=\"3.0\"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[adm2]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></Wrapper></Ad></VAST>"
	vastXmlWithoutImpression            = "<VAST version=\"3.0\"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[adm2]]></VASTAdTagURI><Creatives></Creatives></Wrapper></Ad></VAST>"
)

// Mock pbs cache client
type vtrackMockCacheClient struct {
	Fail  bool
	Error error
	Uuids []string
}

func (m *vtrackMockCacheClient) PutJson(ctx context.Context, values []prebid_cache_client.Cacheable) ([]string, []error) {
	if m.Fail {
		return []string{}, []error{m.Error}
	}
	return m.Uuids, []error{}
}
func (m *vtrackMockCacheClient) GetExtCacheData() (scheme string, host string, path string) {
	return
}

// Test
func TestShouldRespondWithBadRequestWhenAccountParameterIsMissing(t *testing.T) {
	// mock pbs cache client
	mockCacheClient := &vtrackMockCacheClient{}

	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{}

	// mock config
	cfg := &config.Configuration{
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// prepare
	reqData := ""

	req := httptest.NewRequest("POST", "/vtrack", strings.NewReader(reqData))
	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: nil,
		Cache:       mockCacheClient,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	d, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	assert.Equal(t, 400, recorder.Result().StatusCode, "Expected 400 on request with missing account parameter")
	assert.Equal(t, "Account 'a' is required query parameter and can't be empty", string(d))
}

func TestShouldRespondWithBadRequestWhenRequestBodyIsEmpty(t *testing.T) {
	// mock pbs cache client
	mockCacheClient := &vtrackMockCacheClient{}

	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{}

	// config
	cfg := &config.Configuration{
		MaxRequestSize:  maxSize,
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// prepare
	reqData := ""

	req := httptest.NewRequest("POST", "/vtrack?a=events_enabled", strings.NewReader(reqData))

	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: nil,
		Cache:       mockCacheClient,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	d, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	assert.Equal(t, 400, recorder.Result().StatusCode, "Expected 400 on request with empty body")
	assert.Equal(t, "Invalid request: request body is empty\n", string(d))
}

func TestShouldRespondWithBadRequestWhenRequestBodyIsInvalid(t *testing.T) {
	// mock pbs cache client
	mockCacheClient := &vtrackMockCacheClient{}

	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{}

	// config
	cfg := &config.Configuration{
		MaxRequestSize:  maxSize,
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// prepare
	reqData := "invalid"

	req := httptest.NewRequest("POST", "/vtrack?a=events_enabled", strings.NewReader(reqData))

	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: nil,
		Cache:       mockCacheClient,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	// validate
	assert.Equal(t, 400, recorder.Result().StatusCode, "Expected 400 on request with invalid body")
}

func TestShouldRespondWithBadRequestWhenBidIdIsMissing(t *testing.T) {
	// mock pbs cache client
	mockCacheClient := &vtrackMockCacheClient{}

	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{}

	// config
	cfg := &config.Configuration{
		MaxRequestSize:  maxSize,
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// prepare
	data := &BidCacheRequest{
		Puts: []prebid_cache_client.Cacheable{
			{},
		},
	}

	reqData, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/vtrack?a=events_enabled", strings.NewReader(string(reqData)))

	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: nil,
		Cache:       mockCacheClient,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	d, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	assert.Equal(t, 400, recorder.Result().StatusCode, "Expected 400 on request with elements missing bidid")
	assert.Equal(t, "Invalid request: 'bidid' is required field and can't be empty\n", string(d))
}

func TestShouldRespondWithBadRequestWhenBidderIsMissing(t *testing.T) {
	// mock pbs cache client
	mockCacheClient := &vtrackMockCacheClient{}

	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{}

	// config
	cfg := &config.Configuration{
		MaxRequestSize:  maxSize,
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// prepare
	data := &BidCacheRequest{
		Puts: []prebid_cache_client.Cacheable{
			{
				BidID: "test",
			},
		},
	}

	reqData, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/vtrack?a=events_enabled", strings.NewReader(string(reqData)))

	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: nil,
		Cache:       mockCacheClient,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	d, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	assert.Equal(t, 400, recorder.Result().StatusCode, "Expected 400 on request with elements missing bidder")
	assert.Equal(t, "Invalid request: 'bidder' is required field and can't be empty\n", string(d))
}

func TestShouldRespondWithInternalServerErrorWhenPbsCacheClientFails(t *testing.T) {
	// mock pbs cache client
	mockCacheClient := &vtrackMockCacheClient{
		Fail:  true,
		Error: fmt.Errorf("pbs cache client failed"),
	}

	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{}

	// config
	cfg := &config.Configuration{
		MaxRequestSize: maxSize, VTrack: config.VTrack{
			TimeoutMS: int64(2000), AllowUnknownBidder: true,
		},
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// prepare
	data, err := getValidVTrackRequestBody(false, false)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/vtrack?a=events_enabled", strings.NewReader(data))

	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: nil,
		Cache:       mockCacheClient,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	d, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	assert.Equal(t, 500, recorder.Result().StatusCode, "Expected 500 when pbs cache client fails")
	assert.Equal(t, "Error(s) updating vast: pbs cache client failed\n", string(d))
}

func TestShouldTolerateAccountNotFound(t *testing.T) {
	// mock pbs cache client
	mockCacheClient := &vtrackMockCacheClient{}

	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{
		Fail:  true,
		Error: stored_requests.NotFoundError{},
	}

	// config
	cfg := &config.Configuration{
		MaxRequestSize: maxSize, VTrack: config.VTrack{
			TimeoutMS: int64(2000), AllowUnknownBidder: false,
		},
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// prepare
	data, err := getValidVTrackRequestBody(true, false)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/vtrack?a=1235", strings.NewReader(data))

	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: nil,
		Cache:       mockCacheClient,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	// validate
	assert.Equal(t, 200, recorder.Result().StatusCode, "Expected 200 when account is not found and request is valid")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestShouldSendToCacheExpectedPutsAndUpdatableBiddersWhenBidderVastNotAllowed(t *testing.T) {
	// mock pbs cache client
	mockCacheClient := &vtrackMockCacheClient{
		Fail:  false,
		Uuids: []string{"uuid1"},
	}

	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{
		Fail: false,
	}

	// config
	cfg := &config.Configuration{
		MaxRequestSize: maxSize, VTrack: config.VTrack{
			TimeoutMS: int64(2000), AllowUnknownBidder: false,
		},
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// bidder info
	bidderInfos := make(adapters.BidderInfos)
	bidderInfos["bidder"] = adapters.BidderInfo{
		Status:                  adapters.StatusActive,
		ModifyingVastXmlAllowed: false,
	}
	bidderInfos["updatable_bidder"] = adapters.BidderInfo{
		Status:                  adapters.StatusActive,
		ModifyingVastXmlAllowed: true,
	}

	// prepare
	data, err := getValidVTrackRequestBody(false, false)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/vtrack?a=events_enabled", strings.NewReader(data))

	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: bidderInfos,
		Cache:       mockCacheClient,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	d, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	assert.Equal(t, 200, recorder.Result().StatusCode, "Expected 200 when account is not found and request is valid")
	assert.Equal(t, "{\"responses\":[{\"uuid\":\"uuid1\"}]}", string(d), "Expected 200 when account is found and request is valid")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestShouldSendToCacheExpectedPutsAndUpdatableBiddersWhenBidderVastAllowed(t *testing.T) {
	// mock pbs cache client
	mockCacheClient := &vtrackMockCacheClient{
		Fail:  false,
		Uuids: []string{"uuid1", "uuid2"},
	}

	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{
		Fail: false,
	}

	// config
	cfg := &config.Configuration{
		MaxRequestSize: maxSize, VTrack: config.VTrack{
			TimeoutMS: int64(2000), AllowUnknownBidder: false,
		},
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// bidder info
	bidderInfos := make(adapters.BidderInfos)
	bidderInfos["bidder"] = adapters.BidderInfo{
		Status:                  adapters.StatusActive,
		ModifyingVastXmlAllowed: true,
	}
	bidderInfos["updatable_bidder"] = adapters.BidderInfo{
		Status:                  adapters.StatusActive,
		ModifyingVastXmlAllowed: true,
	}

	// prepare
	data, err := getValidVTrackRequestBody(true, true)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/vtrack?a=events_enabled", strings.NewReader(data))

	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: bidderInfos,
		Cache:       mockCacheClient,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	d, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	assert.Equal(t, 200, recorder.Result().StatusCode, "Expected 200 when account is not found and request is valid")
	assert.Equal(t, "{\"responses\":[{\"uuid\":\"uuid1\"},{\"uuid\":\"uuid2\"}]}", string(d), "Expected 200 when account is found and request is valid")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestShouldSendToCacheExpectedPutsAndUpdatableUnknownBiddersWhenUnknownBidderIsAllowed(t *testing.T) {
	// mock pbs cache client
	mockCacheClient := &vtrackMockCacheClient{
		Fail:  false,
		Uuids: []string{"uuid1", "uuid2"},
	}

	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{
		Fail: false,
	}

	// config
	cfg := &config.Configuration{
		MaxRequestSize: maxSize, VTrack: config.VTrack{
			TimeoutMS: int64(2000), AllowUnknownBidder: true,
		},
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// bidder info
	bidderInfos := make(adapters.BidderInfos)

	// prepare
	data, err := getValidVTrackRequestBody(true, false)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/vtrack?a=events_enabled", strings.NewReader(data))

	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: bidderInfos,
		Cache:       mockCacheClient,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	d, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	assert.Equal(t, 200, recorder.Result().StatusCode, "Expected 200 when account is not found and request is valid")
	assert.Equal(t, "{\"responses\":[{\"uuid\":\"uuid1\"},{\"uuid\":\"uuid2\"}]}", string(d), "Expected 200 when account is found, request has unknown bidders but allowUnknownBidders is enabled")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestShouldReturnBadRequestWhenRequestExceedsMaxRequestSize(t *testing.T) {
	// mock pbs cache client
	mockCacheClient := &vtrackMockCacheClient{
		Fail:  false,
		Uuids: []string{"uuid1", "uuid2"},
	}

	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{
		Fail: false,
	}

	// config
	cfg := &config.Configuration{
		MaxRequestSize: 1,
		VTrack: config.VTrack{
			TimeoutMS: int64(2000), AllowUnknownBidder: true,
		},
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// bidder info
	bidderInfos := make(adapters.BidderInfos)

	// prepare
	data, err := getValidVTrackRequestBody(true, false)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/vtrack?a=events_enabled", strings.NewReader(data))

	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: bidderInfos,
		Cache:       mockCacheClient,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	d, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	assert.Equal(t, 400, recorder.Result().StatusCode, "Expected 400 when request exceeds max request size")
	assert.Equal(t, "Invalid request: request size exceeded max size of 1 bytes\n", string(d))
}

func TestShouldRespondWithInternalErrorPbsCacheIsNotConfigured(t *testing.T) {
	// mock AccountsFetcher
	mockAccountsFetcher := &mockAccountsFetcher{
		Fail: false,
	}

	// config
	cfg := &config.Configuration{
		MaxRequestSize: maxSize, VTrack: config.VTrack{
			TimeoutMS: int64(2000), AllowUnknownBidder: false,
		},
		AccountDefaults: config.Account{},
	}
	cfg.MarshalAccountDefaults()

	// prepare
	data, err := getValidVTrackRequestBody(true, true)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/vtrack?a=events_enabled", strings.NewReader(data))
	recorder := httptest.NewRecorder()

	e := vtrackEndpoint{
		Cfg:         cfg,
		BidderInfos: nil,
		Cache:       nil,
		Accounts:    mockAccountsFetcher,
	}

	// execute
	e.Handle(recorder, req, nil)

	d, err := ioutil.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	// validate
	assert.Equal(t, 500, recorder.Result().StatusCode, "Expected 500 when pbs cache is not configured")
	assert.Equal(t, "PBS Cache client is not configured", string(d))
}

func TestVastUrlShouldReturnExpectedUrl(t *testing.T) {
	url := GetVastUrlTracking("http://external-url", "bidId", "bidder", "accountId", 1000)
	assert.Equal(t, "http://external-url/event?t=imp&b=bidId&a=accountId&bidder=bidder&f=b&ts=1000", url, "Invalid vast url")
}

func getValidVTrackRequestBody(withImpression bool, withContent bool) (string, error) {
	d, e := getVTrackRequestData(withImpression, withContent)

	if e != nil {
		return "", e
	}

	req := &BidCacheRequest{
		Puts: []prebid_cache_client.Cacheable{
			{
				Type:       prebid_cache_client.TypeXML,
				BidID:      "bidId1",
				Bidder:     "bidder",
				Data:       d,
				TTLSeconds: 3600,
				Timestamp:  1000,
			},
			{
				Type:       prebid_cache_client.TypeXML,
				BidID:      "bidId2",
				Bidder:     "updatable_bidder",
				Data:       d,
				TTLSeconds: 3600,
				Timestamp:  1000,
			},
		},
	}

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	e = enc.Encode(req)

	return buf.String(), e
}

func getVTrackRequestData(wi bool, wic bool) (db []byte, e error) {
	data := &bytes.Buffer{}
	enc := json.NewEncoder(data)
	enc.SetEscapeHTML(false)

	if wi && wic {
		e = enc.Encode(vastXmlWithImpressionWithContent)
		return data.Bytes(), e
	} else if wi {
		e = enc.Encode(vastXmlWithImpressionWithoutContent)
	} else {
		enc.Encode(vastXmlWithoutImpression)
	}

	return data.Bytes(), e
}

func TestInjectVideoEventTrackers(t *testing.T) {
	type args struct {
		externalURL string
		bid         *openrtb.Bid
		req         *openrtb.BidRequest
		callBack    func(string, *openrtb.BidRequest, string, *openrtb.Bid) map[string]string
	}
	type want struct {
		eventURLs map[string][]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "linear_creative",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[APPBUNDLE]",
				callBack: func(event string, req *openrtb.BidRequest, bidder string, bid *openrtb.Bid) map[string]string {
					companyEventIDMap := map[string]string{
						"midpoint":      "1003",
						"firstQuartile": "1004",
						"thirdQuartile": "1005",
						"complete":      "1006",
					}
					return map[string]string{
						"[EVENT_ID]": companyEventIDMap[event], // overrides PBS default value
					}
				},
				bid: &openrtb.Bid{
					AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative>
					                              <Linear>                      
					                                      <TrackingEvents>
					                                              <Tracking event="firstQuartile"><![CDATA[http://example.com/tracking/firstQuartile?k1=v1&k2=v2]]></Tracking>
					                                              <Tracking event="midpoint">http://example.com/tracking/midpoint</Tracking>
					                                              <Tracking event="thirdQuartile">http://example.com/tracking/thirdQuartile</Tracking>
					                                              <Tracking event="complete">http://example.com/tracking/complete</Tracking>
					                                      </TrackingEvents>
					                              </Linear>
					                     </Creative></Creatives></InLine></Ad></VAST>`,
				},
				req: &openrtb.BidRequest{App: &openrtb.App{Bundle: "abc"}},
			},
			want: want{
				eventURLs: map[string][]string{
					// "firstQuartile": {"http://example.com/tracking/firstQuartile?k1=v1&k2=v2", "http://company.tracker.com?eventId=1004&appbundle=abc"},
					// "midpoint":      {"http://example.com/tracking/midpoint", "http://company.tracker.com?eventId=1003&appbundle=abc"},
					// "thirdQuartile": {"http://example.com/tracking/thirdQuartile", "http://company.tracker.com?eventId=1005&appbundle=abc"},
					// "complete":      {"http://example.com/tracking/complete", "http://company.tracker.com?eventId=1006&appbundle=abc"},
					"firstQuartile": {"http://example.com/tracking/firstQuartile?k1=v1&k2=v2", "http://company.tracker.com?eventId=4&appbundle=abc"},
					"midpoint":      {"http://example.com/tracking/midpoint", "http://company.tracker.com?eventId=3&appbundle=abc"},
					"thirdQuartile": {"http://example.com/tracking/thirdQuartile", "http://company.tracker.com?eventId=5&appbundle=abc"},
					"complete":      {"http://example.com/tracking/complete", "http://company.tracker.com?eventId=6&appbundle=abc"},
				},
			},
		},
		{
			name: "non_linear_creative",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[APPBUNDLE]",
				callBack: func(event string, req *openrtb.BidRequest, bidder string, bid *openrtb.Bid) map[string]string {
					companyEventIDMap := map[string]string{
						"midpoint":      "1003",
						"firstQuartile": "1004",
						"thirdQuartile": "1005",
						"complete":      "1006",
					}
					return map[string]string{
						"[EVENT_ID]": companyEventIDMap[event], // overrides PBS default value
					}
				},
				bid: &openrtb.Bid{ // Adm contains to TrackingEvents tag
					AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative>
				<NonLinearAds>
					<TrackingEvents>
					<Tracking event="firstQuartile">http://something.com</Tracking>
					</TrackingEvents>
				</NonLinearAds>
			</Creative></Creatives></InLine></Ad></VAST>`,
				},
				req: &openrtb.BidRequest{App: &openrtb.App{Bundle: "abc"}},
			},
			want: want{
				eventURLs: map[string][]string{
					// "firstQuartile": {"http://something.com", "http://company.tracker.com?eventId=1004&appbundle=abc"},
					// "midpoint":      {"http://company.tracker.com?eventId=1003&appbundle=abc"},
					// "thirdQuartile": {"http://company.tracker.com?eventId=1005&appbundle=abc"},
					// "complete":      {"http://company.tracker.com?eventId=1006&appbundle=abc"},
					"firstQuartile": {"http://something.com", "http://company.tracker.com?eventId=4&appbundle=abc"},
					"midpoint":      {"http://company.tracker.com?eventId=3&appbundle=abc"},
					"thirdQuartile": {"http://company.tracker.com?eventId=5&appbundle=abc"},
					"complete":      {"http://company.tracker.com?eventId=6&appbundle=abc"},
				},
			},
		}, {
			name: "no_traker_url_configured", // expect no injection
			args: args{
				externalURL: "",
				bid: &openrtb.Bid{ // Adm contains to TrackingEvents tag
					AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative>
				<Linear>                      
				</Linear>
			</Creative></Creatives></InLine></Ad></VAST>`,
				},
				req: &openrtb.BidRequest{App: &openrtb.App{Bundle: "abc"}},
			},
			want: want{
				eventURLs: map[string][]string{},
			},
		},
		{
			name: "wrapper_vast_xml_from_partner", // expect we are injecting trackers inside wrapper
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[APPBUNDLE]",
				bid: &openrtb.Bid{ // Adm contains to TrackingEvents tag
					AdM: `<VAST version="4.2" xmlns="http://www.iab.com/VAST">
					<Ad id="20011" sequence="1" >
					  <Wrapper followAdditionalWrappers="0" allowMultipleAds="1" fallbackOnNoAd="0">
						<AdSystem version="4.0">iabtechlab</AdSystem>
					  <VASTAdTagURI>http://somevasturl</VASTAdTagURI>
						<Impression id="Impression-ID"><![CDATA[https://example.com/track/impression]]></Impression>
						<Creatives>
						  <Creative id="5480" sequence="1" adId="2447226">
							 <Linear></Linear>
						 </Creative>
				  </Creatives></Wrapper></Ad></VAST>`,
				},
				req: &openrtb.BidRequest{App: &openrtb.App{Bundle: "abc"}},
			},
			want: want{
				eventURLs: map[string][]string{
					// "firstQuartile": {"http://company.tracker.com?eventId=firstQuartile&appbundle=abc"},
					// "midpoint":      {"http://company.tracker.com?eventId=midpoint&appbundle=abc"},
					// "thirdQuartile": {"http://company.tracker.com?eventId=thirdQuartile&appbundle=abc"},
					// "complete":      {"http://company.tracker.com?eventId=complete&appbundle=abc"},
					"firstQuartile": {"http://company.tracker.com?eventId=4&appbundle=abc"},
					"midpoint":      {"http://company.tracker.com?eventId=3&appbundle=abc"},
					"thirdQuartile": {"http://company.tracker.com?eventId=5&appbundle=abc"},
					"complete":      {"http://company.tracker.com?eventId=6&appbundle=abc"},
				},
			},
		},
		// {
		// 	name: "vast_tag_uri_response_from_partner",
		// 	args: args{
		// 		externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[APPBUNDLE]",
		// 		bid: &openrtb.Bid{ // Adm contains to TrackingEvents tag
		// 			AdM: `<![CDATA[http://hostedvasttag.url&k=v]]>`,
		// 		},
		// 		req: &openrtb.BidRequest{App: &openrtb.App{Bundle: "abc"}},
		// 	},
		// 	want: want{
		// 		eventURLs: map[string][]string{
		// 			"firstQuartile": {"http://company.tracker.com?eventId=firstQuartile&appbundle=abc"},
		// 			"midpoint":      {"http://company.tracker.com?eventId=midpoint&appbundle=abc"},
		// 			"thirdQuartile": {"http://company.tracker.com?eventId=thirdQuartile&appbundle=abc"},
		// 			"complete":      {"http://company.tracker.com?eventId=complete&appbundle=abc"},
		// 		},
		// 	},
		// },
		// {
		// 	name: "adm_empty",
		// 	args: args{
		// 		externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[APPBUNDLE]",
		// 		bid: &openrtb.Bid{ // Adm contains to TrackingEvents tag
		// 			AdM:  "",
		// 			NURL: "nurl_contents",
		// 		},
		// 		req: &openrtb.BidRequest{App: &openrtb.App{Bundle: "abc"}},
		// 	},
		// 	want: want{
		// 		eventURLs: map[string][]string{
		// 			"firstQuartile": {"http://company.tracker.com?eventId=firstQuartile&appbundle=abc"},
		// 			"midpoint":      {"http://company.tracker.com?eventId=midpoint&appbundle=abc"},
		// 			"thirdQuartile": {"http://company.tracker.com?eventId=thirdQuartile&appbundle=abc"},
		// 			"complete":      {"http://company.tracker.com?eventId=complete&appbundle=abc"},
		// 		},
		// 	},
		// },
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// if tc.name != "vast_tag_uri_response_from_partner" {
			// 	return
			// }
			vast := ""
			if nil != tc.args.bid {
				vast = tc.args.bid.AdM // original vast
			}
			// if nil == tc.args.callBack {
			// 	config.TrackerMacros = nil
			// } else {
			// 	config.TrackerMacros = tc.args.callBack
			// }

			// bind this bid id with imp object
			tc.args.req.Imp = []openrtb.Imp{{ID: "123", Video: &openrtb.Video{}}}
			tc.args.bid.ImpID = tc.args.req.Imp[0].ID
			accountID := ""
			timestamp := int64(0)
			biddername := "test_bidder"
			injectedVast, injected, ierr := InjectVideoEventTrackers(tc.args.externalURL, vast, tc.args.bid, biddername, accountID, timestamp, tc.args.req)

			if !injected {
				// expect no change in input vast if tracking events are not injected
				assert.Equal(t, vast, string(injectedVast))
				assert.NotNil(t, ierr)
			} else {
				assert.Nil(t, ierr)
			}
			actualVastDoc := etree.NewDocument()

			err := actualVastDoc.ReadFromBytes(injectedVast)
			if nil != err {
				assert.Fail(t, err.Error())
			}

			// fmt.Println(string(injectedVast))
			actualTrackingEvents := actualVastDoc.FindElements("VAST/Ad/InLine/Creatives/Creative/Linear/TrackingEvents/Tracking")
			actualTrackingEvents = append(actualTrackingEvents, actualVastDoc.FindElements("VAST/Ad/InLine/Creatives/Creative/NonLinearAds/TrackingEvents/Tracking")...)
			actualTrackingEvents = append(actualTrackingEvents, actualVastDoc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/Linear/TrackingEvents/Tracking")...)
			actualTrackingEvents = append(actualTrackingEvents, actualVastDoc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/NonLinearAds/TrackingEvents/Tracking")...)

			totalURLCount := 0
			for event, URLs := range tc.want.eventURLs {

				for _, expectedURL := range URLs {
					present := false
					for _, te := range actualTrackingEvents {
						if te.SelectAttr("event").Value == event && te.Text() == expectedURL {
							present = true
							totalURLCount++
							break // expected URL present. check for next expected URL
						}
					}
					if !present {
						assert.Fail(t, "Expected tracker URL '"+expectedURL+"' is not present")
					}
				}
			}
			// ensure all total of events are injected
			assert.Equal(t, totalURLCount, len(actualTrackingEvents), fmt.Sprintf("Expected '%v' event trackers. But found '%v'", len(tc.want.eventURLs), len(actualTrackingEvents)))

		})
	}
}

func TestGetVideoEventTracking(t *testing.T) {
	type args struct {
		trackerURL string
		bid        *openrtb.Bid
		bidder     string
		accountId  string
		timestamp  int64
		req        *openrtb.BidRequest
		doc        *etree.Document
		callBack   func(string, *openrtb.BidRequest, string, *openrtb.Bid) map[string]string
	}
	type want struct {
		trackerURLMap map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid_scenario",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[APPBUNDLE]",
				bid:        &openrtb.Bid{
					// AdM: vastXMLWith2Creatives,
				},
				req: &openrtb.BidRequest{
					App: &openrtb.App{
						Bundle: "someappbundle",
					},
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					// "firstQuartile": "http://company.tracker.com?eventId=firstQuartile&appbundle=someappbundle",
					// "midpoint":      "http://company.tracker.com?eventId=midpoint&appbundle=someappbundle",
					// "thirdQuartile": "http://company.tracker.com?eventId=thirdQuartile&appbundle=someappbundle",
					// "complete":      "http://company.tracker.com?eventId=complete&appbundle=someappbundle"},
					"firstQuartile": "http://company.tracker.com?eventId=4&appbundle=someappbundle",
					"midpoint":      "http://company.tracker.com?eventId=3&appbundle=someappbundle",
					"thirdQuartile": "http://company.tracker.com?eventId=5&appbundle=someappbundle",
					"complete":      "http://company.tracker.com?eventId=6&appbundle=someappbundle"},
			},
		},
		{
			name: "no_macro_value", // expect no replacement
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[APPBUNDLE]",
				bid:        &openrtb.Bid{},
				req: &openrtb.BidRequest{
					App: &openrtb.App{}, // no app bundle value
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					// "firstQuartile": "http://company.tracker.com?eventId=firstQuartile&appbundle=[APPBUNDLE]",
					// "midpoint":      "http://company.tracker.com?eventId=midpoint&appbundle=[APPBUNDLE]",
					// "thirdQuartile": "http://company.tracker.com?eventId=thirdQuartile&appbundle=[APPBUNDLE]",
					// "complete":      "http://company.tracker.com?eventId=complete&appbundle=[APPBUNDLE]"},
					"firstQuartile": "http://company.tracker.com?eventId=4&appbundle=[APPBUNDLE]",
					"midpoint":      "http://company.tracker.com?eventId=3&appbundle=[APPBUNDLE]",
					"thirdQuartile": "http://company.tracker.com?eventId=5&appbundle=[APPBUNDLE]",
					"complete":      "http://company.tracker.com?eventId=6&appbundle=[APPBUNDLE]"},
			},
		},
		{
			name: "replace_using_callback_fuction",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&param1=[company_custom_macro]",
				callBack: func(string, *openrtb.BidRequest, string, *openrtb.Bid) map[string]string {
					return map[string]string{
						"[company_custom_macro]": "macro_value",
					}
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					// "firstQuartile": "http://company.tracker.com?eventId=firstQuartile&param1=macro_value",
					// "midpoint":      "http://company.tracker.com?eventId=midpoint&param1=macro_value",
					// "thirdQuartile": "http://company.tracker.com?eventId=thirdQuartile&param1=macro_value",
					// "complete":      "http://company.tracker.com?eventId=complete&param1=macro_value"},
					"firstQuartile": "http://company.tracker.com?eventId=4&param1=[company_custom_macro]",
					"midpoint":      "http://company.tracker.com?eventId=3&param1=[company_custom_macro]",
					"thirdQuartile": "http://company.tracker.com?eventId=5&param1=[company_custom_macro]",
					"complete":      "http://company.tracker.com?eventId=6&param1=[company_custom_macro]"},
			},
		}, {
			name: "prefer_company_value_for_standard_macro",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[APPBUNDLE]",
				callBack: func(string, *openrtb.BidRequest, string, *openrtb.Bid) map[string]string {
					return map[string]string{
						"[APPBUNDLE]": "my_custom_value", // expect this value for macro
					}
				},
				req: &openrtb.BidRequest{
					App: &openrtb.App{
						Bundle: "myapp", // do not expect this value
					},
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					// "firstQuartile": "http://company.tracker.com?eventId=firstQuartile&appbundle=my_custom_value",
					// "midpoint":      "http://company.tracker.com?eventId=midpoint&appbundle=my_custom_value",
					// "thirdQuartile": "http://company.tracker.com?eventId=thirdQuartile&appbundle=my_custom_value",
					// "complete":      "http://company.tracker.com?eventId=complete&appbundle=my_custom_value"},
					"firstQuartile": "http://company.tracker.com?eventId=4&appbundle=myapp",
					"midpoint":      "http://company.tracker.com?eventId=3&appbundle=myapp",
					"thirdQuartile": "http://company.tracker.com?eventId=5&appbundle=myapp",
					"complete":      "http://company.tracker.com?eventId=6&appbundle=myapp"},
			},
		}, {
			name: "multireplace_macro",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[APPBUNDLE]&parameter2=[APPBUNDLE]",
				req: &openrtb.BidRequest{
					App: &openrtb.App{
						Bundle: "myapp123",
					},
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					// "firstQuartile": "http://company.tracker.com?eventId=firstQuartile&appbundle=myapp123&parameter2=myapp123",
					// "midpoint":      "http://company.tracker.com?eventId=midpoint&appbundle=myapp123&parameter2=myapp123",
					// "thirdQuartile": "http://company.tracker.com?eventId=thirdQuartile&appbundle=myapp123&parameter2=myapp123",
					// "complete":      "http://company.tracker.com?eventId=complete&appbundle=myapp123&parameter2=myapp123"},
					"firstQuartile": "http://company.tracker.com?eventId=4&appbundle=myapp123&parameter2=myapp123",
					"midpoint":      "http://company.tracker.com?eventId=3&appbundle=myapp123&parameter2=myapp123",
					"thirdQuartile": "http://company.tracker.com?eventId=5&appbundle=myapp123&parameter2=myapp123",
					"complete":      "http://company.tracker.com?eventId=6&appbundle=myapp123&parameter2=myapp123"},
			},
		},
		{
			name: "callback_with_macro_without_prefix_and_suffix",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&param1=[CUSTOM_MACRO]",
				callBack: func(string, *openrtb.BidRequest, string, *openrtb.Bid) map[string]string {
					return map[string]string{
						"CUSTOM_MACRO": "my_custom_value", // invalid macro syntax missing [ and ]
					}
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					// "firstQuartile": "http://company.tracker.com?eventId=firstQuartile&param1=[CUSTOM_MACRO]",
					// "midpoint":      "http://company.tracker.com?eventId=midpoint&param1=[CUSTOM_MACRO]",
					// "thirdQuartile": "http://company.tracker.com?eventId=thirdQuartile&param1=[CUSTOM_MACRO]",
					// "complete":      "http://company.tracker.com?eventId=complete&param1=[CUSTOM_MACRO]"},
					"firstQuartile": "http://company.tracker.com?eventId=4&param1=[CUSTOM_MACRO]",
					"midpoint":      "http://company.tracker.com?eventId=3&param1=[CUSTOM_MACRO]",
					"thirdQuartile": "http://company.tracker.com?eventId=5&param1=[CUSTOM_MACRO]",
					"complete":      "http://company.tracker.com?eventId=6&param1=[CUSTOM_MACRO]"},
			},
		},
		{
			name: "callback_with_empty_macro",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&param1=[CUSTOM_MACRO]",
				callBack: func(string, *openrtb.BidRequest, string, *openrtb.Bid) map[string]string {
					return map[string]string{
						"": "my_custom_value", // invalid macro .. its empty
					}
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					// "firstQuartile": "http://company.tracker.com?eventId=firstQuartile&param1=[CUSTOM_MACRO]",
					// "midpoint":      "http://company.tracker.com?eventId=midpoint&param1=[CUSTOM_MACRO]",
					// "thirdQuartile": "http://company.tracker.com?eventId=thirdQuartile&param1=[CUSTOM_MACRO]",
					// "complete":      "http://company.tracker.com?eventId=complete&param1=[CUSTOM_MACRO]"},
					"firstQuartile": "http://company.tracker.com?eventId=4&param1=[CUSTOM_MACRO]",
					"midpoint":      "http://company.tracker.com?eventId=3&param1=[CUSTOM_MACRO]",
					"thirdQuartile": "http://company.tracker.com?eventId=5&param1=[CUSTOM_MACRO]",
					"complete":      "http://company.tracker.com?eventId=6&param1=[CUSTOM_MACRO]"},
			},
		},
		{
			name: "macro_is_case_sensitive",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&param1=[CUSTOM_MACRO]",
				callBack: func(string, *openrtb.BidRequest, string, *openrtb.Bid) map[string]string {
					return map[string]string{
						"[custom_MACRO]": "my_custom_value", // case sensitivity fail w.r.t. trackerURL
					}
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					// "firstQuartile": "http://company.tracker.com?eventId=firstQuartile&param1=[CUSTOM_MACRO]",
					// "midpoint":      "http://company.tracker.com?eventId=midpoint&param1=[CUSTOM_MACRO]",
					// "thirdQuartile": "http://company.tracker.com?eventId=thirdQuartile&param1=[CUSTOM_MACRO]",
					// "complete":      "http://company.tracker.com?eventId=complete&param1=[CUSTOM_MACRO]"},
					"firstQuartile": "http://company.tracker.com?eventId=4&param1=[CUSTOM_MACRO]",
					"midpoint":      "http://company.tracker.com?eventId=3&param1=[CUSTOM_MACRO]",
					"thirdQuartile": "http://company.tracker.com?eventId=5&param1=[CUSTOM_MACRO]",
					"complete":      "http://company.tracker.com?eventId=6&param1=[CUSTOM_MACRO]"},
			},
		},
		{
			name: "company_specific_event_id_for_EVENT_ID_macro",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]",
				callBack: func(event string, req *openrtb.BidRequest, bidder string, bid *openrtb.Bid) map[string]string {
					companyEventIDMap := map[string]string{
						"midpoint":      "1003",
						"firstQuartile": "1004",
						"thirdQuartile": "1005",
						"complete":      "1006",
					}
					return map[string]string{
						"[EVENT_ID]": companyEventIDMap[event],
					}
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					// "firstQuartile": "http://company.tracker.com?eventId=1004",
					// "midpoint":      "http://company.tracker.com?eventId=1003",
					// "thirdQuartile": "http://company.tracker.com?eventId=1005",
					// "complete":      "http://company.tracker.com?eventId=1006"},
					"firstQuartile": "http://company.tracker.com?eventId=4",
					"midpoint":      "http://company.tracker.com?eventId=3",
					"thirdQuartile": "http://company.tracker.com?eventId=5",
					"complete":      "http://company.tracker.com?eventId=6"},
			},
		},
		{
			name: "company_specific_event_id_macro",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[COMPANY_EVENT_ID]",
				callBack: func(event string, req *openrtb.BidRequest, bidder string, bid *openrtb.Bid) map[string]string {
					companyEventIDMap := map[string]string{
						"midpoint":      "1003",
						"firstQuartile": "1004",
						"thirdQuartile": "1005",
						"complete":      "1006",
					}
					return map[string]string{
						"[COMPANY_EVENT_ID]": companyEventIDMap[event],
					}
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					// "firstQuartile": "http://company.tracker.com?eventId=1004",
					// "midpoint":      "http://company.tracker.com?eventId=1003",
					// "thirdQuartile": "http://company.tracker.com?eventId=1005",
					// "complete":      "http://company.tracker.com?eventId=1006"},
					"firstQuartile": "http://company.tracker.com?eventId=[COMPANY_EVENT_ID]",
					"midpoint":      "http://company.tracker.com?eventId=[COMPANY_EVENT_ID]",
					"thirdQuartile": "http://company.tracker.com?eventId=[COMPANY_EVENT_ID]",
					"complete":      "http://company.tracker.com?eventId=[COMPANY_EVENT_ID]"},
			},
		}, {
			name: "empty_tracker_url",
			args: args{trackerURL: "    "},
			want: want{trackerURLMap: make(map[string]string)},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// if tc.name != "macro_is_case_sensitive" {
			// 	return
			// }
			// assign callback function if present
			// if nil == tc.args.callBack {
			// 	config.TrackerMacros = nil
			// } else {
			// 	config.TrackerMacros = tc.args.callBack
			// }
			eventURLMap := GetVideoEventTracking(tc.args.trackerURL, tc.args.bid, tc.args.bidder, tc.args.accountId, tc.args.timestamp, tc.args.req, tc.args.doc)
			assert.True(t, reflect.DeepEqual(tc.want.trackerURLMap, eventURLMap), "tracker url map comparison failed")
		})
	}
}

func TestReplaceMacro(t *testing.T) {
	type args struct {
		trackerURL string
		macro      string
		value      string
	}
	type want struct {
		trackerURL string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "empty_tracker_url", args: args{trackerURL: "", macro: "[TEST]", value: "testme"}, want: want{trackerURL: ""}},
		{name: "tracker_url_with_macro", args: args{trackerURL: "http://something.com?test=[TEST]", macro: "[TEST]", value: "testme"}, want: want{trackerURL: "http://something.com?test=testme"}},
		{name: "tracker_url_with_invalid_macro", args: args{trackerURL: "http://something.com?test=TEST]", macro: "[TEST]", value: "testme"}, want: want{trackerURL: "http://something.com?test=TEST]"}},
		{name: "tracker_url_with_repeating_macro", args: args{trackerURL: "http://something.com?test=[TEST]&test1=[TEST]", macro: "[TEST]", value: "testme"}, want: want{trackerURL: "http://something.com?test=testme&test1=testme"}},
		{name: "empty_macro", args: args{trackerURL: "http://something.com?test=[TEST]", macro: "", value: "testme"}, want: want{trackerURL: "http://something.com?test=[TEST]"}},
		{name: "macro_without_[", args: args{trackerURL: "http://something.com?test=[TEST]", macro: "TEST]", value: "testme"}, want: want{trackerURL: "http://something.com?test=[TEST]"}},
		{name: "macro_without_]", args: args{trackerURL: "http://something.com?test=[TEST]", macro: "[TEST", value: "testme"}, want: want{trackerURL: "http://something.com?test=[TEST]"}},
		{name: "empty_value", args: args{trackerURL: "http://something.com?test=[TEST]", macro: "[TEST]", value: ""}, want: want{trackerURL: "http://something.com?test=[TEST]"}},
		{name: "nested_macro_value", args: args{trackerURL: "http://something.com?test=[TEST]", macro: "[TEST]", value: "[TEST][TEST]"}, want: want{trackerURL: "http://something.com?test=%5BTEST%5D%5BTEST%5D"}},
		{name: "url_as_macro_value", args: args{trackerURL: "http://something.com?test=[TEST]", macro: "[TEST]", value: "http://iamurl.com"}, want: want{trackerURL: "http://something.com?test=http%3A%2F%2Fiamurl.com"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			trackerURL := replaceMacro(tc.args.trackerURL, tc.args.macro, tc.args.value)
			assert.Equal(t, tc.want.trackerURL, trackerURL)
		})
	}

}
