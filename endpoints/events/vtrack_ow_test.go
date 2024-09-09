package events

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
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
		wantErr     error
		metrics     *openrtb_ext.FastXMLMetrics
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     errEventURLNotConfigured,
				metrics:     nil,
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
				wantErr:     nil,
				metrics:     &openrtb_ext.FastXMLMetrics{IsRespMismatch: false},
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
			injectedVast, metrics, err := InjectVideoEventTrackers(tc.args.req, tc.args.bid, tc.args.vastXML, tc.args.externalURL, tc.args.genbidID, "test_bidder", "test_core_bidder", int64(0), true)
			assert.Equal(t, tc.want.wantErr, err)
			assert.Equal(t, tc.want.wantVastXml, string(injectedVast))
			if tc.want.metrics != nil {
				assert.NotNil(t, metrics)
				assert.Equal(t, tc.want.metrics.IsRespMismatch, metrics.IsRespMismatch)
			}
		})
	}
}

func quoteUnescape[T []byte | string](s T) string {
	buf := bytes.Buffer{}
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '\\' {
			if i+1 < len(s) {
				nextCh := s[i+1]
				if nextCh == '\\' || nextCh == '"' || nextCh == '\'' {
					i++
					ch = nextCh
				}
			}
		}
		buf.WriteByte(ch)
	}
	return buf.String()
}

func TestCompareXMLParsers(t *testing.T) {
	//fileName := `./test/base64_vast.txt`
	fileName := `./test/raw_vast.txt`

	base64Decode := strings.Contains(fileName, "base64")

	file, err := os.Open(fileName)
	if !assert.Nil(t, err) {
		return
	}

	defer file.Close()
	var mismatched []int
	var debugLines map[int]bool
	line := 0
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	//debugLines = map[int]bool{29: true, 30: true, 33: true, 50: true, 93: true}

	for scanner.Scan() {
		line++
		vast := scanner.Text()
		if len(debugLines) > 0 {
			if debugLines[line] == false {
				continue
			}
		}
		if base64Decode {
			data, err := base64.StdEncoding.DecodeString(vast)
			if !assert.Nil(t, err) {
				continue
			}
			vast = quoteUnescape(data)
		}
		t.Run(fmt.Sprintf("vast_%d", line), func(t *testing.T) {
			etreeXML, _ := injectVideoEventsETree(vast, eventURLMap, false, adcom1.LinearityLinear)
			fastXML, _ := injectVideoEventsFastXML(vast, eventURLMap, false, adcom1.LinearityLinear)
			if vast != fastXML {
				//replace only if trackers are injected
				fastXML = strings.ReplaceAll(fastXML, " >", ">")
			}

			if !assert.Equal(t, etreeXML, fastXML) {
				mismatched = append(mismatched, line)
			}
		})
	}
	t.Logf("\n total:[%v] mismatched:[%v] lines:[%v]", line, len(mismatched), mismatched)
	assert.Equal(t, 0, len(mismatched))
	assert.Nil(t, scanner.Err())
}

func TestCompare(t *testing.T) {
	vastBytes, err := os.ReadFile(`./test/vast.txt`)
	assert.NoError(t, err)

	vast := string(vastBytes)
	etreeXML, err := injectVideoEventsETree(vast, eventURLMap, false, adcom1.LinearityLinear)
	assert.NoError(t, err)

	fastXML, err := injectVideoEventsFastXML(vast, eventURLMap, false, adcom1.LinearityLinear)
	assert.NoError(t, err)

	if vast != fastXML {
		//replace only if trackers are injected
		fastXML = strings.ReplaceAll(fastXML, " >", ">")
	}
	assert.Equal(t, etreeXML, fastXML)
}
