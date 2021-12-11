package constant

const (
	CTVImpressionIDSeparator = `_`
	CTVImpressionIDFormat    = `%v` + CTVImpressionIDSeparator + `%v`
	CTVUniqueBidIDFormat     = `%v-%v`
	HTTPPrefix               = `http`

	//VAST Constants
	VASTDefaultVersion    = 2.0
	VASTMaxVersion        = 4.0
	VASTDefaultVersionStr = `2.0`
	VASTDefaultTag        = `<VAST version="` + VASTDefaultVersionStr + `"/>`
	VASTElement           = `VAST`
	VASTAdElement         = `Ad`
	VASTWrapperElement    = `Wrapper`
	VASTAdTagURIElement   = `VASTAdTagURI`
	VASTVersionAttribute  = `version`
	VASTSequenceAttribute = `sequence`

	CTVAdpod  = `adpod`
	CTVOffset = `offset`
)

var (
	VASTVersionsStr = []string{"0", "1.0", "2.0", "3.0", "4.0"}
)

const (
	UnableToGenerateImpressions = `prebid_ctv unable to generate impressions for adpod`
	DurationMismatchError       = `prebid_ctv all bids filtered while matching lineitem duration`
	UnableToGenerateAdPod       = `prebid_ctv unable to generate adpod from bids combinations`
)

//BidStatus contains bids filtering reason
type BidStatus = int

const (
	//StatusOK ...
	StatusOK BidStatus = 0
	//StatusWinningBid ...
	StatusWinningBid BidStatus = 1
	//StatusCategoryExclusion ...
	StatusCategoryExclusion BidStatus = 2
	//StatusDomainExclusion ...
	StatusDomainExclusion BidStatus = 3
)

// MonitorKey provides the unique key for moniroting the algorithms
type MonitorKey string

const (
	// CombinationGeneratorV1 ...
	CombinationGeneratorV1 MonitorKey = "comp_exclusion_v1"
	// CompetitiveExclusionV1 ...
	CompetitiveExclusionV1 MonitorKey = "comp_exclusion_v1"
)
