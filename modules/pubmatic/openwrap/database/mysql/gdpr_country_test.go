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

func Test_mySqlDB_GetGDPRCountryCodes(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	tests := []struct {
		name    string
		fields  fields
		want    models.HashSet
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
						GetGDPRCountryCodes: "^SELECT (.+) FROM KomliAdServer.geo (.+)",
					},
				},
			},
			want: models.HashSet{
				"DE": {},
				"LV": {},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"countrycode"}).
					AddRow(`DE`).
					AddRow(`LV`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM KomliAdServer.geo (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "no rows returned from DB",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetGDPRCountryCodes: "^SELECT (.+) FROM KomliAdServer.geo (.+)",
					},
				},
			},
			want:    models.HashSet{},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"countrycode"})
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM KomliAdServer.geo (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "partial row scan error",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 1000000,
					Queries: config.Queries{
						GetGDPRCountryCodes: "^SELECT (.+) FROM KomliAdServer.geo (.+)",
					},
				},
			},
			want:    models.HashSet{},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"countrycode", "extra_column"}).
					AddRow(`DE`, `12`)
				rows = rows.RowError(1, errors.New("error in row scan"))
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM KomliAdServer.geo (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "error in row scan",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetGDPRCountryCodes: "^SELECT (.+) FROM KomliAdServer.geo (.+)",
					},
				},
			},
			want:    nil,
			wantErr: errors.New("error in row scan"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"countrycode"}).
					AddRow(`DE`).
					AddRow(`LV`)
				rows = rows.RowError(1, errors.New("error in row scan"))
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM KomliAdServer.geo (.+)")).WillReturnRows(rows)
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
			got, err := db.GetGDPRCountryCodes()
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
