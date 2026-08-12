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

func Test_mySqlDB_GetInViewEnabledPublishers(t *testing.T) {
	type fields struct {
		cfg config.Database
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[int]struct{}
		wantErr error
		setup   func() *sql.DB
	}{
		{
			name:    "query_context_error",
			want:    nil,
			wantErr: errors.New("query context error"),
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetInViewEnabledPublishersQuery: "SELECT pub_id FROM inview_enabled_publishers",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT pub_id FROM inview_enabled_publishers")).
					WillReturnError(errors.New("query context error"))
				return db
			},
		},
		{
			name:    "context_timeout_error",
			want:    nil,
			wantErr: errors.New("context deadline exceeded"),
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetInViewEnabledPublishersQuery: "SELECT pub_id FROM inview_enabled_publishers",
					},
					MaxDbContextTimeout: 1, // Very short timeout
				},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT pub_id FROM inview_enabled_publishers")).
					WillReturnError(errors.New("context deadline exceeded"))
				return db
			},
		},
		{
			name: "valid_rows_with_multiple_publishers",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetInViewEnabledPublishersQuery: "SELECT pub_id FROM inview_enabled_publishers",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want: map[int]struct{}{
				1001: {},
				1002: {},
				1003: {},
				1004: {},
				1005: {},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id"}).
					AddRow(1001).
					AddRow(1002).
					AddRow(1003).
					AddRow(1004).
					AddRow(1005)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT pub_id FROM inview_enabled_publishers")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "single_publisher",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetInViewEnabledPublishersQuery: "SELECT pub_id FROM inview_enabled_publishers",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want: map[int]struct{}{
				5890: {},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id"}).
					AddRow(5890)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT pub_id FROM inview_enabled_publishers")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "empty_result_set",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetInViewEnabledPublishersQuery: "SELECT pub_id FROM inview_enabled_publishers",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want:    map[int]struct{}{},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id"})
				mock.ExpectQuery(regexp.QuoteMeta("SELECT pub_id FROM inview_enabled_publishers")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "scan_error_returns_error",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetInViewEnabledPublishersQuery: "SELECT pub_id FROM inview_enabled_publishers",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want:    nil,
			wantErr: errors.New("scan error"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id"}).
					AddRow(1001).
					AddRow(1002).
					AddRow(1003)
				// Add row error for the second row - this will be caught by rows.Err()
				rows = rows.RowError(1, errors.New("scan error"))
				mock.ExpectQuery(regexp.QuoteMeta("SELECT pub_id FROM inview_enabled_publishers")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "rows_error_after_iteration",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetInViewEnabledPublishersQuery: "SELECT pub_id FROM inview_enabled_publishers",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want:    nil,
			wantErr: errors.New("rows iteration error"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id"}).
					AddRow(1001).
					AddRow(1002).
					CloseError(errors.New("rows iteration error"))
				mock.ExpectQuery(regexp.QuoteMeta("SELECT pub_id FROM inview_enabled_publishers")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "large_number_of_publishers",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetInViewEnabledPublishersQuery: "SELECT pub_id FROM inview_enabled_publishers",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want: map[int]struct{}{
				1:    {},
				100:  {},
				500:  {},
				1000: {},
				5000: {},
				9999: {},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id"}).
					AddRow(1).
					AddRow(100).
					AddRow(500).
					AddRow(1000).
					AddRow(5000).
					AddRow(9999)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT pub_id FROM inview_enabled_publishers")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "duplicate_publisher_ids_handled",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetInViewEnabledPublishersQuery: "SELECT pub_id FROM inview_enabled_publishers",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want: map[int]struct{}{
				1001: {},
				1002: {},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"pub_id"}).
					AddRow(1001).
					AddRow(1002).
					AddRow(1001). // Duplicate
					AddRow(1002)  // Duplicate
				mock.ExpectQuery(regexp.QuoteMeta("SELECT pub_id FROM inview_enabled_publishers")).
					WillReturnRows(rows)
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
			got, err := db.GetInViewEnabledPublishers()
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
