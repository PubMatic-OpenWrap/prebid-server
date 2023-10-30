package pubmatic

import (
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestGetAdFormat(t *testing.T) {
	tests := []struct {
		name   string
		adm    string
		format string
	}{
		{
			name:   "Empty Bid Adm",
			adm:    "",
			format: models.Banner,
		},
		{
			name:   "Banner Adm has json assets keyword",
			adm:    `u003cscript src='mraid.js'></script><script>document.createElement('IMG').src="https://tlx.3lift.com/s2s/notify?px=1&pr=${AUCTION_PRICE}&ts=1686039380&aid=45036723730776858540010&ec=2711_67927_9991141&n=GFABCP4AhSIAwCSAwQwMTNimAMAoAOfqByoAwA%3D";window.tl_auction_response_559942={"settings":{"viewability":{},"additional_data":{"pr":"${AUCTION_PRICE}","bc":"AAABiI_HQr-J9SWLTbGinTR6NNuHz29x102WBw==","aid":"45036723730776858540010","bmid":"2711","biid":"7295","sid":"67927","brid":"82983","adid":"9991141","crid":"737729","ts":"1686039380","bcud":"1240","ss":"20"},"template_id":210,"payable_event":1,"billable_event":1,"billable_pixel":"https:\/\/tlx.3lift.com\/s2s\/notify?px=1&pr=${AUCTION_PRICE}&ts=1686039380&aid=45036723730776858540010&ec=2711_67927_9991141&n=GpIGaCP%3D&b=1","adchoices_url":"https:\/\/optout.aboutads.info\/","format_id":10,"render_options_bm":0,"cta":"Learn more"},"assets":[{"asset_id":0,"cta":"Learn more","banner_width":300,"banner_height":600,"banner_markup":"<script type='text\/javascript' src='https:\/\/ads.as.criteo.com\/delivery\/r\/ajs.php?z=AAABiI_HQr-J9SWLTbGinTR6NNuHz29x102WBw==&u=%7Cgud5gNZYq-lw&ct0={clickurl_enc}'><\/script>"}]};</script><script src="https://ib.3lift.com/ttj?inv_code=HK01_Android_Opening_InListBox_3_336x280&tid=210" data-auction-response-id="559942" data-ss-id="20"></script>`,
			format: models.Banner,
		},
		{
			name:   "VAST Ad",
			adm:    "<VAST version='3.0'><Ad id='601364'><InLine><AdSystem>Acudeo Compatible</AdSystem><AdTitle>VAST 2.0 Instream Test 1</AdTitle><Description>VAST 2.0 Instream Test 1</Description><Impression>http://172.16.4.213/AdServer/AdDisplayTrackerServlet</Impression><Impression>https://dsptracker.com/{PSPM}</Impression><Error>http://172.16.4.213/track</Error><Error>https://Errortrack.com</Error><Creatives><Creative AdID='601364'><Linear skipoffset='20%'><Duration>00:00:04</Duration><VideoClicks><ClickTracking>http://172.16.4.213/track</ClickTracking><ClickThrough>https://www.pubmatic.com</ClickThrough></VideoClicks><MediaFiles><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'>https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]</MediaFile><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'>https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]</MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>",
			format: models.Video,
		},
		{
			name:   "VAST Ad xml",
			adm:    "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?><VAST version=\"4.0\"><Ad id=\"97517771\"><Wrapper><AdSystem version=\"4.0\">adnxs</AdSystem><VASTAdTagURI><![CDATA[http://sin3-ib.adnxs.com/ab?an_audit=0&test=1&referrer=http%3A%2F%2Fprebid.org%2Fexamples%2Fvideo%2Fserver%2Fjwplayer%2Fpbs-ve-jwplayer-hosted.html&e=wqT_3QKUC6CUBQAAAwDWAAUBCObImfIFEN_CrZCNxam9JhjNy6T75_2f4igqNgkAAAECCBRAEQEHNAAAFEAZAAAAwB6FPUAhERIAKREJADERG6gw6dGnBjjtSEDtSEgCUMuBwC5YnPFbYABozbp1eMu4BYABAYoBA1VTRJIBAQbwUpgBAaABAagBAbABALgBA8ABA8gBAtABANgBAOABAfABAIoCO3VmKCdhJywgMjUyOTg4NSwgMTU4MTY3MTUyNik7dWYoJ3InLCA5NzUxNzc3MSwgLh4A9A4BkgLRAiE0RUQzNndpMi1Md0tFTXVCd0M0WUFDQ2M4VnN3QURnQVFBUkk3VWhRNmRHbkJsZ0FZUF9fX184UGFBQndBWGdCZ0FFQmlBRUJrQUVCbUFFQm9BRUJxQUVEc0FFQXVRSHpyV3FrQUFBVVFNRUI4NjFxcEFBQUZFREpBVlBTU2JZNVlPY18yUUVBQUFBQUFBRHdQLUFCQVBVQkFBQUFBSmdDQUtBQ0FMVUNBQUFBQUwwQ0FBQUFBTUFDQWNnQ0FkQUNBZGdDQWVBQ0FPZ0NBUGdDQUlBREFaZ0RBYWdEdHZpOENyb0RDVk5KVGpNNk5EZ3pOdUFEbEJ1SUJBQ1FCQUNZQkFIQkJBQUFBQQmDCHlRUQkJAQEYTmdFQVBFRQELCQEgQ0lCZVFscVFVCQ8YQUR3UDdFRg0NAQEsLpoCiQEheXc3WjlnNlUBJG5QRmJJQVFvQUQVUFRVUURvSlUwbE9Nem8wT0RNMlFKUWJTEYAMUEFfVREMDEFBQVcdDABZHQwAYR0MAGMdDPBSZUFBLsICP2h0dHA6Ly9wcmViaWQub3JnL2Rldi1kb2NzL3Nob3ctdmlkZW8td2l0aC1hLWRmcC12aWRlby10YWcuaHRtbNgCAOACrZhI6gJMaHQ-SgAgZXhhbXBsZXMvBUVcL3NlcnZlci9qd3BsYXllci9wYnMtdmUtERAYLWhvc3RlZAVXNPICEQoGQURWX0lEEgcySbsFFAhDUEcFFBg1NzU5MzY0ARQIBUNQARM0CDIxOTY5OTc08gINCggBPBhGUkVREgEwBRAcUkVNX1VTRVIFEAAMCSAYQ09ERRIA8gEPAVcRDxALCgdDUBUOEAkKBUlPAWAEAPIBGgRJTxUaOBMKD0NVU1RPTV9NT0RFTA0kCBoKFjIWABxMRUFGX05BTQVqCB4KGjYdAAhBU1QBPhBJRklFRAFiHA0KCFNQTElUAU3wgQEwgAMAiAMBkAMAmAMUoAMBqgMAwAPgqAHIAwDYAwDgAwDoAwD4AwOABACSBAkvb3BlbnJ0YjKYBACiBA0xODIuNzQuMzkuMjUwqAQAsgQOCAAQBBiABSDoAjAAOAS4BADABADIBADSBA45MzI1I1NJTjM6NDgzNtoEAggA4AQA8ASBgSCIBQGYBQCgBf8RAbABqgUkZDc0MzQ3ZDUtYzY3Mi00NTM5LWIxNDEtOWVjMWMzMzJiZTI2wAUAyQWJ4hTwP9IFCQkJDHgAANgFAeAFAfAFw5UL-gUECAAQAJAGAZgGALgGAMEGCSUo8D_QBvUv2gYWChAJERkBUBAAGADgBgTyBgIIAIAHAYgHAKAHQA..&s=dcc685e3549971224cbd8615ff729bcb19107ec0&pp=${AUCTION_PRICE}]]></VASTAdTagURI><Impression><![CDATA[http://ib.adnxs.com/nop]]></Impression><Creatives><Creative adID=\"97517771\"><Linear></Linear></Creative></Creatives></Wrapper></Ad></VAST>",
			format: models.Video,
		},
		{
			name:   "Banner Ad",
			adm:    `"<span class=\"PubAPIAd\"  id=\"4E733404-CC2E-48A2-BC83-4DD5F38FE9BB\"><script type=\"text/javascript\"> document.writeln('<iframe width=\"300\" scrolling=\"no\" height=\"250\" frameborder=\"0\" name=\"iframe0\" allowtransparency=\"true\" marginheight=\"0\" marginwidth=\"0\" vspace=\"0\" hspace=\"0\" src=\"https://ads.pubmatic.com/AdTag/300x250.png\"></iframe>');</script><iframe width=\"0\" scrolling=\"no\" height=\"0\" frameborder=\"0\" src=\"https://st.pubmatic.com/AdServer/AdDisplayTrackerServlet?pubId=5890\" style=\"position:absolute;top:-15000px;left:-15000px\" vspace=\"0\" hspace=\"0\" marginwidth=\"0\" marginheight=\"0\" allowtransparency=\"true\" name=\"pbeacon\"></iframe></span> <!-- PubMatic Ad Ends --><div style=\"position:absolute;left:0px;top:0px;visibility:hidden;\">`,
			format: models.Banner,
		},
		// {
		// 	name:   "Native Adm with `assets` Object",
		// 	adm:    `{"assets":[{"id":0,"img":{"type":3,"url":"//ads.pubmatic.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":1,"data":{"type":1,"value":"Sponsored By PubMatic"}},{"id":2,"img":{"type":1,"url":"//ads.pubmatic.com/AdTag/native/728x90.png","w":728,"h":90}},{"id":3,"title":{"text":"Native Test Title"}},{"id":4,"data":{"type":2,"value":"Sponsored By PubMatic"}}],"link":{"url":"//www.pubmatic.com","clicktrackers":["http://clicktracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.pubmatic.com"},"imptrackers":["http://clicktracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/ads.pubmatic.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/ads.pubmatic.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://aktrack.pubmatic.com/AdServer/AdDisplayTrackerServlet?operId=1\u0026pubId=157962\u0026siteId=547907\u0026adId=1947750\u0026adType=12\u0026adServerId=243\u0026kefact=0.010000\u0026kaxefact=0.010000\u0026kadNetFrequecy=0\u0026kadwidth=0\u0026kadheight=0\u0026kadsizeid=0\u0026kltstamp=1658395546\u0026indirectAdId=0\u0026adServerOptimizerId=2\u0026ranreq=0.1\u0026kpbmtpfact=0.010000\u0026dcId=1\u0026tldId=0\u0026passback=0\u0026svr=pritiads\u0026adsver=_3952513331\u0026adsabzcid=0\u0026cls=pri\u0026ekefact=mhvZYu3OCACapb_iETG3xdxtC1tPzzUJu-KGCg7FolbaREk4\u0026ekaxefact=mhvZYvzOCACzeBW8kmVTsJknqYBPJWS55RNrwnAXq55kxLgA\u0026ekpbmtpfact=mhvZYgjPCAAp7x5RPBYGHfni3ntITSDe7G4kwxt3A2ZtoEPx\u0026enpp=mhvZYhXPCAAZEEUDckm1AOVbVrc4p9LF534kZuebuZ4ATSXX\u0026pfi=2\u0026dc=VA2\u0026crID=119_4987704\u0026lpu=ableunited.com\u0026ucrid=15711954930062968336\u0026campaignId=23041\u0026creativeId=0\u0026pctr=0.000000\u0026wDSPByrId=494\u0026wDspId=632\u0026wbId=0\u0026wrId=0\u0026wAdvID=700327\u0026wDspCampId=16530\u0026isRTB=1\u0026rtbId=3012556727136858784\u0026imprId=7BAD05C7-994C-49AF-95B8-4C5CC9542025\u0026oid=7BAD05C7-994C-49AF-95B8-4C5CC9542025\u0026mobflag=2\u0026country=IN\u0026pAuSt=2\u0026wops=0\u0026sURL=gamerch.com"}]}`,
		// 	format: models.Native,
		// },
		{
			name:   "Native Adm with `native` Object",
			adm:    `{"native":{"ver":1.2,"link":{"url":"https://dummyimage.com/1x1/000000/fff.jpg&text=420x420+Creative.jpg","clicktrackers":["http://image3.pubmatic.com/AdServer/layer?a={PUBMATIC_SECOND_PRICE}&ucrid=9335447642416814892&t=FNOZW09VkdSTVM0eU5BPT09JmlkPTAmY2lkPTIyNzcyJnhwcj0xLjAwMDAwMCZmcD00JnBwPTIuMzcxMiZ0cD0yJnBlPTAuMDA9=","http://image3.pubmatic.com/AdServer/layer?a={PUBMATIC_SECOND_PRICE}&ucrid=9335447642416814892&t=FNOZW09VkdSTVM0eU5BPT09JmlkPTAmY2lkPTIyNzcyJnhwcj0xLjAwMDAwMCZmcD00JnBwPTIuMzcxMiZ0cD0yJnBlPTAuMDA9="]},"eventtrackers":[{"event":1,"method":1,"url":"http://image3.pubmatic.com/AdServer/layer?a={PUBMATIC_SECOND_PRICE}&ucrid=9335447642416814892&t=FNOZW09VkdSTVM0eU5BPT09JmlkPTAmY2lkPTIyNzcyJnhwcj0xLjAwMDAwMCZmcD00JnBwPTIuMzcxMiZ0cD0yJnBlPTAuMDA9="}]}}`,
			format: models.Native,
		},
		{
			name:   "Native Adm with `native` and `assets` Object",
			adm:    "{\"native\":{\"assets\":[{\"id\":1,\"required\":0,\"title\":{\"text\":\"Lexus - Luxury vehicles company\"}},{\"id\":2,\"img\":{\"h\":150,\"url\":\"https://stagingnyc.pubmatic.com:8443//sdk/lexus_logo.png\",\"w\":150},\"required\":0},{\"id\":3,\"img\":{\"h\":428,\"url\":\"https://stagingnyc.pubmatic.com:8443//sdk/28f48244cafa0363b03899f267453fe7%20copy.png\",\"w\":214},\"required\":0},{\"data\":{\"value\":\"Goto PubMatic\"},\"id\":4,\"required\":0},{\"data\":{\"value\":\"Lexus - Luxury vehicles company\"},\"id\":5,\"required\":0},{\"data\":{\"value\":\"4\"},\"id\":6,\"required\":0}],\"imptrackers\":[\"http://phtrack.pubmatic.com/?ts=1496043362&r=84137f17-eefd-4f06-8380-09138dc616e6&i=c35b1240-a0b3-4708-afca-54be95283c61&a=130917&t=9756&au=10002949&p=&c=10014299&o=10002476&wl=10009731&ty=1\"],\"link\":{\"clicktrackers\":[\"http://ct.pubmatic.com/track?ts=1496043362&r=84137f17-eefd-4f06-8380-09138dc616e6&i=c35b1240-a0b3-4708-afca-54be95283c61&a=130917&t=9756&au=10002949&p=&c=10014299&o=10002476&wl=10009731&ty=3&url=\"],\"url\":\"http://www.lexus.com/\"},\"ver\":1}}",
			format: models.Native,
		},
		// {
		// 	name:   "Native Adm with `link` Object",
		// 	adm:    `{"link":{"url":"//www.pubmatic.com","clicktrackers":["http://clicktracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"fallback":"http://www.pubmatic.com"},"imptrackers":["http://clicktracker.com/AdTag/9bde02d0-6017-11e4-9df7-005056967c35"],"jstracker":"\u003cscript src='\\/\\/ads.pubmatic.com\\/AdTag\\/native\\/tempReseponse.js'\u003e\u003cscript src='\\/\\/ads.pubmatic.com\\/AdTag\\/native\\/tempReseponse.js'\u003e","eventtrackers":[{"event":1,"method":1,"url":"http://aktrack.pubmatic.com/AdServer/AdDisplayTrackerServlet?operId=1\u0026pubId=157962\u0026siteId=547907\u0026adId=1947750\u0026adType=12\u0026adServerId=243\u0026kefact=0.010000\u0026kaxefact=0.010000\u0026kadNetFrequecy=0\u0026kadwidth=0\u0026kadheight=0\u0026kadsizeid=0\u0026kltstamp=1658395546\u0026indirectAdId=0\u0026adServerOptimizerId=2\u0026ranreq=0.1\u0026kpbmtpfact=0.010000\u0026dcId=1\u0026tldId=0\u0026passback=0\u0026svr=pritiads\u0026adsver=_3952513331\u0026adsabzcid=0\u0026cls=pri\u0026ekefact=mhvZYu3OCACapb_iETG3xdxtC1tPzzUJu-KGCg7FolbaREk4\u0026ekaxefact=mhvZYvzOCACzeBW8kmVTsJknqYBPJWS55RNrwnAXq55kxLgA\u0026ekpbmtpfact=mhvZYgjPCAAp7x5RPBYGHfni3ntITSDe7G4kwxt3A2ZtoEPx\u0026enpp=mhvZYhXPCAAZEEUDckm1AOVbVrc4p9LF534kZuebuZ4ATSXX\u0026pfi=2\u0026dc=VA2\u0026crID=119_4987704\u0026lpu=ableunited.com\u0026ucrid=15711954930062968336\u0026campaignId=23041\u0026creativeId=0\u0026pctr=0.000000\u0026wDSPByrId=494\u0026wDspId=632\u0026wbId=0\u0026wrId=0\u0026wAdvID=700327\u0026wDspCampId=16530\u0026isRTB=1\u0026rtbId=3012556727136858784\u0026imprId=7BAD05C7-994C-49AF-95B8-4C5CC9542025\u0026oid=7BAD05C7-994C-49AF-95B8-4C5CC9542025\u0026mobflag=2\u0026country=IN\u0026pAuSt=2\u0026wops=0\u0026sURL=gamerch.com"}]}`,
		// 	format: models.Native,
		// },
		{
			name:   "Video Adm",
			adm:    "<VAST ></VAST>",
			format: models.Video,
		},
		{
			name:   "Video Adm \t",
			adm:    "<VAST\t></VAST>",
			format: models.Video,
		},
		{
			name:   "Video Adm \r",
			adm:    "<VAST\r></VAST>",
			format: models.Video,
		},
		{
			name:   "Video Adm \n",
			adm:    "<VAST\n></VAST>",
			format: models.Video,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format := GetAdFormat(tt.adm)
			assert.Equal(t, tt.format, format, tt.name)
		})
	}
}

func TestGetRevenueShare(t *testing.T) {
	tests := []struct {
		name          string
		partnerConfig map[string]string
		revshare      float64
	}{
		{
			name:          "Empty partnerConfig",
			partnerConfig: make(map[string]string),
			revshare:      0,
		},
		{
			name: "partnerConfig without rev_share",
			partnerConfig: map[string]string{
				"anykey": "anyval",
			},
			revshare: 0,
		},
		{
			name: "partnerConfig with invalid rev_share",
			partnerConfig: map[string]string{
				models.REVSHARE: "invalid",
			},
			revshare: 0,
		},
		{
			name: "partnerConfig with valid rev_share",
			partnerConfig: map[string]string{
				models.REVSHARE: "10",
			},
			revshare: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			revshare := GetRevenueShare(tt.partnerConfig)
			assert.Equal(t, tt.revshare, revshare, tt.name)
		})
	}
}

func TestGetGdprEnabledFlag(t *testing.T) {
	tests := []struct {
		name          string
		partnerConfig map[int]map[string]string
		gdprFlag      int
	}{
		{
			name:          "Empty partnerConfig",
			partnerConfig: make(map[int]map[string]string),
			gdprFlag:      0,
		},
		{
			name: "partnerConfig without versionlevel cfg",
			partnerConfig: map[int]map[string]string{
				2: {models.GDPR_ENABLED: "1"},
			},
			gdprFlag: 0,
		},
		{
			name: "partnerConfig without GDPR_ENABLED",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {"any": "1"},
			},
			gdprFlag: 0,
		},
		{
			name: "partnerConfig with invalid GDPR_ENABLED",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {models.GDPR_ENABLED: "non-int"},
			},
			gdprFlag: 0,
		},
		{
			name: "partnerConfig with GDPR_ENABLED=1",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {models.GDPR_ENABLED: "1"},
			},
			gdprFlag: 1,
		},
		{
			name: "partnerConfig with GDPR_ENABLED=2",
			partnerConfig: map[int]map[string]string{
				models.VersionLevelConfigID: {models.GDPR_ENABLED: "2"},
			},
			gdprFlag: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdprFlag := GetGdprEnabledFlag(tt.partnerConfig)
			assert.Equal(t, tt.gdprFlag, gdprFlag, tt.name)
		})
	}
}

func TestGetNetEcpm(t *testing.T) {
	type args struct {
		price, revShare float64
	}
	tests := []struct {
		name    string
		args    args
		netecpm float64
	}{
		{
			name: "revshare is 0",
			args: args{
				revShare: 0,
				price:    10,
			},
			netecpm: 10,
		},
		{
			name: "revshare is int",
			args: args{
				revShare: 2,
				price:    2,
			},
			netecpm: 1.96,
		},
		{
			name: "revshare is float",
			args: args{
				revShare: 3.338,
				price:    100,
			},
			netecpm: 96.66,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netecpm := GetNetEcpm(tt.args.price, tt.args.revShare)
			assert.Equal(t, tt.netecpm, netecpm, tt.name)
		})
	}
}

func TestGetGrossEcpm(t *testing.T) {

	tests := []struct {
		name      string
		price     float64
		grossecpm float64
	}{
		{
			name:      "grossecpm ceiling",
			price:     18.998,
			grossecpm: 19,
		},
		{
			name:      "grossecpm floor",
			price:     18.901,
			grossecpm: 18.90,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grossecpm := GetGrossEcpm(tt.price)
			assert.Equal(t, tt.grossecpm, grossecpm, tt.name)
		})
	}
}

func TestExtractDomain(t *testing.T) {
	type want struct {
		domain string
		err    bool
	}
	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "url without http prefix",
			url:  "google.com",
			want: want{
				domain: "google.com",
			},
		},
		{
			name: "url with http prefix",
			url:  "http://google.com",
			want: want{
				domain: "google.com",
			},
		},
		{
			name: "url with https prefix",
			url:  "https://google.com",
			want: want{
				domain: "google.com",
			},
		},
		{
			name: "invalid",
			url:  "https://google:com?a=1;b=2",
			want: want{
				domain: "",
				err:    true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, err := ExtractDomain(tt.url)
			assert.Equal(t, tt.want.domain, domain, tt.name)
			assert.Equal(t, tt.want.err, err != nil, tt.name)
		})
	}
}
