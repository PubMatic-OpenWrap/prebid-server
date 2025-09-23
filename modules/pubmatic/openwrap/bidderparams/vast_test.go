package bidderparams

import (
	"fmt"
	"sort"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestGetVASTBidderSlotKeys(t *testing.T) {
	type args struct {
		imp         *openrtb2.Imp
		slotKey     string
		slotMap     map[string]models.SlotMapping
		pubVASTTags models.PublisherVASTTags
		adpodCtx    models.AdpodCtx
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: `empty-imp-object`,
			args: args{},
			want: nil,
		},
		{
			name: `non-video-request`,
			args: args{
				imp: &openrtb2.Imp{ID: "123", Banner: &openrtb2.Banner{}},
			},
			want: nil,
		},
		{
			name: `empty-mappings`,
			args: args{
				imp: &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
			},
			want: nil,
		},
		{
			name: `key-not-present`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `key-not-present@com.pubmatic.openbid.app@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
			},
			want: nil,
		},
		{
			name: `invalid-vast-tag-id`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@invalid-vast-tag-id": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@invalid-vast-tag-id",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
			},
			want: nil,
		},
		{
			name: `vast-tag-details-not-present`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
			},
			want: nil,
		},
		{
			name: `invalid-vast-tag`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
				pubVASTTags: models.PublisherVASTTags{
					101: &models.VASTTag{},
					102: &models.VASTTag{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: `valid-row-1`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
				pubVASTTags: models.PublisherVASTTags{
					101: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
					102: &models.VASTTag{URL: `vast-tag-url-2`, Duration: 20},
				},
			},
			want: []string{
				`/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101`,
				`/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102`,
			},
		},
		{
			name: `mixed-mappings`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@101": createSlotMapping("/15671365/MG_VideoAdUnit@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@102": createSlotMapping("/15671365/MG_VideoAdUnit@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
				pubVASTTags: models.PublisherVASTTags{
					101: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
					102: &models.VASTTag{URL: `vast-tag-url-2`, Duration: 20},
				},
			},
			want: []string{
				`/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101`,
				`/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102`,
			},
		},
		{
			name: `default-mappings-only`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit1@com.pubmatic.openbid.app@101": createSlotMapping("/15671365/MG_VideoAdUnit1@com.pubmatic.openbid.app@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit1@com.pubmatic.openbid.app@102": createSlotMapping("/15671365/MG_VideoAdUnit1@com.pubmatic.openbid.app@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@@101": createSlotMapping("/15671365/MG_VideoAdUnit@@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@@102": createSlotMapping("/15671365/MG_VideoAdUnit@@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
				pubVASTTags: models.PublisherVASTTags{
					101: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
					102: &models.VASTTag{URL: `vast-tag-url-2`, Duration: 20},
				},
			},
			want: []string{
				`/15671365/MG_VideoAdUnit@@101`,
				`/15671365/MG_VideoAdUnit@@102`,
			},
		},
		{
			name: `no-site-bundle`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `/15671365/MG_VideoAdUnit@@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@@101": createSlotMapping("/15671365/MG_VideoAdUnit@@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@@102": createSlotMapping("/15671365/MG_VideoAdUnit@@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
				pubVASTTags: models.PublisherVASTTags{
					101: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
					102: &models.VASTTag{URL: `vast-tag-url-2`, Duration: 20},
				},
			},
			want: []string{
				`/15671365/MG_VideoAdUnit@@101`,
				`/15671365/MG_VideoAdUnit@@102`,
			},
		},
		{
			name: `different-site-bundle`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@not-present-domain@102": createSlotMapping("/15671365/MG_VideoAdUnit@not-present-domain@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@@101": createSlotMapping("/15671365/MG_VideoAdUnit@@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@@102": createSlotMapping("/15671365/MG_VideoAdUnit@@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
				pubVASTTags: models.PublisherVASTTags{
					101: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
					102: &models.VASTTag{URL: `vast-tag-url-2`, Duration: 20},
				},
			},
			want: []string{
				`/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101`,
			},
		},
		{
			name: `no-adunit`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `@@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@101": createSlotMapping("/15671365/MG_VideoAdUnit@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@102": createSlotMapping("/15671365/MG_VideoAdUnit@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
				pubVASTTags: models.PublisherVASTTags{
					101: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
					102: &models.VASTTag{URL: `vast-tag-url-2`, Duration: 20},
				},
			},
			want: nil,
		},
		{
			name: `case in-sensitive Ad Unit mapping`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `/15671365/mg_VideoAdUnit@com.pubmatic.openbid.app@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VIDEOAdUnit@com.pubmatic.openbid.app@102": createSlotMapping("/15671365/MG_VIDEOAdUnit@com.pubmatic.openbid.app@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@COM.pubmatic.openbid.app@103": createSlotMapping("/15671365/MG_VideoAdUnit@COM.pubmatic.openbid.app@103",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@101": createSlotMapping("/15671365/MG_VideoAdUnit@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@102": createSlotMapping("/15671365/MG_VideoAdUnit@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
				pubVASTTags: models.PublisherVASTTags{
					101: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
					102: &models.VASTTag{URL: `vast-tag-url-2`, Duration: 20},
				},
			},
			want: []string{
				`/15671365/MG_VIDEOAdUnit@com.pubmatic.openbid.app@102`,
				`/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101`,
			},
		},
		{
			name: `case sensitive site-bundle`,
			args: args{
				imp:     &openrtb2.Imp{ID: "123", Video: &openrtb2.Video{}},
				slotKey: `/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@`,
				slotMap: map[string]models.SlotMapping{
					"/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101": createSlotMapping("/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@not-present-domain@102": createSlotMapping("/15671365/MG_VideoAdUnit@not-present-domain@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@COM.pubmatic.openbid.app@103": createSlotMapping("/15671365/MG_VideoAdUnit@COM.pubmatic.openbid.app@103",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@@101": createSlotMapping("/15671365/MG_VideoAdUnit@@101",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
					"/15671365/MG_VideoAdUnit@@102": createSlotMapping("/15671365/MG_VideoAdUnit@@102",
						map[string]interface{}{"param1": "85394", "param2": "test", "param3": "example1"}),
				},
				pubVASTTags: models.PublisherVASTTags{
					101: &models.VASTTag{URL: `vast-tag-url-1`, Duration: 15},
					102: &models.VASTTag{URL: `vast-tag-url-2`, Duration: 20},
				},
			},
			want: []string{
				`/15671365/MG_VideoAdUnit@com.pubmatic.openbid.app@101`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getVASTBidderSlotKeys(tt.args.imp, tt.args.slotKey, tt.args.slotMap, tt.args.pubVASTTags, tt.args.adpodCtx)
			if (err != nil) != tt.wantErr {
				assert.Equal(t, tt.wantErr, err != nil)
				return
			}
			sort.Strings(got)
			sort.Strings(tt.want)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateVASTTag(t *testing.T) {
	type args struct {
		vastTag          *models.VASTTag
		videoMinDuration int64
		videoMaxDuration int64
		podDur           int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    `empty_vast_tag`,
			args:    args{},
			wantErr: fmt.Errorf("Empty vast tag"),
		},
		{
			name: `empty_url`,
			args: args{
				vastTag: &models.VASTTag{
					ID: 101,
				},
			},
			wantErr: fmt.Errorf("VAST tag mandatory parameter 'url' missing: 101"),
		},
		{
			name: `empty_duration`,
			args: args{
				vastTag: &models.VASTTag{
					ID:  101,
					URL: `vast_tag_url`,
				},
			},
			wantErr: fmt.Errorf("VAST tag mandatory parameter 'duration' missing: 101"),
		},
		{
			name: `valid-without-duration-checks`,
			args: args{
				vastTag: &models.VASTTag{
					ID:       101,
					URL:      `vast_tag_url`,
					Duration: 15,
				},
			},
			wantErr: nil,
		},
		{
			name: `max_duration_check`,
			args: args{
				vastTag: &models.VASTTag{
					ID:       101,
					URL:      `vast_tag_url`,
					Duration: 15,
				},
				videoMaxDuration: 10,
			},
			wantErr: fmt.Errorf("VAST tag 'duration' validation failed 'tag.duration > video.maxduration' vastTagID:101, tag.duration:15, video.maxduration:10"),
		},
		{
			name: `min_duration_check`,
			args: args{
				vastTag: &models.VASTTag{
					ID:       101,
					URL:      `vast_tag_url`,
					Duration: 15,
				},
				videoMaxDuration: 30,
				videoMinDuration: 20,
			},
			wantErr: fmt.Errorf("VAST tag 'duration' validation failed 'tag.duration < video.minduration' vastTagID:101, tag.duration:15, video.minduration:20"),
		},
		{
			name: `valid_non_adpod`,
			args: args{
				vastTag: &models.VASTTag{
					ID:       101,
					URL:      `vast_tag_url`,
					Duration: 25,
				},
				videoMaxDuration: 30,
				videoMinDuration: 20,
			},
			wantErr: nil,
		},
		{
			name: `valid_non_adpod_exact`,
			args: args{
				vastTag: &models.VASTTag{
					ID:       101,
					URL:      `vast_tag_url`,
					Duration: 25,
				},
				videoMaxDuration: 25,
				videoMinDuration: 25,
			},
			wantErr: nil,
		},
		{
			name: `empty_adpod`,
			args: args{
				vastTag: &models.VASTTag{
					ID:       101,
					URL:      `vast_tag_url`,
					Duration: 25,
				},
				videoMaxDuration: 25,
				videoMinDuration: 25,
			},
			wantErr: fmt.Errorf(`VAST tag 'duration' validation failed 'tag.duration > adpod.maxduration' vastTagID:101, tag.duration:25, adpod.maxduration:0`),
		},
		{
			name: `adpod_min_duration_check`,
			args: args{
				vastTag: &models.VASTTag{
					ID:       101,
					URL:      `vast_tag_url`,
					Duration: 5,
				},
				videoMaxDuration: 25,
				podDur:           10,
			},
			wantErr: fmt.Errorf(`VAST tag 'duration' validation failed 'tag.duration < adpod.minduration' vastTagID:101, tag.duration:5, adpod.minduration:10`),
		},
		{
			name: `adpod_max_duration_check`,
			args: args{
				vastTag: &models.VASTTag{
					ID:       101,
					URL:      `vast_tag_url`,
					Duration: 15,
				},
				videoMaxDuration: 25,
				podDur:           10,
			},
			wantErr: fmt.Errorf(`VAST tag 'duration' validation failed 'tag.duration > imp.video.PodDur' vastTagID:101, tag.duration:15, imp.video.PodDur:10`),
		},
		{
			name: `adpod_exact_duration_check`,
			args: args{
				vastTag: &models.VASTTag{
					ID:       101,
					URL:      `vast_tag_url`,
					Duration: 15,
				},
				videoMaxDuration: 25,
				podDur:           15,
			},
			wantErr: nil,
		},
		{
			name: `mixed-check-1`,
			args: args{
				vastTag: &models.VASTTag{
					ID:       101,
					URL:      `vast_tag_url`,
					Duration: 15,
				},
				videoMaxDuration: 25,
				videoMinDuration: 5,
				podDur:           15,
			},
			wantErr: nil,
		},
		{
			name: `video-min-duration-check-skipped-incase-of-adpod`,
			args: args{
				vastTag: &models.VASTTag{
					ID:       101,
					URL:      `vast_tag_url`,
					Duration: 10,
				},
				videoMaxDuration: 25,
				videoMinDuration: 15,
				podDur:           10,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateVASTTag(tt.args.vastTag, tt.args.videoMinDuration, tt.args.videoMaxDuration, tt.args.podDur); err != nil {
				assert.EqualError(t, err, string(tt.wantErr.Error()), "Expected error:%v but got:%v", tt.wantErr, err)
			}
		})
	}
}
