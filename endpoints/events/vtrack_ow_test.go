package events

import (
	"errors"
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
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
	tests := []struct {
		name        string
		args        args
		wantVastXml string
		wantErr     error
	}{
		{
			name: "linear_creative",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="firstQuartile"><![CDATA[http://example.com/tracking/firstQuartile?k1=v1&k2=v2]]></Tracking><Tracking event="midpoint">http://example.com/tracking/midpoint</Tracking><Tracking event="thirdQuartile">http://example.com/tracking/thirdQuartile</Tracking><Tracking event="complete">http://example.com/tracking/complete</Tracking><Tracking event="start">http://partner.tracking.url</Tracking></TrackingEvents></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				bid: &openrtb2.Bid{
					AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="firstQuartile"><![CDATA[http://example.com/tracking/firstQuartile?k1=v1&k2=v2]]></Tracking><Tracking event="midpoint">http://example.com/tracking/midpoint</Tracking><Tracking event="thirdQuartile">http://example.com/tracking/thirdQuartile</Tracking><Tracking event="complete">http://example.com/tracking/complete</Tracking><Tracking event="start">http://partner.tracking.url</Tracking></TrackingEvents></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				req: &openrtb2.BidRequest{App: &openrtb2.App{Bundle: "abc"}},
			},
			wantVastXml: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="firstQuartile"><![CDATA[http://example.com/tracking/firstQuartile?k1=v1&k2=v2]]></Tracking><Tracking event="midpoint"><![CDATA[http://example.com/tracking/midpoint]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://example.com/tracking/thirdQuartile]]></Tracking><Tracking event="complete"><![CDATA[http://example.com/tracking/complete]]></Tracking><Tracking event="start"><![CDATA[http://partner.tracking.url]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			wantErr:     nil,
		},
		{
			name: "non_linear_creative",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><NonLinearAds><TrackingEvents><Tracking event="firstQuartile">http://something.com</Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine></Ad></VAST>`,
				bid: &openrtb2.Bid{
					AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><NonLinearAds><TrackingEvents><Tracking event="firstQuartile">http://something.com</Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine></Ad></VAST>`,
				},
				req: &openrtb2.BidRequest{App: &openrtb2.App{Bundle: "abc"}},
			},
			wantVastXml: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><NonLinearAds><TrackingEvents><Tracking event="firstQuartile"><![CDATA[http://something.com]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine></Ad></VAST>`,
			wantErr:     nil,
		},
		{
			name: "all_inline_wrapper_liner_and_non_linear_creative",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></InLine><Wrapper><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{
					AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></InLine><Wrapper><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression></Wrapper></Ad></VAST>`,
				},
				req: &openrtb2.BidRequest{App: &openrtb2.App{Bundle: "abc"}},
			},
			wantVastXml: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression/></InLine><Wrapper><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression/></Wrapper></Ad></VAST>`,
			wantErr:     nil,
		},
		{
			name: "no_traker_url_configured", // expect no injection
			args: args{
				externalURL: "",
				vastXML:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				bid: &openrtb2.Bid{
					AdM: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				req: &openrtb2.BidRequest{App: &openrtb2.App{Bundle: "abc"}},
			},
			wantVastXml: `<VAST version="3.0"><Ad><InLine><Creatives><Creative><Linear></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			wantErr:     errors.New("event URLs not found"),
		},
		{
			name: "wrapper_vast_xml_from_partner", // expect we are injecting trackers inside wrapper
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<VAST version="4.2" xmlns="http://www.iab.com/VAST"><Ad id="20011" sequence="1" ><Wrapper followAdditionalWrappers="0" allowMultipleAds="1" fallbackOnNoAd="0"><AdSystem version="4.0">iabtechlab</AdSystem><VASTAdTagURI>http://somevasturl</VASTAdTagURI><Impression id="Impression-ID"><![CDATA[https://example.com/track/impression]]></Impression><Creatives><Creative id="5480" sequence="1" adId="2447226"><Linear></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
				bid: &openrtb2.Bid{ // Adm contains to TrackingEvents tag
					AdM: `<VAST version="4.2" xmlns="http://www.iab.com/VAST"><Ad id="20011" sequence="1" ><Wrapper followAdditionalWrappers="0" allowMultipleAds="1" fallbackOnNoAd="0"><AdSystem version="4.0">iabtechlab</AdSystem><VASTAdTagURI>http://somevasturl</VASTAdTagURI><Impression id="Impression-ID"><![CDATA[https://example.com/track/impression]]></Impression><Creatives><Creative id="5480" sequence="1" adId="2447226"><Linear></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
				},
				req: &openrtb2.BidRequest{App: &openrtb2.App{Bundle: "abc"}},
			},
			wantVastXml: `<VAST version="4.2" xmlns="http://www.iab.com/VAST"><Ad id="20011" sequence="1"><Wrapper followAdditionalWrappers="0" allowMultipleAds="1" fallbackOnNoAd="0"><AdSystem version="4.0"><![CDATA[iabtechlab]]></AdSystem><VASTAdTagURI><![CDATA[http://somevasturl]]></VASTAdTagURI><Impression id="Impression-ID"><![CDATA[https://example.com/track/impression]]></Impression><Creatives><Creative id="5480" sequence="1" adId="2447226"><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantErr:     nil,
		},
		{
			name: "vast_tag_uri_response_from_partner",
			args: args{
				externalURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				vastXML:     `<![CDATA[http://hostedvasttag.url&k=v]]>`,
				bid: &openrtb2.Bid{
					AdM: `<![CDATA[http://hostedvasttag.url&k=v]]>`,
				},
				req: &openrtb2.BidRequest{App: &openrtb2.App{Bundle: "abc"}},
			},
			wantVastXml: `<![CDATA[http://hostedvasttag.url&k=v]]>`,
			wantErr:     nil,
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
		// 	wantVastXml: "",
		// 	wantErr:     errors.New("error parsing VAST XML. 'EOF'"),
		// },
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
			wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression/><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantErr:     nil,
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
					Imp: []openrtb2.Imp{{ID: "123"}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></Wrapper></Ad></VAST>`,
			wantErr:     nil,
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
					Imp: []openrtb2.Imp{{ID: "123"}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></Wrapper></Ad></VAST>`,
			wantErr:     nil,
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
					Imp: []openrtb2.Imp{{ID: "123"}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantErr:     nil,
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
					Imp: []openrtb2.Imp{{ID: "123"}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression/><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantErr:     nil,
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
					Imp: []openrtb2.Imp{{ID: "123"}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			wantVastXml: `<VAST version="3.0"><Ad><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression/><Creatives><Creative><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantErr:     nil,
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
					Imp: []openrtb2.Imp{{ID: "123"}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></InLine></Ad></VAST>`,
			wantErr:     nil,
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
					Imp: []openrtb2.Imp{{ID: "123"}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem>prebid.org wrapper</AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression></Impression><Creatives><Creative></Creative></Creatives></InLine></Ad></VAST>`,
			wantErr:     nil,
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
					Imp: []openrtb2.Imp{{ID: "123"}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression/><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			wantErr:     nil,
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
					Imp: []openrtb2.Imp{{ID: "123"}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression/><Creatives><Creative><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine></Ad></VAST>`,
			wantErr:     nil,
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
					Imp: []openrtb2.Imp{{ID: "123"}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression/><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression/><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantErr:     nil,
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
					Imp: []openrtb2.Imp{{ID: "123"}},
					App: &openrtb2.App{Bundle: "abc"},
				},
			},
			wantVastXml: `<VAST version="3.0"><Ad><InLine><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression/><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></InLine><Wrapper><AdSystem><![CDATA[prebid.org wrapper]]></AdSystem><VASTAdTagURI><![CDATA[nurl_contents]]></VASTAdTagURI><Impression/><Creatives><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative><Creative><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></Linear><NonLinearAds><TrackingEvents><Tracking event="start"><![CDATA[http://company1.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company1.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company1.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company1.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company1.tracker.com?eventId=6&appbundle=abc]]></Tracking><Tracking event="start"><![CDATA[http://company.tracker.com?eventId=2&appbundle=abc]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://company.tracker.com?eventId=4&appbundle=abc]]></Tracking><Tracking event="midpoint"><![CDATA[http://company.tracker.com?eventId=3&appbundle=abc]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://company.tracker.com?eventId=5&appbundle=abc]]></Tracking><Tracking event="complete"><![CDATA[http://company.tracker.com?eventId=6&appbundle=abc]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantErr:     nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			injectedVast, err := InjectVideoEventTrackers(tc.args.externalURL, tc.args.vastXML, tc.args.bid, tc.args.genbidID, "test_bidder", "test_core_bidder", "", int64(0), tc.args.req)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVastXml, string(injectedVast))
		})
	}
}

func TestGetVideoEventTracking(t *testing.T) {
	type args struct {
		trackerURL       string
		bid              *openrtb2.Bid
		requestingBidder string
		gen_bidid        string
		bidderCoreName   string
		accountId        string
		timestamp        int64
		req              *openrtb2.BidRequest
	}
	type want struct {
		trackerURLMap map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid_scenario",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				bid:        &openrtb2.Bid{
					// AdM: vastXMLWith2Creatives,
				},
				req: &openrtb2.BidRequest{
					App: &openrtb2.App{
						Bundle: "someappbundle",
					},
					Imp: []openrtb2.Imp{},
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					"firstQuartile": "http://company.tracker.com?eventId=4&appbundle=someappbundle",
					"midpoint":      "http://company.tracker.com?eventId=3&appbundle=someappbundle",
					"thirdQuartile": "http://company.tracker.com?eventId=5&appbundle=someappbundle",
					"start":         "http://company.tracker.com?eventId=2&appbundle=someappbundle",
					"complete":      "http://company.tracker.com?eventId=6&appbundle=someappbundle"},
			},
		},
		{
			name: "no_macro_value", // expect no replacement
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				bid:        &openrtb2.Bid{},
				req: &openrtb2.BidRequest{
					App: &openrtb2.App{}, // no app bundle value
					Imp: []openrtb2.Imp{},
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					"firstQuartile": "http://company.tracker.com?eventId=4&appbundle=",
					"midpoint":      "http://company.tracker.com?eventId=3&appbundle=",
					"thirdQuartile": "http://company.tracker.com?eventId=5&appbundle=",
					"start":         "http://company.tracker.com?eventId=2&appbundle=",
					"complete":      "http://company.tracker.com?eventId=6&appbundle="},
			},
		},
		{
			name: "prefer_company_value_for_standard_macro",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]",
				bid:        &openrtb2.Bid{},
				req: &openrtb2.BidRequest{
					App: &openrtb2.App{
						Bundle: "myapp", // do not expect this value
					},
					Imp: []openrtb2.Imp{},
					Ext: []byte(`{"prebid":{
								"macros": {
									"[DOMAIN]": "my_custom_value"
								}
						}}`),
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					"firstQuartile": "http://company.tracker.com?eventId=4&appbundle=my_custom_value",
					"midpoint":      "http://company.tracker.com?eventId=3&appbundle=my_custom_value",
					"thirdQuartile": "http://company.tracker.com?eventId=5&appbundle=my_custom_value",
					"start":         "http://company.tracker.com?eventId=2&appbundle=my_custom_value",
					"complete":      "http://company.tracker.com?eventId=6&appbundle=my_custom_value"},
			},
		},
		{
			name: "multireplace_macro",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&appbundle=[DOMAIN]&parameter2=[DOMAIN]",
				bid:        &openrtb2.Bid{},
				req: &openrtb2.BidRequest{
					App: &openrtb2.App{
						Bundle: "myapp123",
					},
					Imp: []openrtb2.Imp{},
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					"firstQuartile": "http://company.tracker.com?eventId=4&appbundle=myapp123&parameter2=myapp123",
					"midpoint":      "http://company.tracker.com?eventId=3&appbundle=myapp123&parameter2=myapp123",
					"thirdQuartile": "http://company.tracker.com?eventId=5&appbundle=myapp123&parameter2=myapp123",
					"start":         "http://company.tracker.com?eventId=2&appbundle=myapp123&parameter2=myapp123",
					"complete":      "http://company.tracker.com?eventId=6&appbundle=myapp123&parameter2=myapp123"},
			},
		},
		{
			name: "custom_macro_without_prefix_and_suffix",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&param1=[CUSTOM_MACRO]",
				bid:        &openrtb2.Bid{},
				req: &openrtb2.BidRequest{
					Ext: []byte(`{"prebid":{
							"macros": {
								"CUSTOM_MACRO": "my_custom_value"
							}
					}}`),
					Imp: []openrtb2.Imp{},
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					"firstQuartile": "http://company.tracker.com?eventId=4&param1=[CUSTOM_MACRO]",
					"midpoint":      "http://company.tracker.com?eventId=3&param1=[CUSTOM_MACRO]",
					"thirdQuartile": "http://company.tracker.com?eventId=5&param1=[CUSTOM_MACRO]",
					"start":         "http://company.tracker.com?eventId=2&param1=[CUSTOM_MACRO]",
					"complete":      "http://company.tracker.com?eventId=6&param1=[CUSTOM_MACRO]"},
			},
		},
		{
			name: "empty_macro",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&param1=[CUSTOM_MACRO]",
				bid:        &openrtb2.Bid{},
				req: &openrtb2.BidRequest{
					Ext: []byte(`{"prebid":{
							"macros": {
								"": "my_custom_value"
							}
					}}`),
					Imp: []openrtb2.Imp{},
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					"firstQuartile": "http://company.tracker.com?eventId=4&param1=[CUSTOM_MACRO]",
					"midpoint":      "http://company.tracker.com?eventId=3&param1=[CUSTOM_MACRO]",
					"thirdQuartile": "http://company.tracker.com?eventId=5&param1=[CUSTOM_MACRO]",
					"start":         "http://company.tracker.com?eventId=2&param1=[CUSTOM_MACRO]",
					"complete":      "http://company.tracker.com?eventId=6&param1=[CUSTOM_MACRO]"},
			},
		},
		{
			name: "macro_is_case_sensitive",
			args: args{
				trackerURL: "http://company.tracker.com?eventId=[EVENT_ID]&param1=[CUSTOM_MACRO]",
				bid:        &openrtb2.Bid{},
				req: &openrtb2.BidRequest{
					Ext: []byte(`{"prebid":{
							"macros": {
								"": "my_custom_value"
							}
					}}`),
					Imp: []openrtb2.Imp{},
				},
			},
			want: want{
				trackerURLMap: map[string]string{
					"firstQuartile": "http://company.tracker.com?eventId=4&param1=[CUSTOM_MACRO]",
					"midpoint":      "http://company.tracker.com?eventId=3&param1=[CUSTOM_MACRO]",
					"thirdQuartile": "http://company.tracker.com?eventId=5&param1=[CUSTOM_MACRO]",
					"start":         "http://company.tracker.com?eventId=2&param1=[CUSTOM_MACRO]",
					"complete":      "http://company.tracker.com?eventId=6&param1=[CUSTOM_MACRO]"},
			},
		},
		{
			name: "empty_tracker_url",
			args: args{trackerURL: "    ", bid: &openrtb2.Bid{}, req: &openrtb2.BidRequest{Imp: []openrtb2.Imp{}}},
			want: want{trackerURLMap: nil},
		},
		{
			name: "site_domain_tracker_url",
			args: args{
				trackerURL: "https://company.tracker.com?operId=8&e=[EVENT_ID]&p=[PBS-ACCOUNT]&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=[PBS-BIDDER]&advertiser_id=[ADVERTISER_NAME]&sURL=[DOMAIN]&pfi=[PLATFORM]&af=[ADTYPE]&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=[AD_UNIT]&bidid=[PBS-BIDID]",
				bid:        &openrtb2.Bid{}, req: &openrtb2.BidRequest{Site: &openrtb2.Site{Name: "test", Domain: "www.test.com", Publisher: &openrtb2.Publisher{ID: "5890"}}, Imp: []openrtb2.Imp{}}},
			want: want{
				map[string]string{
					"complete":      "https://company.tracker.com?operId=8&e=6&p=5890&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=&advertiser_id=&sURL=www.test.com&pfi=[PLATFORM]&af=video&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=&bidid=",
					"firstQuartile": "https://company.tracker.com?operId=8&e=4&p=5890&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=&advertiser_id=&sURL=www.test.com&pfi=[PLATFORM]&af=video&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=&bidid=",
					"midpoint":      "https://company.tracker.com?operId=8&e=3&p=5890&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=&advertiser_id=&sURL=www.test.com&pfi=[PLATFORM]&af=video&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=&bidid=",
					"start":         "https://company.tracker.com?operId=8&e=2&p=5890&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=&advertiser_id=&sURL=www.test.com&pfi=[PLATFORM]&af=video&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=&bidid=",
					"thirdQuartile": "https://company.tracker.com?operId=8&e=5&p=5890&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=&advertiser_id=&sURL=www.test.com&pfi=[PLATFORM]&af=video&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=&bidid=",
				},
			},
		},
		{
			name: "site_page_tracker_url",
			args: args{trackerURL: "https://company.tracker.com?operId=8&e=[EVENT_ID]&p=[PBS-ACCOUNT]&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=[PBS-BIDDER]&advertiser_id=[ADVERTISER_NAME]&sURL=[DOMAIN]&pfi=[PLATFORM]&af=[ADTYPE]&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=[AD_UNIT]&bidid=[PBS-BIDID]",
				bid: &openrtb2.Bid{}, req: &openrtb2.BidRequest{Site: &openrtb2.Site{Name: "test", Page: "https://www.test.com/", Publisher: &openrtb2.Publisher{ID: "5890"}}, Imp: []openrtb2.Imp{}}},
			want: want{
				map[string]string{
					"complete":      "https://company.tracker.com?operId=8&e=6&p=5890&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=&advertiser_id=&sURL=www.test.com&pfi=[PLATFORM]&af=video&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=&bidid=",
					"firstQuartile": "https://company.tracker.com?operId=8&e=4&p=5890&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=&advertiser_id=&sURL=www.test.com&pfi=[PLATFORM]&af=video&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=&bidid=",
					"midpoint":      "https://company.tracker.com?operId=8&e=3&p=5890&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=&advertiser_id=&sURL=www.test.com&pfi=[PLATFORM]&af=video&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=&bidid=",
					"start":         "https://company.tracker.com?operId=8&e=2&p=5890&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=&advertiser_id=&sURL=www.test.com&pfi=[PLATFORM]&af=video&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=&bidid=",
					"thirdQuartile": "https://company.tracker.com?operId=8&e=5&p=5890&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=&advertiser_id=&sURL=www.test.com&pfi=[PLATFORM]&af=video&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=&bidid=",
				},
			},
		},
		{
			name: "all_macros with generated_bidId", // expect encoding for WRAPPER_IMPRESSION_ID macro
			args: args{
				trackerURL: "https://company.tracker.com?operId=8&e=[EVENT_ID]&p=[PBS-ACCOUNT]&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=[PBS-BIDDER]&advertiser_id=[ADVERTISER_NAME]&sURL=[DOMAIN]&pfi=[PLATFORM]&af=[ADTYPE]&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=[AD_UNIT]&bidid=[PBS-BIDID]&origbidid=[PBS-ORIG_BIDID]&bc=[BIDDER_CODE]",
				req: &openrtb2.BidRequest{
					App: &openrtb2.App{Bundle: "com.someapp.com", Publisher: &openrtb2.Publisher{ID: "5890"}},
					Ext: []byte(`{
						"prebid": {
								"macros": {
									"[PROFILE_ID]": "100",
									"[PROFILE_VERSION]": "2",
									"[UNIX_TIMESTAMP]": "1234567890",
									"[PLATFORM]": "7",
									"[WRAPPER_IMPRESSION_ID]": "abc~!@#$%^&&*()_+{}|:\"<>?[]\\;',./"
								}
						}
					}`),
					Imp: []openrtb2.Imp{
						{TagID: "/testadunit/1", ID: "imp_1"},
					},
				},
				bid:              &openrtb2.Bid{ADomain: []string{"http://a.com/32?k=v", "b.com"}, ImpID: "imp_1", ID: "test_bid_id"},
				gen_bidid:        "random_bid_id",
				requestingBidder: "test_bidder:234",
				bidderCoreName:   "test_core_bidder:234",
			},
			want: want{
				trackerURLMap: map[string]string{
					"firstQuartile": "https://company.tracker.com?operId=8&e=4&p=5890&pid=100&v=2&ts=1234567890&pn=test_core_bidder%3A234&advertiser_id=a.com&sURL=com.someapp.com&pfi=7&af=video&iid=abc~%21%40%23%24%25%5E%26%26%2A%28%29_%2B%7B%7D%7C%3A%22%3C%3E%3F%5B%5D%5C%3B%27%2C.%2F&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2Ftestadunit%2F1&bidid=random_bid_id&origbidid=test_bid_id&bc=test_bidder%3A234",
					"midpoint":      "https://company.tracker.com?operId=8&e=3&p=5890&pid=100&v=2&ts=1234567890&pn=test_core_bidder%3A234&advertiser_id=a.com&sURL=com.someapp.com&pfi=7&af=video&iid=abc~%21%40%23%24%25%5E%26%26%2A%28%29_%2B%7B%7D%7C%3A%22%3C%3E%3F%5B%5D%5C%3B%27%2C.%2F&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2Ftestadunit%2F1&bidid=random_bid_id&origbidid=test_bid_id&bc=test_bidder%3A234",
					"thirdQuartile": "https://company.tracker.com?operId=8&e=5&p=5890&pid=100&v=2&ts=1234567890&pn=test_core_bidder%3A234&advertiser_id=a.com&sURL=com.someapp.com&pfi=7&af=video&iid=abc~%21%40%23%24%25%5E%26%26%2A%28%29_%2B%7B%7D%7C%3A%22%3C%3E%3F%5B%5D%5C%3B%27%2C.%2F&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2Ftestadunit%2F1&bidid=random_bid_id&origbidid=test_bid_id&bc=test_bidder%3A234",
					"complete":      "https://company.tracker.com?operId=8&e=6&p=5890&pid=100&v=2&ts=1234567890&pn=test_core_bidder%3A234&advertiser_id=a.com&sURL=com.someapp.com&pfi=7&af=video&iid=abc~%21%40%23%24%25%5E%26%26%2A%28%29_%2B%7B%7D%7C%3A%22%3C%3E%3F%5B%5D%5C%3B%27%2C.%2F&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2Ftestadunit%2F1&bidid=random_bid_id&origbidid=test_bid_id&bc=test_bidder%3A234",
					"start":         "https://company.tracker.com?operId=8&e=2&p=5890&pid=100&v=2&ts=1234567890&pn=test_core_bidder%3A234&advertiser_id=a.com&sURL=com.someapp.com&pfi=7&af=video&iid=abc~%21%40%23%24%25%5E%26%26%2A%28%29_%2B%7B%7D%7C%3A%22%3C%3E%3F%5B%5D%5C%3B%27%2C.%2F&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2Ftestadunit%2F1&bidid=random_bid_id&origbidid=test_bid_id&bc=test_bidder%3A234"},
			},
		},
		{
			name: "all_macros with empty generated_bidId", // expect encoding for WRAPPER_IMPRESSION_ID macro
			args: args{
				trackerURL: "https://company.tracker.com?operId=8&e=[EVENT_ID]&p=[PBS-ACCOUNT]&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=[PBS-BIDDER]&advertiser_id=[ADVERTISER_NAME]&sURL=[DOMAIN]&pfi=[PLATFORM]&af=[ADTYPE]&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=[AD_UNIT]&bidid=[PBS-BIDID]&origbidid=[PBS-ORIG_BIDID]&bc=[BIDDER_CODE]",
				req: &openrtb2.BidRequest{
					App: &openrtb2.App{Bundle: "com.someapp.com", Publisher: &openrtb2.Publisher{ID: "5890"}},
					Ext: []byte(`{
						"prebid": {
								"macros": {
									"[PROFILE_ID]": "100",
									"[PROFILE_VERSION]": "2",
									"[UNIX_TIMESTAMP]": "1234567890",
									"[PLATFORM]": "7",
									"[WRAPPER_IMPRESSION_ID]": "abc~!@#$%^&&*()_+{}|:\"<>?[]\\;',./"
								}
						}
					}`),
					Imp: []openrtb2.Imp{
						{TagID: "/testadunit/1", ID: "imp_1"},
					},
				},
				bid:              &openrtb2.Bid{ADomain: []string{"http://a.com/32?k=v", "b.com"}, ImpID: "imp_1", ID: "test_bid_id"},
				gen_bidid:        "",
				requestingBidder: "test_bidder:234",
				bidderCoreName:   "test_core_bidder:234",
			},
			want: want{
				trackerURLMap: map[string]string{
					"firstQuartile": "https://company.tracker.com?operId=8&e=4&p=5890&pid=100&v=2&ts=1234567890&pn=test_core_bidder%3A234&advertiser_id=a.com&sURL=com.someapp.com&pfi=7&af=video&iid=abc~%21%40%23%24%25%5E%26%26%2A%28%29_%2B%7B%7D%7C%3A%22%3C%3E%3F%5B%5D%5C%3B%27%2C.%2F&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2Ftestadunit%2F1&bidid=test_bid_id&origbidid=test_bid_id&bc=test_bidder%3A234",
					"midpoint":      "https://company.tracker.com?operId=8&e=3&p=5890&pid=100&v=2&ts=1234567890&pn=test_core_bidder%3A234&advertiser_id=a.com&sURL=com.someapp.com&pfi=7&af=video&iid=abc~%21%40%23%24%25%5E%26%26%2A%28%29_%2B%7B%7D%7C%3A%22%3C%3E%3F%5B%5D%5C%3B%27%2C.%2F&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2Ftestadunit%2F1&bidid=test_bid_id&origbidid=test_bid_id&bc=test_bidder%3A234",
					"thirdQuartile": "https://company.tracker.com?operId=8&e=5&p=5890&pid=100&v=2&ts=1234567890&pn=test_core_bidder%3A234&advertiser_id=a.com&sURL=com.someapp.com&pfi=7&af=video&iid=abc~%21%40%23%24%25%5E%26%26%2A%28%29_%2B%7B%7D%7C%3A%22%3C%3E%3F%5B%5D%5C%3B%27%2C.%2F&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2Ftestadunit%2F1&bidid=test_bid_id&origbidid=test_bid_id&bc=test_bidder%3A234",
					"complete":      "https://company.tracker.com?operId=8&e=6&p=5890&pid=100&v=2&ts=1234567890&pn=test_core_bidder%3A234&advertiser_id=a.com&sURL=com.someapp.com&pfi=7&af=video&iid=abc~%21%40%23%24%25%5E%26%26%2A%28%29_%2B%7B%7D%7C%3A%22%3C%3E%3F%5B%5D%5C%3B%27%2C.%2F&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2Ftestadunit%2F1&bidid=test_bid_id&origbidid=test_bid_id&bc=test_bidder%3A234",
					"start":         "https://company.tracker.com?operId=8&e=2&p=5890&pid=100&v=2&ts=1234567890&pn=test_core_bidder%3A234&advertiser_id=a.com&sURL=com.someapp.com&pfi=7&af=video&iid=abc~%21%40%23%24%25%5E%26%26%2A%28%29_%2B%7B%7D%7C%3A%22%3C%3E%3F%5B%5D%5C%3B%27%2C.%2F&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2Ftestadunit%2F1&bidid=test_bid_id&origbidid=test_bid_id&bc=test_bidder%3A234"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			impMap := map[string]*openrtb2.Imp{}
			for _, imp := range tc.args.req.Imp {
				impMap[imp.ID] = &imp
			}

			eventURLMap := GetVideoEventTracking(tc.args.trackerURL, tc.args.bid, tc.args.gen_bidid, tc.args.requestingBidder, tc.args.bidderCoreName, tc.args.accountId, tc.args.timestamp, tc.args.req, impMap)
			assert.True(t, reflect.DeepEqual(tc.want.trackerURLMap, eventURLMap))
		})
	}
}

func TestExtractDomain(t *testing.T) {
	testCases := []struct {
		description    string
		url            string
		expectedDomain string
		expectedErr    error
	}{
		{description: "a.com", url: "a.com", expectedDomain: "a.com", expectedErr: nil},
		{description: "a.com/123", url: "a.com/123", expectedDomain: "a.com", expectedErr: nil},
		{description: "http://a.com/123", url: "http://a.com/123", expectedDomain: "a.com", expectedErr: nil},
		{description: "https://a.com/123", url: "https://a.com/123", expectedDomain: "a.com", expectedErr: nil},
		{description: "c.b.a.com", url: "c.b.a.com", expectedDomain: "c.b.a.com", expectedErr: nil},
		{description: "url_encoded_http://c.b.a.com", url: "http%3A%2F%2Fc.b.a.com", expectedDomain: "c.b.a.com", expectedErr: nil},
		{description: "url_encoded_with_www_http://c.b.a.com", url: "http%3A%2F%2Fwww.c.b.a.com", expectedDomain: "c.b.a.com", expectedErr: nil},
	}
	for _, test := range testCases {
		t.Run(test.description, func(t *testing.T) {
			domain, err := extractDomain(test.url)
			assert.Equal(t, test.expectedDomain, domain)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

// replaceMacros copied test cases from older replaceMacro(), will use gofuzzy once golang is upgraded
func Test_replaceMacros(t *testing.T) {
	type args struct {
		trackerURL string
		macroMap   map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty_tracker_url",
			args: args{
				trackerURL: "",
				macroMap: map[string]string{
					"[TEST]": "testme",
				},
			},
			want: "",
		},
		{
			name: "tracker_url_with_macro",
			args: args{
				trackerURL: "http://something.com?test=[TEST]",
				macroMap: map[string]string{
					"[TEST]": "testme",
				},
			},
			want: "http://something.com?test=testme",
		},
		{
			name: "tracker_url_with_invalid_macro",
			args: args{
				trackerURL: "http://something.com?test=TEST]",
				macroMap: map[string]string{
					"[TEST]": "testme",
				},
			},
			want: "http://something.com?test=TEST]",
		},
		{
			name: "tracker_url_with_repeating_macro",
			args: args{
				trackerURL: "http://something.com?test=[TEST]&test1=[TEST]",
				macroMap: map[string]string{
					"[TEST]": "testme",
				},
			},
			want: "http://something.com?test=testme&test1=testme",
		},
		{
			name: "empty_macro",
			args: args{
				trackerURL: "http://something.com?test=[TEST]",
				macroMap: map[string]string{
					"": "testme",
				},
			},
			want: "http://something.com?test=[TEST]",
		},
		{
			name: "macro_without_[",
			args: args{
				trackerURL: "http://something.com?test=[TEST]",
				macroMap: map[string]string{
					"TEST]": "testme",
				},
			},
			want: "http://something.com?test=[TEST]",
		},
		{
			name: "macro_without_]",
			args: args{
				trackerURL: "http://something.com?test=[TEST]",
				macroMap: map[string]string{
					"[TEST": "testme",
				},
			},
			want: "http://something.com?test=[TEST]",
		},
		{
			name: "empty_value",
			args: args{
				trackerURL: "http://something.com?test=[TEST]",
				macroMap: map[string]string{
					"[TEST]": ""},
			},
			want: "http://something.com?test=",
		},
		{
			name: "nested_macro_value",
			args: args{
				trackerURL: "http://something.com?test=[TEST]",
				macroMap: map[string]string{
					"[TEST]": "[TEST][TEST]",
				},
			},
			want: "http://something.com?test=%5BTEST%5D%5BTEST%5D",
		},
		{
			name: "url_as_macro_value",
			args: args{
				trackerURL: "http://something.com?test=[TEST]",
				macroMap: map[string]string{
					"[TEST]": "http://iamurl.com",
				},
			},
			want: "http://something.com?test=http%3A%2F%2Fiamurl.com",
		},
		// { Moved this responsiblity to GetVideoEventTracking()
		// 	name: "macro_with_spaces",
		// 	args: args{
		// 		trackerURL: "http://something.com?test=[TEST]",
		// 		macroMap: map[string]string{
		// 			"  [TEST]  ": "http://iamurl.com",
		// 		},
		// 	},
		// 	want: "http://something.com?test=http%3A%2F%2Fiamurl.com",
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceMacros(tt.args.trackerURL, tt.args.macroMap)
			assert.Equal(t, tt.want, got)
		})
	}
}

// BenchmarkGetVideoEventTracking
//
// Original:
// Running tool: /usr/local/go/bin/go test -benchmem -run=^$ -bench ^BenchmarkGetVideoEventTracking$ github.com/PubMatic-OpenWrap/prebid-server/endpoints/events

// goos: linux
// goarch: arm64
// pkg: github.com/PubMatic-OpenWrap/prebid-server/endpoints/events
// BenchmarkGetVideoEventTracking-8   	   19048	     78882 ns/op	   31590 B/op	     128 allocs/op
// BenchmarkGetVideoEventTracking-8   	   27333	     40491 ns/op	   31589 B/op	     128 allocs/op
// BenchmarkGetVideoEventTracking-8   	   28392	     45111 ns/op	   31586 B/op	     128 allocs/op
// BenchmarkGetVideoEventTracking-8   	   18160	     83581 ns/op	   31585 B/op	     128 allocs/op
// BenchmarkGetVideoEventTracking-8   	   16633	     77993 ns/op	   31591 B/op	     128 allocs/op
// PASS
// ok  	github.com/PubMatic-OpenWrap/prebid-server/endpoints/events	1.807s

// Refactored-GetVideoEventTracking:
// BenchmarkGetVideoEventTracking-8   	   10000	    108697 ns/op	   33489 B/op	     131 allocs/op
// BenchmarkGetVideoEventTracking-8   	   10000	    115349 ns/op	   33489 B/op	     131 allocs/op
// BenchmarkGetVideoEventTracking-8   	   12678	     80833 ns/op	   33486 B/op	     131 allocs/op
// BenchmarkGetVideoEventTracking-8   	   18840	     60841 ns/op	   33493 B/op	     131 allocs/op
// BenchmarkGetVideoEventTracking-8   	   20086	     57733 ns/op	   33482 B/op	     131 allocs/op

// Refactored-GetVideoEventTracking-using-replaceMacros:
// BenchmarkGetVideoEventTracking-8   	   65928	     16866 ns/op	   10434 B/op	      96 allocs/op
// BenchmarkGetVideoEventTracking-8   	   66710	     18611 ns/op	   10433 B/op	      96 allocs/op
// BenchmarkGetVideoEventTracking-8   	   66448	     17244 ns/op	   10433 B/op	      96 allocs/op
// BenchmarkGetVideoEventTracking-8   	   35112	     38085 ns/op	   10433 B/op	      96 allocs/op
// BenchmarkGetVideoEventTracking-8   	   40941	     27584 ns/op	   10434 B/op	      96 allocs/op
func BenchmarkGetVideoEventTracking(b *testing.B) {
	//  all_macros with generated_bidId
	trackerURL := "https://company.tracker.com?operId=8&e=[EVENT_ID]&p=[PBS-ACCOUNT]&pid=[PROFILE_ID]&v=[PROFILE_VERSION]&ts=[UNIX_TIMESTAMP]&pn=[PBS-BIDDER]&advertiser_id=[ADVERTISER_NAME]&sURL=[DOMAIN]&pfi=[PLATFORM]&af=[ADTYPE]&iid=[WRAPPER_IMPRESSION_ID]&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=[AD_UNIT]&bidid=[PBS-BIDID]&origbidid=[PBS-ORIG_BIDID]&bc=[BIDDER_CODE]"
	req := &openrtb2.BidRequest{
		App: &openrtb2.App{Bundle: "com.someapp.com", Publisher: &openrtb2.Publisher{ID: "5890"}},
		Ext: []byte(`{
				"prebid": {
						"macros": {
							"[PROFILE_ID]": "100",
							"[PROFILE_VERSION]": "2",
							"[UNIX_TIMESTAMP]": "1234567890",
							"[PLATFORM]": "7",
							"[WRAPPER_IMPRESSION_ID]": "abc~!@#$%^&&*()_+{}|:\"<>?[]\\;',./"
						}
				}
			}`),
		Imp: []openrtb2.Imp{
			{TagID: "/testadunit/1", ID: "imp_1"},
		},
	}
	bid := &openrtb2.Bid{ADomain: []string{"http://a.com/32?k=v", "b.com"}, ImpID: "imp_1", ID: "test_bid_id"}
	gen_bidid := "random_bid_id"
	requestingBidder := "test_bidder:234"
	bidderCoreName := "test_core_bidder:234"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetVideoEventTracking(trackerURL, bid, gen_bidid, requestingBidder, bidderCoreName, "", 0, req, nil)
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// no need to validate, using vast from test TestInjectVideoEventTrackers/all_inline_wrapper_liner_and_non_linear_creative
		_, _ = InjectVideoEventTrackers(trackerURL, vastXML, bid, "", "test_bidder", "test_core_bidder", "", int64(0), req)
	}
}
