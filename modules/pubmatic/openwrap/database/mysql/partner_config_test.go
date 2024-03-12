package mysql

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/stretchr/testify/assert"
)

func Test_mySqlDB_GetActivePartnerConfigurations(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	type args struct {
		pubID          int
		profileID      int
		displayVersion int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[int]map[string]string
		wantErr bool
		setup   func() *sql.DB
	}{
		{
			name: "invalid verison id",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: "^SELECT (.+) FROM wrapper_version (.+) LIVE",
					},
				},
			},
			args: args{
				pubID:          5890,
				profileID:      19109,
				displayVersion: 0,
			},

			want:    nil,
			wantErr: true,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId", "platform"}).AddRow("25_1", "9", "display")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+) LIVE")).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "error getting partnercofnig",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: "^SELECT (.+) FROM wrapper_version (.+) LIVE",
					},
				},
			},
			args: args{
				pubID:          5890,
				profileID:      19109,
				displayVersion: 0,
			},

			want:    nil,
			wantErr: true,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId", "platform"}).AddRow("251", "9", "display")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+) LIVE")).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "valid partnerconfig with displayversion is 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: "^SELECT (.+) FROM wrapper_version (.+) LIVE",
						GetParterConfig:       "^SELECT (.+) FROM wrapper_config_map (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubID:          5890,
				profileID:      19109,
				displayVersion: 0,
			},

			want: map[int]map[string]string{
				101: {
					"bidderCode":        "pubmatic",
					"prebidPartnerName": "pubmatic",
					"timeout":           "200",
					"kgp":               "_AU_@_W_x_H_",
					"serverSideEnabled": "1",
					"isAlias":           "0",
					"partnerId":         "101",
					"vendorId":          "76",
				},
				-1: {
					"bidderCode":        "ALL",
					"prebidPartnerName": "ALL",
					"gdpr":              "0",
					"isAlias":           "0",
					"partnerId":         "-1",
					"displayVersionId":  "9",
					"platform":          "display",
					"vendorId":          "-1",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId", "platform"}).AddRow("251", "9", "display")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+) LIVE")).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				rowsPartnerConfig := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "platform", "display").
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "gdpr", "0").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "kgp", "_AU_@_W_x_H_").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "timeout", "200").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "serverSideEnabled", "1")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_config_map (.+)")).WillReturnRows(rowsPartnerConfig)
				return db
			},
		},
		{
			name: "valid partnerconfig with displayversion is not 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						DisplayVersionInnerQuery: "^SELECT (.+) FROM wrapper_version (.+)",
						GetParterConfig:          "^SELECT (.+) FROM wrapper_config_map (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubID:          5890,
				profileID:      19109,
				displayVersion: 3,
			},

			want: map[int]map[string]string{
				101: {
					"bidderCode":        "pubmatic",
					"prebidPartnerName": "pubmatic",
					"timeout":           "200",
					"kgp":               "_AU_@_W_x_H_",
					"serverSideEnabled": "1",
					"isAlias":           "0",
					"partnerId":         "101",
					"vendorId":          "76",
				},
				-1: {
					"bidderCode":        "ALL",
					"prebidPartnerName": "ALL",
					"gdpr":              "0",
					"isAlias":           "0",
					"partnerId":         "-1",
					"displayVersionId":  "9",
					"platform":          "display",
					"vendorId":          "-1",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId", "platform"}).AddRow("251", "9", "display")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+)")).WithArgs(19109, 3, 5890).WillReturnRows(rowsWrapperVersion)

				rowsPartnerConfig := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "platform", "display").
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "gdpr", "0").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "kgp", "_AU_@_W_x_H_").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "timeout", "200").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "serverSideEnabled", "1")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_config_map (.+)")).WillReturnRows(rowsPartnerConfig)
				return db
			},
		},
		{
			name: "vastbidder present with publisher key vendorId",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						DisplayVersionInnerQuery: "^SELECT (.+) FROM wrapper_version (.+)",
						GetParterConfig:          "^SELECT (.+) FROM wrapper_config_map (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubID:          5890,
				profileID:      19109,
				displayVersion: 3,
			},

			want: map[int]map[string]string{
				234: {
					"bidderCode":        "test-vastbidder",
					"prebidPartnerName": "vastbidder",
					"serverSideEnabled": "1",
					"isAlias":           "0",
					"partnerId":         "234",
					"vendorId":          "546",
				},
				101: {
					"bidderCode":        "pubmatic",
					"prebidPartnerName": "pubmatic",
					"timeout":           "200",
					"kgp":               "_AU_@_W_x_H_",
					"serverSideEnabled": "1",
					"isAlias":           "0",
					"partnerId":         "101",
					"vendorId":          "76",
				},
				-1: {
					"bidderCode":        "ALL",
					"prebidPartnerName": "ALL",
					"gdpr":              "0",
					"isAlias":           "0",
					"partnerId":         "-1",
					"displayVersionId":  "9",
					"platform":          "display",
					"vendorId":          "-1",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId", "platform"}).AddRow("251", "9", "display")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+)")).WithArgs(19109, 3, 5890).WillReturnRows(rowsWrapperVersion)

				rowsPartnerConfig := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "platform", "display").
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "gdpr", "0").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "kgp", "_AU_@_W_x_H_").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "timeout", "200").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "serverSideEnabled", "1").
					AddRow("234", "vastbidder", "test-vastbidder", 0, 3, 0, 546, "vendorId", "999").
					AddRow("234", "vastbidder", "test-vastbidder", 0, 3, 0, 546, "serverSideEnabled", "1")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_config_map (.+)")).WillReturnRows(rowsPartnerConfig)
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
			gotPartnerConfigMap, err := db.GetActivePartnerConfigurations(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetActivePartnerConfigurations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, gotPartnerConfigMap)
		})
	}
}

func Test_mySqlDB_getActivePartnerConfigurations(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	type args struct {
		versionID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[int]map[string]string
		wantErr bool
		setup   func() *sql.DB
	}{
		{
			name:    "empty query in config file",
			want:    nil,
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
			name: "incorrect datatype of partner_id ",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: "^SELECT (.+) FROM wrapper_config_map (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				versionID: 1,
			},
			want:    map[int]map[string]string{},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow("11_11", "openx", "openx", 0, -1, 0, -1, "k1", "v1")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_config_map (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "default display version",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: "^SELECT (.+) FROM wrapper_config_map (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				versionID: 123,
			},
			want: map[int]map[string]string{
				101: {
					"k1":                "v1",
					"k2":                "v2",
					"partnerId":         "101",
					"prebidPartnerName": "openx",
					"bidderCode":        "openx",
					"isAlias":           "0",
					"vendorId":          "152",
				},
				102: {
					"k1":                "v2",
					"partnerId":         "102",
					"prebidPartnerName": "pubmatic",
					"bidderCode":        "pubmatic",
					"isAlias":           "0",
					"vendorId":          "76",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow(101, "openx", "openx", 0, -1, 0, 152, "k1", "v1").
					AddRow(101, "openx", "openx", 0, -1, 0, 152, "k2", "v2").
					AddRow(102, "pubmatic", "pubmatic", 0, -1, 0, 76, "k1", "v2")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_config_map (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "account params present",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: "^SELECT (.+) FROM wrapper_config_map (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				versionID: 123,
			},
			want: map[int]map[string]string{
				101: {
					"accountId":         "9876",
					"pubId":             "8888",
					"rev_share":         "10",
					"partnerId":         "101",
					"prebidPartnerName": "FirstPartnerName",
					"bidderCode":        "FirstBidder",
					"isAlias":           "0",
					"vendorId":          "152",
				},
				102: {
					"k1":                "v1",
					"k2":                "v2",
					"partnerId":         "102",
					"prebidPartnerName": "SecondPartnerName",
					"bidderCode":        "SecondBidder",
					"isAlias":           "0",
					"vendorId":          "100",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, 152, "accountId", "1234").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, 152, "accountId", "9876").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, 152, "pubId", "9999").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, 152, "pubId", "8888").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, 152, "rev_share", "10").
					AddRow(102, "SecondPartnerName", "SecondBidder", 0, -1, 0, 100, "k1", "v1").
					AddRow(102, "SecondPartnerName", "SecondBidder", 0, -1, 0, 100, "k2", "v2")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_config_map (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "AB Test Enabled",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: "^SELECT (.+) FROM wrapper_config_map (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				versionID: 123,
			},
			want: map[int]map[string]string{
				101: {
					"accountId":         "1234",
					"pubId":             "8888",
					"rev_share":         "10",
					"partnerId":         "101",
					"prebidPartnerName": "FirstPartnerName",
					"bidderCode":        "FirstBidder",
					"sstimeout":         "200",
					"sstimeout_test":    "350",
					"testEnabled":       "1",
					"isAlias":           "0",
					"vendorId":          "76",
				},
				102: {
					"k1":                "v1",
					"k2":                "v2",
					"partnerId":         "102",
					"prebidPartnerName": "SecondPartnerName",
					"bidderCode":        "SecondBidder",
					"isAlias":           "0",
					"vendorId":          "100",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, 76, "accountId", "1234").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, 76, "sstimeout", "200").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 1, 76, "sstimeout", "350").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, 76, "pubId", "8888").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, 76, "rev_share", "10").
					AddRow(102, "SecondPartnerName", "SecondBidder", 0, -1, 0, 100, "k1", "v1").
					AddRow(102, "SecondPartnerName", "SecondBidder", 0, -1, 0, 100, "k2", "v2")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_config_map (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "bidder alias present",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: "^SELECT (.+) FROM wrapper_config_map (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				versionID: 123,
			},

			want: map[int]map[string]string{
				101: {
					"accountId":         "1234",
					"pubId":             "8888",
					"rev_share":         "10",
					"partnerId":         "101",
					"prebidPartnerName": "FirstPartnerName",
					"bidderCode":        "FirstBidder",
					"sstimeout":         "200",
					"isAlias":           "0",
					"vendorId":          "76",
				},
				102: {
					"k1":                "v1",
					"k2":                "v2",
					"partnerId":         "102",
					"prebidPartnerName": "SecondPartnerName",
					"bidderCode":        "SecondBidder",
					"isAlias":           "1",
					"vendorId":          "100",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, 76, "accountId", "1234").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, 76, "sstimeout", "200").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, 76, "pubId", "8888").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, 76, "rev_share", "10").
					AddRow(102, "SecondPartnerName", "SecondBidder", 1, -1, 0, 100, "k1", "v1").
					AddRow(102, "SecondPartnerName", "SecondBidder", 1, -1, 0, 100, "k2", "v2")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_config_map (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "partnerName as `-`",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: "^SELECT (.+) FROM wrapper_config_map (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				versionID: 123,
			},
			want: map[int]map[string]string{
				101: {
					"accountId":         "1234",
					"pubId":             "12345",
					"rev_share":         "10",
					"partnerId":         "101",
					"prebidPartnerName": "FirstPartnerName",
					"bidderCode":        "FirstBidder",
					"sstimeout":         "200",
					"isAlias":           "0",
					"vendorId":          "76",
				},
				102: {
					"k1":                "v1",
					"k2":                "v2",
					"partnerId":         "102",
					"prebidPartnerName": "SecondPartnerName",
					"bidderCode":        "SecondBidder",
					"isAlias":           "0",
					"vendorId":          "100",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow(101, "-", "-", 0, 1, 0, 76, "accountId", "1234").
					AddRow(101, "-", "-", 0, 1, 0, 76, "sstimeout", "200").
					AddRow(101, "-", "-", 0, 1, 0, 76, "pubId", "8888").
					AddRow(101, "-", "-", 0, 3, 0, 76, "pubId", "12345").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, 76, "rev_share", "10").
					AddRow(102, "-", "-", 0, -1, 0, 100, "k1", "v1").
					AddRow(102, "SecondPartnerName", "SecondBidder", 0, -1, 0, 100, "k2", "v2")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_config_map (.+)")).WillReturnRows(rows)
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
			gotPartnerConfigMap, err := db.getActivePartnerConfigurations(tt.args.versionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.getActivePartnerConfigurations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, gotPartnerConfigMap)
		})
	}
}

func Test_mySqlDB_getVersionID(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	type args struct {
		profileID      int
		displayVersion int
		pubID          int
	}
	tests := []struct {
		name                           string
		fields                         fields
		args                           args
		expectedVersionID              int
		expectedDisplayVersionIDFromDB int
		expectedPlatform               string
		wantErr                        bool
		setup                          func() *sql.DB
	}{
		{
			name: "invalid verison id",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: "^SELECT (.+) FROM wrapper_version (.+) LIVE",
					},
				},
			},
			args: args{
				profileID:      19109,
				displayVersion: 0,
				pubID:          5890,
			},
			expectedVersionID:              0,
			expectedDisplayVersionIDFromDB: 0,
			expectedPlatform:               "",
			wantErr:                        true,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY}).AddRow("25_1", "9", models.PLATFORM_APP)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+) LIVE")).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "displayversion is 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: "^SELECT (.+) FROM wrapper_version (.+) LIVE",
					},
				},
			},
			args: args{
				profileID:      19109,
				displayVersion: 0,
				pubID:          5890,
			},

			expectedVersionID:              251,
			expectedDisplayVersionIDFromDB: 9,
			expectedPlatform:               "in-app",
			wantErr:                        false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY}).AddRow("251", "9", models.PLATFORM_APP)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+) LIVE")).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "displayversion is not 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						DisplayVersionInnerQuery: "^SELECT (.+) FROM wrapper_version (.+)",
					},
				},
			},
			args: args{
				profileID:      19109,
				displayVersion: 3,
				pubID:          5890,
			},

			expectedVersionID:              251,
			expectedDisplayVersionIDFromDB: 9,
			expectedPlatform:               "in-app",
			wantErr:                        false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY}).AddRow("251", "9", models.PLATFORM_APP)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+)")).WithArgs(19109, 3, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "Platform is null",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						DisplayVersionInnerQuery: "^SELECT (.+) FROM wrapper_version (.+)",
					},
				},
			},
			args: args{
				profileID:      19109,
				displayVersion: 2,
				pubID:          5890,
			},
			expectedVersionID:              123,
			expectedDisplayVersionIDFromDB: 12,
			expectedPlatform:               "",
			wantErr:                        false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY}).AddRow("123", "12", nil)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+)")).WithArgs(19109, 2, 5890).WillReturnRows(rowsWrapperVersion)
				return db
			},
		},
		{
			name: "Platform is not null",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: "^SELECT (.+) FROM wrapper_version (.+) LIVE",
					},
				},
			},
			args: args{
				profileID:      19109,
				displayVersion: 0,
				pubID:          5890,
			},
			expectedVersionID:              251,
			expectedDisplayVersionIDFromDB: 9,
			expectedPlatform:               "in-app",
			wantErr:                        false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY}).AddRow("251", "9", models.PLATFORM_APP)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+) LIVE")).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)
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
			gotVersionID, gotDisplayVersionID, gotPlatform, err := db.getVersionID(tt.args.profileID, tt.args.displayVersion, tt.args.pubID)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.getVersionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.expectedVersionID, gotVersionID, "mySqlDB.getVersionID() gotVersionID and expectedVersionID mismatch")
			assert.Equal(t, tt.expectedDisplayVersionIDFromDB, gotDisplayVersionID, "mySqlDB.getVersionID() gotDisplayVersionID and expectedDisplayVersionIDFromDB mismatch")
			assert.Equal(t, tt.expectedPlatform, gotPlatform, "mySqlDB.getVersionID() gotPlatform and expectedPlatform mismatch")
		})
	}
}
