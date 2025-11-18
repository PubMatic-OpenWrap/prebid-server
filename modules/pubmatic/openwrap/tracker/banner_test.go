package tracker

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestInjectBannerTracker(t *testing.T) {
	type args struct {
		rctx    models.RequestCtx
		tracker models.OWTracker
		bid     openrtb2.Bid
		seat    string
		pixels  []adunitconfig.UniversalPixel
	}
	type want struct {
		adm  string
		burl string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "endpoint_applovinmax",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					Endpoint: models.EndpointAppLovinMax,
				},
				tracker: models.OWTracker{
					TrackerURL: `sample.com/track?tid=1234`,
				},
				bid: openrtb2.Bid{
					AdM:  `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="sample.com"></div>`,
					BURL: `http://burl.com`,
				},
				seat: "pubmatic",
			},
			want: want{
				adm:  `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="sample.com"></div>`,
				burl: `sample.com/track?tid=1234&owsspburl=http%3A%2F%2Fburl.com`,
			},
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
			want: want{
				adm: `sample_creative<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="Tracking URL"></div>`,
			},
		},
		{
			name: "app_platform_OM_Inactive_pubmatic",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
				},
				tracker: models.OWTracker{
					TrackerURL:  `Tracking URL`,
					IsOMEnabled: false,
				},
				bid: openrtb2.Bid{
					AdM: `sample_creative`,
				},
				seat: models.BidderPubMatic,
			},
			want: want{
				adm: `sample_creative<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="Tracking URL"></div>`,
			},
		},
		{
			name: "app_platform_OM_Active_pubmatic",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
				},
				tracker: models.OWTracker{
					TrackerURL:  `Tracking URL`,
					IsOMEnabled: true,
				},
				bid: openrtb2.Bid{
					AdM: `sample_creative`,
				},
				seat: models.BidderPubMatic,
			},
			want: want{
				adm: `sample_creative<script id="OWPubOMVerification" data-owurl="Tracking URL" src="${OMScript}"></script>`,
			},
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
			want: want{
				adm: `<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="Tracking URL"></div>sample_creative`,
			},
		},
		{
			name: "app_platform_partner_other_than_pubmatic_imp_counting_method_enabled",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					ImpCountingMethodEnabledBidders: map[string]struct{}{
						string(openrtb_ext.BidderIx): {},
					},
				},
				tracker: models.OWTracker{
					TrackerURL:  `Tracking URL`,
					IsOMEnabled: true,
				},
				bid: openrtb2.Bid{
					AdM: `sample_creative`,
				},
				seat: string(openrtb_ext.BidderIx),
			},
			want: want{
				adm: `sample_creative<script id="OWPubOMVerification" data-owurl="Tracking URL" src="${OMScript}"></script>`,
			},
		},
		{
			name: "app_platform_partner_other_than_pubmatic_imp_counting_method_disabled",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					ImpCountingMethodEnabledBidders: map[string]struct{}{
						string(openrtb_ext.BidderIx): {},
					},
				},
				tracker: models.OWTracker{
					TrackerURL: `Tracking URL`,
				},
				bid: openrtb2.Bid{
					AdM: `sample_creative`,
				},
				seat: string(openrtb_ext.BidderAppnexus),
			},
			want: want{
				adm: `sample_creative<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="Tracking URL"></div>`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdm, gotBurl := injectBannerTracker(tt.args.rctx, tt.args.tracker, tt.args.bid, tt.args.seat, tt.args.pixels)
			assert.Equal(t, tt.want.adm, gotAdm)
			assert.Equal(t, tt.want.burl, gotBurl)
		})
	}
}

func TestTrackerWithOM(t *testing.T) {
	type args struct {
		rctx              models.RequestCtx
		prebidPartnerName string
		dspID             int
		bidExt            json.RawMessage
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "in-app_partner_other_than_pubmatic",
			args: args{

				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
				},
				prebidPartnerName: "test",
			},
			want: false,
		},
		{
			name: "in-app_partner_pubmatic_other_dv360",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
				},
				prebidPartnerName: models.BidderPubMatic,
				dspID:             -1,
			},
			want: false,
		},
		{
			name: "display_partner_pubmatic_dv360",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_DISPLAY,
				},
				prebidPartnerName: models.BidderPubMatic,
				dspID:             models.DspId_DV360,
			},
			want: false,
		},
		{
			name: "in-app_partner_pubmatic_dv360",
			args: args{

				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
				},
				prebidPartnerName: models.BidderPubMatic,
				dspID:             models.DspId_DV360,
			},
			want: true,
		},
		{
			name: "in-app_partner_other_than_pubmatic_imp_counting_method_enabled",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					ImpCountingMethodEnabledBidders: map[string]struct{}{
						"ix": {},
					},
				},
				prebidPartnerName: "ix",
			},
			want: true,
		},
		{
			name: "in-app_partner_other_than_pubmatic_imp_counting_method_disabled",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
					ImpCountingMethodEnabledBidders: map[string]struct{}{
						"ix": {},
					},
				},
				prebidPartnerName: "magnite",
			},
			want: false,
		},
		{
			name: "in-app_partner_pubmatic_inview_enabled_publisher_performance_dsp",
			args: args{
				rctx: models.RequestCtx{
					Platform:                models.PLATFORM_APP,
					PubID:                   5890,
					InViewEnabledPublishers: map[int]struct{}{5890: {}},
					PerformanceDSPs:         map[int]struct{}{101: {}},
				},
				prebidPartnerName: models.BidderPubMatic,
				dspID:             101,
			},
			want: true,
		},
		{
			name: "in-app_partner_pubmatic_inview_enabled_publisher_non_performance_dsp",
			args: args{
				rctx: models.RequestCtx{
					Platform:                models.PLATFORM_APP,
					PubID:                   5890,
					InViewEnabledPublishers: map[int]struct{}{5890: {}},
					PerformanceDSPs:         map[int]struct{}{101: {}},
				},
				prebidPartnerName: models.BidderPubMatic,
				dspID:             999,
			},
			want: false,
		},
		{
			name: "in-app_partner_pubmatic_inview_disabled_publisher_non_performance_dsp",
			args: args{
				rctx: models.RequestCtx{
					Platform:                models.PLATFORM_APP,
					PubID:                   5890,
					InViewEnabledPublishers: map[int]struct{}{1234: {}},
					PerformanceDSPs:         map[int]struct{}{101: {}},
				},
				prebidPartnerName: models.BidderPubMatic,
				dspID:             999,
			},
			want: false,
		},
		{
			name: "in-app_partner_pubmatic_empty_inview_publishers_performance_dsp",
			args: args{
				rctx: models.RequestCtx{
					Platform:                models.PLATFORM_APP,
					PubID:                   5890,
					InViewEnabledPublishers: map[int]struct{}{},
					PerformanceDSPs:         map[int]struct{}{101: {}},
				},
				prebidPartnerName: models.BidderPubMatic,
				dspID:             101,
			},
			want: false,
		},
		{
			name: "in-app_partner_pubmatic_inview_enabled_empty_performance_dsps",
			args: args{
				rctx: models.RequestCtx{
					Platform:                models.PLATFORM_APP,
					PubID:                   5890,
					InViewEnabledPublishers: map[int]struct{}{5890: {}},
					PerformanceDSPs:         map[int]struct{}{},
				},
				prebidPartnerName: models.BidderPubMatic,
				dspID:             101,
			},
			want: false,
		},
		{
			name: "in-app_partner_other_than_pubmatic_with_imp_counting_method_flag_1",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
				},
				prebidPartnerName: "appnexus",
				bidExt:            json.RawMessage(`{"imp_ct_mthd":1}`),
			},
			want: true,
		},
		{
			name: "in-app_partner_other_than_pubmatic_with_invalid_bidext_json",
			args: args{
				rctx: models.RequestCtx{
					Platform: models.PLATFORM_APP,
				},
				prebidPartnerName: "appnexus",
				bidExt:            json.RawMessage(`{invalid json}`),
			},
			want: false,
		},
		{
			name: "in-app_partner_pubmatic_dv360_with_inview_and_performance_dsp",
			args: args{
				rctx: models.RequestCtx{
					Platform:                models.PLATFORM_APP,
					PubID:                   5890,
					InViewEnabledPublishers: map[int]struct{}{5890: {}},
					PerformanceDSPs:         map[int]struct{}{models.DspId_DV360: {}},
				},
				prebidPartnerName: models.BidderPubMatic,
				dspID:             models.DspId_DV360,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trackerWithOM(tt.args.rctx, tt.args.prebidPartnerName, tt.args.dspID, tt.args.bidExt); got != tt.want {
				assert.Equal(t, tt.want, got)
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
