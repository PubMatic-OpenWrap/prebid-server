package tracker

import (
	"strconv"
	"testing"
	"time"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/ptrutil"
)

func TestGetTrackerInfo(t *testing.T) {
	startTime := int64(time.Now().Unix())
	type args struct {
		rCtx        models.RequestCtx
		responseExt openrtb_ext.ExtBidResponse
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "all_tracker_info_without_floors",
			args: args{
				rCtx: models.RequestCtx{
					TrackerEndpoint:     "localhost:8080/wt",
					PubID:               123,
					ProfileID:           1,
					VersionID:           1,
					PageURL:             "www.test.com",
					LoggerImpressionID:  "iid123",
					StartTime:           startTime,
					DevicePlatform:      models.DevicePlatformMobileAppAndroid,
					Origin:              "www.publisher.com",
					ABTestConfigApplied: 1,
				},
				responseExt: openrtb_ext.ExtBidResponse{},
			},
			want: "localhost:8080/wt?adv=&af=&aps=0&au=%24%7BADUNIT%7D&bc=%24%7BBIDDER_CODE%7D&bidid=%24%7BBID_ID%7D&di=&eg=%24%7BG_ECPM%7D&en=%24%7BN_ECPM%7D&ft=0&iid=iid123&kgpv=%24%7BKGPV%7D&orig=www.publisher.com&origbidid=%24%7BORIGBID_ID%7D&pdvid=0&pid=1&plt=5&pn=%24%7BPARTNER_NAME%7D&psz=&pubid=123&purl=www.test.com&rwrd=%24%7BREWARDED%7D&sl=1&slot=%24%7BSLOT_ID%7D&ss=0&tgid=1&tst=" + strconv.FormatInt(startTime, 10),
		},
		{
			name: "all_tracker_info_with_floors",
			args: args{
				rCtx: models.RequestCtx{
					TrackerEndpoint:     "localhost:8080/wt",
					PubID:               123,
					ProfileID:           1,
					VersionID:           1,
					PageURL:             "www.test.com",
					LoggerImpressionID:  "iid123",
					StartTime:           startTime,
					DevicePlatform:      models.DevicePlatformMobileAppAndroid,
					Origin:              "www.publisher.com",
					ABTestConfigApplied: 1,
				},
				responseExt: openrtb_ext.ExtBidResponse{
					Prebid: &openrtb_ext.ExtResponsePrebid{
						Floors: &openrtb_ext.PriceFloorRules{
							Skipped: ptrutil.ToPtr(true),
							Data: &openrtb_ext.PriceFloorData{
								ModelGroups: []openrtb_ext.PriceFloorModelGroup{
									{
										ModelVersion: "version 1",
									},
								},
							},
							PriceFloorLocation: openrtb_ext.FetchLocation,
							Enforcement: &openrtb_ext.PriceFloorEnforcement{
								EnforcePBS: ptrutil.ToPtr(true),
							},
						},
					},
				},
			},
			want: "localhost:8080/wt?adv=&af=&aps=0&au=%24%7BADUNIT%7D&bc=%24%7BBIDDER_CODE%7D&bidid=%24%7BBID_ID%7D&di=&eg=%24%7BG_ECPM%7D&en=%24%7BN_ECPM%7D&fmv=version+1&fskp=1&fsrc=2&ft=1&iid=iid123&kgpv=%24%7BKGPV%7D&orig=www.publisher.com&origbidid=%24%7BORIGBID_ID%7D&pdvid=0&pid=1&plt=5&pn=%24%7BPARTNER_NAME%7D&psz=&pubid=123&purl=www.test.com&rwrd=%24%7BREWARDED%7D&sl=1&slot=%24%7BSLOT_ID%7D&ss=0&tgid=1&tst=" + strconv.FormatInt(startTime, 10),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTrackerInfo(tt.args.rCtx, tt.args.responseExt); got != tt.want {
				t.Errorf("GetTrackerInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
