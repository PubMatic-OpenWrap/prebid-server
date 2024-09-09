package metrics

import (
	"time"

	"github.com/prebid/openrtb/v20/openrtb3"
)

// RecordAdapterDuplicateBidID as noop
func (me *Metrics) RecordAdapterDuplicateBidID(adaptor string, collisions int) {
}

// RecordRequestHavingDuplicateBidID as noop
func (me *Metrics) RecordRequestHavingDuplicateBidID() {
}

// RecordPodImpGenTime as a noop
func (me *Metrics) RecordPodImpGenTime(labels PodLabels, startTime time.Time) {
}

// RecordPodCombGenTime as a noop
func (me *Metrics) RecordPodCombGenTime(labels PodLabels, elapsedTime time.Duration) {
}

// RecordPodCompititveExclusionTime as a noop
func (me *Metrics) RecordPodCompititveExclusionTime(labels PodLabels, elapsedTime time.Duration) {
}

// RecordAdapterVideoBidDuration as a noop
func (me *Metrics) RecordAdapterVideoBidDuration(labels AdapterLabels, videoBidDuration int) {
}

// RecordAdapterVideoBidDuration as a noop
func (me *Metrics) RecordRejectedBids(pubid, biddder, code string) {
}

// RecordBids as a noop
func (me *Metrics) RecordBids(pubid, profileid, biddder, deal string) {
}

// RecordVastVersion as a noop
func (me *Metrics) RecordVastVersion(biddder, vastVersion string) {
}

// RecordVASTTagType as a noop
func (me *Metrics) RecordVASTTagType(biddder, vastTag string) {
}

// RecordPanic as a noop
func (me *Metrics) RecordPanic(hostname, method string) {
}

// RecordBadRequest as a noop
func (me *Metrics) RecordBadRequest(endpoint string, pubId string, nbr *openrtb3.NoBidReason) {
}

// RecordXMLParserResponseTime records execution time for multiple parsers
func (me *Metrics) RecordXMLParserResponseTime(parser string, method string, respTime time.Duration) {
}

// RecordXMLParserResponseMismatch as a noop
func (me *Metrics) RecordXMLParserResponseMismatch(method string, isMismatch bool) {
}
