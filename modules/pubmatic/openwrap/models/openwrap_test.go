package models

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func TestRequestCtx_GetVersionLevelKey(t *testing.T) {
	type fields struct {
		PubID                     int
		ProfileID                 int
		DisplayID                 int
		VersionID                 int
		SSAuction                 int
		SummaryDisable            int
		LogInfoFlag               int
		SSAI                      string
		PartnerConfigMap          map[int]map[string]string
		SupportDeals              bool
		Platform                  string
		LoggerImpressionID        string
		ClientConfigFlag          int
		IP                        string
		TMax                      int64
		IsTestRequest             int8
		ABTestConfig              int
		ABTestConfigApplied       int
		IsCTVRequest              bool
		TrackerEndpoint           string
		VideoErrorTrackerEndpoint string
		UA                        string
		Cookies                   string
		UidCookie                 *http.Cookie
		KADUSERCookie             *http.Cookie
		OriginCookie              string
		Debug                     bool
		Trace                     bool
		PageURL                   string
		StartTime                 int64
		DevicePlatform            DevicePlatform
		Trackers                  map[string]OWTracker
		PrebidBidderCode          map[string]string
		ImpBidCtx                 map[string]ImpCtx
		Aliases                   map[string]string
		NewReqExt                 json.RawMessage
		ResponseExt               json.RawMessage
		MarketPlaceBidders        map[string]struct{}
		AdapterThrottleMap        map[string]struct{}
		AdUnitConfig              *adunitconfig.AdUnitConfig
		Source                    string
		Origin                    string
		SendAllBids               bool
		WinningBids               map[string][]OwBid
		DroppedBids               map[string][]openrtb2.Bid
		DefaultBids               map[string]map[string][]openrtb2.Bid
		SeatNonBids               map[string][]openrtb_ext.NonBid
		BidderResponseTimeMillis  map[string]int
		Endpoint                  string
		PubIDStr                  string
		ProfileIDStr              string
		MetricsEngine             metrics.MetricsEngine
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "get_version_level_platform_key",
			fields: fields{
				PartnerConfigMap: map[int]map[string]string{
					-1: {
						"platform": "in-app",
					},
				},
			},
			args: args{
				key: "platform",
			},
			want: "in-app",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RequestCtx{
				PubID:                     tt.fields.PubID,
				ProfileID:                 tt.fields.ProfileID,
				DisplayID:                 tt.fields.DisplayID,
				VersionID:                 tt.fields.VersionID,
				SSAuction:                 tt.fields.SSAuction,
				SummaryDisable:            tt.fields.SummaryDisable,
				LogInfoFlag:               tt.fields.LogInfoFlag,
				SSAI:                      tt.fields.SSAI,
				PartnerConfigMap:          tt.fields.PartnerConfigMap,
				SupportDeals:              tt.fields.SupportDeals,
				Platform:                  tt.fields.Platform,
				LoggerImpressionID:        tt.fields.LoggerImpressionID,
				ClientConfigFlag:          tt.fields.ClientConfigFlag,
				IP:                        tt.fields.IP,
				TMax:                      tt.fields.TMax,
				IsTestRequest:             tt.fields.IsTestRequest,
				ABTestConfig:              tt.fields.ABTestConfig,
				ABTestConfigApplied:       tt.fields.ABTestConfigApplied,
				IsCTVRequest:              tt.fields.IsCTVRequest,
				TrackerEndpoint:           tt.fields.TrackerEndpoint,
				VideoErrorTrackerEndpoint: tt.fields.VideoErrorTrackerEndpoint,
				UA:                        tt.fields.UA,
				Cookies:                   tt.fields.Cookies,
				UidCookie:                 tt.fields.UidCookie,
				KADUSERCookie:             tt.fields.KADUSERCookie,
				OriginCookie:              tt.fields.OriginCookie,
				Debug:                     tt.fields.Debug,
				Trace:                     tt.fields.Trace,
				PageURL:                   tt.fields.PageURL,
				StartTime:                 tt.fields.StartTime,
				DevicePlatform:            tt.fields.DevicePlatform,
				Trackers:                  tt.fields.Trackers,
				PrebidBidderCode:          tt.fields.PrebidBidderCode,
				ImpBidCtx:                 tt.fields.ImpBidCtx,
				Aliases:                   tt.fields.Aliases,
				NewReqExt:                 tt.fields.NewReqExt,
				ResponseExt:               tt.fields.ResponseExt,
				MarketPlaceBidders:        tt.fields.MarketPlaceBidders,
				AdapterThrottleMap:        tt.fields.AdapterThrottleMap,
				AdUnitConfig:              tt.fields.AdUnitConfig,
				Source:                    tt.fields.Source,
				Origin:                    tt.fields.Origin,
				SendAllBids:               tt.fields.SendAllBids,
				WinningBids:               tt.fields.WinningBids,
				DroppedBids:               tt.fields.DroppedBids,
				DefaultBids:               tt.fields.DefaultBids,
				SeatNonBids:               tt.fields.SeatNonBids,
				BidderResponseTimeMillis:  tt.fields.BidderResponseTimeMillis,
				Endpoint:                  tt.fields.Endpoint,
				PubIDStr:                  tt.fields.PubIDStr,
				ProfileIDStr:              tt.fields.ProfileIDStr,
				MetricsEngine:             tt.fields.MetricsEngine,
			}
			if got := r.GetVersionLevelKey(tt.args.key); got != tt.want {
				t.Errorf("RequestCtx.GetVersionLevelKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
