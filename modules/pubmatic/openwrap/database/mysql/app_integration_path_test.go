package mysql

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/magiconair/properties/assert"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
)

func Test_mySqlDB_GetAppIntegrationPath(t *testing.T) {
	type fields struct {
		conn *sql.DB
		cfg  config.Database
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]int
		wantErr bool
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
			name: "valid rows returned from DB",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetAppIntegrationPathQuery: "^SELECT (.+) FROM app_integration_path (.+)",
					},
				},
			},
			want: map[string]int{
				"test1": 1,
				"test2": 2,
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id"}).
					AddRow(`test1`, `1`).
					AddRow(`test2`, `2`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM app_integration_path (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "valid rows returned from DB",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetAppIntegrationPathQuery: "^SELECT (.+) FROM app_integration_path (.+)",
					},
				},
			},
			want: map[string]int{
				"test1": 1,
				"test2": 2,
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id"}).
					AddRow(`test1`, `1`).
					AddRow(`test2`, `2`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM app_integration_path (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "Invalid id from DB",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetAppIntegrationPathQuery: "^SELECT (.+) FROM app_integration_path (.+)",
					},
				},
			},
			want: map[string]int{
				"test2": 2,
			},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id"}).
					AddRow(`test1`, `1,4`).
					AddRow(`test2`, `2`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM app_integration_path (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "no rows returned from DB",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetAppIntegrationPathQuery: "^SELECT (.+) FROM app_integration_path (.+)",
					},
				},
			},
			want:    map[string]int{},
			wantErr: false,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id"})
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM app_integration_path (.+)")).WillReturnRows(rows)
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
			got, err := db.GetAppIntegrationPath()
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetAppIntegrationPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
