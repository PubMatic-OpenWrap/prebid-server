package openrtb2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/analytics/pubmatic"
	"github.com/prebid/prebid-server/v3/hooks/hookanalytics"
	"github.com/prebid/prebid-server/v3/hooks/hookexecution"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/wakanda"
	"github.com/prebid/prebid-server/v3/util/ptrutil"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/analytics"
	"github.com/prebid/prebid-server/v3/exchange"
	"github.com/prebid/prebid-server/v3/metrics"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRecordRejectedBids(t *testing.T) {

	type args struct {
		pubid       string
		seatNonBids []openrtb_ext.SeatNonBid
	}

	type want struct {
		expectedCalls int
	}

	tests := []struct {
		description string
		args        args
		want        want
	}{
		{
			description: "empty rejected bids",
			args: args{
				seatNonBids: []openrtb_ext.SeatNonBid{},
			},
			want: want{
				expectedCalls: 0,
			},
		},
		{
			description: "rejected bids",
			args: args{
				pubid: "1010",
				seatNonBids: []openrtb_ext.SeatNonBid{
					{
						NonBid: []openrtb_ext.NonBid{
							{
								StatusCode: int(exchange.ResponseRejectedCreativeAdvertiserExclusions),
							},
							{
								StatusCode: int(exchange.ResponseRejectedBelowDealFloor),
							},
							{
								StatusCode: int(exchange.ResponseRejectedCreativeAdvertiserExclusions),
							},
						},
						Seat: "pubmatic",
					},
					{
						NonBid: []openrtb_ext.NonBid{
							{
								StatusCode: int(exchange.ResponseRejectedBelowDealFloor),
							},
						},
						Seat: "appnexus",
					},
				},
			},
			want: want{
				expectedCalls: 4,
			},
		},
	}

	for _, test := range tests {
		me := &metrics.MetricsEngineMock{}
		me.On("RecordRejectedBids", mock.Anything, mock.Anything, mock.Anything).Return()

		recordRejectedBids(test.args.pubid, test.args.seatNonBids, me)
		me.AssertNumberOfCalls(t, "RecordRejectedBids", test.want.expectedCalls)
	}
}

func TestUpdateResponseExtOW(t *testing.T) {
	uuidFunc := pubmatic.GetUUID
	defer func() {
		pubmatic.GetUUID = uuidFunc
	}()

	pubmatic.GetUUID = func() string { return "uuid" }
	type args struct {
		w           http.ResponseWriter
		bidResponse *openrtb2.BidResponse
		ao          analytics.AuctionObject
	}
	tests := []struct {
		name             string
		args             args
		want             *openrtb2.BidResponse
		RestoredResponse *openrtb2.BidResponse
		rejectResponse   bool
	}{
		{
			name: "empty bid response",
			args: args{
				bidResponse: nil,
				ao: analytics.AuctionObject{
					Response: nil,
				},
			},
			want:             nil,
			RestoredResponse: nil,
		},
		{
			name: "rctx is nil",
			args: args{
				bidResponse: &openrtb2.BidResponse{},
				ao: analytics.AuctionObject{
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
																	"request-ctx": nil,
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
					Response: &openrtb2.BidResponse{},
				},
			},
			want:             &openrtb2.BidResponse{},
			RestoredResponse: &openrtb2.BidResponse{},
		},
		{
			name: "debug is enabled and endpoint is other than applovinmax",
			args: args{
				bidResponse: &openrtb2.BidResponse{
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
				ao: analytics.AuctionObject{
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
																		PubID:    5890,
																		Debug:    true,
																		Endpoint: models.EndpointV25,
																		ImpBidCtx: map[string]models.ImpCtx{
																			"imp_1": {
																				IncomingSlots:     []string{"0x0", "100x200"},
																				IsRewardInventory: ptrutil.ToPtr(int8(1)),
																				SlotName:          "imp_1_tagid_1",
																				AdUnitName:        "tagid_1",
																			},
																		},
																		WakandaDebug: &wakanda.Debug{},
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
					RequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{
							Imp: []openrtb2.Imp{
								{
									ID:    "imp_1",
									TagID: "tagid_1",
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
									},
								},
							},
						},
						Ext: json.RawMessage(`{"matchedimpression":{"appnexus":50,"pubmatic":50}}`),
					},
				},
			},
			want: &openrtb2.BidResponse{
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
				Ext: json.RawMessage(`{"matchedimpression":{"appnexus":50,"pubmatic":50},"owlogger":"?json=%7B%22pubid%22%3A5890%2C%22pid%22%3A%220%22%2C%22pdvid%22%3A%220%22%2C%22sl%22%3A1%2C%22s%22%3A%5B%7B%22sid%22%3A%22uuid%22%2C%22sn%22%3A%22imp_1_tagid_1%22%2C%22sz%22%3A%5B%220x0%22%2C%22100x200%22%5D%2C%22au%22%3A%22tagid_1%22%2C%22ps%22%3A%5B%7B%22pn%22%3A%22pubmatic%22%2C%22bc%22%3A%22pubmatic%22%2C%22kgpv%22%3A%22%22%2C%22kgpsv%22%3A%22%22%2C%22psz%22%3A%220x0%22%2C%22af%22%3A%22%22%2C%22eg%22%3A0%2C%22en%22%3A0%2C%22l1%22%3A0%2C%22l2%22%3A0%2C%22t%22%3A0%2C%22wb%22%3A0%2C%22bidid%22%3A%22bid-id-1%22%2C%22origbidid%22%3A%22bid-id-1%22%2C%22di%22%3A%22-1%22%2C%22dc%22%3A%22%22%2C%22db%22%3A1%2C%22ss%22%3A1%2C%22mi%22%3A0%2C%22ocpm%22%3A0%2C%22ocry%22%3A%22USD%22%7D%5D%2C%22rwrd%22%3A1%7D%5D%2C%22dvc%22%3A%7B%7D%2C%22ft%22%3A0%2C%22it%22%3A%22sdk%22%2C%22geo%22%3A%7B%7D%7D&pubid=5890"}`),
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
			name: "debug is enabled and endpoint is AppLovinMax",
			args: args{
				bidResponse: &openrtb2.BidResponse{
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
				ao: analytics.AuctionObject{
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
																		PubID:    5890,
																		Debug:    true,
																		Endpoint: models.EndpointAppLovinMax,
																		ImpBidCtx: map[string]models.ImpCtx{
																			"imp_1": {
																				IncomingSlots:     []string{"0x0", "100x200"},
																				IsRewardInventory: ptrutil.ToPtr(int8(1)),
																				SlotName:          "imp_1_tagid_1",
																				AdUnitName:        "tagid_1",
																			},
																		},
																		WakandaDebug: &wakanda.Debug{},
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
					RequestWrapper: &openrtb_ext.RequestWrapper{
						BidRequest: &openrtb2.BidRequest{
							Imp: []openrtb2.Imp{
								{
									ID:    "imp_1",
									TagID: "tagid_1",
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
			},
			want: &openrtb2.BidResponse{
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
				Ext: json.RawMessage(`{"owlogger":"?json=%7B%22pubid%22%3A5890%2C%22pid%22%3A%220%22%2C%22pdvid%22%3A%220%22%2C%22sl%22%3A1%2C%22s%22%3A%5B%7B%22sid%22%3A%22uuid%22%2C%22sn%22%3A%22imp_1_tagid_1%22%2C%22sz%22%3A%5B%220x0%22%2C%22100x200%22%5D%2C%22au%22%3A%22tagid_1%22%2C%22ps%22%3A%5B%7B%22pn%22%3A%22pubmatic%22%2C%22bc%22%3A%22pubmatic%22%2C%22kgpv%22%3A%22%22%2C%22kgpsv%22%3A%22%22%2C%22psz%22%3A%220x0%22%2C%22af%22%3A%22%22%2C%22eg%22%3A0%2C%22en%22%3A0%2C%22l1%22%3A0%2C%22l2%22%3A0%2C%22t%22%3A0%2C%22wb%22%3A0%2C%22bidid%22%3A%22bid-id-1%22%2C%22origbidid%22%3A%22bid-id-1%22%2C%22di%22%3A%22-1%22%2C%22dc%22%3A%22%22%2C%22db%22%3A1%2C%22ss%22%3A1%2C%22mi%22%3A0%2C%22ocpm%22%3A0%2C%22ocry%22%3A%22USD%22%7D%5D%2C%22rwrd%22%3A1%7D%5D%2C%22dvc%22%3A%7B%7D%2C%22ft%22%3A0%2C%22it%22%3A%22sdk%22%2C%22geo%22%3A%7B%7D%7D&pubid=5890"}`),
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
			name: "debug is not enabled and request is other than AppLovinMax",
			args: args{
				bidResponse: &openrtb2.BidResponse{
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
				ao: analytics.AuctionObject{
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
																		PubID:        5890,
																		Debug:        false,
																		Endpoint:     models.EndpintInappVideo,
																		WakandaDebug: &wakanda.Debug{},
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
									},
								},
							},
						},
						Ext: json.RawMessage(`{"matchedimpression":{"appnexus":50,"pubmatic":50}}`),
					},
				},
			},
			want: &openrtb2.BidResponse{
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
			name: "debug is not enabled and endpoint is AppLovinMax but AppLovinMax reject is false",
			args: args{
				bidResponse: &openrtb2.BidResponse{
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
				ao: analytics.AuctionObject{
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
																		PubID:    5890,
																		Debug:    false,
																		Endpoint: models.EndpointAppLovinMax,
																		AppLovinMax: models.AppLovinMax{
																			Reject: false,
																		},
																		WakandaDebug: &wakanda.Debug{},
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
									},
								},
							},
						},
						Ext: json.RawMessage(`{"matchedimpression":{"appnexus":50,"pubmatic":50}}`),
					},
				},
			},
			want: &openrtb2.BidResponse{
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
				Ext: nil,
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
			name: "debug is not enabled and endpoint is AppLovinMax but AppLovinMax reject is true",
			args: args{
				w: httptest.NewRecorder(),
				bidResponse: &openrtb2.BidResponse{
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
				ao: analytics.AuctionObject{
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
																		PubID:    5890,
																		Debug:    false,
																		Endpoint: models.EndpointAppLovinMax,
																		AppLovinMax: models.AppLovinMax{
																			Reject: true,
																		},
																		WakandaDebug: &wakanda.Debug{},
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
									},
								},
							},
						},
						Ext: json.RawMessage(`{"matchedimpression":{"appnexus":50,"pubmatic":50}}`),
					},
				},
			},
			rejectResponse: true,
			want: &openrtb2.BidResponse{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateResponseExtOW(tt.args.w, tt.args.bidResponse, tt.args.ao)
			assert.Equal(t, tt.want, tt.args.bidResponse, tt.name)
			assert.Equal(t, tt.RestoredResponse, tt.args.ao.Response)
			if tt.rejectResponse {
				assert.Equal(t, http.StatusNoContent, tt.args.w.(*httptest.ResponseRecorder).Code, tt.name)
			}
		})
	}
}

func TestGetGoogleSDKRejectedResponse(t *testing.T) {
	tests := []struct {
		name     string
		response *openrtb2.BidResponse
		ao       analytics.AuctionObject
		want     *openrtb2.BidResponse
	}{
		{
			name:     "nil response",
			response: nil,
			ao:       analytics.AuctionObject{},
			want:     nil,
		},
		{
			name: "nil request context",
			response: &openrtb2.BidResponse{
				ID:  "test-id",
				Ext: json.RawMessage(`{"existing":"value"}`),
			},
			ao: analytics.AuctionObject{},
			want: &openrtb2.BidResponse{
				ID:  "test-id",
				Ext: json.RawMessage(`{"existing":"value"}`),
			},
		},
		{
			name: "google-sdk reject false and NBR nil",
			response: &openrtb2.BidResponse{
				ID:  "test-id",
				Ext: json.RawMessage(`{"existing":"value"}`),
			},
			ao: analytics.AuctionObject{
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
																	Endpoint: models.EndpointGoogleSDK,
																	GoogleSDK: models.GoogleSDK{
																		Reject: false,
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
			want: &openrtb2.BidResponse{
				ID:  "test-id",
				Ext: json.RawMessage(`{"existing":"value"}`),
			},
		},
		{
			name: "google-sdk reject true with debug",
			response: &openrtb2.BidResponse{
				ID:  "test-id",
				NBR: openrtb3.NoBidUnknownError.Ptr(),
				Ext: json.RawMessage(`{"existing":"value"}`),
			},
			ao: analytics.AuctionObject{
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
																	Endpoint: models.EndpointGoogleSDK,
																	GoogleSDK: models.GoogleSDK{
																		Reject: true,
																	},
																	Debug: true,
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
			want: &openrtb2.BidResponse{
				ID:  "test-id",
				NBR: openrtb3.NoBidUnknownError.Ptr(),
				Ext: json.RawMessage(`{"existing":"value", "processing_time_ms": 1}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getGoogleSDKRejectedResponse(tt.response, tt.ao)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.NBR, got.NBR)
			assert.NotNil(t, got.Ext)
		})
	}
}

func TestRemoveDefaultBidsFromSeatNonBid(t *testing.T) {
	t.Cleanup(func() { glog.Flush() })

	tests := []struct {
		name             string
		inputSB          *openrtb_ext.SeatNonBidBuilder
		wantSB           openrtb_ext.SeatNonBidBuilder
		wantAOSeatNonBid []openrtb_ext.SeatNonBid
	}{
		{
			name:             "Empty_SeatNonBid",
			inputSB:          &openrtb_ext.SeatNonBidBuilder{},
			wantSB:           openrtb_ext.SeatNonBidBuilder{},
			wantAOSeatNonBid: nil,
		},
		{
			name: "SeatNonBid_with_default_bid",
			inputSB: &openrtb_ext.SeatNonBidBuilder{
				"seat1": {
					{ImpId: "imp1", StatusCode: 0},
				},
			},
			wantSB: openrtb_ext.SeatNonBidBuilder{
				"seat1": {
					{ImpId: "imp1", StatusCode: 0},
				},
			},
			wantAOSeatNonBid: []openrtb_ext.SeatNonBid{
				{Seat: "seat1", NonBid: []openrtb_ext.NonBid{
					{ImpId: "imp1", StatusCode: 0},
				}},
			},
		},
		{
			name: "Default_bid_and_bid_rejected_due_to_floors_for_impA",
			inputSB: &openrtb_ext.SeatNonBidBuilder{
				"seat1": {
					{ImpId: "impA", StatusCode: 0},
					{ImpId: "impA", StatusCode: 301},
				},
			},
			wantSB: openrtb_ext.SeatNonBidBuilder{
				"seat1": {
					{ImpId: "impA", StatusCode: 301},
				},
			},
			wantAOSeatNonBid: []openrtb_ext.SeatNonBid{
				{Seat: "seat1", NonBid: []openrtb_ext.NonBid{
					{ImpId: "impA", StatusCode: 301},
				}},
			},
		},
		{
			name: "Default_bid_and_bid_rejected_due_to_floors_for_multi_imp",
			inputSB: &openrtb_ext.SeatNonBidBuilder{
				"seat1": {
					{ImpId: "impA", StatusCode: 0},
					{ImpId: "impA", StatusCode: 301},
					{ImpId: "impB", StatusCode: 0},
				},
			},
			wantSB: openrtb_ext.SeatNonBidBuilder{
				"seat1": {
					{ImpId: "impA", StatusCode: 301},
					{ImpId: "impB", StatusCode: 0},
				},
			},
			wantAOSeatNonBid: []openrtb_ext.SeatNonBid{
				{Seat: "seat1", NonBid: []openrtb_ext.NonBid{
					{ImpId: "impA", StatusCode: 301},
					{ImpId: "impB", StatusCode: 0},
				}},
			},
		},
		{
			name: "Multiple_seats_with_all_bids_rejected_due_to_floors",
			inputSB: &openrtb_ext.SeatNonBidBuilder{
				"seat1": {
					{ImpId: "impA", StatusCode: 301},
				},
				"seat2": {
					{ImpId: "impA", StatusCode: 301},
				},
			},
			wantSB: openrtb_ext.SeatNonBidBuilder{
				"seat1": {
					{ImpId: "impA", StatusCode: 301},
				},
				"seat2": {
					{ImpId: "impA", StatusCode: 301},
				},
			},
			wantAOSeatNonBid: []openrtb_ext.SeatNonBid{
				{Seat: "seat1", NonBid: []openrtb_ext.NonBid{
					{ImpId: "impA", StatusCode: 301},
				}},
				{Seat: "seat2", NonBid: []openrtb_ext.NonBid{
					{ImpId: "impA", StatusCode: 301},
				}},
			},
		},
		{
			name: "Multiple_seats_independent",
			inputSB: &openrtb_ext.SeatNonBidBuilder{
				"seat1": {
					{ImpId: "impA", StatusCode: 0},
					{ImpId: "impB", StatusCode: 301},
				},
				"seat2": {
					{ImpId: "impA", StatusCode: 302},
				},
			},
			wantSB: openrtb_ext.SeatNonBidBuilder{
				"seat1": {
					{ImpId: "impA", StatusCode: 0},
					{ImpId: "impB", StatusCode: 301},
				},
				"seat2": {
					{ImpId: "impA", StatusCode: 302},
				},
			},
			wantAOSeatNonBid: []openrtb_ext.SeatNonBid{
				{Seat: "seat1", NonBid: []openrtb_ext.NonBid{
					{ImpId: "impA", StatusCode: 0},
					{ImpId: "impB", StatusCode: 301},
				}},
				{Seat: "seat2", NonBid: []openrtb_ext.NonBid{
					{ImpId: "impA", StatusCode: 302},
				}},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ao := &analytics.AuctionObject{}
			removeDefaultBidsFromSeatNonBid(tc.inputSB, ao)

			normalizeSeatNonBidBuilder(tc.inputSB)
			normalizeSeatNonBidBuilder(&tc.wantSB)
			assert.Equal(t, &tc.wantSB, tc.inputSB, "SeatNonBidBuilder mismatch")
			assert.ElementsMatch(t, tc.wantAOSeatNonBid, ao.SeatNonBid, "AuctionObject.SeatNonBid mismatch")
		})
	}
}

func normalizeSeatNonBidBuilder(sb *openrtb_ext.SeatNonBidBuilder) {
	if sb == nil {
		return
	}
	for _, nonBids := range *sb {
		sort.Slice(nonBids, func(i, j int) bool {
			return nonBids[i].ImpId < nonBids[j].ImpId
		})
	}
}
