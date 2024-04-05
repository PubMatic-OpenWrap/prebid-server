package mysql

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
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
		wantErr bool
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
			name: "Invalid feature_id",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPublisherFeatureMapQuery: "^SELECT (.+) FROM test_wrapper_table (.+)",
					},
				},
			},
			want:    map[int]map[int]models.FeatureData{},
			wantErr: false,
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
			wantErr: false,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mySqlDB{
				conn: tt.setup(),
				cfg:  tt.fields.cfg,
			}
			got, err := db.GetPublisherFeatureMap()
			if (err != nil) != tt.wantErr {
				t.Errorf("mySqlDB.GetPublisherFeatureMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
