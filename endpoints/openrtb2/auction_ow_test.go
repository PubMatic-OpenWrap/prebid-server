package openrtb2

import (
	"encoding/json"
	"fmt"

	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	analyticsConf "github.com/prebid/prebid-server/analytics/config"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/hooks"
	"github.com/prebid/prebid-server/metrics"
	metricsConfig "github.com/prebid/prebid-server/metrics/config"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/stored_requests/backends/empty_fetcher"
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
		analyticsConf.NewPBSAnalytics(&config.Analytics{}),
		map[string]string{"disabledbidder": "The bidder 'disabledbidder' has been disabled."},
		false,
		[]byte{},
		openrtb_ext.BuildBidderMap(),
		nil,
		nil,
		hardcodedResponseIPValidator{response: true},
		empty_fetcher.EmptyFetcher{},
		hooks.EmptyPlanBuilder{},
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
		pubid   string
		rejBids []analytics.RejectedBid
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
				rejBids: []analytics.RejectedBid{},
			},
			want: want{
				expectedCalls: 0,
			},
		},
		{
			description: "rejected bids",
			args: args{
				pubid: "1010",
				rejBids: []analytics.RejectedBid{
					{
						Seat:            "pubmatic",
						RejectionReason: openrtb3.LossBidAdvertiserExclusions,
					},
					{
						Seat:            "pubmatic",
						RejectionReason: openrtb3.LossBidBelowDealFloor,
					},
					{
						Seat:            "pubmatic",
						RejectionReason: openrtb3.LossBidAdvertiserExclusions,
					},
					{
						Seat:            "appnexus",
						RejectionReason: openrtb3.LossBidBelowDealFloor,
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

		recordRejectedBids(test.args.pubid, test.args.rejBids, me)
		me.AssertNumberOfCalls(t, "RecordRejectedBids", test.want.expectedCalls)
	}
}
