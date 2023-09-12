package adunitconfig

import (
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/util/ptrutil"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func Test_getRegexMatch(t *testing.T) {
	type args struct {
		rctx     models.RequestCtx
		slotName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Matching_Slotname_with_regex_expression,_returing_valid_values",
			args: args{
				rctx: models.RequestCtx{
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Regex: true,
						Config: map[string]*adunitconfig.AdConfig{
							"^/15671365/MG_VideoAdUnit[0-9]*$": {
								Video: &adunitconfig.Video{
									Enabled: ptrutil.ToPtr[bool](true),
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
									},
								},
							},
						},
					},
				},
				slotName: "/15671365/MG_VideoAdUnit12349",
			},
			want: "^/15671365/MG_VideoAdUnit[0-9]*$",
		},
		{
			name: "Slotname_and_regex_dont_match",
			args: args{
				rctx: models.RequestCtx{
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Regex: true,
						Config: map[string]*adunitconfig.AdConfig{
							"^/15671365/MG_VideoAdUnit[0-9]*$": {
								Video: &adunitconfig.Video{
									Enabled: ptrutil.ToPtr[bool](true),
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
									},
								},
							},
						},
					},
				},
				slotName: "/15627/Regex_Not_Registered",
			},
			want: "",
		},
		{
			name: "Empty_AdunitConfig",
			args: args{
				rctx: models.RequestCtx{
					AdUnitConfig: &adunitconfig.AdUnitConfig{
						Config: map[string]*adunitconfig.AdConfig{},
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRegexMatch(tt.args.rctx, tt.args.slotName); got != tt.want {
				t.Errorf("getRegexMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
