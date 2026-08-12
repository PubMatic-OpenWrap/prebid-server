package mysql

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/stretchr/testify/assert"
)

const testFscActQuery = "SELECT m.dsp_id, m.value, k.key_name FROM wrapper_feature_dsp_mapping m JOIN wrapper_key_master k ON m.key_id = k.id WHERE k.key_name IN ('fsc_pcnt', 'act_pcnt')"

func Test_mySqlDB_GetFSCAndACTThresholdsPerDSP(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	tests := []struct {
		name    string
		fields  fields
		wantFsc map[int]int
		wantAct map[int]int
		wantErr error
		setup   func() *sql.DB
	}{
		{
			name: "empty query in config file",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAllDspFscAndActPcntQuery: "",
					},
				},
			},
			wantFsc: map[int]int{},
			wantAct: map[int]int{},
			wantErr: nil,
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
						GetAllDspFscAndActPcntQuery: testFscActQuery,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			wantFsc: map[int]int{},
			wantAct: map[int]int{},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"dsp_id", "value", "key_name"}).
					AddRow("5,23", "24", "fsc_pcnt")
				mock.ExpectQuery(regexp.QuoteMeta(testFscActQuery)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "Valid rows returned, avoiding floating pcnt values",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAllDspFscAndActPcntQuery: testFscActQuery,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			wantFsc: map[int]int{5: 24, 8: 20},
			wantAct: map[int]int{5: 80, 8: 60},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"dsp_id", "value", "key_name"}).
					AddRow(5, "24", "fsc_pcnt").
					AddRow(5, "80", "act_pcnt").
					AddRow(8, "20", "fsc_pcnt").
					AddRow(8, "60", "act_pcnt").
					AddRow(9, "12.12", "fsc_pcnt") // invalid for int, skipped
				mock.ExpectQuery(regexp.QuoteMeta(testFscActQuery)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "error in row scan",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAllDspFscAndActPcntQuery: testFscActQuery,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			wantFsc: nil,
			wantAct: nil,
			wantErr: errors.New("error in row scan"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"dsp_id", "value", "key_name"}).
					AddRow(5, "24", "fsc_pcnt").
					AddRow(8, "20", "fsc_pcnt").
					AddRow(8, "60", "act_pcnt")
				rows = rows.RowError(1, errors.New("error in row scan"))
				mock.ExpectQuery(regexp.QuoteMeta(testFscActQuery)).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "query context error",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAllDspFscAndActPcntQuery: testFscActQuery,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			wantFsc: nil,
			wantAct: nil,
			wantErr: errors.New("query context error"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectQuery(regexp.QuoteMeta(testFscActQuery)).
					WillReturnError(errors.New("query context error"))
				return db
			},
		},
		{
			name: "empty result set",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAllDspFscAndActPcntQuery: testFscActQuery,
					},
					MaxDbContextTimeout: 1000,
				},
			},
			wantFsc: map[int]int{},
			wantAct: map[int]int{},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"dsp_id", "value", "key_name"})
				mock.ExpectQuery(regexp.QuoteMeta(testFscActQuery)).WillReturnRows(rows)
				return db
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := tt.setup()
			defer conn.Close()
			db := &mySqlDB{
				conn: conn,
				cfg:  tt.fields.cfg,
			}
			gotFsc, gotAct, err := db.GetFSCAndACTThresholdsPerDSP()
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.wantFsc, gotFsc, tt.name)
			assert.Equal(t, tt.wantAct, gotAct, tt.name)
		})
	}
}
