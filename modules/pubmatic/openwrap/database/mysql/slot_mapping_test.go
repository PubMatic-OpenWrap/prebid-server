package mysql

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func Test_mySqlDB_GetPublisherSlotNameHash(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	type args struct {
		pubID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr error
		setup   func() *sql.DB
	}{
		{
			name:    "empty query in config file",
			want:    map[string]string{},
			wantErr: errors.New("context deadline exceeded"),
			setup: func() *sql.DB {
				db, _, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				return db
			},
		},
		{
			name: "duplicate slotname in publisher slotnamehash",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetSlotNameHash: "^SELECT (.+) FROM  wrapper_publisher_slot (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubID: 5890,
			},
			want: map[string]string{
				"/43743431/DMDemo1@160x600": "2fb84286ede5b20e82b0601df0c7e454",
				"/43743431/DMDemo2@160x600": "2aa34b52a9e941c1594af7565e599c8d",
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "hash"}).
					AddRow("/43743431/DMDemo1@160x600", "eb15e9be2d65f0268ff498572d3bb53e").
					AddRow("/43743431/DMDemo1@160x600", "f514eb9f174485f850b7e92d2a40baf6").
					AddRow("/43743431/DMDemo1@160x600", "2fb84286ede5b20e82b0601df0c7e454").
					AddRow("/43743431/DMDemo2@160x600", "2aa34b52a9e941c1594af7565e599c8d")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM  wrapper_publisher_slot (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "valid publisher slotnamehash",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetSlotNameHash: "^SELECT (.+) FROM  wrapper_publisher_slot (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubID: 5890,
			},
			want: map[string]string{
				"/43743431/DMDemo1@160x600": "2fb84286ede5b20e82b0601df0c7e454",
				"/43743431/DMDemo2@160x600": "2aa34b52a9e941c1594af7565e599c8d",
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "hash"}).
					AddRow("/43743431/DMDemo1@160x600", "2fb84286ede5b20e82b0601df0c7e454").
					AddRow("/43743431/DMDemo2@160x600", "2aa34b52a9e941c1594af7565e599c8d")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM  wrapper_publisher_slot (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "error in row scan",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetSlotNameHash: "^SELECT (.+) FROM  wrapper_publisher_slot (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubID: 5890,
			},
			want:    map[string]string(nil),
			wantErr: errors.New("error in row scan"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "hash"}).
					AddRow("/43743431/DMDemo1@160x600", "2fb84286ede5b20e82b0601df0c7e454").
					AddRow("/43743431/DMDemo2@160x600", "2aa34b52a9e941c1594af7565e599c8d")
				rows = rows.RowError(1, errors.New("error in row scan"))
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM  wrapper_publisher_slot (.+)")).WillReturnRows(rows)

				return db
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{
				conn: tt.setup(),
				cfg:  tt.fields.cfg,
			}
			got, err := db.GetPublisherSlotNameHash(tt.args.pubID)
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_mySqlDB_GetWrapperSlotMappings(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	type args struct {
		partnerConfigMap map[int]map[string]string
		profileID        int
		displayVersion   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[int][]models.SlotMapping
		wantErr error
		setup   func() *sql.DB
	}{
		{
			name:    "empty query in config file",
			want:    map[int][]models.SlotMapping{},
			wantErr: errors.New("all expectations were already fulfilled, call to Query '' with args [] was not expected"),
			setup: func() *sql.DB {
				db, _, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				return db
			},
		},
		{
			name: "invalid partnerId in wrapper slot mapping with displayversion 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetWrapperLiveVersionSlotMappings: "^SELECT (.+) FROM wrapper_partner_slot_mapping (.+) LIVE",
					},
				},
			},
			args: args{
				partnerConfigMap: formTestPartnerConfig(),
				profileID:        19109,
				displayVersion:   0,
			},
			want: map[int][]models.SlotMapping{
				10: {
					{
						PartnerId:    10,
						AdapterId:    1,
						VersionId:    1,
						SlotName:     "/43743431/DMDemo2@160x600",
						MappingJson:  "{\"adtag\":\"1405192\",\"site\":\"47124\"}",
						SlotMappings: nil,
						Hash:         "",
						OrderID:      0,
					},
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"PartnerId", "AdapterId", "VersionId", "SlotName", "MappingJson", "OrderId"}).
					AddRow("10_112", 1, 1, "/43743431/DMDemo1@160x600", "{\"adtag\":\"1405192\",\"site\":\"47124\"}", 0).
					AddRow(10, 1, 1, "/43743431/DMDemo2@160x600", "{\"adtag\":\"1405192\",\"site\":\"47124\"}", 0)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_partner_slot_mapping (.+) LIVE")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "valid wrapper slot mapping with displayversion 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetWrapperLiveVersionSlotMappings: "^SELECT (.+) FROM wrapper_partner_slot_mapping (.+) LIVE",
					},
				},
			},
			args: args{
				partnerConfigMap: formTestPartnerConfig(),
				profileID:        19109,
				displayVersion:   0,
			},
			want: map[int][]models.SlotMapping{
				10: {
					{
						PartnerId:    10,
						AdapterId:    1,
						VersionId:    1,
						SlotName:     "/43743431/DMDemo1@160x600",
						MappingJson:  "{\"adtag\":\"1405192\",\"site\":\"47124\"}",
						SlotMappings: nil,
						Hash:         "",
						OrderID:      0,
					},
					{
						PartnerId:    10,
						AdapterId:    1,
						VersionId:    1,
						SlotName:     "/43743431/DMDemo2@160x600",
						MappingJson:  "{\"adtag\":\"1405192\",\"site\":\"47124\"}",
						SlotMappings: nil,
						Hash:         "",
						OrderID:      0,
					},
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"PartnerId", "AdapterId", "VersionId", "SlotName", "MappingJson", "OrderId"}).
					AddRow(10, 1, 1, "/43743431/DMDemo1@160x600", "{\"adtag\":\"1405192\",\"site\":\"47124\"}", 0).
					AddRow(10, 1, 1, "/43743431/DMDemo2@160x600", "{\"adtag\":\"1405192\",\"site\":\"47124\"}", 0)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_partner_slot_mapping (.+) LIVE")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "valid wrapper slot mapping with displayversion non-zero",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetWrapperSlotMappingsQuery: "^SELECT (.+) FROM wrapper_partner_slot_mapping (.+)",
					},
				},
			},
			args: args{
				partnerConfigMap: formTestPartnerConfig(),
				profileID:        19109,
				displayVersion:   4,
			},
			want: map[int][]models.SlotMapping{
				10: {
					{
						PartnerId:    10,
						AdapterId:    1,
						VersionId:    1,
						SlotName:     "/43743431/DMDemo1@160x600",
						MappingJson:  "{\"adtag\":\"1405192\",\"site\":\"47124\"}",
						SlotMappings: nil,
						Hash:         "",
						OrderID:      0,
					},
					{
						PartnerId:    10,
						AdapterId:    1,
						VersionId:    1,
						SlotName:     "/43743431/DMDemo2@160x600",
						MappingJson:  "{\"adtag\":\"1405192\",\"site\":\"47124\"}",
						SlotMappings: nil,
						Hash:         "",
						OrderID:      0,
					},
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"PartnerId", "AdapterId", "VersionId", "SlotName", "MappingJson", "OrderId"}).
					AddRow(10, 1, 1, "/43743431/DMDemo1@160x600", "{\"adtag\":\"1405192\",\"site\":\"47124\"}", 0).
					AddRow(10, 1, 1, "/43743431/DMDemo2@160x600", "{\"adtag\":\"1405192\",\"site\":\"47124\"}", 0)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_partner_slot_mapping (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "error in row scan with displayversion 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetWrapperLiveVersionSlotMappings: "^SELECT (.+) FROM wrapper_partner_slot_mapping (.+) LIVE",
					},
				},
			},
			args: args{
				partnerConfigMap: formTestPartnerConfig(),
				profileID:        19109,
				displayVersion:   0,
			},
			want:    map[int][]models.SlotMapping(nil),
			wantErr: errors.New("error in row scan"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"PartnerId", "AdapterId", "VersionId", "SlotName", "MappingJson", "OrderId"}).
					AddRow(10, 1, 1, "/43743431/DMDemo1@160x600", "{\"adtag\":\"1405192\",\"site\":\"47124\"}", 0).
					AddRow(10, 1, 1, "/43743431/DMDemo2@160x600", "{\"adtag\":\"1405192\",\"site\":\"47124\"}", 0)
				rows = rows.RowError(1, errors.New("error in row scan"))
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_partner_slot_mapping (.+) LIVE")).WillReturnRows(rows)

				return db
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{
				conn: tt.setup(),
				cfg:  tt.fields.cfg,
			}
			got, err := db.GetWrapperSlotMappings(tt.args.partnerConfigMap, tt.args.profileID, tt.args.displayVersion)
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_mySqlDB_GetMappings(t *testing.T) {
	type args struct {
		slotKey string
		slotMap map[string]models.SlotMapping
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr error
	}{
		{
			name:    "empty_data",
			args:    args{},
			want:    nil,
			wantErr: errors.New("No mapping found for slot:"),
		},
		{
			name: "slotmapping_notfound",
			args: args{
				slotKey: "key1",
				slotMap: map[string]models.SlotMapping{
					"slot1": {},
				},
			},
			want:    nil,
			wantErr: errors.New("No mapping found for slot:key1"),
		},
		{
			name: "slotmapping_found_with_empty_fieldmap",
			args: args{
				slotKey: "slot1",
				slotMap: map[string]models.SlotMapping{
					"slot1": {},
				},
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "slotmapping_found_with_fieldmap",
			args: args{
				slotKey: "slot1",
				slotMap: map[string]models.SlotMapping{
					"slot1": {
						SlotMappings: map[string]interface{}{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: nil,
		},
		{
			name: "key_case_sensitive",
			args: args{
				slotKey: "SLOT1",
				slotMap: map[string]models.SlotMapping{
					"slot1": {
						SlotMappings: map[string]interface{}{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{}
			got, err := db.GetMappings(tt.args.slotKey, tt.args.slotMap)
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func formTestPartnerConfig() map[int]map[string]string {

	partnerConfigMap := make(map[int]map[string]string)

	partnerConfigMap[0] = map[string]string{
		"partnerId":         "10",
		"prebidPartnerName": "pubmatic",
		"serverSideEnabled": "1",
		"level":             "multi",
		"kgp":               "_AU_@_W_x_H",
		"timeout":           "220",
	}

	return partnerConfigMap
}
