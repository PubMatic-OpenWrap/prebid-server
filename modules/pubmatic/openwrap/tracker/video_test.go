package tracker

import (
	"fmt"
	"testing"

	"github.com/beevik/etree"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestInjectVideoCreativeTrackers(t *testing.T) {
	type args struct {
		bid                     openrtb2.Bid
		videoParams             []models.OWTracker
		injectImpressionTracker bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "empty_bid",
			args: args{
				bid: openrtb2.Bid{},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracking URL`,
						ErrorURL:   `Error URL`,
					},
				},
				injectImpressionTracker: true,
			},
			want:    ``,
			wantErr: true,
		},
		{
			name: "nil_bid.adm",
			args: args{

				bid: openrtb2.Bid{},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracking URL`,
						ErrorURL:   `Error URL`,
					},
				},
				injectImpressionTracker: true,
			},
			want:    ``,
			wantErr: true,
		},
		{
			name: "empty_bid.adm",
			args: args{

				bid: openrtb2.Bid{
					AdM: ``,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracking URL`,
						ErrorURL:   `Error URL`,
					},
				},
				injectImpressionTracker: true,
			},
			want:    ``,
			wantErr: true,
		},
		{
			name: "empty_bid.adm.partner_url",
			args: args{

				bid: openrtb2.Bid{
					AdM: `https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracking URL`,
						ErrorURL:   `Error URL`,
					},
				},
				injectImpressionTracker: true,
			},
			want:    `<VAST version="3.0"><Ad id="1"><Wrapper><AdSystem>PubMatic Wrapper</AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error></Wrapper></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "empty_vast_params",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[https://t.pubmatic.com/wt?pubid=156021&purl=https%3A%2F%2Fwww.financialexpress.com%2F&tst=1533048550&iid=9fc7e570-9b01-4cfb-8381-0dc1eec16d27-dieuf&bidid=36110151cad7636&pid=116&pdvid=24&slot=div-gpt-ad-1478842567868-0&pn=rubicon&en=0.02&eg=0.02&kgpv=300x250%40300x250]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[http://172.16.4.213/track]]></Error><Error><![CDATA[https://Errortrack.com]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]]></MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				videoParams:             []models.OWTracker{},
				injectImpressionTracker: true,
			},
			want:    ``,
			wantErr: true,
		},
		{
			name: "invalid_vast",
			args: args{

				bid: openrtb2.Bid{
					AdM: `invalid_vast_creative`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracking URL`,
						ErrorURL:   `Error URL`,
					},
				},
				injectImpressionTracker: true,
			},
			want:    `invalid_vast_creative`,
			wantErr: true,
		},
		{
			name: "no_vast_ad",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST><Ad></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracking URL`,
						ErrorURL:   `Error URL`,
					},
				},
				injectImpressionTracker: true,
			},
			want:    `<VAST><Ad></Ad></VAST>`,
			wantErr: true,
		},
		{
			name: "vast_2.0_inline_pricing",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST><Ad><InLine></InLine></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracking URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				injectImpressionTracker: true,
			},
			want:    `<VAST><Ad><InLine><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error><Extensions><Extension><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></Extension></Extensions></InLine></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "vast_3.0_inline_pricing",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST version="3.0"><Ad><InLine></InLine></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracking URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				injectImpressionTracker: true,
			},
			want:    `<VAST version="3.0"><Ad><InLine><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></InLine></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "inline_vast_3.0",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST version='3.0'><Ad id='601364'><InLine><AdSystem>Acudeo Compatible</AdSystem><AdTitle>VAST 2.0 Instream Test 1</AdTitle><Description>VAST 2.0 Instream Test 1</Description><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet?operId=1&pubId=5890&siteId=47163&adId=1405268&adType=13&adServerId=243&kefact=70.000000&kaxefact=70.000000&kadNetFrequecy=0&kadwidth=0&kadheight=0&kadsizeid=97&kltstamp=1529929473&indirectAdId=0&adServerOptimizerId=2&ranreq=0.1&kpbmtpfact=100.000000&dcId=1&tldId=0&passback=0&svr=MADS1107&ekefact=Ad8wW91TCwCmdG0jlfjXn7Tyzh20hnTVx-m5DoNSep-RXGDr&ekaxefact=Ad8wWwRUCwAGir4Zzl1eF0bKiC-qrCV0D0yp_eE7YizB_BQk&ekpbmtpfact=Ad8wWxRUCwD7qgzwwPE2LnS5-Ou19uO5amJl1YT6-XVFvQ41&imprId=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&oid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&crID=creative-1_1_2&ucrid=160175026529250297&campaignId=17050&creativeId=0&pctr=0.000000&wDSPByrId=511&wDspId=6&wbId=0&wrId=0&wAdvID=3170&isRTB=1&rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&pmZoneId=zone1&pageURL=www.yahoo.com&lpu=ae.com]]></Impression><Impression>https://dsptracker.com/{PSPM}</Impression><Error><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&er=[ERRORCODE]]]></Error><Error><![CDATA[https://Errortrack.com?p=1234&er=[ERRORCODE]]]></Error><Creatives><Creative AdID='601364'><Linear skipoffset='20%'><Duration>00:00:04</Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=99]]></ClickTracking><ClickThrough>https://www.sample.com</ClickThrough></VideoClicks><MediaFiles><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]></MediaFile><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				injectImpressionTracker: true,
			},
			want:    `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[Tracker URL]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet?operId=1&pubId=5890&siteId=47163&adId=1405268&adType=13&adServerId=243&kefact=70.000000&kaxefact=70.000000&kadNetFrequecy=0&kadwidth=0&kadheight=0&kadsizeid=97&kltstamp=1529929473&indirectAdId=0&adServerOptimizerId=2&ranreq=0.1&kpbmtpfact=100.000000&dcId=1&tldId=0&passback=0&svr=MADS1107&ekefact=Ad8wW91TCwCmdG0jlfjXn7Tyzh20hnTVx-m5DoNSep-RXGDr&ekaxefact=Ad8wWwRUCwAGir4Zzl1eF0bKiC-qrCV0D0yp_eE7YizB_BQk&ekpbmtpfact=Ad8wWxRUCwD7qgzwwPE2LnS5-Ou19uO5amJl1YT6-XVFvQ41&imprId=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&oid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&crID=creative-1_1_2&ucrid=160175026529250297&campaignId=17050&creativeId=0&pctr=0.000000&wDSPByrId=511&wDspId=6&wbId=0&wrId=0&wAdvID=3170&isRTB=1&rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&pmZoneId=zone1&pageURL=www.yahoo.com&lpu=ae.com]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&er=[ERRORCODE]]]></Error><Error><![CDATA[https://Errortrack.com?p=1234&er=[ERRORCODE]]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=99]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></InLine></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "wrapper_vast_2.0",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST version='2.0'><Ad id='123'><Wrapper><AdSystem>DSP</AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Error><![CDATA[https://track.dsp.com/er=[ERRORCODE]/tracker/error]]></Error><Impression><![CDATA[https://track.dsp.com?e=impression]]></Impression><Creatives><Creative><Linear><TrackingEvents></TrackingEvents><VideoClicks><ClickTracking><![CDATA[http://track.dsp.com/tracker/click]]></ClickTracking></VideoClicks></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				injectImpressionTracker: true,
			},
			want:    `<VAST version="2.0"><Ad id="123"><Wrapper><AdSystem><![CDATA[DSP]]></AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Error><![CDATA[Error URL]]></Error><Error><![CDATA[https://track.dsp.com/er=[ERRORCODE]/tracker/error]]></Error><Impression><![CDATA[Tracker URL]]></Impression><Impression><![CDATA[https://track.dsp.com?e=impression]]></Impression><Creatives><Creative><Linear><TrackingEvents/><VideoClicks><ClickTracking><![CDATA[http://track.dsp.com/tracker/click]]></ClickTracking></VideoClicks></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "inline_vast_with_no_cdata",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST version='3.0'><Ad id='601364'><InLine><AdSystem>Acudeo Compatible</AdSystem><AdTitle>VAST 2.0 Instream Test 1</AdTitle><Description>VAST 2.0 Instream Test 1</Description><Impression>http://172.16.4.213/AdServer/AdDisplayTrackerServlet</Impression><Impression>https://dsptracker.com/{PSPM}</Impression><Error>http://172.16.4.213/track</Error><Error>https://Errortrack.com</Error><Creatives><Creative AdID='601364'><Linear skipoffset='20%'><Duration>00:00:04</Duration><VideoClicks><ClickTracking>http://172.16.4.213/track</ClickTracking><ClickThrough>https://www.sample.com</ClickThrough></VideoClicks><MediaFiles><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'>https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]</MediaFile><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'>https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]</MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				injectImpressionTracker: true,
			},
			want:    `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[Tracker URL]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://172.16.4.213/track]]></Error><Error><![CDATA[https://Errortrack.com]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]]></MediaFile></MediaFiles></Linear></Creative></Creatives><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></InLine></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "wrapper_vast_with_no_cdata",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST version='2.0'><Ad id='123'><Wrapper><AdSystem>DSP</AdSystem><VASTAdTagURI>https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml</VASTAdTagURI><Error>https://track.dsp.com/er=[ERRORCODE]/tracker/error</Error><Impression>https://track.dsp.com?e=impression</Impression><Creatives><Creative><Linear><TrackingEvents></TrackingEvents><VideoClicks><ClickTracking>http://track.dsp.com/tracker/click</ClickTracking></VideoClicks></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				injectImpressionTracker: true,
			},
			want:    `<VAST version="2.0"><Ad id="123"><Wrapper><AdSystem><![CDATA[DSP]]></AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Error><![CDATA[Error URL]]></Error><Error><![CDATA[https://track.dsp.com/er=[ERRORCODE]/tracker/error]]></Error><Impression><![CDATA[Tracker URL]]></Impression><Impression><![CDATA[https://track.dsp.com?e=impression]]></Impression><Creatives><Creative><Linear><TrackingEvents/><VideoClicks><ClickTracking><![CDATA[http://track.dsp.com/tracker/click]]></ClickTracking></VideoClicks></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "spaces_in_creative_TET-8024",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST version="3.0">   <Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[https://t.pubmatic.com/wt?pubid=156021&purl=https%3A%2F%2Fwww.financialexpress.com%2F&tst=1533048550&iid=9fc7e570-9b01-4cfb-8381-0dc1eec16d27-dieuf&bidid=36110151cad7636&pid=116&pdvid=24&slot=div-gpt-ad-1478842567868-0&pn=rubicon&en=0.02&eg=0.02&kgpv=300x250%40300x250]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[http://172.16.4.213/track]]></Error><Error><![CDATA[https://Errortrack.com]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]]></MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracking URL`,
						ErrorURL:   `Error URL`,
					},
				},
				injectImpressionTracker: true,
			},
			want:    `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[Tracking URL]]></Impression><Impression><![CDATA[https://t.pubmatic.com/wt?pubid=156021&purl=https%3A%2F%2Fwww.financialexpress.com%2F&tst=1533048550&iid=9fc7e570-9b01-4cfb-8381-0dc1eec16d27-dieuf&bidid=36110151cad7636&pid=116&pdvid=24&slot=div-gpt-ad-1478842567868-0&pn=rubicon&en=0.02&eg=0.02&kgpv=300x250%40300x250]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://172.16.4.213/track]]></Error><Error><![CDATA[https://Errortrack.com]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]]></MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "inline_vast_3.0_with_inject_impression_tracker_false",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST version='3.0'><Ad id='601364'><InLine><AdSystem>Acudeo Compatible</AdSystem><AdTitle>VAST 2.0 Instream Test 1</AdTitle><Description>VAST 2.0 Instream Test 1</Description><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet?operId=1&pubId=5890&siteId=47163&adId=1405268&adType=13&adServerId=243&kefact=70.000000&kaxefact=70.000000&kadNetFrequecy=0&kadwidth=0&kadheight=0&kadsizeid=97&kltstamp=1529929473&indirectAdId=0&adServerOptimizerId=2&ranreq=0.1&kpbmtpfact=100.000000&dcId=1&tldId=0&passback=0&svr=MADS1107&ekefact=Ad8wW91TCwCmdG0jlfjXn7Tyzh20hnTVx-m5DoNSep-RXGDr&ekaxefact=Ad8wWwRUCwAGir4Zzl1eF0bKiC-qrCV0D0yp_eE7YizB_BQk&ekpbmtpfact=Ad8wWxRUCwD7qgzwwPE2LnS5-Ou19uO5amJl1YT6-XVFvQ41&imprId=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&oid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&crID=creative-1_1_2&ucrid=160175026529250297&campaignId=17050&creativeId=0&pctr=0.000000&wDSPByrId=511&wDspId=6&wbId=0&wrId=0&wAdvID=3170&isRTB=1&rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&pmZoneId=zone1&pageURL=www.yahoo.com&lpu=ae.com]]></Impression><Impression>https://dsptracker.com/{PSPM}</Impression><Error><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&er=[ERRORCODE]]]></Error><Error><![CDATA[https://Errortrack.com?p=1234&er=[ERRORCODE]]]></Error><Creatives><Creative AdID='601364'><Linear skipoffset='20%'><Duration>00:00:04</Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=99]]></ClickTracking><ClickThrough>https://www.sample.com</ClickThrough></VideoClicks><MediaFiles><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]></MediaFile><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				injectImpressionTracker: false,
			},
			want:    `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet?operId=1&pubId=5890&siteId=47163&adId=1405268&adType=13&adServerId=243&kefact=70.000000&kaxefact=70.000000&kadNetFrequecy=0&kadwidth=0&kadheight=0&kadsizeid=97&kltstamp=1529929473&indirectAdId=0&adServerOptimizerId=2&ranreq=0.1&kpbmtpfact=100.000000&dcId=1&tldId=0&passback=0&svr=MADS1107&ekefact=Ad8wW91TCwCmdG0jlfjXn7Tyzh20hnTVx-m5DoNSep-RXGDr&ekaxefact=Ad8wWwRUCwAGir4Zzl1eF0bKiC-qrCV0D0yp_eE7YizB_BQk&ekpbmtpfact=Ad8wWxRUCwD7qgzwwPE2LnS5-Ou19uO5amJl1YT6-XVFvQ41&imprId=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&oid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&crID=creative-1_1_2&ucrid=160175026529250297&campaignId=17050&creativeId=0&pctr=0.000000&wDSPByrId=511&wDspId=6&wbId=0&wrId=0&wAdvID=3170&isRTB=1&rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&pmZoneId=zone1&pageURL=www.yahoo.com&lpu=ae.com]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&er=[ERRORCODE]]]></Error><Error><![CDATA[https://Errortrack.com?p=1234&er=[ERRORCODE]]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=99]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></InLine></Ad></VAST>`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := injectVideoCreativeTrackers(tt.args.bid, tt.args.videoParams, tt.args.injectImpressionTracker)
			if (err != nil) != tt.wantErr {
				t.Errorf("injectVideoCreativeTrackers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func getXMLDocument(tag string) *etree.Document {
	doc := etree.NewDocument()
	err := doc.ReadFromString(tag)
	if err != nil {
		return nil
	}
	return doc
}

func Test_injectPricingNodeVAST20(t *testing.T) {
	type args struct {
		doc      *etree.Document
		price    float64
		model    string
		currency string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "pricing_node_missing",
			args: args{
				doc:      getXMLDocument(`<Impressions/>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Impressions/><Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "extensions_present_pricing_node_missing",
			args: args{
				doc:      getXMLDocument(`<Extensions/>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "extension_present_pricing_node_missing",
			args: args{
				doc:      getXMLDocument(`<Extensions><Extension/></Extensions>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Extensions><Extension/><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "override_pricing_cpm",
			args: args{
				doc:      getXMLDocument(`<Impressions/><Extensions><Extension><Pricing model="CPM" currency="USD">1.23</Pricing></Extension></Extensions>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Impressions/><Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "override_pricing_cpm_add_currency",
			args: args{
				doc:      getXMLDocument(`<Impressions/><Extensions><Extension><Pricing model="CPM">1.23</Pricing></Extension></Extensions>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Impressions/><Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "override_pricing_cpm_add_attributes",
			args: args{
				doc:      getXMLDocument(`<Impressions/><Extensions><Extension><Pricing>1.23</Pricing></Extension></Extensions>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Impressions/><Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "override_pricing_node",
			args: args{
				doc:      getXMLDocument(`<Impressions/><Extensions><Extension><Pricing model="CPC" currency="INR">1.23</Pricing></Extension></Extensions>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Impressions/><Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injectPricingNodeVAST20(&tt.args.doc.Element, tt.args.price, tt.args.model, tt.args.currency)
			actual, _ := tt.args.doc.WriteToString()
			assert.Equal(t, tt.want, actual)
		})
	}
}

func Test_injectPricingNodeVAST3x(t *testing.T) {
	type args struct {
		doc      *etree.Document
		price    float64
		model    string
		currency string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "override_cpm_pricing",
			args: args{
				doc:      getXMLDocument(`<Pricing model="CPM" currency="USD">1.23</Pricing>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[4.5]]></Pricing>`,
		},
		{
			name: "override_cpc_pricing",
			args: args{
				doc:      getXMLDocument(`<Pricing model="CPC" currency="INR">1.23</Pricing>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[4.5]]></Pricing>`,
		},
		{
			name: "add_currency",
			args: args{
				doc:      getXMLDocument(`<Pricing model="CPM">1.23</Pricing>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[4.5]]></Pricing>`,
		},
		{
			name: "add_all_attributes",
			args: args{
				doc:      getXMLDocument(`<Pricing>1.23</Pricing>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[4.5]]></Pricing>`,
		},
		{
			name: "pricing_node_not_present",
			args: args{
				doc:      getXMLDocument(`<Impressions></Impressions>`),
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: models.VideoPricingCurrencyUSD,
			},
			want: `<Impressions/><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injectPricingNodeVAST3x(&tt.args.doc.Element, tt.args.price, tt.args.model, tt.args.currency)
			actual, _ := tt.args.doc.WriteToString()
			assert.Equal(t, tt.want, actual)
		})
	}
}

func Test_updatePricingNode(t *testing.T) {
	type args struct {
		doc      *etree.Document
		price    float64
		model    string
		currency string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "overwrite_existing_price",
			args: args{
				doc:      getXMLDocument(`<Pricing>4.5</Pricing>`),
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: models.VideoPricingCurrencyUSD,
			},
			want: `<Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "empty_attributes",
			args: args{
				doc:      getXMLDocument(`<Pricing>4.5</Pricing>`),
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: models.VideoPricingCurrencyUSD,
			},
			want: `<Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "overwrite_model",
			args: args{
				doc:      getXMLDocument(`<Pricing model="CPM">4.5</Pricing>`),
				price:    1.2,
				model:    "CPC",
				currency: models.VideoPricingCurrencyUSD,
			},
			want: `<Pricing model="CPC" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "overwrite_currency",
			args: args{
				doc:      getXMLDocument(`<Pricing currency="USD">4.5</Pricing>`),
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: "INR",
			},
			want: `<Pricing currency="INR" model="` + models.VideoPricingModelCPM + `"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "default_values_attribute",
			args: args{
				doc:      getXMLDocument(`<Pricing>4.5</Pricing>`),
				price:    1.2,
				model:    "",
				currency: "",
			},
			want: `<Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[1.2]]></Pricing>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatePricingNode(tt.args.doc.ChildElements()[0], tt.args.price, tt.args.model, tt.args.currency)
			actual, _ := tt.args.doc.WriteToString()
			assert.Equal(t, tt.want, actual)
		})
	}
}

func Test_newPricingNode(t *testing.T) {
	type args struct {
		price    float64
		model    string
		currency string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "node",
			args: args{
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: models.VideoPricingCurrencyUSD,
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "empty_currency",
			args: args{
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: "",
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "other_currency",
			args: args{
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: `INR`,
			},
			want: `<Pricing model="CPM" currency="INR"><![CDATA[1.2]]></Pricing>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newPricingNode(tt.args.price, tt.args.model, tt.args.currency)
			doc := etree.NewDocument()
			doc.InsertChild(nil, got)
			actual, _ := doc.WriteToString()
			assert.Equal(t, tt.want, actual)
		})
	}
}
