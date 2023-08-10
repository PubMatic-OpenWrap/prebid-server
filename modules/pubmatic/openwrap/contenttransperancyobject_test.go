package openwrap

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func TestSetContentTransparencyObject(t *testing.T) {
	type args struct {
		rctx   models.RequestCtx
		reqExt models.RequestExt
	}
	tests := []struct {
		name                   string
		args                   args
		wantPrebidTransparency *openrtb_ext.TransparencyExt
	}{
		{
			name: "Transparency object present in request",
			args: args{
				reqExt: models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{},
						},
					},
				},
			},
			wantPrebidTransparency: nil,
		},
		{
			name: "Transparency object not present in request and AdUnit",
			args: args{
				reqExt: models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{},
					},
				},
				rctx: models.RequestCtx{
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BannerAdUnitCtx: models.AdUnitCtx{},
						},
					},
				},
			},
			wantPrebidTransparency: nil,
		},
		{
			name: "All bidders throttled",
			args: args{
				reqExt: models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{},
					},
				},
				rctx: models.RequestCtx{
					Source: "test.com",
					PartnerConfigMap: map[int]map[string]string{
						1: {models.BidderCode: "123", "serverSideEnabled": "1"},
						2: {models.BidderCode: "456", "serverSideEnabled": "0"},
					},
					AdapterThrottleMap: map[string]struct{}{
						"123": {},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							BannerAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Transparency: &adunitconfig.Transparency{
										Content: adunitconfig.Content{
											Mappings: map[string]openrtb_ext.TransparencyRule{
												"test.com|pubmatic": {
													Include: true,
													Keys:    []string{"title"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantPrebidTransparency: nil,
		},
		{
			name: "Content Object Present",
			args: args{
				reqExt: models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{},
					},
				},
				rctx: models.RequestCtx{
					Source: "test.com",
					PartnerConfigMap: map[int]map[string]string{
						1: {models.BidderCode: "pubmatic", "serverSideEnabled": "1"},
						2: {models.BidderCode: "456", "serverSideEnabled": "1"},
					},
					AdapterThrottleMap: map[string]struct{}{
						"456": {},
					},
					ImpBidCtx: map[string]models.ImpCtx{
						"imp1": {
							VideoAdUnitCtx: models.AdUnitCtx{
								AppliedSlotAdUnitConfig: &adunitconfig.AdConfig{
									Transparency: &adunitconfig.Transparency{
										Content: adunitconfig.Content{
											Mappings: map[string]openrtb_ext.TransparencyRule{
												"test.com|pubmatic": {
													Include: true,
													Keys:    []string{"title"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantPrebidTransparency: &openrtb_ext.TransparencyExt{
				Content: map[string]openrtb_ext.TransparencyRule{
					"pubmatic": {
						Include: true,
						Keys:    []string{"title"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrebidTransparency := setContentTransparencyObject(tt.args.rctx, tt.args.reqExt)
			assert.Equal(t, gotPrebidTransparency, tt.wantPrebidTransparency, tt.name)
		})
	}
}
