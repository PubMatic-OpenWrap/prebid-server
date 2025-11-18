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

func Test_mySqlDB_GetPerformanceDSPs(t *testing.T) {
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
						GetPerformanceDSPQuery: "SELECT dsp_id FROM performance_dsp",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id FROM performance_dsp")).
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
						GetPerformanceDSPQuery: "SELECT dsp_id FROM performance_dsp",
					},
					MaxDbContextTimeout: 1, // Very short timeout
				},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id FROM performance_dsp")).
					WillReturnError(errors.New("context deadline exceeded"))
				return db
			},
		},
		{
			name: "valid_rows_with_enabled_dsps",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id FROM performance_dsp",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want: map[int]struct{}{
				101: {},
				102: {},
				105: {},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"dsp_id"}).
					AddRow(101).
					AddRow(102).
					AddRow(105)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "scan_error_returns_error",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id FROM performance_dsp",
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
				rows := sqlmock.NewRows([]string{"dsp_id"}).
					AddRow(101).
					AddRow(102).
					AddRow(103)
				// Add row error for the second row - this will be caught by rows.Err()
				rows = rows.RowError(1, errors.New("scan error"))
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "empty_result_set",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id FROM performance_dsp",
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
				rows := sqlmock.NewRows([]string{"dsp_id"})
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "rows_error_after_iteration",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id FROM performance_dsp",
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
				rows := sqlmock.NewRows([]string{"dsp_id"}).
					AddRow(101).
					AddRow(102).
					CloseError(errors.New("rows iteration error"))
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "row_error_stops_iteration",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id FROM performance_dsp",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want:    nil,
			wantErr: errors.New("rows next error"),
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				// RowError causes rows.Err() to return error after iteration
				rows := sqlmock.NewRows([]string{"dsp_id"}).
					AddRow(101).
					RowError(1, errors.New("rows next error")).
					AddRow(102)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "scan_error_skips_invalid_row_continues_processing",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id FROM performance_dsp",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want: map[int]struct{}{
				101: {},
				103: {},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"dsp_id"}).
					AddRow(101).
					AddRow("invalid_dsp_id"). // This will cause scan error, should be skipped
					AddRow(103)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "multiple_scan_errors_continue_processing",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id FROM performance_dsp",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want: map[int]struct{}{
				101: {},
				105: {},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"dsp_id"}).
					AddRow(101).
					AddRow("invalid1"). // Scan error - skipped
					AddRow(nil).        // Scan error - skipped
					AddRow("invalid2"). // Scan error - skipped
					AddRow(105)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "all_rows_have_scan_errors",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id FROM performance_dsp",
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
				rows := sqlmock.NewRows([]string{"dsp_id"}).
					AddRow("invalid1").
					AddRow("invalid2").
					AddRow(nil)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id FROM performance_dsp")).
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
			got, err := db.GetPerformanceDSPs()
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error(), tt.name)
			}
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
