package metrics

import "time"

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

// RecordVastTag mock
func (me *MetricsEngineMock) RecordVastTag(bidder, vastTag string) {
	me.Called(bidder, vastTag)
}
