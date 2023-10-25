package adapters

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
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
		imp.Video.W = 300
		imp.Video.H = 250
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
			name: "nil video input",
			args: args{
				imp: formImp(false),
				pubVASTTags: models.PublisherVASTTags{
					101: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
					102: &models.VASTTag{URL: `vast-tag-url-2`, Duration: 20},
				},
				matchedSlotKeys: []string{"abc@123"},
				slotMap: map[string]models.SlotMapping{
					"abc@123": createSlotMapping("abc@123",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
			},
			want: nil,
		},

		{
			name: "video input",
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
			got := PrepareVASTBidderParamJSON(tt.args.imp, tt.args.pubVASTTags, tt.args.matchedSlotKeys, tt.args.slotMap)
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
