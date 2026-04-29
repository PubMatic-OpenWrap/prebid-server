package aps

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	adcom1 "github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	mock_metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPS(t *testing.T) {
	a := NewAPS(nil)
	require.NotNil(t, a)
	assert.Nil(t, a.metricsEngine)
}

func TestModifyRequestWithAPSParams(t *testing.T) {
	rctx := models.RequestCtx{}

	signalBR := &openrtb2.BidRequest{
		Imp: []openrtb2.Imp{{
			ID:                "si1",
			Instl:             1,
			DisplayManager:    "dm",
			DisplayManagerVer: "2.0.0",
			Ext:               json.RawMessage(`{"skadn":{"versions":["v1"]},"owsdk":{"x":1}}`),
		}},
		Device: &openrtb2.Device{UA: "Mozilla"},
		App:    &openrtb2.App{Name: "SignalApp"},
	}
	validSig := mustEncodeSignalBidRequest(t, signalBR)
	badSignal := base64.StdEncoding.EncodeToString([]byte(`not-json`))

	tests := []struct {
		name             string
		requestBody      []byte
		expectedResponse []byte
		expectedError    bool
		expectNilBody    bool
		metricsSetup     func(*mock_metrics.MockMetricsEngine)
	}{
		{
			name:             "empty_request_body",
			requestBody:      nil,
			expectedResponse: nil,
		},
		{
			name:             "invalid_json_returns_original_bytes",
			requestBody:      []byte(`{broken`),
			expectedResponse: []byte(`{broken`),
			expectedError:    true,
		},
		{
			name:             "static_data_sets_secure_clears_native_and_video_on_imp[0]",
			requestBody:      []byte(`{"id":"r1","imp":[{"id":"i1","tagid":"t1","secure":0,"banner":{"w":300,"h":250},"video":{"mimes":["video/mp4"]},"native":{"request":"n"}}],"app":{"publisher":{"id":"pub-9"}}}`),
			expectedResponse: []byte(`{"id":"r1","imp":[{"id":"i1","banner":{"w":300,"h":250},"tagid":"t1","secure":1}],"app":{"publisher":{"id":"pub-9"}}}`),
		},
		{
			name:             "reward_video_sets_rwdd_drops_banner_when_video.ext.videotype_is_rewarded",
			requestBody:      []byte(`{"id":"r1","imp":[{"id":"i1","tagid":"t1","banner":{"w":1,"h":1},"video":{"ext":{"videotype":"rewarded"}}}],"app":{"publisher":{"id":"pub"}}}`),
			expectedResponse: []byte(`{"id":"r1","imp":[{"id":"i1","tagid":"t1","secure":1,"rwdd":1}],"app":{"publisher":{"id":"pub"}}}`),
		},
		{
			name:             "removes_app.ext.sessionDepth",
			requestBody:      []byte(`{"id":"x","imp":[{"id":"i","tagid":"t"}],"app":{"ext":{"sessionDepth":3},"publisher":{"id":"p"}}}`),
			expectedResponse: []byte(`{"id":"x","imp":[{"id":"i","tagid":"t","secure":1}],"app":{"publisher":{"id":"p"},"ext":{}}}`),
		},
		{
			name:        "missing_signal_records_metric",
			requestBody: []byte(`{"id":"r1","imp":[{"id":"i1","tagid":"t"}],"app":{"publisher":{"id":"pub-1"}},"ext":{"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":42}}}}},"user":{"buyeruid":""}}`),
			metricsSetup: func(m *mock_metrics.MockMetricsEngine) {
				m.EXPECT().RecordSignalDataStatus("pub-1", "42", models.MissingSignal)
			},
			expectedResponse: []byte(`{"id":"r1","imp":[{"id":"i1","tagid":"t","secure":1}],"app":{"publisher":{"id":"pub-1"}},"user":{},"ext":{"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":42}}}}}}`),
		},
		{
			name:        "invalid_base64_signal_records_metric",
			requestBody: []byte(`{"id":"r1","imp":[{"id":"i1","tagid":"t"}],"app":{"publisher":{"id":"p"}},"ext":{"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":9}}}}},"user":{"buyeruid":"@@@not-base64@@@"}}`),
			metricsSetup: func(m *mock_metrics.MockMetricsEngine) {
				m.EXPECT().RecordSignalDataStatus("p", "9", models.InvalidSignal)
			},
			expectedResponse: []byte(`{"id":"r1","imp":[{"id":"i1","tagid":"t","secure":1}],"app":{"publisher":{"id":"p"}},"user":{"buyeruid":"@@@not-base64@@@"},"ext":{"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":9}}}}}}`),
		},
		{
			name:        "invalid_json_inside_signal_records_metric",
			requestBody: []byte(fmt.Sprintf(`{"id":"r1","imp":[{"id":"i1","tagid":"t"}],"app":{"publisher":{"id":"p"}},"ext":{"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":1}}}}},"user":{"buyeruid":%q}}`, badSignal)),
			metricsSetup: func(m *mock_metrics.MockMetricsEngine) {
				m.EXPECT().RecordSignalDataStatus("p", "1", models.InvalidSignal)
			},
			expectedResponse: []byte(fmt.Sprintf(`{"id":"r1","imp":[{"id":"i1","tagid":"t","secure":1}],"app":{"publisher":{"id":"p"}},"user":{"buyeruid":%q},"ext":{"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":1}}}}}}`, badSignal)),
		},
		{
			name:             "valid_signal_merges_impression_app_and_device_from_signal",
			requestBody:      []byte(fmt.Sprintf(`{"id":"base","imp":[{"id":"i1","tagid":"t1","ext":{}}],"app":{"publisher":{"id":"pubx"}},"device":{"ua":"orig"},"ext":{"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":100}}}}},"user":{"buyeruid":%q}}`, validSig)),
			expectedResponse: []byte(fmt.Sprintf(`{"id":"base","imp":[{"id":"i1","displaymanager":"dm","displaymanagerver":"2.0.0","instl":1,"tagid":"t1","secure":1,"ext":{"skadn":{"versions":["v1"]},"owsdk":{"x":1}}}],"app":{"name":"SignalApp","publisher":{"id":"pubx"}},"device":{"ua":"Mozilla"},"user":{"buyeruid":%q},"ext":{"prebid":{"bidderparams":{"pubmatic":{"wrapper":{"profileid":100}}}}}}`, validSig)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMetrics := mock_metrics.NewMockMetricsEngine(ctrl)
			if tt.metricsSetup != nil {
				tt.metricsSetup(mockMetrics)
			}

			a := NewAPS(mockMetrics)
			response := a.ModifyRequestWithAPSParams(tt.requestBody, rctx)
			if tt.expectedError {
				assert.Equal(t, tt.expectedResponse, response)
				return
			}

			if tt.expectedResponse == nil {
				assert.Empty(t, response)
			} else {
				assert.JSONEq(t, string(tt.expectedResponse), string(response))
			}
		})
	}
}

func TestModifyBanner(t *testing.T) {
	tests := []struct {
		name     string
		request  *openrtb2.Banner
		signal   *openrtb2.Banner
		expected *openrtb2.Banner
	}{
		{
			name:     "nil request is a no-op",
			request:  nil,
			signal:   &openrtb2.Banner{API: []adcom1.APIFramework{5}},
			expected: nil,
		},
		{
			name:     "nil signal is a no-op",
			request:  &openrtb2.Banner{W: ptrutil.ToPtr[int64](1)},
			signal:   nil,
			expected: &openrtb2.Banner{W: ptrutil.ToPtr[int64](1)},
		},
		{
			name:    "copies API frameworks from signal",
			request: &openrtb2.Banner{},
			signal:  &openrtb2.Banner{API: []adcom1.APIFramework{7}},
			expected: &openrtb2.Banner{
				API: []adcom1.APIFramework{7},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifyBanner(tt.request, tt.signal)
			assert.Equal(t, tt.expected, tt.request)
		})
	}
}

func TestUpdateImpExtension(t *testing.T) {
	tests := []struct {
		name             string
		reqExt           []byte
		sigExt           []byte
		expectedResponse string
	}{
		{
			name:             "nil signal returns request ext unchanged",
			reqExt:           []byte(`{"prebid":1}`),
			sigExt:           nil,
			expectedResponse: `{"prebid":1}`,
		},
		{
			name:             "empty request ext receives skadn and owsdk from signal",
			reqExt:           nil,
			sigExt:           []byte(`{"skadn":{"version":"2"},"owsdk":{"a":1}}`),
			expectedResponse: `{"skadn":{"version":"2"},"owsdk":{"a":1}}`,
		},
		{
			name:             "merges skadn paths and owsdk into existing ext",
			reqExt:           []byte(`{"foo":1}`),
			sigExt:           []byte(`{"skadn":{"skoverlay":true,"productpage":7},"owsdk":{"k":2}}`),
			expectedResponse: `{"foo":1,"owsdk":{"k":2},"skadn":{"productpage":7,"skoverlay":true}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := updateImpExtension(tt.reqExt, tt.sigExt)
			assert.JSONEq(t, tt.expectedResponse, string(out))
		})
	}
}

func TestUpdateRegs(t *testing.T) {
	tests := []struct {
		name     string
		req      *openrtb2.BidRequest
		sig      *openrtb2.Regs
		expected string
	}{
		{
			name:     "copies COPPA and reg ext paths from signal",
			req:      &openrtb2.BidRequest{},
			sig:      &openrtb2.Regs{COPPA: 1, Ext: json.RawMessage(`{"gdpr":1,"gpp":"x"}`)},
			expected: `{"coppa":1,"ext":{"gdpr":1,"gpp":"x"}}`,
		},
		{
			name: "nil signal leaves request unchanged",
			req: &openrtb2.BidRequest{
				Regs: &openrtb2.Regs{Ext: json.RawMessage(`{"keep":true}`)},
			},
			sig:      nil,
			expected: `{"ext":{"keep":true}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateRegs(tt.req, tt.sig)
			require.NotNil(t, tt.req.Regs)
			b, err := json.Marshal(tt.req.Regs)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(b))
		})
	}
}

func TestUpdateApp(t *testing.T) {
	req := &openrtb2.BidRequest{App: &openrtb2.App{Bundle: "b"}}
	sig := &openrtb2.App{
		Domain: "d.example",
		Cat:    []string{"IAB1"},
		Name:   "App",
		Ver:    "1.0",
	}
	updateApp(req, sig)

	b, err := json.Marshal(req.App)
	require.NoError(t, err)
	assert.JSONEq(t, `{"bundle":"b","cat":["IAB1"],"domain":"d.example","name":"App","ver":"1.0"}`, string(b))
}

func TestUpdateDevice(t *testing.T) {
	req := &openrtb2.BidRequest{Device: &openrtb2.Device{UA: "a", Ext: json.RawMessage(`{}`)}}
	sig := &openrtb2.Device{
		OS:  "ios",
		Ext: json.RawMessage(`{"atts":3}`),
	}
	updateDevice(req, sig)

	b, err := json.Marshal(req.Device)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ua":"a","os":"ios","ext":{"atts":3}}`, string(b))
}

func TestUpdateUser(t *testing.T) {
	req := &openrtb2.BidRequest{User: &openrtb2.User{BuyerUID: "keep", Ext: json.RawMessage(`{}`)}}
	sig := &openrtb2.User{
		Yob:      1990,
		Gender:   "M",
		Keywords: "k",
		Ext:      json.RawMessage(`{"sessionduration":120}`),
	}
	updateUser(req, sig)

	b, err := json.Marshal(req.User)
	require.NoError(t, err)
	assert.JSONEq(t, `{"buyeruid":"keep","yob":1990,"gender":"M","keywords":"k","ext":{"sessionduration":120}}`, string(b))
}

func TestUpdateSource(t *testing.T) {
	req := &openrtb2.BidRequest{}
	sig := &openrtb2.Source{Ext: json.RawMessage(`{"omidpn":"p","omidpv":"v"}`)}
	updateSource(req, sig)

	require.NotNil(t, req.Source)
	b, err := json.Marshal(req.Source)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ext":{"omidpn":"p","omidpv":"v"}}`, string(b))
}

func TestUpdateImpression(t *testing.T) {
	tests := []struct {
		name     string
		req      *openrtb2.BidRequest
		sig      []openrtb2.Imp
		expected string
	}{
		{
			name: "merges first signal imp into first request imp",
			req: &openrtb2.BidRequest{
				Imp: []openrtb2.Imp{{
					ID:     "i1",
					TagID:  "t",
					Banner: &openrtb2.Banner{W: ptrutil.ToPtr[int64](300)},
				}},
			},
			sig: []openrtb2.Imp{{
				Instl:             1,
				DisplayManager:    "PubMatic",
				DisplayManagerVer: "4.0",
				Video:             &openrtb2.Video{MIMEs: []string{"video/mp4"}},
			}},
			expected: `{"id":"i1","tagid":"t","banner":{"w":300},"instl":1,"displaymanager":"PubMatic","displaymanagerver":"4.0","video":{"mimes":["video/mp4"]}}`,
		},
		{
			name: "nil signal imps leaves request unchanged",
			req: &openrtb2.BidRequest{
				Imp: []openrtb2.Imp{{ID: "a"}},
			},
			sig:      nil,
			expected: `{"id":"a"}`,
		},
		{
			name:     "empty request imps returns early",
			req:      &openrtb2.BidRequest{},
			sig:      []openrtb2.Imp{{ID: "x"}},
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateImpression(tt.req, tt.sig)
			var impJSON string
			if tt.req.Imp == nil {
				impJSON = "null"
			} else if len(tt.req.Imp) == 0 {
				impJSON = "[]"
			} else {
				b, err := json.Marshal(tt.req.Imp[0])
				require.NoError(t, err)
				impJSON = string(b)
			}
			assert.JSONEq(t, tt.expected, impJSON)
		})
	}
}

func mustEncodeSignalBidRequest(t *testing.T, br *openrtb2.BidRequest) string {
	t.Helper()
	b, err := json.Marshal(br)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(b)
}
