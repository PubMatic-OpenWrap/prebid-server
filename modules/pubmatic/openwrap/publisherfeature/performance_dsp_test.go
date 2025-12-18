package publisherfeature

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	"github.com/stretchr/testify/assert"
)

func Test_newPerformanceDSPs(t *testing.T) {
	tests := []struct {
		name string
		want performanceDSPs
	}{
		{
			name: "initialize_performance_dsps_with_default_values",
			want: performanceDSPs{
				dsps: [2]map[int]struct{}{
					{},
					{},
				},
				index: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newPerformanceDSPs()
			assert.Equal(t, tt.want.index, got.index, "index should be 0")
			assert.NotNil(t, got.dsps[0], "dsps[0] should be initialized")
			assert.NotNil(t, got.dsps[1], "dsps[1] should be initialized")
			assert.Equal(t, 0, len(got.dsps[0]), "dsps[0] should be empty")
			assert.Equal(t, 0, len(got.dsps[1]), "dsps[1] should be empty")
		})
	}
}

func TestFeature_updatePerformanceDSPs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type fields struct {
		performanceDSPs performanceDSPs
	}
	type want struct {
		activeMap   map[int]struct{}
		activeIdx   int
		inactiveMap map[int]struct{}
	}
	tests := []struct {
		name   string
		fields fields
		setup  func()
		want   want
	}{
		{
			name: "cache_returns_valid_dsps_updates_inactive_map_and_toggles_index",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{100: {}},
						{200: {}},
					},
					index: 0,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{
					101: {},
					102: {},
					103: {},
				}, nil)
			},
			want: want{
				activeMap: map[int]struct{}{
					101: {},
					102: {},
					103: {},
				},
				activeIdx: 1,
				inactiveMap: map[int]struct{}{
					100: {},
				},
			},
		},
		{
			name: "cache_returns_error_no_update_occurs",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{100: {}, 101: {}},
						{200: {}},
					},
					index: 0,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetPerformanceDSPs().Return(nil, errors.New("DB error"))
			},
			want: want{
				activeMap: map[int]struct{}{
					100: {},
					101: {},
				},
				activeIdx: 0,
				inactiveMap: map[int]struct{}{
					200: {},
				},
			},
		},
		{
			name: "cache_returns_nil_no_update_occurs",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{100: {}},
						{200: {}},
					},
					index: 0,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetPerformanceDSPs().Return(nil, nil)
			},
			want: want{
				activeMap: map[int]struct{}{
					100: {},
				},
				activeIdx: 0,
				inactiveMap: map[int]struct{}{
					200: {},
				},
			},
		},
		{
			name: "cache_returns_empty_map_updates_and_toggles",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{100: {}},
						{200: {}},
					},
					index: 0,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{}, nil)
			},
			want: want{
				activeMap:   map[int]struct{}{},
				activeIdx:   1,
				inactiveMap: map[int]struct{}{100: {}},
			},
		},
		{
			name: "toggle_from_index_1_to_0",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{100: {}},
						{200: {}, 201: {}},
					},
					index: 1,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{
					300: {},
					301: {},
				}, nil)
			},
			want: want{
				activeMap: map[int]struct{}{
					300: {},
					301: {},
				},
				activeIdx: 0,
				inactiveMap: map[int]struct{}{
					200: {},
					201: {},
				},
			},
		},
		{
			name: "multiple_updates_toggle_correctly",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{},
						{},
					},
					index: 0,
				},
			},
			setup: func() {
				// First update
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{
					101: {},
				}, nil)
			},
			want: want{
				activeMap: map[int]struct{}{
					101: {},
				},
				activeIdx:   1,
				inactiveMap: map[int]struct{}{},
			},
		},
		{
			name: "large_dsp_set_update",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{},
						{},
					},
					index: 0,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{
					1: {}, 2: {}, 3: {}, 4: {}, 5: {},
					6: {}, 7: {}, 8: {}, 9: {}, 10: {},
				}, nil)
			},
			want: want{
				activeMap: map[int]struct{}{
					1: {}, 2: {}, 3: {}, 4: {}, 5: {},
					6: {}, 7: {}, 8: {}, 9: {}, 10: {},
				},
				activeIdx:   1,
				inactiveMap: map[int]struct{}{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			fe := &feature{
				cache:           mockCache,
				performanceDSPs: tt.fields.performanceDSPs,
			}
			fe.updatePerformanceDSPs()

			assert.Equal(t, tt.want.activeIdx, fe.performanceDSPs.index, "index should match expected")
			assert.Equal(t, tt.want.activeMap, fe.performanceDSPs.dsps[fe.performanceDSPs.index], "active map should match expected")
			assert.Equal(t, tt.want.inactiveMap, fe.performanceDSPs.dsps[fe.performanceDSPs.index^1], "inactive map should match expected")
		})
	}
}

func TestFeature_GetEnabledPerformanceDSPs(t *testing.T) {
	type fields struct {
		performanceDSPs performanceDSPs
	}
	tests := []struct {
		name   string
		fields fields
		want   map[int]struct{}
	}{
		{
			name: "returns_active_map_when_index_is_0",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{101: {}, 102: {}, 103: {}},
						{201: {}, 202: {}},
					},
					index: 0,
				},
			},
			want: map[int]struct{}{
				101: {},
				102: {},
				103: {},
			},
		},
		{
			name: "returns_active_map_when_index_is_1",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{101: {}, 102: {}},
						{201: {}, 202: {}, 203: {}},
					},
					index: 1,
				},
			},
			want: map[int]struct{}{
				201: {},
				202: {},
				203: {},
			},
		},
		{
			name: "returns_empty_map_when_active_is_empty",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{},
						{201: {}},
					},
					index: 0,
				},
			},
			want: map[int]struct{}{},
		},
		{
			name: "returns_single_dsp",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{100: {}},
						{},
					},
					index: 0,
				},
			},
			want: map[int]struct{}{
				100: {},
			},
		},
		{
			name: "returns_large_dsp_set",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{},
						{
							1: {}, 2: {}, 3: {}, 4: {}, 5: {},
							6: {}, 7: {}, 8: {}, 9: {}, 10: {},
							11: {}, 12: {}, 13: {}, 14: {}, 15: {},
						},
					},
					index: 1,
				},
			},
			want: map[int]struct{}{
				1: {}, 2: {}, 3: {}, 4: {}, 5: {},
				6: {}, 7: {}, 8: {}, 9: {}, 10: {},
				11: {}, 12: {}, 13: {}, 14: {}, 15: {},
			},
		},
		{
			name: "returns_map_with_negative_ids",
			fields: fields{
				performanceDSPs: performanceDSPs{
					dsps: [2]map[int]struct{}{
						{-1: {}, 0: {}, 100: {}},
						{},
					},
					index: 0,
				},
			},
			want: map[int]struct{}{
				-1:  {},
				0:   {},
				100: {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				performanceDSPs: tt.fields.performanceDSPs,
			}
			got := fe.GetEnabledPerformanceDSPs()
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestFeature_updatePerformanceDSPs_DoubleBuffering(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	t.Run("double_buffering_prevents_race_conditions", func(t *testing.T) {
		fe := &feature{
			cache: mockCache,
			performanceDSPs: performanceDSPs{
				dsps: [2]map[int]struct{}{
					{100: {}},
					{200: {}},
				},
				index: 0,
			},
		}

		// First update
		mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{
			101: {},
			102: {},
		}, nil)

		// Get current active before update
		activeBefore := fe.GetEnabledPerformanceDSPs()
		assert.Equal(t, map[int]struct{}{100: {}}, activeBefore)

		// Update
		fe.updatePerformanceDSPs()

		// Get current active after update
		activeAfter := fe.GetEnabledPerformanceDSPs()
		assert.Equal(t, map[int]struct{}{101: {}, 102: {}}, activeAfter)
		assert.Equal(t, 1, fe.performanceDSPs.index)

		// Second update
		mockCache.EXPECT().GetPerformanceDSPs().Return(map[int]struct{}{
			201: {},
			202: {},
			203: {},
		}, nil)

		fe.updatePerformanceDSPs()

		// Get current active after second update
		activeAfterSecond := fe.GetEnabledPerformanceDSPs()
		assert.Equal(t, map[int]struct{}{201: {}, 202: {}, 203: {}}, activeAfterSecond)
		assert.Equal(t, 0, fe.performanceDSPs.index)
	})
}
