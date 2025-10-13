package openwrap

import (
	"fmt"
	"net/http"
	"testing"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters"
	"github.com/prebid/prebid-server/v3/hooks/hookstage"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	mock_metrics "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/unwrap"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

var vastXMLAdM = "<VAST version='3.0'><Ad id='1'><Wrapper><AdSystem>PubMatic</AdSystem><VASTAdTagURI><![CDATA[https://owsdk-stagingams.pubmatic.com:8443/openwrap/video/Shashank/dspResponse/vastInline.php?m=1&x=3&y=3&p=11&va=3&sc=1]]></VASTAdTagURI><Error><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&er=[ERRORCODE]]]></Error><Impression><![CDATA[https://aktrack.pubmatic.com/AdServer/AdDisplayTrackerServlet?operId=1&pubId=64195&siteId=47105&adId=1405154&adType=13&adServerId=243&kefact=1.000000&kaxefact=1.000000&kadNetFrequecy=0&kadwidth=0&kadheight=0&kadsizeid=97&kltstamp=1536933242&indirectAdId=0&adServerOptimizerId=2&ranreq=0.05969169352174375&kpbmtpfact=11.000000&dcId=1&tldId=0&passback=0&svr=ktk57&ekefact=er2bW2sDAwCra06ACbsIQySn5nqBtYsTl8fy5lupAexh37D_&ekaxefact=er2bW4EDAwB_LQpJJ23Fq0DcNC-NSAFXdpSQC8XBk_S33_Fa&ekpbmtpfact=er2bW5MDAwDJHdBnLBt5IrRuh7x0oqp_tjIALv_VvSQDAl6R&crID=m:1_x:3_y:3_p:11_va:3&lpu=ae.com&ucrid=678722001014421372&campaignId=16774&creativeId=0&pctr=0.000000&wDSPByrId=511&wDspId=27&wbId=0&wrId=0&wAdvID=3170&isRTB=1&rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&imprId=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&oid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&pageURL=http%253A%252F%252Fowsdk-stagingams.pubmatic.com%253A8443%252Fvast-validator%252F%2523&sec=1&pmc=1]]></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event='creativeView'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=1]]></Tracking><Tracking event='start'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=2]]></Tracking><Tracking event='midpoint'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=3]]></Tracking><Tracking event='firstQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=4]]></Tracking><Tracking event='thirdQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=5]]></Tracking><Tracking event='complete'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=6]]></Tracking></TrackingEvents><VideoClicks><ClickTracking><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=99]]></ClickTracking></VideoClicks></Linear></Creative><Creative><NonLinearAds><TrackingEvents><Tracking event='creativeView'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=1]]></Tracking><Tracking event='start'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=2]]></Tracking><Tracking event='midpoint'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=3]]></Tracking><Tracking event='firstQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=4]]></Tracking><Tracking event='thirdQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=5]]></Tracking><Tracking event='complete'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=6]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>"
var invalidVastXMLAdM = "<VAST version='3.0'><Ad id='1'><AdSystem>PubMatic</AdSystem><VASTAdTagURI><![CDATA[https://owsdk-stagingams.pubmatic.com:8443/openwrap/video/Shashank/dspResponse/vastInline.php?m=1&x=3&y=3&p=11&va=3&sc=1]]></VASTAdTagURI><Error><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&er=[ERRORCODE]]]></Error><Impression><![CDATA[https://aktrack.pubmatic.com/AdServer/AdDisplayTrackerServlet?operId=1&pubId=64195&siteId=47105&adId=1405154&adType=13&adServerId=243&kefact=1.000000&kaxefact=1.000000&kadNetFrequecy=0&kadwidth=0&kadheight=0&kadsizeid=97&kltstamp=1536933242&indirectAdId=0&adServerOptimizerId=2&ranreq=0.05969169352174375&kpbmtpfact=11.000000&dcId=1&tldId=0&passback=0&svr=ktk57&ekefact=er2bW2sDAwCra06ACbsIQySn5nqBtYsTl8fy5lupAexh37D_&ekaxefact=er2bW4EDAwB_LQpJJ23Fq0DcNC-NSAFXdpSQC8XBk_S33_Fa&ekpbmtpfact=er2bW5MDAwDJHdBnLBt5IrRuh7x0oqp_tjIALv_VvSQDAl6R&crID=m:1_x:3_y:3_p:11_va:3&lpu=ae.com&ucrid=678722001014421372&campaignId=16774&creativeId=0&pctr=0.000000&wDSPByrId=511&wDspId=27&wbId=0&wrId=0&wAdvID=3170&isRTB=1&rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&imprId=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&oid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&pageURL=http%253A%252F%252Fowsdk-stagingams.pubmatic.com%253A8443%252Fvast-validator%252F%2523&sec=1&pmc=1]]></Impression><Creatives><Creative><Linear><TrackingEvents><Tracking event='creativeView'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=1]]></Tracking><Tracking event='start'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=2]]></Tracking><Tracking event='midpoint'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=3]]></Tracking><Tracking event='firstQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=4]]></Tracking><Tracking event='thirdQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=5]]></Tracking><Tracking event='complete'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=6]]></Tracking></TrackingEvents><VideoClicks><ClickTracking><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=99]]></ClickTracking></VideoClicks></Linear></Creative><Creative><NonLinearAds><TrackingEvents><Tracking event='creativeView'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=1]]></Tracking><Tracking event='start'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=2]]></Tracking><Tracking event='midpoint'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=3]]></Tracking><Tracking event='firstQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=4]]></Tracking><Tracking event='thirdQuartile'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=5]]></Tracking><Tracking event='complete'><![CDATA[https://aktrack.pubmatic.com/track?operId=7&p=64195&s=47105&a=1405154&wa=243&ts=1536933242&wc=16774&crId=m:1_x:3_y:3_p:11_va:3&ucrid=678722001014421372&impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&advertiser_id=3170&ecpm=1.000000&e=6]]></Tracking></TrackingEvents></NonLinearAds></Creative></Creatives></Wrapper></Ad></VAST>"
var inlineXMLAdM = "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\" ?><VAST version=\"3.0\"><Ad id=\"1329167\"><InLine><AdSystem>Acudeo Compatible</AdSystem><AdTitle>VAST 2.0 Instream Test 1</AdTitle><Description>VAST 2.0 Instream Test 1</Description><Error>https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;er=[ERRORCODE]</Error><Error>https://track.dsptracker.com?p=1234&amp;er=[ERRORCODE]</Error><Impression>https://aktrack.pubmatic.com/AdServer/AdDisplayTrackerServlet?operId=1&amp;pubId=64195&amp;siteId=47105&amp;adId=1405154&amp;adType=13&amp;adServerId=243&amp;kefact=1.000000&amp;kaxefact=1.000000&amp;kadNetFrequecy=0&amp;kadwidth=0&amp;kadheight=0&amp;kadsizeid=97&amp;kltstamp=1536933242&amp;indirectAdId=0&amp;adServerOptimizerId=2&amp;ranreq=0.05969169352174375&amp;kpbmtpfact=11.000000&amp;dcId=1&amp;tldId=0&amp;passback=0&amp;svr=ktk57&amp;ekefact=er2bW2sDAwCra06ACbsIQySn5nqBtYsTl8fy5lupAexh37D_&amp;ekaxefact=er2bW4EDAwB_LQpJJ23Fq0DcNC-NSAFXdpSQC8XBk_S33_Fa&amp;ekpbmtpfact=er2bW5MDAwDJHdBnLBt5IrRuh7x0oqp_tjIALv_VvSQDAl6R&amp;crID=m:1_x:3_y:3_p:11_va:3&amp;lpu=ae.com&amp;ucrid=678722001014421372&amp;campaignId=16774&amp;creativeId=0&amp;pctr=0.000000&amp;wDSPByrId=511&amp;wDspId=27&amp;wbId=0&amp;wrId=0&amp;wAdvID=3170&amp;isRTB=1&amp;rtbId=EBCA079F-8D7C-45B8-B733-92951F670AA1&amp;imprId=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;oid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;pageURL=http%253A%252F%252Fowsdk-stagingams.pubmatic.com%253A8443%252Fvast-validator%252F%2523&amp;sec=1&amp;pmc=1</Impression><Impression>https://DspImpressionTracker.com/</Impression><Creatives><Creative AdID=\"1329167\"><Linear skipoffset=\"20%\"><TrackingEvents><Tracking event=\"close\">https://mytracking.com/linear/close</Tracking><Tracking event=\"skip\">https://mytracking.com/linear/skip</Tracking><Tracking event=\"creativeView\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=1</Tracking><Tracking event=\"start\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=2</Tracking><Tracking event=\"midpoint\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=3</Tracking><Tracking event=\"firstQuartile\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=4</Tracking><Tracking event=\"thirdQuartile\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=5</Tracking><Tracking event=\"complete\">https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=6</Tracking></TrackingEvents><Duration>00:00:04</Duration><VideoClicks><ClickThrough>https://www.automationtester.in</ClickThrough><ClickTracking>https://aktrack.pubmatic.com/track?operId=7&amp;p=64195&amp;s=47105&amp;a=1405154&amp;wa=243&amp;ts=1536933242&amp;wc=16774&amp;crId=m:1_x:3_y:3_p:11_va:3&amp;ucrid=678722001014421372&amp;impid=24D9FEDA-C97D-4DF7-B747-BD3CFF5AC7B5&amp;advertiser_id=3170&amp;ecpm=1.000000&amp;e=99</ClickTracking></VideoClicks><MediaFiles><MediaFile delivery=\"progressive\" type=\"video/mp4\" bitrate=\"500\" width=\"400\" height=\"300\" scalable=\"true\" maintainAspectRatio=\"true\">https://stagingams.pubmatic.com:8443/openwrap/media/pubmatic.mp4</MediaFile><MediaFile delivery=\"progressive\" type=\"video/mp4\" bitrate=\"500\" width=\"400\" height=\"300\" scalable=\"true\" maintainAspectRatio=\"true\">https://stagingams.pubmatic.com:8443/openwrap/media/pubmatic.mp4</MediaFile><MediaFile delivery=\"progressive\" type=\"video/mp4\" bitrate=\"500\" width=\"400\" height=\"300\" scalable=\"true\" maintainAspectRatio=\"true\">https://stagingams.pubmatic.com:8443/openwrap/media/mp4-sample-3.mp4</MediaFile></MediaFiles></Linear></Creative></Creatives></InLine></Ad></VAST>"

func TestHandleRawBidderResponseHook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMetricsEngine := mock_metrics.NewMockMetricsEngine(ctrl)

	type args struct {
		module              OpenWrap
		payload             hookstage.RawBidderResponsePayload
		moduleInvocationCtx hookstage.ModuleInvocationContext
		isAdmUpdated        bool
	}
	tests := []struct {
		name           string
		args           args
		wantResult     hookstage.HookResult[hookstage.RawBidderResponsePayload]
		setup          func()
		wantSeatNonBid openrtb_ext.SeatNonBidBuilder
		mockHandler    http.HandlerFunc
		wantBids       []*adapters.TypedBid
	}{
		{
			name: "Empty_Request_Context",
			args: args{
				module: OpenWrap{
					cfg: config.Config{VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
						MaxWrapperSupport: 5,
						StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
						APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
					}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
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
							}}}},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.NewModuleContext()},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{DebugMessages: []string{"error: request-ctx not found in handleRawBidderResponseHook()"}},
		},
		{
			name: "VASTUnwrap_Disabled_Video_Bids",
			args: args{
				module: OpenWrap{
					cfg: config.Config{VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
						MaxWrapperSupport: 5,
						StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
						APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
					}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
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
							}}}},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: false}})
					return moduleCtx
				}()},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
		},
		{
			name: "VASTUnwrap_Enabled_Single_Video_Bid_Invalid_Vast_xml",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapPercent: 50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
						Bids: []*adapters.TypedBid{
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-123",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   invalidVastXMLAdM,
									CrID:  "Cr-234",
									W:     100,
									H:     50,
								},
								BidType: "video",
							},
						},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
			},
			mockHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("unwrap-status", "1")
				w.WriteHeader(http.StatusNoContent)
			}),
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "1")
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
			},
			wantSeatNonBid: openrtb_ext.SeatNonBidBuilder{},
			wantBids: []*adapters.TypedBid{
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   invalidVastXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
			},
		},
		{
			name: "VASTUnwrap_Enabled_Single_Video_Bid",
			mockHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("unwrap-status", "0")
				w.Header().Add("unwrap-count", "1")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(inlineXMLAdM))
			}),
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapPercent: 50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
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
							},
						},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
				isAdmUpdated: true,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0")
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "1")
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "1", gomock.Any())
			},
			wantSeatNonBid: openrtb_ext.SeatNonBidBuilder{},
			wantBids: []*adapters.TypedBid{
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   inlineXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
			},
		},
		{
			name: "VASTUnwrap_Enabled_Multiple_Video_Bids",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapPercent: 100,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
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
							},
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-456",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   vastXMLAdM,
									CrID:  "Cr-789",
									W:     100,
									H:     50,
								},
								BidType: "video",
							}},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
				isAdmUpdated: true,
			},
			mockHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("unwrap-status", "0")
				w.Header().Add("unwrap-count", "1")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(inlineXMLAdM))
			}),
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0").Times(2)
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "1").Times(2)
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any()).Times(2)
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "1", gomock.Any()).Times(2)
			},
			wantSeatNonBid: openrtb_ext.SeatNonBidBuilder{},
			wantBids: []*adapters.TypedBid{
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-456",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   inlineXMLAdM,
						CrID:  "Cr-789",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   inlineXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
			},
		},
		{
			name: "VASTUnwrap_Enabled_Video_and_Banner_Bids",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapPercent: 50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
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
							},
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-456",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   "This is banner creative",
									CrID:  "Cr-789",
									W:     100,
									H:     50,
								},
								BidType: "banner",
							}},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
				isAdmUpdated: true,
			},
			mockHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("unwrap-status", "0")
				w.Header().Add("unwrap-count", "0")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(inlineXMLAdM))
			}),
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0")
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "0")
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "0", gomock.Any())
			},
			wantSeatNonBid: openrtb_ext.SeatNonBidBuilder{},
			wantBids: []*adapters.TypedBid{
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-456",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   "This is banner creative",
						CrID:  "Cr-789",
						W:     100,
						H:     50,
					},
					BidType: "banner",
				},
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   inlineXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
			},
		},
		{
			name: "VASTUnwrap_Enabled_Video_and_Native_Bids",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapPercent: 50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
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
							},
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-456",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   "This is native creative",
									CrID:  "Cr-789",
									W:     100,
									H:     50,
								},
								BidType: "native",
							}},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
				isAdmUpdated: true,
			},
			mockHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("unwrap-status", "0")
				w.Header().Add("unwrap-count", "0")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(inlineXMLAdM))
			}),
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0")
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "0")
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "0", gomock.Any())
			},
			wantSeatNonBid: openrtb_ext.SeatNonBidBuilder{},
			wantBids: []*adapters.TypedBid{
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-456",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   "This is native creative",
						CrID:  "Cr-789",
						W:     100,
						H:     50,
					},
					BidType: "native",
				},
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   inlineXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
			},
		},
		{
			name: "VASTUnwrap_Enabled_Single_Video_bid_and_source_owsdk",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapPercent: 50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
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
							},
						},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
				isAdmUpdated: true,
			},
			mockHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("unwrap-status", "0")
				w.Header().Add("unwrap-count", "1")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(inlineXMLAdM))
			}),
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0")
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "1")
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "1", gomock.Any())
			},
			wantSeatNonBid: openrtb_ext.SeatNonBidBuilder{},
			wantBids: []*adapters.TypedBid{
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   inlineXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
			},
		},
		{
			name: "VASTUnwrap_Enabled_Native_and_Banner_Bids",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapPercent: 50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
						Bids: []*adapters.TypedBid{
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-123",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   "This is banner creative",
									CrID:  "Cr-234",
									W:     100,
									H:     50,
								},
								BidType: "banner",
							},
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-456",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   "This is native creative",
									CrID:  "Cr-789",
									W:     100,
									H:     50,
								},
								BidType: "native",
							}},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
			},
			wantResult:     hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			wantSeatNonBid: openrtb_ext.SeatNonBidBuilder{},
			wantBids: []*adapters.TypedBid{
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   "This is banner creative",
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "banner",
				},
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-456",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   "This is native creative",
						CrID:  "Cr-789",
						W:     100,
						H:     50,
					},
					BidType: "native",
				},
			},
		},
		{
			name: "bid_with_InvalidVAST_should_be_discarded_and_should_be_present_in_seatNonBid",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapPercent: 50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
						Bids: []*adapters.TypedBid{
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-123",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   invalidVastXMLAdM,
									CrID:  "Cr-234",
									W:     100,
									H:     50,
								},
								BidType: "video",
							},
						},
						BidderAlias: "pubmatic2",
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
			},
			mockHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("unwrap-status", models.UnwrapInvalidVASTStatus)
				w.WriteHeader(http.StatusNoContent)
			}),
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", models.UnwrapInvalidVASTStatus)
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
			},
			wantBids: []*adapters.TypedBid{},
			wantSeatNonBid: func() openrtb_ext.SeatNonBidBuilder {
				seatNonBid := openrtb_ext.SeatNonBidBuilder{}
				seatNonBid.AddBid(openrtb_ext.NonBid{
					ImpId:      fmt.Sprintf("div-adunit-%d", 123),
					StatusCode: int(nbr.LossBidLostInVastUnwrap),
					Ext: openrtb_ext.ExtNonBid{
						Prebid: openrtb_ext.ExtNonBidPrebid{
							Bid: openrtb_ext.ExtNonBidPrebidBid{
								Price: 2.1,
								ID:    "Bid-123",
								W:     100,
								H:     50,
								Type:  openrtb_ext.BidTypeVideo,
							},
						},
					},
				}, "pubmatic2")
				return seatNonBid
			}(),
		},
		{
			name: "bid_with_EmptyVAST_should_be_discarded_and_should_be_present_in_seatNonBid",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapPercent: 50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
						Bids: []*adapters.TypedBid{
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-123",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   invalidVastXMLAdM,
									CrID:  "Cr-234",
									W:     100,
									H:     50,
								},
								BidType: "video",
							},
						},
						BidderAlias: "pubmatic2",
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
			},
			mockHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("unwrap-status", models.UnwrapEmptyVASTStatus)
				w.WriteHeader(http.StatusNoContent)
			}),
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", models.UnwrapEmptyVASTStatus)
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
			},
			wantBids: []*adapters.TypedBid{},
			wantSeatNonBid: func() openrtb_ext.SeatNonBidBuilder {
				seatNonBid := openrtb_ext.SeatNonBidBuilder{}
				seatNonBid.AddBid(openrtb_ext.NonBid{
					ImpId:      fmt.Sprintf("div-adunit-%d", 123),
					StatusCode: int(nbr.LossBidLostInVastUnwrap),
					Ext: openrtb_ext.ExtNonBid{
						Prebid: openrtb_ext.ExtNonBidPrebid{
							Bid: openrtb_ext.ExtNonBidPrebidBid{
								Price: 2.1,
								ID:    "Bid-123",
								W:     100,
								H:     50,
								Type:  openrtb_ext.BidTypeVideo,
							},
						},
					},
				}, "pubmatic2")
				return seatNonBid
			}(),
		},
		{
			name: "VASTUnwrap_Disabled_Video_Bids_Valid_XML",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
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
							}}},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: false}})
					return moduleCtx
				}()},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			wantBids: []*adapters.TypedBid{
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
				},
			},
		},
		{
			name: "VASTUnwrap_Disabled_Video_and_Banner_Bids",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
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
							},
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-456",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   "This is banner creative",
									CrID:  "Cr-789",
									W:     100,
									H:     50,
								},
								BidType: "banner",
							}},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: false}})
					return moduleCtx
				}()},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			wantBids: []*adapters.TypedBid{
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
				},
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-456",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   "This is banner creative",
						CrID:  "Cr-789",
						W:     100,
						H:     50,
					},
					BidType: "banner",
				},
			},
		},
		{
			name: "VASTUnwrap_Disabled_Banner_Bids",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
						Bids: []*adapters.TypedBid{
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-456",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   "This is banner creative",
									CrID:  "Cr-789",
									W:     100,
									H:     50,
								},
								BidType: "banner",
							}},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: false}})
					return moduleCtx
				}()},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			wantBids: []*adapters.TypedBid{
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-456",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   "This is banner creative",
						CrID:  "Cr-789",
						W:     100,
						H:     50,
					},
					BidType: "banner",
				},
			},
		},
		{
			name: "VASTUnwrap_Enabled_Banner_Bids",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
						Bids: []*adapters.TypedBid{
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-456",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   "This is banner creative",
									CrID:  "Cr-789",
									W:     100,
									H:     50,
								},
								BidType: "banner",
							}},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
			},
			wantResult:     hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			wantSeatNonBid: openrtb_ext.SeatNonBidBuilder{},
			wantBids: []*adapters.TypedBid{
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-456",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   "This is banner creative",
						CrID:  "Cr-789",
						W:     100,
						H:     50,
					},
					BidType: "banner",
				},
			},
		},
		{
			name: "VASTUnwrap_Enabled_Invalid_Video_and_Banner_Bids",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						Features: config.FeatureToggle{
							VASTUnwrapPercent: 50,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
						Bids: []*adapters.TypedBid{
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-123",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   invalidVastXMLAdM,
									CrID:  "Cr-234",
									W:     100,
									H:     50,
								},
								BidType: "video",
							},
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-456",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   "This is banner creative",
									CrID:  "Cr-789",
									W:     100,
									H:     50,
								},
								BidType: "banner",
							},
						},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
			},
			mockHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("unwrap-status", "1")
				w.WriteHeader(http.StatusNoContent)
			}),
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "1")
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
			},
			wantSeatNonBid: openrtb_ext.SeatNonBidBuilder{},
			wantBids: []*adapters.TypedBid{
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-456",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   "This is banner creative",
						CrID:  "Cr-789",
						W:     100,
						H:     50,
					},
					BidType: "banner",
				},
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   invalidVastXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
			},
		},
		{
			name: "VASTUnwrap_Enabled_valid_Video_and_Banner_Bids_update_bidtype",
			args: args{
				module: OpenWrap{
					cfg: config.Config{
						ResponseOverride: config.ResponseOverride{BidType: []string{"pubmatic"}},
						Features: config.FeatureToggle{
							VASTUnwrapPercent: 100,
						},
						VastUnwrapCfg: unWrapCfg.VastUnWrapCfg{
							MaxWrapperSupport: 5,
							StatConfig:        unWrapCfg.StatConfig{Endpoint: "http://10.172.141.13:8080", PublishInterval: 1},
							APPConfig:         unWrapCfg.AppConfig{UnwrapDefaultTimeout: 1500},
						}},
					metricEngine: mockMetricsEngine,
				},
				payload: hookstage.RawBidderResponsePayload{
					BidderResponse: &adapters.BidderResponse{
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
								BidType: "native",
							},
							{
								Bid: &openrtb2.Bid{
									ID:    "Bid-456",
									ImpID: fmt.Sprintf("div-adunit-%d", 123),
									Price: 2.1,
									AdM:   "This is banner creative",
									CrID:  "Cr-789",
									W:     100,
									H:     50,
								},
								BidType: "video",
							},
						},
					},
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: func() *hookstage.ModuleContext {
					moduleCtx := hookstage.NewModuleContext()
					moduleCtx.Set("rctx", models.RequestCtx{VastUnWrap: models.VastUnWrap{Enabled: true}})
					return moduleCtx
				}()},
			},
			mockHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("unwrap-status", "0")
				w.Header().Add("unwrap-count", "0")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(inlineXMLAdM))
			}),
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0")
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "0")
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "0", gomock.Any())
			},
			wantSeatNonBid: openrtb_ext.SeatNonBidBuilder{},
			wantBids: []*adapters.TypedBid{
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-456",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   "This is banner creative",
						CrID:  "Cr-789",
						W:     100,
						H:     50,
					},
					BidType: "banner",
				},
				{
					Bid: &openrtb2.Bid{
						ID:    "Bid-123",
						ImpID: fmt.Sprintf("div-adunit-%d", 123),
						Price: 2.1,
						AdM:   inlineXMLAdM,
						CrID:  "Cr-234",
						W:     100,
						H:     50,
					},
					BidType: "video",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			m := tt.args.module
			m.unwrap = unwrap.NewUnwrap("http://localhost:8001/unwrap", 200, tt.mockHandler, m.metricEngine)
			hookResult, _ := m.handleRawBidderResponseHook(tt.args.moduleInvocationCtx, tt.args.payload)
			if tt.args.moduleInvocationCtx.ModuleContext != nil && tt.args.isAdmUpdated {
				assert.Equal(t, inlineXMLAdM, tt.args.payload.BidderResponse.Bids[0].Bid.AdM, "AdM is not updated correctly after executing RawBidderResponse hook.")
			}
			for _, mut := range hookResult.ChangeSet.Mutations() {
				newPayload, err := mut.Apply(tt.args.payload)
				assert.NoError(t, err)
				tt.args.payload = newPayload
			}

			if tt.wantBids != nil {
				assert.ElementsMatch(t, tt.wantBids, tt.args.payload.BidderResponse.Bids, "Mismatched response bids")
			}

			assert.Equal(t, tt.wantSeatNonBid, hookResult.SeatNonBid, "mismatched seatNonBids")
		})
	}
}

func TestIsEligibleForUnwrap(t *testing.T) {
	type args struct {
		result rawBidderResponseHookResult
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "bid_is_nil",
			args: args{result: rawBidderResponseHookResult{bid: nil}},
			want: false,
		},
		{
			name: "bid_bid_is_nil",
			args: args{result: rawBidderResponseHookResult{bid: &adapters.TypedBid{Bid: nil}}},
			want: false,
		},
		{
			name: "empty_adm",
			args: args{result: rawBidderResponseHookResult{bid: &adapters.TypedBid{Bid: &openrtb2.Bid{AdM: ""}}}},
			want: false,
		},
		{
			name: "bidType_video",
			args: args{result: rawBidderResponseHookResult{bid: &adapters.TypedBid{Bid: &openrtb2.Bid{AdM: "some_adm"}}, bidtype: openrtb_ext.BidTypeBanner}},
			want: false,
		},
		{
			name: "bid_is_eligible_for_unwrap",
			args: args{result: rawBidderResponseHookResult{bid: &adapters.TypedBid{Bid: &openrtb2.Bid{AdM: "some_adm"}}, bidtype: openrtb_ext.BidTypeVideo}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEligibleForUnwrap(tt.args.result)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUpdateCreativeType(t *testing.T) {
	tests := []struct {
		name     string
		result   *rawBidderResponseHookResult
		expected *rawBidderResponseHookResult
	}{
		{
			name: "bidder_not_in_list",
			result: &rawBidderResponseHookResult{
				bid:     &adapters.TypedBid{Bid: &openrtb2.Bid{ID: "1", Ext: []byte(`{}`)}},
				bidtype: openrtb_ext.BidTypeBanner,
				bidExt:  []byte(`{}`),
			},
			expected: &rawBidderResponseHookResult{
				bid:     &adapters.TypedBid{Bid: &openrtb2.Bid{ID: "1", Ext: []byte(`{}`)}},
				bidtype: openrtb_ext.BidTypeBanner,
				bidExt:  []byte(`{}`),
			},
		},
		{
			name: "bidder_in_list_no_creative_type",
			result: &rawBidderResponseHookResult{
				bid:     &adapters.TypedBid{Bid: &openrtb2.Bid{ID: "2", Ext: []byte(`{}`)}},
				bidtype: openrtb_ext.BidTypeBanner,
				bidExt:  []byte(`{}`),
			},
			expected: &rawBidderResponseHookResult{
				bid:     &adapters.TypedBid{Bid: &openrtb2.Bid{ID: "2", Ext: []byte(`{}`)}},
				bidtype: openrtb_ext.BidTypeBanner,
				bidExt:  []byte(`{}`),
			},
		},
		{
			name: "bidder_in_list_creative_type_updated",
			result: &rawBidderResponseHookResult{
				bid:     &adapters.TypedBid{Bid: &openrtb2.Bid{ID: "3", AdM: "<VAST version=\"3.0\"></VAST>", Ext: []byte(`{}`)}},
				bidtype: openrtb_ext.BidTypeBanner,
				bidExt:  []byte(`{}`),
			},
			expected: &rawBidderResponseHookResult{
				bid:     &adapters.TypedBid{Bid: &openrtb2.Bid{ID: "3", AdM: "<VAST version=\"3.0\"></VAST>", Ext: []byte(`{}`)}},
				bidtype: openrtb_ext.BidTypeVideo,
				bidExt:  []byte(`{"prebid":{"type":"video"}}`),
			},
		},
		{
			name: "error_updating_bid_extension_due_to_malformed_JSON",
			result: &rawBidderResponseHookResult{
				bid:     &adapters.TypedBid{Bid: &openrtb2.Bid{ID: "4", AdM: "{\"native\":{\"link\":{\"url\":\"http://example.com\"},\"assets\":[]}}", Ext: []byte(`"{malformed}"`)}},
				bidtype: openrtb_ext.BidTypeBanner,
				bidExt:  []byte(`"{malformed}"`),
			},
			expected: &rawBidderResponseHookResult{
				bid:     &adapters.TypedBid{Bid: &openrtb2.Bid{ID: "4", AdM: "{\"native\":{\"link\":{\"url\":\"http://example.com\"},\"assets\":[]}}", Ext: []byte(`"{malformed}"`)}},
				bidtype: openrtb_ext.BidTypeNative,
				bidExt:  []byte(`"{malformed}"`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCreativeType(tt.result)
			assert.Equal(t, tt.expected, tt.result)
		})
	}
}

func TestApplyMutation(t *testing.T) {
	tests := []struct {
		name               string
		bidInfo            []*rawBidderResponseHookResult
		payload            hookstage.RawBidderResponsePayload
		expected           []*adapters.TypedBid
		expectedSeatNonBid openrtb_ext.SeatNonBidBuilder
	}{
		{
			name: "Single_bid",
			bidInfo: []*rawBidderResponseHookResult{
				{
					bid: &adapters.TypedBid{
						Bid: &openrtb2.Bid{ID: "1"},
					},
					bidtype: openrtb_ext.BidTypeBanner,
				},
			},
			expected: []*adapters.TypedBid{
				{
					Bid:     &openrtb2.Bid{ID: "1"},
					BidType: openrtb_ext.BidTypeBanner,
				},
			},
		},
		{
			name: "Multiple_bids_with_rejection",
			bidInfo: []*rawBidderResponseHookResult{
				{
					bid: &adapters.TypedBid{
						Bid: &openrtb2.Bid{ID: "1"},
					},
					bidtype: openrtb_ext.BidTypeBanner,
				},
				{
					bid: &adapters.TypedBid{
						Bid: &openrtb2.Bid{ID: "Bid-123", W: 100, H: 50},
					},
					bidtype:      openrtb_ext.BidTypeVideo,
					unwrapStatus: models.UnwrapEmptyVASTStatus,
				},
			},
			expected: []*adapters.TypedBid{
				{
					Bid:     &openrtb2.Bid{ID: "1"},
					BidType: openrtb_ext.BidTypeBanner,
				},
			},
			expectedSeatNonBid: func() openrtb_ext.SeatNonBidBuilder {
				seatNonBid := openrtb_ext.SeatNonBidBuilder{}
				seatNonBid.AddBid(openrtb_ext.NonBid{
					StatusCode: int(nbr.LossBidLostInVastUnwrap),
					Ext: openrtb_ext.ExtNonBid{
						Prebid: openrtb_ext.ExtNonBidPrebid{
							Bid: openrtb_ext.ExtNonBidPrebidBid{
								Price: 2.1,
								ID:    "Bid-123",
								W:     100,
								H:     50,
								Type:  openrtb_ext.BidTypeVideo,
							},
						},
					},
				}, "pubmatic")
				return seatNonBid
			}(),
		},
		{
			name:     "No_bids",
			bidInfo:  []*rawBidderResponseHookResult{},
			expected: []*adapters.TypedBid{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hookstage.HookResult[hookstage.RawBidderResponsePayload]{}
			payload := hookstage.RawBidderResponsePayload{
				BidderResponse: &adapters.BidderResponse{
					Bids:     []*adapters.TypedBid{},
					Currency: "USD",
				},
			}
			applyMutation(tt.bidInfo, &result, payload)

			for _, mut := range result.ChangeSet.Mutations() {
				newPayload, err := mut.Apply(payload)
				assert.NoError(t, err)
				payload = newPayload
			}

			assert.Equal(t, tt.expected, payload.BidderResponse.Bids)
		})
	}
}

func TestApplyPrivacyMaskingToIP(t *testing.T) {
	tests := []struct {
		name     string
		vastWrap models.VastUnWrap
		ip       string
		expected string
	}{
		{
			name:     "IPv4 with consent",
			vastWrap: models.VastUnWrap{IsPrivacyEnforced: true},
			ip:       "192.168.1.1",
			expected: "192.168.1.0",
		},
		{
			name:     "IPv4 without consent",
			vastWrap: models.VastUnWrap{IsPrivacyEnforced: false},
			ip:       "192.168.1.1",
			expected: "192.168.1.1",
		},
		{
			name:     "IPv6 without consent",
			vastWrap: models.VastUnWrap{IsPrivacyEnforced: false},
			ip:       "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expected: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		},
		{
			name:     "IPv6 with consent",
			vastWrap: models.VastUnWrap{IsPrivacyEnforced: true},
			ip:       "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expected: "2001:db8:85a3::",
		},
		{
			name:     "Invalid IP with consent",
			vastWrap: models.VastUnWrap{IsPrivacyEnforced: true},
			ip:       "invalid-ip",
			expected: "invalid-ip",
		},
		{
			name:     "Empty IP",
			vastWrap: models.VastUnWrap{IsPrivacyEnforced: false},
			ip:       "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := applyPrivacyMaskingToIP(tt.vastWrap, tt.ip)
			assert.Equal(t, tt.expected, actual, "applyPrivacyMaskingToIP() = %v, want %v", actual, tt.expected)
		})
	}
}
