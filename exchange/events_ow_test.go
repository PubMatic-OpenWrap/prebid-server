package exchange

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/exchange/entities"
	metricConfig "github.com/prebid/prebid-server/v3/metrics/config"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestInjectVideoEventTrackers(t *testing.T) {
	type args struct {
		externalURL string
		vastXML     string
		genbidID    string
		bid         *openrtb2.Bid
		req         *openrtb2.BidRequest
	}
	type want struct {
		wantVastXml string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "linear_creative",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="firstQuartile"><![CDATA[http://example.com/tracking/firstQuartile?k1=v1&k2=v2]]></Tracking><Tracking event="midpoint">http://example.com/tracking/midpoint</Tracking><Tracking event="thirdQuartile">http://example.com/tracking/thirdQuartile</Tracking><Tracking event="complete">http://example.com/tracking/complete</Tracking><Tracking event="start">http://partner.tracking.url</Tracking></TrackingEvents></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				bid: &openrtb2.Bid{
					AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="firstQuartile"><![CDATA[http://example.com/tracking/firstQuartile?k1=v1&k2=v2]]></Tracking><Tracking event="midpoint">http://example.com/tracking/midpoint</Tracking><Tracking event="thirdQuartile">http://example.com/tracking/thirdQuartile</Tracking><Tracking event="complete">http://example.com/tracking/complete</Tracking><Tracking event="start">http://partner.tracking.url</Tracking></TrackingEvents></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="firstQuartile"><![CDATA[http://example.com/tracking/firstQuartile?k1=v1&k2=v2]]></Tracking><Tracking event="midpoint"><![CDATA[http://example.com/tracking/midpoint]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://example.com/tracking/thirdQuartile]]></Tracking><Tracking event="complete"><![CDATA[http://example.com/tracking/complete]]></Tracking><Tracking event="start"><![CDATA[http://partner.tracking.url]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			},
		},
		{
			name: "non_linear_creative",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><NonLinearAds><TrackingEvents><Tracking event="firstQuartile">http://something.com</Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine></Ad></VAST>`,
				bid: &openrtb2.Bid{
					AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><NonLinearAds><TrackingEvents><Tracking event="firstQuartile">http://something.com</Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine></Ad></VAST>`,
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><NonLinearAds><TrackingEvents><Tracking event="firstQuartile"><![CDATA[http://something.com]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine></Ad></VAST>`,
			},
		},
		{
			name: "all_inline_wrapper_liner_and_non_linear_creative",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></InLine><Wrapper><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{
					AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></InLine><Wrapper><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></Wrapper></Ad></VAST>`,
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></InLine><Wrapper><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></Wrapper></Ad></VAST>`,
			},
		},
		{
			name: "no_traker_url_configured", // expect no injection
			args: args{
				externalURL: "",
				vastXML:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				bid: &openrtb2.Bid{
					AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			},
		},
		{
			name: "wrapper_vast_xml_from_partner", // expect we are injecting trackers inside wrapper
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="4.2" xmlns="http://www.iab.com/VAST"><Ad id="20011" sequence="1"><Wrapper followAdditionalWrappers="0" allowMultipleAds="1" fallbackOnNoAd="0"><AdSystem version="4.0">iabtechlab</AdSystem><VASTAdTagURI>http://somevasturl</VASTAdTagURI><Impression id="Impression-ID"><![CDATA[https://example.com/track/impression]]></Impression><Creatives><Creative id="5480" sequence="1" adId="2447226"><Linear></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{ // Adm contains to TrackingEvents tag
					AdM: `<VAST version="4.2" xmlns="http://www.iab.com/VAST"><Ad id="20011" sequence="1"><Wrapper followAdditionalWrappers="0" allowMultipleAds="1" fallbackOnNoAd="0"><AdSystem version="4.0">iabtechlab</AdSystem><VASTAdTagURI>http://somevasturl</VASTAdTagURI><Impression id="Impression-ID"><![CDATA[https://example.com/track/impression]]></Impression><Creatives><Creative id="5480" sequence="1" adId="2447226"><Linear></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="4.2" xmlns="http://www.iab.com/VAST"><Ad id="20011" sequence="1"><Wrapper followAdditionalWrappers="0" allowMultipleAds="1" fallbackOnNoAd="0"><AdSystem version="4.0"><![CDATA[iabtechlab]]></AdSystem><VASTAdTagURI><![CDATA[http://somevasturl]]></VASTAdTagURI><Impression id="Impression-ID"><![CDATA[https://example.com/track/impression]]></Impression><Creatives><Creative id="5480" sequence="1" adId="2447226"><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
			},
		},
		{
			name: "vast_tag_uri_response_from_partner",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<![CDATA[http://hostedvasttag.url&k=v]]>`,
				bid: &openrtb2.Bid{
					AdM: `<![CDATA[http://hostedvasttag.url&k=v]]>`,
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<![CDATA[http://hostedvasttag.url&k=v]]>`,
			},
		},
		{
			name: "adm_empty_with_vast_build_from_modifyBidVAST",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   "",
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
			},
		},
		{
			name: "adm_empty_with_vast_build_from_modifyBidVAST_non_video",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   "",
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
			},
		},
		{
			name: "vast_without_creative_tracking_node.Only_till_xpath_VAST/Ad/Wrapper/Creatives",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></Wrapper></Ad></VAST>`,
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></Wrapper></Ad></VAST>`,
			},
		},
		{
			name: "vast_without_linear_node.Only_till_xpath_VAST/Ad/Wrapper/Creatives/Creative",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative></Creative></Creatives></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative></Creative></Creatives></Wrapper></Ad></VAST>`,
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative></Creative></Creatives></Wrapper></Ad></VAST>`,
			},
		},
		{
			name: "vast_without_tracking_node.Only_till_xpath_VAST/Ad/Wrapper/Creatives/Creative/Linear",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
			},
		},
		{
			name: "vast_without_tracking_node.Only_till_xpath_VAST/Ad/Wrapper/Creatives/Creative/NonLinear",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><NonLinearAds></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><NonLinearAds></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
			},
		},
		{
			name: "vast_without_creative_tracking_node.Only_till_xpath_VAST/Ad/InLine/Creatives",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></InLine></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></InLine></Ad></VAST>`,
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></InLine></Ad></VAST>`,
			},
		},
		{
			name: "vast_without_linear_node.Only_till_xpath_VAST/Ad/InLine/Creatives/Creative",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative></Creative></Creatives></InLine></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative></Creative></Creatives></InLine></Ad></VAST>`,
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative></Creative></Creatives></InLine></Ad></VAST>`,
			},
		},
		{
			name: "vast_without_tracking_node.Only_till_xpath_VAST/Ad/InLine/Creatives/Creative/Linear",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear></Linear></Creative></Creatives></InLine></Ad></VAST>`,
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			},
		},
		{
			name: "vast_without_tracking_node.Only_till_xpath_VAST/Ad/InLine/Creatives/Creative/NonLinear",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><NonLinearAds></NonLinearAds></Creative></Creatives></InLine></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><NonLinearAds></NonLinearAds></Creative></Creatives></InLine></Ad></VAST>`,
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine></Ad></VAST>`,
			},
		},
		{
			name: "vast_without_tracking_node_and_multiple_creative.All_4_xpath_Wrapper_InLine_Linear_NonLinear",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear></Linear><NonLinearAds></NonLinearAds></Creative><Creative><Linear></Linear><NonLinearAds></NonLinearAds></Creative></Creatives></InLine><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear></Linear><NonLinearAds></NonLinearAds></Creative><Creative><Linear></Linear><NonLinearAds></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear></Linear><NonLinearAds></NonLinearAds></Creative><Creative><Linear></Linear><NonLinearAds></NonLinearAds></Creative></Creatives></InLine><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear></Linear><NonLinearAds></NonLinearAds></Creative><Creative><Linear></Linear><NonLinearAds></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
			},
		},
		{
			name: "vast_with_tracking_node_and_multiple_creative.All_4_xpath_Wrapper_InLine_Linear_NonLinear",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{
					ImpID: "123",
					AdM:   `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
					NURL:  "nurl_contents",
				},
				req: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			want: want{
				wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
			},
		},
		// {
		// 	name: "vast_and_adm_empty - This should never be the case as modifyBidVAST always updates AdM with tempate vast",
		// 	args: args{
		// 		externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
		// 		vastXML:     "",
		// 		bid: &openrtb2.Bid{ // Adm contains to TrackingEvents tag
		// 			AdM:  "",
		// 			NURL: "nurl_contents",
		// 		},
		// 		req: &openrtb2.BidRequest{App: &openrtb2.App{Bundle: "abc"}},
		// 	},
		//	want : want{
		//  wantVastXml: "",
		// 	wantErr:     errors.New("error parsing VAST XML. 'EOF'"),
		//  metrics: &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
		// },
		// },
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ev := eventTracking{OpenWrapEventTracking: OpenWrapEventTracking{enabledVideoEvents: true, me: &metricConfig.NilMetricsEngine{}}}
			ev.injectVideoEvents(tc.args.req, tc.args.bid, tc.args.vastXML, tc.args.externalURL, tc.args.genbidID, "test_bidder", "test_core_bidder")
			assert.Equal(t, tc.want.wantVastXml, tc.args.bid.AdM)
		})
	}
}

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
					me:                 &metricConfig.NilMetricsEngine{},
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

// BenchmarkInjectVideoEventTrackers
// nilesh@9fc43242aec1: git checkout origin/ci
// nilesh@9fc43242aec1:~/go/src/github.com/PubMatic-OpenWrap/prebid-server/endpoints/events$ go test -bench=BenchmarkInjectVideoEventTrackers -count 200 -run=^# | tee old1.txt
// nilesh@9fc43242aec1: git checkout origin/UOE-8632-ci-1
// nilesh@9fc43242aec1:~/go/src/github.com/PubMatic-OpenWrap/prebid-server/endpoints/events$ go test -bench=BenchmarkInjectVideoEventTrackers -count 200 -run=^# | tee new1.txt
// nilesh@9fc43242aec1:~/go/src/github.com/PubMatic-OpenWrap/prebid-server/endpoints/events$ benchstat old1.txt new1.txt
// goos: linux
// goarch: arm64
// pkg: github.com/PubMatic-OpenWrap/prebid-server/endpoints/events
//
//	│   old1.txt   │          new1.txt           │
//	│    sec/op    │   sec/op     vs base        │
//
// InjectVideoEventTrackers-8   107.83µ ± 1%   97.62µ ± 1%  -9.47% (n=200)
func BenchmarkInjectVideoEventTrackers(b *testing.B) {
	trackerURL := "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]"
	vastXML := `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></InLine><Wrapper><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></Wrapper></Ad></VAST>`
	bid := &openrtb2.Bid{
		AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></InLine><Wrapper><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></Wrapper></Ad></VAST>`,
	}
	req := &openrtb2.BidRequest{
		App: &openrtb2.App{Bundle: "abc"},
		Imp: []openrtb2.Imp{{ID: "123", Video: &openrtb2.Video{}}},
	}
	ev := eventTracking{OpenWrapEventTracking: OpenWrapEventTracking{enabledVideoEvents: true, me: &metricConfig.NilMetricsEngine{}}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// no need to validate, using vast from test TestInjectVideoEventTrackers/all_inline_wrapper_liner_and_non_linear_creative
		ev.injectVideoEvents(req, bid, vastXML, trackerURL, "", "test_bidder", "test_core_bidder")
	}
}
