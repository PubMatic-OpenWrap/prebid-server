// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/metrics (interfaces: MetricsEngine)

// Package mock_metrics is a generated GoMock package.
package mock_metrics

import (
	gomock "github.com/golang/mock/gomock"
	metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	reflect "reflect"
	time "time"
)

// MockMetricsEngine is a mock of MetricsEngine interface
type MockMetricsEngine struct {
	ctrl     *gomock.Controller
	recorder *MockMetricsEngineMockRecorder
}

// MockMetricsEngineMockRecorder is the mock recorder for MockMetricsEngine
type MockMetricsEngineMockRecorder struct {
	mock *MockMetricsEngine
}

// NewMockMetricsEngine creates a new mock instance
func NewMockMetricsEngine(ctrl *gomock.Controller) *MockMetricsEngine {
	mock := &MockMetricsEngine{ctrl: ctrl}
	mock.recorder = &MockMetricsEngineMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMetricsEngine) EXPECT() *MockMetricsEngineMockRecorder {
	return m.recorder
}

// RecordAdPodGeneratedImpressionsCount mocks base method
func (m *MockMetricsEngine) RecordAdPodGeneratedImpressionsCount(arg0 int, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordAdPodGeneratedImpressionsCount", arg0, arg1)
}

// RecordAdPodGeneratedImpressionsCount indicates an expected call of RecordAdPodGeneratedImpressionsCount
func (mr *MockMetricsEngineMockRecorder) RecordAdPodGeneratedImpressionsCount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordAdPodGeneratedImpressionsCount", reflect.TypeOf((*MockMetricsEngine)(nil).RecordAdPodGeneratedImpressionsCount), arg0, arg1)
}

// RecordAdPodImpressionYield mocks base method
func (m *MockMetricsEngine) RecordAdPodImpressionYield(arg0, arg1 int, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordAdPodImpressionYield", arg0, arg1, arg2)
}

// RecordAdPodImpressionYield indicates an expected call of RecordAdPodImpressionYield
func (mr *MockMetricsEngineMockRecorder) RecordAdPodImpressionYield(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordAdPodImpressionYield", reflect.TypeOf((*MockMetricsEngine)(nil).RecordAdPodImpressionYield), arg0, arg1, arg2)
}

// RecordBadRequests mocks base method
func (m *MockMetricsEngine) RecordBadRequests(arg0 string, arg1 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordBadRequests", arg0, arg1)
}

// RecordBadRequests indicates an expected call of RecordBadRequests
func (mr *MockMetricsEngineMockRecorder) RecordBadRequests(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordBadRequests", reflect.TypeOf((*MockMetricsEngine)(nil).RecordBadRequests), arg0, arg1)
}

// RecordBidResponseByDealCountInHB mocks base method
func (m *MockMetricsEngine) RecordBidResponseByDealCountInHB(arg0, arg1, arg2, arg3 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordBidResponseByDealCountInHB", arg0, arg1, arg2, arg3)
}

// RecordBidResponseByDealCountInHB indicates an expected call of RecordBidResponseByDealCountInHB
func (mr *MockMetricsEngineMockRecorder) RecordBidResponseByDealCountInHB(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordBidResponseByDealCountInHB", reflect.TypeOf((*MockMetricsEngine)(nil).RecordBidResponseByDealCountInHB), arg0, arg1, arg2, arg3)
}

// RecordBidResponseByDealCountInPBS mocks base method
func (m *MockMetricsEngine) RecordBidResponseByDealCountInPBS(arg0, arg1, arg2, arg3 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordBidResponseByDealCountInPBS", arg0, arg1, arg2, arg3)
}

// RecordBidResponseByDealCountInPBS indicates an expected call of RecordBidResponseByDealCountInPBS
func (mr *MockMetricsEngineMockRecorder) RecordBidResponseByDealCountInPBS(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordBidResponseByDealCountInPBS", reflect.TypeOf((*MockMetricsEngine)(nil).RecordBidResponseByDealCountInPBS), arg0, arg1, arg2, arg3)
}

// RecordBids mocks base method
func (m *MockMetricsEngine) RecordBids(arg0, arg1, arg2, arg3 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordBids", arg0, arg1, arg2, arg3)
}

// RecordBids indicates an expected call of RecordBids
func (mr *MockMetricsEngineMockRecorder) RecordBids(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordBids", reflect.TypeOf((*MockMetricsEngine)(nil).RecordBids), arg0, arg1, arg2, arg3)
}

// RecordCTVHTTPMethodRequests mocks base method
func (m *MockMetricsEngine) RecordCTVHTTPMethodRequests(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordCTVHTTPMethodRequests", arg0, arg1, arg2)
}

// RecordCTVHTTPMethodRequests indicates an expected call of RecordCTVHTTPMethodRequests
func (mr *MockMetricsEngineMockRecorder) RecordCTVHTTPMethodRequests(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordCTVHTTPMethodRequests", reflect.TypeOf((*MockMetricsEngine)(nil).RecordCTVHTTPMethodRequests), arg0, arg1, arg2)
}

// RecordCTVInvalidReasonCount mocks base method
func (m *MockMetricsEngine) RecordCTVInvalidReasonCount(arg0 int, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordCTVInvalidReasonCount", arg0, arg1)
}

// RecordCTVInvalidReasonCount indicates an expected call of RecordCTVInvalidReasonCount
func (mr *MockMetricsEngineMockRecorder) RecordCTVInvalidReasonCount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordCTVInvalidReasonCount", reflect.TypeOf((*MockMetricsEngine)(nil).RecordCTVInvalidReasonCount), arg0, arg1)
}

// RecordCTVReqCountWithAdPod mocks base method
func (m *MockMetricsEngine) RecordCTVReqCountWithAdPod(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordCTVReqCountWithAdPod", arg0, arg1)
}

// RecordCTVReqCountWithAdPod indicates an expected call of RecordCTVReqCountWithAdPod
func (mr *MockMetricsEngineMockRecorder) RecordCTVReqCountWithAdPod(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordCTVReqCountWithAdPod", reflect.TypeOf((*MockMetricsEngine)(nil).RecordCTVReqCountWithAdPod), arg0, arg1)
}

// RecordCTVReqImpsWithDbConfigCount mocks base method
func (m *MockMetricsEngine) RecordCTVReqImpsWithDbConfigCount(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordCTVReqImpsWithDbConfigCount", arg0)
}

// RecordCTVReqImpsWithDbConfigCount indicates an expected call of RecordCTVReqImpsWithDbConfigCount
func (mr *MockMetricsEngineMockRecorder) RecordCTVReqImpsWithDbConfigCount(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordCTVReqImpsWithDbConfigCount", reflect.TypeOf((*MockMetricsEngine)(nil).RecordCTVReqImpsWithDbConfigCount), arg0)
}

// RecordCTVReqImpsWithReqConfigCount mocks base method
func (m *MockMetricsEngine) RecordCTVReqImpsWithReqConfigCount(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordCTVReqImpsWithReqConfigCount", arg0)
}

// RecordCTVReqImpsWithReqConfigCount indicates an expected call of RecordCTVReqImpsWithReqConfigCount
func (mr *MockMetricsEngineMockRecorder) RecordCTVReqImpsWithReqConfigCount(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordCTVReqImpsWithReqConfigCount", reflect.TypeOf((*MockMetricsEngine)(nil).RecordCTVReqImpsWithReqConfigCount), arg0)
}

// RecordCTVRequests mocks base method
func (m *MockMetricsEngine) RecordCTVRequests(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordCTVRequests", arg0, arg1)
}

// RecordCTVRequests indicates an expected call of RecordCTVRequests
func (mr *MockMetricsEngineMockRecorder) RecordCTVRequests(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordCTVRequests", reflect.TypeOf((*MockMetricsEngine)(nil).RecordCTVRequests), arg0, arg1)
}

// RecordCacheErrorRequests mocks base method
func (m *MockMetricsEngine) RecordCacheErrorRequests(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordCacheErrorRequests", arg0, arg1, arg2)
}

// RecordCacheErrorRequests indicates an expected call of RecordCacheErrorRequests
func (mr *MockMetricsEngineMockRecorder) RecordCacheErrorRequests(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordCacheErrorRequests", reflect.TypeOf((*MockMetricsEngine)(nil).RecordCacheErrorRequests), arg0, arg1, arg2)
}

// RecordCtvUaAccuracy mocks base method
func (m *MockMetricsEngine) RecordCtvUaAccuracy(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordCtvUaAccuracy", arg0, arg1)
}

// RecordCtvUaAccuracy indicates an expected call of RecordCtvUaAccuracy
func (mr *MockMetricsEngineMockRecorder) RecordCtvUaAccuracy(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordCtvUaAccuracy", reflect.TypeOf((*MockMetricsEngine)(nil).RecordCtvUaAccuracy), arg0, arg1)
}

// RecordDBQueryFailure mocks base method
func (m *MockMetricsEngine) RecordDBQueryFailure(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordDBQueryFailure", arg0, arg1, arg2)
}

// RecordDBQueryFailure indicates an expected call of RecordDBQueryFailure
func (mr *MockMetricsEngineMockRecorder) RecordDBQueryFailure(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordDBQueryFailure", reflect.TypeOf((*MockMetricsEngine)(nil).RecordDBQueryFailure), arg0, arg1, arg2)
}

// RecordGeoDBInitStatus mocks base method
func (m *MockMetricsEngine) RecordGeoDBInitStatus(arg0, arg1, arg2 string, arg3 float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordGeoDBInitStatus", arg0, arg1, arg2, arg3)
}

// RecordGeoDBInitStatus indicates an expected call of RecordGeoDBInitStatus
func (mr *MockMetricsEngineMockRecorder) RecordGeoDBInitStatus(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordGeoDBInitStatus", reflect.TypeOf((*MockMetricsEngine)(nil).RecordGeoDBInitStatus), arg0, arg1, arg2, arg3)
}

// RecordGetProfileDataTime mocks base method
func (m *MockMetricsEngine) RecordGetProfileDataTime(arg0, arg1 string, arg2 time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordGetProfileDataTime", arg0, arg1, arg2)
}

// RecordGetProfileDataTime indicates an expected call of RecordGetProfileDataTime
func (mr *MockMetricsEngineMockRecorder) RecordGetProfileDataTime(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordGetProfileDataTime", reflect.TypeOf((*MockMetricsEngine)(nil).RecordGetProfileDataTime), arg0, arg1, arg2)
}

// RecordImpDisabledViaConfigStats mocks base method
func (m *MockMetricsEngine) RecordImpDisabledViaConfigStats(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordImpDisabledViaConfigStats", arg0, arg1, arg2)
}

// RecordImpDisabledViaConfigStats indicates an expected call of RecordImpDisabledViaConfigStats
func (mr *MockMetricsEngineMockRecorder) RecordImpDisabledViaConfigStats(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordImpDisabledViaConfigStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordImpDisabledViaConfigStats), arg0, arg1, arg2)
}

// RecordInjectTrackerErrorCount mocks base method
func (m *MockMetricsEngine) RecordInjectTrackerErrorCount(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordInjectTrackerErrorCount", arg0, arg1, arg2)
}

// RecordInjectTrackerErrorCount indicates an expected call of RecordInjectTrackerErrorCount
func (mr *MockMetricsEngineMockRecorder) RecordInjectTrackerErrorCount(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordInjectTrackerErrorCount", reflect.TypeOf((*MockMetricsEngine)(nil).RecordInjectTrackerErrorCount), arg0, arg1, arg2)
}

// RecordInvalidCreativeStats mocks base method
func (m *MockMetricsEngine) RecordInvalidCreativeStats(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordInvalidCreativeStats", arg0, arg1)
}

// RecordInvalidCreativeStats indicates an expected call of RecordInvalidCreativeStats
func (mr *MockMetricsEngineMockRecorder) RecordInvalidCreativeStats(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordInvalidCreativeStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordInvalidCreativeStats), arg0, arg1)
}

// RecordLurlBatchSent mocks base method
func (m *MockMetricsEngine) RecordLurlBatchSent(arg0 metrics.LurlBatchStatusLabels) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordLurlBatchSent", arg0)
}

// RecordLurlBatchSent indicates an expected call of RecordLurlBatchSent
func (mr *MockMetricsEngineMockRecorder) RecordLurlBatchSent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordLurlBatchSent", reflect.TypeOf((*MockMetricsEngine)(nil).RecordLurlBatchSent), arg0)
}

// RecordLurlSent mocks base method
func (m *MockMetricsEngine) RecordLurlSent(arg0 metrics.LurlStatusLabels) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordLurlSent", arg0)
}

// RecordLurlSent indicates an expected call of RecordLurlSent
func (mr *MockMetricsEngineMockRecorder) RecordLurlSent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordLurlSent", reflect.TypeOf((*MockMetricsEngine)(nil).RecordLurlSent), arg0)
}

// RecordNobidErrPrebidServerRequests mocks base method
func (m *MockMetricsEngine) RecordNobidErrPrebidServerRequests(arg0 string, arg1 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordNobidErrPrebidServerRequests", arg0, arg1)
}

// RecordNobidErrPrebidServerRequests indicates an expected call of RecordNobidErrPrebidServerRequests
func (mr *MockMetricsEngineMockRecorder) RecordNobidErrPrebidServerRequests(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordNobidErrPrebidServerRequests", reflect.TypeOf((*MockMetricsEngine)(nil).RecordNobidErrPrebidServerRequests), arg0, arg1)
}

// RecordNobidErrPrebidServerResponse mocks base method
func (m *MockMetricsEngine) RecordNobidErrPrebidServerResponse(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordNobidErrPrebidServerResponse", arg0)
}

// RecordNobidErrPrebidServerResponse indicates an expected call of RecordNobidErrPrebidServerResponse
func (mr *MockMetricsEngineMockRecorder) RecordNobidErrPrebidServerResponse(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordNobidErrPrebidServerResponse", reflect.TypeOf((*MockMetricsEngine)(nil).RecordNobidErrPrebidServerResponse), arg0)
}

// RecordOWServerPanic mocks base method
func (m *MockMetricsEngine) RecordOWServerPanic(arg0, arg1, arg2, arg3 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordOWServerPanic", arg0, arg1, arg2, arg3)
}

// RecordOWServerPanic indicates an expected call of RecordOWServerPanic
func (mr *MockMetricsEngineMockRecorder) RecordOWServerPanic(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordOWServerPanic", reflect.TypeOf((*MockMetricsEngine)(nil).RecordOWServerPanic), arg0, arg1, arg2, arg3)
}

// RecordOpenWrapServerPanicStats mocks base method
func (m *MockMetricsEngine) RecordOpenWrapServerPanicStats(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordOpenWrapServerPanicStats", arg0, arg1)
}

// RecordOpenWrapServerPanicStats indicates an expected call of RecordOpenWrapServerPanicStats
func (mr *MockMetricsEngineMockRecorder) RecordOpenWrapServerPanicStats(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordOpenWrapServerPanicStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordOpenWrapServerPanicStats), arg0, arg1)
}

// RecordPBSAuctionRequestsStats mocks base method
func (m *MockMetricsEngine) RecordPBSAuctionRequestsStats() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPBSAuctionRequestsStats")
}

// RecordPBSAuctionRequestsStats indicates an expected call of RecordPBSAuctionRequestsStats
func (mr *MockMetricsEngineMockRecorder) RecordPBSAuctionRequestsStats() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPBSAuctionRequestsStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPBSAuctionRequestsStats))
}

// RecordPartnerConfigErrors mocks base method
func (m *MockMetricsEngine) RecordPartnerConfigErrors(arg0, arg1, arg2 string, arg3 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPartnerConfigErrors", arg0, arg1, arg2, arg3)
}

// RecordPartnerConfigErrors indicates an expected call of RecordPartnerConfigErrors
func (mr *MockMetricsEngineMockRecorder) RecordPartnerConfigErrors(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPartnerConfigErrors", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPartnerConfigErrors), arg0, arg1, arg2, arg3)
}

// RecordPartnerResponseErrors mocks base method
func (m *MockMetricsEngine) RecordPartnerResponseErrors(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPartnerResponseErrors", arg0, arg1, arg2)
}

// RecordPartnerResponseErrors indicates an expected call of RecordPartnerResponseErrors
func (mr *MockMetricsEngineMockRecorder) RecordPartnerResponseErrors(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPartnerResponseErrors", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPartnerResponseErrors), arg0, arg1, arg2)
}

// RecordPartnerResponseTimeStats mocks base method
func (m *MockMetricsEngine) RecordPartnerResponseTimeStats(arg0, arg1 string, arg2 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPartnerResponseTimeStats", arg0, arg1, arg2)
}

// RecordPartnerResponseTimeStats indicates an expected call of RecordPartnerResponseTimeStats
func (mr *MockMetricsEngineMockRecorder) RecordPartnerResponseTimeStats(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPartnerResponseTimeStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPartnerResponseTimeStats), arg0, arg1, arg2)
}

// RecordPartnerTimeoutInPBS mocks base method
func (m *MockMetricsEngine) RecordPartnerTimeoutInPBS(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPartnerTimeoutInPBS", arg0, arg1, arg2)
}

// RecordPartnerTimeoutInPBS indicates an expected call of RecordPartnerTimeoutInPBS
func (mr *MockMetricsEngineMockRecorder) RecordPartnerTimeoutInPBS(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPartnerTimeoutInPBS", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPartnerTimeoutInPBS), arg0, arg1, arg2)
}

// RecordPartnerTimeoutRequests mocks base method
func (m *MockMetricsEngine) RecordPartnerTimeoutRequests(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPartnerTimeoutRequests", arg0, arg1, arg2)
}

// RecordPartnerTimeoutRequests indicates an expected call of RecordPartnerTimeoutRequests
func (mr *MockMetricsEngineMockRecorder) RecordPartnerTimeoutRequests(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPartnerTimeoutRequests", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPartnerTimeoutRequests), arg0, arg1, arg2)
}

// RecordPlatformPublisherPartnerReqStats mocks base method
func (m *MockMetricsEngine) RecordPlatformPublisherPartnerReqStats(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPlatformPublisherPartnerReqStats", arg0, arg1, arg2)
}

// RecordPlatformPublisherPartnerReqStats indicates an expected call of RecordPlatformPublisherPartnerReqStats
func (mr *MockMetricsEngineMockRecorder) RecordPlatformPublisherPartnerReqStats(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPlatformPublisherPartnerReqStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPlatformPublisherPartnerReqStats), arg0, arg1, arg2)
}

// RecordPlatformPublisherPartnerResponseStats mocks base method
func (m *MockMetricsEngine) RecordPlatformPublisherPartnerResponseStats(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPlatformPublisherPartnerResponseStats", arg0, arg1, arg2)
}

// RecordPlatformPublisherPartnerResponseStats indicates an expected call of RecordPlatformPublisherPartnerResponseStats
func (mr *MockMetricsEngineMockRecorder) RecordPlatformPublisherPartnerResponseStats(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPlatformPublisherPartnerResponseStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPlatformPublisherPartnerResponseStats), arg0, arg1, arg2)
}

// RecordPreProcessingTimeStats mocks base method
func (m *MockMetricsEngine) RecordPreProcessingTimeStats(arg0 string, arg1 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPreProcessingTimeStats", arg0, arg1)
}

// RecordPreProcessingTimeStats indicates an expected call of RecordPreProcessingTimeStats
func (mr *MockMetricsEngineMockRecorder) RecordPreProcessingTimeStats(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPreProcessingTimeStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPreProcessingTimeStats), arg0, arg1)
}

// RecordPrebidTimeoutRequests mocks base method
func (m *MockMetricsEngine) RecordPrebidTimeoutRequests(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPrebidTimeoutRequests", arg0, arg1)
}

// RecordPrebidTimeoutRequests indicates an expected call of RecordPrebidTimeoutRequests
func (mr *MockMetricsEngineMockRecorder) RecordPrebidTimeoutRequests(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPrebidTimeoutRequests", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPrebidTimeoutRequests), arg0, arg1)
}

// RecordPublisherInvalidProfileImpressions mocks base method
func (m *MockMetricsEngine) RecordPublisherInvalidProfileImpressions(arg0, arg1 string, arg2 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPublisherInvalidProfileImpressions", arg0, arg1, arg2)
}

// RecordPublisherInvalidProfileImpressions indicates an expected call of RecordPublisherInvalidProfileImpressions
func (mr *MockMetricsEngineMockRecorder) RecordPublisherInvalidProfileImpressions(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPublisherInvalidProfileImpressions", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPublisherInvalidProfileImpressions), arg0, arg1, arg2)
}

// RecordPublisherInvalidProfileRequests mocks base method
func (m *MockMetricsEngine) RecordPublisherInvalidProfileRequests(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPublisherInvalidProfileRequests", arg0, arg1, arg2)
}

// RecordPublisherInvalidProfileRequests indicates an expected call of RecordPublisherInvalidProfileRequests
func (mr *MockMetricsEngineMockRecorder) RecordPublisherInvalidProfileRequests(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPublisherInvalidProfileRequests", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPublisherInvalidProfileRequests), arg0, arg1, arg2)
}

// RecordPublisherPartnerNoCookieStats mocks base method
func (m *MockMetricsEngine) RecordPublisherPartnerNoCookieStats(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPublisherPartnerNoCookieStats", arg0, arg1)
}

// RecordPublisherPartnerNoCookieStats indicates an expected call of RecordPublisherPartnerNoCookieStats
func (mr *MockMetricsEngineMockRecorder) RecordPublisherPartnerNoCookieStats(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPublisherPartnerNoCookieStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPublisherPartnerNoCookieStats), arg0, arg1)
}

// RecordPublisherProfileRequests mocks base method
func (m *MockMetricsEngine) RecordPublisherProfileRequests(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPublisherProfileRequests", arg0, arg1)
}

// RecordPublisherProfileRequests indicates an expected call of RecordPublisherProfileRequests
func (mr *MockMetricsEngineMockRecorder) RecordPublisherProfileRequests(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPublisherProfileRequests", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPublisherProfileRequests), arg0, arg1)
}

// RecordPublisherRequests mocks base method
func (m *MockMetricsEngine) RecordPublisherRequests(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPublisherRequests", arg0, arg1, arg2)
}

// RecordPublisherRequests indicates an expected call of RecordPublisherRequests
func (mr *MockMetricsEngineMockRecorder) RecordPublisherRequests(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPublisherRequests", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPublisherRequests), arg0, arg1, arg2)
}

// RecordPublisherResponseEncodingErrorStats mocks base method
func (m *MockMetricsEngine) RecordPublisherResponseEncodingErrorStats(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPublisherResponseEncodingErrorStats", arg0)
}

// RecordPublisherResponseEncodingErrorStats indicates an expected call of RecordPublisherResponseEncodingErrorStats
func (mr *MockMetricsEngineMockRecorder) RecordPublisherResponseEncodingErrorStats(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPublisherResponseEncodingErrorStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPublisherResponseEncodingErrorStats), arg0)
}

// RecordPublisherResponseTimeStats mocks base method
func (m *MockMetricsEngine) RecordPublisherResponseTimeStats(arg0 string, arg1 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPublisherResponseTimeStats", arg0, arg1)
}

// RecordPublisherResponseTimeStats indicates an expected call of RecordPublisherResponseTimeStats
func (mr *MockMetricsEngineMockRecorder) RecordPublisherResponseTimeStats(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPublisherResponseTimeStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPublisherResponseTimeStats), arg0, arg1)
}

// RecordPublisherWrapperLoggerFailure mocks base method
func (m *MockMetricsEngine) RecordPublisherWrapperLoggerFailure(arg0, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordPublisherWrapperLoggerFailure", arg0, arg1, arg2)
}

// RecordPublisherWrapperLoggerFailure indicates an expected call of RecordPublisherWrapperLoggerFailure
func (mr *MockMetricsEngineMockRecorder) RecordPublisherWrapperLoggerFailure(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordPublisherWrapperLoggerFailure", reflect.TypeOf((*MockMetricsEngine)(nil).RecordPublisherWrapperLoggerFailure), arg0, arg1, arg2)
}

// RecordReqImpsWithContentCount mocks base method
func (m *MockMetricsEngine) RecordReqImpsWithContentCount(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordReqImpsWithContentCount", arg0, arg1)
}

// RecordReqImpsWithContentCount indicates an expected call of RecordReqImpsWithContentCount
func (mr *MockMetricsEngineMockRecorder) RecordReqImpsWithContentCount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordReqImpsWithContentCount", reflect.TypeOf((*MockMetricsEngine)(nil).RecordReqImpsWithContentCount), arg0, arg1)
}

// RecordRequest mocks base method
func (m *MockMetricsEngine) RecordRequest(arg0 metrics.Labels) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordRequest", arg0)
}

// RecordRequest indicates an expected call of RecordRequest
func (mr *MockMetricsEngineMockRecorder) RecordRequest(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordRequest", reflect.TypeOf((*MockMetricsEngine)(nil).RecordRequest), arg0)
}

// RecordRequestAdPodGeneratedImpressionsCount mocks base method
func (m *MockMetricsEngine) RecordRequestAdPodGeneratedImpressionsCount(arg0 int, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordRequestAdPodGeneratedImpressionsCount", arg0, arg1)
}

// RecordRequestAdPodGeneratedImpressionsCount indicates an expected call of RecordRequestAdPodGeneratedImpressionsCount
func (mr *MockMetricsEngineMockRecorder) RecordRequestAdPodGeneratedImpressionsCount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordRequestAdPodGeneratedImpressionsCount", reflect.TypeOf((*MockMetricsEngine)(nil).RecordRequestAdPodGeneratedImpressionsCount), arg0, arg1)
}

// RecordRequestTime mocks base method
func (m *MockMetricsEngine) RecordRequestTime(arg0 string, arg1 time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordRequestTime", arg0, arg1)
}

// RecordRequestTime indicates an expected call of RecordRequestTime
func (mr *MockMetricsEngineMockRecorder) RecordRequestTime(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordRequestTime", reflect.TypeOf((*MockMetricsEngine)(nil).RecordRequestTime), arg0, arg1)
}

// RecordSSTimeoutRequests mocks base method
func (m *MockMetricsEngine) RecordSSTimeoutRequests(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordSSTimeoutRequests", arg0, arg1)
}

// RecordSSTimeoutRequests indicates an expected call of RecordSSTimeoutRequests
func (mr *MockMetricsEngineMockRecorder) RecordSSTimeoutRequests(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordSSTimeoutRequests", reflect.TypeOf((*MockMetricsEngine)(nil).RecordSSTimeoutRequests), arg0, arg1)
}

// RecordSendLoggerDataTime mocks base method
func (m *MockMetricsEngine) RecordSendLoggerDataTime(arg0, arg1 string, arg2 time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordSendLoggerDataTime", arg0, arg1, arg2)
}

// RecordSendLoggerDataTime indicates an expected call of RecordSendLoggerDataTime
func (mr *MockMetricsEngineMockRecorder) RecordSendLoggerDataTime(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordSendLoggerDataTime", reflect.TypeOf((*MockMetricsEngine)(nil).RecordSendLoggerDataTime), arg0, arg1, arg2)
}

// RecordStatsKeyCTVPrebidFailedImpression mocks base method
func (m *MockMetricsEngine) RecordStatsKeyCTVPrebidFailedImpression(arg0 int, arg1, arg2 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordStatsKeyCTVPrebidFailedImpression", arg0, arg1, arg2)
}

// RecordStatsKeyCTVPrebidFailedImpression indicates an expected call of RecordStatsKeyCTVPrebidFailedImpression
func (mr *MockMetricsEngineMockRecorder) RecordStatsKeyCTVPrebidFailedImpression(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordStatsKeyCTVPrebidFailedImpression", reflect.TypeOf((*MockMetricsEngine)(nil).RecordStatsKeyCTVPrebidFailedImpression), arg0, arg1, arg2)
}

// RecordUidsCookieNotPresentErrorStats mocks base method
func (m *MockMetricsEngine) RecordUidsCookieNotPresentErrorStats(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordUidsCookieNotPresentErrorStats", arg0, arg1)
}

// RecordUidsCookieNotPresentErrorStats indicates an expected call of RecordUidsCookieNotPresentErrorStats
func (mr *MockMetricsEngineMockRecorder) RecordUidsCookieNotPresentErrorStats(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordUidsCookieNotPresentErrorStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordUidsCookieNotPresentErrorStats), arg0, arg1)
}

// RecordVideoImpDisabledViaConnTypeStats mocks base method
func (m *MockMetricsEngine) RecordVideoImpDisabledViaConnTypeStats(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordVideoImpDisabledViaConnTypeStats", arg0, arg1)
}

// RecordVideoImpDisabledViaConnTypeStats indicates an expected call of RecordVideoImpDisabledViaConnTypeStats
func (mr *MockMetricsEngineMockRecorder) RecordVideoImpDisabledViaConnTypeStats(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordVideoImpDisabledViaConnTypeStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordVideoImpDisabledViaConnTypeStats), arg0, arg1)
}

// RecordVideoInstlImpsStats mocks base method
func (m *MockMetricsEngine) RecordVideoInstlImpsStats(arg0, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordVideoInstlImpsStats", arg0, arg1)
}

// RecordVideoInstlImpsStats indicates an expected call of RecordVideoInstlImpsStats
func (mr *MockMetricsEngineMockRecorder) RecordVideoInstlImpsStats(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordVideoInstlImpsStats", reflect.TypeOf((*MockMetricsEngine)(nil).RecordVideoInstlImpsStats), arg0, arg1)
}

// Shutdown mocks base method
func (m *MockMetricsEngine) Shutdown() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Shutdown")
}

// Shutdown indicates an expected call of Shutdown
func (mr *MockMetricsEngineMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockMetricsEngine)(nil).Shutdown))
}
