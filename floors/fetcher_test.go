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

func TestFetchQueueLen(t *testing.T) {
	tests := []struct {
		name string
		fq   FetchQueue
		want int
	}{
		{
			name: "Queue is empty",
			fq:   make(FetchQueue, 0),
			want: 0,
		},
		{
			name: "Queue is of lenght 1",
			fq:   make(FetchQueue, 1),
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fq.Len(); got != tt.want {
				t.Errorf("FetchQueue.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchQueueLess(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		fq   FetchQueue
		args args
		want bool
	}{
		{
			name: "first fetchperiod is less than second",
			fq:   FetchQueue{&FetchInfo{FetchTime: 10}, &FetchInfo{FetchTime: 20}},
			args: args{i: 0, j: 1},
			want: true,
		},
		{
			name: "first fetchperiod is greater than second",
			fq:   FetchQueue{&FetchInfo{FetchTime: 30}, &FetchInfo{FetchTime: 10}},
			args: args{i: 0, j: 1},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fq.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("FetchQueue.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchQueueSwap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		fq   FetchQueue
		args args
	}{
		{
			name: "Swap two elements at index i and j",
			fq:   FetchQueue{&FetchInfo{FetchTime: 30}, &FetchInfo{FetchTime: 10}},
			args: args{i: 0, j: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fInfo1, fInfo2 := tt.fq[0], tt.fq[1]
			tt.fq.Swap(tt.args.i, tt.args.j)
			assert.Equal(t, fInfo1, tt.fq[1], "elements are not swapped")
			assert.Equal(t, fInfo2, tt.fq[0], "elements are not swapped")
		})
	}
}

func TestFetchQueuePush(t *testing.T) {
	type args struct {
		element interface{}
	}
	tests := []struct {
		name string
		fq   *FetchQueue
		args args
	}{
		{
			name: "Push element to queue",
			fq:   &FetchQueue{},
			args: args{element: &FetchInfo{FetchTime: 10}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fq.Push(tt.args.element)
			q := *tt.fq
			assert.Equal(t, q[0], &FetchInfo{FetchTime: 10})
		})
	}
}

func TestFetchQueuePop(t *testing.T) {
	tests := []struct {
		name string
		fq   *FetchQueue
		want interface{}
	}{
		{
			name: "Pop element from queue",
			fq:   &FetchQueue{&FetchInfo{FetchTime: 10}},
			want: &FetchInfo{FetchTime: 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fq.Pop(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchQueue.Pop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchQueueTop(t *testing.T) {
	tests := []struct {
		name string
		fq   *FetchQueue
		want *FetchInfo
	}{
		{
			name: "Get top element from queue",
			fq:   &FetchQueue{&FetchInfo{FetchTime: 20}},
			want: &FetchInfo{FetchTime: 20},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fq.Top(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchQueue.Top() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
				data := `{"currency":"USD","modelgroups":[{"modelweight":40,"modelversion":"version1","default":5,"values":{"banner|300x600|www.website.com":3,"banner|728x90|www.website.com":5,"banner|300x600|*":4,"banner|300x250|*":2,"*|*|*":16,"*|300x250|*":10,"*|300x600|*":12,"*|300x600|www.website.com":11,"banner|*|*":8,"banner|300x250|www.website.com":1,"*|728x90|www.website.com":13,"*|300x250|www.website.com":9,"*|728x90|*":14,"banner|728x90|*":6,"banner|*|www.website.com":7,"*|*|www.website.com":15},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]}`
				return []byte(data)
			}(),
			responseStatus: 200,
			want: func() *openrtb_ext.PriceFloorRules {
				var res openrtb_ext.PriceFloorRules
				data := `{"currency":"USD","modelgroups":[{"modelweight":40,"modelversion":"version1","default":5,"values":{"banner|300x600|www.website.com":3,"banner|728x90|www.website.com":5,"banner|300x600|*":4,"banner|300x250|*":2,"*|*|*":16,"*|300x250|*":10,"*|300x600|*":12,"*|300x600|www.website.com":11,"banner|*|*":8,"banner|300x250|www.website.com":1,"*|728x90|www.website.com":13,"*|300x250|www.website.com":9,"*|728x90|*":14,"banner|728x90|*":6,"banner|*|www.website.com":7,"*|*|www.website.com":15},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]}`
				_ = json.Unmarshal([]byte(data), &res.Data)
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
					MaxFileSize: 1,
					MaxRules:    30,
					MaxAge:      60,
					Period:      40,
				},
			},
			response: func() []byte {
				data := `{"currency":"USD","floorProvider":"PM","floorsSchemaVersion":2,"modelGroups":[{"modelVersion":"M_0","modelWeight":1,"schema":{"fields":["domain"]},"values":{"missyusa.com":0.85,"www.missyusa.com":0.7}},{"modelVersion":"M_1","modelWeight":1,"schema":{"fields":["domain"]},"values":{"missyusa.com":1,"www.missyusa.com":1.85}},{"modelVersion":"M_2","modelWeight":5,"schema":{"fields":["domain"]},"values":{"missyusa.com":1.6,"www.missyusa.com":0.7}},{"modelVersion":"M_3","modelWeight":2,"schema":{"fields":["domain"]},"values":{"missyusa.com":1.9,"www.missyusa.com":0.75}},{"modelVersion":"M_4","modelWeight":1,"schema":{"fields":["domain"]},"values":{"www.missyusa.com":1.35,"missyusa.com":1.75}},{"modelVersion":"M_5","modelWeight":2,"schema":{"fields":["domain"]},"values":{"missyusa.com":1.4,"www.missyusa.com":0.9}},{"modelVersion":"M_6","modelWeight":43,"schema":{"fields":["domain"]},"values":{"www.missyusa.com":2,"missyusa.com":2}},{"modelVersion":"M_7","modelWeight":1,"schema":{"fields":["domain"]},"values":{"missyusa.com":1.4,"www.missyusa.com":1.85}},{"modelVersion":"M_8","modelWeight":3,"schema":{"fields":["domain"]},"values":{"www.missyusa.com":1.7,"missyusa.com":0.1}},{"modelVersion":"M_9","modelWeight":7,"schema":{"fields":["domain"]},"values":{"missyusa.com":1.9,"www.missyusa.com":1.05}},{"modelVersion":"M_10","modelWeight":9,"schema":{"fields":["domain"]},"values":{"www.missyusa.com":2,"missyusa.com":0.1}},{"modelVersion":"M_11","modelWeight":1,"schema":{"fields":["domain"]},"values":{"missyusa.com":0.45,"www.missyusa.com":1.5}},{"modelVersion":"M_12","modelWeight":8,"schema":{"fields":["domain"]},"values":{"missyusa.com":1.2,"www.missyusa.com":1.7}},{"modelVersion":"M_13","modelWeight":8,"schema":{"fields":["domain"]},"values":{"missyusa.com":0.85,"www.missyusa.com":0.75}},{"modelVersion":"M_14","modelWeight":1,"schema":{"fields":["domain"]},"values":{"missyusa.com":1.8,"www.missyusa.com":1}},{"modelVersion":"M_15","modelWeight":1,"schema":{"fields":["domain"]},"values":{"www.missyusa.com":1.2,"missyusa.com":1.75}},{"modelVersion":"M_16","modelWeight":2,"schema":{"fields":["domain"]},"values":{"missyusa.com":1,"www.missyusa.com":0.7}},{"modelVersion":"M_17","modelWeight":1,"schema":{"fields":["domain"]},"values":{"missyusa.com":0.45,"www.missyusa.com":0.35}},{"modelVersion":"M_18","modelWeight":3,"schema":{"fields":["domain"]},"values":{"missyusa.com":1.2,"www.missyusa.com":1.05}}],"skipRate":10}`
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

	refetchCheckInterval = 1

	response := []byte(`{"currency":"USD","modelgroups":[{"modelweight":40,"modelversion":"version1","default":5,"values":{"banner|300x600|www.website.com":3,"banner|728x90|www.website.com":5,"banner|300x600|*":4,"banner|300x250|*":2,"*|*|*":16,"*|300x250|*":10,"*|300x600|*":12,"*|300x600|www.website.com":11,"banner|*|*":8,"banner|300x250|www.website.com":1,"*|728x90|www.website.com":13,"*|300x250|www.website.com":9,"*|728x90|*":14,"banner|728x90|*":6,"banner|*|www.website.com":7,"*|*|www.website.com":15},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]}`)
	mockHandler := func(mockResponse []byte, mockStatus int) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(mockStatus)
			w.Write(mockResponse)
		})
	}

	mockHttpServer := httptest.NewServer(mockHandler(response, 200))
	defer mockHttpServer.Close()

	fectherInstance := NewPriceFloorFetcher(5, 10, 1, 20)
	defer fectherInstance.Stop()
	defer fectherInstance.pool.Stop()

	fetchConfig := config.AccountPriceFloors{
		Enabled:        true,
		UseDynamicData: true,
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

	for i := 0; i < 500; i++ {
		fectherInstance.Fetch(fetchConfig)
	}

	assert.Never(t, func() bool { return len(fectherInstance.fetchQueue) > 1 }, time.Duration(5*time.Second), 100*time.Millisecond, "Queue Got more than one entry")
	assert.Never(t, func() bool { return len(fectherInstance.fetchInprogress) > 1 }, time.Duration(5*time.Second), 100*time.Millisecond, "Map Got more than one entry")

}
