package adunitconfig

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func getAdunitConfigWithRx() *adunitconfig.AdUnitConfig {
	return &adunitconfig.AdUnitConfig{
		ConfigPattern: "_AU_",
		Regex:         true,
		Config: map[string]*adunitconfig.AdConfig{
			"default": {
				Video: &adunitconfig.Video{
					Enabled: ptrutil.ToPtr(true),
				},
			},
			"^/15671365/test_adunit[0-9]*$": {
				Video: &adunitconfig.Video{
					Enabled: ptrutil.ToPtr(true),
					Config: &adunitconfig.VideoConfig{
						Video: openrtb2.Video{
							SkipAfter:   16,
							MaxDuration: 57,
							Skip:        ptrutil.ToPtr[int8](2),
							SkipMin:     11,
							MinDuration: 15,
							MIMEs: []string{
								"video/mp4",
								"video/x-flv",
								"video/mp4",
								"video/webm",
							},
						},
						ConnectionType: []int{
							1,
							2,
							6,
						},
					},
				},
			},
			"/15671365/test_adunit1": {
				Video: &adunitconfig.Video{
					Enabled: ptrutil.ToPtr(true),
					Config:  &adunitconfig.VideoConfig{},
				},
			},
		},
	}
}

func TestSelectSlot(t *testing.T) {
	type args struct {
		rCtx   models.RequestCtx
		h      int64
		w      int64
		tagid  string
		div    string
		source string
	}
	type want struct {
		slotAdUnitConfig *adunitconfig.AdConfig
		slotName         string
		isRegex          bool
		matchedRegex     string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Matching_Slot_config_when_regex_is_present_and_slotconfig_is_absent",
			args: args{
				rCtx: models.RequestCtx{
					AdUnitConfig: func() *adunitconfig.AdUnitConfig {
						auc := getAdunitConfigWithRx()

						// Temporary fix to make UT execution consistent.
						// TODO: make getRegexMatch()'s loop consistent.
						delete(auc.Config, "/15671365/test_adunit1")
						return auc
					}(),
				},
				h:      300,
				w:      200,
				tagid:  "/15671365/Test_AdUnit92349",
				div:    "Div1",
				source: "test.com",
			},
			want: want{
				slotAdUnitConfig: &adunitconfig.AdConfig{
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr(true),
						Config: &adunitconfig.VideoConfig{
							Video: openrtb2.Video{
								SkipAfter:   16,
								MaxDuration: 57,
								Skip:        ptrutil.ToPtr[int8](2),
								SkipMin:     11,
								MinDuration: 15,
								MIMEs: []string{
									"video/mp4",
									"video/x-flv",
									"video/mp4",
									"video/webm",
								},
							},
							ConnectionType: []int{
								1,
								2,
								6,
							},
						},
					},
				},
				slotName:     "/15671365/Test_AdUnit92349",
				isRegex:      true,
				matchedRegex: "^/15671365/test_adunit[0-9]*$",
			},
		},
		{
			name: "Priority_to_Exact_Match_for_Slot_config_when_regex_is_also_present",
			args: args{
				rCtx: models.RequestCtx{
					AdUnitConfig: getAdunitConfigWithRx(),
				},
				h:      300,
				w:      200,
				tagid:  "/15671365/Test_AdUnit1",
				div:    "Div1",
				source: "test.com",
			},
			want: want{
				slotAdUnitConfig: &adunitconfig.AdConfig{
					Video: &adunitconfig.Video{
						Enabled: ptrutil.ToPtr(true),
						Config:  &adunitconfig.VideoConfig{},
					},
				},
				slotName:     "/15671365/Test_AdUnit1",
				isRegex:      false,
				matchedRegex: "",
			},
		},
		{
			name: "when_slot_name_does_not_match_slot_as_well_as_not_found_matched_regex",
			args: args{
				rCtx: models.RequestCtx{
					AdUnitConfig: getAdunitConfigWithRx(),
				},
				tagid: "/15627/Regex_Not_Registered",
			},
			want: want{
				slotAdUnitConfig: nil,
				slotName:         "",
				isRegex:          false,
				matchedRegex:     "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSlotAdUnitConfig, gotSlotName, gotIsRegex, gotMatchedRegex := selectSlot(tt.args.rCtx, tt.args.h, tt.args.w, tt.args.tagid, tt.args.div, tt.args.source)
			assert.Equal(t, tt.want.slotAdUnitConfig, gotSlotAdUnitConfig)
			if gotSlotName != tt.want.slotName {
				t.Errorf("selectSlot() gotSlotName = %v, want %v", gotSlotName, tt.want.slotName)
			}
			if gotIsRegex != tt.want.isRegex {
				t.Errorf("selectSlot() gotIsRegex = %v, want %v", gotIsRegex, tt.want.isRegex)
			}
			if gotMatchedRegex != tt.want.matchedRegex {
				t.Errorf("selectSlot() gotMatchedRegex = %v, want %v", gotMatchedRegex, tt.want.matchedRegex)
			}
		})
	}
}
