package exchange

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/exchange/entities"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestModifyBidVAST(t *testing.T) {
	type args struct {
		enabledVideoEvents bool
		bidReq             *openrtb2.BidRequest
		bid                *openrtb2.Bid
	}
	tests := []struct {
		name    string
		args    args
		wantAdM string
	}{
		{
			name: "empty_adm", // expect adm contain vast tag with tracking events and  VASTAdTagURI nurl contents
			args: args{
				enabledVideoEvents: true,
				bidReq: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
				},
				bid: &openrtb2.Bid{
					AdM:   "",
					NURL:  "nurl_contents",
					ImpID: "123",
				},
			},
			wantAdM: `<VAST version="3.0"><Ad><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?e=2]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?e=4]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?e=3]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?e=5]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?e=6]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?e=2]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?e=4]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?e=3]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?e=5]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?e=6]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
		},
		{
			name: "adm_containing_url", // expect adm contain vast tag with tracking events and  VASTAdTagURI adm url (previous value) contents
			args: args{
				enabledVideoEvents: true,
				bidReq: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
				},
				bid: &openrtb2.Bid{
					AdM:   "http://vast_tag_inline.xml",
					NURL:  "nurl_contents",
					ImpID: "123",
				},
			},
			wantAdM: `<VAST version="3.0"><Ad><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[http://vast_tag_inline.xml]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?e=2]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?e=4]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?e=3]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?e=5]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?e=6]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?e=2]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?e=4]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?e=3]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?e=5]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?e=6]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ev := eventTracking{
				bidderInfos: config.BidderInfos{
					"somebidder": config.BidderInfo{
						ModifyingVastXmlAllowed: false,
					},
				},
				OpenWrapEventTracking: OpenWrapEventTracking{
					enabledVideoEvents: tc.args.enabledVideoEvents,
				},
			}
			ev.modifyBidVAST(&entities.PbsOrtbBid{
				Bid:     tc.args.bid,
				BidType: openrtb_ext.BidTypeVideo,
			}, "somebidder", "coreBidder", tc.args.bidReq, "http://company.tracker.com?e=[EVENT_ID]")
			assert.Equal(t, tc.wantAdM, tc.args.bid.AdM)
		})
	}
}
