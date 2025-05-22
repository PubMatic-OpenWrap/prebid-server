package feature

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/stretchr/testify/assert"
)

// Mocks and helpers

type mockDB struct {
	sqlmock.Sqlmock
	*sql.DB
}

func newMockDB(t *testing.T) *mockDB {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %s", err)
	}
	return &mockDB{mock, db}
}

func TestLoadGoogleSDKFeatures(t *testing.T) {
	type testCase struct {
		name           string
		setupMock      func(mock *mockDB)
		expectedResult []Feature
		expectNil      bool
	}

	tests := []testCase{
		{
			name: "Success",
			setupMock: func(mock *mockDB) {
				rows := sqlmock.NewRows([]string{"slot_size"}).
					AddRow("300x250").
					AddRow("728x90")
				mock.ExpectQuery("SELECT .* FROM banner_sizes").WillReturnRows(rows)
			},
			expectedResult: []Feature{{
				Name: FeatureFlexSlot,
				Data: []string{"300x250", "728x90"},
			}},
			expectNil: false,
		},
		{
			name: "QueryError",
			setupMock: func(mock *mockDB) {
				mock.ExpectQuery("SELECT slot_size FROM banner_sizes").WillReturnError(errors.New("db error"))
			},
			expectNil: true,
		},
		{
			name: "ScanError",
			setupMock: func(mock *mockDB) {
				rows := sqlmock.NewRows([]string{"slot_size"}).
					AddRow(nil) // will cause Scan error for string
				mock.ExpectQuery("SELECT slot_size FROM banner_sizes").WillReturnRows(rows)
			},
			expectNil: true,
		},
		{
			name: "RowsError",
			setupMock: func(mock *mockDB) {
				rows := sqlmock.NewRows([]string{"slot_size"}).
					AddRow("300x250")
				rows.RowError(0, errors.New("row error"))
				mock.ExpectQuery("SELECT slot_size FROM banner_sizes").WillReturnRows(rows)
			},
			expectNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := newMockDB(t)
			defer mock.Close()
			tc.setupMock(mock)

			loader := &FeatureLoader{
				db: mock.DB,
				cfg: config.Database{
					MaxDbContextTimeout: 100,
					Queries: config.Queries{
						GetBannerSizesQuery: "SELECT slot_size FROM banner_sizes",
					},
				},
			}

			features := loader.LoadGoogleSDKFeatures()
			if tc.expectNil {
				assert.Nil(t, features)
			} else {
				assert.Len(t, features, len(tc.expectedResult))
				assert.Equal(t, tc.expectedResult[0].Name, features[0].Name)
				assert.ElementsMatch(t, tc.expectedResult[0].Data, features[0].Data)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
