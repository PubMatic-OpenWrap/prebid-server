package tracker

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prebid/openrtb/v19/openrtb2"
	mock_metrics "github.com/prebid/prebid-server/modules/pubmatic/openwrap/metrics/mock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func TestInjectTrackers(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockEngine := mock_metrics.NewMockMetricsEngine(ctrl)
	defer ctrl.Finish()

	type args struct {
		rctx        models.RequestCtx
		bidResponse *openrtb2.BidResponse
	}
	tests := []struct {
		name    string
		args    args
		want    *openrtb2.BidResponse
		wantErr bool
	}{
		{
			name: "no_bidresponse",
			args: args{
				bidResponse: &openrtb2.BidResponse{},
			},
			want:    &openrtb2.BidResponse{},
			wantErr: false,
		},
		{
			name: "invalid_adformat",
			args: args{
				rctx: models.RequestCtx{
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "invalid",
							TrackerURL: `Tracking URL`,
						},
					},
					MetricsEngine: mockEngine,
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `creative`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `creative`,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty_tracker_params",
			args: args{
				rctx: models.RequestCtx{
					MetricsEngine: mockEngine,
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `creative`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `creative`,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "adformat_is_banner",
			args: args{
				rctx: models.RequestCtx{
					Platform: "",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "banner",
							TrackerURL: `Tracking URL`,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `sample_creative`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `sample_creative<div style="position:absolute;left:0px;top:0px;visibility:hidden;"><img src="Tracking URL"></div>`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "adformat_is_video",
			args: args{
				rctx: models.RequestCtx{
					Platform: "",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "video",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `<VAST version="3.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `<VAST version="3.0"><Ad><Wrapper><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error></Wrapper></Ad></VAST>`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "platform_is_video",
			args: args{
				rctx: models.RequestCtx{
					Platform: "video",
					Trackers: map[string]models.OWTracker{
						"12345": {
							BidType:    "video",
							TrackerURL: `Tracking URL`,
							ErrorURL:   `Error URL`,
						},
					},
				},
				bidResponse: &openrtb2.BidResponse{
					SeatBid: []openrtb2.SeatBid{
						{
							Bid: []openrtb2.Bid{
								{
									ID:  "12345",
									AdM: `<VAST version="3.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
								},
							},
						},
					},
				},
			},
			want: &openrtb2.BidResponse{
				SeatBid: []openrtb2.SeatBid{
					{
						Bid: []openrtb2.Bid{
							{
								ID:  "12345",
								AdM: `<VAST version="3.0"><Ad><Wrapper><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error></Wrapper></Ad></VAST>`,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.rctx.MetricsEngine != nil {
				mockEngine.EXPECT().RecordInjectTrackerErrorCount(gomock.Any(), gomock.Any(), gomock.Any())
			}
			got, err := InjectTrackers(tt.args.rctx, tt.args.bidResponse)
			if (err != nil) != tt.wantErr {
				t.Errorf("InjectTrackers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InjectTrackers() = %v, want %v", got, tt.want)
			}
		})
	}
}
