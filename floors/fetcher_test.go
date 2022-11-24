package floors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestValidatePriceFloorRules(t *testing.T) {

	zero := 0
	one_o_one := 101
	type args struct {
		configs     config.AccountFloorFetch
		priceFloors *openrtb_ext.PriceFloorRules
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Price floor data is empty",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    5,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{},
			},
			wantErr: true,
		},
		{
			name: "Model group array is empty",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    5,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{
					Data: &openrtb_ext.PriceFloorData{},
				},
			},
			wantErr: true,
		},
		{
			name: "floor rules is empty",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    5,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{
					Data: &openrtb_ext.PriceFloorData{
						ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
							Values: map[string]float64{},
						}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "floor rules is grater than max floor rules",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    0,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{
					Data: &openrtb_ext.PriceFloorData{
						ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
							Values: map[string]float64{
								"*|*|www.website.com": 15.01,
							},
						}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Modelweight is zero",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    1,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{
					Data: &openrtb_ext.PriceFloorData{
						ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
							Values: map[string]float64{
								"*|*|www.website.com": 15.01,
							},
							ModelWeight: &zero,
						}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Modelweight is 101",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    1,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{
					Data: &openrtb_ext.PriceFloorData{
						ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
							Values: map[string]float64{
								"*|*|www.website.com": 15.01,
							},
							ModelWeight: &one_o_one,
						}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "skiprate is 101",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    1,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{
					Data: &openrtb_ext.PriceFloorData{
						ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
							Values: map[string]float64{
								"*|*|www.website.com": 15.01,
							},
							SkipRate: 101,
						}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Default is -1",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					URL:         "abc.com",
					Timeout:     5,
					MaxFileSize: 20,
					MaxRules:    1,
					MaxAge:      20,
					Period:      10,
				},
				priceFloors: &openrtb_ext.PriceFloorRules{
					Data: &openrtb_ext.PriceFloorData{
						ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
							Values: map[string]float64{
								"*|*|www.website.com": 15.01,
							},
							Default: -1,
						}},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateRules(tt.args.configs, tt.args.priceFloors); (err != nil) != tt.wantErr {
				t.Errorf("validatePriceFloorRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFetchFloorRulesFromURL(t *testing.T) {

	mockHandler := func(mockResponse []byte, mockStatus int) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Length", "645")
			w.WriteHeader(mockStatus)
			w.Write(mockResponse)
		})
	}

	type args struct {
		URL     string
		timeout int
	}
	tests := []struct {
		name           string
		args           args
		response       []byte
		responseStatus int
		want           []byte
		wantErr        bool
	}{
		{
			name: "Floor data is successfully returned",
			args: args{
				URL:     "",
				timeout: 60,
			},
			response: func() []byte {
				data := `{"data":{"currency":"USD","modelgroups":[{"modelweight":40,"modelversion":"version1","default":5,"values":{"banner|300x600|www.website.com":3,"banner|728x90|www.website.com":5,"banner|300x600|*":4,"banner|300x250|*":2,"*|*|*":16,"*|300x250|*":10,"*|300x600|*":12,"*|300x600|www.website.com":11,"banner|*|*":8,"banner|300x250|www.website.com":1,"*|728x90|www.website.com":13,"*|300x250|www.website.com":9,"*|728x90|*":14,"banner|728x90|*":6,"banner|*|www.website.com":7,"*|*|www.website.com":15},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"floormin":1,"enforcement":{"enforcepbs":false,"floordeals":true}}`
				return []byte(data)
			}(),
			responseStatus: 200,
			want: func() []byte {
				data := `{"data":{"currency":"USD","modelgroups":[{"modelweight":40,"modelversion":"version1","default":5,"values":{"banner|300x600|www.website.com":3,"banner|728x90|www.website.com":5,"banner|300x600|*":4,"banner|300x250|*":2,"*|*|*":16,"*|300x250|*":10,"*|300x600|*":12,"*|300x600|www.website.com":11,"banner|*|*":8,"banner|300x250|www.website.com":1,"*|728x90|www.website.com":13,"*|300x250|www.website.com":9,"*|728x90|*":14,"banner|728x90|*":6,"banner|*|www.website.com":7,"*|*|www.website.com":15},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"floormin":1,"enforcement":{"enforcepbs":false,"floordeals":true}}`
				return []byte(data)
			}(),
			wantErr: false,
		},
		{
			name: "Time out occured",
			args: args{
				URL:     "",
				timeout: 0,
			},
			responseStatus: 200,
			wantErr:        true,
		},
		{
			name: "Invalid URL",
			args: args{
				URL:     "%%",
				timeout: 10,
			},
			responseStatus: 200,
			wantErr:        true,
		},
		{
			name: "No response from server",
			args: args{
				URL:     "",
				timeout: 10,
			},
			responseStatus: 500,
			wantErr:        true,
		},
		{
			name: "Invalid response",
			args: args{
				URL:     "",
				timeout: 10,
			},
			response:       []byte("1"),
			responseStatus: 200,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpServer := httptest.NewServer(mockHandler(tt.response, tt.responseStatus))
			defer mockHttpServer.Close()

			var url string
			if tt.args.URL != "" {
				url = tt.args.URL
			} else {
				url = mockHttpServer.URL
			}
			got, err := fetchFloorRulesFromURL(url, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchFloorRulesFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Success fetchFloorRulesFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchAndValidate(t *testing.T) {

	mockHandler := func(mockResponse []byte, mockStatus int) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(mockStatus)
			w.Write(mockResponse)
		})
	}

	type args struct {
		configs config.AccountFloorFetch
	}
	tests := []struct {
		name           string
		args           args
		response       []byte
		responseStatus int
		want           *openrtb_ext.PriceFloorRules
	}{
		{
			name: "Recieved valid price floor rules response",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					Timeout:     30,
					MaxFileSize: 700,
					MaxRules:    30,
					MaxAge:      60,
					Period:      40,
				},
			},
			response: func() []byte {
				data := `{"data":{"currency":"USD","modelgroups":[{"modelweight":40,"modelversion":"version1","default":5,"values":{"banner|300x600|www.website.com":3,"banner|728x90|www.website.com":5,"banner|300x600|*":4,"banner|300x250|*":2,"*|*|*":16,"*|300x250|*":10,"*|300x600|*":12,"*|300x600|www.website.com":11,"banner|*|*":8,"banner|300x250|www.website.com":1,"*|728x90|www.website.com":13,"*|300x250|www.website.com":9,"*|728x90|*":14,"banner|728x90|*":6,"banner|*|www.website.com":7,"*|*|www.website.com":15},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"floormin":1,"enforcement":{"enforcepbs":false,"floordeals":true}}`
				return []byte(data)
			}(),
			responseStatus: 200,
			want: func() *openrtb_ext.PriceFloorRules {
				var res openrtb_ext.PriceFloorRules
				data := `{"data":{"currency":"USD","modelgroups":[{"modelweight":40,"modelversion":"version1","default":5,"values":{"banner|300x600|www.website.com":3,"banner|728x90|www.website.com":5,"banner|300x600|*":4,"banner|300x250|*":2,"*|*|*":16,"*|300x250|*":10,"*|300x600|*":12,"*|300x600|www.website.com":11,"banner|*|*":8,"banner|300x250|www.website.com":1,"*|728x90|www.website.com":13,"*|300x250|www.website.com":9,"*|728x90|*":14,"banner|728x90|*":6,"banner|*|www.website.com":7,"*|*|www.website.com":15},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"floormin":1,"enforcement":{"enforcepbs":false,"floordeals":true}}`
				_ = json.Unmarshal([]byte(data), &res)
				return &res
			}(),
		},
		{
			name: "No response from server",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					Timeout:     30,
					MaxFileSize: 700,
					MaxRules:    30,
					MaxAge:      60,
					Period:      40,
				},
			},
			response:       []byte{},
			responseStatus: 500,
			want:           nil,
		},
		{
			name: "File is greater than MaxFileSize",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					Timeout:     30,
					MaxFileSize: 100,
					MaxRules:    30,
					MaxAge:      60,
					Period:      40,
				},
			},
			response: func() []byte {
				data := `{"data":{"currency":"USD","modelgroups":[{"modelweight":40,"modelversion":"version1","default":5,"values":{"banner|300x600|www.website.com":3,"banner|728x90|www.website.com":5,"banner|300x600|*":4,"banner|300x250|*":2,"*|*|*":16,"*|300x250|*":10,"*|300x600|*":12,"*|300x600|www.website.com":11,"banner|*|*":8,"banner|300x250|www.website.com":1,"*|728x90|www.website.com":13,"*|300x250|www.website.com":9,"*|728x90|*":14,"banner|728x90|*":6,"banner|*|www.website.com":7,"*|*|www.website.com":15},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"floormin":1,"enforcement":{"enforcepbs":false,"floordeals":true}}`
				return []byte(data)
			}(),
			responseStatus: 200,
			want:           nil,
		},
		{
			name: "Malformed response : json unmarshalling failed",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					Timeout:     30,
					MaxFileSize: 800,
					MaxRules:    30,
					MaxAge:      60,
					Period:      40,
				},
			},
			response: func() []byte {
				data := `{"data":nil?}`
				return []byte(data)
			}(),
			responseStatus: 200,
			want:           nil,
		},
		{
			name: "Validations failed for price floor rules response",
			args: args{
				configs: config.AccountFloorFetch{
					Enabled:     true,
					Timeout:     30,
					MaxFileSize: 700,
					MaxRules:    30,
					MaxAge:      60,
					Period:      40,
				},
			},
			response: func() []byte {
				data := `{"data":{"currency":"USD","modelgroups":[]},"enabled":true,"floormin":1,"enforcement":{"enforcepbs":false,"floordeals":true}}`
				return []byte(data)
			}(),
			responseStatus: 200,
			want:           nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpServer := httptest.NewServer(mockHandler(tt.response, tt.responseStatus))
			defer mockHttpServer.Close()

			tt.args.configs.URL = mockHttpServer.URL
			if got := fetchAndValidate(tt.args.configs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fetchAndValidate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetcherWhenRequestGetSameURLInrequest(t *testing.T) {

	response := []byte(`{"data":{"currency":"USD","modelgroups":[{"modelweight":40,"modelversion":"version1","default":5,"values":{"banner|300x600|www.website.com":3,"banner|728x90|www.website.com":5,"banner|300x600|*":4,"banner|300x250|*":2,"*|*|*":16,"*|300x250|*":10,"*|300x600|*":12,"*|300x600|www.website.com":11,"banner|*|*":8,"banner|300x250|www.website.com":1,"*|728x90|www.website.com":13,"*|300x250|www.website.com":9,"*|728x90|*":14,"banner|728x90|*":6,"banner|*|www.website.com":7,"*|*|www.website.com":15},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"floormin":1,"enforcement":{"enforcepbs":false,"floordeals":true}}`)
	mockHandler := func(mockResponse []byte, mockStatus int) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(mockStatus)
			w.Write(mockResponse)
		})
	}

	mockHttpServer := httptest.NewServer(mockHandler(response, 200))
	defer mockHttpServer.Close()

	fectherInstance := NewPriceFloorFetcher(5, 10)
	defer fectherInstance.Stop()
	defer fectherInstance.pool.Stop()

	fetchConfig := config.AccountPriceFloors{
		Enabled: true,
		Fetch: config.AccountFloorFetch{
			Enabled:     true,
			URL:         mockHttpServer.URL,
			Timeout:     100,
			MaxFileSize: 1000,
			MaxRules:    100,
			MaxAge:      20,
			Period:      1,
		},
	}

	for i := 0; i < 50; i++ {
		fectherInstance.Fetch(fetchConfig)
	}

	assert.Never(t, func() bool { return len(fectherInstance.fetchQueue) > 1 }, time.Duration(5*time.Second), 100*time.Millisecond, "Queue Got more than one entry")
	assert.Never(t, func() bool { return len(fectherInstance.fetchInprogress) > 1 }, time.Duration(5*time.Second), 100*time.Millisecond, "Map Got more than one entry")

}
