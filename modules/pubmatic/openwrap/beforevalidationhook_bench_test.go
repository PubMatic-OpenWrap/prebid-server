package openwrap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v4/currency"
	"github.com/prebid/prebid-server/v4/hooks/hookstage"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/adapters"
	mock_cache "github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/cache/mock"
	mock_metrics "github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/models/adunitconfig"
	mock_profilemetadata "github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/profilemetadata/mock"
	mock_feature "github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/publisherfeature/mock"
	"github.com/prebid/prebid-server/v4/modules/pubmatic/openwrap/wakanda"
	"github.com/prebid/prebid-server/v4/openrtb_ext"
)

func BenchmarkHandleBeforeValidationHook_1Imp1Partner(b *testing.B) {
	benchmarkHandleBeforeValidationHook(b, 1, 1, false)
}

func BenchmarkHandleBeforeValidationHook_5Imp8Partner(b *testing.B) {
	benchmarkHandleBeforeValidationHook(b, 5, 8, false)
}

func BenchmarkHandleBeforeValidationHook_RegexKGP(b *testing.B) {
	benchmarkHandleBeforeValidationHook(b, 1, 1, true)
}

func benchmarkHandleBeforeValidationHook(b *testing.B, impCount, partnerCount int, regexKGP bool) {
	b.Helper()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockCache := mock_cache.NewMockCache(ctrl)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	mockFeature := mock_feature.NewMockFeature(ctrl)
	mockProfileMetaData := mock_profilemetadata.NewMockProfileMetaData(ctrl)

	mockEngine.EXPECT().RecordPublisherProfileRequests(gomock.Any(), gomock.Any()).AnyTimes()
	mockEngine.EXPECT().RecordPublisherRequests(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockEngine.EXPECT().RecordPlatformPublisherPartnerReqStats(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockEngine.EXPECT().RecordPreProcessingTimeStats(gomock.Any(), gomock.Any()).AnyTimes()
	mockEngine.EXPECT().RecordPreProcessingPhaseTime(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockEngine.EXPECT().RecordPartnerConfigErrors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockEngine.EXPECT().RecordUidsCookieNotPresentErrorStats(gomock.Any(), gomock.Any()).AnyTimes()
	mockEngine.EXPECT().RecordBadRequests(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockEngine.EXPECT().RecordNobidErrPrebidServerRequests(gomock.Any(), gomock.Any()).AnyTimes()
	mockEngine.EXPECT().RecordPublisherInvalidProfileRequests(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockEngine.EXPECT().RecordImpDisabledViaConfigStats(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockFeature.EXPECT().IsEDSBlockedCountry(gomock.Any()).Return(false).AnyTimes()
	mockFeature.EXPECT().IsDynamicFloorEnabledPublisher(gomock.Any()).Return(false).AnyTimes()
	mockFeature.EXPECT().IsTBFFeatureEnabled(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
	mockFeature.EXPECT().IsAnalyticsTrackingThrottled(gomock.Any(), gomock.Any()).Return(false, false).AnyTimes()
	mockFeature.EXPECT().IsMBMFPublisherEnabled(gomock.Any()).Return(false).AnyTimes()
	mockProfileMetaData.EXPECT().GetProfileTypePlatform(gomock.Any()).Return(0, false).AnyTimes()
	mockProfileMetaData.EXPECT().GetAppIntegrationPath(gomock.Any()).Return(0, false).AnyTimes()
	mockProfileMetaData.EXPECT().GetAppSubIntegrationPath(gomock.Any()).Return(models.AppSubIntegrationPath{}, false).AnyTimes()

	partnerConfigMap := make(map[int]map[string]string, partnerCount+1)
	partnerConfigMap[models.VersionLevelConfigID] = map[string]string{
		models.DisplayVersionID: "1",
		models.PLATFORM_KEY:     models.PLATFORM_APP,
	}
	kgp := models.ADUNIT_SIZE_KGP
	if regexKGP {
		kgp = models.REGEX_KGP
	}
	for i := 1; i <= partnerCount; i++ {
		cfg := map[string]string{
			models.PARTNER_ID:          fmt.Sprintf("%d", i),
			models.PREBID_PARTNER_NAME: "appnexus",
			models.BidderCode:          "appnexus",
			models.SERVER_SIDE_FLAG:    "1",
			models.KEY_GEN_PATTERN:     kgp,
			models.TIMEOUT:             "200",
		}
		if i > 1 {
			cfg[models.BidderCode] = fmt.Sprintf("appnexus%d", i)
			cfg[models.IsAlias] = "1"
		}
		partnerConfigMap[i] = cfg
	}
	mockCache.EXPECT().GetPartnerConfigMap(gomock.Any(), gomock.Any(), gomock.Any()).Return(partnerConfigMap, nil).AnyTimes()
	mockCache.EXPECT().GetAdunitConfigFromCache(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&adunitconfig.AdUnitConfig{}).AnyTimes()
	mockCache.EXPECT().GetThrottlePartnersWithCriteria(gomock.Any()).Return(map[string]struct{}{}, nil).AnyTimes()

	slotName := "adunit@300x250"
	if regexKGP {
		mockCache.EXPECT().Get(gomock.Any()).Return(nil, false).AnyTimes()
		mockCache.EXPECT().Set(gomock.Any(), gomock.Any()).AnyTimes()
		mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
			"adunit.*@div.*@.*": {
				SlotName: "adunit.*@div.*@.*",
				SlotMappings: map[string]interface{}{
					models.SITE_CACHE_KEY: "12313",
					models.TAG_CACHE_KEY:  "45343",
				},
			},
		}).AnyTimes()
		mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
			OrderedSlotList: []string{"adunit.*@div.*@.*"},
			HashValueMap:    map[string]string{"adunit.*@div.*@.*": "hash"},
		}).AnyTimes()
	} else {
		mockCache.EXPECT().GetMappingsFromCacheV25(gomock.Any(), gomock.Any()).Return(map[string]models.SlotMapping{
			slotName: {
				SlotName: slotName,
				SlotMappings: map[string]interface{}{
					models.SITE_CACHE_KEY: "12313",
					models.TAG_CACHE_KEY:  "45343",
				},
			},
		}).AnyTimes()
		mockCache.EXPECT().GetSlotToHashValueMapFromCacheV25(gomock.Any(), gomock.Any()).Return(models.SlotMappingInfo{
			OrderedSlotList: []string{slotName},
			HashValueMap:    map[string]string{slotName: "hash"},
		}).AnyTimes()
	}

	adapters.InitBidders("./static/bidder-params/")
	m := OpenWrap{
		cache:           mockCache,
		metricEngine:    mockEngine,
		pubFeatures:     mockFeature,
		rateConvertor:   &currency.RateConverter{},
		profileMetaData: mockProfileMetaData,
	}

	impExt := json.RawMessage(`{"prebid":{}}`)
	if regexKGP {
		impExt = json.RawMessage(`{"wrapper":{"div":"div1"},"prebid":{}}`)
	}
	imps := make([]openrtb2.Imp, impCount)
	for i := 0; i < impCount; i++ {
		imps[i] = openrtb2.Imp{
			ID:    fmt.Sprintf("imp-%d", i+1),
			TagID: "adunit",
			Banner: &openrtb2.Banner{
				W:      ptrInt64(300),
				H:      ptrInt64(250),
				Format: []openrtb2.Format{{W: 300, H: 250}},
			},
			Ext: impExt,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reqCtx := cloneBenchRequestCtx()
		payload := hookstage.BeforeValidationRequestPayload{
			BidRequest: &openrtb2.BidRequest{
				ID:  "bench",
				Imp: append([]openrtb2.Imp(nil), imps...),
				Site: &openrtb2.Site{
					Domain:    "test.com",
					Page:      "www.test.com",
					Publisher: &openrtb2.Publisher{ID: "5890"},
				},
				Device: &openrtb2.Device{
					UA: "Mozilla/5.0",
					IP: "123.145.167.10",
				},
				Ext: json.RawMessage(`{"prebid":{},"wrapper":{"profileid":1234,"versionid":1}}`),
			},
		}
		moduleCtx := hookstage.ModuleInvocationContext{
			ModuleContext: moduleContextFromMap(map[string]any{"rctx": reqCtx}),
		}
		b.StartTimer()
		result, err := m.handleBeforeValidationHook(context.Background(), moduleCtx, payload)
		if err != nil {
			b.Fatal(err)
		}
		if result.Reject {
			b.Fatalf("hook rejected NbrCode=%d errors=%v", result.NbrCode, result.Errors)
		}
		for _, mut := range result.ChangeSet.Mutations() {
			if _, err := mut.Apply(payload); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func cloneBenchRequestCtx() models.RequestCtx {
	return models.RequestCtx{
		PubIDStr:     "5890",
		PubID:        5890,
		ProfileID:    1234,
		DisplayID:    1,
		ProfileIDStr: "1234",
		Platform:     "in-app",
		Endpoint:     models.EndpointV25,
		DeviceCtx:    models.DeviceCtx{UA: "go-test", IP: "127.0.0.1", Country: "IND", DerivedCountryCode: "IN"},
		UidCookie: &http.Cookie{
			Name:  "uids",
			Value: "bench",
		},
		Aliases:          make(map[string]string),
		ImpBidCtx:        make(map[string]models.ImpCtx),
		PrebidBidderCode: make(map[string]string),
		SeatNonBids:      make(map[string][]openrtb_ext.NonBid),
		WakandaDebug:     &wakanda.Debug{},
	}
}

func ptrInt64(v int64) *int64 { return &v }
