package openrtb2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters"
	analyticsBuild "github.com/prebid/prebid-server/v3/analytics/build"
	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/currency"
	"github.com/prebid/prebid-server/v3/errortypes"
	"github.com/prebid/prebid-server/v3/exchange"
	"github.com/prebid/prebid-server/v3/experiment/adscert"
	"github.com/prebid/prebid-server/v3/floors"
	"github.com/prebid/prebid-server/v3/hooks"
	"github.com/prebid/prebid-server/v3/macros"
	"github.com/prebid/prebid-server/v3/metrics"
	metricsConfig "github.com/prebid/prebid-server/v3/metrics/config"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/ortb"
	"github.com/prebid/prebid-server/v3/stored_requests"
	"github.com/prebid/prebid-server/v3/stored_requests/backends/empty_fetcher"
)

type ctvtestCase struct {
	Description             string               `json:"description"`
	Config                  *ctvtestConfigValues `json:"config"`
	BidRequest              json.RawMessage      `json:"mockBidRequest"`
	ExpectedValidatedBidReq json.RawMessage      `json:"expectedValidatedBidRequest"`
	ExpectedReturnCode      int                  `json:"expectedReturnCode,omitempty"`
	ExpectedErrorMessage    string               `json:"expectedErrorMessage"`
	Query                   string               `json:"query"`
	planBuilder             hooks.ExecutionPlanBuilder
	ExpectedBidResponse     json.RawMessage `json:"expectedBidResponse"`
}

type ctvtestConfigValues struct {
	AccountRequired     bool                          `json:"accountRequired"`
	AliasJSON           string                        `json:"aliases"`
	BlockedApps         []string                      `json:"blockedApps"`
	DisabledAdapters    []string                      `json:"disabledAdapters"`
	CurrencyRates       map[string]map[string]float64 `json:"currencyRates"`
	MockBidders         []ctvMockBidderHandler        `json:"mockBidders"`
	RealParamsValidator bool                          `json:"realParamsValidator"`
	AssertBidExt        bool                          `json:"assertbidext"`
}

func (tc *ctvtestConfigValues) getBlockedAppsLookup() map[string]bool {
	var blockedAppsLookup map[string]bool

	if len(tc.BlockedApps) > 0 {
		blockedAppsLookup = make(map[string]bool, len(tc.BlockedApps))
		for _, app := range tc.BlockedApps {
			blockedAppsLookup[app] = true
		}
	}
	return blockedAppsLookup
}

type ctvMockBidderHandler struct {
	BidderName string    `json:"bidderName"`
	Currency   string    `json:"currency"`
	Bids       []mockBid `json:"bids"`
}

type mockBid struct {
	ImpId   string   `json:"impid"`
	Price   float64  `json:"price"`
	DealID  string   `json:"dealid,omitempty"`
	Cat     []string `json:"cat,omitempty"`
	ADomain []string `json:"adomain,omitempty"`
}

func (b ctvMockBidderHandler) bid(w http.ResponseWriter, req *http.Request) {
	// Read request Body
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)

	// Unmarshal exit if error
	var openrtb2Request openrtb2.BidRequest
	if err := json.Unmarshal(buf.Bytes(), &openrtb2Request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var openrtb2ImpExt map[string]json.RawMessage
	if err := json.Unmarshal(openrtb2Request.Imp[0].Ext, &openrtb2ImpExt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, exists := openrtb2ImpExt["bidder"]
	if !exists {
		http.Error(w, "This request is not meant for this bidder", http.StatusBadRequest)
		return
	}

	var bids []openrtb2.Bid
	var seq int
	for _, bid := range b.Bids {
		bids = append(bids, openrtb2.Bid{
			ID:      b.BidderName + "-bid-" + strconv.Itoa(seq),
			ImpID:   bid.ImpId,
			Price:   bid.Price,
			AdM:     "<VAST version=\"3.0\"><Ad id=\"601364\"><InLine><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[https://Errortrack.com?p=1234&er=[ERRORCODE]]]></Error><Creatives><Creative AdID=\"601364\"><Linear skipoffset=\"70%\"><Duration><![CDATA[00:00:04]]></Duration><MediaFiles><MediaFile delivery=\"progressive\" type=\"video/mp4\" bitrate=\"500\" width=\"400\" height=\"300\" scalable=\"true\" maintainAspectRatio=\"true\"><![CDATA[https://owsdk-stagingams.pubmatic.com:8443/openwrap/media/pubmatic.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives><Pricing model=\"CPM\" currency=\"USD\"><![CDATA[18]]></Pricing></InLine></Ad></VAST>",
			DealID:  bid.DealID,
			Cat:     bid.Cat,
			ADomain: bid.ADomain,
		})
		seq++

	}

	// Create bid service openrtb2.BidResponse with one bid according to JSON test file values
	var serverResponseObject = openrtb2.BidResponse{
		ID:  openrtb2Request.ID,
		Cur: b.Currency,
		SeatBid: []openrtb2.SeatBid{
			{
				Bid:  bids,
				Seat: b.BidderName,
			},
		},
	}

	// Marshal the response and http write it
	serverJsonResponse, err := json.Marshal(&serverResponseObject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(serverJsonResponse)
	return
}

func ctvTestEndpoint(test ctvtestCase, cfg *config.Configuration) (httprouter.Handle, *exchangeTestWrapper, []*httptest.Server, *httptest.Server, error) {
	if test.Config == nil {
		test.Config = &ctvtestConfigValues{}
	}

	var paramValidator openrtb_ext.BidderParamValidator
	if test.Config.RealParamsValidator {
		var err error
		paramValidator, err = openrtb_ext.NewBidderParamsValidator("../../static/bidder-params")
		if err != nil {
			return nil, nil, nil, nil, err
		}
	} else {
		paramValidator = mockBidderParamValidator{}
	}

	bidderInfos := getBidderInfos(test.Config.DisabledAdapters, openrtb_ext.CoreBidderNames())
	bidderMap := exchange.GetActiveBidders(bidderInfos)
	disabledBidders := exchange.GetDisabledBidderWarningMessages(bidderInfos)
	met := &metricsConfig.NilMetricsEngine{}
	mockFetcher := empty_fetcher.EmptyFetcher{}

	// Adapter map with mock adapters needed to run JSON test cases
	adapterMap := make(map[openrtb_ext.BidderName]exchange.AdaptedBidder, 0)
	mockBidServersArray := make([]*httptest.Server, 0, 3)

	// Mock prebid Server's currency converter, instantiate and start
	mockCurrencyConversionService := mockCurrencyRatesClient{
		currencyInfo{
			Conversions: test.Config.CurrencyRates,
		},
	}
	mockCurrencyRatesServer := httptest.NewServer(http.HandlerFunc(mockCurrencyConversionService.handle))

	requestValidator := ortb.NewRequestValidator(bidderMap, disabledBidders, paramValidator)
	testExchange, mockBidServersArray := testCTVExchange(test.Config, adapterMap, mockBidServersArray, mockCurrencyRatesServer, bidderInfos, cfg, met, mockFetcher, requestValidator)

	planBuilder := test.planBuilder
	if planBuilder == nil {
		planBuilder = hooks.EmptyPlanBuilder{}
	}

	endpoint, err := NewCTVEndpoint(
		testExchange,
		requestValidator,
		&mockStoredReqFetcher{},
		&mockStoredReqFetcher{},
		&mockAccountFetcher{},
		cfg,
		met,
		analyticsBuild.New(&config.Analytics{}),
		disabledBidders,
		[]byte(test.Config.AliasJSON),
		bidderMap,
		planBuilder,
		nil,
	)

	return endpoint, testExchange.(*exchangeTestWrapper), mockBidServersArray, mockCurrencyRatesServer, err
}

func testCTVExchange(testCfg *ctvtestConfigValues, adapterMap map[openrtb_ext.BidderName]exchange.AdaptedBidder, mockBidServersArray []*httptest.Server, mockCurrencyRatesServer *httptest.Server, bidderInfos config.BidderInfos, cfg *config.Configuration, met metrics.MetricsEngine, mockFetcher stored_requests.CategoryFetcher, paramValidator ortb.RequestValidator) (exchange.Exchange, []*httptest.Server) {
	if len(testCfg.MockBidders) == 0 {
		testCfg.MockBidders = append(testCfg.MockBidders, ctvMockBidderHandler{BidderName: "pubmatic", Currency: "USD", Bids: []mockBid{
			{
				ImpId: "imp1",
				Price: 0,
			},
		}})
	}
	for _, mockBidder := range testCfg.MockBidders {
		bidServer := httptest.NewServer(http.HandlerFunc(mockBidder.bid))
		bidderAdapter := ctvMockAdapter{mockServerURL: bidServer.URL}
		bidderName := openrtb_ext.BidderName(mockBidder.BidderName)

		adapterMap[bidderName] = exchange.AdaptBidder(bidderAdapter, bidServer.Client(), &config.Configuration{}, &metricsConfig.NilMetricsEngine{}, bidderName, nil, "")
		mockBidServersArray = append(mockBidServersArray, bidServer)
	}

	mockCurrencyConverter := currency.NewRateConverter(mockCurrencyRatesServer.Client(), mockCurrencyRatesServer.URL, time.Second)
	mockCurrencyConverter.Run()

	gdprPermsBuilder := fakePermissionsBuilder{
		permissions: &fakePermissions{},
	}.Builder

	testExchange := exchange.NewExchange(adapterMap,
		&wellBehavedCache{},
		cfg,
		paramValidator,
		nil,
		met,
		bidderInfos,
		gdprPermsBuilder,
		mockCurrencyConverter,
		mockFetcher,
		&adscert.NilSigner{},
		macros.NewStringIndexBasedReplacer(),
		&floors.PriceFloorFetcher{},
	)

	testExchange = &exchangeTestWrapper{
		ex: testExchange,
	}

	return testExchange, mockBidServersArray
}

type ctvMockAdapter struct {
	mockServerURL string
	Server        config.Server
}

func CTVBuilder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	adapter := &ctvMockAdapter{
		mockServerURL: config.Endpoint,
		Server:        server,
	}
	return adapter, nil
}

func (a ctvMockAdapter) MakeRequests(request *openrtb2.BidRequest, requestInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	var requests []*adapters.RequestData
	var errors []error

	requestJSON, err := json.Marshal(request)
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	requestData := &adapters.RequestData{
		Method: "POST",
		Uri:    a.mockServerURL,
		Body:   requestJSON,
	}
	requests = append(requests, requestData)
	return requests, errors
}

func (a ctvMockAdapter) MakeBids(request *openrtb2.BidRequest, requestData *adapters.RequestData, responseData *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if responseData.StatusCode != http.StatusOK {
		switch responseData.StatusCode {
		case http.StatusNoContent:
			return nil, nil
		case http.StatusBadRequest:
			return nil, []error{&errortypes.BadInput{
				Message: "Unexpected status code: 400. Bad request from publisher. Run with request.debug = 1 for more info.",
			}}
		default:
			return nil, []error{&errortypes.BadServerResponse{
				Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info.", responseData.StatusCode),
			}}
		}
	}

	var publisherResponse openrtb2.BidResponse
	if err := json.Unmarshal(responseData.Body, &publisherResponse); err != nil {
		return nil, []error{err}
	}

	rv := adapters.NewBidderResponseWithBidsCapacity(len(request.Imp))
	rv.Currency = publisherResponse.Cur
	for _, seatBid := range publisherResponse.SeatBid {
		for i, bid := range seatBid.Bid {
			for _, imp := range request.Imp {
				if imp.ID == bid.ImpID {
					b := &adapters.TypedBid{
						Bid:     &seatBid.Bid[i],
						BidType: openrtb_ext.BidTypeVideo,
					}
					rv.Bids = append(rv.Bids, b)
				}
			}
		}
	}
	return rv, nil
}
