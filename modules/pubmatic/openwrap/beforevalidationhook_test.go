package openwrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/adcom1"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/hooks/hookanalytics"
	"github.com/prebid/prebid-server/hooks/hookstage"
	ow_adapters "github.com/prebid/prebid-server/modules/pubmatic/openwrap/adapters"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	mock_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
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
				W:     200,
				H:     300,
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
					DisplayID:          1234,
					StartTime:          1234,
					DevicePlatform:     1234,
					LoggerImpressionID: "1234",
					SSAI:               "",
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
					DisplayID:          1234,
					StartTime:          1234,
					DevicePlatform:     1234,
					LoggerImpressionID: "1234",
					SSAI:               "1234",
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
			want: 10,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSlotEnabled(tt.args.videoAdUnitCtx, tt.args.bannerAdUnitCtx)
			assert.Equal(t, tt.want, got)
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
							W:     200,
							H:     300,
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
							W:     200,
							H:     300,
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
	type fields struct {
		cfg          config.Config
		cache        cache.Cache
		metricEngine metrics.MetricsEngine
	}
	type args struct {
		rCtx models.RequestCtx
		imp  *openrtb2.Imp
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *openrtb2.Imp
	}{
		{
			name: "imp.video_is_nil",
			args: args{
				imp: &openrtb2.Imp{
					Video: nil,
				},
			},
			want: &openrtb2.Imp{
				Video: nil,
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
			want: &openrtb2.Imp{
				ID:    "testImp",
				Video: &openrtb2.Video{},
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
			want: &openrtb2.Imp{
				ID:          "testImp",
				Video:       &openrtb2.Video{},
				BidFloor:    2.0,
				BidFloorCur: "USD",
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
			want: &openrtb2.Imp{
				ID:    "testImp",
				Video: &openrtb2.Video{},
				Exp:   10,
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
						W: 200,
						H: 300,
					},
				},
			},
			want: &openrtb2.Imp{
				ID: "testImp",
				Video: &openrtb2.Video{
					W: 200,
					H: 300,
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
						W: 200,
						H: 300,
					},
				},
			},
			want: &openrtb2.Imp{
				ID:    "testImp",
				Video: nil,
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
												W:              640,
												H:              480,
												Sequence:       2,
												BoxingAllowed:  1,
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
			want: &openrtb2.Imp{
				ID: "testImp",
				Video: &openrtb2.Video{
					W:              640,
					H:              480,
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
					BoxingAllowed:  1,
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
						W:           640,
						H:           480,
						MinDuration: 20,
						MaxDuration: 60,
						Skip:        ptrutil.ToPtr(int8(2)),
						SkipMin:     10,
						SkipAfter:   20,
					},
				},
			},
			want: &openrtb2.Imp{
				ID: "testImp",
				Video: &openrtb2.Video{
					W:           640,
					H:           480,
					MinDuration: 20,
					MaxDuration: 60,
					Skip:        ptrutil.ToPtr(int8(2)),
					SkipMin:     10,
					SkipAfter:   20,
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
			m.applyVideoAdUnitConfig(tt.args.rCtx, tt.args.imp)
			assert.Equal(t, tt.args.imp, tt.want, "Imp video is not upadted as expected from adunit config")
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
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *openrtb2.Imp
	}{
		{
			name: "imp.banner_is_nil",
			args: args{
				imp: &openrtb2.Imp{
					Banner: nil,
				},
			},
			want: &openrtb2.Imp{
				Banner: nil,
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
			want: &openrtb2.Imp{
				ID:     "testImp",
				Banner: &openrtb2.Banner{},
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
			want: &openrtb2.Imp{
				ID:          "testImp",
				Banner:      &openrtb2.Banner{},
				BidFloor:    2.0,
				BidFloorCur: "USD",
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
			want: &openrtb2.Imp{
				ID:     "testImp",
				Banner: &openrtb2.Banner{},
				Exp:    10,
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
			want: &openrtb2.Imp{
				ID: "testImp",
				Banner: &openrtb2.Banner{
					W: ptrutil.ToPtr[int64](200),
					H: ptrutil.ToPtr[int64](300),
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
			want: &openrtb2.Imp{
				ID:     "testImp",
				Banner: nil,
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
			assert.Equal(t, tt.args.imp, tt.want, "Imp banner is not upadted as expected from adunit config")
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

func TestOpenWrap_handleBeforeValidationHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)

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
		name                 string
		fields               fields
		args                 args
		want                 hookstage.HookResult[hookstage.BeforeValidationRequestPayload]
		setup                func()
		isRequestNotRejected bool
		wantErr              bool
	}{
		{
			name: "empty_module_context",
			args: args{
				ctx:       context.Background(),
				moduleCtx: hookstage.ModuleInvocationContext{},
			},
			fields: fields{
				cache:        mockCache,
				metricEngine: mockEngine,
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:        false,
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
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("", nbr.InvalidPublisherID)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: nbr.InvalidPublisherID,
				Errors:  []string{"ErrInvalidPublisherID"},
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
				mockEngine.EXPECT().RecordBadRequests(rctx.Endpoint, getPubmaticErrorCode(nbr.InvalidRequest))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", nbr.InvalidRequest)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: nbr.InvalidRequest,
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
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", nbr.InvalidProfileConfiguration)
				mockEngine.EXPECT().RecordPublisherInvalidProfileRequests(rctx.Endpoint, "5890", rctx.ProfileIDStr)
				mockEngine.EXPECT().RecordPublisherInvalidProfileImpressions("5890", rctx.ProfileIDStr, gomock.Any())
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: nbr.InvalidProfileConfiguration,
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
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", nbr.InvalidProfileConfiguration)
				mockEngine.EXPECT().RecordPublisherInvalidProfileRequests(rctx.Endpoint, "5890", rctx.ProfileIDStr)
				mockEngine.EXPECT().RecordPublisherInvalidProfileImpressions("5890", rctx.ProfileIDStr, gomock.Any())
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: nbr.InvalidProfileConfiguration,
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
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", nbr.InvalidPlatform)
				mockEngine.EXPECT().RecordPublisherInvalidProfileRequests(rctx.Endpoint, "5890", rctx.ProfileIDStr)
				mockEngine.EXPECT().RecordPublisherInvalidProfileImpressions("5890", rctx.ProfileIDStr, gomock.Any())
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: nbr.InvalidPlatform,
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
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", nbr.AllPartnerThrottled)
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: nbr.AllPartnerThrottled,
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
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", nbr.InvalidImpressionTagID)
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: nbr.InvalidImpressionTagID,
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
							testRctx.Endpoint = models.EndpointOWS2S
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
				mockEngine.EXPECT().RecordBadRequests(models.EndpointOWS2S, getPubmaticErrorCode(nbr.InvalidImpressionTagID))
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", nbr.InvalidImpressionTagID)
				mockEngine.EXPECT().RecordPublisherRequests(models.EndpointOWS2S, "5890", rctx.Platform)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: nbr.InvalidImpressionTagID,
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
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", nbr.InternalError)
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: nbr.InternalError,
				Errors:  []string{"failed to parse imp.ext: 123"},
			},
			wantErr: true,
		},
		{
			name: "allSotsDisabled",
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
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", nbr.AllSlotsDisabled)
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeVideo, "5890", "1234")
				mockEngine.EXPECT().RecordImpDisabledViaConfigStats(models.ImpTypeBanner, "5890", "1234")
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: nbr.AllSlotsDisabled,
				Errors:  []string{"All slots disabled"},
			},
			wantErr: false,
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
				mockEngine.EXPECT().RecordNobidErrPrebidServerRequests("5890", nbr.ServerSidePartnerNotConfigured)
				mockEngine.EXPECT().RecordPublisherRequests(rctx.Endpoint, "5890", rctx.Platform)
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:  true,
				NbrCode: nbr.ServerSidePartnerNotConfigured,
				Errors:  []string{"server side partner not found"},
			},
			wantErr: false,
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
				bidrequest: json.RawMessage(`{"id":"123-456-789","imp":[{"id":"123","banner":{"format":[{"w":728,"h":90},{"w":300,"h":250}],"w":700,"h":900},"video":{"mimes":["video/mp4","video/mpeg"],"w":640,"h":480},"tagid":"adunit","ext":{"bidder":{"pubmatic":{"keywords":[{"key":"pmzoneid","value":["val1","val2"]}]}},"prebid":{}}}],"site":{"domain":"test.com","page":"www.test.com","publisher":{"id":"5890"}},"device":{"ua":"Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36","ip":"123.145.167.10"},"user":{"id":"119208432","buyeruid":"1rwe432","yob":1980,"gender":"F","geo":{"country":"US","region":"CA","metro":"90001","city":"Alamo"}},"wseat":["Wseat_0","Wseat_1"],"bseat":["Bseat_0","Bseat_1"],"cur":["cur_0","cur_1"],"wlang":["Wlang_0","Wlang_1"],"bcat":["bcat_0","bcat_1"],"badv":["badv_0","badv_1"],"bapp":["bapp_0","bapp_1"],"source":{"ext":{"omidpn":"MyIntegrationPartner","omidpv":"7.1"}},"ext":{"prebid":{},"wrapper":{"test":123,"profileid":123,"versionid":1,"wiid":"test_display_wiid"}}}`),
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
			},
			want: hookstage.HookResult[hookstage.BeforeValidationRequestPayload]{
				Reject:        false,
				NbrCode:       0,
				ChangeSet:     hookstage.ChangeSet[hookstage.BeforeValidationRequestPayload]{},
				DebugMessages: []string{"new imp: {\"123\":{\"ImpID\":\"\",\"TagID\":\"adunit\",\"Div\":\"\",\"Secure\":0,\"IsRewardInventory\":null,\"Banner\":true,\"Video\":{\"mimes\":[\"video/mp4\",\"video/mpeg\"],\"w\":640,\"h\":480},\"IncomingSlots\":[\"700x900\",\"728x90\",\"300x250\",\"640x480v\"],\"Type\":\"video\",\"Bidders\":{\"appnexus\":{\"PartnerID\":2,\"PrebidBidderCode\":\"appnexus\",\"MatchedSlot\":\"adunit@700x900\",\"KGP\":\"_AU_@_W_x_H_\",\"KGPV\":\"\",\"IsRegex\":false,\"Params\":{\"placementId\":0,\"site\":\"12313\",\"adtag\":\"45343\"}}},\"NonMapped\":{},\"NewExt\":{\"prebid\":{\"bidder\":{\"appnexus\":{\"placementId\":0,\"site\":\"12313\",\"adtag\":\"45343\"}}}},\"BidCtx\":{},\"BannerAdUnitCtx\":{\"MatchedSlot\":\"\",\"IsRegex\":false,\"MatchedRegex\":\"\",\"SelectedSlotAdUnitConfig\":null,\"AppliedSlotAdUnitConfig\":null,\"UsingDefaultConfig\":false,\"AllowedConnectionTypes\":null},\"VideoAdUnitCtx\":{\"MatchedSlot\":\"\",\"IsRegex\":false,\"MatchedRegex\":\"\",\"SelectedSlotAdUnitConfig\":null,\"AppliedSlotAdUnitConfig\":null,\"UsingDefaultConfig\":false,\"AllowedConnectionTypes\":null}}}", "new request.ext: {\"prebid\":{\"bidderparams\":{\"pubmatic\":{\"wiid\":\"\"}},\"debug\":true,\"targeting\":{\"pricegranularity\":{\"precision\":2,\"ranges\":[{\"min\":0,\"max\":5,\"increment\":0.05},{\"min\":5,\"max\":10,\"increment\":0.1},{\"min\":10,\"max\":20,\"increment\":0.5}]},\"mediatypepricegranularity\":{},\"includewinners\":true,\"includebidderkeys\":true},\"floors\":{\"enforcement\":{\"enforcepbs\":true},\"enabled\":true},\"macros\":{\"[PLATFORM]\":\"3\",\"[PROFILE_ID]\":\"1234\",\"[PROFILE_VERSION]\":\"1\",\"[UNIX_TIMESTAMP]\":\"0\",\"[WRAPPER_IMPRESSION_ID]\":\"\"}}}"},
				AnalyticsTags: hookanalytics.Analytics{},
			},
			wantErr:              false,
			isRequestNotRejected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			ow_adapters.InitBidders(tt.fields.cfg)
			m := OpenWrap{
				cfg:          tt.fields.cfg,
				cache:        tt.fields.cache,
				metricEngine: tt.fields.metricEngine,
			}
			bidrequest := &openrtb2.BidRequest{}
			json.Unmarshal(tt.args.bidrequest, bidrequest)
			tt.args.payload.BidRequest = bidrequest
			got, err := m.handleBeforeValidationHook(tt.args.ctx, tt.args.moduleCtx, tt.args.payload)
			if tt.isRequestNotRejected {
				assert.Equal(t, tt.want.Reject, got.Reject)
				assert.Equal(t, tt.want.NbrCode, got.NbrCode)
				assert.NotEmpty(t, tt.want.DebugMessages)
				return
			}
			fmt.Println(got.DebugMessages)
			if (err != nil) != tt.wantErr {
				assert.Equal(t, tt.wantErr, err != nil)
				return
			}
			assert.Equal(t, tt.want, got)
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
					Gpid: "/7578294/adunit1",
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
					Gpid: "/7578294/adunit123",
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
					Gpid: "/7578294/adunit123",
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
					Gpid: "/7578294/adunit123",
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
					Gpid: "/43743431/DMDemo#Div1",
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
