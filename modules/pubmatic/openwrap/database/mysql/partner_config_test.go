package mysql

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId"}).AddRow("25_1", "9")
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

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId"}).AddRow("251", "9")
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
				},
				-1: {
					"bidderCode":        "ALL",
					"prebidPartnerName": "ALL",
					"gdpr":              "0",
					"isAlias":           "0",
					"partnerId":         "-1",
					"displayVersionId":  "9",
					"platform":          "display",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId"}).AddRow("251", "9")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+) LIVE")).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				rowsPartnerConfig := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "keyName", "value"}).
					AddRow("-1", "ALL", "ALL", 0, -1, 0, "platform", "display").
					AddRow("-1", "ALL", "ALL", 0, -1, 0, "gdpr", "0").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "kgp", "_AU_@_W_x_H_").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "timeout", "200").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "serverSideEnabled", "1")
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
				},
				-1: {
					"bidderCode":        "ALL",
					"prebidPartnerName": "ALL",
					"gdpr":              "0",
					"isAlias":           "0",
					"partnerId":         "-1",
					"displayVersionId":  "9",
					"platform":          "display",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId"}).AddRow("251", "9")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+)")).WithArgs(19109, 3, 5890).WillReturnRows(rowsWrapperVersion)

				rowsPartnerConfig := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "keyName", "value"}).
					AddRow("-1", "ALL", "ALL", 0, -1, 0, "platform", "display").
					AddRow("-1", "ALL", "ALL", 0, -1, 0, "gdpr", "0").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "kgp", "_AU_@_W_x_H_").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "timeout", "200").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "serverSideEnabled", "1")
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
			got, err := db.GetActivePartnerConfigurations(tt.args.pubID, tt.args.profileID, tt.args.displayVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetActivePartnerConfigurations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
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
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "keyName", "value"}).
					AddRow("11_11", "openx", "openx", 0, -1, 0, "k1", "v1")
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
				},
				102: {
					"k1":                "v2",
					"partnerId":         "102",
					"prebidPartnerName": "pubmatic",
					"bidderCode":        "pubmatic",
					"isAlias":           "0",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "keyName", "value"}).
					AddRow(101, "openx", "openx", 0, -1, 0, "k1", "v1").
					AddRow(101, "openx", "openx", 0, -1, 0, "k2", "v2").
					AddRow(102, "pubmatic", "pubmatic", 0, -1, 0, "k1", "v2")
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
				},
				102: {
					"k1":                "v1",
					"k2":                "v2",
					"partnerId":         "102",
					"prebidPartnerName": "SecondPartnerName",
					"bidderCode":        "SecondBidder",
					"isAlias":           "0",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "keyName", "value"}).
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, "accountId", "1234").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, "accountId", "9876").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, "pubId", "9999").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, "pubId", "8888").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, "rev_share", "10").
					AddRow(102, "SecondPartnerName", "SecondBidder", 0, -1, 0, "k1", "v1").
					AddRow(102, "SecondPartnerName", "SecondBidder", 0, -1, 0, "k2", "v2")
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
				},
				102: {
					"k1":                "v1",
					"k2":                "v2",
					"partnerId":         "102",
					"prebidPartnerName": "SecondPartnerName",
					"bidderCode":        "SecondBidder",
					"isAlias":           "0",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "keyName", "value"}).
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, "accountId", "1234").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, "sstimeout", "200").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 1, "sstimeout", "350").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, "pubId", "8888").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, "rev_share", "10").
					AddRow(102, "SecondPartnerName", "SecondBidder", 0, -1, 0, "k1", "v1").
					AddRow(102, "SecondPartnerName", "SecondBidder", 0, -1, 0, "k2", "v2")
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
				},
				102: {
					"k1":                "v1",
					"k2":                "v2",
					"partnerId":         "102",
					"prebidPartnerName": "SecondPartnerName",
					"bidderCode":        "SecondBidder",
					"isAlias":           "1",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "keyName", "value"}).
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, "accountId", "1234").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 1, 0, "sstimeout", "200").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, "pubId", "8888").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, "rev_share", "10").
					AddRow(102, "SecondPartnerName", "SecondBidder", 1, -1, 0, "k1", "v1").
					AddRow(102, "SecondPartnerName", "SecondBidder", 1, -1, 0, "k2", "v2")
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
				},
				102: {
					"k1":                "v1",
					"k2":                "v2",
					"partnerId":         "102",
					"prebidPartnerName": "SecondPartnerName",
					"bidderCode":        "SecondBidder",
					"isAlias":           "0",
				},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "keyName", "value"}).
					AddRow(101, "-", "-", 0, 1, 0, "accountId", "1234").
					AddRow(101, "-", "-", 0, 1, 0, "sstimeout", "200").
					AddRow(101, "-", "-", 0, 1, 0, "pubId", "8888").
					AddRow(101, "-", "-", 0, 3, 0, "pubId", "12345").
					AddRow(101, "FirstPartnerName", "FirstBidder", 0, 3, 0, "rev_share", "10").
					AddRow(102, "-", "-", 0, -1, 0, "k1", "v1").
					AddRow(102, "SecondPartnerName", "SecondBidder", 0, -1, 0, "k2", "v2")
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
			got, err := db.getActivePartnerConfigurations(tt.args.versionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.getActivePartnerConfigurations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
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
			wantErr:                        true,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId"}).AddRow("25_1", "9")
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
			wantErr:                        false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId"}).AddRow("251", "9")
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
			wantErr:                        false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rowsWrapperVersion := sqlmock.NewRows([]string{"versionId", "displayVersionId"}).AddRow("251", "9")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_version (.+)")).WithArgs(19109, 3, 5890).WillReturnRows(rowsWrapperVersion)

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
			got, got1, err := db.getVersionID(tt.args.profileID, tt.args.displayVersion, tt.args.pubID)
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.getVersionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expectedVersionID {
				t.Errorf("mySqlDB.getVersionID() got = %v, want %v", got, tt.expectedVersionID)
			}
			if got1 != tt.expectedDisplayVersionIDFromDB {
				t.Errorf("mySqlDB.getVersionID() got1 = %v, want %v", got1, tt.expectedDisplayVersionIDFromDB)
			}
		})
	}
}
