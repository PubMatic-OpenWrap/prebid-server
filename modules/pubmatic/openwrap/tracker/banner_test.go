package tracker

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/stretchr/testify/assert"
)

func Test_injectBannerTracker(t *testing.T) {
	type args struct {
		rctx    models.RequestCtx
		tracker models.OWTracker
		bid     openrtb2.Bid
		seat    string
		pixels  []adunitconfig.UniversalPixel
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "endpoint_applovinmax",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					Endpoint: models.EndpointAppLovinMax,
				},
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				bid: openrtb2.Bid{
					AdM: `sample_creative`,
				},
				seat: "test",
			},
			want: `sample_creative`,
		},
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
		{
			name: "tbf_feature_enabled",
			args: args{
				rctx: models.RequestCtx{
					PubID:               5890,
					ProfileID:           1234,
					IsTBFFeatureEnabled: true,
				},
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				bid: openrtb2.Bid{
					AdM: `sample_creative`,
				},
			},
			want: `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="Tracking URL"></div>sample_creative`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := injectBannerTracker(tt.args.rctx, tt.args.tracker, tt.args.bid, tt.args.seat, tt.args.pixels); got != tt.want {
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
					PubID:               5890,
					ProfileID:           100,
					IsTBFFeatureEnabled: false,
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
					PubID:               5890,
					ProfileID:           1234,
					IsTBFFeatureEnabled: true,
				},
				bid: openrtb2.Bid{
					AdM: "<start>bid_AdM<end>",
				},
				tracker: "<start>tracker_url<end>",
			},
			want: "<start>tracker_url<end><start>bid_AdM<end>",
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

func Test_appendUPixelinBanner(t *testing.T) {
	type args struct {
		adm            string
		universalPixel []adunitconfig.UniversalPixel
	}
	type want struct {
		creative string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty universal pixel",
			args: args{
				adm: `sample_creative`,
			},
			want: want{
				creative: `sample_creative`,
			},
		},
		{
			name: "valid insertion of upixel",
			args: args{
				adm: `sample_creative`,
				universalPixel: []adunitconfig.UniversalPixel{
					{
						Id:        123,
						Pixel:     `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="sample.com"></div>`,
						PixelType: models.PixelTypeUrl,
						Pos:       models.PixelPosAbove,
						MediaType: "banner",
						Partners:  []string{"pubmatic", "appnexus"},
					},
					{
						Id:        123,
						Pixel:     "<script>__script__</script>",
						PixelType: models.PixelTypeJS,
						Pos:       models.PixelPosBelow,
						MediaType: "banner",
						Partners:  []string{"pubmatic", "appnexus"},
					},
					{
						Id:        123,
						Pixel:     `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="sample.com"></div>`,
						PixelType: "url",
						MediaType: "banner",
						Partners:  []string{"pubmatic", "appnexus"},
					},
				},
			},
			want: want{
				creative: `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="sample.com"></div>sample_creative<script>__script__</script><div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="sample.com"></div>`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendUPixelinBanner(tt.args.adm, tt.args.universalPixel)
			assert.Equal(t, tt.want.creative, got)
		})
	}
}
