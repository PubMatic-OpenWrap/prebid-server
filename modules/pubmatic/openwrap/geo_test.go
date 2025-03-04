package openwrap

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb"
	mock_geodb "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb/mock"
	mock_metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	mock_feature "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/publisherfeature/mock"
	"github.com/stretchr/testify/assert"
)

const (
	geoWithPubid      = `http://localhost:8001/geo?pubid=23105`
	geoWithoutPubid   = `http://localhost:8001/geo`
	geoWithWrongPubid = `http://localhost:8001/geo?pubid=bad`
)

type headerType int

const (
	NoHeaders headerType = iota
	NonEURequest
	EURequest
	USPRequest
	GPPCountryRequest
	InvalidIP
	InvalidUID
	NilGeo
	GeoLookupFail
	EmptyGeo
)

func getTestHTTPRequest(url string, headers http.Header) *http.Request {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic("error: creating an http request")
	}

	for k, v := range headers {
		r.Header.Add(k, v[0])
	}

	if err := r.ParseForm(); err != nil {
		panic("error: parsing http request")
	}
	return r
}

func getHeaders(headerType headerType) http.Header {
	switch headerType {
	case NoHeaders:
		return nil
	case NonEURequest:
		return http.Header{
			"User-Agent": []string{"golang sample request"},
			"Cookie":     []string{"KADUSERCOOKIE=pmuserid"},
			"SOURCE_IP":  []string{"115.114.134.174"},
		}
	case EURequest:
		return http.Header{
			"User-Agent": []string{"golang sample request"},
			"Cookie":     []string{"KADUSERCOOKIE=pmuserid"},
			"SOURCE_IP":  []string{"2.16.1.255"},
		}
	case USPRequest:
		return http.Header{
			"User-Agent": []string{"golang sample request"},
			"Cookie":     []string{"KADUSERCOOKIE=pmuserid"},
			"SOURCE_IP":  []string{"43.135.143.132"},
		}
	case GPPCountryRequest:
		return http.Header{
			"User-Agent": []string{"golang sample request"},
			"Cookie":     []string{"KADUSERCOOKIE=pmuserid"},
			"SOURCE_IP":  []string{"208.253.114.165"},
		}
	case InvalidIP:
		return http.Header{
			"User-Agent": []string{"golang sample request"},
			"Cookie":     []string{"KADUSERCOOKIE=pmuserid"},
			"SOURCE_IP":  []string{"115.114.134"},
		}
	case InvalidUID:
		return http.Header{
			"User-Agent": []string{"golang sample request"},
			"Cookie":     []string{"KADUSERCOOKIE=pmuserid}"},
		}
	case EmptyGeo:
		return http.Header{
			"SOURCE_IP": []string{"30.30.30.30"},
		}
	case GeoLookupFail:
		return http.Header{
			"SOURCE_IP": []string{"20.20.20.20"},
		}
	case NilGeo:
		return http.Header{
			"SOURCE_IP": []string{"10.10.10.10"},
		}
	}
	return nil
}

func TestHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetrics := mock_metrics.NewMockMetricsEngine(ctrl)
	mockgeodb := mock_geodb.NewMockGeography(ctrl)
	mockFeature := mock_feature.NewMockFeature(ctrl)

	originalOw := ow
	defer func() { ow = originalOw }()
	ow = &OpenWrap{
		metricEngine:   mockMetrics,
		geoInfoFetcher: mockgeodb,
		pubFeatures:    mockFeature,
	}

	type args struct {
		r     *http.Request
		setup func()
	}
	type want struct {
		geo       string
		statuCode int
		header    []string
	}
	tests := []struct {
		name  string
		args  args
		setup geodb.Geography
		want  want
	}{
		{
			name: "POST request",
			args: args{
				r: func() *http.Request {
					r, err := http.NewRequest("POST", geoWithoutPubid, nil)
					if err != nil {
						panic("error: creating an http request")
					}
					return r
				}(),
				setup: func() {
					mockMetrics.EXPECT().RecordBadRequests(models.EndpointGeo, "", 0)
					mockMetrics.EXPECT().RecordPublisherRequests(models.EndpointGeo, "", "")
				},
			},
			want: want{
				geo:       "",
				statuCode: http.StatusMethodNotAllowed,
				header:    nil,
			},
		},
		{
			name: "invalid pubid",
			args: args{
				r: getTestHTTPRequest(geoWithWrongPubid, getHeaders(NoHeaders)),
				setup: func() {
					mockMetrics.EXPECT().RecordBadRequests(models.EndpointGeo, "bad", 0)
					mockMetrics.EXPECT().RecordPublisherRequests(models.EndpointGeo, "bad", "")
				},
			},
			want: want{
				geo:       "",
				statuCode: http.StatusBadRequest,
				header:    nil,
			},
		},
		{
			name: "pubid not present",
			args: args{
				r: getTestHTTPRequest(geoWithoutPubid, getHeaders(NoHeaders)),
				setup: func() {
					mockMetrics.EXPECT().RecordBadRequests(models.EndpointGeo, "", 0)
					mockMetrics.EXPECT().RecordPublisherRequests(models.EndpointGeo, "", "")
				},
			},
			want: want{
				geo:       "",
				statuCode: http.StatusBadRequest,
				header:    nil,
			},
		},
		{
			name: "valid pubid with no ip in headers",
			args: args{
				r: getTestHTTPRequest(geoWithPubid, getHeaders(NoHeaders)),
				setup: func() {
					mockMetrics.EXPECT().RecordPublisherRequests(models.EndpointGeo, "23105", "")
					mockMetrics.EXPECT().RecordGeoLookupFailure(models.EndpointGeo)
					mockgeodb.EXPECT().LookUp(gomock.Any()).Return(&geodb.GeoInfo{}, nil)
				},
			},
			want: want{
				geo:       "",
				statuCode: http.StatusOK,
				header:    nil,
			},
		},
		{
			name: "ip Lookup fail",
			args: args{
				r: getTestHTTPRequest(geoWithPubid, getHeaders(GeoLookupFail)),
				setup: func() {
					mockMetrics.EXPECT().RecordPublisherRequests(models.EndpointGeo, "23105", "")
					mockMetrics.EXPECT().RecordGeoLookupFailure(models.EndpointGeo)
					mockgeodb.EXPECT().LookUp(gomock.Any()).Return(nil, errors.New("ErrDummy"))
				},
			},
			want: want{
				geo:       "",
				statuCode: http.StatusOK,
				header:    nil,
			},
		},
		{
			name: "ip lookup returns nil",
			args: args{
				r: getTestHTTPRequest(geoWithPubid, getHeaders(NilGeo)),
				setup: func() {
					mockMetrics.EXPECT().RecordPublisherRequests(models.EndpointGeo, "23105", "")
					mockMetrics.EXPECT().RecordGeoLookupFailure(models.EndpointGeo)
					mockgeodb.EXPECT().LookUp(gomock.Any()).Return(nil, nil)
				},
			},
			want: want{
				geo:       "",
				statuCode: http.StatusOK,
				header:    nil,
			},
		},
		{
			name: "empty countrycode",
			args: args{
				r: getTestHTTPRequest(geoWithPubid, getHeaders(NonEURequest)),
				setup: func() {
					mockMetrics.EXPECT().RecordPublisherRequests(models.EndpointGeo, "23105", "")
					mockMetrics.EXPECT().RecordGeoLookupFailure(models.EndpointGeo)
					mockgeodb.EXPECT().LookUp(gomock.Any()).Return(&geodb.GeoInfo{ISOCountryCode: ""}, nil)
				},
			},
			want: want{
				geo:       "",
				statuCode: http.StatusOK,
				header:    nil,
			},
		},
		{
			name: "EU region request",
			args: args{
				r: getTestHTTPRequest(geoWithPubid, getHeaders(EURequest)),
				setup: func() {
					mockMetrics.EXPECT().RecordPublisherRequests(models.EndpointGeo, "23105", "")
					mockFeature.EXPECT().IsCountryGDPREnabled(gomock.Any()).Return(true)
					mockgeodb.EXPECT().LookUp(gomock.Any()).
						Return(&geodb.GeoInfo{ISOCountryCode: "UK", CountryCode: "uk", RegionCode: "lnd"}, nil)
				},
			},
			want: want{
				geo:       "{\"cc\":\"UK\",\"sc\":\"lnd\",\"gc\":1}\n",
				statuCode: http.StatusOK,
				header:    []string{"max-age=172800"},
			},
		},
		{
			name: "non-EU region request",
			args: args{
				r: getTestHTTPRequest(geoWithPubid, getHeaders(NonEURequest)),
				setup: func() {
					mockMetrics.EXPECT().RecordPublisherRequests(models.EndpointGeo, "23105", "")
					mockFeature.EXPECT().IsCountryGDPREnabled(gomock.Any()).Return(false)
					mockgeodb.EXPECT().LookUp(gomock.Any()).
						Return(&geodb.GeoInfo{ISOCountryCode: "IN", CountryCode: "in", RegionCode: "mh"}, nil)
				},
			},
			want: want{
				geo:       "{\"cc\":\"IN\",\"sc\":\"mh\"}\n",
				statuCode: http.StatusOK,
				header:    []string{"max-age=172800"},
			},
		},
		{
			name: "statecode is california(USP compliance)",
			args: args{
				r: getTestHTTPRequest(geoWithPubid, getHeaders(EURequest)),
				setup: func() {
					mockMetrics.EXPECT().RecordPublisherRequests(models.EndpointGeo, "23105", "")
					mockFeature.EXPECT().IsCountryGDPREnabled(gomock.Any()).Return(false)
					mockgeodb.EXPECT().LookUp(gomock.Any()).
						Return(&geodb.GeoInfo{ISOCountryCode: "US", CountryCode: "us", RegionCode: "ca"}, nil)
				},
			},
			want: want{
				geo:       "{\"cc\":\"US\",\"sc\":\"ca\",\"gc\":2}\n",
				statuCode: http.StatusOK,
				header:    []string{"max-age=172800"},
			},
		},
		{
			name: "gpp country request",
			args: args{
				r: getTestHTTPRequest(geoWithPubid, getHeaders(EURequest)),
				setup: func() {
					mockMetrics.EXPECT().RecordPublisherRequests(models.EndpointGeo, "23105", "")
					mockFeature.EXPECT().IsCountryGDPREnabled(gomock.Any()).Return(false)
					mockgeodb.EXPECT().LookUp(gomock.Any()).
						Return(&geodb.GeoInfo{ISOCountryCode: "US", CountryCode: "us", RegionCode: "va"}, nil)
				},
			},
			want: want{
				geo:       "{\"cc\":\"US\",\"sc\":\"va\",\"gc\":3,\"gsId\":9}\n",
				statuCode: http.StatusOK,
				header:    []string{"max-age=172800"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setup()
			w := httptest.NewRecorder()
			Handler(w, tt.args.r)
			assert.Equal(t, tt.want.statuCode, w.Result().StatusCode)
			assert.Equal(t, tt.want.geo, w.Body.String())
			assert.Equal(t, tt.want.header, w.Header().Values("Cache-Control"))
		})
	}
}

func TestUpdateGeoObject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFeature := mock_feature.NewMockFeature(ctrl)

	originalOw := ow
	defer func() { ow = originalOw }()
	ow = &OpenWrap{pubFeatures: mockFeature}

	type args struct {
		geo   *geo
		setup func()
	}
	tests := []struct {
		name string
		args args
		want *geo
	}{
		{
			name: "gdpr compliance",
			args: args{
				geo: &geo{
					CountryCode: "country",
					StateCode:   "state",
				},
				setup: func() {
					mockFeature.EXPECT().IsCountryGDPREnabled(gomock.Any()).Return(true)
				},
			},
			want: &geo{
				CountryCode: "country",
				StateCode:   "state",
				Compliance:  gdprCompliance,
			},
		},
		{
			name: "usp compliance",
			args: args{
				geo: &geo{
					CountryCode: "US",
					StateCode:   "ca",
				},
				setup: func() {
					mockFeature.EXPECT().IsCountryGDPREnabled(gomock.Any()).Return(false)
				},
			},
			want: &geo{
				CountryCode: "US",
				StateCode:   "ca",
				Compliance:  uspCompliance,
			},
		},
		{
			name: "gpp compliance",
			args: args{
				geo: &geo{
					CountryCode: "US",
					StateCode:   "ct",
				},
				setup: func() {
					mockFeature.EXPECT().IsCountryGDPREnabled(gomock.Any()).Return(false)
				},
			},
			want: &geo{
				CountryCode: "US",
				StateCode:   "ct",
				Compliance:  gppCompliance,
				SectionID:   12,
			},
		},
		{
			name: "no compliance",
			args: args{
				geo: &geo{
					CountryCode: "country",
					StateCode:   "state",
				},
				setup: func() {
					mockFeature.EXPECT().IsCountryGDPREnabled(gomock.Any()).Return(false)
				},
			},
			want: &geo{
				CountryCode: "country",
				StateCode:   "state",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setup()
			updateGeoObject(tt.args.geo)
			assert.Equal(t, tt.want, tt.args.geo, tt.name)
		})
	}
}
