package mysql

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetTBFTrafficForPublishers(t *testing.T) {
	type want struct {
		trafficDetails map[int]map[int]int
		err            error
	}

	tests := []struct {
		name  string
		setup func(db *mySqlDB)
		want  want
	}{
		{
			name: "db_query_fail",
			setup: func(db *mySqlDB) {
				conn, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rows := sqlmock.NewRows([]string{"value"})
				rows.AddRow("{'5890': 12}")
				mock.ExpectQuery("").WillReturnError(sql.ErrConnDone)
				db.conn = conn
			},
			want: want{
				trafficDetails: nil,
				err:            sql.ErrConnDone,
			},
		},
		{
			name: "query_returns_empty_rows",
			setup: func(db *mySqlDB) {
				conn, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				rows := sqlmock.NewRows([]string{"value"})
				mock.ExpectQuery("").WillReturnRows(rows)
				db.conn = conn
			},
			want: want{
				trafficDetails: map[int]map[int]int{},
				err:            nil,
			},
		},
		{
			name: "row_scan_failure",
			setup: func(db *mySqlDB) {
				conn, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				row := mock.NewRows([]string{"value"}).AddRow(nil)
				mock.ExpectQuery("").WillReturnRows(row)
				db.conn = conn
			},
			want: want{
				trafficDetails: map[int]map[int]int{},
				err:            nil,
			},
		},
		{
			name: "json_unmarshal_fail",
			setup: func(db *mySqlDB) {
				conn, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				row := mock.NewRows([]string{"pubid", "value"}).AddRow("5890", "{1234:10}")
				mock.ExpectQuery("").WillReturnRows(row)
				db.conn = conn
			},
			want: want{
				trafficDetails: map[int]map[int]int{},
				err:            nil,
			},
		},
		{
			name: "valid_single_row",
			setup: func(db *mySqlDB) {
				conn, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				row := mock.NewRows([]string{"pubid", "value"}).AddRow("5890", "{\"1234\":10}")
				mock.ExpectQuery("").WillReturnRows(row)
				db.conn = conn
			},
			want: want{
				trafficDetails: map[int]map[int]int{
					5890: {1234: 10},
				},
				err: nil,
			},
		},
		{
			name: "multi_row_response",
			setup: func(db *mySqlDB) {
				conn, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				row := mock.NewRows([]string{"pubid", "value"})
				row.AddRow("5890", "{\"1234\":10  ,\"4321\": 90}")
				row.AddRow("5891", "{\"5678\":20}")
				mock.ExpectQuery("").WillReturnRows(row)
				db.conn = conn
			},
			want: want{
				trafficDetails: map[int]map[int]int{
					5890: {1234: 10, 4321: 90},
					5891: {5678: 20},
				},
				err: nil,
			},
		},
		{
			name: "one_invalid_row_in_multi_row_response",
			setup: func(db *mySqlDB) {
				conn, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}

				row := mock.NewRows([]string{"pubid", "value"})
				row.AddRow("5890", "{\"1234\":10}")
				row.AddRow("5890", "invalid_row")
				row.AddRow("5890", "{\"5678\":20}")

				mock.ExpectQuery("").WillReturnRows(row)
				db.conn = conn
			},
			want: want{
				trafficDetails: map[int]map[int]int{
					5890: {5678: 20},
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mySQLDB := mySqlDB{}
			tt.setup(&mySQLDB)

			trafficDetails, err := mySQLDB.GetTBFTrafficForPublishers()
			assert.Equalf(t, tt.want.trafficDetails, trafficDetails, tt.name)
			assert.Equalf(t, tt.want.err, err, tt.name)
		})
	}
}
