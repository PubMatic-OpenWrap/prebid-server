package metrics

import (
	"time"

	"github.com/prebid/openrtb/v20/openrtb3"
)

const (
	XMLParserLabelFastXML = "fastxml"
	XMLParserLabelETree   = "etree"
)

type OWMetricsEngine interface {
	//RecordBids records the bidder deal bids labeled by pubid, profile, bidder and deal
	RecordBids(pubid, profileid, bidder, deal string)
	//RecordVastVersion record the count of vast version labelled by bidder and vast version
	RecordVastVersion(coreBidder, vastVersion string)
	//RecordVASTTagType record the count of vast tag type labeled by bidder and vast tag
	RecordVASTTagType(bidder, vastTagType string)

	RecordPanic(hostname, method string)
	RecordBadRequest(endpoint string, pubId string, nbr *openrtb3.NoBidReason)
	//RecordXMLParserResponseTime records execution time for multiple parsers
	RecordXMLParserResponseTime(parser string, method string, respTime time.Duration)
	//RecordXMLParserResponseMismatch records number of response mismatch
	RecordXMLParserResponseMismatch(method string, isMismatch bool)
}
