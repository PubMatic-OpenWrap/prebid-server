package mysql

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func Test_mySqlDB_GetAppSubIntegrationPath(t *testing.T) {
	type fields struct {
		conn *sql.DB
		cfg  config.Database
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]models.AppSubIntegrationPath
		wantErr error
		setup   func() *sql.DB
	}{
		{
			name: "empty query in config file",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
				},
			},
			want:    nil,
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
			name: "Invalid id from db",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetAppSubIntegrationPathMapQuery: "^SELECT (.+) FROM app_sub_integration_path (.+)",
					},
				},
			},
			want: map[string]models.AppSubIntegrationPath{
				"test_sub_2": {ID: 2, MediationName: "AppLovin", MediationType: "waterfall"},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id", "mediation_name", "mediation_type"}).
					AddRow(`test_sub_1`, `1,3`, "AdMob", "bidding").
					AddRow(`test_sub_2`, `2`, "AppLovin", "waterfall")
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM app_sub_integration_path (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "null mediation columns",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetAppSubIntegrationPathMapQuery: "^SELECT (.+) FROM app_sub_integration_path (.+)",
					},
				},
			},
			want: map[string]models.AppSubIntegrationPath{
				"DFP": {ID: 1},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id", "mediation_name", "mediation_type"}).
					AddRow("DFP", 1, nil, nil)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM app_sub_integration_path (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "no rows returned from DB",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetAppSubIntegrationPathMapQuery: "^SELECT (.+) FROM app_sub_integration_path (.+)",
					},
				},
			},
			want:    map[string]models.AppSubIntegrationPath{},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id", "mediation_name", "mediation_type"})
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM app_sub_integration_path (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "error in row scan",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetAppSubIntegrationPathMapQuery: "^SELECT (.+) FROM app_sub_integration_path (.+)",
					},
				},
			},
			want:    map[string]models.AppSubIntegrationPath(nil),
			wantErr: errors.New("error in row scan"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id", "mediation_name", "mediation_type"}).
					AddRow(`test_sub_1`, `1`, "AdMob", "bidding").
					AddRow(`test_sub_2`, `2`, "AppLovin", "waterfall")
				rows = rows.RowError(1, errors.New("error in row scan"))
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM app_sub_integration_path (.+)")).WillReturnRows(rows)
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
			got, err := db.GetAppSubIntegrationPaths()
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
