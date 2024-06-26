package metrics

import "time"

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

	//RecordXMLParserResponseTime records execution time for multiple parsers
	RecordXMLParserResponseTime(parser string, method string, bidder string, respTime time.Duration)
	//RecordXMLParserResponseMismatch records number of response mismatch
	RecordXMLParserResponseMismatch(method string, bidder string, isMismatch bool)
}
