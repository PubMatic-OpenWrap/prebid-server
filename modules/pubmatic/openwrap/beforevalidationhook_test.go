package openwrap

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/hooks/hookanalytics"
	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	adapters "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics"
	mock_metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	modelsAdunitConfig "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	mock_feature "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/publisherfeature/mock"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"

	"github.com/stretchr/testify/assert"
)

var rctx = models.RequestCtx{
	ProfileID:                 1234,
	DisplayID:                 1,
	SSAuction:                 -1,
	Platform:                  "in-app",
	Debug:                     true,
	UA:                        "go-test",
	IP:                        "127.0.0.1",
	IsCTVRequest:              false,
	TrackerEndpoint:           "t.pubmatic.com",
	VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
	UidCookie: &http.Cookie{
		Name:  "uids",
		Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
	},
	KADUSERCookie: &http.Cookie{
		Name:  "KADUSERCOOKIE",
		Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
	},
	OriginCookie:             "go-test",
	Aliases:                  make(map[string]string),
	ImpBidCtx:                make(map[string]models.ImpCtx),
	PrebidBidderCode:         make(map[string]string),
	BidderResponseTimeMillis: make(map[string]int),
	ProfileIDStr:             "1234",
	Endpoint:                 models.EndpointV25,
	SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
}

func getTestBidRequest(isSite bool) *openrtb2.BidRequest {

	testReq := &openrtb2.BidRequest{}

	testReq.ID = "testID"

	testReq.Imp = []openrtb2.Imp{
		{
			ID: "testImp1",
			Banner: &openrtb2.Banner{
				W: ptrutil.ToPtr[int64](200),
				H: ptrutil.ToPtr[int64](300),
			},
			Video: &openrtb2.Video{
				W:     ptrutil.ToPtr[int64](200),
				H:     ptrutil.ToPtr[int64](300),
				Plcmt: 1,
			},
		},
	}
	if !isSite {
		testReq.App = &openrtb2.App{
			Publisher: &openrtb2.Publisher{
				ID: "1010",
			},
			Content: &openrtb2.Content{
				Language: "english",
			},
		}
	} else {
		testReq.Site = &openrtb2.Site{
			Publisher: &openrtb2.Publisher{
				ID: "1010",
			},
			Content: &openrtb2.Content{
				Language: "english",
			},
		}
	}
	testReq.Cur = []string{"EUR"}
	testReq.WLang = []string{"english", "hindi"}
	testReq.Device = &openrtb2.Device{
		DeviceType: 1,
		Language:   "english",
	}
	return testReq
}

func TestGetPageURL(t *testing.T) {
	type args struct {
		bidRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "App_storeurl_is_not_empty",
			args: args{
				bidRequest: &openrtb2.BidRequest{
					App: &openrtb2.App{
						StoreURL: "testurlApp",
					},
				},
			},
			want: "testurlApp",
		},
		{
			name: "Site_page_is_not_empty",
			args: args{
				bidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Page: "testurlSite",
					},
				},
			},
			want: "testurlSite",
		},
		{
			name: "both_app_and_site_are_nil",
			args: args{
				bidRequest: &openrtb2.BidRequest{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPageURL(tt.args.bidRequest)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetVASTEventMacros(t *testing.T) {
	type args struct {
		rctx models.RequestCtx
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "SSAI_is_empty",
			args: args{
				rctx: models.RequestCtx{
					ProfileID:          1234,
					DisplayVersionID:   1234,
					StartTime:          1234,
					LoggerImpressionID: "1234",
					SSAI:               "",
					DeviceCtx: models.DeviceCtx{
						Platform: 1234,
					},
				},
			},
			want: map[string]string{
				"[PROFILE_ID]":            "1234",
				"[PROFILE_VERSION]":       "1234",
				"[UNIX_TIMESTAMP]":        "1234",
				"[PLATFORM]":              "1234",
				"[WRAPPER_IMPRESSION_ID]": "1234",
			},
		},
		{
			name: "SSAI_is_not_empty",
			args: args{
				rctx: models.RequestCtx{
					ProfileID:          1234,
					DisplayVersionID:   1234,
					StartTime:          1234,
					LoggerImpressionID: "1234",
					SSAI:               "1234",
					DeviceCtx: models.DeviceCtx{
						Platform: 1234,
					},
				},
			},
			want: map[string]string{
				"[PROFILE_ID]":            "1234",
				"[PROFILE_VERSION]":       "1234",
				"[UNIX_TIMESTAMP]":        "1234",
				"[PLATFORM]":              "1234",
				"[WRAPPER_IMPRESSION_ID]": "1234",
				"[SSAI]":                  "1234",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getVASTEventMacros(tt.args.rctx)
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestUpdateAliasGVLIds(t *testing.T) {
	type args struct {
		aliasgvlids   map[string]uint16
		bidderCode    string
		partnerConfig map[string]string
	}
	type want struct {
		aliasgvlids map[string]uint16
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "vendorId not present in config",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{},
			},
			want: want{
				aliasgvlids: map[string]uint16{},
			},
		},
		{
			name: "Empty vendorID",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{models.VENDORID: ""},
			},
			want: want{
				aliasgvlids: map[string]uint16{},
			},
		},
		{
			name: "Error parsing vendorID",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{models.VENDORID: "abc"},
			},
		},
		{
			name: "VendorID is 0",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{models.VENDORID: "0"},
			},
			want: want{
				aliasgvlids: map[string]uint16{},
			},
		},
		{
			name: "Negative vendorID",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{models.VENDORID: "-76"},
			},
		},
		{
			name: "Valid vendorID",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "vastbidder1",
				partnerConfig: map[string]string{models.VENDORID: "76"},
			},
			want: want{
				aliasgvlids: map[string]uint16{"vastbidder1": uint16(76)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateAliasGVLIds(tt.args.aliasgvlids, tt.args.bidderCode, tt.args.partnerConfig)
		})
	}
}

func TestOpenWrap_setTimeout(t *testing.T) {
	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		rCtx       models.RequestCtx
		bidRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		{
			name: "Highest_priority_to_request_tmax_parameter",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"ssTimeout": "250",
						},
					},
				},
				bidRequest: &openrtb2.BidRequest{
					TMax: 220,
				},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 200,
						MaxTimeout: 300,
					},
				},
			},
			want: 220,
		},
		{
			name: "tmax_parameter_less_than_minTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"ssTimeout": "250",
						},
					},
				},
				bidRequest: &openrtb2.BidRequest{
					TMax: 10,
				},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 200,
						MaxTimeout: 300,
					},
				},
			},
			want: 200,
		},
		{
			name: "ssTimeout_greater_than_minTimeout_and_less_than_maxTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"ssTimeout": "250",
						},
					},
				},
				bidRequest: &openrtb2.BidRequest{},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 200,
						MaxTimeout: 300,
					},
				},
			},
			want: 250,
		},
		{
			name: "ssTimeout_less_than_minTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"ssTimeout": "250",
						},
					},
				},
				bidRequest: &openrtb2.BidRequest{},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 300,
						MaxTimeout: 400,
					},
				},
			},
			want: 300,
		},
		{
			name: "ssTimeout_greater_than_minTimeout_and_also_greater_than_maxTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"ssTimeout": "500",
						},
					},
				},
				bidRequest: &openrtb2.BidRequest{},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 300,
						MaxTimeout: 400,
					},
				},
			},
			want: 400,
		},
		{
			name: "ssTimeout_greater_than_minTimeout_and_less_than_maxTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"ssTimeout": "400",
						},
					},
				},
				bidRequest: &openrtb2.BidRequest{},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 300,
						MaxTimeout: 500,
					},
				},
			},
			want: 400,
		},
		//Below piece of code is applicable for older profiles where ssTimeout is not set
		//Here we will check the partner timeout and select max timeout considering timeout range
		{
			name: "at_lease_one_partner_timeout_greater_than_cofig_maxTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							"timeout": "500",
						},
						2: {
							"timeout": "250",
						},
					},
				},
				bidRequest: &openrtb2.BidRequest{},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 200,
						MaxTimeout: 300,
					},
				},
			},
			want: 300,
		},
		{
			name: "all_partner_timeout_less_than_cofig_maxTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							"timeout": "230",
						},
						2: {
							"timeout": "250",
						},
						3: {
							"timeout": "280",
						},
					},
				},
				bidRequest: &openrtb2.BidRequest{},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 200,
						MaxTimeout: 300,
					},
				},
			},
			want: 280,
		},
		{
			name: "all_partner_timeout_less_than_cofig_minTimeout",
			args: args{
				rCtx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							"timeout": "100",
						},
						2: {
							"timeout": "150",
						},
						3: {
							"timeout": "180",
						},
					},
				},
				bidRequest: &openrtb2.BidRequest{},
			},
			fields: fields{
				cfg: config.Config{
					Timeout: config.Timeout{
						MinTimeout: 200,
						MaxTimeout: 300,
					},
				},
			},
			want: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: tt.fields.metricEngine,
			}
			got := m.setTimeout(tt.args.rCtx, tt.args.bidRequest)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsSendAllBids(t *testing.T) {
	type args struct {
		rctx models.RequestCtx
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "sendallbids_always_true_for_webs2s_endpoint",
			args: args{
				rctx: models.RequestCtx{
					Endpoint: models.EndpointWebS2S,
				},
			},
			want: true,
		},
		{
			name: "Don't_do_ssauction",
			args: args{
				rctx: models.RequestCtx{
					SSAuction: 0,
				},
			},
			want: true,
		},
		{
			name: "SSAuction_flag_not_set_In-app_sendAllbids_flag_1",
			args: args{
				rctx: models.RequestCtx{
					SSAuction: -1,
					Platform:  models.PLATFORM_APP,
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"sendAllBids": "1",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "SSAuction_flag_not_set_In-app_sendAllbids_flag_other_than_1",
			args: args{
				rctx: models.RequestCtx{
					SSAuction: -1,
					Platform:  models.PLATFORM_APP,
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"sendAllBids": "5",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "Random_value_of_ssauctionflag",
			args: args{
				rctx: models.RequestCtx{
					SSAuction: 5,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSendAllBids(tt.args.rctx)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetValidLanguage(t *testing.T) {
	type args struct {
		language string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Language_of_length_less_than_2",
			args: args{
				language: "te",
			},
			want: "te",
		},
		{
			name: "Language_of_length_greater_than_2_and_it_is_valid_code",
			args: args{
				language: "hindi",
			},
			want: "hi",
		},
		{
			name: "Language_of_length_greater_than_2_and_it_is_Invalid_code",
			args: args{
				language: "xyz",
			},
			want: "xyz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getValidLanguage(tt.args.language)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsSlotEnabled(t *testing.T) {
	type args struct {
		videoAdUnitCtx  models.AdUnitCtx
		bannerAdUnitCtx models.AdUnitCtx
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Video_enabled_in_Video_adunit_context",
			args: args{
				videoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{
							Enabled: ptrutil.ToPtr(true),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Banner_enabled_in_banner_adunit_context",
			args: args{
				bannerAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Banner: &adunitconfig.Banner{
							Enabled: ptrutil.ToPtr(true),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "both_banner_and_video_enabled_in_adunit_context",
			args: args{
				bannerAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Banner: &adunitconfig.Banner{
							Enabled: ptrutil.ToPtr(true),
						},
					},
				},
				videoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{
							Enabled: ptrutil.ToPtr(true),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "both_banner_and_video_disabled_in_adunit_context",
			args: args{
				bannerAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Banner: &adunitconfig.Banner{
							Enabled: ptrutil.ToPtr(false),
						},
					},
				},
				videoAdUnitCtx: models.AdUnitCtx{
					AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
						Video: &adunitconfig.Video{
							Enabled: ptrutil.ToPtr(false),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "both_banner_and_video_context_are empty",
			args: args{},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSlotEnabled(tt.args.videoAdUnitCtx, tt.args.bannerAdUnitCtx)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestGetPubID(t *testing.T) {
	type args struct {
		bidRequest openrtb2.BidRequest
	}
	type want struct {
		wantErr bool
		pubID   int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "publisher_id_present_in_site_object_and_it_is_valid_integer",
			args: args{
				bidRequest: openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{
							ID: "5890",
						},
					},
				},
			},
			want: want{
				wantErr: false,
				pubID:   5890,
			},
		},
		{
			name: "publisher_id_present_in_site_object_but_it_is_not_valid_integer",
			args: args{
				bidRequest: openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{
							ID: "test",
						},
					},
				},
			},
			want: want{
				wantErr: true,
				pubID:   0,
			},
		},
		{
			name: "publisher_id_present_in_App_object_and_it_is_valid_integer",
			args: args{
				bidRequest: openrtb2.BidRequest{
					App: &openrtb2.App{
						Publisher: &openrtb2.Publisher{
							ID: "5890",
						},
					},
				},
			},
			want: want{
				wantErr: false,
				pubID:   5890,
			},
		},
		{
			name: "publisher_id_present_in_App_object_but_it_is_not_valid_integer",
			args: args{
				bidRequest: openrtb2.BidRequest{
					App: &openrtb2.App{
						Publisher: &openrtb2.Publisher{
							ID: "test",
						},
					},
				},
			},
			want: want{
				wantErr: true,
				pubID:   0,
			},
		},
		{
			name: "publisher_id_present_in_both_Site_and_App_object",
			args: args{
				bidRequest: openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{
							ID: "1234",
						},
					},
					App: &openrtb2.App{
						Publisher: &openrtb2.Publisher{
							ID: "5890",
						},
					},
				},
			},
			want: want{
				wantErr: false,
				pubID:   1234,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPubID(tt.args.bidRequest)
			if (err != nil) != tt.want.wantErr {
				assert.Equal(t, tt.want.wantErr, err != nil)
				return
			}
			if got != tt.want.pubID {
				assert.Equal(t, tt.want.pubID, got)
			}
		})
	}
}

func TestOpenWrap_applyProfileChanges(t *testing.T) {
	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		rctx       models.RequestCtx
		bidRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *openrtb2.BidRequest
		wantErr bool
	}{
		{
			name: "Request_with_App_object",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 1,
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.AdServerCurrency: "USD",
							models.SChainDBKey:      "1",
						},
					},
					TMax:     500,
					IP:       "127.0.0.1",
					Platform: models.PLATFORM_APP,
					KADUSERCookie: &http.Cookie{
						Name:  "KADUSERCOOKIE",
						Value: "123456789",
					},
				},
				bidRequest: getTestBidRequest(false),
			},
			want: &openrtb2.BidRequest{
				ID:   "testID",
				Test: 1,
				Cur:  []string{"EUR", "USD"},
				TMax: 500,
				Source: &openrtb2.Source{
					TID: "testID",
				},
				Imp: []openrtb2.Imp{
					{
						ID: "testImp1",
						Banner: &openrtb2.Banner{
							W: ptrutil.ToPtr[int64](200),
							H: ptrutil.ToPtr[int64](300),
						},
						Video: &openrtb2.Video{
							W:     ptrutil.ToPtr[int64](200),
							H:     ptrutil.ToPtr[int64](300),
							Plcmt: 1,
						},
					},
				},
				Device: &openrtb2.Device{
					IP:         "127.0.0.1",
					Language:   "en",
					DeviceType: 1,
				},
				WLang: []string{"en", "hi"},
				User: &openrtb2.User{
					CustomData: "123456789",
				},
				App: &openrtb2.App{
					Publisher: &openrtb2.Publisher{
						ID: "1010",
					},
					Content: &openrtb2.Content{
						Language: "en",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Request_with_Site_object",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 1,
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.AdServerCurrency: "USD",
							models.SChainDBKey:      "1",
						},
					},
					TMax:     500,
					IP:       "127.0.0.1",
					Platform: models.PLATFORM_APP,
					KADUSERCookie: &http.Cookie{
						Name:  "KADUSERCOOKIE",
						Value: "123456789",
					},
				},
				bidRequest: getTestBidRequest(true),
			},
			want: &openrtb2.BidRequest{
				ID:   "testID",
				Test: 1,
				Cur:  []string{"EUR", "USD"},
				TMax: 500,
				Source: &openrtb2.Source{
					TID: "testID",
				},
				Imp: []openrtb2.Imp{
					{
						ID: "testImp1",
						Banner: &openrtb2.Banner{
							W: ptrutil.ToPtr[int64](200),
							H: ptrutil.ToPtr[int64](300),
						},
						Video: &openrtb2.Video{
							W:     ptrutil.ToPtr[int64](200),
							H:     ptrutil.ToPtr[int64](300),
							Plcmt: 1,
						},
					},
				},
				Device: &openrtb2.Device{
					IP:         "127.0.0.1",
					Language:   "en",
					DeviceType: 1,
				},
				WLang: []string{"en", "hi"},
				User: &openrtb2.User{
					CustomData: "123456789",
				},
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{
						ID: "1010",
					},
					Content: &openrtb2.Content{
						Language: "en",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "For_amp_request_banner_can_not_be_disabled_through_adunit_config",
			args: args{
				rctx: models.RequestCtx{
					IsTestRequest: 1,
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.AdServerCurrency: "USD",
							models.SChainDBKey:      "1",
						},
					},
					TMax:     500,
					IP:       "127.0.0.1",
					Platform: models.PLATFORM_APP,
					KADUSERCookie: &http.Cookie{
						Name:  "KADUSERCOOKIE",
						Value: "123456789",
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Banner: &adunitconfig.Banner{
										Enabled: ptrutil.ToPtr(false),
									},
								},
							},
						},
					},
					Endpoint: models.EndpointAMP,
				},
				bidRequest: getTestBidRequest(true),
			},
			want: &openrtb2.BidRequest{
				ID:   "testID",
				Test: 1,
				Cur:  []string{"EUR", "USD"},
				TMax: 500,
				Source: &openrtb2.Source{
					TID: "testID",
				},
				Imp: []openrtb2.Imp{
					{
						ID: "testImp1",
						Banner: &openrtb2.Banner{
							W: ptrutil.ToPtr[int64](200),
							H: ptrutil.ToPtr[int64](300),
						},
						Video: &openrtb2.Video{
							W:     ptrutil.ToPtr[int64](200),
							H:     ptrutil.ToPtr[int64](300),
							Plcmt: 1,
						},
					},
				},
				Device: &openrtb2.Device{
					IP:         "127.0.0.1",
					Language:   "en",
					DeviceType: 1,
				},
				WLang: []string{"en", "hi"},
				User: &openrtb2.User{
					CustomData: "123456789",
				},
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{
						ID: "1010",
					},
					Content: &openrtb2.Content{
						Language: "en",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: tt.fields.metricEngine,
			}
			got, err := m.applyProfileChanges(tt.args.rctx, tt.args.bidRequest)
			if (err != nil) != tt.wantErr {
				assert.Equal(t, tt.wantErr, err != nil)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOpenWrap_applyVideoAdUnitConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFeature := mock_feature.NewMockFeature(ctrl)

	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		rCtx models.RequestCtx
		imp  *openrtb2.Imp
	}
	type want struct {
		rCtx models.RequestCtx
		imp  *openrtb2.Imp
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "imp.video_is_nil",
			args: args{
				imp: &openrtb2.Imp{
					Video: nil,
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					Video: nil,
				},
			},
		},
		{
			name: "empty_adunitCfg",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: nil,
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:    "testImp",
					Video: &openrtb2.Video{},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID:    "testImp",
					Video: &openrtb2.Video{},
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: nil,
							},
						},
					},
				},
			},
		},
		{
			name: "imp.BidFloor_and_BidFloorCur_updated_from_adunit_config",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									BidFloor:    ptrutil.ToPtr(2.0),
									BidFloorCur: ptrutil.ToPtr("USD"),
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:          "testImp",
					BidFloor:    0,
					BidFloorCur: "",
					Video:       &openrtb2.Video{},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID:          "testImp",
					Video:       &openrtb2.Video{},
					BidFloor:    2.0,
					BidFloorCur: "USD",
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									BidFloor:    ptrutil.ToPtr(2.0),
									BidFloorCur: ptrutil.ToPtr("USD"),
								},
							},
							BidFloor:    2,
							BidFloorCur: "USD",
						},
					},
				},
			},
		},
		{
			name: "imp.Exp_updated_from_adunit_config",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Exp: ptrutil.ToPtr(10),
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:    "testImp",
					Video: &openrtb2.Video{},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID:    "testImp",
					Video: &openrtb2.Video{},
					Exp:   10,
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Exp: ptrutil.ToPtr(10),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "imp_has_video_object_but_adunitConfig_video_is_nil._imp_video_will_not_be_updated",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: nil,
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID: "testImp",
					Video: &openrtb2.Video{
						W: ptrutil.ToPtr[int64](200),
						H: ptrutil.ToPtr[int64](300),
					},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID: "testImp",
					Video: &openrtb2.Video{
						W: ptrutil.ToPtr[int64](200),
						H: ptrutil.ToPtr[int64](300),
					},
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: nil,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "imp_has_video_object_but_video_is_disabled_from_adunitConfig_then_remove_video_object_from_imp",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(false),
									},
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID: "testImp",
					Video: &openrtb2.Video{
						W: ptrutil.ToPtr[int64](200),
						H: ptrutil.ToPtr[int64](300),
					},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID:    "testImp",
					Video: nil,
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(false),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "imp_has_empty_video_object_and_adunitCofig_for_video_is_enable._all_absent_video_parameters_will_be_updated",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(true),
										Config: &adunitconfig.VideoConfig{
											Video: openrtb2.Video{
												MinDuration:    10,
												MaxDuration:    40,
												Skip:           ptrutil.ToPtr(int8(1)),
												SkipMin:        5,
												SkipAfter:      10,
												Plcmt:          1,
												Placement:      1,
												MinBitRate:     100,
												MaxBitRate:     200,
												MaxExtended:    50,
												Linearity:      1,
												Protocol:       1,
												W:              ptrutil.ToPtr[int64](640),
												H:              ptrutil.ToPtr[int64](480),
												Sequence:       2,
												BoxingAllowed:  ptrutil.ToPtr[int8](1),
												PlaybackEnd:    2,
												MIMEs:          []string{"mimes"},
												API:            []adcom1.APIFramework{1, 2},
												Delivery:       []adcom1.DeliveryMethod{1, 2},
												PlaybackMethod: []adcom1.PlaybackMethod{1, 2},
												BAttr:          []adcom1.CreativeAttribute{1, 2},
												StartDelay:     ptrutil.ToPtr(adcom1.StartDelay(2)),
												Protocols:      []adcom1.MediaCreativeSubtype{1, 2},
												Pos:            ptrutil.ToPtr(adcom1.PlacementPosition(1)),
												CompanionType:  []adcom1.CompanionType{1, 2},
											},
										},
									},
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:    "testImp",
					Video: &openrtb2.Video{},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID: "testImp",
					Video: &openrtb2.Video{
						W:              ptrutil.ToPtr[int64](640),
						H:              ptrutil.ToPtr[int64](480),
						MinDuration:    10,
						MaxDuration:    40,
						Skip:           ptrutil.ToPtr(int8(1)),
						SkipMin:        5,
						SkipAfter:      10,
						Plcmt:          1,
						Placement:      1,
						MinBitRate:     100,
						MaxBitRate:     200,
						MaxExtended:    50,
						Linearity:      1,
						Protocol:       1,
						Sequence:       2,
						BoxingAllowed:  ptrutil.ToPtr[int8](1),
						PlaybackEnd:    2,
						MIMEs:          []string{"mimes"},
						API:            []adcom1.APIFramework{1, 2},
						Delivery:       []adcom1.DeliveryMethod{1, 2},
						PlaybackMethod: []adcom1.PlaybackMethod{1, 2},
						BAttr:          []adcom1.CreativeAttribute{1, 2},
						StartDelay:     ptrutil.ToPtr(adcom1.StartDelay(2)),
						Protocols:      []adcom1.MediaCreativeSubtype{1, 2},
						Pos:            ptrutil.ToPtr(adcom1.PlacementPosition(1)),
						CompanionType:  []adcom1.CompanionType{1, 2},
					},
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(true),
										Config: &adunitconfig.VideoConfig{
											Video: openrtb2.Video{
												MinDuration:    10,
												MaxDuration:    40,
												Skip:           ptrutil.ToPtr(int8(1)),
												SkipMin:        5,
												SkipAfter:      10,
												Plcmt:          1,
												Placement:      1,
												MinBitRate:     100,
												MaxBitRate:     200,
												MaxExtended:    50,
												Linearity:      1,
												Protocol:       1,
												W:              ptrutil.ToPtr[int64](640),
												H:              ptrutil.ToPtr[int64](480),
												Sequence:       2,
												BoxingAllowed:  ptrutil.ToPtr[int8](1),
												PlaybackEnd:    2,
												MIMEs:          []string{"mimes"},
												API:            []adcom1.APIFramework{1, 2},
												Delivery:       []adcom1.DeliveryMethod{1, 2},
												PlaybackMethod: []adcom1.PlaybackMethod{1, 2},
												BAttr:          []adcom1.CreativeAttribute{1, 2},
												StartDelay:     ptrutil.ToPtr(adcom1.StartDelay(2)),
												Protocols:      []adcom1.MediaCreativeSubtype{1, 2},
												Pos:            ptrutil.ToPtr(adcom1.PlacementPosition(1)),
												CompanionType:  []adcom1.CompanionType{1, 2},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "imp_has_video_object_and_adunitConfig_alos_have_parameter_present_then_priority_to_request",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(true),
										Config: &adunitconfig.VideoConfig{
											Video: openrtb2.Video{
												MinDuration: 10,
												MaxDuration: 40,
												Skip:        ptrutil.ToPtr(int8(1)),
												SkipMin:     5,
												SkipAfter:   10,
											},
										},
									},
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID: "testImp",
					Video: &openrtb2.Video{
						W:           ptrutil.ToPtr[int64](640),
						H:           ptrutil.ToPtr[int64](480),
						MinDuration: 20,
						MaxDuration: 60,
						Skip:        ptrutil.ToPtr(int8(2)),
						SkipMin:     10,
						SkipAfter:   20,
					},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID: "testImp",
					Video: &openrtb2.Video{
						W:           ptrutil.ToPtr[int64](640),
						H:           ptrutil.ToPtr[int64](480),
						MinDuration: 20,
						MaxDuration: 60,
						Skip:        ptrutil.ToPtr(int8(2)),
						SkipMin:     10,
						SkipAfter:   20,
					},
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(true),
										Config: &adunitconfig.VideoConfig{
											Video: openrtb2.Video{
												MinDuration: 10,
												MaxDuration: 40,
												Skip:        ptrutil.ToPtr(int8(1)),
												SkipMin:     5,
												SkipAfter:   10,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "imp.video_is_nil_but_AmpVideoEnabled_true_update_and_no_video_config_update_imp.Video_to_default_video",
			args: args{
				imp: &openrtb2.Imp{
					ID: "testImp",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](250),
						H: ptrutil.ToPtr[int64](300),
					},
				},
				rCtx: models.RequestCtx{
					PubID: 5890,
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(true),
										Config: &adunitconfig.VideoConfig{
											Video: openrtb2.Video{
												MIMEs:          []string{"video/mpev"},
												MinDuration:    10,
												MaxDuration:    50,
												StartDelay:     adcom1.StartMidRoll.Ptr(),
												Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper},
												Placement:      adcom1.VideoPlacementInArticle,
												Plcmt:          adcom1.VideoPlcmtInstream,
												Linearity:      adcom1.LinearityNonLinear,
												Skip:           ptrutil.ToPtr[int8](1),
												SkipMin:        1,
												SkipAfter:      1,
												PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackClickSoundOn},
												PlaybackEnd:    adcom1.PlaybackFloating,
												Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryProgressive},
												W:              ptrutil.ToPtr[int64](300),
												H:              ptrutil.ToPtr[int64](400),
											},
										},
									},
								},
							},
						},
					},
					AmpVideoEnabled: true,
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID: "testImp",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](250),
						H: ptrutil.ToPtr[int64](300),
					},
					Video: &openrtb2.Video{
						MIMEs:          []string{"video/mpev"},
						MinDuration:    10,
						MaxDuration:    50,
						StartDelay:     adcom1.StartMidRoll.Ptr(),
						Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper},
						Placement:      adcom1.VideoPlacementInArticle,
						Plcmt:          adcom1.VideoPlcmtInstream,
						Linearity:      adcom1.LinearityNonLinear,
						Skip:           ptrutil.ToPtr[int8](1),
						SkipMin:        1,
						SkipAfter:      1,
						PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackClickSoundOn},
						PlaybackEnd:    adcom1.PlaybackFloating,
						Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryProgressive},
						W:              ptrutil.ToPtr[int64](300),
						H:              ptrutil.ToPtr[int64](400),
					},
				},
				rCtx: models.RequestCtx{
					PubID: 5890,
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(true),
										Config: &adunitconfig.VideoConfig{
											Video: openrtb2.Video{
												MIMEs:          []string{"video/mpev"},
												MinDuration:    10,
												MaxDuration:    50,
												StartDelay:     adcom1.StartMidRoll.Ptr(),
												Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper},
												Placement:      adcom1.VideoPlacementInArticle,
												Plcmt:          adcom1.VideoPlcmtInstream,
												Linearity:      adcom1.LinearityNonLinear,
												Skip:           ptrutil.ToPtr[int8](1),
												SkipMin:        1,
												SkipAfter:      1,
												PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackClickSoundOn},
												PlaybackEnd:    adcom1.PlaybackFloating,
												Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryProgressive},
												W:              ptrutil.ToPtr[int64](300),
												H:              ptrutil.ToPtr[int64](400),
											},
										},
									},
								},
							},
						},
					},
					AmpVideoEnabled: true,
				},
			},
		},
		{
			name: "imp.video_is_nil_but_AmpVideoEnabled_true_update_and_video_config_is_also_non_nil_update_imp.Video_to_video_config",
			args: args{
				imp: &openrtb2.Imp{
					ID: "testImp",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](250),
						H: ptrutil.ToPtr[int64](300),
					},
				},
				rCtx: models.RequestCtx{
					PubID: 5890,
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(true),
									},
								},
							},
						},
					},
					AmpVideoEnabled: true,
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID: "testImp",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](250),
						H: ptrutil.ToPtr[int64](300),
					},
					Video: &openrtb2.Video{
						MIMEs:          []string{"video/mp4"},
						MinDuration:    0,
						MaxDuration:    30,
						StartDelay:     adcom1.StartPreRoll.Ptr(),
						Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper, adcom1.CreativeVAST30Wrapper, adcom1.CreativeVAST40, adcom1.CreativeVAST40Wrapper, adcom1.CreativeVAST41, adcom1.CreativeVAST41Wrapper, adcom1.CreativeVAST42, adcom1.CreativeVAST42Wrapper},
						Placement:      adcom1.VideoPlacementInBanner,
						Plcmt:          adcom1.VideoPlcmtNoContent,
						Linearity:      adcom1.LinearityLinear,
						Skip:           ptrutil.ToPtr[int8](0),
						SkipMin:        0,
						SkipAfter:      0,
						PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackPageLoadSoundOff},
						PlaybackEnd:    adcom1.PlaybackCompletion,
						Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryProgressive, adcom1.DeliveryDownload},
						W:              ptrutil.ToPtr[int64](250),
						H:              ptrutil.ToPtr[int64](300),
					},
				},
				rCtx: models.RequestCtx{
					PubID: 5890,
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Video: &adunitconfig.Video{
										Enabled: ptrutil.ToPtr(true),
									},
								},
							},
						},
					},
					AmpVideoEnabled: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &OpenWrap{
				cfg:           tt.fields.cfg,
				cache:         tt.fields.cache,
				metricEngine:  tt.fields.metricEngine,
				featureConfig: mockFeature,
			}
			m.applyVideoAdUnitConfig(tt.args.rCtx, tt.args.imp)
			assert.Equal(t, tt.args.imp, tt.want.imp, "Imp video is not upadted as expected from adunit config")
			assert.Equal(t, tt.args.rCtx, tt.want.rCtx, "rctx is not upadted as expected from adunit config")
		})
	}
}

func TestOpenWrap_applyBannerAdUnitConfig(t *testing.T) {
	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		rCtx models.RequestCtx
		imp  *openrtb2.Imp
	}
	type want struct {
		rCtx models.RequestCtx
		imp  *openrtb2.Imp
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "imp.banner_is_nil",
			args: args{
				imp: &openrtb2.Imp{
					Banner: nil,
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					Banner: nil,
				},
			},
		},
		{
			name: "empty_adunitCfg",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: nil,
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:     "testImp",
					Banner: &openrtb2.Banner{},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID:     "testImp",
					Banner: &openrtb2.Banner{},
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: nil,
							},
						},
					},
				},
			},
		},
		{
			name: "imp.BidFloor_and_BidFloorCur_updated_from_adunit_config",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									BidFloor:    ptrutil.ToPtr(2.0),
									BidFloorCur: ptrutil.ToPtr("USD"),
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:          "testImp",
					BidFloor:    0,
					BidFloorCur: "",
					Banner:      &openrtb2.Banner{},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID:          "testImp",
					Banner:      &openrtb2.Banner{},
					BidFloor:    2.0,
					BidFloorCur: "USD",
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									BidFloor:    ptrutil.ToPtr(2.0),
									BidFloorCur: ptrutil.ToPtr("USD"),
								},
							},
							BidFloor:    2,
							BidFloorCur: "USD",
						},
					},
				},
			},
		},
		{
			name: "imp.Exp_updated_from_adunit_config",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Exp: ptrutil.ToPtr(10),
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID:     "testImp",
					Banner: &openrtb2.Banner{},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID:     "testImp",
					Banner: &openrtb2.Banner{},
					Exp:    10,
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Exp: ptrutil.ToPtr(10),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "imp_has_banner_object_but_adunitConfig_banner_is_nil._imp_banner_will_not_be_updated",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Banner: nil,
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID: "testImp",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](200),
						H: ptrutil.ToPtr[int64](300),
					},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID: "testImp",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](200),
						H: ptrutil.ToPtr[int64](300),
					},
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Banner: nil,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "imp_has_banner_object_but_banner_is_disabled_from_adunitConfig_then_remove_banner_object_from_imp",
			args: args{
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Banner: &adunitconfig.Banner{
										Enabled: ptrutil.ToPtr(false),
									},
								},
							},
						},
					},
				},
				imp: &openrtb2.Imp{
					ID: "testImp",
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](200),
						H: ptrutil.ToPtr[int64](300),
					},
				},
			},
			want: want{
				imp: &openrtb2.Imp{
					ID:     "testImp",
					Banner: nil,
				},
				rCtx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"testImp": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Banner: &adunitconfig.Banner{
										Enabled: ptrutil.ToPtr(false),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: tt.fields.metricEngine,
			}
			m.applyBannerAdUnitConfig(tt.args.rCtx, tt.args.imp)
			assert.Equal(t, tt.args.imp, tt.want.imp, "Imp banner is not upadted as expected from adunit config")
			assert.Equal(t, tt.args.rCtx, tt.want.rCtx, "rctx is not upadted as expected from adunit config")
		})
	}
}

func TestGetDomainFromUrl(t *testing.T) {
	type args struct {
		pageUrl string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test_case_1",
			args: args{
				pageUrl: "http://ebay.com/inte/automation/s2s/pwt_parameter_validation_muti_slot_multi_size.html?pwtvc=1&pwtv=1&profileid=3277",
			},
			want: "ebay.com",
		},
		{
			name: "test_case_2",
			args: args{
				pageUrl: "http://ebay.co.in/inte/automation/s2s/pwt_parameter_validation_muti_slot_multi_size.html?pwtvc=1&pwtv=1&profileid=3277",
			},
			want: "ebay.co.in",
		},
		{
			name: "test_case_3",
			args: args{
				pageUrl: "site@sit.com",
			},
			want: "",
		},
		{
			name: "test_case_4",
			args: args{
				pageUrl: " 12 44",
			},
			want: "",
		},
		{
			name: "test_case_5",
			args: args{
				pageUrl: " ",
			},
			want: "",
		},
		{
			name: "test_case_6",
			args: args{
				pageUrl: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDomainFromUrl(tt.args.pageUrl); got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUpdateRequestExtBidderParamsPubmatic(t *testing.T) {
	type args struct {
		bidderParams json.RawMessage
		cookie       string
		loggerID     string
		bidderCode   string
	}
	tests := []struct {
		name    string
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name: "empty_cookie",
			args: args{
				bidderParams: json.RawMessage(`{"pubmatic":{"pmzoneid":"zone1","adSlot":"38519891"}}`),
				loggerID:     "b441a46e-8c1f-428b-9c29-44e2a408a954",
				bidderCode:   "pubmatic",
			},
			want:    json.RawMessage(`{"pubmatic":{"wiid":"b441a46e-8c1f-428b-9c29-44e2a408a954"}}`),
			wantErr: false,
		},
		{
			name: "empty_loggerID",
			args: args{
				bidderParams: json.RawMessage(`{"pubmatic":{"pmzoneid":"zone1","adSlot":"38519891"}}`),
				cookie:       "test_cookie",
				bidderCode:   "pubmatic",
			},
			want: json.RawMessage(`{"pubmatic":{"Cookie":"test_cookie","wiid":""}}`),
		},
		{
			name: "both_cookie_and_loogerID_are_empty",
			args: args{
				bidderParams: json.RawMessage(`{"pubmatic":{"pmzoneid":"zone1","adSlot":"38519891"}}`),
				cookie:       "",
				loggerID:     "",
				bidderCode:   "pubmatic",
			},
			want: json.RawMessage(`{"pubmatic":{"wiid":""}}`),
		},
		{
			name: "both_cookie_and_loogerID_are_present",
			args: args{
				bidderParams: json.RawMessage(`{"pubmatic":{"pmzoneid":"zone1","adSlot":"38519891"}}`),
				cookie:       "test_cookie",
				loggerID:     "b441a46e-8c1f-428b-9c29-44e2a408a954",
				bidderCode:   "pubmatic",
			},
			want: json.RawMessage(`{"pubmatic":{"Cookie":"test_cookie","wiid":"b441a46e-8c1f-428b-9c29-44e2a408a954"}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := updateRequestExtBidderParamsPubmatic(tt.args.bidderParams, tt.args.cookie, tt.args.loggerID, tt.args.bidderCode)
			if (err != nil) != tt.wantErr {
				assert.Equal(t, tt.wantErr, err != nil)
				return
			}
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestOpenWrapHandleBeforeValidationHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	mockFeature := mock_feature.NewMockFeature(ctrl)
	adapters.InitBidders("./static/bidder-params/")

	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		ctx        context.Context
		moduleCtx  hookstage.ModuleInvocationContext
		payload    hookstage.BeforeValidationRequestPayload
		bidrequest json.RawMessage
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		want           hookstage.HookResult[hookstage.BeforeValidationRequestPayload]
		setup          func()
		wantErr        bool
		wantBidRequest json.RawMessage
		doMutate       bool
	}{
		{
			name: "request_with_sshb=1",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							Sshb: "1",
						},
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject: false,
			},
			wantErr: false,
		},
		{
			name: "empty_module_context",
			args: args{
				ctx:        context.Background(),
				moduleCtx:  hookstage.ModuleInvocationContext{},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:        true,
				DebugMessages: []string{"error: module-ctx not found in handleBeforeValidationHook()"},
			},
			wantErr: false,
		},
		{
			name: "rctx_is_not_present",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"test_rctx": "test",
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:        true,
				DebugMessages: []string{"error: request-ctx not found in handleBeforeValidationHook()"},
			},
			wantErr: false,
		},
		{
			name: "hybrid_request_module_should_not_reject_request_and_return_without_executing_module",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							Endpoint: models.EndpointHybrid,
						},
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject: false,
			},
			wantErr: false,
		},
		{
			name: "Invalid_PubID_in_request",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"test"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidPublisherID))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("", int(nbr.InvalidPublisherID))
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:        true,
				NbrCode:       int(nbr.InvalidPublisherID),
				Errors:        []string{"ErrInvalidPublisherID"},
				DebugMessages: nil,
			},
			wantErr: true,
		},
		{
			name: "Invalid_request_ext",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":1}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidRequestExt))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InvalidRequestExt))
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: int(nbr.InvalidRequestExt),
				Errors:  []string{"failed to get request ext: failed to decode request.ext : json: cannot unmarshal number into Go value of type models.RequestExt"},
			},
			wantErr: true,
		},
		{
			name: "Error_in_getting_profile_data",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.PLATFORM_KEY:     models.PLATFORM_APP,
						models.DisplayVersionID: "1",
					},
				}, errors.New("test"))
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidProfileConfiguration))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InvalidProfileConfiguration))
				mockEngine.EXPECT().RecordPublisherInvalidProfileRequests(rctx.Endpoint, "5890", rctx.ProfileIDStr)
				mockEngine.EXPECT().RecordPublisherInvalidProfileImpressions("5890", rctx.ProfileIDStr, gomock.Any())
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)

			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: int(nbr.InvalidProfileConfiguration),
				Errors:  []string{"failed to get profile data: test"},
			},
			wantErr: true,
		},
		{
			name: "got_empty_profileData_from_DB",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{}, nil)
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidProfileConfiguration))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InvalidProfileConfiguration))
				mockEngine.EXPECT().RecordPublisherInvalidProfileRequests(rctx.Endpoint, "5890", rctx.ProfileIDStr)
				mockEngine.EXPECT().RecordPublisherInvalidProfileImpressions("5890", rctx.ProfileIDStr, gomock.Any())
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)

			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: int(nbr.InvalidProfileConfiguration),
				Errors:  []string{"failed to get profile data: received empty data"},
			},
			wantErr: true,
		},
		{
			name: "platform_is_not_present_in_request_then_reject_the_request",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
					},
				}, nil)
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidPlatform))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InvalidPlatform))
				mockEngine.EXPECT().RecordPublisherInvalidProfileRequests(rctx.Endpoint, "5890", rctx.ProfileIDStr)
				mockEngine.EXPECT().RecordPublisherInvalidProfileImpressions("5890", rctx.ProfileIDStr, gomock.Any())
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)

			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: int(nbr.InvalidPlatform),
				Errors:  []string{"failed to get platform data"},
			},
			wantErr: true,
		},
		{
			name: "All_partners_throttled",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"","ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
						models.THROTTLE:            "0",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.AllPartnerThrottled))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.AllPartnerThrottled))
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: int(nbr.AllPartnerThrottled),
				Errors:  []string{"All adapters throttled"},
			},
			wantErr: false,
		},
		{
			name: "TagID_not_present_in_imp",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidImpressionTagID))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InvalidImpressionTagID))
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: int(nbr.InvalidImpressionTagID),
				Errors:  []string{"tagid missing for imp: 123"},
			},
			wantErr: true,
		},
		{
			name: "TagID_not_present_in_imp_and_not_found_for_client_request",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": func() models.RequestCtx {
							testRctx := rctx
							testRctx.Endpoint = models.EndpointWebS2S
							return testRctx
						}(),
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(models.EndpointWebS2S, getPubmaticErrorCode(nbr.InvalidImpressionTagID))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InvalidImpressionTagID))
				mockEngine.EXPECT().RecordPublisherRequests(models.EndpointWebS2S, "5890", rctx.Platform)
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: int(nbr.InvalidImpressionTagID),
				Errors:  []string{"tagid missing for imp: 123"},
			},
			wantErr: true,
		},
		{
			name: "invalid_impExt",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":"1"}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InternalError))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InternalError))
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: int(nbr.InternalError),
				Errors:  []string{"failed to parse imp.ext: 123"},
			},
			wantErr: true,
		},
		{
			name: "allSotsDisabled-native-not-present",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							ProfileID:                 1234,
							DisplayID:                 1,
							SSAuction:                 -1,
							Platform:                  "in-app",
							Debug:                     true,
							UA:                        "go-test",
							IP:                        "127.0.0.1",
							IsCTVRequest:              false,
							TrackerEndpoint:           "t.pubmatic.com",
							VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
							UidCookie: &http.Cookie{
								Name:  "uids",
								Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
							},
							KADUSERCookie: &http.Cookie{
								Name:  "KADUSERCOOKIE",
								Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
							},
							OriginCookie:             "go-test",
							Aliases:                  make(map[string]string),
							ImpBidCtx:                make(map[string]models.ImpCtx),
							PrebidBidderCode:         make(map[string]string),
							BidderResponseTimeMillis: make(map[string]int),
							ProfileIDStr:             "1234",
							Endpoint:                 models.EndpointV25,
							SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
							MetricsEngine:            mockEngine,
						},
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{
					ConfigPattern: "_AU_@_W_x_H_",
					Config: map[string]*adunitconfig.AdConfig{
						"adunit@700x900": {
							Banner: &adunitconfig.Banner{
								Enabled: ptrutil.ToPtr(false),
							},
						},
						"adunit@640x480": {
							Video: &adunitconfig.Video{
								Enabled: ptrutil.ToPtr(false),
							},
						},
					},
				})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.AllSlotsDisabled))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.AllSlotsDisabled))
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeVideo, "5890", "1234")
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeBanner, "5890", "1234")
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: int(nbr.AllSlotsDisabled),
				Errors:  []string{"All slots disabled"},
			},
			wantErr: false,
		},
		{
			name: "allSotsDisabled-native-present",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							ProfileID:                 1234,
							DisplayID:                 1,
							SSAuction:                 -1,
							Platform:                  "in-app",
							Debug:                     true,
							UA:                        "go-test",
							IP:                        "127.0.0.1",
							IsCTVRequest:              false,
							TrackerEndpoint:           "t.pubmatic.com",
							VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
							UidCookie: &http.Cookie{
								Name:  "uids",
								Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
							},
							KADUSERCookie: &http.Cookie{
								Name:  "KADUSERCOOKIE",
								Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
							},
							OriginCookie:             "go-test",
							Aliases:                  make(map[string]string),
							ImpBidCtx:                make(map[string]models.ImpCtx),
							PrebidBidderCode:         make(map[string]string),
							BidderResponseTimeMillis: make(map[string]int),
							ProfileIDStr:             "1234",
							Endpoint:                 models.EndpointV25,
							SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
							MetricsEngine:            mockEngine,
						},
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","native": {},"banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{
					ConfigPattern: "_AU_@_W_x_H_",
					Config: map[string]*adunitconfig.AdConfig{
						"adunit@700x900": {
							Banner: &adunitconfig.Banner{
								Enabled: ptrutil.ToPtr(false),
							},
						},
						"adunit@640x480": {
							Video: &adunitconfig.Video{
								Enabled: ptrutil.ToPtr(false),
							},
						},
					},
				})
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"adunit@700x900": {
						SlotName: "adunit@700x900",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@700x900"},
					HashValueMap: map[string]string{
						"adunit@700x900": "1232433543534543",
					},
				})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(rctx.Platform, "5890", "appnexus")
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeVideo, "5890", "1234")
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeBanner, "5890", "1234")
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:        false,
				NbrCode:       0,
				Message:       "",
				ChangeSet:     hookstage.ChangeSet[hookstage.BeforeValidationRequestPayload]{},
				DebugMessages: []string{`new imp: {"123":{"ImpID":"123","TagID":"adunit","Div":"","SlotName":"adunit","AdUnitName":"adunit","Secure":0,"BidFloor":0,"BidFloorCur":"","IsRewardInventory":null,"Banner":true,"Video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"Native":{"request":""},"IncomingSlots":["640x480v","700x900","728x90","300x250"],"Type":"video","Bidders":{"appnexus":{"PartnerID":2,"PrebidBidderCode":"appnexus","MatchedSlot":"adunit@700x900","KGP":"_AU_@_W_x_H_","KGPV":"","IsRegex":false,"Params":{"placementId":0,"site":"12313","adtag":"45343"},"VASTTagFlag":false,"VASTTagFlags":null}},"NonMapped":{},"NewExt":{"data":{"pbadslot":"adunit"},"prebid":{"bidder":{"appnexus":{"placementId":0,"site":"12313","adtag":"45343"}}}},"BidCtx":{},"BannerAdUnitCtx":{"MatchedSlot":"adunit@700x900","IsRegex":false,"MatchedRegex":"","SelectedSlotAdUnitConfig":{"banner":{"enabled":false}},"AppliedSlotAdUnitConfig":{"banner":{"enabled":false}},"UsingDefaultConfig":false,"AllowedConnectionTypes":null},"VideoAdUnitCtx":{"MatchedSlot":"adunit@640x480","IsRegex":false,"MatchedRegex":"","SelectedSlotAdUnitConfig":{"video":{"enabled":false}},"AppliedSlotAdUnitConfig":{"video":{"enabled":false}},"UsingDefaultConfig":false,"AllowedConnectionTypes":null},"BidderError":"","IsAdPodRequest":false}}`, `new request.ext: {"prebid":{"bidadjustmentfactors":{"appnexus":1},"bidderparams":{"pubmatic":{"wiid":""}},"debug":true,"floors":{"enforcement":{"enforcepbs":true},"enabled":true},"targeting":{"pricegranularity":{"precision":2,"ranges":[{"min":0,"max":5,"increment":0.05},{"min":5,"max":10,"increment":0.1},{"min":10,"max":20,"increment":0.5}]},"mediatypepricegranularity":{},"includewinners":true,"includebidderkeys":true},"macros":{"[PLATFORM]":"3","[PROFILE_ID]":"1234","[PROFILE_VERSION]":"1","[UNIX_TIMESTAMP]":"0","[WRAPPER_IMPRESSION_ID]":""}}}`},
				AnalyticsTags: hookanalytics.Analytics{Activities: nil},
			},
			wantErr:        false,
			doMutate:       true,
			wantBidRequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","native":{"request":""},"tagid":"adunit","ext":{"data":{"pbadslot":"adunit"},"prebid":{"bidder":{"appnexus":{"placementId":0,"site":"12313","adtag":"45343"}}}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","customdata":"7D75D25F-FAC9-443D-B2D1-B17FEE11E027","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"tid":"123-456-789","ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{"bidadjustmentfactors":{"appnexus":1},"bidderparams":{"pubmatic":{"wiid":""}},"debug":true,"floors":{"enforcement":{"enforcepbs":true},"enabled":true},"targeting":{"pricegranularity":{"precision":2,"ranges":[{"min":0,"max":5,"increment":0.05},{"min":5,"max":10,"increment":0.1},{"min":10,"max":20,"increment":0.5}]},"mediatypepricegranularity":{},"includewinners":true,"includebidderkeys":true},"macros":{"[PLATFORM]":"3","[PROFILE_ID]":"1234","[PROFILE_VERSION]":"1","[UNIX_TIMESTAMP]":"0","[WRAPPER_IMPRESSION_ID]":""}}}}`),
		},
		{
			name: "no_serviceSideBidderPresent",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							ProfileID:                 1234,
							DisplayID:                 1,
							SSAuction:                 -1,
							Platform:                  "in-app",
							Debug:                     true,
							UA:                        "go-test",
							IP:                        "127.0.0.1",
							IsCTVRequest:              false,
							TrackerEndpoint:           "t.pubmatic.com",
							VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
							UidCookie: &http.Cookie{
								Name:  "uids",
								Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
							},
							KADUSERCookie: &http.Cookie{
								Name:  "KADUSERCOOKIE",
								Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
							},
							OriginCookie:             "go-test",
							Aliases:                  make(map[string]string),
							ImpBidCtx:                make(map[string]models.ImpCtx),
							PrebidBidderCode:         make(map[string]string),
							BidderResponseTimeMillis: make(map[string]int),
							ProfileIDStr:             "1234",
							Endpoint:                 models.EndpointV25,
							SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
							MetricsEngine:            mockEngine,
						},
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.ServerSidePartnerNotConfigured))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.ServerSidePartnerNotConfigured))
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)

			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: int(nbr.ServerSidePartnerNotConfigured),
				Errors:  []string{"server side partner not found"},
			},
			wantErr: false,
		},
		{
			name: "if_partner_is_alias_update_req.ext.prebid.aliasgvlid",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							ProfileID:                 1234,
							DisplayID:                 1,
							SSAuction:                 -1,
							Platform:                  "in-app",
							Debug:                     true,
							UA:                        "go-test",
							IP:                        "127.0.0.1",
							IsCTVRequest:              false,
							TrackerEndpoint:           "t.pubmatic.com",
							VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
							UidCookie: &http.Cookie{
								Name:  "uids",
								Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
							},
							KADUSERCookie: &http.Cookie{
								Name:  "KADUSERCOOKIE",
								Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
							},
							OriginCookie:             "go-test",
							Aliases:                  make(map[string]string),
							ImpBidCtx:                make(map[string]models.ImpCtx),
							PrebidBidderCode:         make(map[string]string),
							BidderResponseTimeMillis: make(map[string]int),
							ProfileIDStr:             "1234",
							Endpoint:                 models.EndpointV25,
							SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
							MetricsEngine:            mockEngine,
						},
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"adunit@700x900": {
						SlotName: "adunit@700x900",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				}).Times(3)
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@700x900"},
					HashValueMap: map[string]string{
						"adunit@700x900": "1232433543534543",
					},
				}).Times(3)
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					1: {
						models.PARTNER_ID:          "1",
						models.PREBID_PARTNER_NAME: "pubmatic2",
						models.BidderCode:          "pub2-alias",
						models.IsAlias:             "1",
						models.TIMEOUT:             "200",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.SERVER_SIDE_FLAG:    "1",
						models.VENDORID:            "130",
					},
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
						models.VENDORID:            "100",
					},
					3: {
						models.PARTNER_ID:          "3",
						models.PREBID_PARTNER_NAME: "districtm",
						models.BidderCode:          "dm-alias",
						models.IsAlias:             "1",
						models.TIMEOUT:             "200",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.SERVER_SIDE_FLAG:    "1",
						models.VENDORID:            "99",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(rctx.Platform, "5890", "pub2-alias")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(rctx.Platform, "5890", "appnexus")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(rctx.Platform, "5890", "dm-alias")
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:        false,
				NbrCode:       0,
				ChangeSet:     hookstage.ChangeSet[hookstage.BeforeValidationRequestPayload]{},
				DebugMessages: []string{`new imp: {"123":{"ImpID":"123","TagID":"adunit","Div":"","SlotName":"adunit","AdUnitName":"adunit","Secure":0,"BidFloor":4.3,"BidFloorCur":"USD","IsRewardInventory":null,"Banner":true,"Video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"Native":null,"IncomingSlots":["640x480v","700x900","728x90","300x250"],"Type":"video","Bidders":{"appnexus":{"PartnerID":2,"PrebidBidderCode":"appnexus","MatchedSlot":"adunit@700x900","KGP":"_AU_@_W_x_H_","KGPV":"","IsRegex":false,"Params":{"placementId":0,"adtag":"45343","site":"12313"},"VASTTagFlag":false,"VASTTagFlags":null},"dm-alias":{"PartnerID":3,"PrebidBidderCode":"districtm","MatchedSlot":"adunit@700x900","KGP":"_AU_@_W_x_H_","KGPV":"","IsRegex":false,"Params":{"placementId":0,"site":"12313","adtag":"45343"},"VASTTagFlag":false,"VASTTagFlags":null},"pub2-alias":{"PartnerID":1,"PrebidBidderCode":"pubmatic2","MatchedSlot":"adunit@700x900","KGP":"_AU_@_W_x_H_","KGPV":"","IsRegex":false,"Params":{"publisherId":"5890","adSlot":"adunit@700x900","wrapper":{"version":1,"profile":1234}},"VASTTagFlag":false,"VASTTagFlags":null}},"NonMapped":{},"NewExt":{"data":{"pbadslot":"adunit"},"prebid":{"bidder":{"appnexus":{"placementId":0,"adtag":"45343","site":"12313"},"dm-alias":{"placementId":0,"site":"12313","adtag":"45343"},"pub2-alias":{"publisherId":"5890","adSlot":"adunit@700x900","wrapper":{"version":1,"profile":1234}}}}},"BidCtx":{},"BannerAdUnitCtx":{"MatchedSlot":"","IsRegex":false,"MatchedRegex":"","SelectedSlotAdUnitConfig":null,"AppliedSlotAdUnitConfig":null,"UsingDefaultConfig":false,"AllowedConnectionTypes":null},"VideoAdUnitCtx":{"MatchedSlot":"","IsRegex":false,"MatchedRegex":"","SelectedSlotAdUnitConfig":null,"AppliedSlotAdUnitConfig":null,"UsingDefaultConfig":false,"AllowedConnectionTypes":null},"BidderError":"","IsAdPodRequest":false}}`, `new request.ext: {"prebid":{"aliases":{"dm-alias":"appnexus","pub2-alias":"pubmatic"},"aliasgvlids":{"dm-alias":99,"pub2-alias":130},"bidadjustmentfactors":{"appnexus":1,"dm-alias":1,"pub2-alias":1},"bidderparams":{"pubmatic":{"wiid":""}},"debug":true,"floors":{"enforcement":{"enforcepbs":true},"enabled":true},"targeting":{"pricegranularity":{"precision":2,"ranges":[{"min":0,"max":5,"increment":0.05},{"min":5,"max":10,"increment":0.1},{"min":10,"max":20,"increment":0.5}]},"mediatypepricegranularity":{},"includewinners":true,"includebidderkeys":true},"macros":{"[PLATFORM]":"3","[PROFILE_ID]":"1234","[PROFILE_VERSION]":"1","[UNIX_TIMESTAMP]":"0","[WRAPPER_IMPRESSION_ID]":""}}}`},
				AnalyticsTags: hookanalytics.Analytics{},
			},
			wantBidRequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"data":{"pbadslot":"adunit"},"prebid":{"bidder":{"appnexus":{"placementId":0,"adtag":"45343","site":"12313"},"dm-alias":{"placementId":0,"site":"12313","adtag":"45343"},"pub2-alias":{"publisherId":"5890","adSlot":"adunit@700x900","wrapper":{"version":1,"profile":1234}}}}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","customdata":"7D75D25F-FAC9-443D-B2D1-B17FEE11E027","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"tid":"123-456-789","ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{"aliases":{"dm-alias":"appnexus","pub2-alias":"pubmatic"},"aliasgvlids":{"dm-alias":99,"pub2-alias":130},"bidadjustmentfactors":{"appnexus":1,"dm-alias":1,"pub2-alias":1},"bidderparams":{"pubmatic":{"wiid":""}},"debug":true,"floors":{"enforcement":{"enforcepbs":true},"enabled":true},"targeting":{"pricegranularity":{"precision":2,"ranges":[{"min":0,"max":5,"increment":0.05},{"min":5,"max":10,"increment":0.1},{"min":10,"max":20,"increment":0.5}]},"mediatypepricegranularity":{},"includewinners":true,"includebidderkeys":true},"macros":{"[PLATFORM]":"3","[PROFILE_ID]":"1234","[PROFILE_VERSION]":"1","[UNIX_TIMESTAMP]":"0","[WRAPPER_IMPRESSION_ID]":""}}}}`),
			wantErr:        false,
			doMutate:       true,
		},
		{
			name: "happy_path_request_not_rejected_and_successfully_updted_from_DB",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							ProfileID:                 1234,
							DisplayID:                 1,
							SSAuction:                 -1,
							Platform:                  "in-app",
							Debug:                     true,
							UA:                        "go-test",
							IP:                        "127.0.0.1",
							IsCTVRequest:              false,
							TrackerEndpoint:           "t.pubmatic.com",
							VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
							UidCookie: &http.Cookie{
								Name:  "uids",
								Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
							},
							KADUSERCookie: &http.Cookie{
								Name:  "KADUSERCOOKIE",
								Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
							},
							OriginCookie:             "go-test",
							Aliases:                  make(map[string]string),
							ImpBidCtx:                make(map[string]models.ImpCtx),
							PrebidBidderCode:         make(map[string]string),
							BidderResponseTimeMillis: make(map[string]int),
							ProfileIDStr:             "1234",
							Endpoint:                 models.EndpointV25,
							SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
							MetricsEngine:            mockEngine,
						},
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"adunit@700x900": {
						SlotName: "adunit@700x900",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@700x900"},
					HashValueMap: map[string]string{
						"adunit@700x900": "1232433543534543",
					},
				})
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(rctx.Platform, "5890", "appnexus")
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:        false,
				NbrCode:       0,
				ChangeSet:     hookstage.ChangeSet[hookstage.BeforeValidationRequestPayload]{},
				DebugMessages: []string{`new imp: {"123":{"ImpID":"123","TagID":"adunit","Div":"","SlotName":"adunit","AdUnitName":"adunit","Secure":0,"BidFloor":4.3,"BidFloorCur":"USD","IsRewardInventory":null,"Banner":true,"Video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"Native":null,"IncomingSlots":["300x250","640x480v","700x900","728x90"],"Type":"video","Bidders":{"appnexus":{"PartnerID":2,"PrebidBidderCode":"appnexus","MatchedSlot":"adunit@700x900","KGP":"_AU_@_W_x_H_","KGPV":"","IsRegex":false,"Params":{"placementId":0,"site":"12313","adtag":"45343"},"VASTTagFlag":false,"VASTTagFlags":null}},"NonMapped":{},"NewExt":{"data":{"pbadslot":"adunit"},"prebid":{"bidder":{"appnexus":{"placementId":0,"site":"12313","adtag":"45343"}}}},"BidCtx":{},"BannerAdUnitCtx":{"MatchedSlot":"","IsRegex":false,"MatchedRegex":"","SelectedSlotAdUnitConfig":null,"AppliedSlotAdUnitConfig":null,"UsingDefaultConfig":false,"AllowedConnectionTypes":null},"VideoAdUnitCtx":{"MatchedSlot":"","IsRegex":false,"MatchedRegex":"","SelectedSlotAdUnitConfig":null,"AppliedSlotAdUnitConfig":null,"UsingDefaultConfig":false,"AllowedConnectionTypes":null},"BidderError":"","IsAdPodRequest":false}}`, `new request.ext: {"prebid":{"bidadjustmentfactors":{"appnexus":1},"bidderparams":{"pubmatic":{"wiid":""}},"debug":true,"floors":{"enforcement":{"enforcepbs":true},"enabled":true},"targeting":{"pricegranularity":{"precision":2,"ranges":[{"min":0,"max":5,"increment":0.05},{"min":5,"max":10,"increment":0.1},{"min":10,"max":20,"increment":0.5}]},"mediatypepricegranularity":{},"includewinners":true,"includebidderkeys":true},"macros":{"[PLATFORM]":"3","[PROFILE_ID]":"1234","[PROFILE_VERSION]":"1","[UNIX_TIMESTAMP]":"0","[WRAPPER_IMPRESSION_ID]":""}}}`},
				AnalyticsTags: hookanalytics.Analytics{},
			},
			wantBidRequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"data":{"pbadslot":"adunit"},"prebid":{"bidder":{"appnexus":{"placementId":0,"site":"12313","adtag":"45343"}}}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","customdata":"7D75D25F-FAC9-443D-B2D1-B17FEE11E027","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"tid":"123-456-789","ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{"bidadjustmentfactors":{"appnexus":1},"bidderparams":{"pubmatic":{"wiid":""}},"debug":true,"floors":{"enforcement":{"enforcepbs":true},"enabled":true},"targeting":{"pricegranularity":{"precision":2,"ranges":[{"min":0,"max":5,"increment":0.05},{"min":5,"max":10,"increment":0.1},{"min":10,"max":20,"increment":0.5}]},"mediatypepricegranularity":{},"includewinners":true,"includebidderkeys":true},"macros":{"[PLATFORM]":"3","[PROFILE_ID]":"1234","[PROFILE_VERSION]":"1","[UNIX_TIMESTAMP]":"0","[WRAPPER_IMPRESSION_ID]":""}}}}`),
			wantErr:        false,
			doMutate:       true,
		},
		{
			name: "prebid-validation-errors-imp-missing",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							PubIDStr: "1234",
							Endpoint: models.EndpointV25,
						},
					},
				},
				bidrequest: json.RawMessage(`{}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockEngine.EXPECT().RecordBadRequests(models.EndpointV25, 18)
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("1234", 604)
			},
			want:    hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{},
			wantErr: false,
		},
		{
			name: "prebid-validation-errors-site-and-app-missing",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							PubIDStr: "1234",
							Endpoint: models.EndpointV25,
						},
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockEngine.EXPECT().RecordBadRequests(models.EndpointV25, 18)
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("1234", 604)
			},
			want:    hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{},
			wantErr: false,
		},
		{
			name: "AMP_request_successfully_update_video_object_from_adunit_config_and_updated_remaining_feilds_from_default",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							ProfileID:                 1234,
							PubID:                     5890,
							DisplayID:                 1,
							SSAuction:                 -1,
							Platform:                  "amp",
							Debug:                     true,
							UA:                        "go-test",
							IP:                        "127.0.0.1",
							IsCTVRequest:              false,
							TrackerEndpoint:           "t.pubmatic.com",
							VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
							UidCookie: &http.Cookie{
								Name:  "uids",
								Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
							},
							KADUSERCookie: &http.Cookie{
								Name:  "KADUSERCOOKIE",
								Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
							},
							OriginCookie:             "go-test",
							Aliases:                  make(map[string]string),
							ImpBidCtx:                make(map[string]models.ImpCtx),
							PrebidBidderCode:         make(map[string]string),
							BidderResponseTimeMillis: make(map[string]int),
							ProfileIDStr:             "1234",
							Endpoint:                 models.EndpointAMP,
							SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
							MetricsEngine:            mockEngine,
						},
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"adunit@700x900": {
						SlotName: "adunit@700x900",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@700x900"},
					HashValueMap: map[string]string{
						"adunit@700x900": "1232433543534543",
					},
				})
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_AMP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{
					ConfigPattern: "_AU_",
					Config: map[string]*adunitconfig.AdConfig{
						"adunit": {
							Video: &adunitconfig.Video{
								Enabled:              ptrutil.ToPtr(true),
								AmpTrafficPercentage: ptrutil.ToPtr(100),
								Config: &adunitconfig.VideoConfig{
									Video: openrtb2.Video{
										MIMEs: []string{"video/mp4", "video/mpeg"},
										W:     ptrutil.ToPtr[int64](640),
										H:     ptrutil.ToPtr[int64](480),
									},
								},
							},
						},
					},
				})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordPublisherRequests(models.EndpointAMP, "5890", "amp")
				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats("amp", "5890", "appnexus")
				mockFeature.EXPECT().IsAmpMultiformatEnabled(5890).Return(true)
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:        false,
				NbrCode:       0,
				ChangeSet:     hookstage.ChangeSet[hookstage.BeforeValidationRequestPayload]{},
				DebugMessages: []string{`new imp: {"123":{"ImpID":"123","TagID":"adunit","Div":"","SlotName":"adunit","AdUnitName":"adunit","Secure":0,"BidFloor":4.3,"BidFloorCur":"USD","IsRewardInventory":null,"Banner":true,"Video":{"mimes":null},"Native":null,"IncomingSlots":["700x900","728x90","300x250"],"Type":"video","Bidders":{"appnexus":{"PartnerID":2,"PrebidBidderCode":"appnexus","MatchedSlot":"adunit@700x900","KGP":"_AU_@_W_x_H_","KGPV":"","IsRegex":false,"Params":{"placementId":0,"adtag":"45343","site":"12313"},"VASTTagFlag":false,"VASTTagFlags":null}},"NonMapped":{},"NewExt":{"data":{"pbadslot":"adunit"},"prebid":{"bidder":{"appnexus":{"placementId":0,"adtag":"45343","site":"12313"}}}},"BidCtx":{},"BannerAdUnitCtx":{"MatchedSlot":"","IsRegex":false,"MatchedRegex":"","SelectedSlotAdUnitConfig":null,"AppliedSlotAdUnitConfig":null,"UsingDefaultConfig":false,"AllowedConnectionTypes":null},"VideoAdUnitCtx":{"MatchedSlot":"adunit","IsRegex":false,"MatchedRegex":"","SelectedSlotAdUnitConfig":{"video":{"enabled":true,"amptrafficpercentage":100,"config":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480}}},"AppliedSlotAdUnitConfig":{"video":{"enabled":true,"amptrafficpercentage":100,"config":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480}}},"UsingDefaultConfig":false,"AllowedConnectionTypes":null},"BidderError":"","IsAdPodRequest":false}}`, `new request.ext: {"prebid":{"bidadjustmentfactors":{"appnexus":1},"bidderparams":{"pubmatic":{"wiid":""}},"debug":true,"floors":{"enforcement":{"enforcepbs":true},"enabled":true},"targeting":{"pricegranularity":{"precision":2,"ranges":[{"min":0,"max":5,"increment":0.05},{"min":5,"max":10,"increment":0.1},{"min":10,"max":20,"increment":0.5}]},"mediatypepricegranularity":{},"includewinners":true,"includebidderkeys":true},"macros":{"[PLATFORM]":"2","[PROFILE_ID]":"1234","[PROFILE_VERSION]":"1","[UNIX_TIMESTAMP]":"0","[WRAPPER_IMPRESSION_ID]":""}}}`},
				AnalyticsTags: hookanalytics.Analytics{},
			},
			wantBidRequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"maxduration":30,"startdelay":0,"protocols":[1,2,3,4,5,6,7,8,11,12,13,14],"w":640,"h":480,"placement":2,"plcmt":4,"linearity":1,"skip":0,"playbackmethod":[2],"playbackend":1,"delivery":[2,3]},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"data":{"pbadslot":"adunit"},"prebid":{"bidder":{"appnexus":{"placementId":0,"site":"12313","adtag":"45343"}}}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","customdata":"7D75D25F-FAC9-443D-B2D1-B17FEE11E027","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"tid":"123-456-789","ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{"bidadjustmentfactors":{"appnexus":1},"bidderparams":{"pubmatic":{"wiid":""}},"debug":true,"floors":{"enforcement":{"enforcepbs":true},"enabled":true},"targeting":{"pricegranularity":{"precision":2,"ranges":[{"min":0,"max":5,"increment":0.05},{"min":5,"max":10,"increment":0.1},{"min":10,"max":20,"increment":0.5}]},"mediatypepricegranularity":{},"includewinners":true,"includebidderkeys":true},"macros":{"[PLATFORM]":"2","[PROFILE_ID]":"1234","[PROFILE_VERSION]":"1","[UNIX_TIMESTAMP]":"0","[WRAPPER_IMPRESSION_ID]":""}}}}`),
			wantErr:        false,
			doMutate:       true,
		},
		{
			name: "applovin_max_request_successfully_updated_from_DB",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": models.RequestCtx{
							ProfileID:                 1234,
							DisplayID:                 1,
							SSAuction:                 -1,
							Platform:                  "in-app",
							Debug:                     true,
							UA:                        "go-test",
							IP:                        "127.0.0.1",
							IsCTVRequest:              false,
							TrackerEndpoint:           "t.pubmatic.com",
							VideoErrorTrackerEndpoint: "t.pubmatic.com/error",
							UidCookie: &http.Cookie{
								Name:  "uids",
								Value: `eyJ0ZW1wVUlEcyI6eyIzM2Fjcm9zcyI6eyJ1aWQiOiIxMTkxNzkxMDk5Nzc2NjEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTo0My4zODg4Nzk5NVoifSwiYWRmIjp7InVpZCI6IjgwNDQ2MDgzMzM3Nzg4MzkwNzgiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMS4wMzMwNTQ3MjdaIn0sImFka2VybmVsIjp7InVpZCI6IkE5MTYzNTAwNzE0OTkyOTMyOTkwIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuMzczMzg1NjYyWiJ9LCJhZGtlcm5lbEFkbiI6eyJ1aWQiOiJBOTE2MzUwMDcxNDk5MjkzMjk5MCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEzLjQzNDkyNTg5NloifSwiYWRtaXhlciI6eyJ1aWQiOiIzNjZhMTdiMTJmMjI0ZDMwOGYzZTNiOGRhOGMzYzhhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjU5MjkxNDgwMVoifSwiYWRueHMiOnsidWlkIjoiNDE5Mjg5ODUzMDE0NTExOTMiLCJleHBpcmVzIjoiMjAyMy0wMS0xOFQwOTo1MzowOC44MjU0NDI2NzZaIn0sImFqYSI6eyJ1aWQiOiJzMnN1aWQ2RGVmMFl0bjJveGQ1aG9zS1AxVmV3IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTMuMjM5MTc2MDU0WiJ9LCJlcGxhbm5pbmciOnsidWlkIjoiQUoxRjBTOE5qdTdTQ0xWOSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjkyOTk2MDQ3M1oifSwiZ2Ftb3NoaSI6eyJ1aWQiOiJndXNyXzM1NmFmOWIxZDhjNjQyYjQ4MmNiYWQyYjdhMjg4MTYxIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuNTI0MTU3MjI1WiJ9LCJncmlkIjp7InVpZCI6IjRmYzM2MjUwLWQ4NTItNDU5Yy04NzcyLTczNTZkZTE3YWI5NyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE0LjY5NjMxNjIyN1oifSwiZ3JvdXBtIjp7InVpZCI6IjdENzVEMjVGLUZBQzktNDQzRC1CMkQxLUIxN0ZFRTExRTAyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjM5LjIyNjIxMjUzMloifSwiaXgiOnsidWlkIjoiWW9ORlNENlc5QkphOEh6eEdtcXlCUUFBXHUwMDI2Mjk3IiwiZXhwaXJlcyI6IjIwMjMtMDUtMzFUMDc6NTM6MzguNTU1ODI3MzU0WiJ9LCJqaXhpZSI6eyJ1aWQiOiI3MzY3MTI1MC1lODgyLTExZWMtYjUzOC0xM2FjYjdhZjBkZTQiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi4xOTEwOTk3MzJaIn0sImxvZ2ljYWQiOnsidWlkIjoiQVZ4OVROQS11c25pa3M4QURzTHpWa3JvaDg4QUFBR0JUREh0UUEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS40NTUxNDk2MTZaIn0sIm1lZGlhbmV0Ijp7InVpZCI6IjI5Nzg0MjM0OTI4OTU0MTAwMDBWMTAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMy42NzIyMTUxMjhaIn0sIm1naWQiOnsidWlkIjoibTU5Z1hyN0xlX1htIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTcuMDk3MDAxNDcxWiJ9LCJuYW5vaW50ZXJhY3RpdmUiOnsidWlkIjoiNmFlYzhjMTAzNzlkY2I3ODQxMmJjODBiNmRkOWM5NzMxNzNhYjdkNzEyZTQzMWE1YTVlYTcwMzRlNTZhNThhMCIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjE2LjcxNDgwNzUwNVoifSwib25ldGFnIjp7InVpZCI6IjdPelZoVzFOeC1LOGFVak1HMG52NXVNYm5YNEFHUXZQbnVHcHFrZ3k0ckEiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OTowOS4xNDE3NDEyNjJaIn0sIm9wZW54Ijp7InVpZCI6IjVkZWNlNjIyLTBhMjMtMGRhYi0zYTI0LTVhNzcwMTBlNDU4MiIsImV4cGlyZXMiOiIyMDIzLTA1LTMxVDA3OjUyOjQ3LjE0MDQxNzM2M1oifSwicHVibWF0aWMiOnsidWlkIjoiN0Q3NUQyNUYtRkFDOS00NDNELUIyRDEtQjE3RkVFMTFFMDI3IiwiZXhwaXJlcyI6IjIwMjItMTAtMzFUMDk6MTQ6MjUuNzM3MjU2ODk5WiJ9LCJyaWNoYXVkaWVuY2UiOnsidWlkIjoiY2I2YzYzMjAtMzNlMi00Nzc0LWIxNjAtMXp6MTY1NDg0MDc0OSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjUyNTA3NDE4WiJ9LCJzbWFydHlhZHMiOnsidWlkIjoiMTJhZjE1ZTQ0ZjAwZDA3NjMwZTc0YzQ5MTU0Y2JmYmE0Zjg0N2U4ZDRhMTU0YzhjM2Q1MWY1OGNmNzJhNDYyNyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjgyNTAzMTg4NFoifSwic21pbGV3YW50ZWQiOnsidWlkIjoiZGQ5YzNmZTE4N2VmOWIwOWNhYTViNzExNDA0YzI4MzAiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNC4yNTU2MDkzNjNaIn0sInN5bmFjb3JtZWRpYSI6eyJ1aWQiOiJHRFBSIiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MDkuOTc5NTgzNDM4WiJ9LCJ0cmlwbGVsaWZ0Ijp7InVpZCI6IjcwMjE5NzUwNTQ4MDg4NjUxOTQ2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA4Ljk4OTY3MzU3NFoifSwidmFsdWVpbXByZXNzaW9uIjp7InVpZCI6IjlkMDgxNTVmLWQ5ZmUtNGI1OC04OThlLWUyYzU2MjgyYWIzZSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjA5LjA2NzgzOTE2NFoifSwidmlzeCI6eyJ1aWQiOiIyN2UwYWMzYy1iNDZlLTQxYjMtOTkyYy1mOGQyNzE0OTQ5NWUiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxMi45ODk1MjM1NzNaIn0sInlpZWxkbGFiIjp7InVpZCI6IjY5NzE0ZDlmLWZiMDAtNGE1Zi04MTljLTRiZTE5MTM2YTMyNSIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjExLjMwMzAyNjYxNVoifSwieWllbGRtbyI6eyJ1aWQiOiJnOTZjMmY3MTlmMTU1MWIzMWY2MyIsImV4cGlyZXMiOiIyMDIyLTA2LTI0VDA1OjU5OjEwLjExMDUyODYwOVoifSwieWllbGRvbmUiOnsidWlkIjoiMmE0MmZiZDMtMmM3MC00ZWI5LWIxYmQtMDQ2OTY2NTBkOTQ4IiwiZXhwaXJlcyI6IjIwMjItMDYtMjRUMDU6NTk6MTAuMzE4MzMzOTM5WiJ9LCJ6ZXJvY2xpY2tmcmF1ZCI6eyJ1aWQiOiJiOTk5NThmZS0yYTg3LTJkYTQtOWNjNC05NjFmZDExM2JlY2UiLCJleHBpcmVzIjoiMjAyMi0wNi0yNFQwNTo1OToxNS43MTk1OTQ1NjZaIn19LCJiZGF5IjoiMjAyMi0wNS0xN1QwNjo0ODozOC4wMTc5ODgyMDZaIn0=`,
							},
							KADUSERCookie: &http.Cookie{
								Name:  "KADUSERCOOKIE",
								Value: `7D75D25F-FAC9-443D-B2D1-B17FEE11E027`,
							},
							OriginCookie:             "go-test",
							Aliases:                  make(map[string]string),
							ImpBidCtx:                make(map[string]models.ImpCtx),
							PrebidBidderCode:         make(map[string]string),
							BidderResponseTimeMillis: make(map[string]int),
							ProfileIDStr:             "1234",
							Endpoint:                 models.EndpointV25,
							SeatNonBids:              make(map[string][]openrtb_ext.NonBid),
							MetricsEngine:            mockEngine,
							SignalData:               `{"device":{"devicetype":4,"w":393,"h":852,"ifa":"F5BA1637-7156-4369-BA7E-3C45033D9F61","mccmnc":"311-480","js":1,"osv":"17.3.1","connectiontype":5,"os":"iOS","pxratio":3,"geo":{"lastfix":8,"lat":37.48773508935608,"utcoffset":-480,"lon":-122.22855027909678,"type":1},"language":"en","make":"Apple","ext":{"atts":3},"ua":"Mozilla/5.0 (iPhone; CPU iPhone OS 17_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148","model":"iPhone15,2","carrier":"Verizon"},"source":{"ext":{"omidpn":"Pubmatic","omidpv":"3.1.0"}},"id":"CE204A0E-31C3-4D7F-A1A0-D34AF5ED1A7F","app":{"id":"406719683","paid":1,"keywords":"k1=v1","domain":"abc.com","bundle":"406719683","storeurl":"https://apps.apple.com/us/app/gasbuddy-find-pay-for-gas/id406719683","name":"GasBuddy","publisher":{"id":"160361"},"ver":"700.89.22927"},"ext":{"wrapper":{"sumry_disable":1,"clientconfig":1,"profileid":3422}},"imp":[{"secure":1,"tagid":"Mobile_iPhone_List_Screen_Bottom","banner":{"pos":0,"format":[{"w":300,"h":250}],"api":[5,6,7]},"id":"98D9318E-5276-402F-BAA4-CDBD8A364957","ext":{"skadn":{"sourceapp":"406719683","versions":["2.0","2.1","2.2","3.0","4.0"],"skadnetids":["cstr6suwn9.skadnetwork","7ug5zh24hu.skadnetwork","uw77j35x4d.skadnetwork","c6k4g5qg8m.skadnetwork","hs6bdukanm.skadnetwork","yclnxrl5pm.skadnetwork","3sh42y64q3.skadnetwork","cj5566h2ga.skadnetwork","klf5c3l5u5.skadnetwork","8s468mfl3y.skadnetwork","2u9pt9hc89.skadnetwork","7rz58n8ntl.skadnetwork","ppxm28t8ap.skadnetwork","mtkv5xtk9e.skadnetwork","cg4yq2srnc.skadnetwork","wzmmz9fp6w.skadnetwork","k674qkevps.skadnetwork","v72qych5uu.skadnetwork","578prtvx9j.skadnetwork","3rd42ekr43.skadnetwork","g28c52eehv.skadnetwork","2fnua5tdw4.skadnetwork","9nlqeag3gk.skadnetwork","5lm9lj6jb7.skadnetwork","97r2b46745.skadnetwork","e5fvkxwrpn.skadnetwork","4pfyvq9l8r.skadnetwork","tl55sbb4fm.skadnetwork","t38b2kh725.skadnetwork","prcb7njmu6.skadnetwork","mlmmfzh3r3.skadnetwork","9t245vhmpl.skadnetwork","9rd848q2bz.skadnetwork","4fzdc2evr5.skadnetwork","4468km3ulz.skadnetwork","m8dbw4sv7c.skadnetwork","ejvt5qm6ak.skadnetwork","5lm9lj6jb7.skadnetwork","44jx6755aq.skadnetwork","6g9af3uyq4.skadnetwork","u679fj5vs4.skadnetwork","rx5hdcabgc.skadnetwork","275upjj5gd.skadnetwork","p78axxw29g.skadnetwork"],"productpage":1,"version":"2.0"}},"displaymanagerver":"3.1.0","clickbrowser":1,"video":{"companionad":[{"pos":0,"format":[{"w":300,"h":250}],"vcm":1}],"protocols":[2,3,5,6,7,8,11,12,13,14],"h":250,"w":300,"linearity":1,"pos":0,"boxingallowed":1,"placement":2,"mimes":["video/3gpp2","video/quicktime","video/mp4","video/x-m4v","video/3gpp"],"companiontype":[1,2,3],"delivery":[2],"startdelay":0,"playbackend":1,"api":[7]},"displaymanager":"PubMatic_OpenWrap_SDK","instl":0}],"at":1,"cur":["USD"],"regs":{"coppa":1,"ext":{"ccpa":0,"gdpr":1,"gpp":"gpp_string","gpp_sid":[7],"us_privacy":"uspConsentString","consent":"0"}},"user":{"id":"119208432","name":"Pubmatic_sdk","yob":2000,"gender":"M","data":[{"id":"1","name":"PubMatic_SDK","segment":[{"id":"seg_id","name":"segment_sdk","ext":{"segtax":90}}]}],"geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}}}`,
							IsMaxRequest:             true,
						},
					},
				},
				bidrequest: json.RawMessage(`{"id":"{BID_ID}","at":1,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"tmax":3000,"app":{"publisher":{"name":"New Story Inc.","id":"5890","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"name":"DrawHappyAngel","ver":"0.5.4","bundle":"com.newstory.DrawHappyAngel","cat":["IAB9-30"],"id":"{NETWORK_APP_ID}","ext":{"orientation":1}},"device":{"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ip":"38.158.207.171","carrier":"MYTEL","language":"en_US","hwv":"ruby","ppi":440,"pxratio":2.75,"devicetype":4,"connectiontype":2,"js":1,"h":2400,"w":1080,"geo":{"type":2,"ipservice":3,"lat":40.7429,"lon":-73.9392,"long":-73.9392,"city":"Queens","country":"USA","region":"ny","dma":"501","metro":"501","zip":"11101","ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ext":{},"osv":"13.0.0","ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","make":"xiaomi","model":"22101316c","os":"android"},"imp":[{"id":"1","displaymanager":"applovin_mediation","displaymanagerver":"11.8.2","instl":0,"secure":0,"tagid":"adunit","bidfloor":0.01,"bidfloorcur":"USD","exp":14400,"banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"rwdd":0}],"user":{"data":[{"id":"1","name":"Publisher Passed","segment":[{"signal":"{BIDDING_SIGNAL}"}]}],"ext":{"gdpr":0}},"regs":{"ext":{"ccpa":0,"gdpr":1,"consent":"0","tcf_consent_string":"{TCF_STRING}"}},"source":{"ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]}}},"ext":{"wrapper":{"profileid":1234}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"adunit@700x900": {
						SlotName: "adunit@700x900",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@700x900"},
					HashValueMap: map[string]string{
						"adunit@700x900": "1232433543534543",
					},
				})
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(rctx.Platform, "5890", "appnexus")
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
				mockEngine.EXPECT().RecordMaxSDKRequests("5890", "1234")
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:        false,
				NbrCode:       0,
				ChangeSet:     hookstage.ChangeSet[hookstage.BeforeValidationRequestPayload]{},
				DebugMessages: []string{`new imp: {"1":{"ImpID":"1","TagID":"adunit","Div":"","SlotName":"adunit","AdUnitName":"adunit","Secure":0,"BidFloor":0.01,"BidFloorCur":"USD","IsRewardInventory":null,"Banner":true,"Video":{"mimes":["video/3gpp2","video/quicktime","video/mp4","video/x-m4v","video/3gpp"],"startdelay":0,"protocols":[2,3,5,6,7,8,11,12,13,14],"w":300,"h":250,"placement":2,"linearity":1,"boxingallowed":1,"playbackend":1,"delivery":[2],"pos":0,"companionad":[{"format":[{"w":300,"h":250}],"pos":0,"vcm":1}],"api":[7],"companiontype":[1,2,3]},"Native":null,"IncomingSlots":["300x250","300x250v","700x900","728x90"],"Type":"video","Bidders":{"appnexus":{"PartnerID":2,"PrebidBidderCode":"appnexus","MatchedSlot":"adunit@700x900","KGP":"_AU_@_W_x_H_","KGPV":"","IsRegex":false,"Params":{"placementId":0,"site":"12313","adtag":"45343"},"VASTTagFlag":false,"VASTTagFlags":null}},"NonMapped":{},"NewExt":{"skadn":{"sourceapp":"406719683","versions":["2.0","2.1","2.2","3.0","4.0"],"skadnetids":["cstr6suwn9.skadnetwork","7ug5zh24hu.skadnetwork","uw77j35x4d.skadnetwork","c6k4g5qg8m.skadnetwork","hs6bdukanm.skadnetwork","yclnxrl5pm.skadnetwork","3sh42y64q3.skadnetwork","cj5566h2ga.skadnetwork","klf5c3l5u5.skadnetwork","8s468mfl3y.skadnetwork","2u9pt9hc89.skadnetwork","7rz58n8ntl.skadnetwork","ppxm28t8ap.skadnetwork","mtkv5xtk9e.skadnetwork","cg4yq2srnc.skadnetwork","wzmmz9fp6w.skadnetwork","k674qkevps.skadnetwork","v72qych5uu.skadnetwork","578prtvx9j.skadnetwork","3rd42ekr43.skadnetwork","g28c52eehv.skadnetwork","2fnua5tdw4.skadnetwork","9nlqeag3gk.skadnetwork","5lm9lj6jb7.skadnetwork","97r2b46745.skadnetwork","e5fvkxwrpn.skadnetwork","4pfyvq9l8r.skadnetwork","tl55sbb4fm.skadnetwork","t38b2kh725.skadnetwork","prcb7njmu6.skadnetwork","mlmmfzh3r3.skadnetwork","9t245vhmpl.skadnetwork","9rd848q2bz.skadnetwork","4fzdc2evr5.skadnetwork","4468km3ulz.skadnetwork","m8dbw4sv7c.skadnetwork","ejvt5qm6ak.skadnetwork","5lm9lj6jb7.skadnetwork","44jx6755aq.skadnetwork","6g9af3uyq4.skadnetwork","u679fj5vs4.skadnetwork","rx5hdcabgc.skadnetwork","275upjj5gd.skadnetwork","p78axxw29g.skadnetwork"],"productpage":1,"version":"2.0"},"data":{"pbadslot":"adunit"},"prebid":{"bidder":{"appnexus":{"placementId":0,"site":"12313","adtag":"45343"}}}},"BidCtx":{},"BannerAdUnitCtx":{"MatchedSlot":"","IsRegex":false,"MatchedRegex":"","SelectedSlotAdUnitConfig":null,"AppliedSlotAdUnitConfig":null,"UsingDefaultConfig":false,"AllowedConnectionTypes":null},"VideoAdUnitCtx":{"MatchedSlot":"","IsRegex":false,"MatchedRegex":"","SelectedSlotAdUnitConfig":null,"AppliedSlotAdUnitConfig":null,"UsingDefaultConfig":false,"AllowedConnectionTypes":null},"BidderError":"","IsAdPodRequest":false}}`, `new request.ext: {"prebid":{"bidadjustmentfactors":{"appnexus":1},"bidderparams":{"pubmatic":{"wiid":""}},"debug":true,"floors":{"enforcement":{"enforcepbs":true},"enabled":true},"targeting":{"pricegranularity":{"precision":2,"ranges":[{"min":0,"max":5,"increment":0.05},{"min":5,"max":10,"increment":0.1},{"min":10,"max":20,"increment":0.5}]},"mediatypepricegranularity":{},"includewinners":true,"includebidderkeys":true},"macros":{"[PLATFORM]":"5","[PROFILE_ID]":"1234","[PROFILE_VERSION]":"1","[UNIX_TIMESTAMP]":"0","[WRAPPER_IMPRESSION_ID]":""}}}`},
				AnalyticsTags: hookanalytics.Analytics{},
			},
			wantBidRequest: json.RawMessage(`{"id":"{BID_ID}","imp":[{"id":"1","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900,"api":[5,6,7]},"video":{"mimes":["video/3gpp2","video/quicktime","video/mp4","video/x-m4v","video/3gpp"],"startdelay":0,"protocols":[2,3,5,6,7,8,11,12,13,14],"w":300,"h":250,"placement":2,"linearity":1,"boxingallowed":1,"playbackend":1,"delivery":[2],"pos":0,"companionad":[{"format":[{"w":300,"h":250}],"pos":0,"vcm":1}],"api":[7],"companiontype":[1,2,3]},"displaymanager":"PubMatic_OpenWrap_SDK","displaymanagerver":"3.1.0","tagid":"adunit","bidfloor":0.01,"bidfloorcur":"USD","clickbrowser":1,"secure":0,"exp":14400,"ext":{"skadn":{"sourceapp":"406719683","versions":["2.0","2.1","2.2","3.0","4.0"],"skadnetids":["cstr6suwn9.skadnetwork","7ug5zh24hu.skadnetwork","uw77j35x4d.skadnetwork","c6k4g5qg8m.skadnetwork","hs6bdukanm.skadnetwork","yclnxrl5pm.skadnetwork","3sh42y64q3.skadnetwork","cj5566h2ga.skadnetwork","klf5c3l5u5.skadnetwork","8s468mfl3y.skadnetwork","2u9pt9hc89.skadnetwork","7rz58n8ntl.skadnetwork","ppxm28t8ap.skadnetwork","mtkv5xtk9e.skadnetwork","cg4yq2srnc.skadnetwork","wzmmz9fp6w.skadnetwork","k674qkevps.skadnetwork","v72qych5uu.skadnetwork","578prtvx9j.skadnetwork","3rd42ekr43.skadnetwork","g28c52eehv.skadnetwork","2fnua5tdw4.skadnetwork","9nlqeag3gk.skadnetwork","5lm9lj6jb7.skadnetwork","97r2b46745.skadnetwork","e5fvkxwrpn.skadnetwork","4pfyvq9l8r.skadnetwork","tl55sbb4fm.skadnetwork","t38b2kh725.skadnetwork","prcb7njmu6.skadnetwork","mlmmfzh3r3.skadnetwork","9t245vhmpl.skadnetwork","9rd848q2bz.skadnetwork","4fzdc2evr5.skadnetwork","4468km3ulz.skadnetwork","m8dbw4sv7c.skadnetwork","ejvt5qm6ak.skadnetwork","5lm9lj6jb7.skadnetwork","44jx6755aq.skadnetwork","6g9af3uyq4.skadnetwork","u679fj5vs4.skadnetwork","rx5hdcabgc.skadnetwork","275upjj5gd.skadnetwork","p78axxw29g.skadnetwork"],"productpage":1,"version":"2.0"},"data":{"pbadslot":"adunit"},"prebid":{"bidder":{"appnexus":{"placementId":0,"site":"12313","adtag":"45343"}}}}}],"app":{"id":"{NETWORK_APP_ID}","name":"DrawHappyAngel","bundle":"com.newstory.DrawHappyAngel","domain":"abc.com","cat":["IAB9-30"],"ver":"0.5.4","paid":1,"publisher":{"id":"5890","name":"New Story Inc.","ext":{"installed_sdk":{"id":"MOLOCO_BIDDING","sdk_version":{"major":1,"minor":0,"micro":0},"adapter_version":{"major":1,"minor":0,"micro":0}}}},"keywords":"k1=v1","ext":{"orientation":1}},"device":{"geo":{"city":"Queens","lat":40.7429,"lon":-73.9392,"type":2,"ipservice":3,"country":"USA","region":"ny","metro":"501","zip":"11101","utcoffset":-480,"ext":{"org":"Myanmar Broadband Telecom Co.","isp":"Myanmar Broadband Telecom Co."}},"ua":"Mozilla/5.0 (Linux; Android 13; 22101316C Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.230 Mobile Safari/537.36","ip":"38.158.207.171","devicetype":4,"make":"xiaomi","model":"22101316c","os":"android","osv":"13.0.0","hwv":"ruby","h":2400,"w":1080,"ppi":440,"pxratio":2.75,"js":1,"language":"en","carrier":"MYTEL","mccmnc":"311-480","connectiontype":5,"ifa":"497a10d6-c4dd-4e04-a986-c32b7180d462","ext":{"atts":3}},"user":{"customdata":"7D75D25F-FAC9-443D-B2D1-B17FEE11E027","yob":2000,"gender":"M","data":[{"id":"1","name":"PubMatic_SDK","segment":[{"id":"seg_id","name":"segment_sdk","ext":{"segtax":90}}]}],"ext":{"gdpr":0}},"at":1,"tmax":3000,"bcat":["IAB26-4","IAB26-2","IAB25-6","IAB25-5","IAB25-4","IAB25-3","IAB25-1","IAB25-7","IAB8-18","IAB26-3","IAB26-1","IAB8-5","IAB25-2","IAB11-4"],"source":{"tid":"{BID_ID}","ext":{"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"applovin.com","sid":"53bf468f18c5a0e2b7d4e3f748c677c1","rid":"494dbe15a3ce08c54f4e456363f35a022247f997","hp":1}]},"omidpn":"Pubmatic","omidpv":"3.1.0"}},"regs":{"coppa":1,"ext":{"ccpa":0,"gdpr":1,"consent":"0","tcf_consent_string":"{TCF_STRING}","gpp":"gpp_string","gpp_sid":[7],"us_privacy":"uspConsentString"}},"ext":{"prebid":{"bidadjustmentfactors":{"appnexus":1},"bidderparams":{"pubmatic":{"wiid":""}},"debug":true,"floors":{"enforcement":{"enforcepbs":true},"enabled":true},"targeting":{"pricegranularity":{"precision":2,"ranges":[{"min":0,"max":5,"increment":0.05},{"min":5,"max":10,"increment":0.1},{"min":10,"max":20,"increment":0.5}]},"mediatypepricegranularity":{},"includewinners":true,"includebidderkeys":true},"macros":{"[PLATFORM]":"5","[PROFILE_ID]":"1234","[PROFILE_VERSION]":"1","[UNIX_TIMESTAMP]":"0","[WRAPPER_IMPRESSION_ID]":""}}}}`),
			wantErr:        false,
			doMutate:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			m := OpenWrap{
				cfg:           tt.fields.cfg,
				cache:         tt.fields.cache,
				metricEngine:  tt.fields.metricEngine,
				featureConfig: mockFeature,
			}

			bidrequest := &openrtb2.BidRequest{}
			json.Unmarshal(tt.args.bidrequest, bidrequest)
			tt.args.payload.BidRequest = bidrequest
			got, err := m.handleBeforeValidationHook(tt.args.ctx, tt.args.moduleCtx, tt.args.payload)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want.Reject, got.Reject)
			assert.Equal(t, tt.want.NbrCode, got.NbrCode)
			for i := 0; i < len(got.DebugMessages); i++ {
				gotDebugMessage, _ := json.Marshal(got.DebugMessages[i])
				wantDebugMessage, _ := json.Marshal(tt.want.DebugMessages[i])
				sort.Slice(gotDebugMessage, func(i, j int) bool {
					return gotDebugMessage[i] < gotDebugMessage[j]
				})
				sort.Slice(wantDebugMessage, func(i, j int) bool {
					return wantDebugMessage[i] < wantDebugMessage[j]
				})
				assert.Equal(t, wantDebugMessage, gotDebugMessage)
			}

			if tt.doMutate {
				mutations := got.ChangeSet.Mutations()
				assert.NotEmpty(t, mutations, tt.name)
				for _, mut := range mutations {
					result, err := mut.Apply(tt.args.payload)
					assert.Nil(t, err, tt.name)
					gotBidRequest, _ := json.Marshal(result.BidRequest)
					assert.JSONEq(t, string(tt.wantBidRequest), string(gotBidRequest))
				}
			}
		})
	}
}

func TestUserAgent_handleBeforeValidationHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	mockFeature := mock_feature.NewMockFeature(ctrl)

	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		ctx        context.Context
		moduleCtx  hookstage.ModuleInvocationContext
		payload    hookstage.BeforeValidationRequestPayload
		bidrequest json.RawMessage
	}
	type want struct {
		rctx  *models.RequestCtx
		error bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
		setup  func()
	}{
		{
			name: "bidRequest.Device.UA_is_present",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":1}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidRequestExt))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InvalidRequestExt))
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)

			},
			want: want{
				rctx: &models.RequestCtx{
					UA: "Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36",
				},
				error: true,
			},
		},
		{
			name: "bidRequest.Device.UA_is_absent",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":1}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidRequestExt))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InvalidRequestExt))
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: want{
				rctx: &models.RequestCtx{
					UA:    "go-test",
					PubID: 1,
				},
				error: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			adapters.InitBidders("./static/bidder-params/")
			m := OpenWrap{
				cfg:           tt.fields.cfg,
				cache:         tt.fields.cache,
				metricEngine:  tt.fields.metricEngine,
				featureConfig: mockFeature,
			}
			tt.args.payload.BidRequest = &openrtb2.BidRequest{}
			json.Unmarshal(tt.args.bidrequest, tt.args.payload.BidRequest)

			_, err := m.handleBeforeValidationHook(tt.args.ctx, tt.args.moduleCtx, tt.args.payload)
			assert.Equal(t, tt.want.error, err != nil, "mismatched error received from handleBeforeValidationHook")
			iRctx := tt.args.moduleCtx.ModuleContext["rctx"]
			assert.Equal(t, tt.want.rctx == nil, iRctx == nil, "mismatched rctx received from handleBeforeValidationHook")
			gotRctx := iRctx.(models.RequestCtx)
			assert.Equal(t, tt.want.rctx.UA, gotRctx.UA, "mismatched rctx.UA received from handleBeforeValidationHook")
		})
	}
}

func TestVASTUnwrap_handleBeforeValidationHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	mockFeature := mock_feature.NewMockFeature(ctrl)

	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		ctx          context.Context
		moduleCtx    hookstage.ModuleInvocationContext
		payload      hookstage.BeforeValidationRequestPayload
		bidrequest   json.RawMessage
		randomNumber int
	}
	type want struct {
		rctx  *models.RequestCtx
		error bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
		setup  func()
	}{
		{
			name: "VAST Unwrap Disabled in DB, traffic percent  present in config",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
				cfg: config.Config{
					Features: config.FeatureToggle{
						VASTUnwrapPercent: 10,
					},
				},
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"adunit@700x900": {
						SlotName: "adunit@700x900",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@700x900"},
					HashValueMap: map[string]string{
						"adunit@700x900": "1232433543534543",
					},
				})
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID:       "1",
						models.PLATFORM_KEY:           models.PLATFORM_APP,
						models.VastUnwrapperEnableKey: "0",
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})

				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(rctx.Platform, "5890", "appnexus")
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: want{
				rctx: &models.RequestCtx{
					VastUnwrapEnabled: false,
				},
				error: false,
			},
		},
		{
			name: "VAST Unwrap Enabled in DB, traffic percent not present in config and DB",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest:   json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
				randomNumber: 20,
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"adunit@700x900": {
						SlotName: "adunit@700x900",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@700x900"},
					HashValueMap: map[string]string{
						"adunit@700x900": "1232433543534543",
					},
				})
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID:       "1",
						models.PLATFORM_KEY:           models.PLATFORM_APP,
						models.VastUnwrapperEnableKey: "1",
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})

				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(rctx.Platform, "5890", "appnexus")
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: want{
				rctx: &models.RequestCtx{
					VastUnwrapEnabled: false,
				},
				error: false,
			},
		},
		{
			name: "VAST Unwrap Enabled in DB, traffic percent present in config",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest:   json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
				randomNumber: 20,
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
				cfg: config.Config{
					Features: config.FeatureToggle{
						VASTUnwrapPercent: 100,
					},
				},
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"adunit@700x900": {
						SlotName: "adunit@700x900",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@700x900"},
					HashValueMap: map[string]string{
						"adunit@700x900": "1232433543534543",
					},
				})
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID:       "1",
						models.PLATFORM_KEY:           models.PLATFORM_APP,
						models.VastUnwrapperEnableKey: "1",
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})

				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(rctx.Platform, "5890", "appnexus")
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: want{
				rctx: &models.RequestCtx{
					VastUnwrapEnabled: true,
				},
				error: false,
			},
		},
		{
			name: "VAST Unwrap Enabled in DB, traffic percent present in config and DB",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest:   json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
				randomNumber: 20,
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
				cfg: config.Config{
					Features: config.FeatureToggle{
						VASTUnwrapPercent: 10,
					},
				},
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"adunit@700x900": {
						SlotName: "adunit@700x900",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@700x900"},
					HashValueMap: map[string]string{
						"adunit@700x900": "1232433543534543",
					},
				})
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID:            "1",
						models.PLATFORM_KEY:                models.PLATFORM_APP,
						models.VastUnwrapperEnableKey:      "1",
						models.VastUnwrapTrafficPercentKey: "50",
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})

				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(rctx.Platform, "5890", "appnexus")
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: want{
				rctx: &models.RequestCtx{
					VastUnwrapEnabled: true,
				},
				error: false,
			},
		},
		{
			name: "VAST Unwrap Enabled DB, traffic percent not present in config",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest:   json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","bidfloor":4.3,"bidfloorcur":"USD","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
				randomNumber: 20,
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
					"adunit@700x900": {
						SlotName: "adunit@700x900",
						SlotMappings: map[string]interface{}{
							models.SITE_CACHE_KEY: "12313",
							models.TAG_CACHE_KEY:  "45343",
						},
					},
				})
				mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
					OrderedSlotList: []string{"adunit@700x900"},
					HashValueMap: map[string]string{
						"adunit@700x900": "1232433543534543",
					},
				})
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID:            "1",
						models.PLATFORM_KEY:                models.PLATFORM_APP,
						models.VastUnwrapperEnableKey:      "1",
						models.VastUnwrapTrafficPercentKey: "100",
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})

				mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(rctx.Platform, "5890", "appnexus")
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: want{
				rctx: &models.RequestCtx{
					VastUnwrapEnabled: true,
				},
				error: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			GetRandomNumberIn1To100 = func() int {
				return tt.args.randomNumber
			}

			adapters.InitBidders("./static/bidder-params/")
			m := OpenWrap{
				cfg:           tt.fields.cfg,
				cache:         tt.fields.cache,
				metricEngine:  tt.fields.metricEngine,
				featureConfig: mockFeature,
			}
			tt.args.payload.BidRequest = &openrtb2.BidRequest{}
			json.Unmarshal(tt.args.bidrequest, tt.args.payload.BidRequest)

			_, err := m.handleBeforeValidationHook(tt.args.ctx, tt.args.moduleCtx, tt.args.payload)
			assert.Equal(t, tt.want.error, err != nil, "mismatched error received from handleBeforeValidationHook")
			iRctx := tt.args.moduleCtx.ModuleContext["rctx"]
			assert.Equal(t, tt.want.rctx == nil, iRctx == nil, "mismatched rctx received from handleBeforeValidationHook")
			gotRctx := iRctx.(models.RequestCtx)
			assert.Equal(t, tt.want.rctx.VastUnwrapEnabled, gotRctx.VastUnwrapEnabled, "mismatched rctx.VastUnwrapEnabled received from handleBeforeValidationHook")
		})
	}
}
func TestImpBidCtx_handleBeforeValidationHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	mockFeature := mock_feature.NewMockFeature(ctrl)
	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		ctx        context.Context
		moduleCtx  hookstage.ModuleInvocationContext
		payload    hookstage.BeforeValidationRequestPayload
		bidrequest json.RawMessage
	}
	type want struct {
		rctx  *models.RequestCtx
		error bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
		setup  func()
	}{
		{
			name: "default_impctx_if_getProfileData_fails",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.PLATFORM_KEY:     models.PLATFORM_APP,
						models.DisplayVersionID: "1",
					},
				}, errors.New("test"))
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidProfileConfiguration))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InvalidProfileConfiguration))
				mockEngine.EXPECT().RecordPublisherInvalidProfileRequests(rctx.Endpoint, "5890", rctx.ProfileIDStr)
				mockEngine.EXPECT().RecordPublisherInvalidProfileImpressions("5890", rctx.ProfileIDStr, gomock.Any())
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: want{
				rctx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"123": {
							IncomingSlots: []string{
								"640x480v",
							},
							SlotName:   "adunit",
							AdUnitName: "adunit",
						},
					},
				},
				error: true,
			},
		},
		{
			name: "default_impctx_if_platform_is_missing",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
					},
				}, nil)
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidPlatform))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InvalidPlatform))
				mockEngine.EXPECT().RecordPublisherInvalidProfileRequests(rctx.Endpoint, "5890", rctx.ProfileIDStr)
				mockEngine.EXPECT().RecordPublisherInvalidProfileImpressions("5890", rctx.ProfileIDStr, gomock.Any())
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)

			},
			want: want{
				rctx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"123": {
							IncomingSlots: []string{
								"640x480v",
							},
							SlotName:   "adunit",
							AdUnitName: "adunit",
						},
					},
				},
				error: true,
			},
		},
		{
			name: "default_impctx_if_all_partners_throttled",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
						models.THROTTLE:            "0",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.AllPartnerThrottled))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.AllPartnerThrottled))
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: want{
				error: false,
				rctx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"123": {
							IncomingSlots: []string{
								"640x480v",
							},
							SlotName:   "adunit",
							AdUnitName: "adunit",
						},
					},
				},
			},
		},
		{
			name: "empty_impctx_if_TagID_not_present_in_imp",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}},{"id":"456","video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432"},"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidImpressionTagID))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InvalidImpressionTagID))
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: want{
				rctx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{},
				},
				error: true,
			},
		},
		{
			name: "empty_impctx_if_imp_ext_parse_fails",
			args: args{
				ctx: context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{
					ModuleContext: hookstage.ModuleContext{
						"rctx": rctx,
					},
				},
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"ext":{"wrapper":"invalid","bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}},{"id":"456","video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"ext":{"wrapper":{"div":"div"},"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432"},"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			setup: func() {
				mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(map[int]map[string]string{
					2: {
						models.PARTNER_ID:          "2",
						models.PREBID_PARTNER_NAME: "appnexus",
						models.BidderCode:          "appnexus",
						models.SERVER_SIDE_FLAG:    "1",
						models.KEY_GEN_PATTERN:     "_AU_@_W_x_H_",
						models.TIMEOUT:             "200",
					},
					-1: {
						models.DisplayVersionID: "1",
						models.PLATFORM_KEY:     models.PLATFORM_APP,
					},
				}, nil)
				mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{})
				//prometheus metrics
				mockEngine.EXPECT().RecordPublisherProfileRequests("5890", "1234")
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InternalError))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", int(nbr.InternalError))
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false)
			},
			want: want{
				rctx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{},
				},
				error: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			adapters.InitBidders("./static/bidder-params/")
			m := OpenWrap{
				cfg:           tt.fields.cfg,
				cache:         tt.fields.cache,
				metricEngine:  tt.fields.metricEngine,
				featureConfig: mockFeature,
			}
			tt.args.payload.BidRequest = &openrtb2.BidRequest{}
			json.Unmarshal(tt.args.bidrequest, tt.args.payload.BidRequest)

			_, err := m.handleBeforeValidationHook(tt.args.ctx, tt.args.moduleCtx, tt.args.payload)
			assert.Equal(t, tt.want.error, err != nil, "mismatched error")
			iRctx := tt.args.moduleCtx.ModuleContext["rctx"]
			gotRctx := iRctx.(models.RequestCtx)
			assert.Equal(t, tt.want.rctx.ImpBidCtx, gotRctx.ImpBidCtx, "mismatched rctx.ImpBidCtx")
		})
	}
}

func TestGetSlotName(t *testing.T) {
	type args struct {
		tagId  string
		impExt *models.ImpExtension
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Slot_name_from_gpid",
			args: args{
				tagId: "some-tagid",
				impExt: &models.ImpExtension{
					GpId: "some-gpid",
				},
			},
			want: "some-gpid",
		},
		{
			name: "Slot_name_from_tagid",
			args: args{
				tagId: "some-tagid",
				impExt: &models.ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
					},
				},
			},
			want: "some-tagid",
		},
		{
			name: "Slot_name_from_pbadslot",
			args: args{
				tagId: "",
				impExt: &models.ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
					},
				},
			},
			want: "some-pbadslot",
		},
		{
			name: "Slot_name_from_stored_request_id",
			args: args{
				tagId: "",
				impExt: &models.ImpExtension{
					Prebid: openrtb_ext.ExtImpPrebid{
						StoredRequest: &openrtb_ext.ExtStoredRequest{
							ID: "stored-req-id",
						},
					},
				},
			},
			want: "stored-req-id",
		},
		{
			name: "imp_ext_nil_slot_name_from_tag_id",
			args: args{
				tagId:  "some-tagid",
				impExt: nil,
			},
			want: "some-tagid",
		},
		{
			name: "empty_slot_name",
			args: args{
				tagId:  "",
				impExt: &models.ImpExtension{},
			},
			want: "",
		},
		{
			name: "all_level_information_is_present_slot_name_picked_by_preference",
			args: args{
				tagId: "some-tagid",
				impExt: &models.ImpExtension{
					GpId: "some-gpid",
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
					},
					Prebid: openrtb_ext.ExtImpPrebid{
						StoredRequest: &openrtb_ext.ExtStoredRequest{
							ID: "stored-req-id",
						},
					},
				},
			},
			want: "some-gpid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSlotName(tt.args.tagId, tt.args.impExt)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestGetAdunitName(t *testing.T) {
	type args struct {
		tagId  string
		impExt *models.ImpExtension
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "adunit_from_adserver_slot",
			args: args{
				tagId: "some-tagid",
				impExt: &models.ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
						AdServer: &openrtb_ext.ExtImpDataAdServer{
							Name:   models.GamAdServer,
							AdSlot: "gam-unit",
						},
					},
				},
			},
			want: "gam-unit",
		},
		{
			name: "adunit_from_pbadslot",
			args: args{
				tagId: "some-tagid",
				impExt: &models.ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
						AdServer: &openrtb_ext.ExtImpDataAdServer{
							Name:   models.GamAdServer,
							AdSlot: "",
						},
					},
				},
			},
			want: "some-pbadslot",
		},
		{
			name: "adunit_from_pbadslot_when_gam_is_absent",
			args: args{
				tagId: "some-tagid",
				impExt: &models.ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
						AdServer: &openrtb_ext.ExtImpDataAdServer{
							Name:   "freewheel",
							AdSlot: "freewheel-unit",
						},
					},
				},
			},
			want: "some-pbadslot",
		},
		{
			name: "adunit_from_TagId",
			args: args{
				tagId: "some-tagid",
				impExt: &models.ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "",
						AdServer: &openrtb_ext.ExtImpDataAdServer{
							Name:   models.GamAdServer,
							AdSlot: "",
						},
					},
				},
			},
			want: "some-tagid",
		},
		{
			name: "adunit_from_TagId_imp_ext_nil",
			args: args{
				tagId:  "some-tagid",
				impExt: nil,
			},
			want: "some-tagid",
		},
		{
			name: "adunit_from_TagId_imp_ext_nil",
			args: args{
				tagId:  "some-tagid",
				impExt: &models.ImpExtension{},
			},
			want: "some-tagid",
		},
		{
			name: "all_level_information_is_present_adunit_name_picked_by_preference",
			args: args{
				tagId: "some-tagid",
				impExt: &models.ImpExtension{
					GpId: "some-gpid",
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "some-pbadslot",
						AdServer: &openrtb_ext.ExtImpDataAdServer{
							Name:   models.GamAdServer,
							AdSlot: "gam-unit",
						},
					},
				},
			},
			want: "gam-unit",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAdunitName(tt.args.tagId, tt.args.impExt)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestGetTagID(t *testing.T) {
	type args struct {
		imp    openrtb2.Imp
		impExt *models.ImpExtension
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "tagId_not_found",
			args: args{
				imp:    openrtb2.Imp{},
				impExt: &models.ImpExtension{},
			},
			want: "",
		},
		{
			name: "tagId_present_in_gpid",
			args: args{
				imp: openrtb2.Imp{},
				impExt: &models.ImpExtension{
					GpId: "/7578294/adunit1",
				},
			},
			want: "/7578294/adunit1",
		},
		{
			name: "tagId_set_by_publisher_on_page",
			args: args{
				imp: openrtb2.Imp{
					TagID: "/7578294/adunit1",
				},
				impExt: &models.ImpExtension{},
			},
			want: "/7578294/adunit1",
		},
		{
			name: "tagId_present_in_pbadslot",
			args: args{
				imp: openrtb2.Imp{},
				impExt: &models.ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "/7578294/adunit1",
					},
				},
			},
			want: "/7578294/adunit1",
		},
		{
			name: "tagId_present_in_pbadslot_and_gpid",
			args: args{
				imp: openrtb2.Imp{},
				impExt: &models.ImpExtension{
					GpId: "/7578294/adunit123",
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "/7578294/adunit",
					},
				},
			},
			want: "/7578294/adunit123",
		},
		{
			name: "tagId_present_in_imp.TagId_and_gpid",
			args: args{
				imp: openrtb2.Imp{
					TagID: "/7578294/adunit",
				},
				impExt: &models.ImpExtension{
					GpId: "/7578294/adunit123",
				},
			},
			want: "/7578294/adunit123",
		},
		{
			name: "tagId_present_in_imp.TagId_and_pbadslot",
			args: args{
				imp: openrtb2.Imp{
					TagID: "/7578294/adunit123",
				},
				impExt: &models.ImpExtension{
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "/7578294/adunit",
					},
				},
			},
			want: "/7578294/adunit123",
		},
		{
			name: "tagId_present_in_imp.TagId_and_pbadslot_and_gpid",
			args: args{
				imp: openrtb2.Imp{
					TagID: "/7578294/adunit",
				},
				impExt: &models.ImpExtension{
					GpId: "/7578294/adunit123",
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "/7578294/adunit12345",
					},
				},
			},
			want: "/7578294/adunit123",
		},
		{
			name: "GpId_contains_'#'",
			args: args{
				imp: openrtb2.Imp{
					TagID: "/7578294/adunit",
				},
				impExt: &models.ImpExtension{
					GpId: "/43743431/DMDemo#Div1",
					Data: openrtb_ext.ExtImpData{
						PbAdslot: "/7578294/adunit12345",
					},
				},
			},
			want: "/43743431/DMDemo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTagID(tt.args.imp, tt.args.impExt); got != tt.want {
				t.Errorf("getTagID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateImpVideoWithVideoConfig(t *testing.T) {
	type args struct {
		imp                    *openrtb2.Imp
		configObjInVideoConfig *modelsAdunitConfig.VideoConfig
	}
	tests := []struct {
		name         string
		args         args
		wantImpVideo *openrtb2.Video
	}{
		{
			name: "imp video object is empty updated from adunit config",
			args: args{
				imp: &openrtb2.Imp{
					ID:    "123",
					Video: &openrtb2.Video{},
				},
				configObjInVideoConfig: &modelsAdunitConfig.VideoConfig{
					Video: openrtb2.Video{
						W:              ptrutil.ToPtr[int64](300),
						H:              ptrutil.ToPtr[int64](250),
						MIMEs:          []string{"MP4"},
						Linearity:      adcom1.LinearityNonLinear,
						StartDelay:     adcom1.StartMidRoll.Ptr(),
						MinDuration:    20,
						MaxDuration:    50,
						Placement:      adcom1.VideoPlacementInStream,
						Plcmt:          adcom1.VideoPlcmtAccompanyingContent,
						Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper, adcom1.CreativeVAST30Wrapper, adcom1.CreativeVAST40, adcom1.CreativeVAST40Wrapper, adcom1.CreativeVAST41, adcom1.CreativeVAST41Wrapper, adcom1.CreativeVAST42, adcom1.CreativeVAST42Wrapper},
						Skip:           ptrutil.ToPtr(int8(1)),
						SkipMin:        10,
						SkipAfter:      5,
						BoxingAllowed:  ptrutil.ToPtr[int8](2),
						PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackPageLoadSoundOn},
						PlaybackEnd:    adcom1.PlaybackCompletion,
						Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryProgressive},
						Protocol:       adcom1.CreativeVAST10,
						Sequence:       1,
						CompanionType:  []adcom1.CompanionType{adcom1.CompanionHTML},
						Pos:            adcom1.PositionAboveFold.Ptr(),
						API:            []adcom1.APIFramework{adcom1.APIVPAID10},
						CompanionAd:    []openrtb2.Banner{},
						BAttr:          []adcom1.CreativeAttribute{adcom1.AttrAudioAuto},
						MaxExtended:    100,
					},
				},
			},
			wantImpVideo: &openrtb2.Video{
				W:              ptrutil.ToPtr[int64](300),
				H:              ptrutil.ToPtr[int64](250),
				MIMEs:          []string{"MP4"},
				Linearity:      adcom1.LinearityNonLinear,
				StartDelay:     adcom1.StartMidRoll.Ptr(),
				MinDuration:    20,
				MaxDuration:    50,
				Placement:      adcom1.VideoPlacementInStream,
				Plcmt:          adcom1.VideoPlcmtAccompanyingContent,
				Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper, adcom1.CreativeVAST30Wrapper, adcom1.CreativeVAST40, adcom1.CreativeVAST40Wrapper, adcom1.CreativeVAST41, adcom1.CreativeVAST41Wrapper, adcom1.CreativeVAST42, adcom1.CreativeVAST42Wrapper},
				Skip:           ptrutil.ToPtr[int8](1),
				SkipMin:        10,
				SkipAfter:      5,
				BoxingAllowed:  ptrutil.ToPtr[int8](2),
				PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackPageLoadSoundOn},
				PlaybackEnd:    adcom1.PlaybackCompletion,
				Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryProgressive},
				Protocol:       adcom1.CreativeVAST10,
				Sequence:       1,
				CompanionType:  []adcom1.CompanionType{adcom1.CompanionHTML},
				Pos:            ptrutil.ToPtr(adcom1.PositionAboveFold),
				API:            []adcom1.APIFramework{adcom1.APIVPAID10},
				CompanionAd:    []openrtb2.Banner{},
				BAttr:          []adcom1.CreativeAttribute{adcom1.AttrAudioAuto},
				MaxExtended:    100,
			},
		},
		{
			name: "imp video object is not empty and adunit config is also not empty priority to request level parameters",
			args: args{
				imp: &openrtb2.Imp{
					ID: "123",
					Video: &openrtb2.Video{
						W: ptrutil.ToPtr[int64](300),
						H: ptrutil.ToPtr[int64](250),
					},
				},

				configObjInVideoConfig: &modelsAdunitConfig.VideoConfig{
					Video: openrtb2.Video{
						W:              ptrutil.ToPtr[int64](400),
						H:              ptrutil.ToPtr[int64](300),
						MIMEs:          []string{"MP4"},
						Linearity:      adcom1.LinearityNonLinear,
						StartDelay:     adcom1.StartMidRoll.Ptr(),
						MinDuration:    20,
						MaxDuration:    50,
						Placement:      adcom1.VideoPlacementInStream,
						Plcmt:          adcom1.VideoPlcmtAccompanyingContent,
						Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper, adcom1.CreativeVAST30Wrapper, adcom1.CreativeVAST40, adcom1.CreativeVAST40Wrapper, adcom1.CreativeVAST41, adcom1.CreativeVAST41Wrapper, adcom1.CreativeVAST42, adcom1.CreativeVAST42Wrapper},
						Skip:           ptrutil.ToPtr(int8(1)),
						SkipMin:        10,
						SkipAfter:      5,
						BoxingAllowed:  ptrutil.ToPtr[int8](2),
						PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackPageLoadSoundOn},
						PlaybackEnd:    adcom1.PlaybackCompletion,
						Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryProgressive},
						Protocol:       adcom1.CreativeVAST10,
						Sequence:       1,
						CompanionType:  []adcom1.CompanionType{adcom1.CompanionHTML},
						Pos:            adcom1.PositionAboveFold.Ptr(),
						API:            []adcom1.APIFramework{adcom1.APIVPAID10},
						CompanionAd:    []openrtb2.Banner{},
						BAttr:          []adcom1.CreativeAttribute{adcom1.AttrAudioAuto},
						MaxExtended:    100,
					},
				},
			},
			wantImpVideo: &openrtb2.Video{
				W:              ptrutil.ToPtr[int64](300),
				H:              ptrutil.ToPtr[int64](250),
				MIMEs:          []string{"MP4"},
				Linearity:      adcom1.LinearityNonLinear,
				StartDelay:     adcom1.StartMidRoll.Ptr(),
				MinDuration:    20,
				MaxDuration:    50,
				Placement:      adcom1.VideoPlacementInStream,
				Plcmt:          adcom1.VideoPlcmtAccompanyingContent,
				Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper, adcom1.CreativeVAST30Wrapper, adcom1.CreativeVAST40, adcom1.CreativeVAST40Wrapper, adcom1.CreativeVAST41, adcom1.CreativeVAST41Wrapper, adcom1.CreativeVAST42, adcom1.CreativeVAST42Wrapper},
				Skip:           ptrutil.ToPtr[int8](1),
				SkipMin:        10,
				SkipAfter:      5,
				BoxingAllowed:  ptrutil.ToPtr[int8](2),
				PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackPageLoadSoundOn},
				PlaybackEnd:    adcom1.PlaybackCompletion,
				Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryProgressive},
				Protocol:       adcom1.CreativeVAST10,
				Sequence:       1,
				CompanionType:  []adcom1.CompanionType{adcom1.CompanionHTML},
				Pos:            ptrutil.ToPtr(adcom1.PositionAboveFold),
				API:            []adcom1.APIFramework{adcom1.APIVPAID10},
				CompanionAd:    []openrtb2.Banner{},
				BAttr:          []adcom1.CreativeAttribute{adcom1.AttrAudioAuto},
				MaxExtended:    100,
			},
		},
	}
	for _, tt := range tests {
		updateImpVideoWithVideoConfig(tt.args.imp, tt.args.configObjInVideoConfig)
		assert.Equal(t, tt.wantImpVideo, tt.args.imp.Video, tt.name)
	}
}

func TestUpdateAmpImpVideoWithDefault(t *testing.T) {
	type args struct {
		imp *openrtb2.Imp
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.Video
	}{
		{
			name: "banner has the width and height",
			args: args{
				imp: &openrtb2.Imp{
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](300),
						H: ptrutil.ToPtr[int64](250),
					},
					Video: &openrtb2.Video{},
				},
			},
			want: &openrtb2.Video{
				MIMEs:          []string{"video/mp4"},
				MinDuration:    0,
				MaxDuration:    30,
				StartDelay:     adcom1.StartPreRoll.Ptr(),
				Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper, adcom1.CreativeVAST30Wrapper, adcom1.CreativeVAST40, adcom1.CreativeVAST40Wrapper, adcom1.CreativeVAST41, adcom1.CreativeVAST41Wrapper, adcom1.CreativeVAST42, adcom1.CreativeVAST42Wrapper},
				Placement:      adcom1.VideoPlacementInBanner,
				Plcmt:          adcom1.VideoPlcmtNoContent,
				Linearity:      adcom1.LinearityLinear,
				Skip:           ptrutil.ToPtr[int8](0),
				SkipMin:        0,
				SkipAfter:      0,
				BoxingAllowed:  ptrutil.ToPtr[int8](1),
				PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackPageLoadSoundOff},
				PlaybackEnd:    adcom1.PlaybackCompletion,
				Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryProgressive, adcom1.DeliveryDownload},
				W:              ptrutil.ToPtr[int64](300),
				H:              ptrutil.ToPtr[int64](250),
			},
		},
		{
			name: "banner has the width and height in the banner format object",
			args: args{
				imp: &openrtb2.Imp{
					Banner: &openrtb2.Banner{
						Format: []openrtb2.Format{
							{
								W: 300,
								H: 250,
							},
							{
								W: 400,
								H: 300,
							},
						},
					},
					Video: &openrtb2.Video{},
				},
			},
			want: &openrtb2.Video{
				MIMEs:          []string{"video/mp4"},
				MinDuration:    0,
				MaxDuration:    30,
				StartDelay:     adcom1.StartPreRoll.Ptr(),
				Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper, adcom1.CreativeVAST30Wrapper, adcom1.CreativeVAST40, adcom1.CreativeVAST40Wrapper, adcom1.CreativeVAST41, adcom1.CreativeVAST41Wrapper, adcom1.CreativeVAST42, adcom1.CreativeVAST42Wrapper},
				Placement:      adcom1.VideoPlacementInBanner,
				Plcmt:          adcom1.VideoPlcmtNoContent,
				Linearity:      adcom1.LinearityLinear,
				Skip:           ptrutil.ToPtr[int8](0),
				SkipMin:        0,
				SkipAfter:      0,
				BoxingAllowed:  ptrutil.ToPtr[int8](1),
				PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackPageLoadSoundOff},
				PlaybackEnd:    adcom1.PlaybackCompletion,
				Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryProgressive, adcom1.DeliveryDownload},
				W:              ptrutil.ToPtr[int64](300),
				H:              ptrutil.ToPtr[int64](250),
			},
		},
		{
			name: "banner has the width and height in in the both banner and format object",
			args: args{
				imp: &openrtb2.Imp{
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](300),
						H: ptrutil.ToPtr[int64](250),
						Format: []openrtb2.Format{
							{
								W: 200,
								H: 150,
							},
							{
								W: 400,
								H: 300,
							},
						},
					},
					Video: &openrtb2.Video{},
				},
			},
			want: &openrtb2.Video{
				MIMEs:          []string{"video/mp4"},
				MinDuration:    0,
				MaxDuration:    30,
				StartDelay:     adcom1.StartPreRoll.Ptr(),
				Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST10, adcom1.CreativeVAST20, adcom1.CreativeVAST30, adcom1.CreativeVAST10Wrapper, adcom1.CreativeVAST20Wrapper, adcom1.CreativeVAST30Wrapper, adcom1.CreativeVAST40, adcom1.CreativeVAST40Wrapper, adcom1.CreativeVAST41, adcom1.CreativeVAST41Wrapper, adcom1.CreativeVAST42, adcom1.CreativeVAST42Wrapper},
				Placement:      adcom1.VideoPlacementInBanner,
				Plcmt:          adcom1.VideoPlcmtNoContent,
				Linearity:      adcom1.LinearityLinear,
				Skip:           ptrutil.ToPtr[int8](0),
				SkipMin:        0,
				SkipAfter:      0,
				BoxingAllowed:  ptrutil.ToPtr[int8](1),
				PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackPageLoadSoundOff},
				PlaybackEnd:    adcom1.PlaybackCompletion,
				Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryProgressive, adcom1.DeliveryDownload},
				W:              ptrutil.ToPtr[int64](300),
				H:              ptrutil.ToPtr[int64](250),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateAmpImpVideoWithDefault(tt.args.imp)
		})
	}
}

func TestGetW(t *testing.T) {
	type args struct {
		imp *openrtb2.Imp
	}
	tests := []struct {
		name string
		args args
		want *int64
	}{
		{
			name: "Empty banner and format",
			args: args{
				imp: &openrtb2.Imp{
					Banner: &openrtb2.Banner{
						Format: nil,
					},
				},
			},
			want: nil,
		},
		{
			name: "both banner and format are present",
			args: args{
				imp: &openrtb2.Imp{
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr[int64](300),
						Format: []openrtb2.Format{
							{
								W: 400,
							},
						},
					},
				},
			},
			want: ptrutil.ToPtr[int64](300),
		},
		{
			name: "only format is present",
			args: args{
				imp: &openrtb2.Imp{
					Banner: &openrtb2.Banner{
						Format: []openrtb2.Format{
							{
								W: 400,
							},
						},
					},
				},
			},
			want: ptrutil.ToPtr[int64](400),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getW(tt.args.imp)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestGetH(t *testing.T) {
	type args struct {
		imp *openrtb2.Imp
	}
	tests := []struct {
		name string
		args args
		want *int64
	}{
		{
			name: "Empty banner and format",
			args: args{
				imp: &openrtb2.Imp{
					Banner: &openrtb2.Banner{
						Format: nil,
					},
				},
			},
			want: nil,
		},
		{
			name: "both banner and format are present",
			args: args{
				imp: &openrtb2.Imp{
					Banner: &openrtb2.Banner{
						H: ptrutil.ToPtr[int64](300),
						Format: []openrtb2.Format{
							{
								H: 400,
							},
						},
					},
				},
			},
			want: ptrutil.ToPtr[int64](300),
		},
		{
			name: "only format is present",
			args: args{
				imp: &openrtb2.Imp{
					Banner: &openrtb2.Banner{
						Format: []openrtb2.Format{
							{
								H: 400,
							},
						},
					},
				},
			},
			want: ptrutil.ToPtr[int64](400),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getH(tt.args.imp)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestIsVastUnwrapEnabled(t *testing.T) {

	type args struct {
		PartnerConfigMap  map[int]map[string]string
		VASTUnwrapTraffic int
	}
	tests := []struct {
		name         string
		args         args
		randomNumber int
		want         bool
	}{
		{
			name: "vastunwrap is enabled and traffic percent in DB and config, DB percent should be preferred",
			args: args{
				PartnerConfigMap: map[int]map[string]string{
					-1: {
						models.VastUnwrapperEnableKey:      "1",
						models.VastUnwrapTrafficPercentKey: "90",
					},
				},
				VASTUnwrapTraffic: 9,
			},
			randomNumber: 10,
			want:         true,
		},
		{
			name: "vastunwrap is enabled and DB traffic percent is less than random number",
			args: args{
				PartnerConfigMap: map[int]map[string]string{
					-1: {
						models.VastUnwrapperEnableKey:      "1",
						models.VastUnwrapTrafficPercentKey: "90",
					},
				},
				VASTUnwrapTraffic: 0,
			},
			randomNumber: 91,
			want:         false,
		},
		{
			name: "vastunwrap is dissabled and config traffic percent is less than random number",
			args: args{
				PartnerConfigMap: map[int]map[string]string{
					-1: {
						models.VastUnwrapperEnableKey: "0",
					},
				},
				VASTUnwrapTraffic: 5,
			},
			randomNumber: 7,
			want:         false,
		},
		{
			name: "vastunwrap is enabled and traffic percent not present in DB, random num higher than traffic percent",
			args: args{
				PartnerConfigMap: map[int]map[string]string{
					-1: {
						models.VastUnwrapperEnableKey: "1",
					},
				},
				VASTUnwrapTraffic: 5,
			},
			randomNumber: 10,
			want:         false,
		},

		{
			name: "vastunwrap is enabled and traffic percent not present in DB, random num less than traffic percent",
			args: args{
				PartnerConfigMap: map[int]map[string]string{
					-1: {
						models.VastUnwrapperEnableKey: "1",
					},
				},
				VASTUnwrapTraffic: 10,
			},
			randomNumber: 9,
			want:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetRandomNumberIn1To100 = func() int {
				return tt.randomNumber
			}
			got := isVastUnwrapEnabled(tt.args.PartnerConfigMap, tt.args.VASTUnwrapTraffic)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
