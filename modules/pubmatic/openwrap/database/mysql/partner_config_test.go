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
		pubId          int
		profileId      int
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
						LiveVersionInnerQuery: `SELECT wv.id as versionId, display_version as displayVersionId FROM wrapper_version wv JOIN wrapper_status ws on profile_id=? AND version_id=id and status IN ('LIVE','LIVE_PENDING') JOIN wrapper_profile wp ON profile_id=wp.id AND pub_id=?`,
					},
				},
			},
			args: args{
				pubId:          5890,
				profileId:      19109,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT wv.id as versionId, display_version as displayVersionId FROM wrapper_version wv JOIN wrapper_status ws on profile_id=? AND version_id=id and status IN ('LIVE','LIVE_PENDING') JOIN wrapper_profile wp ON profile_id=wp.id AND pub_id=?`)).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "valid partnerconfig with displayversion is 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: `SELECT wv.id as versionId, display_version as displayVersionId FROM wrapper_version wv JOIN wrapper_status ws on profile_id=? AND version_id=id and status IN ('LIVE','LIVE_PENDING') JOIN wrapper_profile wp ON profile_id=wp.id AND pub_id=?`,
						GetParterConfig:       `SELECT /*+ MAX_EXECUTION_TIME(%d) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = %d AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = %d AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = %d LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubId:          5890,
				profileId:      19109,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT wv.id as versionId, display_version as displayVersionId FROM wrapper_version wv JOIN wrapper_status ws on profile_id=? AND version_id=id and status IN ('LIVE','LIVE_PENDING') JOIN wrapper_profile wp ON profile_id=wp.id AND pub_id=?`)).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				rowsPartnerConfig := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "keyName", "value"}).
					AddRow("-1", "ALL", "ALL", 0, -1, 0, "platform", "display").
					AddRow("-1", "ALL", "ALL", 0, -1, 0, "gdpr", "0").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "kgp", "_AU_@_W_x_H_").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "timeout", "200").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "serverSideEnabled", "1")
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT /*+ MAX_EXECUTION_TIME(1000) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = 251 AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = 251 AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = 251 LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`)).WillReturnRows(rowsPartnerConfig)
				return db
			},
		},
		{
			name: "valid partnerconfig with displayversion is not 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						DisplayVersionInnerQuery: `SELECT wv.id as versionId, display_version AS displayVersionId FROM wrapper_version wv JOIN wrapper_profile wp ON profile_id=wp.id WHERE profile_id=? and display_version=? AND pub_id=?`,
						GetParterConfig:          `SELECT /*+ MAX_EXECUTION_TIME(%d) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = %d AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = %d AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = %d LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubId:          5890,
				profileId:      19109,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT wv.id as versionId, display_version AS displayVersionId FROM wrapper_version wv JOIN wrapper_profile wp ON profile_id=wp.id WHERE profile_id=? and display_version=? AND pub_id=?`)).WithArgs(19109, 3, 5890).WillReturnRows(rowsWrapperVersion)

				rowsPartnerConfig := sqlmock.NewRows([]string{"partnerId", "prebidPartnerName", "bidderCode", "isAlias", "entityTypeID", "testConfig", "keyName", "value"}).
					AddRow("-1", "ALL", "ALL", 0, -1, 0, "platform", "display").
					AddRow("-1", "ALL", "ALL", 0, -1, 0, "gdpr", "0").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "kgp", "_AU_@_W_x_H_").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "timeout", "200").
					AddRow("101", "pubmatic", "pubmatic", 0, 3, 0, "serverSideEnabled", "1")
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT /*+ MAX_EXECUTION_TIME(1000) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = 251 AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = 251 AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = 251 LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`)).WillReturnRows(rowsPartnerConfig)
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
			got, err := db.GetActivePartnerConfigurations(tt.args.pubId, tt.args.profileId, tt.args.displayVersion)
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
		pubId     int
		profileId int
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
						GetParterConfig: `SELECT /*+ MAX_EXECUTION_TIME(%d) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = %d AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = %d AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = %d LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubId:     5890,
				profileId: 1234,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT /*+ MAX_EXECUTION_TIME(1000) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = 1 AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = 1 AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = 1 LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "default display version",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: `SELECT /*+ MAX_EXECUTION_TIME(%d) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = %d AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = %d AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = %d LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubId:     5890,
				profileId: 1234,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT /*+ MAX_EXECUTION_TIME(1000) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = 123 AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = 123 AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = 123 LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "account params present",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: `SELECT /*+ MAX_EXECUTION_TIME(%d) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = %d AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = %d AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = %d LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubId:     5890,
				profileId: 1234,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT /*+ MAX_EXECUTION_TIME(1000) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = 123 AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = 123 AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = 123 LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "AB Test Enabled",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: `SELECT /*+ MAX_EXECUTION_TIME(%d) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = %d AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = %d AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = %d LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubId:     5890,
				profileId: 1234,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT /*+ MAX_EXECUTION_TIME(1000) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = 123 AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = 123 AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = 123 LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "bidder alias present",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: `SELECT /*+ MAX_EXECUTION_TIME(%d) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = %d AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = %d AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = %d LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubId:     5890,
				profileId: 1234,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT /*+ MAX_EXECUTION_TIME(1000) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = 123 AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = 123 AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = 123 LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "partnerName as `-`",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetParterConfig: `SELECT /*+ MAX_EXECUTION_TIME(%d) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = %d AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = %d AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = %d LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				pubId:     5890,
				profileId: 1234,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT /*+ MAX_EXECUTION_TIME(1000) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = 123 AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = 123 AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = 123 LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`)).WillReturnRows(rows)
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
			got, err := db.getActivePartnerConfigurations(tt.args.pubId, tt.args.profileId, tt.args.versionID)
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
		profileID        int
		displayVersionID int
		pubID            int
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
						LiveVersionInnerQuery: `SELECT wv.id as versionId, display_version as displayVersionId FROM wrapper_version wv JOIN wrapper_status ws on profile_id=? AND version_id=id and status IN ('LIVE','LIVE_PENDING') JOIN wrapper_profile wp ON profile_id=wp.id AND pub_id=?`,
					},
				},
			},
			args: args{
				profileID:        19109,
				displayVersionID: 0,
				pubID:            5890,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT wv.id as versionId, display_version as displayVersionId FROM wrapper_version wv JOIN wrapper_status ws on profile_id=? AND version_id=id and status IN ('LIVE','LIVE_PENDING') JOIN wrapper_profile wp ON profile_id=wp.id AND pub_id=?`)).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "displayversion is 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						LiveVersionInnerQuery: `SELECT wv.id as versionId, display_version as displayVersionId FROM wrapper_version wv JOIN wrapper_status ws on profile_id=? AND version_id=id and status IN ('LIVE','LIVE_PENDING') JOIN wrapper_profile wp ON profile_id=wp.id AND pub_id=?`,
						GetParterConfig:       `SELECT /*+ MAX_EXECUTION_TIME(%d) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = %d AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = %d AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = %d LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				profileID:        19109,
				displayVersionID: 0,
				pubID:            5890,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT wv.id as versionId, display_version as displayVersionId FROM wrapper_version wv JOIN wrapper_status ws on profile_id=? AND version_id=id and status IN ('LIVE','LIVE_PENDING') JOIN wrapper_profile wp ON profile_id=wp.id AND pub_id=?`)).WithArgs(19109, 5890).WillReturnRows(rowsWrapperVersion)

				return db
			},
		},
		{
			name: "displayversion is not 0",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						DisplayVersionInnerQuery: `SELECT wv.id as versionId, display_version AS displayVersionId FROM wrapper_version wv JOIN wrapper_profile wp ON profile_id=wp.id WHERE profile_id=? and display_version=? AND pub_id=?`,
						GetParterConfig:          `SELECT /*+ MAX_EXECUTION_TIME(%d) */ IFNULL(wcm.partner_id, -1) as partnerId, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wp.prebid_partner_name END AS prebidPartnerName, CASE WHEN (wcm.partner_id is NULL) THEN "ALL" ELSE wpp.bidder_code END AS bidderCode, CASE WHEN (wcm.partner_id is NULL) THEN 0 ELSE wpp.is_alias END AS isAlias, CASE WHEN (wcm.partner_id is NULL) THEN -1 ELSE entity_type_id END AS entityTypeID, test_config as testConfig, key_name AS keyName, value FROM wrapper_config_map wcm JOIN wrapper_key_master wkm ON wkm.id=wcm.config_id AND entity_type_id = 3 AND entity_id = %d AND wcm.is_active=1 LEFT JOIN wrapper_publisher_partner wpp ON wcm.partner_id=wpp.id LEFT JOIN wrapper_partner wp ON wpp.partner_id = wp.id UNION SELECT distinct(wppvt.partner_id) AS partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 as isAlias, -1 as entityTypeID, 0 as testConfig, 'kgp' AS keyName, key_gen_pattern AS value FROM wrapper_mapping_template JOIN wrapper_profile_partner_version_template wppvt ON version_id = %d AND template_id=id UNION SELECT partner_id as partnerId, "-" AS prebidPartnerName, "-" AS bidderCode, 0 AS isAlias, entity_type_id AS entityTypeID, test_config as testConfig, (SELECT key_name FROM wrapper_key_master WHERE id=config_id) AS keyName, value FROM wrapper_config_map wcm WHERE entity_type_id=1 AND partner_id is NOT NULL AND EXISTS (SELECT partner_id FROM wrapper_config_map WHERE partner_id=wcm.partner_id AND is_active=1 AND entity_type_id = 3 AND entity_id = %d LIMIT 1) ORDER BY partnerId, keyName, entityTypeId`,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			args: args{
				profileID:        19109,
				displayVersionID: 3,
				pubID:            5890,
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
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT wv.id as versionId, display_version AS displayVersionId FROM wrapper_version wv JOIN wrapper_profile wp ON profile_id=wp.id WHERE profile_id=? and display_version=? AND pub_id=?`)).WithArgs(19109, 3, 5890).WillReturnRows(rowsWrapperVersion)

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
			got, got1, err := db.getVersionID(tt.args.profileID, tt.args.displayVersionID, tt.args.pubID)
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
