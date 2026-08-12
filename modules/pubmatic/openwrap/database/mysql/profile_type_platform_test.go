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

func Test_mySqlDB_GetProfileTypePlatform(t *testing.T) {
	type fields struct {
		conn *sql.DB
		cfg  config.Database
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]int
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
			name: "valid rows returned from DB",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetProfileTypePlatformMapQuery: "^SELECT (.+) FROM profile_type_platform (.+)",
					},
				},
			},
			want: map[string]int{
				"test1": 1,
				"test2": 2,
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id"}).
					AddRow(`test1`, `1`).
					AddRow(`test2`, `2`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM profile_type_platform (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "Invalid id returned from DB",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetProfileTypePlatformMapQuery: "^SELECT (.+) FROM profile_type_platform (.+)",
					},
				},
			},
			want: map[string]int{
				"test2": 2,
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id"}).
					AddRow(`test1`, `1,5`).
					AddRow(`test2`, `2`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM profile_type_platform (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "no rows returned from DB",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetProfileTypePlatformMapQuery: "^SELECT (.+) FROM profile_type_platform (.+)",
					},
				},
			},
			want:    map[string]int{},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id"})
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM profile_type_platform (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "error in row scan",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetProfileTypePlatformMapQuery: "^SELECT (.+) FROM profile_type_platform (.+)",
					},
				},
			},
			want:    map[string]int(nil),
			wantErr: errors.New("error in row scan"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"name", "id"}).
					AddRow(`test1`, `1`).
					AddRow(`test2`, `2`)
				rows = rows.RowError(1, errors.New("error in row scan"))
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM profile_type_platform (.+)")).WillReturnRows(rows)
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
			got, err := db.GetProfileTypePlatforms()
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
