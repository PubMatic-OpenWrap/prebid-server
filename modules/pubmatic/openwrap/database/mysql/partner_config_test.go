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
		wantErr error
		setup   func() *sql.DB
	}{
		{
			name: "invalid verison id",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: models.TestQuery,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubID:          5890,
				profileID:      19109,
				displayVersion: 0,
			},

			want:    nil,
			wantErr: errors.New("LiveVersionInnerQuery/DisplayVersionInnerQuery Failure Error: sql: Scan error on column index 0, name \"versionID\": converting driver.Value type string (\"25_1\") to a int: invalid syntax"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY, models.ProfileTypeKey}).AddRow("25_1", "9", models.PLATFORM_DISPLAY, "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "error getting partnercofnig",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: models.TestQuery,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubID:          5890,
				profileID:      19109,
				displayVersion: 0,
			},
			want:    nil,
			wantErr: errors.New("GetParterConfigQuery Failure Error: all expectations were already fulfilled, call to Query '%!(EXTRA int=1000, int=251, int=19109, int=251, int=251)' with args [] was not expected"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY, models.ProfileTypeKey}).AddRow("251", "9", models.PLATFORM_DISPLAY, "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)
				return db
			},
		},
		{
			name: "valid partnerconfig with displayversion is 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: models.TestQuery,
						GetParterConfig:       models.TestQuery,
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
					"type":              "1",
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY, models.ProfileTypeKey}).AddRow("251", "9", models.PLATFORM_DISPLAY, "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)
				rowsPartnerConfig := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "platform", "display").
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "gdpr", "0").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "kgp", "_AU_@_W_x_H_").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "timeout", "200").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "serverSideEnabled", "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WillReturnRows(rowsPartnerConfig)
				return db
			},
		},
		{
			name: "valid partnerconfig with displayversion is not 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						DisplayVersionInnerQuery: models.TestQuery,
						GetParterConfig:          models.TestQuery,
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
					"type":              "1",
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY, models.ProfileTypeKey}).AddRow("251", "9", models.PLATFORM_DISPLAY, "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WithArgs(19109, 3, 5890).WillReturnRows(rowsWrapperVersion)
				rowsPartnerConfig := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "platform", "display").
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "gdpr", "0").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "kgp", "_AU_@_W_x_H_").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "timeout", "200").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "serverSideEnabled", "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WillReturnRows(rowsPartnerConfig)
				return db
			},
		},
		{
			name: "vastbidder present with publisher key vendorId",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						DisplayVersionInnerQuery: models.TestQuery,
						GetParterConfig:          models.TestQuery,
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
					"type":              "1",
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY, models.ProfileTypeKey}).AddRow("251", "9", models.PLATFORM_DISPLAY, "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WithArgs(19109, 3, 5890).WillReturnRows(rowsWrapperVersion)
				rowsPartnerConfig := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "platform", "display").
					AddRow("-1", "ALL", "ALL", 0, -1, 0, -1, "gdpr", "0").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "kgp", "_AU_@_W_x_H_").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "timeout", "200").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, 76, "serverSideEnabled", "1").
					AddRow("234", "vastbidder", "test-vastbidder", 0, 3, 0, -1, "serverSideEnabled", "1").
					AddRow("234", "vastbidder", "test-vastbidder", 0, 3, 0, -1, "vendorId", "546")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WillReturnRows(rowsPartnerConfig)
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
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				assert.ErrorIs(t, tt.wantErr, err)
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
		profileID int
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
						GetParterConfig: models.TestQuery,
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
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "default display version",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: models.TestQuery,
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
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "account params present",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: models.TestQuery,
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
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "AB Test Enabled",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: models.TestQuery,
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
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "bidder alias present",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: models.TestQuery,
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
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "partnerName as `-`",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: models.TestQuery,
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
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "error in row scan",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: models.TestQuery,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				versionID: 123,
			},
			want:    map[int]map[string]string(nil),
			wantErr: true,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "vendorId", "keyName", "value"}).
					AddRow(101, "openx", "openx", 0, -1, 0, 152, "k1", "v1").
					AddRow(101, "openx", "openx", 0, -1, 0, 152, "k2", "v2").
					AddRow(102, "pubmatic", "pubmatic", 0, -1, 0, 76, "k1", "v2")
				rows = rows.RowError(1, errors.New("error in row scan"))
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WillReturnRows(rows)
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
			gotPartnerConfigMap, err := db.getActivePartnerConfigurations(tt.args.profileID, tt.args.versionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.getActivePartnerConfigurations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, gotPartnerConfigMap)
		})
	}
}

func Test_mySqlDB_getVersionIdAndProfileDeatails(t *testing.T) {
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
		expectedProfileType            int
		wantErr                        bool
		setup                          func() *sql.DB
	}{
		{
			name: "invalid verison id",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: models.TestQuery,
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
			expectedProfileType:            0,
			wantErr:                        true,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY, models.ProfileTypeKey}).AddRow("25_1", "9", models.PLATFORM_APP, "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "displayversion is 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: models.TestQuery,
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
			expectedProfileType:            1,
			wantErr:                        false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY, models.ProfileTypeKey}).AddRow("251", "9", models.PLATFORM_APP, "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "displayversion is not 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						DisplayVersionInnerQuery: models.TestQuery,
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
			expectedProfileType:            1,
			wantErr:                        false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY, models.ProfileId}).AddRow("251", "9", models.PLATFORM_APP, "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WithArgs(19109, 3, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "Platform is null",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						DisplayVersionInnerQuery: models.TestQuery,
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
			expectedProfileType:            1,
			wantErr:                        false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY, models.ProfileTypeKey}).AddRow("123", "12", nil, "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WithArgs(19109, 2, 5890).WillReturnRows(rowsWrapperVersion)
				return db
			},
		},
		{
			name: "Platform is empty string",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: models.TestQuery,
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
			expectedPlatform:               "",
			expectedProfileType:            1,
			wantErr:                        false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY, models.ProfileTypeKey}).AddRow("251", "9", "", "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)
				return db
			},
		},
		{
			name: "Platform is not null",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: models.TestQuery,
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
			expectedProfileType:            1,
			wantErr:                        false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rowsWrapperVersion := sqlmock.NewRows([]string{models.VersionID, models.DisplayVersionID, models.PLATFORM_KEY, models.ProfileTypeKey}).AddRow("251", "9", models.PLATFORM_APP, "1")
				mock.ExpectQuery(regexp.QuoteMeta(models.TestQuery)).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)
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
			gotVersionID, gotDisplayVersionID, gotPlatform, gotProfileType, err := db.getVersionIdAndProfileDetails(tt.args.profileID, tt.args.displayVersion, tt.args.pubID)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.getVersionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.expectedVersionID, gotVersionID)
			assert.Equal(t, tt.expectedDisplayVersionIDFromDB, gotDisplayVersionID)
			assert.Equal(t, tt.expectedPlatform, gotPlatform)
			assert.Equal(t, tt.expectedProfileType, gotProfileType)
		})
	}
}
