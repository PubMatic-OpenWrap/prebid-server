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

func Test_mySqlDB_GetPublisherFeatureMap(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	tests := []struct {
		name    string
		fields  fields
		setup   func() *sql.DB
		want    map[int]map[int]models.FeatureData
		wantErr error
	}{
		{
			name:    "empty query in config file",
			want:    nil,
			wantErr: errors.New("context deadline exceeded"),
			setup: func() *sql.DB {
				db, _, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				return db
			},
		},
		{
			name: "Invalid feature_id",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPublisherFeatureMapQuery: "^SELECT (.+) FROM test_wrapper_table (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want:    map[int]map[int]models.FeatureData{},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id", "feature_id", "is_enabled", "value"}).AddRow(`5890`, `3.5`, `1`, `24`)
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM test_wrapper_table (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "Valid rows returned",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPublisherFeatureMapQuery: "^SELECT (.+) FROM test_wrapper_table (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want: map[int]map[int]models.FeatureData{
				5890: {
					1: {
						Enabled: 0,
					},
					2: {
						Enabled: 1,
						Value:   `{"1234": 100}`,
					},
					3: {
						Enabled: 1,
					},
				},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id", "feature_id", "is_enabled", "value"}).
					AddRow(`5890`, `1`, `0`, sql.NullString{}).
					AddRow(`5890`, `2`, `1`, `{"1234": 100}`).
					AddRow(`5890`, `3`, `1`, sql.NullString{})
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM test_wrapper_table (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "no rows returned from DB",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPublisherFeatureMapQuery: "^SELECT (.+) FROM test_wrapper_table (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want:    map[int]map[int]models.FeatureData{},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id", "feature_id", "is_enabled", "value"})
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM test_wrapper_table (.+)")).WillReturnRows(rows)
				return db
			},
		},
		{
			name: "error in row scan",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPublisherFeatureMapQuery: "^SELECT (.+) FROM test_wrapper_table (.+)",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want:    map[int]map[int]models.FeatureData(nil),
			wantErr: errors.New("error in row scan"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id", "feature_id", "is_enabled", "value"}).
					AddRow(`5890`, `1`, `0`, sql.NullString{}).
					AddRow(`5890`, `2`, `1`, `{"1234": 100}`).
					AddRow(`5890`, `3`, `1`, sql.NullString{})
				rows = rows.RowError(1, errors.New("error in row scan"))
				mock.ExpectQuery(regexp.QuoteMeta("^SELECT (.+) FROM test_wrapper_table (.+)")).WillReturnRows(rows)
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
			got, err := db.GetPublisherFeatureMap()
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
