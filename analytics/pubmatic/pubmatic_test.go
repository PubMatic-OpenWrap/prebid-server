package pubmatic

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/analytics"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/hooks/hookanalytics"
	"github.com/prebid/prebid-server/v2/hooks/hookexecution"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
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
	tests := []struct {
		name             string
		ao               *analytics.AuctionObject
		RestoredResponse *openrtb2.BidResponse
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
			name: "AppLovinMax request . RestoreBidResponse for logger and wakanda enable",
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
																	PubID:    5890,
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
		},
	}
	for _, tt := range tests {
		HTTPLogger{}.LogAuctionObject(tt.ao)
		assert.Equal(t, tt.RestoredResponse, tt.ao.Response, tt.name)
	}
}
