package models

import (
	"fmt"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestErrorWrap(t *testing.T) {
	type args struct {
		cErr error
		nErr error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "current error as nil",
			args: args{
				cErr: nil,
				nErr: fmt.Errorf("error found for %d", 1234),
			},
			wantErr: true,
		},
		{
			name: "wrap error",
			args: args{
				cErr: fmt.Errorf("current error found for %d", 1234),
				nErr: fmt.Errorf("new error found for %d", 1234),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ErrorWrap(tt.args.cErr, tt.args.nErr); (err != nil) != tt.wantErr {
				t.Errorf("ErrorWrap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetCreativeType(t *testing.T) {
	type args struct {
		bid    *openrtb2.Bid
		bidExt *BidExt
		impCtx *ImpCtx
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "bid.ext.prebid.type absent",
			args: args{
				bid: &openrtb2.Bid{},
				bidExt: &BidExt{
					ExtBid: openrtb_ext.ExtBid{
						Prebid: &openrtb_ext.ExtBidPrebid{},
					},
				},
				impCtx: &ImpCtx{},
			},
			want: "",
		},
		{
			name: "bid.ext.prebid.type empty",
			args: args{
				bid: &openrtb2.Bid{},
				bidExt: &BidExt{
					ExtBid: openrtb_ext.ExtBid{
						Prebid: &openrtb_ext.ExtBidPrebid{
							Type: "",
						},
					},
				},
				impCtx: &ImpCtx{},
			},
			want: "",
		},
		{
			name: "bid.ext.prebid.type is banner",
			args: args{
				bid: &openrtb2.Bid{},
				bidExt: &BidExt{
					ExtBid: openrtb_ext.ExtBid{
						Prebid: &openrtb_ext.ExtBidPrebid{
							Type: Banner,
						},
					},
				},
				impCtx: &ImpCtx{},
			},
			want: Banner,
		},
		{
			name: "bid.ext.prebid.type is video",
			args: args{
				bid: &openrtb2.Bid{},
				bidExt: &BidExt{
					ExtBid: openrtb_ext.ExtBid{
						Prebid: &openrtb_ext.ExtBidPrebid{
							Type: Video,
						},
					},
				},
				impCtx: &ImpCtx{},
			},
			want: Video,
		},
		{
			name: "Banner Adm has json assets keyword",
			args: args{
				bid: &openrtb2.Bid{
					AdM: `u003cscript src='mraid.js'></script><script>document.createElement('IMG').src="https://tlx.3lift.com/s2s/notify?px=1&pr=${AUCTION_PRICE}&ts=1686039380&aid=45036723730776858540010&ec=2711_67927_9991141&n=GFABCP4AhSIAwCSAwQwMTNimAMAoAOfqByoAwA%3D";window.tl_auction_response_559942={"settings":{"viewability":{},"additional_data":{"pr":"${AUCTION_PRICE}","bc":"AAABiI_HQr-J9SWLTbGinTR6NNuHz29x102WBw==","aid":"45036723730776858540010","bmid":"2711","biid":"7295","sid":"67927","brid":"82983","adid":"9991141","crid":"737729","ts":"1686039380","bcud":"1240","ss":"20"},"template_id":210,"payable_event":1,"billable_event":1,"billable_pixel":"https:\/\/tlx.3lift.com\/s2s\/notify?px=1&pr=${AUCTION_PRICE}&ts=1686039380&aid=45036723730776858540010&ec=2711_67927_9991141&n=GpIGaCP%3D&b=1","adchoices_url":"https:\/\/optout.aboutads.info\/","format_id":10,"render_options_bm":0,"cta":"Learn more"},"assets":[{"asset_id":0,"cta":"Learn more","banner_width":300,"banner_height":600,"banner_markup":"<script type='text\/javascript' src='https:\/\/ads.as.criteo.com\/delivery\/r\/ajs.php?z=AAABiI_HQr-J9SWLTbGinTR6NNuHz29x102WBw==&u=%7Cgud5gNZYq-lw&ct0={clickurl_enc}'><\/script>"}]};</script><script src="https://ib.3lift.com/ttj?inv_code=HK01_Android_Opening_InListBox_3_336x280&tid=210" data-auction-response-id="559942" data-ss-id="20"></script>`,
				},
				bidExt: &BidExt{
					ExtBid: openrtb_ext.ExtBid{
						Prebid: &openrtb_ext.ExtBidPrebid{},
					},
				},
				impCtx: &ImpCtx{},
			},
			want: Banner,
		},
		{
			name: "Empty Bid Adm",
			args: args{
				bid: &openrtb2.Bid{
					AdM: "",
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: "",
		},
		{
			name: "VAST Ad",
			args: args{
				bid: &openrtb2.Bid{
					AdM: "<VAST version='3.0'><Ad id='601364'><InLine><AdSystem>Acudeo Compatible</AdSystem><AdTitle>VAST 2.0 Instream Test 1</AdTitle><Description>VAST 2.0 Instream Test 1</Description><Impression>http://172.16.4.213/AdServer/AdDisplayTrackerServlet</Impression><Impression>https://dsptracker.com/{PSPM}</Impression><Error>http://172.16.4.213/track</Error><Error>https://Errortrack.com</Error><Creatives><Creative AdID='601364'><Linear skipoffset='20%'><Duration>00:00:04</Duration><VideoClicks><ClickTracking>http://172.16.4.213/track</ClickTracking><ClickThrough>https://www.pubmatic.com</ClickThrough></VideoClicks><MediaFiles><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'>https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]</MediaFile><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'>https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]</MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>",
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: Video,
		},
		{
			name: "VAST Ad xml",
			args: args{
				bid: &openrtb2.Bid{
					AdM: "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?><VAST version=\"4.0\"><Ad id=\"97517771\"><Wrapper><AdSystem version=\"4.0\">adnxs</AdSystem><VASTAdTagURI><![CDATA[http://sin3-ib.adnxs.com/ab?an_audit=0&test=1&referrer=http%3A%2F%2Fprebid.org%2Fexamples%2Fvideo%2Fserver%2Fjwplayer%2Fpbs-ve-jwplayer-hosted.html&e=wqT_3QKUC6CUBQAAAwDWAAUBCObImfIFEN_CrZCNxam9JhjNy6T75_2f4igqNgkAAAECCBRAEQEHNAAAFEAZAAAAwB6FPUAhERIAKREJADERG6gw6dGnBjjtSEDtSEgCUMuBwC5YnPFbYABozbp1eMu4BYABAYoBA1VTRJIBAQbwUpgBAaABAagBAbABALgBA8ABA8gBAtABANgBAOABAfABAIoCO3VmKCdhJywgMjUyOTg4NSwgMTU4MTY3MTUyNik7dWYoJ3InLCA5NzUxNzc3MSwgLh4A9A4BkgLRAiE0RUQzNndpMi1Md0tFTXVCd0M0WUFDQ2M4VnN3QURnQVFBUkk3VWhRNmRHbkJsZ0FZUF9fX184UGFBQndBWGdCZ0FFQmlBRUJrQUVCbUFFQm9BRUJxQUVEc0FFQXVRSHpyV3FrQUFBVVFNRUI4NjFxcEFBQUZFREpBVlBTU2JZNVlPY18yUUVBQUFBQUFBRHdQLUFCQVBVQkFBQUFBSmdDQUtBQ0FMVUNBQUFBQUwwQ0FBQUFBTUFDQWNnQ0FkQUNBZGdDQWVBQ0FPZ0NBUGdDQUlBREFaZ0RBYWdEdHZpOENyb0RDVk5KVGpNNk5EZ3pOdUFEbEJ1SUJBQ1FCQUNZQkFIQkJBQUFBQQmDCHlRUQkJAQEYTmdFQVBFRQELCQEgQ0lCZVFscVFVCQ8YQUR3UDdFRg0NAQEsLpoCiQEheXc3WjlnNlUBJG5QRmJJQVFvQUQVUFRVUURvSlUwbE9Nem8wT0RNMlFKUWJTEYAMUEFfVREMDEFBQVcdDABZHQwAYR0MAGMdDPBSZUFBLsICP2h0dHA6Ly9wcmViaWQub3JnL2Rldi1kb2NzL3Nob3ctdmlkZW8td2l0aC1hLWRmcC12aWRlby10YWcuaHRtbNgCAOACrZhI6gJMaHQ-SgAgZXhhbXBsZXMvBUVcL3NlcnZlci9qd3BsYXllci9wYnMtdmUtERAYLWhvc3RlZAVXNPICEQoGQURWX0lEEgcySbsFFAhDUEcFFBg1NzU5MzY0ARQIBUNQARM0CDIxOTY5OTc08gINCggBPBhGUkVREgEwBRAcUkVNX1VTRVIFEAAMCSAYQ09ERRIA8gEPAVcRDxALCgdDUBUOEAkKBUlPAWAEAPIBGgRJTxUaOBMKD0NVU1RPTV9NT0RFTA0kCBoKFjIWABxMRUFGX05BTQVqCB4KGjYdAAhBU1QBPhBJRklFRAFiHA0KCFNQTElUAU3wgQEwgAMAiAMBkAMAmAMUoAMBqgMAwAPgqAHIAwDYAwDgAwDoAwD4AwOABACSBAkvb3BlbnJ0YjKYBACiBA0xODIuNzQuMzkuMjUwqAQAsgQOCAAQBBiABSDoAjAAOAS4BADABADIBADSBA45MzI1I1NJTjM6NDgzNtoEAggA4AQA8ASBgSCIBQGYBQCgBf8RAbABqgUkZDc0MzQ3ZDUtYzY3Mi00NTM5LWIxNDEtOWVjMWMzMzJiZTI2wAUAyQWJ4hTwP9IFCQkJDHgAANgFAeAFAfAFw5UL-gUECAAQAJAGAZgGALgGAMEGCSUo8D_QBvUv2gYWChAJERkBUBAAGADgBgTyBgIIAIAHAYgHAKAHQA..&s=dcc685e3549971224cbd8615ff729bcb19107ec0&pp=${AUCTION_PRICE}]]></VASTAdTagURI><Impression><![CDATA[http://ib.adnxs.com/nop]]></Impression><Creatives><Creative adID=\"97517771\"><Linear></Linear></Creative></Creatives></Wrapper></Ad></VAST>",
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: Video,
		},
		{
			name: "Banner Ad",
			args: args{
				bid: &openrtb2.Bid{
					AdM: "<span class=\"PubAPIAd\"  id=\"4E733404-CC2E-48A2-BC83-4DD5F38FE9BB\"><script type=\"text/javascript\"> document.writeln('<iframe width=\"300\" scrolling=\"no\" height=\"250\" frameborder=\"0\" name=\"iframe0\" allowtransparency=\"true\" marginheight=\"0\" marginwidth=\"0\" vspace=\"0\" hspace=\"0\" src=\"https://ads.pubmatic.com/AdTag/300x250.png\"></iframe>');</script><iframe width=\"0\" scrolling=\"no\" height=\"0\" frameborder=\"0\" src=\"https://st.pubmatic.com/AdServer/AdDisplayTrackerServlet?pubId=5890\" style=\"position:absolute;top:-15000px;left:-15000px\" vspace=\"0\" hspace=\"0\" marginwidth=\"0\" marginheight=\"0\" allowtransparency=\"true\" name=\"pbeacon\"></iframe></span> <!-- PubMatic Ad Ends --><div style=\"position:absolute;left:0px;top:0px;visibility:hidden;\">",
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: Banner,
		},
		{
			name: "Native Adm with `native` Object",
			args: args{
				bid: &openrtb2.Bid{
					AdM: `{"native":{"ver":1.2,"link":{"url":"https://dummyimage.com/1x1/000000/fff.jpg&text=420x420+Creative.jpg","clicktrackers":["http://image3.pubmatic.com/AdServer/layer?a={PUBMATIC_SECOND_PRICE}&ucrid=9335447642416814892&t=FNOZW09VkdSTVM0eU5BPT09JmlkPTAmY2lkPTIyNzcyJnhwcj0xLjAwMDAwMCZmcD00JnBwPTIuMzcxMiZ0cD0yJnBlPTAuMDA9=","http://image3.pubmatic.com/AdServer/layer?a={PUBMATIC_SECOND_PRICE}&ucrid=9335447642416814892&t=FNOZW09VkdSTVM0eU5BPT09JmlkPTAmY2lkPTIyNzcyJnhwcj0xLjAwMDAwMCZmcD00JnBwPTIuMzcxMiZ0cD0yJnBlPTAuMDA9="]},"eventtrackers":[{"event":1,"method":1,"url":"http://image3.pubmatic.com/AdServer/layer?a={PUBMATIC_SECOND_PRICE}&ucrid=9335447642416814892&t=FNOZW09VkdSTVM0eU5BPT09JmlkPTAmY2lkPTIyNzcyJnhwcj0xLjAwMDAwMCZmcD00JnBwPTIuMzcxMiZ0cD0yJnBlPTAuMDA9="}]}}`,
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{
					Native: &openrtb2.Native{},
				},
			},
			want: Native,
		},
		{
			name: "Native Adm with `native` and `assets` Object",
			args: args{
				bid: &openrtb2.Bid{
					AdM: "{\"native\":{\"assets\":[{\"id\":1,\"required\":0,\"title\":{\"text\":\"Lexus - Luxury vehicles company\"}},{\"id\":2,\"img\":{\"h\":150,\"url\":\"https://stagingnyc.pubmatic.com:8443//sdk/lexus_logo.png\",\"w\":150},\"required\":0},{\"id\":3,\"img\":{\"h\":428,\"url\":\"https://stagingnyc.pubmatic.com:8443//sdk/28f48244cafa0363b03899f267453fe7%20copy.png\",\"w\":214},\"required\":0},{\"data\":{\"value\":\"Goto PubMatic\"},\"id\":4,\"required\":0},{\"data\":{\"value\":\"Lexus - Luxury vehicles company\"},\"id\":5,\"required\":0},{\"data\":{\"value\":\"4\"},\"id\":6,\"required\":0}],\"imptrackers\":[\"http://phtrack.pubmatic.com/?ts=1496043362&r=84137f17-eefd-4f06-8380-09138dc616e6&i=c35b1240-a0b3-4708-afca-54be95283c61&a=130917&t=9756&au=10002949&p=&c=10014299&o=10002476&wl=10009731&ty=1\"],\"link\":{\"clicktrackers\":[\"http://ct.pubmatic.com/track?ts=1496043362&r=84137f17-eefd-4f06-8380-09138dc616e6&i=c35b1240-a0b3-4708-afca-54be95283c61&a=130917&t=9756&au=10002949&p=&c=10014299&o=10002476&wl=10009731&ty=3&url=\"],\"url\":\"http://www.lexus.com/\"},\"ver\":1}}",
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{
					Native: &openrtb2.Native{},
				},
			},
			want: Native,
		},
		{
			name: "Native Adm with `native` Object but Native is missing in impCtx",
			args: args{
				bid: &openrtb2.Bid{
					AdM: `{"native":{"ver":1.2,"link":{"url":"https://dummyimage.com/1x1/000000/fff.jpg&text=420x420+Creative.jpg","clicktrackers":["http://image3.pubmatic.com/AdServer/layer?a={PUBMATIC_SECOND_PRICE}&ucrid=9335447642416814892&t=FNOZW09VkdSTVM0eU5BPT09JmlkPTAmY2lkPTIyNzcyJnhwcj0xLjAwMDAwMCZmcD00JnBwPTIuMzcxMiZ0cD0yJnBlPTAuMDA9=","http://image3.pubmatic.com/AdServer/layer?a={PUBMATIC_SECOND_PRICE}&ucrid=9335447642416814892&t=FNOZW09VkdSTVM0eU5BPT09JmlkPTAmY2lkPTIyNzcyJnhwcj0xLjAwMDAwMCZmcD00JnBwPTIuMzcxMiZ0cD0yJnBlPTAuMDA9="]},"eventtrackers":[{"event":1,"method":1,"url":"http://image3.pubmatic.com/AdServer/layer?a={PUBMATIC_SECOND_PRICE}&ucrid=9335447642416814892&t=FNOZW09VkdSTVM0eU5BPT09JmlkPTAmY2lkPTIyNzcyJnhwcj0xLjAwMDAwMCZmcD00JnBwPTIuMzcxMiZ0cD0yJnBlPTAuMDA9="}]}}`,
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: Banner,
		},
		{
			name: "Video Adm \t",
			args: args{
				bid: &openrtb2.Bid{
					AdM: "<VAST\t></VAST>",
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: Video,
		},
		{
			name: "Video Adm \r",
			args: args{
				bid: &openrtb2.Bid{
					AdM: "<VAST\r></VAST>",
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: Video,
		},
		{
			name: "Video AdM \n",
			args: args{
				bid: &openrtb2.Bid{
					AdM: "<VAST\n></VAST>",
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: Video,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creativeType := GetCreativeType(tt.args.bid, tt.args.bidExt, tt.args.impCtx)
			assert.Equal(t, tt.want, creativeType, tt.name)
		})
	}
}

func TestGetAdFormat(t *testing.T) {
	type args struct {
		bid    *openrtb2.Bid
		bidExt *BidExt
		impCtx *ImpCtx
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no bid object",
			args: args{
				bid:    nil,
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: "",
		},
		{
			name: "no bidExt object",
			args: args{
				bid:    nil,
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: "",
		},
		{
			name: "no impctx object",
			args: args{
				bid:    nil,
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: "",
		},
		{
			name: "Default bid and banner present",
			args: args{
				bid: &openrtb2.Bid{
					DealID: "",
					Price:  0,
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{
					Banner: true,
				},
			},
			want: Banner,
		},
		{
			name: "Default bid and video present",
			args: args{
				bid: &openrtb2.Bid{
					DealID: "",
					Price:  0,
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{
					Video: &openrtb2.Video{},
				},
			},
			want: Video,
		},
		{
			name: "Default bid and native present",
			args: args{
				bid: &openrtb2.Bid{
					DealID: "",
					Price:  0,
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{
					Native: &openrtb2.Native{},
				},
			},
			want: Native,
		},
		{
			name: "Default bid and banner and video and native present",
			args: args{
				bid: &openrtb2.Bid{
					DealID: "",
					Price:  0,
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{
					Banner: true,
					Native: &openrtb2.Native{},
					Video:  &openrtb2.Video{},
				},
			},
			want: Banner,
		},
		{
			name: "Default bid and none of banner/video/native present",
			args: args{
				bid: &openrtb2.Bid{
					DealID: "",
					Price:  0,
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: "",
		},
		{
			name: "Empty Bid Adm",
			args: args{
				bid: &openrtb2.Bid{
					Price: 10,
					AdM:   "",
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: "",
		},
		{
			name: "VAST Ad",
			args: args{
				bid: &openrtb2.Bid{
					DealID: "dl",
					AdM:    "<VAST version='3.0'><Ad id='601364'><InLine><AdSystem>Acudeo Compatible</AdSystem><AdTitle>VAST 2.0 Instream Test 1</AdTitle><Description>VAST 2.0 Instream Test 1</Description><Impression>http://172.16.4.213/AdServer/AdDisplayTrackerServlet</Impression><Impression>https://dsptracker.com/{PSPM}</Impression><Error>http://172.16.4.213/track</Error><Error>https://Errortrack.com</Error><Creatives><Creative AdID='601364'><Linear skipoffset='20%'><Duration>00:00:04</Duration><VideoClicks><ClickTracking>http://172.16.4.213/track</ClickTracking><ClickThrough>https://www.pubmatic.com</ClickThrough></VideoClicks><MediaFiles><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'>https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-1.mp4]</MediaFile><MediaFile delivery='progressive' type='video/mp4' bitrate='500' width='400' height='300' scalable='true' maintainAspectRatio='true'>https://stagingnyc.pubmatic.com:8443/video/Shashank/mediaFileHost/media/mp4-sample-2.mp4]</MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>",
				},
				bidExt: &BidExt{},
				impCtx: &ImpCtx{},
			},
			want: Video,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creativeType := GetAdFormat(tt.args.bid, tt.args.bidExt, tt.args.impCtx)
			assert.Equal(t, tt.want, creativeType, tt.name)
		})
	}
}
