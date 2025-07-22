package gocache

import (
	"testing"

	"github.com/golang/mock/gomock"
	mock_database "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/database/mock"
	"github.com/stretchr/testify/assert"
)

func TestGetThrottlePartnersWithCriteria(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		db *mock_database.MockDatabase
	}
	type args struct {
		country string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "DB_Not_Initialized",
			fields: fields{
				db: nil,
			},
			args:    args{"US"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty_Partner_Filter_Cache",
			fields: func() fields {
				mockDB := mock_database.NewMockDatabase(ctrl)
				mockDB.EXPECT().GetLatestCountryPartnerFilter().Return(nil).AnyTimes()
				return fields{db: mockDB}
			}(),
			args:    args{"US"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Multiple_Matching_Records",
			fields: func() fields {
				mockDB := mock_database.NewMockDatabase(ctrl)
				mockDB.EXPECT().GetLatestCountryPartnerFilter().Return(map[string][]string{
					"US": {
						"partner1",
						"partner2",
					},
				}).AnyTimes()
				return fields{db: mockDB}
			}(),
			args:    args{"US"},
			want:    []string{"partner1", "partner2"},
			wantErr: false,
		},
		{
			name: "	Mismatching_country",
			fields: func() fields {
				mockDB := mock_database.NewMockDatabase(ctrl)
				mockDB.EXPECT().GetLatestCountryPartnerFilter().Return(map[string][]string{
					"US": {
						"partner1",
						"partner2",
					},
				}).AnyTimes()
				return fields{db: mockDB}
			}(),
			args:    args{"IN"},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cache{
				db: tt.fields.db,
			}
			got, err := c.GetThrottlePartnersWithCriteria(tt.args.country)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
