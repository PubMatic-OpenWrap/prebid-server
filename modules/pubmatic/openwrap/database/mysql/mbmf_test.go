package mysql

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestMySqlDBGetProfileAdUnitMultiFloors(t *testing.T) {
	type fields struct {
		cfg config.Database
	}

	tests := []struct {
		name    string
		fields  fields
		setup   func() *sql.DB
		want    models.ProfileAdUnitMultiFloors
		wantErr error
	}{
		{
			name: "Rows error",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 5,
					Queries: config.Queries{
						GetProfileAdUnitMultiFloors: "SELECT",
					},
				},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunit_name", "profile_id", "value"}).
					RowError(0, errors.New("row error")).
					AddRow("adunit1", 123, `{"isActive":true,"tier1":1.0,"tier2":0.8,"tier3":0.6}`)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
				return db
			},
			want:    models.ProfileAdUnitMultiFloors{},
			wantErr: errors.New("row error"),
		},
		{
			name: "Success case with different no. of tiers floors",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 5,
					Queries: config.Queries{
						GetProfileAdUnitMultiFloors: "SELECT",
					},
				},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunit_name", "profile_id", "value"}).
					AddRow("adunit1", 12344, `{"isActive":true,"tier1":1.0,"tier2":0.8,"tier3":0.6}`).
					AddRow("adunit2", 54532, `{"isActive":true,"tier1":2.0,"tier2":1.6,"tier3":1.2,"tier4":2.4}`)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
				return db
			},
			want: models.ProfileAdUnitMultiFloors{
				12344: map[string]*models.MultiFloors{
					"adunit1": {IsActive: true, Tier1: 1.0, Tier2: 0.8, Tier3: 0.6},
				},
				54532: map[string]*models.MultiFloors{
					"adunit2": {IsActive: true, Tier1: 2.0, Tier2: 1.6, Tier3: 1.2, Tier4: 2.4},
				},
			},
			wantErr: nil,
		},
		{
			name: "Query error",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 5,
					Queries: config.Queries{
						GetProfileAdUnitMultiFloors: "SELECT",
					},
				},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectQuery("SELECT").WillReturnError(sql.ErrConnDone)
				return db
			},
			want:    models.ProfileAdUnitMultiFloors{},
			wantErr: sql.ErrConnDone,
		},
		{
			name: "Invalid JSON value",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 5,
					Queries: config.Queries{
						GetProfileAdUnitMultiFloors: "SELECT",
					},
				},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunit_name", "profile_id", "value"}).
					AddRow("adunit1", 1, `invalid json`)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
				return db
			},
			want:    models.ProfileAdUnitMultiFloors{},
			wantErr: nil,
		},
		{
			name: "Row scan error",
			fields: fields{
				cfg: config.Database{
					MaxDbContextTimeout: 5,
					Queries: config.Queries{
						GetProfileAdUnitMultiFloors: "SELECT",
					},
				},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"adunit_name", "profile_id"}).
					AddRow("adunit1", 1)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
				return db
			},
			want:    models.ProfileAdUnitMultiFloors{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup()
			defer db.Close()

			mysqlDB := &mySqlDB{
				conn: db,
				cfg:  tt.fields.cfg,
			}

			got, err := mysqlDB.GetProfileAdUnitMultiFloors()
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
