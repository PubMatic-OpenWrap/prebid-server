package openrtb_ext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCreativeTypeFromCreative(t *testing.T) {
	type args struct {
		adm string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "video_creative",
			args: args{
				adm: "<VAST version=\"3.0\"></VAST>",
			},
			want: Video,
		},
		{
			name: "native_creative",
			args: args{
				adm: "{\"native\":{\"link\":{\"url\":\"http://example.com\"},\"assets\":[]}}",
			},
			want: Native,
		},
		{
			name: "native_creative_with_assets",
			args: args{
				adm: "{\"native\":{\"link\":{\"url\":\"http://example.com\"},\"assets\":[{\"id\":1,\"title\":{\"text\":\"Title\"}}]}}",
			},
			want: Native,
		},
		{
			name: "native_creative_with_link",
			args: args{
				adm: "{\"native\":{\"link\":{\"url\":\"http://example.com\"}}}",
			},
			want: Native,
		},
		{
			name: "native_creative_empty_native",
			args: args{
				adm: "{\"native\":{}}",
			},
			want: Native,
		},
		{
			name: "banner_creative",
			args: args{
				adm: "<div>Banner Ad</div>",
			},
			want: Banner,
		},
		{
			name: "banner_adm_has_json_assets_keyword",
			args: args{
				adm: "<script src='mraid.js'></script><script>document.createElement('IMG').src=\"https://tlx.3lift.com/s2s/notify?px=1&pr=${AUCTION_PRICE}&ts=1686039380&aid=45036723730776858540010&ec=2711_67927_9991141&n=GFABCP4AhSIAwCSAwQwMTNimAMAoAOfqByoAwA%3D\";window.tl_auction_response_559942={\"settings\":{\"viewability\":{},\"additional_data\":{\"pr\":\"${AUCTION_PRICE}\",\"bc\":\"AAABiI_HQr-J9SWLTbGinTR6NNuHz29x102WBw==\",\"aid\":\"45036723730776858540010\",\"bmid\":\"2711\",\"biid\":\"7295\",\"sid\":\"67927\",\"brid\":\"82983\",\"adid\":\"9991141\",\"crid\":\"737729\",\"ts\":\"1686039380\",\"bcud\":\"1240\",\"ss\":\"20\"},\"template_id\":210,\"payable_event\":1,\"billable_event\":1,\"billable_pixel\":\"https://tlx.3lift.com/s2s/notify?px=1&pr=${AUCTION_PRICE}&ts=1686039380&aid=45036723730776858540010&ec=2711_67927_9991141&n=GpIGaCP%3D&b=1\",\"adchoices_url\":\"https://optout.aboutads.info/\",\"format_id\":10,\"render_options_bm\":0,\"cta\":\"Learn more\"},\"assets\":[{\"asset_id\":0,\"cta\":\"Learn more\",\"banner_width\":300,\"banner_height\":600,\"banner_markup\":\"<script type='text/javascript' src='https://ads.as.criteo.com/delivery/r/ajs.php?z=AAABiI_HQr-J9SWLTbGinTR6NNuHz29x102WBw==&u=%7Cgud5gNZYq-lw&ct0={clickurl_enc}'></script>\"}]};</script><script src=\"https://ib.3lift.com/ttj?inv_code=HK01_Android_Opening_InListBox_3_336x280&tid=210\" data-auction-response-id=\"559942\" data-ss-id=\"20\"></script>",
			},
			want: Banner,
		},
		{
			name: "banner_adm_has_video_keyword",
			args: args{
				adm: "<span class='PubAPIAd' id='imp176227948'><div id='66F8570D-FB8D-4929-A9BB-88DFF50C4BE3'></div><script>window.bbwAdUnitCode = 'pubmatic_ibv_159557';window.inBannerVideoConfig = { behaviour: 'COLLAPSE' };window.uniqueElementId = '66F8570D-FB8D-4929-A9BB-88DFF50C4BE3';window.vast_xml= \"<VAST version='4.0'><Ad id='601364'><InLine><AdSystem>Acudeo Compatible</AdSystem><AdTitle>VAST 2.0 Instream Test 1</AdTitle><Description>VAST 2.0 Instream Test 1</Description><Impression><![CDATA[https://st.pubmatic.com/AdServer/AdDisplayTrackerServlet?operId=1&pubId=164871&siteId=1221311&adId=6118201]]></Impression><Creatives><Creative AdID='601364'><Linear skipoffset='70%'><TrackingEvents><Tracking event='close'><![CDATA[https://mytracking.com/linear/close]]></Tracking><Tracking event='skip'><![CDATA[https://mytracking.com/linear/skip]]></Tracking></TrackingEvents><Duration>00:00:04</Duration><VideoClicks><ClickTracking><![CDATA[https://st.pubmatic.com/track?operId=7&p=164871&s=1221311]]></ClickTracking><ClickThrough>https://www.pubmatic.com</ClickThrough></VideoClicks><MediaFiles><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'><![CDATA[https://owsdk-stagingams.pubmatic.com:8443/openwrap/media/pubmatic.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>\";</script><script src='//ads.pubmatic.com/AdServer/js/vastTemplate/vastRenderer_v2.js'></script></span><div style='position:absolute;left:0px;top:0px;visibility:hidden;'><img src='https://t.pubmatic.com/wt?adv=ae.com&af=banner'></div>"},
			want: Banner,
		},
		{
			name: "in_banner_video",
			args: args{
				adm: "<div id=\"imp1\"></div><script>window.bbwAdUnitCode = \"ow_amp_5890\";window.inBannerVideoConfig = { behaviour: \"REPLAY\" };window.uniqueElementId = \"imp1\";window.vast_xml = `<VAST version=\"3.0\"><Ad id=\"601364\"><InLine><AdSystem><![CDATA[Acudeo Compatible]]></AdSystem><AdTitle><![CDATA[VAST 2.0 Instream Test 1]]></AdTitle><Description><![CDATA[VAST 2.0 Instream Test 1]]></Description><Impression><![CDATA[https://t.pubmatic.com/wt?adv=ae.com&af=video&aps=0&au=%2F43743431%2FQAAMP1UC&bc=pubmatic&bidid=5b5b593a-6b07-47a3-bb4d-e904b3ff1d8c&di=PUBDEAL1&dur=30&eg=15&en=13.5&fsrc=0&ft=1&iid=65f9d24f-feeb-4912-9194-979a09798cdc&kgpv=%2F43743431%2FQAAMP1UC%400x0&orig=&origbidid=VIDEO12-89A1-41F1-8708-978FD3C0912A&pdvid=4&pid=36120&plt=2&pn=pubmatic&psz=0x0&pubid=5890&purl=file%3A%2F%2F%2Fhome%2Ftest%2Fautomation%2Fopenwrap-automation%2Fresources%2Fdecision_manager_resources%2Ftest_pages%2FAMP%2Fsingle_slot_single_size.html%3Fpwtvc%3D1%26profileid%3D36120%26pwtv%3D4%26owLoggerDebug%3D1&sl=1&slot=%2F43743431%2FQAAMP1UC&ss=1&tst=1710931540]]></Impression><Impression><![CDATA[http://172.16.4.213/AdServer/AdDisplayTrackerServlet?operId=1&pubId=5890&siteId=47163&adId=1405268&adType=13&adServerId=243&kefact=70.000000&kaxefact=70.000000&kadNetFrequecy=0&kadwidth=0&kadheight=0&kadsizeid=97&kltstamp=1529929473&indirectAdId=0&adServerOptimizerId=2&ranreq=0.1&kpbmtpfact=100.000000&dcId=1&tldId=0&passback=0&svr=MADS1107&ekefact=Ad8wW91TCwCmdG0jlfjXn7Tyzh20hnTVx-m5DoNSep-RXGDr&ekaxefact=Ad8wWwRUCwAGir4Zzl1eF0bKiC-qrCV0D0yp_eE7YizB_BQk&ekpbmtpfact=Ad8wWxRUCwD7qgzwwPE2LnS5-Ou19uO5amJl1YT6-XVFvQ41&imprId=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&oid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&crID=creative-1_1_2&ucrid=160175026529250297&campaignId=17050&creativeId=0&pctr=0.000000&wDSPByrId=511&wDspId=6&wbId=0&wrId=0&wAdvID=3170&isRTB=1&rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&pmZoneId=zone1&pageURL=www.yahoo.com&lpu=ae.com]]></Impression><Impression><![CDATA[https://dsptracker.com/{PSPM}]]></Impression><Error><![CDATA[https://127.0.0.1:8080/wt?operId=8&adv=ae.com&au=%2F43743431%2FQAAMP1UC&bc=pubmatic&crId=creative-1_1_2&p=5890&pfi=2&pid=36120&pn=pubmatic&sURL=&ts=1710931540&v=4&ier=[ERRORCODE]]]></Error><Error><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&er=[ERRORCODE]]]></Error><Error><![CDATA[https://Errortrack.com?p=1234&er=[ERRORCODE]]]></Error><Creatives><Creative AdID=\"601364\"><Linear skipoffset=\"70%\"><TrackingEvents><Tracking event=\"close\"><![CDATA[https://mytracking.com/linear/close]]></Tracking><Tracking event=\"skip\"><![CDATA[https://mytracking.com/linear/skip]]></Tracking><Tracking event=\"creativeView\"><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=1]]></Tracking><Tracking event=\"start\"><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=2]]></Tracking><Tracking event=\"midpoint\"><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=3]]></Tracking><Tracking event=\"firstQuartile\"><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=4]]></Tracking><Tracking event=\"thirdQuartile\"><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=5]]></Tracking><Tracking event=\"complete\"><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=6]]></Tracking><Tracking event=\"complete\"><![CDATA[https://127.0.0.1:8080/wt?operId=8&e=6&p=5890&pid=36120&v=4&ts=1710931540&pn=pubmatic&adv=ae.com&sURL=&pfi=2&af=video&iid=65f9d24f-feeb-4912-9194-979a09798cdc&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2F43743431%2FQAAMP1UC&bidid=5b5b593a-6b07-47a3-bb4d-e904b3ff1d8c&origbidid=VIDEO12-89A1-41F1-8708-978FD3C0912A&bc=pubmatic&ssai=[SSAI]]]></Tracking><Tracking event=\"start\"><![CDATA[https://127.0.0.1:8080/wt?operId=8&e=2&p=5890&pid=36120&v=4&ts=1710931540&pn=pubmatic&adv=ae.com&sURL=&pfi=2&af=video&iid=65f9d24f-feeb-4912-9194-979a09798cdc&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2F43743431%2FQAAMP1UC&bidid=5b5b593a-6b07-47a3-bb4d-e904b3ff1d8c&origbidid=VIDEO12-89A1-41F1-8708-978FD3C0912A&bc=pubmatic&ssai=[SSAI]]]></Tracking><Tracking event=\"firstQuartile\"><![CDATA[https://127.0.0.1:8080/wt?operId=8&e=4&p=5890&pid=36120&v=4&ts=1710931540&pn=pubmatic&adv=ae.com&sURL=&pfi=2&af=video&iid=65f9d24f-feeb-4912-9194-979a09798cdc&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2F43743431%2FQAAMP1UC&bidid=5b5b593a-6b07-47a3-bb4d-e904b3ff1d8c&origbidid=VIDEO12-89A1-41F1-8708-978FD3C0912A&bc=pubmatic&ssai=[SSAI]]]></Tracking><Tracking event=\"midpoint\"><![CDATA[https://127.0.0.1:8080/wt?operId=8&e=3&p=5890&pid=36120&v=4&ts=1710931540&pn=pubmatic&adv=ae.com&sURL=&pfi=2&af=video&iid=65f9d24f-feeb-4912-9194-979a09798cdc&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2F43743431%2FQAAMP1UC&bidid=5b5b593a-6b07-47a3-bb4d-e904b3ff1d8c&origbidid=VIDEO12-89A1-41F1-8708-978FD3C0912A&bc=pubmatic&ssai=[SSAI]]]></Tracking><Tracking event=\"thirdQuartile\"><![CDATA[https://127.0.0.1:8080/wt?operId=8&e=5&p=5890&pid=36120&v=4&ts=1710931540&pn=pubmatic&adv=ae.com&sURL=&pfi=2&af=video&iid=65f9d24f-feeb-4912-9194-979a09798cdc&pseq=[PODSEQUENCE]&adcnt=[ADCOUNT]&cb=[CACHEBUSTING]&au=%2F43743431%2FQAAMP1UC&bidid=5b5b593a-6b07-47a3-bb4d-e904b3ff1d8c&origbidid=VIDEO12-89A1-41F1-8708-978FD3C0912A&bc=pubmatic&ssai=[SSAI]]]></Tracking></TrackingEvents><Duration><![CDATA[00:00:04]]></Duration><VideoClicks><ClickTracking><![CDATA[http://172.16.4.213/track?operId=7&p=5890&s=47163&a=1405268&wa=243&ts=1529929473&wc=17050&crId=creative-1_1_2&ucrid=160175026529250297&impid=48F73E1A-7F23-443D-A53C-30EE6BBF5F7F&advertiser_id=3170&ecpm=70.000000&e=99]]></ClickTracking><ClickThrough><![CDATA[https://www.pubmatic.com]]></ClickThrough></VideoClicks><MediaFiles><MediaFile delivery=\"progressive\" type=\"video/mp4\" bitrate=\"500\" width=\"400\" height=\"300\" scalable=\"true\" maintainAspectRatio=\"true\"><![CDATA[https://owsdk-stagingams.pubmatic.com:8443/openwrap/media/pubmatic.mp4]]></MediaFile></MediaFiles></Linear></Creative></Creatives><Pricing model=\"CPM\" currency=\"USD\"><![CDATA[13.5]]></Pricing></InLine></Ad></VAST>`</script><script src=\"https://ads.pubmatic.com/AdServer/js/vastTemplate/vastRenderer_v2.js\"></script>",
			},
			want: Banner,
		},
		{
			name: "in_banner_video_adm",
			args: args{
				adm: "<script><VAST version=\"3.0\"></VAST></script>",
			},
			want: Banner,
		},
		{
			name: "empty_AdM",
			args: args{
				adm: "",
			},
			want: "",
		},
		{
			name: "empty_Vast_AdM",
			args: args{
				adm: "<VAST></VAST>",
			},
			want: Video,
		},
		{
			name: "invalid_json_in_adm",
			args: args{
				adm: "{\"native\":{\"link\":{\"url\":\"http://example.com\"},\"assets\":[]",
			},
			want: Banner,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCreativeTypeFromCreative(tt.args.adm)
			assert.Equal(t, tt.want, got)
		})
	}
}
