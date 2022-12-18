package openrtb2

import (
	"encoding/json"
	"fmt"
	"strconv"

	"testing"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/openrtb/v17/openrtb3"
	"github.com/prebid/prebid-server/analytics"
	analyticsConf "github.com/prebid/prebid-server/analytics/config"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/metrics"
	metricsConfig "github.com/prebid/prebid-server/metrics/config"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/stored_requests/backends/empty_fetcher"
	"github.com/stretchr/testify/assert"
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
					impExt:         json.RawMessage(`{"appnexus":{"placement_id":"A"}}`),
					expectedImpExt: `{"appnexus":{"placement_id":"A"}}`,
					expectedErrs: []error{&errortypes.BidderFailedSchemaValidation{Message: "request.imp[0].ext.appnexus failed validation.\nplacement_id: Invalid type. Expected: integer, given: string"},
						fmt.Errorf("request.imp[%d].ext must contain at least one bidder", 0)},
				},
				{
					description:    "Valid Bidder params + Invalid bidder params",
					impExt:         json.RawMessage(`{"appnexus":{"placement_id":"A"},"pubmatic":{"publisherId":"156209"}}`),
					expectedImpExt: `{"appnexus":{"placement_id":"A"},"pubmatic":{"publisherId":"156209"}}`,
					expectedErrs:   []error{&errortypes.BidderFailedSchemaValidation{Message: "request.imp[0].ext.appnexus failed validation.\nplacement_id: Invalid type. Expected: integer, given: string"}},
				},
				{
					description:    "Valid Bidder + Disabled Bidder + Invalid bidder params",
					impExt:         json.RawMessage(`{"pubmatic":{"publisherId":156209},"appnexus":{"placement_id":555},"disabledbidder":{"foo":"bar"}}`),
					expectedImpExt: `{"pubmatic":{"publisherId":156209},"appnexus":{"placement_id":555},"disabledbidder":{"foo":"bar"}}`,
					expectedErrs: []error{&errortypes.BidderTemporarilyDisabled{Message: "The bidder 'disabledbidder' has been disabled."},
						&errortypes.BidderFailedSchemaValidation{Message: "request.imp[0].ext.pubmatic failed validation.\npublisherId: Invalid type. Expected: string, given: integer"}},
				},
				{
					description:    "Valid Bidder + Disabled Bidder + Invalid bidder params",
					impExt:         json.RawMessage(`{"pubmatic":{"publisherId":156209},"disabledbidder":{"foo":"bar"}}`),
					expectedImpExt: `{"pubmatic":{"publisherId":156209},"disabledbidder":{"foo":"bar"}}`,
					expectedErrs: []error{&errortypes.BidderFailedSchemaValidation{Message: "request.imp[0].ext.pubmatic failed validation.\npublisherId: Invalid type. Expected: string, given: integer"},
						&errortypes.BidderTemporarilyDisabled{Message: "The bidder 'disabledbidder' has been disabled."},
						fmt.Errorf("request.imp[%d].ext must contain at least one bidder", 0)},
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

	me := metricsConfig.NewMetricsEngine(&config.Configuration{
		Metrics: config.Metrics{Prometheus: config.PrometheusMetrics{Port: 1}},
	}, nil, nil)

	type args struct {
		pubid   string
		rejBids []analytics.RejectedBid
		engine  metrics.MetricsEngine
	}

	type want struct {
		bidderLossCount   map[string]map[openrtb3.LossReason]float64
		expectToGetRecord bool
	}

	tests := []struct {
		description string
		arg         args
		want        want
	}{
		{
			description: "empty rejected bids",
			arg: args{
				rejBids: []analytics.RejectedBid{},
				engine:  me,
			},
			want: want{
				expectToGetRecord: false,
			},
		},
		{
			description: "rejected bids",
			arg: args{
				pubid: "1010",
				rejBids: []analytics.RejectedBid{
					analytics.RejectedBid{
						Seat:            "pubmatic",
						RejectionReason: openrtb3.LossAdvertiserExclusions,
					},
					analytics.RejectedBid{
						Seat:            "pubmatic",
						RejectionReason: openrtb3.LossBelowDealFloor,
					},
					analytics.RejectedBid{
						Seat:            "pubmatic",
						RejectionReason: openrtb3.LossAdvertiserExclusions,
					},
					analytics.RejectedBid{
						Seat:            "appnexus",
						RejectionReason: openrtb3.LossBelowDealFloor,
					},
				},
				engine: me,
			},
			want: want{
				bidderLossCount: map[string]map[openrtb3.LossReason]float64{
					"pubmatic": map[openrtb3.LossReason]float64{
						openrtb3.LossAdvertiserExclusions: 2,
						openrtb3.LossBelowDealFloor:       1,
					},
					"appnexus": map[openrtb3.LossReason]float64{
						openrtb3.LossBelowDealFloor: 1,
					},
				},
				expectToGetRecord: true,
			},
		},
	}

	for _, test := range tests {
		recordRejectedBids(test.arg.pubid, test.arg.rejBids, test.arg.engine)

		detailedEngine, _ := test.arg.engine.(*metricsConfig.DetailedMetricsEngine)
		metricFamilies, _ := detailedEngine.PrometheusMetrics.Gatherer.Gather()
		isRecorded := false

		for _, metricFamily := range metricFamilies {
			if metricFamily.GetName() == "rejected_bids" {
				for _, metric := range metricFamily.GetMetric() {
					counter := metric.GetCounter().Value
					current_bidder := ""
					current_code := 0

					// verify labels
					for _, label := range metric.GetLabel() {
						switch *label.Name {
						case "pubid":
							if *label.Value != test.arg.pubid {
								t.Errorf("Expected pubid=[%s], got- [%s] for test - [%s]", test.arg.pubid, *label.Value, test.description)
							}
						case "bidder":
							current_bidder = *label.Value
						case "code":
							current_code, _ = strconv.Atoi(*label.Value)
						default:
							t.Errorf("Unexpected label %s found in metric for test - [%s]", *label.Name, test.description)
						}
					}
					lossCount := test.want.bidderLossCount[current_bidder]

					// verify counter value
					if *counter != lossCount[openrtb3.LossReason(current_code)] {
						t.Errorf("Counter value mismatch for bidder- [%s], code - [%d], expected - [%f], got - [%f], for test - [%s]",
							current_bidder, current_code, lossCount[openrtb3.LossReason(current_code)], *counter, test.description)
					}
					isRecorded = true
				}
			}
		}
		// verify if metric got recorded by metric-engine.
		if test.want.expectToGetRecord != isRecorded {
			t.Errorf("Failed to record rejected_bids for test case - [%s]", test.description)
		}
	}
}
