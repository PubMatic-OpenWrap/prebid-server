package metrics

import (
	"time"

	"github.com/prebid/openrtb/v20/openrtb3"
)

// RecordAdapterDuplicateBidID mock
func (me *MetricsEngineMock) RecordAdapterDuplicateBidID(adaptor string, collisions int) {
	me.Called(adaptor, collisions)
}

// RecordRequestHavingDuplicateBidID mock
func (me *MetricsEngineMock) RecordRequestHavingDuplicateBidID() {
	me.Called()
}

// RecordPodImpGenTime mock
func (me *MetricsEngineMock) RecordPodImpGenTime(labels PodLabels, startTime time.Time) {
	me.Called(labels, startTime)
}

// RecordPodCombGenTime mock
func (me *MetricsEngineMock) RecordPodCombGenTime(labels PodLabels, elapsedTime time.Duration) {
	me.Called(labels, elapsedTime)
}

// RecordPodCompititveExclusionTime mock
func (me *MetricsEngineMock) RecordPodCompititveExclusionTime(labels PodLabels, elapsedTime time.Duration) {
	me.Called(labels, elapsedTime)
}

// RecordAdapterVideoBidDuration mock
func (me *MetricsEngineMock) RecordAdapterVideoBidDuration(labels AdapterLabels, videoBidDuration int) {
	me.Called(labels, videoBidDuration)
}

func (me *MetricsEngineMock) RecordBids(pubid, profileid, biddder, deal string) {
	me.Called(pubid, profileid, biddder, deal)
}

// RecordVastVersion mock
func (me *MetricsEngineMock) RecordVastVersion(coreBidder, vastVersion string) {
	me.Called(coreBidder, vastVersion)
}

// RecordVASTTagType mock
func (me *MetricsEngineMock) RecordVASTTagType(bidder, vastTagType string) {
	me.Called(bidder, vastTagType)
}

// RecordPanic mock
func (me *MetricsEngineMock) RecordPanic(hostname, method string) {
	me.Called(hostname, method)
}

// RecordPanic mock
func (me *MetricsEngineMock) RecordBadRequest(endpoint string, pubId string, nbr *openrtb3.NoBidReason) {
	me.Called(endpoint, pubId, nbr)
}

// RecordXMLParserProcessingTime records execution time for multiple parsers
func (me *MetricsEngineMock) RecordXMLParserProcessingTime(parser string, method string, param string, respTime time.Duration) {
	me.Called(parser, method, param, respTime)
}

// RecordXMLParserResponseMismatch mock
func (me *MetricsEngineMock) RecordXMLParserResponseMismatch(method string, param string, isMismatch bool) {
	me.Called(method, param, isMismatch)
}

// RecordXMLParserResponseTime records execution time for multiple parsers
func (me *MetricsEngineMock) RecordXMLParserResponseTime(parser string, method string, param string, respTime time.Duration) {
	me.Called(parser, method, param, respTime)
}
