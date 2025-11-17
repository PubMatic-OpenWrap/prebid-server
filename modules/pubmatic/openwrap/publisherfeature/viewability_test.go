package publisherfeature

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mock_cache "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache/mock"
	"github.com/stretchr/testify/assert"
)

func Test_newInViewEnabledPublishers(t *testing.T) {
	tests := []struct {
		name string
		want inViewEnabledPublishers
	}{
		{
			name: "initialize_inview_enabled_publishers_with_default_values",
			want: inViewEnabledPublishers{
				pubs: [2]map[int]struct{}{
					{},
					{},
				},
				index: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newInViewEnabledPublishers()
			assert.Equal(t, tt.want.index, got.index, "index should be 0")
			assert.NotNil(t, got.pubs[0], "pubs[0] should be initialized")
			assert.NotNil(t, got.pubs[1], "pubs[1] should be initialized")
			assert.Equal(t, 0, len(got.pubs[0]), "pubs[0] should be empty")
			assert.Equal(t, 0, len(got.pubs[1]), "pubs[1] should be empty")
		})
	}
}

func TestFeature_updateInViewEnabledPublishers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	type fields struct {
		inViewEnabledPublishers inViewEnabledPublishers
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
			name: "cache_returns_valid_publishers_updates_inactive_map_and_toggles_index",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{1000: {}},
						{2000: {}},
					},
					index: 0,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{
					1001: {},
					1002: {},
					1003: {},
				}, nil)
			},
			want: want{
				activeMap: map[int]struct{}{
					1001: {},
					1002: {},
					1003: {},
				},
				activeIdx: 1,
				inactiveMap: map[int]struct{}{
					1000: {},
				},
			},
		},
		{
			name: "cache_returns_error_no_update_occurs",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{1000: {}, 1001: {}},
						{2000: {}},
					},
					index: 0,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(nil, errors.New("DB error"))
			},
			want: want{
				activeMap: map[int]struct{}{
					1000: {},
					1001: {},
				},
				activeIdx: 0,
				inactiveMap: map[int]struct{}{
					2000: {},
				},
			},
		},
		{
			name: "cache_returns_nil_no_update_occurs",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{1000: {}},
						{2000: {}},
					},
					index: 0,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(nil, nil)
			},
			want: want{
				activeMap: map[int]struct{}{
					1000: {},
				},
				activeIdx: 0,
				inactiveMap: map[int]struct{}{
					2000: {},
				},
			},
		},
		{
			name: "cache_returns_empty_map_updates_and_toggles",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{1000: {}},
						{2000: {}},
					},
					index: 0,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{}, nil)
			},
			want: want{
				activeMap:   map[int]struct{}{},
				activeIdx:   1,
				inactiveMap: map[int]struct{}{1000: {}},
			},
		},
		{
			name: "toggle_from_index_1_to_0",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{1000: {}},
						{2000: {}, 2001: {}},
					},
					index: 1,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{
					3000: {},
					3001: {},
				}, nil)
			},
			want: want{
				activeMap: map[int]struct{}{
					3000: {},
					3001: {},
				},
				activeIdx: 0,
				inactiveMap: map[int]struct{}{
					2000: {},
					2001: {},
				},
			},
		},
		{
			name: "multiple_updates_toggle_correctly",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{},
						{},
					},
					index: 0,
				},
			},
			setup: func() {
				// First update
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{
					5890: {},
				}, nil)
			},
			want: want{
				activeMap: map[int]struct{}{
					5890: {},
				},
				activeIdx:   1,
				inactiveMap: map[int]struct{}{},
			},
		},
		{
			name: "large_publisher_set_update",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{},
						{},
					},
					index: 0,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{
					1001: {}, 1002: {}, 1003: {}, 1004: {}, 1005: {},
					1006: {}, 1007: {}, 1008: {}, 1009: {}, 1010: {},
					1011: {}, 1012: {}, 1013: {}, 1014: {}, 1015: {},
				}, nil)
			},
			want: want{
				activeMap: map[int]struct{}{
					1001: {}, 1002: {}, 1003: {}, 1004: {}, 1005: {},
					1006: {}, 1007: {}, 1008: {}, 1009: {}, 1010: {},
					1011: {}, 1012: {}, 1013: {}, 1014: {}, 1015: {},
				},
				activeIdx:   1,
				inactiveMap: map[int]struct{}{},
			},
		},
		{
			name: "single_publisher_update",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{1000: {}, 1001: {}},
						{},
					},
					index: 0,
				},
			},
			setup: func() {
				mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{
					5890: {},
				}, nil)
			},
			want: want{
				activeMap: map[int]struct{}{
					5890: {},
				},
				activeIdx: 1,
				inactiveMap: map[int]struct{}{
					1000: {},
					1001: {},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			fe := &feature{
				cache:                   mockCache,
				inViewEnabledPublishers: tt.fields.inViewEnabledPublishers,
			}
			fe.updateInViewEnabledPublishers()

			assert.Equal(t, tt.want.activeIdx, fe.inViewEnabledPublishers.index, "index should match expected")
			assert.Equal(t, tt.want.activeMap, fe.inViewEnabledPublishers.pubs[fe.inViewEnabledPublishers.index], "active map should match expected")
			assert.Equal(t, tt.want.inactiveMap, fe.inViewEnabledPublishers.pubs[fe.inViewEnabledPublishers.index^1], "inactive map should match expected")
		})
	}
}

func TestFeature_GetInViewEnabledPublishers(t *testing.T) {
	type fields struct {
		inViewEnabledPublishers inViewEnabledPublishers
	}
	tests := []struct {
		name   string
		fields fields
		want   map[int]struct{}
	}{
		{
			name: "returns_active_map_when_index_is_0",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{1001: {}, 1002: {}, 1003: {}},
						{2001: {}, 2002: {}},
					},
					index: 0,
				},
			},
			want: map[int]struct{}{
				1001: {},
				1002: {},
				1003: {},
			},
		},
		{
			name: "returns_active_map_when_index_is_1",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{1001: {}, 1002: {}},
						{2001: {}, 2002: {}, 2003: {}},
					},
					index: 1,
				},
			},
			want: map[int]struct{}{
				2001: {},
				2002: {},
				2003: {},
			},
		},
		{
			name: "returns_empty_map_when_active_is_empty",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{},
						{2001: {}},
					},
					index: 0,
				},
			},
			want: map[int]struct{}{},
		},
		{
			name: "returns_single_publisher",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{5890: {}},
						{},
					},
					index: 0,
				},
			},
			want: map[int]struct{}{
				5890: {},
			},
		},
		{
			name: "returns_large_publisher_set",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{},
						{
							1001: {}, 1002: {}, 1003: {}, 1004: {}, 1005: {},
							1006: {}, 1007: {}, 1008: {}, 1009: {}, 1010: {},
							1011: {}, 1012: {}, 1013: {}, 1014: {}, 1015: {},
							1016: {}, 1017: {}, 1018: {}, 1019: {}, 1020: {},
						},
					},
					index: 1,
				},
			},
			want: map[int]struct{}{
				1001: {}, 1002: {}, 1003: {}, 1004: {}, 1005: {},
				1006: {}, 1007: {}, 1008: {}, 1009: {}, 1010: {},
				1011: {}, 1012: {}, 1013: {}, 1014: {}, 1015: {},
				1016: {}, 1017: {}, 1018: {}, 1019: {}, 1020: {},
			},
		},
		{
			name: "returns_map_with_zero_and_negative_ids",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{-1: {}, 0: {}, 1000: {}},
						{},
					},
					index: 0,
				},
			},
			want: map[int]struct{}{
				-1:   {},
				0:    {},
				1000: {},
			},
		},
		{
			name: "returns_map_with_mixed_publisher_ids",
			fields: fields{
				inViewEnabledPublishers: inViewEnabledPublishers{
					pubs: [2]map[int]struct{}{
						{5890: {}, 1001: {}, 9999: {}, 100: {}},
						{},
					},
					index: 0,
				},
			},
			want: map[int]struct{}{
				5890: {},
				1001: {},
				9999: {},
				100:  {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := &feature{
				inViewEnabledPublishers: tt.fields.inViewEnabledPublishers,
			}
			got := fe.GetInViewEnabledPublishers()
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestFeature_updateInViewEnabledPublishers_DoubleBuffering(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	t.Run("double_buffering_prevents_race_conditions", func(t *testing.T) {
		fe := &feature{
			cache: mockCache,
			inViewEnabledPublishers: inViewEnabledPublishers{
				pubs: [2]map[int]struct{}{
					{1000: {}},
					{2000: {}},
				},
				index: 0,
			},
		}

		// First update
		mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{
			1001: {},
			1002: {},
		}, nil)

		// Get current active before update
		activeBefore := fe.GetInViewEnabledPublishers()
		assert.Equal(t, map[int]struct{}{1000: {}}, activeBefore)

		// Update
		fe.updateInViewEnabledPublishers()

		// Get current active after update
		activeAfter := fe.GetInViewEnabledPublishers()
		assert.Equal(t, map[int]struct{}{1001: {}, 1002: {}}, activeAfter)
		assert.Equal(t, 1, fe.inViewEnabledPublishers.index)

		// Second update
		mockCache.EXPECT().GetInViewEnabledPublishers().Return(map[int]struct{}{
			2001: {},
			2002: {},
			2003: {},
		}, nil)

		fe.updateInViewEnabledPublishers()

		// Get current active after second update
		activeAfterSecond := fe.GetInViewEnabledPublishers()
		assert.Equal(t, map[int]struct{}{2001: {}, 2002: {}, 2003: {}}, activeAfterSecond)
		assert.Equal(t, 0, fe.inViewEnabledPublishers.index)
	})
}

func TestFeature_updateInViewEnabledPublishers_EdgeCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCache := mock_cache.NewMockCache(ctrl)

	t.Run("error_and_nil_both_prevent_update", func(t *testing.T) {
		fe := &feature{
			cache: mockCache,
			inViewEnabledPublishers: inViewEnabledPublishers{
				pubs: [2]map[int]struct{}{
					{1000: {}, 1001: {}},
					{},
				},
				index: 0,
			},
		}

		// Test with error
		mockCache.EXPECT().GetInViewEnabledPublishers().Return(nil, errors.New("DB connection failed"))
		fe.updateInViewEnabledPublishers()
		assert.Equal(t, 0, fe.inViewEnabledPublishers.index, "index should not change on error")
		assert.Equal(t, map[int]struct{}{1000: {}, 1001: {}}, fe.GetInViewEnabledPublishers())

		// Test with nil
		mockCache.EXPECT().GetInViewEnabledPublishers().Return(nil, nil)
		fe.updateInViewEnabledPublishers()
		assert.Equal(t, 0, fe.inViewEnabledPublishers.index, "index should not change on nil")
		assert.Equal(t, map[int]struct{}{1000: {}, 1001: {}}, fe.GetInViewEnabledPublishers())
	})
}
