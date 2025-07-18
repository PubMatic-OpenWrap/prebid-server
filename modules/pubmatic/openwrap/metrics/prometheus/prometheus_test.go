package prometheus

import (
	"strconv"
	"testing"
	"time"

	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func createMetricsForTesting() *Metrics {
	return NewMetrics(&config.PrometheusMetrics{
		Port:      8080,
		Namespace: "prebid",
		Subsystem: "server",
	}, prometheus.NewRegistry())
}

func TestRecordOpenWrapServerPanicStats(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordOpenWrapServerPanicStats("node:pod", "process")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "panics", m.panics,
		expectedCount,
		prometheus.Labels{
			hostLabel:   "node:pod",
			methodLabel: "process",
		})
}

func TestRecordPartnerResponseErrors(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordPartnerResponseErrors("5890", "pubmatic", "timeout")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "partner_response_error", m.pubPartnerRespErrors,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:   "5890",
			partnerLabel: "pubmatic",
			errorLabel:   "timeout",
		})
}

func TestRecordPublisherPartnerNoCookieStats(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordPublisherPartnerNoCookieStats("5890", "pubmatic")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "no_cookie", m.pubPartnerNoCookie,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:   "5890",
			partnerLabel: "pubmatic",
		})
}

func TestRecordPartnerConfigErrors(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordPartnerConfigErrors("5890", "1234", "pubmatic", models.PartnerErrSlotNotMapped)

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "partner_config_errors", m.pubPartnerConfigErrors,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:     "5890",
			partnerLabel:   "pubmatic",
			profileIDLabel: "1234",
			errorLabel:     strconv.Itoa(models.PartnerErrSlotNotMapped),
		})
}

func TestRecordPublisherProfileRequests(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordPublisherProfileRequests("5890", "1234")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "pub_profile_requests", m.pubProfRequests,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:     "5890",
			profileIDLabel: "1234",
		})
}

func TestRecordNobidErrPrebidServerRequests(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordNobidErrPrebidServerRequests("5890", int(nbr.AllPartnerThrottled))

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "request_validation_errors", m.pubRequestValidationErrors,
		expectedCount,
		prometheus.Labels{
			pubIDLabel: "5890",
			nbrLabel:   strconv.Itoa(int(nbr.AllPartnerThrottled)),
		})
}

func TestRecordNobidErrPrebidServerResponse(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordNobidErrPrebidServerResponse("5890")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "no_bid", m.pubNoBidResponseErrors,
		expectedCount,
		prometheus.Labels{
			pubIDLabel: "5890",
		})
}

func TestRecordPlatformPublisherPartnerReqStats(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordPlatformPublisherPartnerReqStats(models.PLATFORM_APP, "5890", "pubmatic")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "platform_requests", m.pubPartnerPlatformRequests,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:    "5890",
			platformLabel: models.PLATFORM_APP,
			partnerLabel:  "pubmatic",
		})
}

func TestRecordPlatformPublisherPartnerResponseStats(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordPlatformPublisherPartnerResponseStats(models.PLATFORM_APP, "5890", "pubmatic")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "platform_responses", m.pubPartnerPlatformResponses,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:    "5890",
			platformLabel: models.PLATFORM_APP,
			partnerLabel:  "pubmatic",
		})
}

func TestRecordPublisherInvalidProfileRequests(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordPublisherInvalidProfileRequests(models.EndpointV25, "5890", "1234")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "invalid_requests", m.pubProfEndpointInvalidRequests,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:     "5890",
			endpointLabel:  models.EndpointV25,
			profileIDLabel: "1234",
		})
}

func TestRecordBadRequests(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordBadRequests(models.EndpointV25, "5890", int(nbr.AllPartnerThrottled))

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "bad_requests", m.endpointBadRequest,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:    "5890",
			endpointLabel: models.EndpointV25,
			nbrLabel:      strconv.Itoa(int(nbr.AllPartnerThrottled)),
		})
}

func TestRecordUidsCookieNotPresentErrorStats(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordUidsCookieNotPresentErrorStats("5890", "1234")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "uids_cookie_absent", m.pubProfUidsCookieAbsent,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:     "5890",
			profileIDLabel: "1234",
		})
}

func TestRecordVideoInstlImpsStats(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordVideoInstlImpsStats("5890", "1234")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "vid_instl_imps", m.pubProfVidInstlImps,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:     "5890",
			profileIDLabel: "1234",
		})
}

func TestRecordImpDisabledViaConfigStats(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordImpDisabledViaConfigStats(models.ImpTypeBanner, "5890", "1234")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "imps_disabled_via_config", m.pubProfImpDisabledViaConfig,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:     "5890",
			profileIDLabel: "1234",
			impFormatLabel: models.ImpTypeBanner,
		})
}

func TestRecordPublisherRequests(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordPublisherRequests(models.EndpointV25, "5890", models.PLATFORM_AMP)

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "endpoint_requests", m.pubPlatformEndpointRequests,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:    "5890",
			platformLabel: models.PLATFORM_AMP,
			endpointLabel: models.EndpointV25,
		})
}

func TestRecordReqImpsWithContentCount(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordReqImpsWithContentCount("5890", models.ContentTypeSite)

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "imps_with_content", m.pubImpsWithContent,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:  "5890",
			sourceLabel: models.ContentTypeSite,
		})
}

func TestRecordInjectTrackerErrorCount(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordInjectTrackerErrorCount(models.Banner, "5890", "pubmatic")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "inject_tracker_errors", m.pubPartnerInjectTrackerErrors,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:    "5890",
			adFormatLabel: models.Banner,
			partnerLabel:  "pubmatic",
		})
}

func TestRecordPartnerResponseTimeStats(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordPartnerResponseTimeStats("5890", "pubmatic", 3000)
	resultingHistogram := getHistogramFromHistogramVecByTwoKeys(m.pubPartnerResponseTimeSecs,
		pubIDLabel, "5890", partnerLabel, "pubmatic")

	assertHistogram(t, "partner_response_time", resultingHistogram, 1, 3)
}

func TestRecordPublisherResponseTimeStats(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordPublisherResponseTimeStats("5890", 3000)
	resultingHistogram := getHistogramFromHistogramVec(m.pubResponseTime,
		pubIDLabel, "5890")

	assertHistogram(t, "pub_response_time", resultingHistogram, 1, 3)
}

func TestRecordGetProfileDataTime(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordGetProfileDataTime(300 * time.Millisecond)
	resultingHistogram := getHistogramFromHistogram(m.getProfileData)

	assertHistogram(t, "sshb_profile_data_get_time", resultingHistogram, 1, 0.3)
}

func TestRecordDBQueryFailure(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordDBQueryFailure(models.AdunitConfigForLiveVersion, "5890", "59201")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "sshb_db_query_failed", m.dbQueryError,
		expectedCount,
		prometheus.Labels{
			queryTypeLabel: models.AdunitConfigForLiveVersion,
			pubIDLabel:     "5890",
			profileIDLabel: "59201",
		})
}

func TestRecordIBVRequest(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordIBVRequest("5890", "59201")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "ibv_requests", m.ibvRequests,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:     "5890",
			profileIDLabel: "59201",
		})
}

func TestRecordSignalDataStatus(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordSignalDataStatus("5890", "1234", models.MissingSignal)

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "signal_status", m.signalStatus,
		expectedCount,
		prometheus.Labels{
			pubIDLabel:      "5890",
			profileIDLabel:  "1234",
			signalTypeLabel: models.MissingSignal,
		})
}

func getHistogramFromHistogram(histogram prometheus.Histogram) dto.Histogram {
	var result dto.Histogram
	processMetrics(histogram, func(m dto.Metric) {
		result = *m.GetHistogram()
	})
	return result
}

func TestRecordAdruleEnabled(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordAdruleEnabled("5890", "123")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "pubProfAdruleEnabled", m.pubProfAdruleEnabled,
		expectedCount, prometheus.Labels{
			pubIdLabel:   "5890",
			profileLabel: "123",
		})
}

func TestRecordAdruleValidationFailure(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordAdruleValidationFailure("5890", "123")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "pubProfAdruleValidationfailure", m.pubProfAdruleValidationfailure,
		expectedCount, prometheus.Labels{
			pubIdLabel:   "5890",
			profileLabel: "123",
		})
}

func TestRecordBidRecoveryStatus(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordBidRecoveryStatus("5890", "123", true)

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "bid_recovery_response_status", m.pubBidRecoveryStatus,
		expectedCount, prometheus.Labels{
			pubIDLabel:     "5890",
			profileIDLabel: "123",
			successLabel:   "true",
		})
}

func TestRecordBidRecoveryResponseTime(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordBidRecoveryResponseTime("5890", "12345", time.Duration(70)*time.Millisecond)
	m.RecordBidRecoveryResponseTime("5890", "12345", time.Duration(130)*time.Millisecond)
	resultingHistogram := getHistogramFromHistogramVecByTwoKeys(m.pubBidRecoveryTime,
		pubIDLabel, "5890", profileIDLabel, "12345")
	assertHistogram(t, "bid_recovery_response_time", resultingHistogram, 2, 200)
}

func TestRecordGeoLookUpFailure(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordGeoLookupFailure("geo")

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "geo_lookup_fail", m.geoLookUpFailure,
		expectedCount, prometheus.Labels{
			endpointLabel: "geo",
		})
}

func TestRecordMBMFRequests(t *testing.T) {
	m := createMetricsForTesting()

	m.RecordMBMFRequests("endpoint", "5890", models.MBMFCountryDisabled)

	expectedCount := float64(1)
	assertCounterVecValue(t, "", "mbmf_requests", m.mbmfRequests,
		expectedCount, prometheus.Labels{
			endpointLabel:  "endpoint",
			pubIDLabel:     "5890",
			errorCodeLabel: strconv.Itoa(models.MBMFCountryDisabled),
		})
}

func getHistogramFromHistogramVec(histogram *prometheus.HistogramVec, labelKey, labelValue string) dto.Histogram {
	var result dto.Histogram
	processMetrics(histogram, func(m dto.Metric) {
		for _, label := range m.GetLabel() {
			if label.GetName() == labelKey && label.GetValue() == labelValue {
				result = *m.GetHistogram()
			}
		}
	})
	return result
}

func getHistogramFromHistogramVecByTwoKeys(histogram *prometheus.HistogramVec, label1Key, label1Value, label2Key, label2Value string) dto.Histogram {
	var result dto.Histogram
	processMetrics(histogram, func(m dto.Metric) {
		for ind, label := range m.GetLabel() {
			if label.GetName() == label1Key && label.GetValue() == label1Value {
				valInd := ind
				if ind == 1 {
					valInd = 0
				} else {
					valInd = 1
				}
				if m.Label[valInd].GetName() == label2Key && m.Label[valInd].GetValue() == label2Value {
					result = *m.GetHistogram()
				}
			}
		}
	})
	return result
}

func processMetrics(collector prometheus.Collector, handler func(m dto.Metric)) {
	collectorChan := make(chan prometheus.Metric)
	go func() {
		collector.Collect(collectorChan)
		close(collectorChan)
	}()

	for metric := range collectorChan {
		dtoMetric := dto.Metric{}
		metric.Write(&dtoMetric)
		handler(dtoMetric)
	}
}

func assertHistogram(t *testing.T, name string, histogram dto.Histogram, expectedCount uint64, expectedSum float64) {
	assert.Equal(t, expectedCount, histogram.GetSampleCount(), name+":count")
	assert.Equal(t, expectedSum, histogram.GetSampleSum(), name+":sum")
}

func assertCounterValue(t *testing.T, description string, counter prometheus.Counter, expected float64) {
	m := dto.Metric{}
	counter.Write(&m)
	actual := *m.GetCounter().Value

	assert.Equal(t, expected, actual, description)
}

func assertCounterVecValue(t *testing.T, description, name string, counterVec *prometheus.CounterVec, expected float64, labels prometheus.Labels) {
	counter := counterVec.With(labels)
	assertCounterValue(t, description, counter, expected)
}
