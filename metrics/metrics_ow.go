package metrics

type OWMetricsEngine interface {
	RecordHttpCounter()
	//RecordBids records the bidder deal bids labeled by pubid, profile, bidder and deal
	RecordBids(pubid, profileid, bidder, deal string)
	//RecordVastVersion record the count of vast version labelled by bidder and vast version
	RecordVastVersion(coreBidder, vastVersion string)
	//RecordVASTTagType record the count of vast tag type labeled by bidder and vast tag
	RecordVASTTagType(bidder, vastTagType string)

	RecordPanic(hostname, method string)
}
