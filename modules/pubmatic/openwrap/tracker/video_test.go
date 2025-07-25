package tracker

import (
	"testing"

	"github.com/beevik/etree"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func getXMLDocument(tag string) *etree.Document {
	doc := etree.NewDocument()
	err := doc.ReadFromString(tag)
	if err != nil {
		return nil
	}
	return doc
}

func TestInjectVideoCreativeTrackers(t *testing.T) {
	type args struct {
		bid         openrtb2.Bid
		videoParams []models.OWTracker
		rctx        models.RequestCtx
	}
	tests := []struct {
		name     string
		args     args
		wantAdm  string
		wantBurl string
		wantErr  bool
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
			},
			wantAdm: ``,
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
			},
			wantAdm: ``,
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
			},
			wantAdm: ``,
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
			},
			wantAdm: `<VAST version="3.0"><Ad id="1"><Wrapper><AdSystem>PubMatic Wrapper</AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error></Wrapper></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "empty_vast_params",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[https://t.pubmatic.com/wt?pubid=156021&purl=https%3A%2F%2Fwww.financialexpress.com%2F&tst=1533048550&iid=9fc7e570-9b01-4cfb-8381-0dc1eec16d27-dieuf&bidid=36110151cad7636&pid=116&pdvid=24&slot=div-gpt-ad-1478842567868-0&pn=rubicon&en=0.02&eg=0.02&kgpv=300x250%40300x250]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[http://172.16.4.213/track]]></Error><Error><![CDATA[https://Errortrack.com]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]]></MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{},
			},
			wantAdm: ``,
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
			},
			wantAdm: `invalid_vast_creative`,
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
			},
			wantAdm: `<VAST><Ad></Ad></VAST>`,
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
			},
			wantAdm: `<VAST><Ad><InLine><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error><Extensions><Extension><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></Extension></Extensions></InLine></Ad></VAST>`,
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
			},
			wantAdm: `<VAST version="3.0"><Ad><InLine><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></InLine></Ad></VAST>`,
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
			},
			wantAdm: `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[Tracker URL]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet?operId=1&pubId=5890&siteId=47163&adId=1405268&adType=13&adServerId=243&kefact=70.000000&kaxefact=70.000000&kadNetFrequecy=0&kadwidth=0&kadheight=0&kadsizeid=97&kltstamp=1529929473&indirectAdId=0&adServerOptimizerId=2&ranreq=0.1&kpbmtpfact=100.000000&dcId=1&tldId=0&passback=0&svr=MADS1107&ekefact=Ad8wW91TCwCmdG0jlfjXn7Tyzh20hnTVx-m5DoNSep-RXGDr&ekaxefact=Ad8wWwRUCwAGir4Zzl1eF0bKiC-qrCV0D0yp_eE7YizB_BQk&ekpbmtpfact=Ad8wWxRUCwD7qgzwwPE2LnS5-Ou19uO5amJl1YT6-XVFvQ41&imprId=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&oid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&crID=creative-1_1_2&ucrid=160175026529250297&campaignId=17050&creativeId=0&pctr=0.000000&wDSPByrId=511&wDspId=6&wbId=0&wrId=0&wAdvID=3170&isRTB=1&rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&pmZoneId=zone1&pageURL=www.yahoo.com&lpu=ae.com]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&er=[ERRORCODE]]]></Error><Error><![CDATA[https://Errortrack.com?p=1234&er=[ERRORCODE]]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=99]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></InLine></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "wrapper_vast_4.0",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST version='4.0' xmlns='http://www.iab.com/VAST'><Ad id='20011' sequence='1' conditionalAd='false'><Wrapper followAdditionalWrappers='0' allowMultipleAds='1' fallbackOnNoAd='0'><AdSystem version='4.0'>iabtechlab</AdSystem><Error>http://example.com/error</Error><Impression id='Impression-ID'>http://example.com/track/impression</Impression><Creatives><Creative id='5480' sequence='1' adId='2447226'><CompanionAds><Companion id='1232' width='100' height='150' assetWidth='250' assetHeight='200' expandedWidth='350' expandedHeight='250' apiFramework='VPAID' adSlotID='3214' pxratio='1400'><StaticResource creativeType='image/png'><![CDATA[https://www.iab.com/wp-content/uploads/2014/09/iab-tech-lab-6-644x290.png]]></StaticResource><CompanionClickThrough><![CDATA[https://iabtechlab.com]]></CompanionClickThrough></Companion></CompanionAds></Creative></Creatives><VASTAdTagURI><![CDATA[https://raw.githubusercontent.com/InteractiveAdvertisingBureau/VAST_Samples/master/VAST%204.0%20Samples/Inline_Companion_Tag-test.xml]]></VASTAdTagURI></Wrapper></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
			},
			wantAdm: `<VAST version="4.0" xmlns="http://www.iab.com/VAST"><Ad id="20011" sequence="1" conditionalAd="false"><Wrapper followAdditionalWrappers="0" allowMultipleAds="1" fallbackOnNoAd="0"><AdSystem version="4.0"><![CDATA[iabtechlab]]></AdSystem><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://example.com/error]]></Error><Impression><![CDATA[Tracker URL]]></Impression><Impression id="Impression-ID"><![CDATA[http://example.com/track/impression]]></Impression><Creatives><Creative id="5480" sequence="1" adId="2447226"><CompanionAds><Companion id="1232" width="100" height="150" assetWidth="250" assetHeight="200" expandedWidth="350" expandedHeight="250" apiFramework="VPAID" adSlotID="3214" pxratio="1400"><StaticResource creativeType="image/png"><![CDATA[https://www.iab.com/wp-content/uploads/2014/09/iab-tech-lab-6-644x290.png]]></StaticResource><CompanionClickThrough><![CDATA[https://iabtechlab.com]]></CompanionClickThrough></Companion></CompanionAds></Creative></Creatives><VASTAdTagURI><![CDATA[https://raw.githubusercontent.com/InteractiveAdvertisingBureau/VAST_Samples/master/VAST%204.0%20Samples/Inline_Companion_Tag-test.xml]]></VASTAdTagURI><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></Wrapper></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "inline_vast_4.0",
			args: args{

				bid: openrtb2.Bid{
					AdM: `<VAST version="4.0" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST"><Ad id="20008" sequence="1" conditionalAd="false"><InLine><AdSystem version="4.0">iabtechlab</AdSystem><Error>http://example.com/error</Error><Extensions><Extension type="iab-Count"><total_available><![CDATA[ 2 ]]></total_available></Extension></Extensions><Impression id="Impression-ID">http://example.com/track/impression</Impression><Pricing model="cpm" currency="USD"><![CDATA[ 25.00 ]]></Pricing><AdTitle>iabtechlab video ad</AdTitle><Category authority="http://www.iabtechlab.com/categoryauthority">AD CONTENT description category</Category><Creatives><Creative id="5480" sequence="1" adId="2447226"><UniversalAdId idRegistry="Ad-ID" idValue="8465">8465</UniversalAdId><Linear><TrackingEvents><Tracking event="start">http://example.com/tracking/start</Tracking><Tracking event="firstQuartile">http://example.com/tracking/firstQuartile</Tracking><Tracking event="midpoint">http://example.com/tracking/midpoint</Tracking><Tracking event="thirdQuartile">http://example.com/tracking/thirdQuartile</Tracking><Tracking event="complete">http://example.com/tracking/complete</Tracking><Tracking event="progress" offset="00:00:10">http://example.com/tracking/progress-10</Tracking></TrackingEvents><Duration>00:00:16</Duration><MediaFiles><MediaFile id="5241" delivery="progressive" type="video/mp4" bitrate="2000" width="1280" height="720" minBitrate="1500" maxBitrate="2500" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro.mp4]]></MediaFile><MediaFile id="5244" delivery="progressive" type="video/mp4" bitrate="1000" width="854" height="480" minBitrate="700" maxBitrate="1500" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro-mid-resolution.mp4]]></MediaFile><MediaFile id="5246" delivery="progressive" type="video/mp4" bitrate="600" width="640" height="360" minBitrate="500" maxBitrate="700" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro-low-resolution.mp4]]></MediaFile></MediaFiles><VideoClicks><ClickThrough id="blog"><![CDATA[https://iabtechlab.com]]></ClickThrough></VideoClicks></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
			},
			wantAdm: `<VAST version="4.0" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST"><Ad id="20008" sequence="1" conditionalAd="false"><InLine><AdSystem version="4.0"><![CDATA[iabtechlab]]></AdSystem><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://example.com/error]]></Error><Extensions><Extension type="iab-Count"><total_available><![CDATA[2]]></total_available></Extension></Extensions><Impression><![CDATA[Tracker URL]]></Impression><Impression id="Impression-ID"><![CDATA[http://example.com/track/impression]]></Impression><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing><AdTitle><![CDATA[iabtechlab video ad]]></AdTitle><Category authority="http://www.iabtechlab.com/categoryauthority"><![CDATA[AD CONTENT description category]]></Category><Creatives><Creative id="5480" sequence="1" adId="2447226"><UniversalAdId idRegistry="Ad-ID" idValue="8465"><![CDATA[8465]]></UniversalAdId><Linear><TrackingEvents><Tracking event="start"><![CDATA[http://example.com/tracking/start]]></Tracking><Tracking event="firstQuartile"><![CDATA[http://example.com/tracking/firstQuartile]]></Tracking><Tracking event="midpoint"><![CDATA[http://example.com/tracking/midpoint]]></Tracking><Tracking event="thirdQuartile"><![CDATA[http://example.com/tracking/thirdQuartile]]></Tracking><Tracking event="complete"><![CDATA[http://example.com/tracking/complete]]></Tracking><Tracking event="progress" offset="00:00:10"><![CDATA[http://example.com/tracking/progress-10]]></Tracking></TrackingEvents><Duration><![CDATA[00:00:16]]></Duration><MediaFiles><MediaFile id="5241" delivery="progressive" type="video/mp4" bitrate="2000" width="1280" height="720" minBitrate="1500" maxBitrate="2500" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro.mp4]]></MediaFile><MediaFile id="5244" delivery="progressive" type="video/mp4" bitrate="1000" width="854" height="480" minBitrate="700" maxBitrate="1500" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro-mid-resolution.mp4]]></MediaFile><MediaFile id="5246" delivery="progressive" type="video/mp4" bitrate="600" width="640" height="360" minBitrate="500" maxBitrate="700" scalable="1" maintainAspectRatio="1" codec="H.264"><![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro-low-resolution.mp4]]></MediaFile></MediaFiles><VideoClicks><ClickThrough id="blog"><![CDATA[https://iabtechlab.com]]></ClickThrough></VideoClicks></Linear></Creative></Creatives></InLine></Ad></VAST>`,
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
			},
			wantAdm: `<VAST version="2.0"><Ad id="123"><Wrapper><AdSystem><![CDATA[DSP]]></AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Error><![CDATA[Error URL]]></Error><Error><![CDATA[https://track.dsp.com/er=[ERRORCODE]/tracker/error]]></Error><Impression><![CDATA[Tracker URL]]></Impression><Impression><![CDATA[https://track.dsp.com?e=impression]]></Impression><Creatives><Creative><Linear><TrackingEvents/><VideoClicks><ClickTracking><![CDATA[http://track.dsp.com/tracker/click]]></ClickTracking></VideoClicks></Linear></Creative></Creatives><Extensions><Extension><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></Extension></Extensions></Wrapper></Ad></VAST>`,
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
			},
			wantAdm: `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[Tracker URL]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://172.16.4.213/track]]></Error><Error><![CDATA[https://Errortrack.com]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]]></MediaFile></MediaFiles></Linear></Creative></Creatives><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></InLine></Ad></VAST>`,
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
			},
			wantAdm: `<VAST version="2.0"><Ad id="123"><Wrapper><AdSystem><![CDATA[DSP]]></AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Error><![CDATA[Error URL]]></Error><Error><![CDATA[https://track.dsp.com/er=[ERRORCODE]/tracker/error]]></Error><Impression><![CDATA[Tracker URL]]></Impression><Impression><![CDATA[https://track.dsp.com?e=impression]]></Impression><Creatives><Creative><Linear><TrackingEvents/><VideoClicks><ClickTracking><![CDATA[http://track.dsp.com/tracker/click]]></ClickTracking></VideoClicks></Linear></Creative></Creatives><Extensions><Extension><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></Extension></Extensions></Wrapper></Ad></VAST>`,
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
			},
			wantAdm: `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[Tracking URL]]></Impression><Impression><![CDATA[https://t.pubmatic.com/wt?pubid=156021&purl=https%3A%2F%2Fwww.financialexpress.com%2F&tst=1533048550&iid=9fc7e570-9b01-4cfb-8381-0dc1eec16d27-dieuf&bidid=36110151cad7636&pid=116&pdvid=24&slot=div-gpt-ad-1478842567868-0&pn=rubicon&en=0.02&eg=0.02&kgpv=300x250%40300x250]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://172.16.4.213/track]]></Error><Error><![CDATA[https://Errortrack.com]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]]></MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			wantErr: false,
		},
		{
			name: "vast_2.0_inline_pricing_with_EndpointAppLovinMax",
			args: args{

				bid: openrtb2.Bid{
					AdM:  `<VAST><Ad><InLine></InLine></Ad></VAST>`,
					BURL: "https://burl.com",
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracking URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				rctx: models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			wantBurl: "Tracking URL&owsspburl=https%3A%2F%2Fburl.com",
			wantAdm:  `<VAST><Ad><InLine><Error><![CDATA[Error URL]]></Error><Extensions><Extension><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></Extension></Extensions></InLine></Ad></VAST>`,
			wantErr:  false,
		},
		{
			name: "vast_3.0_inline_pricing_with_EndpointAppLovinMax",
			args: args{

				bid: openrtb2.Bid{
					AdM:  `<VAST version="3.0"><Ad><InLine></InLine></Ad></VAST>`,
					BURL: "https://burl.com",
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				rctx: models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			wantBurl: "Tracker URL&owsspburl=https%3A%2F%2Fburl.com",
			wantAdm:  `<VAST version="3.0"><Ad><InLine><Error><![CDATA[Error URL]]></Error><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></InLine></Ad></VAST>`,
			wantErr:  false,
		},
		{
			name: "inline_vast_3.0_with_EndpointAppLovinMax",
			args: args{

				bid: openrtb2.Bid{
					BURL: "https://burl.com",
					AdM:  `<VAST version='3.0'><Ad id='601364'><InLine><AdSystem>Acudeo Compatible</AdSystem><AdTitle>VAST 2.0 Instream Test 1</AdTitle><Description>VAST 2.0 Instream Test 1</Description><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet?operId=1&pubId=5890&siteId=47163&adId=1405268&adType=13&adServerId=243&kefact=70.000000&kaxefact=70.000000&kadNetFrequecy=0&kadwidth=0&kadheight=0&kadsizeid=97&kltstamp=1529929473&indirectAdId=0&adServerOptimizerId=2&ranreq=0.1&kpbmtpfact=100.000000&dcId=1&tldId=0&passback=0&svr=MADS1107&ekefact=Ad8wW91TCwCmdG0jlfjXn7Tyzh20hnTVx-m5DoNSep-RXGDr&ekaxefact=Ad8wWwRUCwAGir4Zzl1eF0bKiC-qrCV0D0yp_eE7YizB_BQk&ekpbmtpfact=Ad8wWxRUCwD7qgzwwPE2LnS5-Ou19uO5amJl1YT6-XVFvQ41&imprId=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&oid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&crID=creative-1_1_2&ucrid=160175026529250297&campaignId=17050&creativeId=0&pctr=0.000000&wDSPByrId=511&wDspId=6&wbId=0&wrId=0&wAdvID=3170&isRTB=1&rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&pmZoneId=zone1&pageURL=www.yahoo.com&lpu=ae.com]]></Impression><Impression>https://dsptracker.com/{PSPM}</Impression><Error><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&er=[ERRORCODE]]]></Error><Error><![CDATA[https://Errortrack.com?p=1234&er=[ERRORCODE]]]></Error><Creatives><Creative AdID='601364'><Linear skipoffset='20%'><Duration>00:00:04</Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=99]]></ClickTracking><ClickThrough>https://www.sample.com</ClickThrough></VideoClicks><MediaFiles><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]></MediaFile><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				rctx: models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			wantBurl: "Tracker URL&owsspburl=https%3A%2F%2Fburl.com",
			wantAdm:  `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet?operId=1&pubId=5890&siteId=47163&adId=1405268&adType=13&adServerId=243&kefact=70.000000&kaxefact=70.000000&kadNetFrequecy=0&kadwidth=0&kadheight=0&kadsizeid=97&kltstamp=1529929473&indirectAdId=0&adServerOptimizerId=2&ranreq=0.1&kpbmtpfact=100.000000&dcId=1&tldId=0&passback=0&svr=MADS1107&ekefact=Ad8wW91TCwCmdG0jlfjXn7Tyzh20hnTVx-m5DoNSep-RXGDr&ekaxefact=Ad8wWwRUCwAGir4Zzl1eF0bKiC-qrCV0D0yp_eE7YizB_BQk&ekpbmtpfact=Ad8wWxRUCwD7qgzwwPE2LnS5-Ou19uO5amJl1YT6-XVFvQ41&imprId=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&oid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&crID=creative-1_1_2&ucrid=160175026529250297&campaignId=17050&creativeId=0&pctr=0.000000&wDSPByrId=511&wDspId=6&wbId=0&wrId=0&wAdvID=3170&isRTB=1&rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&pmZoneId=zone1&pageURL=www.yahoo.com&lpu=ae.com]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&er=[ERRORCODE]]]></Error><Error><![CDATA[https://Errortrack.com?p=1234&er=[ERRORCODE]]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=99]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></InLine></Ad></VAST>`,
			wantErr:  false,
		},
		{
			name: "wrapper_vast_2.0_with_EndpointAppLovinMax",
			args: args{

				bid: openrtb2.Bid{
					BURL: "https://burl.com",
					AdM:  `<VAST version='2.0'><Ad id='123'><Wrapper><AdSystem>DSP</AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Error><![CDATA[https://track.dsp.com/er=[ERRORCODE]/tracker/error]]></Error><Impression><![CDATA[https://track.dsp.com?e=impression]]></Impression><Creatives><Creative><Linear><TrackingEvents></TrackingEvents><VideoClicks><ClickTracking><![CDATA[http://track.dsp.com/tracker/click]]></ClickTracking></VideoClicks></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				rctx: models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			wantBurl: "Tracker URL&owsspburl=https%3A%2F%2Fburl.com",
			wantAdm:  `<VAST version="2.0"><Ad id="123"><Wrapper><AdSystem><![CDATA[DSP]]></AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Error><![CDATA[Error URL]]></Error><Error><![CDATA[https://track.dsp.com/er=[ERRORCODE]/tracker/error]]></Error><Impression><![CDATA[https://track.dsp.com?e=impression]]></Impression><Creatives><Creative><Linear><TrackingEvents/><VideoClicks><ClickTracking><![CDATA[http://track.dsp.com/tracker/click]]></ClickTracking></VideoClicks></Linear></Creative></Creatives><Extensions><Extension><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></Extension></Extensions></Wrapper></Ad></VAST>`,
			wantErr:  false,
		},
		{
			name: "inline_vast_with_no_cdata_and_EndpointAppLovinMax",
			args: args{

				bid: openrtb2.Bid{
					BURL: "https://burl.com",
					AdM:  `<VAST version='3.0'><Ad id='601364'><InLine><AdSystem>Acudeo Compatible</AdSystem><AdTitle>VAST 2.0 Instream Test 1</AdTitle><Description>VAST 2.0 Instream Test 1</Description><Impression>http://172.16.4.213/AdServer/AdDisplayTrackerServlet</Impression><Impression>https://dsptracker.com/{PSPM}</Impression><Error>http://172.16.4.213/track</Error><Error>https://Errortrack.com</Error><Creatives><Creative AdID='601364'><Linear skipoffset='20%'><Duration>00:00:04</Duration><VideoClicks><ClickTracking>http://172.16.4.213/track</ClickTracking><ClickThrough>https://www.sample.com</ClickThrough></VideoClicks><MediaFiles><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'>https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]</MediaFile><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'>https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]</MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				rctx: models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			wantBurl: "Tracker URL&owsspburl=https%3A%2F%2Fburl.com",
			wantAdm:  `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://172.16.4.213/track]]></Error><Error><![CDATA[https://Errortrack.com]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]]></MediaFile></MediaFiles></Linear></Creative></Creatives><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></InLine></Ad></VAST>`,
			wantErr:  false,
		},
		{
			name: "wrapper_vast_with_no_cdata_and_EndpointAppLovinMax",
			args: args{

				bid: openrtb2.Bid{
					BURL: "https://burl.com",
					AdM:  `<VAST version='2.0'><Ad id='123'><Wrapper><AdSystem>DSP</AdSystem><VASTAdTagURI>https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml</VASTAdTagURI><Error>https://track.dsp.com/er=[ERRORCODE]/tracker/error</Error><Impression>https://track.dsp.com?e=impression</Impression><Creatives><Creative><Linear><TrackingEvents></TrackingEvents><VideoClicks><ClickTracking>http://track.dsp.com/tracker/click</ClickTracking></VideoClicks></Linear></Creative></Creatives></Wrapper></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
						Price:      1.2,
					},
				},
				rctx: models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			wantBurl: "Tracker URL&owsspburl=https%3A%2F%2Fburl.com",
			wantAdm:  `<VAST version="2.0"><Ad id="123"><Wrapper><AdSystem><![CDATA[DSP]]></AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Error><![CDATA[Error URL]]></Error><Error><![CDATA[https://track.dsp.com/er=[ERRORCODE]/tracker/error]]></Error><Impression><![CDATA[https://track.dsp.com?e=impression]]></Impression><Creatives><Creative><Linear><TrackingEvents/><VideoClicks><ClickTracking><![CDATA[http://track.dsp.com/tracker/click]]></ClickTracking></VideoClicks></Linear></Creative></Creatives><Extensions><Extension><Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing></Extension></Extensions></Wrapper></Ad></VAST>`,
			wantErr:  false,
		},
		{
			name: "spaces_in_creative_with_EndpointAppLovinMax",
			args: args{

				bid: openrtb2.Bid{
					BURL: "https://burl.com",
					AdM:  `<VAST version="3.0">   <Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[https://t.pubmatic.com/wt?pubid=156021&purl=https%3A%2F%2Fwww.financialexpress.com%2F&tst=1533048550&iid=9fc7e570-9b01-4cfb-8381-0dc1eec16d27-dieuf&bidid=36110151cad7636&pid=116&pdvid=24&slot=div-gpt-ad-1478842567868-0&pn=rubicon&en=0.02&eg=0.02&kgpv=300x250%40300x250]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[http://172.16.4.213/track]]></Error><Error><![CDATA[https://Errortrack.com]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]]></MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
				},
				videoParams: []models.OWTracker{
					{
						TrackerURL: `Tracker URL`,
						ErrorURL:   `Error URL`,
					},
				},
				rctx: models.RequestCtx{
					Endpoint: models.EndpointAppLovinMax,
				},
			},
			wantBurl: "Tracker URL&owsspburl=https%3A%2F%2Fburl.com",
			wantAdm:  `<VAST version="3.0"><Ad id="601364"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[https://t.pubmatic.com/wt?pubid=156021&purl=https%3A%2F%2Fwww.financialexpress.com%2F&tst=1533048550&iid=9fc7e570-9b01-4cfb-8381-0dc1eec16d27-dieuf&bidid=36110151cad7636&pid=116&pdvid=24&slot=div-gpt-ad-1478842567868-0&pn=rubicon&en=0.02&eg=0.02&kgpv=300x250%40300x250]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[Error URL]]></Error><Error><![CDATA[http://172.16.4.213/track]]></Error><Error><![CDATA[https://Errortrack.com]]></Error><Creatives><Creative AdID="601364"><Linear skipoffset="20%"><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track]]></ClickTracking><ClickThrough><![CDATA[https://www.sample.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]]]></MediaFile><MediaFile delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" scalable="true" maintainAspectRatio="true"><![CDATA[https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]]]></MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdm, gotBurl, err := injectVideoCreativeTrackers(tt.args.rctx, tt.args.bid, tt.args.videoParams)
			if (err != nil) != tt.wantErr {
				t.Errorf("injectVideoCreativeTrackers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantAdm, gotAdm, tt.name)
			assert.Equal(t, tt.wantBurl, gotBurl, tt.name)
		})
	}
}
