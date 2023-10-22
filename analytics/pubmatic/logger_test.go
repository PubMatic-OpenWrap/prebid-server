package pubmatic

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/exchange"
	"github.com/prebid/prebid-server/hooks/hookanalytics"
	"github.com/prebid/prebid-server/hooks/hookexecution"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestGetRequestCtx(t *testing.T) {
	tests := []struct {
		name                 string
		hookExecutionOutcome []hookexecution.StageOutcome
		rctx                 *models.RequestCtx
	}{
		{
			name: "rctx present",
			hookExecutionOutcome: []hookexecution.StageOutcome{
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
			rctx: &models.RequestCtx{},
		},
		{
			name: "rctx of invalid type",
			hookExecutionOutcome: []hookexecution.StageOutcome{
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
															"request-ctx": []string{},
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
			rctx: nil,
		},
		{
			name: "rctx absent",
			hookExecutionOutcome: []hookexecution.StageOutcome{
				{
					Groups: []hookexecution.GroupOutcome{
						{
							InvocationResults: []hookexecution.HookOutcome{},
						},
					},
				},
			},
			rctx: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rctx := GetRequestCtx(tt.hookExecutionOutcome)
			assert.Equal(t, tt.rctx, rctx, tt.name)
		})
	}
}

func TestConvertNonBidToBid(t *testing.T) {

	tests := []struct {
		name   string
		nonBid openrtb_ext.NonBid
		bid    openrtb2.Bid
	}{
		{
			name: "nonbid to bid",
			nonBid: openrtb_ext.NonBid{
				StatusCode: int(openrtb3.LossBidBelowAuctionFloor),
				ImpId:      "imp1",
				Ext: openrtb_ext.NonBidExt{
					Prebid: openrtb_ext.ExtResponseNonBidPrebid{
						Bid: openrtb_ext.NonBidObject{
							Price:             10,
							ADomain:           []string{"abc.com"},
							DealID:            "d1",
							OriginalBidCPM:    10,
							OriginalBidCur:    models.USD,
							OriginalBidCPMUSD: 0,
							W:                 10,
							H:                 50,
							DealPriority:      1,
							Video: &openrtb_ext.ExtBidPrebidVideo{
								Duration: 10,
							},
						},
					},
				},
			},
			bid: openrtb2.Bid{
				ImpID:   "imp1",
				Price:   10,
				ADomain: []string{"abc.com"},
				DealID:  "d1",
				W:       10,
				H:       50,
				Ext:     json.RawMessage(`{"prebid":{"dealpriority":1,"video":{"duration":10,"primary_category":"","vasttagid":""}},"origbidcpm":10,"origbidcur":"USD","nbr":301}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bid := convertNonBidToBid(tt.nonBid)
			fmt.Printf("%s", bid.Ext)
			assert.Equal(t, tt.bid, bid, tt.name)
		})
	}
}

func TestGetDefaultPartnerRecordsByImp(t *testing.T) {

	tests := []struct {
		name     string
		rCtx     *models.RequestCtx
		partners map[string][]PartnerRecord
	}{
		{
			name:     "empty ImpBidCtx",
			rCtx:     &models.RequestCtx{},
			partners: map[string][]PartnerRecord{},
		},
		{
			name: "multiple imps",
			rCtx: &models.RequestCtx{
				ImpBidCtx: map[string]models.ImpCtx{
					"imp1": {},
					"imp2": {},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					PartnerRecord{
						ServerSide:       1,
						DefaultBidStatus: 1,
						PartnerSize:      "0x0",
						DealID:           "-1",
					},
				},
				"imp2": {
					PartnerRecord{
						ServerSide:       1,
						DefaultBidStatus: 1,
						PartnerSize:      "0x0",
						DealID:           "-1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partners := getDefaultPartnerRecordsByImp(tt.rCtx)
			assert.Equal(t, len(tt.partners), len(partners), tt.name)
			for ind := range partners {
				// ignore order of elements in slice while comparing
				assert.ElementsMatch(t, partners[ind], tt.partners[ind], tt.name)
			}
		})
	}
}

func TestGetPartnerRecordsByImpForDroppedBids(t *testing.T) {
	type args struct {
		ao   analytics.AuctionObject
		rCtx *models.RequestCtx
	}
	tests := []struct {
		name     string
		args     args
		partners map[string][]PartnerRecord
	}{
		{
			name: "all bids got dropped",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{},
				},
				rCtx: &models.RequestCtx{
					DroppedBids: map[string][]openrtb2.Bid{
						"pubmatic": {
							{
								ID:    "bid-id-1",
								ImpID: "imp1",
							},
						},
						"appnexus": {
							{
								ID:    "bid-id-2",
								ImpID: "imp1",
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							Bidders: map[string]models.PartnerData{
								"pubmatic": {
									PartnerID:        1,
									PrebidBidderCode: "pubmatic",
								},
								"appnexus": {
									PartnerID:        2,
									PrebidBidderCode: "appnexus",
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:   "pubmatic",
						BidderCode:  "pubmatic",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: "USD",
					},
					{
						PartnerID:   "appnexus",
						BidderCode:  "appnexus",
						PartnerSize: "0x0",
						BidID:       "bid-id-2",
						OrigBidID:   "bid-id-2",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: "USD",
					},
				},
			},
		},
		{
			name: "1 bid got dropped, 1 bid is present in seatbid",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "pubmatic",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					DroppedBids: map[string][]openrtb2.Bid{
						"appnexus": {
							{
								ID:    "bid-id-2",
								ImpID: "imp1",
							},
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							Bidders: map[string]models.PartnerData{
								"pubmatic": {
									PartnerID:        1,
									PrebidBidderCode: "pubmatic",
								},
								"appnexus": {
									PartnerID:        2,
									PrebidBidderCode: "appnexus",
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:   "pubmatic",
						BidderCode:  "pubmatic",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: "USD",
					},
					{
						PartnerID:   "appnexus",
						BidderCode:  "appnexus",
						PartnerSize: "0x0",
						BidID:       "bid-id-2",
						OrigBidID:   "bid-id-2",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: "USD",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partners := getPartnerRecordsByImp(tt.args.ao, tt.args.rCtx)
			assert.Equal(t, len(tt.partners), len(partners), tt.name)
			for ind := range partners {
				// ignore order of elements in slice while comparing
				assert.ElementsMatch(t, partners[ind], tt.partners[ind], tt.name)
			}
		})
	}
}

func TestGetPartnerRecordsByImpForDefaultBids(t *testing.T) {
	type args struct {
		ao   analytics.AuctionObject
		rCtx *models.RequestCtx
	}
	tests := []struct {
		name     string
		args     args
		partners map[string][]PartnerRecord
	}{
		{
			name: "no default bid",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 10,
									},
								},
								Seat: "pubmatic",
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{},
									},
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:   "pubmatic",
						BidderCode:  "pubmatic",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: "USD",
						NetECPM:     10,
						GrossECPM:   10,
					},
				},
			},
		},
		{
			name: "default bid present but absent in seat-non-bid",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 10,
									},
								},
								Seat: "pubmatic",
							},
						},
					},
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: int(openrtb3.LossBidBelowAuctionFloor),
									Ext: openrtb_ext.NonBidExt{
										Prebid: openrtb_ext.ExtResponseNonBidPrebid{
											Bid: openrtb_ext.NonBidObject{
												ID: "bid-id-2",
											},
										},
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{},
									},
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:   "appnexus",
						BidderCode:  "appnexus",
						PartnerSize: "0x0",
						BidID:       "bid-id-2",
						OrigBidID:   "bid-id-2",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: "USD",
						Nbr:         openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
					},
					{
						PartnerID:   "pubmatic",
						BidderCode:  "pubmatic",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: "USD",
						NetECPM:     10,
						GrossECPM:   10,
					},
				},
			},
		},
		{
			name: "default bid present and same bid is available in seat-non-bid",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 0,
									},
								},
								Seat: "pubmatic",
							},
						},
					},
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "pubmatic",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: int(openrtb3.LossBidBelowAuctionFloor),
									Ext: openrtb_ext.NonBidExt{
										Prebid: openrtb_ext.ExtResponseNonBidPrebid{
											Bid: openrtb_ext.NonBidObject{
												ID: "bid-id-1",
											},
										},
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{},
									},
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:   "pubmatic",
						BidderCode:  "pubmatic",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: "USD",
						NetECPM:     0,
						GrossECPM:   0,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partners := getPartnerRecordsByImp(tt.args.ao, tt.args.rCtx)
			assert.Equal(t, len(tt.partners), len(partners), tt.name)
			for ind := range partners {
				// ignore order of elements in slice while comparing
				if !assert.ElementsMatch(t, partners[ind], tt.partners[ind], tt.name) {
					assert.Equal(t, partners[ind], tt.partners[ind], tt.name)
				}

			}
		})
	}
}

func TestGetPartnerRecordsByImpForSeatNonBid(t *testing.T) {
	type args struct {
		ao   analytics.AuctionObject
		rCtx *models.RequestCtx
	}
	tests := []struct {
		name     string
		args     args
		partners map[string][]PartnerRecord
	}{
		{
			name: "empty seatnonbids, expect empty partnerRecord",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{},
				},
				rCtx: &models.RequestCtx{},
			},
			partners: map[string][]PartnerRecord{},
		},
		{
			name: "logger should not log partner-throttled seat-non-bid",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{},
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "pubmatic",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: int(exchange.RequestBlockedPartnerThrottle),
									Ext: openrtb_ext.NonBidExt{
										Prebid: openrtb_ext.ExtResponseNonBidPrebid{
											Bid: openrtb_ext.NonBidObject{
												Price: 10,
											},
										},
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					AdapterThrottleMap: map[string]struct{}{
						"pubmatic": {},
					},
				},
			},
			partners: map[string][]PartnerRecord{},
		},
		{
			name: "logger should not log non-bid if ImpBidCtx dont have entry in ImpBidCtx",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{},
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "pubmatic",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: int(exchange.RequestBlockedPartnerThrottle),
									Ext: openrtb_ext.NonBidExt{
										Prebid: openrtb_ext.ExtResponseNonBidPrebid{
											Bid: openrtb_ext.NonBidObject{
												Price: 10,
											},
										},
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: make(map[string]models.ImpCtx),
				},
			},
			partners: map[string][]PartnerRecord{},
		},
		{
			name: "logger should not log non-bid for slot-not-mapped reason",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{},
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId:      "imp1",
									StatusCode: int(exchange.RequestBlockedPartnerThrottle),
									Ext: openrtb_ext.NonBidExt{
										Prebid: openrtb_ext.ExtResponseNonBidPrebid{
											Bid: openrtb_ext.NonBidObject{
												Price: 10,
											},
										},
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							NonMapped: map[string]struct{}{
								"appnexus": {},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{},
		},
		{
			name: "log rejected non-bid",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{},
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId: "imp1",
									Ext: openrtb_ext.NonBidExt{
										Prebid: openrtb_ext.ExtResponseNonBidPrebid{
											Bid: openrtb_ext.NonBidObject{
												Price:          10,
												ID:             "bid-id-1",
												W:              10,
												H:              50,
												OriginalBidCPM: 10,
											},
										},
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							Bidders: map[string]models.PartnerData{
								"appnexus": {
									PartnerID:        1,
									PrebidBidderCode: "appnexus",
									KGP:              "kgp",
									KGPV:             "kgpv",
								},
							},
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{},
										Nbr:    openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
									},
								},
							},
							BidFloor:    10.5,
							BidFloorCur: "USD",
						},
					},
					PartnerConfigMap: map[int]map[string]string{
						1: {
							"rev_share": "0",
						},
					},
					WinningBids: make(map[string]models.OwBid),
					Platform:    models.PLATFORM_APP,
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:      "appnexus",
						BidderCode:     "appnexus",
						KGPV:           "kgpv",
						KGPSV:          "kgpv",
						PartnerSize:    "10x50",
						GrossECPM:      10,
						NetECPM:        10,
						BidID:          "bid-id-1",
						OrigBidID:      "bid-id-1",
						DealID:         "-1",
						ServerSide:     1,
						OriginalCPM:    0,
						OriginalCur:    models.USD,
						FloorValue:     10.5,
						FloorRuleValue: 10.5,
						Nbr:            openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partners := getPartnerRecordsByImp(tt.args.ao, tt.args.rCtx)
			assert.Equal(t, tt.partners, partners, tt.name)
		})
	}
}

func TestGetPartnerRecordsByImpForSeatNonBidForFloors(t *testing.T) {
	type args struct {
		ao   analytics.AuctionObject
		rCtx *models.RequestCtx
	}
	tests := []struct {
		name     string
		args     args
		partners map[string][]PartnerRecord
	}{
		{
			name: "bid.ext.prebid.floors has high priority than imp.bidfloor",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{},
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId: "imp1",
									Ext: openrtb_ext.NonBidExt{
										Prebid: openrtb_ext.ExtResponseNonBidPrebid{
											Bid: openrtb_ext.NonBidObject{
												Price: 10,
												ID:    "bid-id-1",
											},
										},
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{
												Floors: &openrtb_ext.ExtBidPrebidFloors{
													FloorRule:      "*|*|ebay.com",
													FloorRuleValue: 1,
													FloorValue:     1,
													FloorCurrency:  models.USD,
												},
											},
										},
										Nbr: openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
									},
								},
							},
							BidFloor:    10.5,
							BidFloorCur: "USD",
						},
					},
					PartnerConfigMap: map[int]map[string]string{
						1: {},
					},
					WinningBids: make(map[string]models.OwBid),
					Platform:    models.PLATFORM_APP,
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:      "appnexus",
						BidderCode:     "appnexus",
						PartnerSize:    "0x0",
						BidID:          "bid-id-1",
						OrigBidID:      "bid-id-1",
						DealID:         "-1",
						GrossECPM:      10,
						NetECPM:        10,
						ServerSide:     1,
						OriginalCPM:    0,
						OriginalCur:    models.USD,
						FloorValue:     1,
						FloorRuleValue: 1,
						Nbr:            openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
					},
				},
			},
		},
		{
			name: "bid.ext.prebid.floors can have 0 value",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{},
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId: "imp1",
									Ext: openrtb_ext.NonBidExt{
										Prebid: openrtb_ext.ExtResponseNonBidPrebid{
											Bid: openrtb_ext.NonBidObject{
												Price: 10,
												ID:    "bid-id-1",
											},
										},
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{
												Floors: &openrtb_ext.ExtBidPrebidFloors{
													FloorRule:      "*|*|ebay.com",
													FloorRuleValue: 0,
													FloorValue:     0,
													FloorCurrency:  models.USD,
												},
											},
										},
										Nbr: openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
									},
								},
							},
							BidFloor:    10.5,
							BidFloorCur: "USD",
						},
					},
					PartnerConfigMap: map[int]map[string]string{
						1: {},
					},
					WinningBids: make(map[string]models.OwBid),
					Platform:    models.PLATFORM_APP,
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:      "appnexus",
						BidderCode:     "appnexus",
						PartnerSize:    "0x0",
						BidID:          "bid-id-1",
						OrigBidID:      "bid-id-1",
						DealID:         "-1",
						GrossECPM:      10,
						NetECPM:        10,
						ServerSide:     1,
						OriginalCPM:    0,
						OriginalCur:    models.USD,
						FloorValue:     0,
						FloorRuleValue: 0,
						Nbr:            openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
					},
				},
			},
		},
		{
			name: "bid.ext.prebid.floors not set, fallback to imp.bidfloor",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{},
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId: "imp1",
									Ext: openrtb_ext.NonBidExt{
										Prebid: openrtb_ext.ExtResponseNonBidPrebid{
											Bid: openrtb_ext.NonBidObject{
												Price: 10,
												ID:    "bid-id-1",
											},
										},
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{},
										},
										Nbr: openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
									},
								},
							},
							BidFloor:    10.567,
							BidFloorCur: "USD",
						},
					},
					PartnerConfigMap: map[int]map[string]string{
						1: {},
					},
					WinningBids: make(map[string]models.OwBid),
					Platform:    models.PLATFORM_APP,
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:      "appnexus",
						BidderCode:     "appnexus",
						PartnerSize:    "0x0",
						BidID:          "bid-id-1",
						OrigBidID:      "bid-id-1",
						DealID:         "-1",
						GrossECPM:      10,
						NetECPM:        10,
						ServerSide:     1,
						OriginalCPM:    0,
						OriginalCur:    models.USD,
						FloorValue:     10.57,
						FloorRuleValue: 10.57,
						Nbr:            openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
					},
				},
			},
		},
		{
			name: "currency conversion when floor value is set to imp.bidfloor",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{},
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId: "imp1",
									Ext: openrtb_ext.NonBidExt{
										Prebid: openrtb_ext.ExtResponseNonBidPrebid{
											Bid: openrtb_ext.NonBidObject{
												Price: 10,
												ID:    "bid-id-1",
											},
										},
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					CurrencyConversion: func(from, to string, value float64) (float64, error) {
						return 1000, nil
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{},
										},
										Nbr: openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
									},
								},
							},
							BidFloor:    10.567,
							BidFloorCur: "JPY",
						},
					},
					PartnerConfigMap: map[int]map[string]string{
						1: {},
					},
					WinningBids: make(map[string]models.OwBid),
					Platform:    models.PLATFORM_APP,
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:      "appnexus",
						BidderCode:     "appnexus",
						PartnerSize:    "0x0",
						BidID:          "bid-id-1",
						OrigBidID:      "bid-id-1",
						DealID:         "-1",
						GrossECPM:      10,
						NetECPM:        10,
						ServerSide:     1,
						OriginalCPM:    0,
						OriginalCur:    models.USD,
						FloorValue:     1000,
						FloorRuleValue: 1000,
						Nbr:            openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
					},
				},
			},
		},
		{
			name: "currency conversion when floor value is set from bid.ext.prebid.floors",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{},
					SeatNonBid: []openrtb_ext.SeatNonBid{
						{
							Seat: "appnexus",
							NonBid: []openrtb_ext.NonBid{
								{
									ImpId: "imp1",
									Ext: openrtb_ext.NonBidExt{
										Prebid: openrtb_ext.ExtResponseNonBidPrebid{
											Bid: openrtb_ext.NonBidObject{
												Price: 10,
												ID:    "bid-id-1",
											},
										},
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					CurrencyConversion: func(from, to string, value float64) (float64, error) {
						return 0.12, nil
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{
												Floors: &openrtb_ext.ExtBidPrebidFloors{
													FloorRule:      "*|*|ebay.com",
													FloorRuleValue: 1,
													FloorValue:     1,
													FloorCurrency:  "JPY",
												},
											},
										},
										Nbr: openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
									},
								},
							},
							BidFloor:    10.567,
							BidFloorCur: "JPY",
						},
					},
					PartnerConfigMap: map[int]map[string]string{
						1: {},
					},
					WinningBids: make(map[string]models.OwBid),
					Platform:    models.PLATFORM_APP,
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:      "appnexus",
						BidderCode:     "appnexus",
						PartnerSize:    "0x0",
						BidID:          "bid-id-1",
						OrigBidID:      "bid-id-1",
						DealID:         "-1",
						GrossECPM:      10,
						NetECPM:        10,
						ServerSide:     1,
						OriginalCPM:    0,
						OriginalCur:    models.USD,
						FloorValue:     0.12,
						FloorRuleValue: 0.12,
						Nbr:            openwrap.GetNonBidStatusCodePtr(openrtb3.LossBidBelowAuctionFloor),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partners := getPartnerRecordsByImp(tt.args.ao, tt.args.rCtx)
			assert.Equal(t, tt.partners, partners, tt.name)
		})
	}
}

func TestGetPartnerRecordsByImpForReserveredBidders(t *testing.T) {
	type args struct {
		ao   analytics.AuctionObject
		rCtx *models.RequestCtx
	}
	tests := []struct {
		name     string
		args     args
		partners map[string][]PartnerRecord
	}{
		{
			name: "ignore prebid_ctv bidder",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "prebid_ctv",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{},
				},
			},
			partners: map[string][]PartnerRecord{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partners := getPartnerRecordsByImp(tt.args.ao, tt.args.rCtx)
			assert.Equal(t, tt.partners, partners, tt.name)
		})
	}
}

func TestGetPartnerRecordsByImpForPostTimeoutBidStatus(t *testing.T) {
	type args struct {
		ao   analytics.AuctionObject
		rCtx *models.RequestCtx
	}
	tests := []struct {
		name     string
		args     args
		partners map[string][]PartnerRecord
	}{
		{
			name: "update 't' when Partner Timed out",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "appnexus",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										Nbr: openwrap.GetNonBidStatusCodePtr(openrtb3.NoBidTimeoutError),
									},
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:            "appnexus",
						BidderCode:           "appnexus",
						PartnerSize:          "0x0",
						BidID:                "bid-id-1",
						OrigBidID:            "bid-id-1",
						DealID:               "-1",
						ServerSide:           1,
						OriginalCur:          models.USD,
						PostTimeoutBidStatus: 1,
						Nbr:                  openwrap.GetNonBidStatusCodePtr(openrtb3.NoBidTimeoutError),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partners := getPartnerRecordsByImp(tt.args.ao, tt.args.rCtx)
			assert.Equal(t, tt.partners, partners, tt.name)
		})
	}
}

func TestGetPartnerRecordsByImpForBidExtPrebidObject(t *testing.T) {
	type args struct {
		ao   analytics.AuctionObject
		rCtx *models.RequestCtx
	}
	tests := []struct {
		name     string
		args     args
		partners map[string][]PartnerRecord
	}{
		{
			name: "log metadata object",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "appnexus",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{
												Meta: &openrtb_ext.ExtBidPrebidMeta{
													NetworkID: 100,
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
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:   "appnexus",
						BidderCode:  "appnexus",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: models.USD,
						MetaData: &MetaData{
							NetworkID: 100,
						},
					},
				},
			},
		},
		{
			name: "dealPriority is 1 but DealTierSatisfied is false",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "appnexus",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{
												DealTierSatisfied: false,
												DealPriority:      1,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:   "appnexus",
						BidderCode:  "appnexus",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: models.USD,
					},
				},
			},
		},
		{
			name: "dealPriority is 1 and DealTierSatisfied is true",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "appnexus",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{
												DealTierSatisfied: true,
												DealPriority:      1,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:    "appnexus",
						BidderCode:   "appnexus",
						PartnerSize:  "0x0",
						BidID:        "bid-id-1",
						OrigBidID:    "bid-id-1",
						DealID:       "-1",
						ServerSide:   1,
						OriginalCur:  models.USD,
						DealPriority: 1,
					},
				},
			},
		},
		{
			name: "dealPriority is 0 and DealTierSatisfied is true",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "appnexus",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{
												DealTierSatisfied: true,
												DealPriority:      0,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:    "appnexus",
						BidderCode:   "appnexus",
						PartnerSize:  "0x0",
						BidID:        "bid-id-1",
						OrigBidID:    "bid-id-1",
						DealID:       "-1",
						ServerSide:   1,
						OriginalCur:  models.USD,
						DealPriority: 0,
					},
				},
			},
		},
		{
			name: "bidExt.Prebid.Video.Duration is 0 ",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "appnexus",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{
												Video: &openrtb_ext.ExtBidPrebidVideo{
													Duration: 0,
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
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:   "appnexus",
						BidderCode:  "appnexus",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: models.USD,
					},
				},
			},
		},
		{
			name: "bidExt.Prebid.Video.Duration is valid, log AdDuration",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "appnexus",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{
												Video: &openrtb_ext.ExtBidPrebidVideo{
													Duration: 10,
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
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:   "appnexus",
						BidderCode:  "appnexus",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: models.USD,
						AdDuration:  ptrutil.ToPtr(10),
					},
				},
			},
		},
		{
			name: "override bidid by bidExt.Prebid.bidID",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "appnexus",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										ExtBid: openrtb_ext.ExtBid{
											Prebid: &openrtb_ext.ExtBidPrebid{
												BidId: "prebid-bid-id-1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						PartnerID:   "appnexus",
						BidderCode:  "appnexus",
						PartnerSize: "0x0",
						BidID:       "prebid-bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
						OriginalCur: models.USD,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partners := getPartnerRecordsByImp(tt.args.ao, tt.args.rCtx)
			assert.Equal(t, tt.partners, partners, tt.name)
		})
	}
}

func TestGetPartnerRecordsByImpForRevShareAndBidCPM(t *testing.T) {
	type args struct {
		ao   analytics.AuctionObject
		rCtx *models.RequestCtx
	}
	tests := []struct {
		name     string
		args     args
		partners map[string][]PartnerRecord
	}{
		{
			name: "origbidcpmusd not present",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "pubmatic",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 1.55,
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										OriginalBidCPM: 1.55,
										OriginalBidCur: "USD",
									},
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						NetECPM:     1.55,
						GrossECPM:   1.55,
						OriginalCPM: 1.55,
						OriginalCur: "USD",
						PartnerID:   "pubmatic",
						BidderCode:  "pubmatic",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
					},
				},
			},
		},
		{
			name: "origbidcpmusd not present and revshare present",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "pubmatic",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 100,
									},
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.REVSHARE: "10",
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										OriginalBidCPM: 100,
										OriginalBidCur: "USD",
									},
								},
							},
							Bidders: map[string]models.PartnerData{
								"pubmatic": {
									PartnerID:        1,
									PrebidBidderCode: "pubmatic",
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						NetECPM:     90,
						GrossECPM:   100,
						OriginalCPM: 100,
						OriginalCur: "USD",
						PartnerID:   "pubmatic",
						BidderCode:  "pubmatic",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
					},
				},
			},
		},
		{
			name: "origbidcpmusd is present",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "pubmatic",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 1.55,
									},
								},
							},
						},
						Cur: "INR",
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										OriginalBidCPM:    125.76829,
										OriginalBidCur:    "INR",
										OriginalBidCPMUSD: 1.76829,
									},
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						NetECPM:     1.77,
						GrossECPM:   1.77,
						OriginalCPM: 125.77,
						OriginalCur: "INR",
						PartnerID:   "pubmatic",
						BidderCode:  "pubmatic",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
					},
				},
			},
		},
		{
			name: "origbidcpmusd not present for non-USD bids",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "pubmatic",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 125.16829,
									},
								},
							},
						},
						Cur: "INR",
					},
				},
				rCtx: &models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										OriginalBidCPM: 125.16829,
										OriginalBidCur: "INR",
									},
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						GrossECPM:   125.17,
						NetECPM:     125.17,
						OriginalCPM: 125.17,
						OriginalCur: "INR",
						PartnerID:   "pubmatic",
						BidderCode:  "pubmatic",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
					},
				},
			},
		},
		{
			name: "origbidcpmusd is present, revshare is present",
			args: args{
				ao: analytics.AuctionObject{
					Response: &openrtb2.BidResponse{
						SeatBid: []openrtb2.SeatBid{
							{
								Seat: "pubmatic",
								Bid: []openrtb2.Bid{
									{
										ID:    "bid-id-1",
										ImpID: "imp1",
										Price: 100,
									},
								},
							},
						},
						Cur: "INR",
					},
				},
				rCtx: &models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						1: {
							models.REVSHARE: "10",
						},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BidCtx: map[string]models.BidCtx{
								"bid-id-1": {
									BidExt: models.BidExt{
										OriginalBidCPM:    200,
										OriginalBidCur:    "INR",
										OriginalBidCPMUSD: 100,
									},
								},
							},
							Bidders: map[string]models.PartnerData{
								"pubmatic": {
									PrebidBidderCode: "pubmatic",
									PartnerID:        1,
								},
							},
						},
					},
				},
			},
			partners: map[string][]PartnerRecord{
				"imp1": {
					{
						NetECPM:     90,
						GrossECPM:   100,
						OriginalCPM: 200,
						OriginalCur: "INR",
						PartnerID:   "pubmatic",
						BidderCode:  "pubmatic",
						PartnerSize: "0x0",
						BidID:       "bid-id-1",
						OrigBidID:   "bid-id-1",
						DealID:      "-1",
						ServerSide:  1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partners := getPartnerRecordsByImp(tt.args.ao, tt.args.rCtx)
			assert.Equal(t, tt.partners, partners, tt.name)
		})
	}
}

func TestGetPartnerRecordsByImpForMarketPlace(t *testing.T) {}
