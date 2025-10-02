package adapters

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func formImp(isVideo bool) *openrtb2.Imp {
	imp := &openrtb2.Imp{}
	imp.ID = "impId"

	banner := new(openrtb2.Banner)
	banner.Format = []openrtb2.Format{
		{
			W: 600,
			H: 800,
		},
	}
	imp.Banner = banner

	if isVideo {
		imp.Video = new(openrtb2.Video)
		imp.Video.W = ptrutil.ToPtr[int64](300)
		imp.Video.H = ptrutil.ToPtr[int64](250)
		imp.Video.MIMEs = []string{"video/mp4"}
	}

	imp.TagID = "4"
	return imp
}
func createSlotMapping(slotName string, mappings map[string]interface{}) models.SlotMapping {
	return models.SlotMapping{
		PartnerId:    0,
		AdapterId:    0,
		VersionId:    0,
		SlotName:     slotName,
		SlotMappings: mappings,
		Hash:         "",
		OrderID:      0,
	}
}

func TestPrepareVASTBidderParamJSON(t *testing.T) {
	type args struct {
		imp             *openrtb2.Imp
		pubVASTTags     models.PublisherVASTTags
		matchedSlotKeys []string
		slotMap         map[string]models.SlotMapping
	}
	tests := []struct {
		name string
		args args
		want json.RawMessage
	}{
		{
			name: "VAST Tag ID not found in slot key",
			args: args{
				imp: formImp(true),
				pubVASTTags: models.PublisherVASTTags{
					101: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
					102: &models.VASTTag{URL: `vast-tag-url-2`, Duration: 20},
				},
				matchedSlotKeys: []string{"`abc@123`"},
				slotMap: map[string]models.SlotMapping{
					"abc@123": createSlotMapping("abc@123",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
			},
			want: nil,
		},
		{
			name: "VAST Tag slot mapping found",
			args: args{
				imp: formImp(true),
				pubVASTTags: models.PublisherVASTTags{
					123: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
				},
				matchedSlotKeys: []string{"abc@123"},
				slotMap: map[string]models.SlotMapping{
					"abc@123": createSlotMapping("abc@123",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
			},
			want: json.RawMessage(`{"tags":[{"tagid":"abc@123","url":"vast-tag-url-1","dur":15,"price":0,"params":{"param1":"85394","param2":"test","param3":"example1"}}]}`),
		},
		{
			name: "VAST Tag slot mapping not found",
			args: args{
				imp: formImp(true),
				pubVASTTags: models.PublisherVASTTags{
					123: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
				},
				matchedSlotKeys: []string{"abc@123"},
				slotMap: map[string]models.SlotMapping{
					"abcd@123": createSlotMapping("abcd@123",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
			},
			want: nil,
		},
		{
			name: "VAST Tag ID not found",
			args: args{
				imp: formImp(true),
				pubVASTTags: models.PublisherVASTTags{
					1234: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
				},
				matchedSlotKeys: []string{"abc@123"},
				slotMap: map[string]models.SlotMapping{
					"abc@123": createSlotMapping("abc@123",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrepareVASTBidderParamJSON(tt.args.pubVASTTags, tt.args.matchedSlotKeys, tt.args.slotMap)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetVASTTagID(t *testing.T) {
	assert.Equal(t, 0, getVASTTagID(""), "key: <empty>")
	assert.Equal(t, 0, getVASTTagID("abc"), "key: abc")
	assert.Equal(t, 0, getVASTTagID("abc@xyz"), "key: abc@xyz")
	assert.Equal(t, 123, getVASTTagID("abc@123"), "key: abc@123")
}

// func TestFilterImpsVastTagsByDuration(t *testing.T) {
// 	type args struct {
// 		rCtx    models.RequestCtx
// 		request *openrtb_ext.RequestWrapper
// 	}
// 	type want struct {
// 		imps      []*openrtb_ext.ImpWrapper
// 		impBidCtx map[string]models.ImpCtx
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want want
// 	}{
// 		{
// 			name: "Update Tag ids according to duration",
// 			args: args{
// 				rCtx: models.RequestCtx{
// 					ImpBidCtx: map[string]models.ImpCtx{
// 						"imp": {
// 							ImpID: "imp",
// 							Bidders: map[string]models.PartnerData{
// 								"pubmatic": {
// 									PartnerID:        1,
// 									PrebidBidderCode: "pubmatic",
// 									Params:           json.RawMessage(`{"adSlot":"/15671365/DMDemo@0x0","publisherId":"5890","wrapper":{"version":1,"profile":23498}}`),
// 								},
// 								"test_vastbidder15": {
// 									PartnerID:        123,
// 									PrebidBidderCode: "vastbidder",
// 									Params:           json.RawMessage(`{"tags":[{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55988","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":20,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55981","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":10,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55985","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":15,"price":0,"params":{"param1":"85394"}}]}`),
// 								},
// 							},
// 						},
// 					},
// 				},
// 				request: &openrtb_ext.RequestWrapper{
// 					Imp: []*openrtb_ext.ImpWrapper{
// 						{
// 							Imp: &openrtb2.Imp{
// 								ID: "imp::1",
// 								Video: &openrtb2.Video{
// 									MinDuration: 10,
// 								MaxDuration: 10,
// 							},
// 							Ext: json.RawMessage(`{"data":{"pbadslot":"/15671365/DMDemo"},"prebid":{"bidder":{"pubmatic":{"adSlot":"/15671365/DMDemo@0x0","publisherId":"5890","wiid":"d3e2e17d-5632-475c-a1d5-a76967aa9e71","wrapper":{"version":1,"profile":23498}},"test_vastbidder15":{"tags":[{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55988","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":20,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55981","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":10,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55985","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":15,"price":0,"params":{"param1":"85394"}}]}}},"tid":"c72aa41f-1174-45d1-a296-c61d094a7de8"}`),
// 						},
// 					},
// 					{
// 						Imp: &openrtb2.Imp{
// 							ID: "imp::2",
// 							Video: &openrtb2.Video{
// 								MinDuration: 10,
// 								MaxDuration: 20,
// 							},
// 							Ext: json.RawMessage(`{"data":{"pbadslot":"/15671365/DMDemo"},"prebid":{"bidder":{"pubmatic":{"adSlot":"/15671365/DMDemo@0x0","publisherId":"5890","wiid":"d3e2e17d-5632-475c-a1d5-a76967aa9e71","wrapper":{"version":1,"profile":23498}},"test_vastbidder15":{"tags":[{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55988","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":20,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55981","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":10,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55985","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":15,"price":0,"params":{"param1":"85394"}}]}}},"tid":"c72aa41f-1174-45d1-a296-c61d094a7de8"}`),
// 						},
// 					},
// 					{
// 						Imp: &openrtb2.Imp{
// 							ID: "imp::3",
// 							Video: &openrtb2.Video{
// 								MinDuration: 10,
// 								MaxDuration: 20,
// 							},
// 							Ext: json.RawMessage(`{"data":{"pbadslot":"/15671365/DMDemo"},"prebid":{"bidder":{"pubmatic":{"adSlot":"/15671365/DMDemo@0x0","publisherId":"5890","wiid":"d3e2e17d-5632-475c-a1d5-a76967aa9e71","wrapper":{"version":1,"profile":23498}},"test_vastbidder15":{"tags":[{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55988","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":20,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55981","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":10,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55985","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":15,"price":0,"params":{"param1":"85394"}}]}}},"tid":"c72aa41f-1174-45d1-a296-c61d094a7de8"}`),
// 						},
// 					},
// 				},
// 			},
// 			want: want{
// 				imps: func() []*openrtb_ext.ImpWrapper {
// 					imp1 := &openrtb_ext.ImpWrapper{
// 						Imp: &openrtb2.Imp{
// 							ID: "imp::1",
// 							Video: &openrtb2.Video{
// 								MinDuration: 10,
// 								MaxDuration: 10,
// 							},
// 							Ext: json.RawMessage(`{"data":{"pbadslot":"/15671365/DMDemo"},"prebid":{"bidder":{"pubmatic":{"adSlot":"/15671365/DMDemo@0x0","publisherId":"5890","wiid":"d3e2e17d-5632-475c-a1d5-a76967aa9e71","wrapper":{"version":1,"profile":23498}},"test_vastbidder15":{"tags":[{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55981","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":10,"price":0,"params":{"param1":"85394"}}]}}},"tid":"c72aa41f-1174-45d1-a296-c61d094a7de8"}`),
// 						},
// 					}
// 					imp1.GetImpExt()
// 					imp2 := &openrtb_ext.ImpWrapper{
// 						Imp: &openrtb2.Imp{
// 							ID: "imp::2",
// 							Video: &openrtb2.Video{
// 								MinDuration: 10,
// 								MaxDuration: 20,
// 							},
// 							Ext: json.RawMessage(`{"data":{"pbadslot":"/15671365/DMDemo"},"prebid":{"bidder":{"pubmatic":{"adSlot":"/15671365/DMDemo@0x0","publisherId":"5890","wiid":"d3e2e17d-5632-475c-a1d5-a76967aa9e71","wrapper":{"version":1,"profile":23498}},"test_vastbidder15":{"tags":[{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55988","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":20,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55981","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":10,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55985","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":15,"price":0,"params":{"param1":"85394"}}]}}},"tid":"c72aa41f-1174-45d1-a296-c61d094a7de8"}`),
// 						},
// 					}
// 					imp2.GetImpExt()

// 					imp3 := &openrtb_ext.ImpWrapper{
// 						Imp: &openrtb2.Imp{
// 							ID: "imp::3",
// 							Video: &openrtb2.Video{
// 								MinDuration: 10,
// 								MaxDuration: 20,
// 							},
// 							Ext: json.RawMessage(`{"data":{"pbadslot":"/15671365/DMDemo"},"prebid":{"bidder":{"pubmatic":{"adSlot":"/15671365/DMDemo@0x0","publisherId":"5890","wiid":"d3e2e17d-5632-475c-a1d5-a76967aa9e71","wrapper":{"version":1,"profile":23498}},"test_vastbidder15":{"tags":[{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55988","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":20,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55981","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":10,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55985","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":15,"price":0,"params":{"param1":"85394"}}]}}},"tid":"c72aa41f-1174-45d1-a296-c61d094a7de8"}`),
// 						},
// 					}
// 					imp3.GetImpExt()
// 					imps := []*openrtb_ext.ImpWrapper{imp1, imp2, imp3}
// 					return imps
// 				}(),
// 				impBidCtx: map[string]models.ImpCtx{
// 					"imp": {
// 						ImpID: "imp",
// 						Bidders: map[string]models.PartnerData{
// 							"pubmatic": {
// 								PartnerID:        1,
// 								PrebidBidderCode: "pubmatic",
// 								Params:           json.RawMessage(`{"adSlot":"/15671365/DMDemo@0x0","publisherId":"5890","wrapper":{"version":1,"profile":23498}}`),
// 							},
// 							"test_vastbidder15": {
// 								PartnerID:        123,
// 								PrebidBidderCode: "vastbidder",
// 								Params:           json.RawMessage(`{"tags":[{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55988","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":20,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55981","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":10,"price":0,"params":{"param1":"85394"}},{"tagid":"/15671365/DMDemo@com.pubmatic.openbid.app@55985","url":"http://10.172.141.11:3141/vastbidder/{param1}?VPI=MP4&app[bundle]={bundle}&app[name]={appname}&app[cat]={cat}&app[domain]={domain}&app[privacypolicy]={privacypolicy}&app[storeurl]={storeurl_ESC}&app[ver]={appver}&cb={cachebuster}&device[devicetype]={devicetype}&device[ifa]={ifa}&device[make]={make}&device[model]={model}&device[dnt]={dnt}&player_height={playerheight}&player_width={playerwidth}&ip_addr={ip}&device[ua]={useragent_ESC}&price=10&ifatype={ifa_type}","dur":15,"price":0,"params":{"param1":"85394"}}]}`),
// 								VASTTagFlags: map[string]bool{
// 									"/15671365/DMDemo@com.pubmatic.openbid.app@55981": false,
// 									"/15671365/DMDemo@com.pubmatic.openbid.app@55988": false,
// 									"/15671365/DMDemo@com.pubmatic.openbid.app@55985": false,
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			FilterImpsVastTagsByDuration(tt.args.imps, tt.args.impBidCtx)
// 			// TODO:: Assert on impression wrapper
// 			// assert.Equal(t, tt.want.imps, tt.args.imps, "Invalid impressions created")
// 			// assert.Equal(t, tt.want.impBidCtx, tt.args.impBidCtx, "Invalid impression context created")
// 		})
// 	}
// }
