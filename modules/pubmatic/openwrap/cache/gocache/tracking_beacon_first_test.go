package gocache

import (
	"testing"

	mock_database "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/database/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetTBFTrafficForPublishers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDatabase := mock_database.NewMockDatabase(ctrl)

	tests := []struct {
		name               string
		setup              func(dbCache *cache)
		wantTrafficDetails map[int]map[int]int
		wantErr            error
	}{
		{
			name: "test_call_forwarding",
			setup: func(dbCache *cache) {
				mockDatabase.EXPECT().GetTBFTrafficForPublishers().Return(map[int]map[int]int{5890: {1234: 100}}, nil)
				dbCache.db = mockDatabase
			},
			wantTrafficDetails: map[int]map[int]int{5890: {1234: 100}},
			wantErr:            nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbCache := cache{}
			tt.setup(&dbCache)
			actualTrafficDetails, err := dbCache.GetTBFTrafficForPublishers()
			assert.Equal(t, actualTrafficDetails, tt.wantTrafficDetails, tt.name)
			assert.Equal(t, tt.wantErr, err, tt.name)
		})
	}
}
