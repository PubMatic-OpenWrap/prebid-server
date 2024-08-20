package openwrap

import (
	"fmt"
	"net/http"
	"testing"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/adapters"
	"github.com/prebid/prebid-server/v2/hooks/hookstage"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	mock_metrics "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/unwrap"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
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
		wantErr        bool
		wantSeatNonBid openrtb_ext.NonBidCollection
		mockHandler    http.HandlerFunc
	}{
		{
			name: "Empty Request Context",
			args: args{
				module: OpenWrap{},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{DebugMessages: []string{"error: request-ctx not found in handleRawBidderResponseHook()"}},
			wantErr:    false,
		},
		{
			name: "Set Vast Unwrapper to false in request context with type video",
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
						}}},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{ModuleContext: hookstage.ModuleContext{models.RequestContext: models.RequestCtx{VastUnwrapEnabled: false}}},
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			wantErr:    false,
		},
		{
			name: "Set Vast Unwrapper to false in request context with type video, stats enabled true",
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
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{models.RequestContext: models.RequestCtx{VastUnwrapEnabled: false}}},
			},
			mockHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Add("unwrap-status", "0")
				w.Header().Add("unwrap-count", "1")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(vastXMLAdM))
			}),
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
			},
			wantErr: false,
		},
		{
			name: "Set Vast Unwrapper to true in request context with invalid vast xml",
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
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{models.RequestContext: models.RequestCtx{VastUnwrapEnabled: true}}},
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
			wantErr: true,
		},
		{
			name: "Set Vast Unwrapper to true in request context with type video",
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
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{models.RequestContext: models.RequestCtx{VastUnwrapEnabled: true}}},
				isAdmUpdated:        true,
			},
			wantResult: hookstage.HookResult[hookstage.RawBidderResponsePayload]{Reject: false},
			setup: func() {
				mockMetricsEngine.EXPECT().RecordUnwrapRequestStatus("5890", "pubmatic", "0")
				mockMetricsEngine.EXPECT().RecordUnwrapWrapperCount("5890", "pubmatic", "1")
				mockMetricsEngine.EXPECT().RecordUnwrapRequestTime("5890", "pubmatic", gomock.Any())
				mockMetricsEngine.EXPECT().RecordUnwrapRespTime("5890", "1", gomock.Any())
			},
			wantErr: false,
		},
		{
			name: "Set Vast Unwrapper to true in request context for multiple bids with type video",
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
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{models.RequestContext: models.RequestCtx{VastUnwrapEnabled: true}}},
				isAdmUpdated:        true,
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
			wantErr: false,
		},
		{
			name: "Set Vast Unwrapper to true in request context for multiple bids with different type",
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
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{models.RequestContext: models.RequestCtx{VastUnwrapEnabled: true}}},
				isAdmUpdated:        true,
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
			wantErr: false,
		},
		{
			name: "Set Vast Unwrapper to true in request context with type video and source owsdk",
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
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{models.RequestContext: models.RequestCtx{VastUnwrapEnabled: true}}},
				isAdmUpdated:        true,
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
			wantErr: false,
		},
		{
			name: "bid with InvalidVAST should be discarded and should be present in seatNonBid",
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
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{models.RequestContext: models.RequestCtx{VastUnwrapEnabled: true}}},
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
			wantErr: true,
			wantSeatNonBid: func() openrtb_ext.NonBidCollection {
				seatNonBid := openrtb_ext.NonBidCollection{}
				seatNonBid.AddBid(openrtb_ext.NonBid{
					ImpId:      fmt.Sprintf("div-adunit-%d", 123),
					StatusCode: int(nbr.LossBidLostInVastUnwrap),
					Ext: openrtb_ext.ExtNonBid{
						Prebid: openrtb_ext.ExtNonBidPrebid{
							Bid: openrtb_ext.ExtNonBidPrebidBid{
								Price:          2.1,
								ID:             "Bid-123",
								W:              100,
								H:              50,
								Type:           openrtb_ext.BidTypeVideo,
								OriginalBidCPM: 2.1,
							},
						},
					},
				}, "pubmatic")
				return seatNonBid
			}(),
		},
		{
			name: "bid with EmptyVAST should be discarded and should be present in seatNonBid",
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
					Bidder: "pubmatic",
				},
				moduleInvocationCtx: hookstage.ModuleInvocationContext{AccountID: "5890", ModuleContext: hookstage.ModuleContext{models.RequestContext: models.RequestCtx{VastUnwrapEnabled: true}}},
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
			wantErr: true,
			wantSeatNonBid: func() openrtb_ext.NonBidCollection {
				seatNonBid := openrtb_ext.NonBidCollection{}
				seatNonBid.AddBid(openrtb_ext.NonBid{
					ImpId:      fmt.Sprintf("div-adunit-%d", 123),
					StatusCode: int(nbr.LossBidLostInVastUnwrap),
					Ext: openrtb_ext.ExtNonBid{
						Prebid: openrtb_ext.ExtNonBidPrebid{
							Bid: openrtb_ext.ExtNonBidPrebidBid{
								Price:          2.1,
								ID:             "Bid-123",
								W:              100,
								H:              50,
								Type:           openrtb_ext.BidTypeVideo,
								OriginalBidCPM: 2.1,
							},
						},
					},
				}, "pubmatic")
				return seatNonBid
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			m := tt.args.module
			m.unwrap = unwrap.NewUnwrap("http://localhost:8001/unwrap", 200, tt.mockHandler, m.metricEngine)
			hookResult, err := m.handleRawBidderResponseHook(tt.args.moduleInvocationCtx, tt.args.payload)
			if !assert.NoError(t, err, tt.wantErr) {
				return
			}
			if tt.args.moduleInvocationCtx.ModuleContext != nil && tt.args.isAdmUpdated {
				assert.Equal(t, inlineXMLAdM, tt.args.payload.Bids[0].Bid.AdM, "AdM is not updated correctly after executing RawBidderResponse hook.")
			}
			assert.Equal(t, tt.wantSeatNonBid, hookResult.SeatNonBid)
		})
	}
}

func TestNotEligibleForUnwrap(t *testing.T) {
	type args struct {
		bid *adapters.TypedBid
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Bid is nil",
			args: args{bid: nil},
			want: true,
		},
		{
			name: "Bid.Bid is nil",
			args: args{bid: &adapters.TypedBid{Bid: nil}},
			want: true,
		},
		{
			name: "AdM is empty",
			args: args{bid: &adapters.TypedBid{Bid: &openrtb2.Bid{AdM: ""}}},
			want: true,
		},
		{
			name: "BidType is not video",
			args: args{bid: &adapters.TypedBid{Bid: &openrtb2.Bid{AdM: "some_adm"}, BidType: openrtb_ext.BidTypeBanner}},
			want: true,
		},
		{
			name: "Bid is eligible for unwrap",
			args: args{bid: &adapters.TypedBid{Bid: &openrtb2.Bid{AdM: "some_adm"}, BidType: openrtb_ext.BidTypeVideo}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := notEligibleForUnwrap(tt.args.bid)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRejectBid(t *testing.T) {
	type args struct {
		unwrapStatus string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Unwrap status is UnwrapEmptyVASTStatus",
			args: args{unwrapStatus: models.UnwrapEmptyVASTStatus},
			want: true,
		},
		{
			name: "Unwrap status is UnwrapInvalidVASTStatus",
			args: args{unwrapStatus: models.UnwrapInvalidVASTStatus},
			want: true,
		},
		{
			name: "Unwrap status is something else",
			args: args{unwrapStatus: "OtherStatus"},
			want: false,
		},
		{
			name: "Unwrap status is empty string",
			args: args{unwrapStatus: ""},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rejectBid(tt.args.unwrapStatus)
			assert.Equal(t, tt.want, got)
		})
	}
}
