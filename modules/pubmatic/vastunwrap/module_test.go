package vastunwrap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
	"github.com/prebid/prebid-server/modules/pubmatic/vastunwrap/models"
	"github.com/stretchr/testify/assert"
)

var vastXMLAdM = "<VAST version='3.0'><Ad id='1'><Wrapper><AdSystem>PubMatic</AdSystem><VASTAdTagURI><![CDATA[https://owsdk-stagingams.pubmatic.com:8443/openwrap/video/Shashank/dspResponse/vastInline.php?m=1&x=3&y=3&p=11&va=3&sc=1]]></VASTAdTagURI><Error><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&er=[ERRORCODE]]]></Error><Impression><![CDATA[https://aktrack.pubmatic.com/AdServer/AdDisplayTrackerServlet?operId=1&pubId=64195&siteId=47105&adId=1405154&adType=13&adServerId=243&kefact=1.000000&kaxefact=1.000000&kadNetFrequecy=0&kadwidth=0&kadheight=0&kadsizeid=97&kltstamp=1536933242&indirectAdId=0&adServerOptimizerId=2&ranreq=0.05969169352174375&kpbmtpfact=11.000000&dcId=1&tldId=0&passback=0&svr=ktk57&ekefact=er2bW2sDAwCra06ACbsIQySn5nqBtYsTl8fy5lupAexh37D_&ekaxefact=er2bW4EDAwB_LQpJJ23Fq0DcNC-NSAFXdpSQC8XBk_S33_Fa&ekpbmtpfact=er2bW5MDAwDJHdBnLBt5IrRuh7x0oqp_tjIALv_VvSQDAl6R&crID=m:1_x:3_y:3_p:11_va:3&lpu=ae.com&ucrid=678722001014421372&campaignId=16774&creativeId=0&pctr=0.000000&wDSPByrId=511&wDspId=27&wbId=0&wrId=0&wAdvID=3170&isRTB=1&rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&imprId=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&oid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&pageURL=http%253A%252F%252Fowsdk-stagingams.pubmatic.com%253A8443%252Fvast-validator%252F%2523&sec=1&pmc=1]]></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event='creativeView'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=1]]></Tracking><Tracking event='start'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=2]]></Tracking><Tracking event='midpoint'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=3]]></Tracking><Tracking event='firstQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=4]]></Tracking><Tracking event='thirdQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=5]]></Tracking><Tracking event='complete'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=6]]></Tracking></TrackingEvents><VideoClicks><ClickTracking><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=99]]></ClickTracking></VideoClicks></Linear></Creative><Creative><NonLinearAds><TrackingEvents><Tracking event='creativeView'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=1]]></Tracking><Tracking event='start'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=2]]></Tracking><Tracking event='midpoint'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=3]]></Tracking><Tracking event='firstQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=4]]></Tracking><Tracking event='thirdQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=5]]></Tracking><Tracking event='complete'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=6]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>"

var inlineXMLAdM = "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\" ?><VAST version=\"3.0\"><Ad id=\"1329167\"><InLine><AdSystem>Acudeo Compatible</AdSystem><AdTitle>VAST 2.0 Instream Test 1</AdTitle><Description>VAST 2.0 Instream Test 1</Description><Error>https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;er=[ERRORCODE]</Error><Error>https://track.dsptracker.com?p=1234&amp;er=[ERRORCODE]</Error><Impression>https://aktrack.pubmatic.com/AdServer/AdDisplayTrackerServlet?operId=1&amp;pubId=64195&amp;siteId=47105&amp;adId=1405154&amp;adType=13&amp;adServerId=243&amp;kefact=1.000000&amp;kaxefact=1.000000&amp;kadNetFrequecy=0&amp;kadwidth=0&amp;kadheight=0&amp;kadsizeid=97&amp;kltstamp=1536933242&amp;indirectAdId=0&amp;adServerOptimizerId=2&amp;ranreq=0.05969169352174375&amp;kpbmtpfact=11.000000&amp;dcId=1&amp;tldId=0&amp;passback=0&amp;svr=ktk57&amp;ekefact=er2bW2sDAwCra06ACbsIQySn5nqBtYsTl8fy5lupAexh37D_&amp;ekaxefact=er2bW4EDAwB_LQpJJ23Fq0DcNC-NSAFXdpSQC8XBk_S33_Fa&amp;ekpbmtpfact=er2bW5MDAwDJHdBnLBt5IrRuh7x0oqp_tjIALv_VvSQDAl6R&amp;crID=m:1_x:3_y:3_p:11_va:3&amp;lpu=ae.com&amp;ucrid=678722001014421372&amp;campaignId=16774&amp;creativeId=0&amp;pctr=0.000000&amp;wDSPByrId=511&amp;wDspId=27&amp;wbId=0&amp;wrId=0&amp;wAdvID=3170&amp;isRTB=1&amp;rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&amp;imprId=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;oid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;pageURL=http%253A%252F%252Fowsdk-stagingams.pubmatic.com%253A8443%252Fvast-validator%252F%2523&amp;sec=1&amp;pmc=1</Impression><Impression>https://DspImpressionTracker.com/</Impression><Creatives><Creative AdID=\"1329167\"><Linear skipoffset=\"20%\"><TrackingEvents><Tracking event=\"close\">https://mytracking.com/linear/close</Tracking><Tracking event=\"skip\">https://mytracking.com/linear/skip</Tracking><Tracking event=\"creativeView\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=1</Tracking><Tracking event=\"start\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=2</Tracking><Tracking event=\"midpoint\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=3</Tracking><Tracking event=\"firstQuartile\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=4</Tracking><Tracking event=\"thirdQuartile\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=5</Tracking><Tracking event=\"complete\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=6</Tracking></TrackingEvents><Duration>00:00:04</Duration><VideoClicks><ClickThrough>https://www.automationtester.in</ClickThrough><ClickTracking>https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=99</ClickTracking></VideoClicks><MediaFiles><MediaFile delivery=\"progressive\" type=\"video/mp4\" bitrate=\"500\" width=\"400\" height=\"300\" scalable=\"true\" maintainAspectRatio=\"true\">https://stagingams.pubmatic.com:8443/openwrap/media/pubmatic.mp4</MediaFile><MediaFile delivery=\"progressive\" type=\"video/mp4\" bitrate=\"500\" width=\"400\" height=\"300\" scalable=\"true\" maintainAspectRatio=\"true\">https://stagingams.pubmatic.com:8443/openwrap/media/pubmatic.mp4</MediaFile><MediaFile delivery=\"progressive\" type=\"video/mp4\" bitrate=\"500\" width=\"400\" height=\"300\" scalable=\"true\" maintainAspectRatio=\"true\">https://stagingams.pubmatic.com:8443/openwrap/media/mp4-sample-3.mp4</MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>"

func TestVastUnwrapModuleHandleEntrypointHook(t *testing.T) {
	type fields struct {
		cfg VastUnwrapModule
	}
	type args struct {
		ctx     context.Context
		miCtx   hookstage.ModuleInvocationContext
		payload hookstage.EntrypointPayload
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    hookstage.HookResult[hookstage.EntrypointPayload]
		wantErr bool
	}{
		{
			name: "Vast unwrap is enabled in the config",
			fields: fields{cfg: VastUnwrapModule{Enabled: true, Cfg: unWrapCfg.VastUnWrapCfg{
				HTTPConfig:   unWrapCfg.HttpConfig{MaxIdleConns: 100, MaxIdleConnsPerHost: 1, IdleConnTimeout: 300},
				APPConfig:    unWrapCfg.AppConfig{Host: "", Port: 0, UnwrapDefaultTimeout: 100, Debug: 1},
				StatConfig:   unWrapCfg.StatConfig{Host: "10.172.141.13", Port: 8080, RefershIntervalInSec: 1},
				ServerConfig: unWrapCfg.ServerConfig{ServerName: "", DCName: "OW_DC"},
				LogConfig:    unWrapCfg.LogConfig{ErrorLogFile: "/home/test/PBSlogs/unwrap/error.log", DebugLogFile: "/home/test/PBSlogs/unwrap/debug.log"},
			},
				TrafficPercentage: 2}},
			args: args{
				miCtx: hookstage.ModuleInvocationContext{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{IsVastUnwrapEnabled: true}}},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), isVastUnWrapEnabled, "1")
						r, _ := http.NewRequestWithContext(ctx, "", "", nil)
						return r
					}(),
				}},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{IsVastUnwrapEnabled: true}}},
		},
		{
			name: "Vast unwrap is disabled in the config",
			fields: fields{
				cfg: VastUnwrapModule{Enabled: false, Cfg: unWrapCfg.VastUnWrapCfg{
					HTTPConfig:   unWrapCfg.HttpConfig{MaxIdleConns: 100, MaxIdleConnsPerHost: 1, IdleConnTimeout: 300},
					APPConfig:    unWrapCfg.AppConfig{Host: "", Port: 0, UnwrapDefaultTimeout: 100, Debug: 1},
					StatConfig:   unWrapCfg.StatConfig{Host: "10.172.141.13", Port: 8080, RefershIntervalInSec: 1},
					ServerConfig: unWrapCfg.ServerConfig{ServerName: "", DCName: "OW_DC"},
					LogConfig:    unWrapCfg.LogConfig{ErrorLogFile: "/home/test/PBSlogs/unwrap/error.log", DebugLogFile: "/home/test/PBSlogs/unwrap/debug.log"},
				},
					TrafficPercentage: 2}},
			args: args{
				miCtx: hookstage.ModuleInvocationContext{},
				payload: hookstage.EntrypointPayload{
					Request: func() *http.Request {
						ctx := context.WithValue(context.Background(), isVastUnWrapEnabled, "1")
						r, _ := http.NewRequestWithContext(ctx, "", "", nil)
						return r
					}(),
				}},
			want: hookstage.HookResult[hookstage.EntrypointPayload]{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := VastUnwrapModule{
				Cfg:     tt.fields.cfg.Cfg,
				Enabled: tt.fields.cfg.Enabled,
			}
			got, err := m.HandleEntrypointHook(tt.args.ctx, tt.args.miCtx, tt.args.payload)
			if !assert.NoError(t, err, tt.wantErr) {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VastUnwrapModule.HandleEntrypointHook() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVastUnwrapModuleHandleRawBidderResponseHook(t *testing.T) {
	type fields struct {
		cfg VastUnwrapModule
	}
	type args struct {
		in0     context.Context
		miCtx   hookstage.ModuleInvocationContext
		payload hookstage.RawBidderResponsePayload
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		expectedBids []*adapters.TypedBid
		want         hookstage.HookResult[hookstage.RawBidderResponsePayload]
		wantErr      bool
	}{
		{
			name: "Vast unwrap is enabled in the config",
			fields: fields{cfg: VastUnwrapModule{Enabled: true, Cfg: unWrapCfg.VastUnWrapCfg{
				HTTPConfig:   unWrapCfg.HttpConfig{MaxIdleConns: 100, MaxIdleConnsPerHost: 1, IdleConnTimeout: 300},
				APPConfig:    unWrapCfg.AppConfig{Host: "", Port: 0, UnwrapDefaultTimeout: 100, Debug: 1},
				StatConfig:   unWrapCfg.StatConfig{Host: "10.172.141.13", Port: 8080, RefershIntervalInSec: 1},
				ServerConfig: unWrapCfg.ServerConfig{ServerName: "", DCName: "OW_DC"},
				LogConfig:    unWrapCfg.LogConfig{ErrorLogFile: "/home/test/PBSlogs/unwrap/error.log", DebugLogFile: "/home/test/PBSlogs/unwrap/debug.log"},
			},
				TrafficPercentage: 2}},
			args: args{
				miCtx: hookstage.ModuleInvocationContext{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{IsVastUnwrapEnabled: true}}},
				payload: hookstage.RawBidderResponsePayload{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    "Bid-123",
								ImpID: fmt.Sprintf("div-adunit-%d", 123),
								Price: 2.1,
								AdM:   vastXMLAdM,
								CrID:  "Cr-234",
								W:     100,
								H:     50,
							},
							BidType: "video",
						}},
				}},
			expectedBids: []*adapters.TypedBid{{
				Bid: &openrtb2.Bid{
					ID:    "Bid-123",
					ImpID: fmt.Sprintf("div-adunit-%d", 123),
					Price: 2.1,
					AdM:   vastXMLAdM,
					CrID:  "Cr-234",
					W:     100,
					H:     50,
				},
				BidType: "video",
			}},
			want: hookstage.HookResult[hookstage.RawBidderResponsePayload]{},
		},
		{
			name: "Vast unwrap is disabled in the config",
			fields: fields{cfg: VastUnwrapModule{Enabled: false, Cfg: unWrapCfg.VastUnWrapCfg{
				HTTPConfig:   unWrapCfg.HttpConfig{MaxIdleConns: 100, MaxIdleConnsPerHost: 1, IdleConnTimeout: 300},
				APPConfig:    unWrapCfg.AppConfig{Host: "", Port: 0, UnwrapDefaultTimeout: 100, Debug: 1},
				StatConfig:   unWrapCfg.StatConfig{Host: "10.172.141.13", Port: 8080, RefershIntervalInSec: 1},
				ServerConfig: unWrapCfg.ServerConfig{ServerName: "", DCName: "OW_DC"},
				LogConfig:    unWrapCfg.LogConfig{ErrorLogFile: "/home/test/PBSlogs/unwrap/error.log", DebugLogFile: "/home/test/PBSlogs/unwrap/debug.log"},
			},
				TrafficPercentage: 2}},
			args: args{
				miCtx: hookstage.ModuleInvocationContext{ModuleContext: hookstage.ModuleContext{"rctx": models.RequestCtx{IsVastUnwrapEnabled: false}}},
				payload: hookstage.RawBidderResponsePayload{
					Bids: []*adapters.TypedBid{
						{
							Bid: &openrtb2.Bid{
								ID:    "Bid-123",
								ImpID: fmt.Sprintf("div-adunit-%d", 123),
								Price: 2.1,
								AdM:   "<div>This is an Ad</div>",
								CrID:  "Cr-234",
								W:     100,
								H:     50,
							},
							BidType: "video",
						}},
				}},
			want: hookstage.HookResult[hookstage.RawBidderResponsePayload]{},
			expectedBids: []*adapters.TypedBid{{
				Bid: &openrtb2.Bid{
					ID:    "Bid-123",
					ImpID: fmt.Sprintf("div-adunit-%d", 123),
					Price: 2.1,
					AdM:   "<div>This is an Ad</div>",
					CrID:  "Cr-234",
					W:     100,
					H:     50,
				},
				BidType: "video",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := VastUnwrapModule{
				Cfg:     tt.fields.cfg.Cfg,
				Enabled: tt.fields.cfg.Enabled,
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				data, _ := json.Marshal(tt.expectedBids)
				_, _ = w.Write([]byte(data))
			}))
			defer server.Close()
			server.URL = UnwrapURL
			doUnwrap(tt.args.payload.Bids[0], "test", 1000, server.URL)
			_, err := m.HandleRawBidderResponseHook(tt.args.in0, tt.args.miCtx, tt.args.payload)

			if !assert.NoError(t, err, tt.wantErr) {
				return
			}

			assert.Equal(t, tt.expectedBids[0].Bid.AdM, tt.args.payload.Bids[0].Bid.AdM, "got, tt.want AdM is not updatd correctly after executing RawBidderResponse hook.")
		})
	}
}

func TestInitVastUnrap(t *testing.T) {
	type args struct {
		rawCfg json.RawMessage
		in1    moduledeps.ModuleDeps
	}
	tests := []struct {
		name    string
		args    args
		want    VastUnwrapModule
		wantErr bool
	}{
		{
			name: "Valid vast unwrap config",
			args: args{
				rawCfg: json.RawMessage(`{"enabled":true,"vastunwrapcfg":{"max_wrapper_support":5,"app_config":{"debug":1,"unwrap_default_timeout":100},"http_config":{"idle_conn_timeout":300,"max_idle_conns":100,"max_idle_conns_per_host":1},"log_config":{"debug_log_file":"/home/test/PBSlogs/unwrap/debug.log","error_log_file":"/home/test/PBSlogs/unwrap/error.log"},"server_config":{"dc_name":"OW_DC"},"stat_config":{"host":"10.172.141.13","port":8080,"referesh_interval_in_sec":1}}}`),
				in1:    moduledeps.ModuleDeps{},
			},
			want: VastUnwrapModule{
				Enabled: true,
				Cfg: unWrapCfg.VastUnWrapCfg{
					MaxWrapperSupport: 5,
					HTTPConfig:        unWrapCfg.HttpConfig{MaxIdleConns: 100, MaxIdleConnsPerHost: 1, IdleConnTimeout: 300},
					APPConfig:         unWrapCfg.AppConfig{Host: "", Port: 0, UnwrapDefaultTimeout: 100, Debug: 1},
					StatConfig:        unWrapCfg.StatConfig{Host: "10.172.141.13", Port: 8080, RefershIntervalInSec: 1},
					ServerConfig:      unWrapCfg.ServerConfig{ServerName: "", DCName: "OW_DC"},
					LogConfig:         unWrapCfg.LogConfig{ErrorLogFile: "/home/test/PBSlogs/unwrap/error.log", DebugLogFile: "/home/test/PBSlogs/unwrap/debug.log"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initVastUnrap(tt.args.rawCfg, tt.args.in1)
			if !assert.NoError(t, err, tt.wantErr) {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initVastUnrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder(t *testing.T) {
	type args struct {
		rawCfg json.RawMessage
		deps   moduledeps.ModuleDeps
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Valid vast unwrap config",
			args: args{
				rawCfg: json.RawMessage(`{"enabled":true,"vastunwrapcfg":{"app_config":{"debug":1,"unwrap_default_timeout":100},"max_wrapper_support":5,"http_config":{"idle_conn_timeout":300,"max_idle_conns":100,"max_idle_conns_per_host":1},"log_config":{"debug_log_file":"/home/test/PBSlogs/unwrap/debug.log","error_log_file":"/home/test/PBSlogs/unwrap/error.log"},"server_config":{"dc_name":"OW_DC"},"stat_config":{"host":"10.172.141.13","port":8080,"referesh_interval_in_sec":1}}}`),
				deps:   moduledeps.ModuleDeps{},
			},
			want: VastUnwrapModule{
				Enabled: true,
				Cfg: unWrapCfg.VastUnWrapCfg{
					MaxWrapperSupport: 5,
					HTTPConfig:        unWrapCfg.HttpConfig{MaxIdleConns: 100, MaxIdleConnsPerHost: 1, IdleConnTimeout: 300},
					APPConfig:         unWrapCfg.AppConfig{Host: "", Port: 0, UnwrapDefaultTimeout: 100, Debug: 1},
					StatConfig:        unWrapCfg.StatConfig{Host: "10.172.141.13", Port: 8080, RefershIntervalInSec: 1},
					ServerConfig:      unWrapCfg.ServerConfig{ServerName: "", DCName: "OW_DC"},
					LogConfig:         unWrapCfg.LogConfig{ErrorLogFile: "/home/test/PBSlogs/unwrap/error.log", DebugLogFile: "/home/test/PBSlogs/unwrap/debug.log"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Builder(tt.args.rawCfg, tt.args.deps)
			if !assert.NoError(t, err, tt.wantErr) {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder() = %v, want %v", got, tt.want)
			}
		})
	}
}
