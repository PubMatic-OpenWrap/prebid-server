package pubmatic

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/analytics"
	"github.com/prebid/prebid-server/v3/analytics/pubmatic/mhttp"
	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/hooks/hookanalytics"
	"github.com/prebid/prebid-server/v3/hooks/hookexecution"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/wakanda"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPLogger(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.PubMaticWL
	}{
		{
			name: "check if NewHTTPLogger returns nil",
			cfg: config.PubMaticWL{
				MaxClients:     5,
				MaxConnections: 50,
				MaxCalls:       5,
				RespTimeout:    50,
			},
		},
	}
	for _, tt := range tests {
		module := NewHTTPLogger(tt.cfg)
		assert.NotNil(t, module, tt.name)
	}
}

// TestLogAuctionObject just increases code coverage, it does not validate anything
func TestLogAuctionObject(t *testing.T) {
	oldSend := send
	send = func(rCtx *models.RequestCtx, url string, headers http.Header, mhc mhttp.MultiHttpContextInterface) {}
	defer func() {
		send = oldSend
	}()
	tests := []struct {
		name             string
		ao               *analytics.AuctionObject
		RestoredResponse *openrtb2.BidResponse
		wantWakanda      wakanda.WakandaDebug
	}{
		{
			name: "rctx is nil",
			ao:   &analytics.AuctionObject{},
		},
		{
			name: "rctx is present",
			ao: &analytics.AuctionObject{
				HookExecutionOutcome: []hookexecution.StageOutcome{
					{
						Groups: []hookexecution.GroupOutcome{
							{
								InvocationResults: []hookexecution.HookOutcome{
									{
										AnalyticsTags: hookanalytics.Analytics{
											Activities: []hookanalytics.Activity{
												{
													Results: []hookanalytics.Result{
														{
															Values: map[string]interface{}{
																"request-ctx": &models.RequestCtx{},
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
					},
				},
			},
		},
		{
			name: "AppLovinMax request . RestoreBidResponse for logger",
			ao: &analytics.AuctionObject{
				HookExecutionOutcome: []hookexecution.StageOutcome{
					{
						Groups: []hookexecution.GroupOutcome{
							{
								InvocationResults: []hookexecution.HookOutcome{
									{
										AnalyticsTags: hookanalytics.Analytics{
											Activities: []hookanalytics.Activity{
												{
													Results: []hookanalytics.Result{
														{
															Values: map[string]interface{}{
																"request-ctx": &models.RequestCtx{
																	Endpoint: models.EndpointAppLovinMax,
																	Debug:    false,
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
						},
					},
				},
				Response: &openrtb2.BidResponse{
					ID:    "123",
					BidID: "bid-id-1",
					Cur:   "USD",
					SeatBid: []openrtb2.SeatBid{
						{
							Seat: "pubmatic",
							Bid: []openrtb2.Bid{
								{
									ID:    "bid-id-1",
									ImpID: "imp_1",
									Ext:   json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"bid-id-1\",\"impid\":\"imp_1\",\"price\":0}],\"seat\":\"pubmatic\"}],\"bidid\":\"bid-id-1\",\"cur\":\"USD\",\"ext\":{\"matchedimpression\":{\"appnexus\":50,\"pubmatic\":50}}}\r\n"}`),
								},
							},
						},
					},
				},
			},
			RestoredResponse: &openrtb2.BidResponse{
				ID:    "123",
				BidID: "bid-id-1",
				Cur:   "USD",
				SeatBid: []openrtb2.SeatBid{
					{
						Seat: "pubmatic",
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-id-1",
								ImpID: "imp_1",
							},
						},
					},
				},
				Ext: json.RawMessage(`{"matchedimpression":{"appnexus":50,"pubmatic":50}}`),
			},
		},
		{
			name: "logger_disabled",
			ao: &analytics.AuctionObject{
				HookExecutionOutcome: []hookexecution.StageOutcome{
					{
						Groups: []hookexecution.GroupOutcome{
							{
								InvocationResults: []hookexecution.HookOutcome{
									{
										AnalyticsTags: hookanalytics.Analytics{
											Activities: []hookanalytics.Activity{
												{
													Results: []hookanalytics.Result{
														{
															Values: map[string]interface{}{
																"request-ctx": &models.RequestCtx{
																	LoggerDisabled: true,
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
						},
					},
				},
			},
		},
		{
			name: "AppLovinMax request, RestoreBidResponse for logger and wakanda enable",
			ao: &analytics.AuctionObject{
				HookExecutionOutcome: []hookexecution.StageOutcome{
					{
						Groups: []hookexecution.GroupOutcome{
							{
								InvocationResults: []hookexecution.HookOutcome{
									{
										AnalyticsTags: hookanalytics.Analytics{
											Activities: []hookanalytics.Activity{
												{
													Results: []hookanalytics.Result{
														{
															Values: map[string]interface{}{
																"request-ctx": &models.RequestCtx{
																	Endpoint:     models.EndpointAppLovinMax,
																	Debug:        false,
																	PubID:        5890,
																	PubIDStr:     "5890",
																	ProfileID:    1234,
																	ProfileIDStr: "1234",
																	WakandaDebug: &wakanda.Debug{
																		Enabled: true,
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
							},
						},
					},
				},
				Response: &openrtb2.BidResponse{
					ID:    "123",
					BidID: "bid-id-1",
					Cur:   "USD",
					SeatBid: []openrtb2.SeatBid{
						{
							Seat: "pubmatic",
							Bid: []openrtb2.Bid{
								{
									ID:    "bid-id-1",
									ImpID: "imp_1",
									Ext:   json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"bid-id-1\",\"impid\":\"imp_1\",\"price\":0}],\"seat\":\"pubmatic\"}],\"bidid\":\"bid-id-1\",\"cur\":\"USD\",\"ext\":{\"matchedimpression\":{\"appnexus\":50,\"pubmatic\":50}}}\r\n"}`),
								},
							},
						},
					},
				},
				RequestWrapper: &openrtb_ext.RequestWrapper{
					BidRequest: &openrtb2.BidRequest{},
				},
			},
			RestoredResponse: &openrtb2.BidResponse{
				ID:    "123",
				BidID: "bid-id-1",
				Cur:   "USD",
				SeatBid: []openrtb2.SeatBid{
					{
						Seat: "pubmatic",
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-id-1",
								ImpID: "imp_1",
								Ext:   json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"bid-id-1\",\"impid\":\"imp_1\",\"price\":0}],\"seat\":\"pubmatic\"}],\"bidid\":\"bid-id-1\",\"cur\":\"USD\",\"ext\":{\"matchedimpression\":{\"appnexus\":50,\"pubmatic\":50}}}\r\n"}`),
							},
						},
					},
				},
			},
			wantWakanda: &wakanda.Debug{
				Enabled:     true,
				FolderPaths: nil,
				DebugLevel:  0,
				DebugData: wakanda.DebugData{
					HTTPRequest:        nil,
					HTTPRequestBody:    nil,
					HTTPResponse:       nil,
					HTTPResponseBody:   "{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"bid-id-1\",\"impid\":\"imp_1\",\"price\":0,\"ext\":{\"signaldata\":\"{\\\"id\\\":\\\"123\\\",\\\"seatbid\\\":[{\\\"bid\\\":[{\\\"id\\\":\\\"bid-id-1\\\",\\\"impid\\\":\\\"imp_1\\\",\\\"price\\\":0}],\\\"seat\\\":\\\"pubmatic\\\"}],\\\"bidid\\\":\\\"bid-id-1\\\",\\\"cur\\\":\\\"USD\\\",\\\"ext\\\":{\\\"matchedimpression\\\":{\\\"appnexus\\\":50,\\\"pubmatic\\\":50}}}\\r\\n\"}}],\"seat\":\"pubmatic\"}],\"bidid\":\"bid-id-1\",\"cur\":\"USD\"}",
					PrebidHTTPRequest:  nil,
					PrebidRequestBody:  nil,
					PrebidHTTPResponse: nil,
					OpenRTB:            &openrtb2.BidRequest{},
					WinningBid:         false,
					Logger:             json.RawMessage(`{"pubid":5890,"pid":"1234","pdvid":"0","sl":1,"dvc":{},"ft":0,"it":"sdk","geo":{}}`),
				},
				Config: wakanda.Wakanda{
					SFTP: wakanda.SFTP{
						User:        "",
						Password:    "",
						ServerIP:    "",
						Destination: "",
					},
					HostName:              "",
					DCName:                "",
					PodName:               "",
					MaxDurationInMin:      0,
					CleanupFrequencyInMin: 0,
				},
			},
		},
		{
			name: "AppLovinMax request, RestoreBidResponse for logger and wakanda disable",
			ao: &analytics.AuctionObject{
				HookExecutionOutcome: []hookexecution.StageOutcome{
					{
						Groups: []hookexecution.GroupOutcome{
							{
								InvocationResults: []hookexecution.HookOutcome{
									{
										AnalyticsTags: hookanalytics.Analytics{
											Activities: []hookanalytics.Activity{
												{
													Results: []hookanalytics.Result{
														{
															Values: map[string]interface{}{
																"request-ctx": &models.RequestCtx{
																	Endpoint:     models.EndpointAppLovinMax,
																	Debug:        false,
																	PubID:        5890,
																	PubIDStr:     "5890",
																	ProfileID:    1234,
																	ProfileIDStr: "1234",
																	WakandaDebug: &wakanda.Debug{
																		Enabled: false,
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
							},
						},
					},
				},
				Response: &openrtb2.BidResponse{
					ID:    "123",
					BidID: "bid-id-1",
					Cur:   "USD",
					SeatBid: []openrtb2.SeatBid{
						{
							Seat: "pubmatic",
							Bid: []openrtb2.Bid{
								{
									ID:    "bid-id-1",
									ImpID: "imp_1",
									Ext:   json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"bid-id-1\",\"impid\":\"imp_1\",\"price\":0}],\"seat\":\"pubmatic\"}],\"bidid\":\"bid-id-1\",\"cur\":\"USD\",\"ext\":{\"matchedimpression\":{\"appnexus\":50,\"pubmatic\":50}}}\r\n"}`),
								},
							},
						},
					},
				},
				RequestWrapper: &openrtb_ext.RequestWrapper{
					BidRequest: &openrtb2.BidRequest{},
				},
			},
			RestoredResponse: &openrtb2.BidResponse{
				ID:    "123",
				BidID: "bid-id-1",
				Cur:   "USD",
				SeatBid: []openrtb2.SeatBid{
					{
						Seat: "pubmatic",
						Bid: []openrtb2.Bid{
							{
								ID:    "bid-id-1",
								ImpID: "imp_1",
								Ext:   json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"bid-id-1\",\"impid\":\"imp_1\",\"price\":0}],\"seat\":\"pubmatic\"}],\"bidid\":\"bid-id-1\",\"cur\":\"USD\",\"ext\":{\"matchedimpression\":{\"appnexus\":50,\"pubmatic\":50}}}\r\n"}`),
							},
						},
					},
				},
			},
			wantWakanda: &wakanda.Debug{},
		},
	}
	for _, tt := range tests {
		HTTPLogger{}.LogAuctionObject(tt.ao)
		assert.Equal(t, tt.RestoredResponse, tt.ao.Response, tt.name)
		var rctx *models.RequestCtx
		if tt.ao != nil && tt.ao.HookExecutionOutcome != nil {
			rctx = tt.ao.HookExecutionOutcome[0].Groups[0].InvocationResults[0].AnalyticsTags.Activities[0].Results[0].Values["request-ctx"].(*models.RequestCtx)
		}
		if rctx != nil {
			assert.Equal(t, tt.wantWakanda, rctx.WakandaDebug, tt.name)
		}
	}
}
