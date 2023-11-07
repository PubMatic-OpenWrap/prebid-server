package tracker

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	mock_cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/tbf"
)

func Test_injectBannerTracker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)
	tbf.SetAndResetTBFConfig(mockCache, map[int]map[int]int{
		5890: {1234: 100},
	})
	type args struct {
		rctx    models.RequestCtx
		tracker models.OWTracker
		bid     openrtb2.Bid
		seat    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "app_platform",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
				},
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				bid: openrtb2.Bid{
					AdM: `sample_creative`,
				},
				seat: "test",
			},
			want: `sample_creative<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="Tracking URL"></div>`,
		},
		{
			name: "app_platform_OM_Inactive_pubmatic",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
				},
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
					DspId:      -1,
				},
				bid: openrtb2.Bid{
					AdM: `sample_creative`,
				},
				seat: models.BidderPubMatic,
			},
			want: `sample_creative<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="Tracking URL"></div>`,
		},
		{
			name: "app_platform_OM_Active_pubmatic",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
				},
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
					DspId:      models.DspId_DV360,
				},
				bid: openrtb2.Bid{
					AdM: `sample_creative`,
				},
				seat: models.BidderPubMatic,
			},
			want: `sample_creative<script id="OWPubOMVerification" data-owurl="Tracking URL" src="${OMScript}"></script>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := injectBannerTracker(tt.args.rctx, tt.args.tracker, tt.args.bid, tt.args.seat); got != tt.want {
				t.Errorf("injectBannerTracker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_trackerWithOM(t *testing.T) {
	type args struct {
		tracker    models.OWTracker
		platform   string
		bidderCode string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "in-app_partner_otherthan_pubmatic",
			args: args{
				tracker: models.OWTracker{
					DspId: models.DspId_DV360,
				},
				platform:   models.PLATFORM_APP,
				bidderCode: "test",
			},
			want: false,
		},
		{
			name: "in-app_partner_pubmatic_other_dv360",
			args: args{
				tracker: models.OWTracker{
					DspId: -1,
				},
				platform:   models.PLATFORM_APP,
				bidderCode: models.BidderPubMatic,
			},
			want: false,
		},
		{
			name: "display_partner_pubmatic_dv360",
			args: args{
				tracker: models.OWTracker{
					DspId: models.DspId_DV360,
				},
				platform:   models.PLATFORM_DISPLAY,
				bidderCode: models.BidderPubMatic,
			},
			want: false,
		},
		{
			name: "in-app_partner_pubmatic_dv360",
			args: args{
				tracker: models.OWTracker{
					DspId: models.DspId_DV360,
				},
				platform:   models.PLATFORM_APP,
				bidderCode: models.BidderPubMatic,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trackerWithOM(tt.args.tracker, tt.args.platform, tt.args.bidderCode); got != tt.want {
				t.Errorf("trackerWithOM() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_applyTBFFeature(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCache := mock_cache.NewMockCache(ctrl)
	defer ctrl.Finish()
	tbf.SetAndResetTBFConfig(mockCache, map[int]map[int]int{
		5890: {1234: 100},
	})

	type args struct {
		rctx    models.RequestCtx
		bid     openrtb2.Bid
		tracker string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "tbf_feature_disabled",
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 100,
				},
				bid: openrtb2.Bid{
					AdM: "<start>bid_AdM<end>",
				},
				tracker: "<start>tracker_url<end>",
			},
			want: "<start>bid_AdM<end><start>tracker_url<end>",
		},
		{
			name: "tbf_feature_enabled",
			args: args{
				rctx: models.RequestCtx{
					PubID:     5890,
					ProfileID: 1234,
				},
				bid: openrtb2.Bid{
					AdM: "<start>bid_AdM<end>",
				},
				tracker: "<start>tracker_url<end>",
			},
			want: "<start>tracker_url<end><start>bid_AdM<end>",
		},
		{
			name: "invalid_pubid",
			args: args{
				rctx: models.RequestCtx{
					PubID:     -1,
					ProfileID: 1234,
				},
				bid: openrtb2.Bid{
					AdM: "<start>bid_AdM<end>",
				},
				tracker: "<start>tracker_url<end>",
			},
			want: "<start>bid_AdM<end><start>tracker_url<end>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := applyTBFFeature(tt.args.rctx, tt.args.bid, tt.args.tracker); got != tt.want {
				t.Errorf("applyTBFFeature() = %v, want %v", got, tt.want)
			}
		})
	}
}
