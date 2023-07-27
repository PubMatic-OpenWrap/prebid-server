package mysql

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/stretchr/testify/assert"
)

func Test_mySqlDB_GetFSCDisabledPublishers(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[int]struct{}
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
			name: "invalid pubid",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAllFscDisabledPublishersQuery: `SELECT pub_id from wrapper_publisher_feature_mapping where feature_id=(SELECT id FROM wrapper_feature WHERE feature_name="fsc") AND is_enabled = 0`,
					},
				},
			},
			want:    map[int]struct{}{},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id"}).AddRow(`5890,5891,5892`)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT pub_id from wrapper_publisher_feature_mapping where feature_id=(SELECT id FROM wrapper_feature WHERE feature_name="fsc") AND is_enabled = 0`)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "Valid rows returned, setting invalid values to 1",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAllFscDisabledPublishersQuery: `SELECT pub_id from wrapper_publisher_feature_mapping where feature_id=(SELECT id FROM wrapper_feature WHERE feature_name="fsc") AND is_enabled = 0`,
					},
				},
			},
			want: map[int]struct{}{
				5890: {},
				5891: {},
				5892: {},
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id"}).
					AddRow(`5890`).
					AddRow(`5891`).
					AddRow(`5892`)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT pub_id from wrapper_publisher_feature_mapping where feature_id=(SELECT id FROM wrapper_feature WHERE feature_name="fsc") AND is_enabled = 0`)).WillReturnRows(rows)
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
			got, err := db.GetFSCDisabledPublishers()
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetFSCDisabledPublishers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_mySqlDB_GetFSCThresholdPerDSP(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[int]int
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
			name: "Invalid dsp_id",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAllDspFscPcntQuery: `SELECT dsp_id,value from wrapper_feature_dsp_mapping WHERE key_id=(SELECT id FROM wrapper_key_master WHERE key_name = "fsc_pcnt")`,
					},
				},
			},
			want:    map[int]int{},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"dsp_id", "value"}).AddRow(`5,23`, `24`)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT dsp_id,value from wrapper_feature_dsp_mapping WHERE key_id=(SELECT id FROM wrapper_key_master WHERE key_name = "fsc_pcnt")`)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "Valid rows returned,avoiding floating pcnt values",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAllDspFscPcntQuery: `SELECT dsp_id,value from wrapper_feature_dsp_mapping WHERE key_id=(SELECT id FROM wrapper_key_master WHERE key_name = "fsc_pcnt")`,
					},
				},
			},
			want: map[int]int{
				5: 24,
				8: 20,
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"dsp_id", "value"}).
					AddRow(`5`, `24`).
					AddRow(`8`, `20`).
					AddRow(`9`, `12.12`)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT dsp_id,value from wrapper_feature_dsp_mapping WHERE key_id=(SELECT id FROM wrapper_key_master WHERE key_name = "fsc_pcnt")`)).WillReturnRows(rows)
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
			got, err := db.GetFSCThresholdPerDSP()
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetFSCThresholdPerDSP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
