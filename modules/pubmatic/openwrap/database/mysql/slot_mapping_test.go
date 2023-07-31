package mysql

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func Test_mySqlDB_GetPubmaticSlotMappings(t *testing.T) {
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
		want    map[string]models.SlotMapping
		wantErr bool
		setup   func() *sql.DB
	}{
		{
			name:    "empty query in config file",
			want:    map[string]models.SlotMapping{},
			wantErr: true,
			setup: func() *sql.DB {
				db, _, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				return db
			},
		},
		{
			name: "empty site_id in pubmatic slotmappings",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPMSlotToMappings: "^SELECT (.+) FROM giym.publisher_slot_to_tag_mapping (.+)",
					},
				},
			},
			args: args{
				pubID: 5890,
			},
			want: map[string]models.SlotMapping{
				"adunit": {
					PartnerId:    1,
					AdapterId:    1,
					VersionId:    0,
					SlotName:     "adunit",
					MappingJson:  "{\"adtag\":\"0\",\"site\":\"0\",\"floor\":\"0.00\",\"gaid\":\"0\"}",
					SlotMappings: map[string]interface{}{"adtag": "0", "floor": "0.00", "gaid": "0", "owSlotName": "adunit", "site": "0"}, Hash: "", OrderID: 0,
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"slot_name", "pm_size", "pm_site_id", "ad_tag_id", "ga_id", "floor"}).
					AddRow("adunit", "300x250", "", 234, 3, 0.12)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM giym.publisher_slot_to_tag_mapping (.+)")).WithArgs(5890, models.MAX_SLOT_COUNT).WillReturnRows(rows)

				return db
			},
		},
		{
			name: "duplicate slotname in pubmatic slotmappings",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPMSlotToMappings: "^SELECT (.+) FROM giym.publisher_slot_to_tag_mapping (.+)",
					},
				},
			},
			args: args{
				pubID: 5890,
			},
			want: map[string]models.SlotMapping{
				"adunit": {
					PartnerId:    1,
					AdapterId:    1,
					VersionId:    0,
					SlotName:     "adunit",
					MappingJson:  "{\"adtag\":\"111\",\"site\":\"555\",\"floor\":\"0.51\",\"gaid\":\"5\"}",
					SlotMappings: map[string]interface{}{"adtag": "111", "floor": "0.51", "gaid": "5", "owSlotName": "adunit", "site": "555"}, Hash: "", OrderID: 0,
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"slot_name", "pm_size", "pm_site_id", "ad_tag_id", "ga_id", "floor"}).
					AddRow("adunit", "300x250", 123, 234, 3, 0.12).
					AddRow("adunit", "400x350", 555, 111, 5, 0.51)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM giym.publisher_slot_to_tag_mapping (.+)")).WithArgs(5890, models.MAX_SLOT_COUNT).WillReturnRows(rows)

				return db
			},
		},
		{
			name: "valid pubmatic slotmappings",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPMSlotToMappings: "^SELECT (.+) FROM giym.publisher_slot_to_tag_mapping (.+)",
					},
				},
			},
			args: args{
				pubID: 5890,
			},
			want: map[string]models.SlotMapping{
				"adunit": {
					PartnerId:    1,
					AdapterId:    1,
					VersionId:    0,
					SlotName:     "adunit",
					MappingJson:  "{\"adtag\":\"234\",\"site\":\"123\",\"floor\":\"0.12\",\"gaid\":\"3\"}",
					SlotMappings: map[string]interface{}{"adtag": "234", "floor": "0.12", "gaid": "3", "owSlotName": "adunit", "site": "123"}, Hash: "", OrderID: 0,
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"slot_name", "pm_size", "pm_site_id", "ad_tag_id", "ga_id", "floor"}).
					AddRow("adunit", "300x250", 123, 234, 3, 0.12)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM giym.publisher_slot_to_tag_mapping (.+)")).WithArgs(5890, models.MAX_SLOT_COUNT).WillReturnRows(rows)

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
			got, err := db.GetPubmaticSlotMappings(tt.args.pubID)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetPubmaticSlotMappings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

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
		wantErr bool
		setup   func() *sql.DB
	}{
		{
			name:    "empty query in config file",
			want:    map[string]string{},
			wantErr: true,
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
				},
			},
			args: args{
				pubID: 5890,
			},
			want: map[string]string{
				"/43743431/DMDemo1@160x600": "2fb84286ede5b20e82b0601df0c7e454",
				"/43743431/DMDemo2@160x600": "2aa34b52a9e941c1594af7565e599c8d",
			},
			wantErr: false,
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
				},
			},
			args: args{
				pubID: 5890,
			},
			want: map[string]string{
				"/43743431/DMDemo1@160x600": "2fb84286ede5b20e82b0601df0c7e454",
				"/43743431/DMDemo2@160x600": "2aa34b52a9e941c1594af7565e599c8d",
			},
			wantErr: false,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{
				conn: tt.setup(),
				cfg:  tt.fields.cfg,
			}
			got, err := db.GetPublisherSlotNameHash(tt.args.pubID)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetPublisherSlotNameHash() error = %v, wantErr %v", err, tt.wantErr)
				return
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
		wantErr bool
		setup   func() *sql.DB
	}{
		{
			name:    "empty query in config file",
			want:    map[int][]models.SlotMapping{},
			wantErr: true,
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{
				conn: tt.setup(),
				cfg:  tt.fields.cfg,
			}
			got, err := db.GetWrapperSlotMappings(tt.args.partnerConfigMap, tt.args.profileID, tt.args.displayVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetWrapperSlotMappings() error = %v, wantErr %v", err, tt.wantErr)
				return
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
		wantErr bool
	}{
		{
			name:    "empty_data",
			args:    args{},
			want:    nil,
			wantErr: true,
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
			wantErr: true,
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{}
			got, err := db.GetMappings(tt.args.slotKey, tt.args.slotMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetMappings() error = %v, wantErr %v", err, tt.wantErr)
				return
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
