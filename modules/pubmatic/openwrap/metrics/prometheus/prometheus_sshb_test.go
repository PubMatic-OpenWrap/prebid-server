package prometheus

import (
	"testing"

	"time"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestRequestMetric(t *testing.T) {
	m := createMetricsForTesting()

	requestType := ReqTypeORTB25Web
	requestStatus := RequestStatusOK

	m.RecordRequest(metrics.Labels{
		RType:         requestType,
		RequestStatus: requestStatus,
	})

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "requests", m.owRequests,
		expectedCount,
		prometheus.Labels{
			requestTypeLabel:   string(requestType),
			requestStatusLabel: string(requestStatus),
		})
}

func TestMetrics_RecordLurlSent(t *testing.T) {
	m := createMetricsForTesting()

	type args struct {
		labels metrics.LurlStatusLabels
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "LurSent success",
			args: args{
				labels: metrics.LurlStatusLabels{
					PublisherID: "123",
					Partner:     "pubmatic",
					Status:      "success",
				},
			},
		},
		{
			name: "LurSent fail",
			args: args{
				labels: metrics.LurlStatusLabels{
					PublisherID: "123",
					Partner:     "pubmatic",
					Status:      "fail",
				},
			},
		},
		{
			name: "LurSent drop",
			args: args{
				labels: metrics.LurlStatusLabels{
					PublisherID: "123",
					Partner:     "pubmatic",
					Status:      "drop",
				},
			},
		},
		{
			name: "LurSent channel_full",
			args: args{
				labels: metrics.LurlStatusLabels{
					PublisherID: "123",
					Partner:     "pubmatic",
					Status:      "channel_full",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.RecordLurlSent(tt.args.labels)
			assertCounterVecValue(t, "", "lurl_sent", m.lurlSent, float64(1), prometheus.Labels{
				pubIdLabel:   tt.args.labels.PublisherID,
				partnerLable: tt.args.labels.Partner,
				statusLabel:  tt.args.labels.Status,
			})
		})
	}
}

func TestMetrics_RecordLurlBatchSent(t *testing.T) {
	m := createMetricsForTesting()

	type args struct {
		labels metrics.LurlBatchStatusLabels
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "LurBatchSent success",
			args: args{
				labels: metrics.LurlBatchStatusLabels{
					Status: "success",
				},
			},
		},
		{
			name: "LurBatchSent fail",
			args: args{
				labels: metrics.LurlBatchStatusLabels{
					Status: "fail",
				},
			},
		},
		{
			name: "LurBatchSent drop",
			args: args{
				labels: metrics.LurlBatchStatusLabels{
					Status: "drop",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.RecordLurlBatchSent(tt.args.labels)
			assertCounterVecValue(t, "", "lurl_batch_sent", m.lurlBatchSent, float64(1), prometheus.Labels{
				statusLabel: tt.args.labels.Status,
			})
		})
	}
}

func TestMetrics_RecordCtvUaAccuracy(t *testing.T) {

	m := createMetricsForTesting()

	type args struct {
		pubId  string
		status string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Regex detect ctv user agent correctly",
			args: args{
				pubId:  "1020",
				status: "success",
			},
		},
		{
			name: "Regex detect ctv user agent incorrectly",
			args: args{
				pubId:  "1020",
				status: "failure",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m.RecordCtvUaAccuracy(tt.args.pubId, tt.args.status)
			assertCounterVecValue(t, "", "ctv user agent accuracy", m.ctvUaAccuracy, float64(1), prometheus.Labels{
				pubIdLabel:  tt.args.pubId,
				statusLabel: tt.args.status,
			})
		})
	}
}

func TestRecordBids(t *testing.T) {
	m := createMetricsForTesting()

	type args struct {
		pubid, profid, bidder, deal string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "call_record_bids",
			args: args{
				pubid:  "1010",
				profid: "11",
				bidder: "pubmatic",
				deal:   "pubdeal",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.RecordBids(tt.args.pubid, tt.args.profid, tt.args.bidder, tt.args.deal)
			assertCounterVecValue(t, "", "bids", m.bids, tt.want, prometheus.Labels{
				pubIDLabel:   tt.args.pubid,
				profileLabel: tt.args.profid,
				bidderLabel:  tt.args.bidder,
				dealLabel:    tt.args.deal,
			})
		})
	}
}

func TestRecordPrebidTimeoutRequests(t *testing.T) {
	m := createMetricsForTesting()

	type args struct {
		pubid, profid string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "record_request_prebid_timeout",
			args: args{
				pubid:  "1010",
				profid: "11",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.RecordPrebidTimeoutRequests(tt.args.pubid, tt.args.profid)
			assertCounterVecValue(t, "", "request_prebid_timeout", m.prebidTimeoutRequests, tt.want, prometheus.Labels{
				pubIDLabel:   tt.args.pubid,
				profileLabel: tt.args.profid,
			})
		})
	}
}

func TestRecordPartnerTimeoutRequests(t *testing.T) {
	m := createMetricsForTesting()

	type args struct {
		pubid, profid, bidder string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "record_request_prebid_timeout",
			args: args{
				pubid:  "1010",
				profid: "11",
				bidder: "pubmatic",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.RecordPartnerTimeoutRequests(tt.args.pubid, tt.args.profid, tt.args.bidder)
			assertCounterVecValue(t, "", "request_partner_timeout", m.partnerTimeoutRequest, tt.want, prometheus.Labels{
				pubIDLabel:   tt.args.pubid,
				profileLabel: tt.args.profid,
				bidderLabel:  tt.args.bidder,
			})
		})
	}
}

func TestRecordSendLoggerDataTime(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordSendLoggerDataTime(300 * time.Millisecond)
	resultingHistogram := getHistogramFromHistogram(m.sendLoggerData)

	assertHistogram(t, "sshb_logger_data_send_time", resultingHistogram, 1, 0.3)
}

func TestRecordRequestTime(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordRequestTime("v25", time.Millisecond*250)

	result := getHistogramFromHistogramVec(m.owRequestTime, apiTypeLabel, "v25")
	assertHistogram(t, "TestRecordRequestTime", result, 1, 0.25)
}

func TestRecordOWServerPanic(t *testing.T) {

	m := createMetricsForTesting()

	type args struct {
		endpoint   string
		methodName string
		nodeName   string
		podName    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Record Panic counts",
			args: args{
				endpoint:   "/test/endpoint",
				methodName: "TestMethodName",
				nodeName:   "sfo2hyp084.sfo2.pubmatic.com",
				podName:    "ssheaderbidding-0-0-38-pr-26-2-k8s-5679748b7b-tqh42",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m.RecordOWServerPanic(tt.args.endpoint, tt.args.methodName, tt.args.nodeName, tt.args.podName)
			assertCounterVecValue(t, "", "panic", m.panicCounts, float64(1), prometheus.Labels{
				endpointLabel:   tt.args.endpoint,
				methodNameLabel: tt.args.methodName,
				nodeNameLabel:   tt.args.nodeName,
				podNameLabel:    tt.args.podName,
			})
		})
	}
}

func TestRegisterLabelPermutations(t *testing.T) {
	testCases := []struct {
		description      string
		labelsWithValues map[string][]string
		expectedLabels   []prometheus.Labels
	}{
		{
			description:      "Empty set.",
			labelsWithValues: map[string][]string{},
			expectedLabels:   []prometheus.Labels{},
		},
		{
			description: "Set of 1 label and 1 value.",
			labelsWithValues: map[string][]string{
				"1": {"A"},
			},
			expectedLabels: []prometheus.Labels{
				{"1": "A"},
			},
		},
		{
			description: "Set of 1 label and 2 values.",
			labelsWithValues: map[string][]string{
				"1": {"A", "B"},
			},
			expectedLabels: []prometheus.Labels{
				{"1": "A"},
				{"1": "B"},
			},
		},
		{
			description: "Set of 2 labels and 2 values.",
			labelsWithValues: map[string][]string{
				"1": {"A", "B"},
				"2": {"C", "D"},
			},
			expectedLabels: []prometheus.Labels{
				{"1": "A", "2": "C"},
				{"1": "A", "2": "D"},
				{"1": "B", "2": "C"},
				{"1": "B", "2": "D"},
			},
		},
	}

	for _, test := range testCases {
		resultLabels := []prometheus.Labels{}
		registerLabelPermutations(test.labelsWithValues, func(label prometheus.Labels) {
			resultLabels = append(resultLabels, label)
		})

		assert.ElementsMatch(t, test.expectedLabels, resultLabels)
	}
}

func TestMetrics_RecordAmpVideoRequets(t *testing.T) {
	m := createMetricsForTesting()

	type args struct {
		pubid     string
		profileid string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Record Amp Video Requests",
			args: args{
				pubid:     "1010",
				profileid: "11",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.RecordAmpVideoRequests(tt.args.pubid, tt.args.profileid)
			assertCounterVecValue(t, "", "sshb_amp_video_requests", m.ampVideoRequests, float64(1), prometheus.Labels{
				pubIDLabel:     tt.args.pubid,
				profileIDLabel: tt.args.profileid,
			})
		})
	}
}

func TestMetrics_RecordAmpVideoResponses(t *testing.T) {
	m := createMetricsForTesting()

	type args struct {
		pubid     string
		profileid string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Record Amp Video Requests",
			args: args{
				pubid:     "1010",
				profileid: "11",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.RecordAmpVideoResponses(tt.args.pubid, tt.args.profileid)
			assertCounterVecValue(t, "", "sshb_amp_video_responses", m.ampVideoResponses, float64(1), prometheus.Labels{
				pubIDLabel:     tt.args.pubid,
				profileIDLabel: tt.args.profileid,
			})
		})
	}
}
