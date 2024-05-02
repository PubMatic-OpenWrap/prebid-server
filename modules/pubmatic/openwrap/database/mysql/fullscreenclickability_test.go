package mysql

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/stretchr/testify/assert"
)

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
						GetAllDspFscPcntQuery: "^SELECT (.+) FROM wrapper_feature_dsp_mapping (.+)",
					},
					MaxDbContextTimeout: 1000,
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
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_feature_dsp_mapping (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "Valid rows returned,avoiding floating pcnt values",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetAllDspFscPcntQuery: "^SELECT (.+) FROM wrapper_feature_dsp_mapping (.+)",
					},
					MaxDbContextTimeout: 1000,
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
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM wrapper_feature_dsp_mapping (.+)")).WillReturnRows(rows)
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
