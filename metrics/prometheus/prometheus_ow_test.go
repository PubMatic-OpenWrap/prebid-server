package prometheusmetrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestRecordRejectedBids(t *testing.T) {
	type testIn struct {
		pubid, bidder, code string
	}
	type testOut struct {
		expCount int
	}
	testCases := []struct {
		description string
		in          testIn
		out         testOut
	}{
		{
			description: "record rejected bids",
			in: testIn{
				pubid:  "1010",
				bidder: "bidder",
				code:   "100",
			},
			out: testOut{
				expCount: 1,
			},
		},
	}
	for _, test := range testCases {
		pm := createMetricsForTesting()
		pm.RecordRejectedBids(test.in.pubid, test.in.bidder, test.in.code)

		assertCounterVecValue(t,
			"",
			"rejected_bids",
			pm.rejectedBids,
			float64(test.out.expCount),
			prometheus.Labels{
				pubIDLabel:  test.in.pubid,
				bidderLabel: test.in.bidder,
				codeLabel:   test.in.code,
			})
	}
}

func TestRecordBids(t *testing.T) {
	type testIn struct {
		pubid, profileid, bidder, deal string
	}
	type testOut struct {
		expCount int
	}
	testCases := []struct {
		description string
		in          testIn
		out         testOut
	}{
		{
			description: "record bids",
			in: testIn{
				pubid:     "1010",
				bidder:    "bidder",
				profileid: "11",
				deal:      "pubdeal",
			},
			out: testOut{
				expCount: 1,
			},
		},
	}
	for _, test := range testCases {
		pm := createMetricsForTesting()
		pm.RecordBids(test.in.pubid, test.in.profileid, test.in.bidder, test.in.deal)

		assertCounterVecValue(t,
			"",
			"bids",
			pm.bids,
			float64(test.out.expCount),
			prometheus.Labels{
				pubIDLabel:   test.in.pubid,
				bidderLabel:  test.in.bidder,
				profileLabel: test.in.profileid,
				dealLabel:    test.in.deal,
			})
	}
}

func TestRecordVastVersion(t *testing.T) {
	type testIn struct {
		coreBidder, vastVersion string
	}
	type testOut struct {
		expCount int
	}
	testCases := []struct {
		description string
		in          testIn
		out         testOut
	}{
		{
			description: "record vast version",
			in: testIn{
				coreBidder:  "bidder",
				vastVersion: "2.0",
			},
			out: testOut{
				expCount: 1,
			},
		},
	}
	for _, test := range testCases {
		pm := createMetricsForTesting()
		pm.RecordVastVersion(test.in.coreBidder, test.in.vastVersion)
		assertCounterVecValue(t,
			"",
			"record vastVersion",
			pm.vastVersion,
			float64(test.out.expCount),
			prometheus.Labels{
				adapterLabel: test.in.coreBidder,
				versionLabel: test.in.vastVersion,
			})
	}
}

func TestRecordVASTTagType(t *testing.T) {
	type testIn struct {
		bidder, vastTag string
	}
	type testOut struct {
		expCount int
	}
	testCases := []struct {
		description string
		in          testIn
		out         testOut
	}{
		{
			description: "record vast tag",
			in: testIn{
				bidder:  "bidder",
				vastTag: "Wrapper",
			},
			out: testOut{
				expCount: 1,
			},
		},
	}
	for _, test := range testCases {
		pm := createMetricsForTesting()
		pm.RecordVASTTagType(test.in.bidder, test.in.vastTag)
		assertCounterVecValue(t,
			"",
			"record vastTag",
			pm.vastTagType,
			float64(test.out.expCount),
			prometheus.Labels{
				bidderLabel:      test.in.bidder,
				vastTagTypeLabel: test.in.vastTag,
			})
	}
}
