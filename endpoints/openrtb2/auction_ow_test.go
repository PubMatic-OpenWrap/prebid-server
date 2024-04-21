package openrtb2

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/analytics"
	analyticsBuild "github.com/prebid/prebid-server/v2/analytics/build"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/errortypes"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/hooks"
	"github.com/prebid/prebid-server/v2/metrics"
	metricsConfig "github.com/prebid/prebid-server/v2/metrics/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/stored_requests/backends/empty_fetcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateImpExtOW(t *testing.T) {
	paramValidator, err := openrtb_ext.NewBidderParamsValidator("../../static/bidder-params")
	if err != nil {
		panic(err.Error())
	}

	type testCase struct {
		description    string
		impExt         json.RawMessage
		expectedImpExt string
		expectedErrs   []error
	}
	testGroups := []struct {
		description string
		testCases   []testCase
	}{
		{
			"Invalid bidder params tests",
			[]testCase{
				{
					description:    "Impression dropped for bidder with invalid bidder params",
					impExt:         json.RawMessage(`{"appnexus":{"placement_id":5.44}}`),
					expectedImpExt: `{"appnexus":{"placement_id":5.44}}`,
					expectedErrs: []error{&errortypes.BidderFailedSchemaValidation{Message: "request.imp[0].ext.prebid.bidder.appnexus failed validation.\nplacement_id: Invalid type. Expected: [integer,string], given: number"},
						fmt.Errorf("request.imp[%d].ext.prebid.bidder must contain at least one bidder", 0)},
				},
				{
					description:    "Valid Bidder params + Invalid bidder params",
					impExt:         json.RawMessage(`{"appnexus":{"placement_id":5.44},"pubmatic":{"publisherId":"156209"}}`),
					expectedImpExt: `{"appnexus":{"placement_id":5.44},"pubmatic":{"publisherId":"156209"}}`,
					expectedErrs:   []error{&errortypes.BidderFailedSchemaValidation{Message: "request.imp[0].ext.prebid.bidder.appnexus failed validation.\nplacement_id: Invalid type. Expected: [integer,string], given: number"}},
				},
				{
					description:    "Valid Bidder + Disabled Bidder + Invalid bidder params",
					impExt:         json.RawMessage(`{"pubmatic":{"publisherId":156209},"appnexus":{"placement_id":555},"disabledbidder":{"foo":"bar"}}`),
					expectedImpExt: `{"pubmatic":{"publisherId":156209},"appnexus":{"placement_id":555},"disabledbidder":{"foo":"bar"}}`,
					expectedErrs: []error{&errortypes.BidderTemporarilyDisabled{Message: "The bidder 'disabledbidder' has been disabled."},
						&errortypes.BidderFailedSchemaValidation{Message: "request.imp[0].ext.prebid.bidder.pubmatic failed validation.\npublisherId: Invalid type. Expected: string, given: integer"}},
				},
				{
					description:    "Valid Bidder + Disabled Bidder + Invalid bidder params",
					impExt:         json.RawMessage(`{"pubmatic":{"publisherId":156209},"disabledbidder":{"foo":"bar"}}`),
					expectedImpExt: `{"pubmatic":{"publisherId":156209},"disabledbidder":{"foo":"bar"}}`,
					expectedErrs: []error{&errortypes.BidderFailedSchemaValidation{Message: "request.imp[0].ext.prebid.bidder.pubmatic failed validation.\npublisherId: Invalid type. Expected: string, given: integer"},
						&errortypes.BidderTemporarilyDisabled{Message: "The bidder 'disabledbidder' has been disabled."},
						fmt.Errorf("request.imp[%d].ext.prebid.bidder must contain at least one bidder", 0)},
				},
			},
		},
	}
	deps := &endpointDeps{
		fakeUUIDGenerator{},
		&nobidExchange{},
		paramValidator,
		&mockStoredReqFetcher{},
		empty_fetcher.EmptyFetcher{},
		empty_fetcher.EmptyFetcher{},
		&config.Configuration{MaxRequestSize: int64(8096)},
		&metricsConfig.NilMetricsEngine{},
		analyticsBuild.New(&config.Analytics{}),
		map[string]string{"disabledbidder": "The bidder 'disabledbidder' has been disabled."},
		false,
		[]byte{},
		openrtb_ext.BuildBidderMap(),
		nil,
		nil,
		hardcodedResponseIPValidator{response: true},
		empty_fetcher.EmptyFetcher{},
		hooks.EmptyPlanBuilder{},
		&exchange.TmaxAdjustmentsPreprocessed{},
		openrtb_ext.NormalizeBidderName,
	}

	for _, group := range testGroups {
		for _, test := range group.testCases {
			impWrapper := &openrtb_ext.ImpWrapper{Imp: &openrtb2.Imp{Ext: test.impExt}}

			errs := deps.validateImpExt(impWrapper, nil, 0, false, nil)

			if len(test.expectedImpExt) > 0 {
				assert.JSONEq(t, test.expectedImpExt, string(impWrapper.Ext), "imp.ext JSON does not match expected. Test: %s. %s\n", group.description, test.description)
			} else {
				assert.Empty(t, impWrapper.Ext, "imp.ext expected to be empty but was: %s. Test: %s. %s\n", string(impWrapper.Ext), group.description, test.description)
			}
			assert.ElementsMatch(t, test.expectedErrs, errs, "errs slice does not match expected. Test: %s. %s\n", group.description, test.description)
		}
	}
}

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

func Test_upadteResponseExtForMax(t *testing.T) {
	type args struct {
		bidResponse *openrtb2.BidResponse
		rCtx        *models.RequestCtx
		ao          analytics.AuctionObject
	}
	tests := []struct {
		name            string
		args            args
		wantResponseExt json.RawMessage
	}{
		{
			name: "debug is disabled",
			args: args{
				bidResponse: &openrtb2.BidResponse{
					ID:  "123",
					Ext: nil,
				},
				rCtx: &models.RequestCtx{
					IsMaxRequest: true,
					Debug:        false,
				},
			},
		},
		{
			name: "empty bid response ext",
			args: args{
				bidResponse: &openrtb2.BidResponse{
					ID: "123",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "123",
									Ext: nil,
								},
							},
						},
					},
				},
				rCtx: &models.RequestCtx{
					IsMaxRequest: true,
					Debug:        true,
				},
			},
		},
		{
			name: "failed to unmarshal SeatBid[0].Bid[0].Ext",
			args: args{
				rCtx: &models.RequestCtx{
					IsMaxRequest: true,
					Debug:        true,
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "123",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "123",
									Ext: json.RawMessage(`{"signaldata":"{\"ID\":\"123\"}`),
								},
							},
						},
					},
				},
			},
			wantResponseExt: json.RawMessage(`{"signaldata":"{\"ID\":\"123\"}`),
		},
		{
			name: "SeatBid[0].Bid[0].Ext do not contain signaldata",
			args: args{
				rCtx: &models.RequestCtx{
					IsMaxRequest: true,
					Debug:        true,
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "123",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "123",
									Ext: json.RawMessage(`{"matchedimpression":"{}"}`),
								},
							},
						},
					},
				},
			},
			wantResponseExt: json.RawMessage(`{"matchedimpression":"{}"}`),
		},
		{
			name: "failed to unmarshal SeatBid[0].Bid[0].Ext signaldata response",
			args: args{
				rCtx: &models.RequestCtx{
					IsMaxRequest: true,
					Debug:        true,
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "123",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "123",
									Ext: json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"456\",\"impid\":\"789\",\"price\":1,\"burl\":\"http://example.com\",\"ext\":{\"key\":\"value\"}]}],\"bidid\":\"456\",\"cur\":\"USD\"}"}`),
								},
							},
						},
					},
				},
			},
			wantResponseExt: json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"456\",\"impid\":\"789\",\"price\":1,\"burl\":\"http://example.com\",\"ext\":{\"key\":\"value\"}]}],\"bidid\":\"456\",\"cur\":\"USD\"}"}`),
		},
		{
			name: "failed to unmarshal SeatBid[0].Bid[0].Ext signaldata response.Ext",
			args: args{
				rCtx: &models.RequestCtx{
					IsMaxRequest: true,
					Debug:        true,
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "123",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "123",
									Ext: json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"456\",\"impid\":\"789\",\"price\":1,\"burl\":\"http:\/\/example.com\",\"ext\":{\"key\":\"value\"}}]}],\"bidid\":\"456\",\"cur\":\"USD\",\"ext\":{\"tmaxrequest\":\"time\"}}"}`),
								},
							},
						},
					},
				},
			},
			wantResponseExt: json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"456\",\"impid\":\"789\",\"price\":1,\"burl\":\"http:\/\/example.com\",\"ext\":{\"key\":\"value\"}}]}],\"bidid\":\"456\",\"cur\":\"USD\",\"ext\":{\"tmaxrequest\":\"time\"}}"}`),
		},
		{
			name: "Successfully updated the logger in response ext",
			args: args{
				rCtx: &models.RequestCtx{
					IsMaxRequest: true,
					Debug:        true,
				},
				bidResponse: &openrtb2.BidResponse{
					ID: "123",
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "123",
									Ext: json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"456\",\"impid\":\"789\",\"price\":1,\"burl\":\"http:\/\/example.com\",\"ext\":{\"key\":\"value\"}}]}],\"bidid\":\"456\",\"cur\":\"USD\",\"ext\":{\"tmaxrequest\":500}}"}`),
								},
							},
						},
					},
				},
			},
			wantResponseExt: json.RawMessage(`{"signaldata":"{\"id\":\"123\",\"seatbid\":[{\"bid\":[{\"id\":\"456\",\"impid\":\"789\",\"price\":1,\"burl\":\"http://example.com\",\"ext\":{\"key\":\"value\"}}]}],\"bidid\":\"456\",\"cur\":\"USD\",\"ext\":{\"tmaxrequest\":500}}"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upadteResponseExtForMax(tt.args.ao, tt.args.rCtx, tt.args.bidResponse)
			if tt.args.bidResponse != nil && len(tt.args.bidResponse.SeatBid) > 0 && len(tt.args.bidResponse.SeatBid[0].Bid) > 0 && tt.args.bidResponse.SeatBid[0].Bid[0].Ext != nil {
				assert.Equal(t, tt.wantResponseExt, tt.args.bidResponse.SeatBid[0].Bid[0].Ext)
			}
		})
	}
}
