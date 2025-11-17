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
						GetPerformanceDSPQuery: "SELECT dsp_id, value FROM performance_dsp",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id, value FROM performance_dsp")).
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
						GetPerformanceDSPQuery: "SELECT dsp_id, value FROM performance_dsp",
					},
					MaxDbContextTimeout: 1, // Very short timeout
				},
			},
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id, value FROM performance_dsp")).
					WillReturnError(errors.New("context deadline exceeded"))
				return db
			},
		},
		{
			name: "valid_rows_with_enabled_dsps",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id, value FROM performance_dsp",
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
				rows := sqlmock.NewRows([]string{"dsp_id", "value"}).
					AddRow(101, "1").
					AddRow(102, "1").
					AddRow(103, "0").
					AddRow(104, "0").
					AddRow(105, "1")
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id, value FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "all_dsps_disabled",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id, value FROM performance_dsp",
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
				rows := sqlmock.NewRows([]string{"dsp_id", "value"}).
					AddRow(101, "0").
					AddRow(102, "0").
					AddRow(103, "0")
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id, value FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "scan_error_returns_error",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id, value FROM performance_dsp",
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
				rows := sqlmock.NewRows([]string{"dsp_id", "value"}).
					AddRow(101, "1").
					AddRow(102, "1").
					AddRow(103, "1")
				// Add row error for the second row - this will be caught by rows.Err()
				rows = rows.RowError(1, errors.New("scan error"))
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id, value FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "invalid_enable_value_skips_row",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id, value FROM performance_dsp",
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
				rows := sqlmock.NewRows([]string{"dsp_id", "value"}).
					AddRow(101, "1").
					AddRow(102, "invalid"). // Invalid value - should be skipped
					AddRow(103, "1")
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id, value FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "non_numeric_enable_value_skips_row",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id, value FROM performance_dsp",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want: map[int]struct{}{
				101: {},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"dsp_id", "value"}).
					AddRow(101, "1").
					AddRow(102, "abc"). // Non-numeric value
					AddRow(103, "1.5"). // Float value
					AddRow(104, "true") // Boolean string
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id, value FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "empty_result_set",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id, value FROM performance_dsp",
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
				rows := sqlmock.NewRows([]string{"dsp_id", "value"})
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id, value FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "rows_error_after_iteration",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id, value FROM performance_dsp",
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
				rows := sqlmock.NewRows([]string{"dsp_id", "value"}).
					AddRow(101, "1").
					AddRow(102, "1").
					CloseError(errors.New("rows iteration error"))
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id, value FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "mixed_enabled_disabled_and_invalid_values",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id, value FROM performance_dsp",
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
				rows := sqlmock.NewRows([]string{"dsp_id", "value"}).
					AddRow(101, "1").       // Enabled
					AddRow(102, "0").       // Disabled
					AddRow(103, "invalid"). // Invalid - skipped
					AddRow(104, "2").       // Value 2 (not 1, so not enabled)
					AddRow(105, "1")        // Enabled
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id, value FROM performance_dsp")).
					WillReturnRows(rows)
				return db
			},
		},
		{
			name: "only_value_1_is_enabled",
			fields: fields{
				cfg: config.Database{
					Queries: config.Queries{
						GetPerformanceDSPQuery: "SELECT dsp_id, value FROM performance_dsp",
					},
					MaxDbContextTimeout: 1000,
				},
			},
			want: map[int]struct{}{
				101: {},
			},
			wantErr: nil,
			setup: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				rows := sqlmock.NewRows([]string{"dsp_id", "value"}).
					AddRow(101, "1"). // Only this should be enabled
					AddRow(102, "2"). // Not enabled (value != 1)
					AddRow(103, "3"). // Not enabled (value != 1)
					AddRow(104, "-1") // Not enabled (value != 1)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT dsp_id, value FROM performance_dsp")).
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
