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
			want: Video,
		},
		{
			name: "empty_AdM",
			args: args{
				adm: "",
			},
			want: "",
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
