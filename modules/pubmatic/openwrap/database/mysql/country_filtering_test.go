package mysql

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

type mockDB struct {
	mock sqlmock.Sqlmock
	db   *sql.DB
}

func newMockDB(t *testing.T) *mockDB {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %s", err)
	}
	return &mockDB{
		mock: mock,
		db:   db,
	}
}
func TestGetCountryPartnerFilteringData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := newMockDB(t)
	mockRows := sqlmock.NewRows([]string{"country", "feature_value", "criteria", "threshold"})

	query := "SELECT country, feature_value, criteria, threshold FROM table"
	maxTimeout := 2 * time.Second

	cpf := &CountryPartnerFilterDB{
		db:                  mockDB.db,
		query:               query,
		MaxDbContextTimeout: maxTimeout,
	}

	tests := []struct {
		name           string
		mockSetup      func()
		expectedResult map[string]map[string]struct{}
		expectErr      bool
	}{
		{
			name: "valid_one_record",
			mockSetup: func() {
				mockDB.mock.ExpectQuery(query).WillReturnRows(mockRows)
				mockRows.AddRow("IN", "bidderA", models.PartnerLevelThrottlingCriteria, models.PartnerLevelThrottlingCriteriaValue)
			},
			expectedResult: map[string]map[string]struct{}{
				"IN": {"bidderA": {}},
			},
			expectErr: false,
		},
		{
			name: "scan_error_skips_row",
			mockSetup: func() {
				mockDB.mock.ExpectQuery(query).WillReturnRows(mockRows)
				mockRows.AddRow("IN", "bidderA", models.PartnerLevelThrottlingCriteria, models.PartnerLevelThrottlingCriteriaValue)
			},
			expectedResult: map[string]map[string]struct{}{},
			expectErr:      false,
		},
		{
			name: "criteria_mismatch_skipped",
			mockSetup: func() {
				mockDB.mock.ExpectQuery(query).WillReturnRows(mockRows)
				mockRows.AddRow("US", "bidderB", "wrong_criteria", models.PartnerLevelThrottlingCriteriaValue)
			},
			expectedResult: map[string]map[string]struct{}{},
			expectErr:      false,
		},
		{
			name: "query_error",
			mockSetup: func() {
				mockDB.mock.ExpectQuery(query).WillReturnError(errors.New("db error"))
			},
			expectedResult: nil,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := cpf.getCountryPartnerFilteringData()
			assert.Equal(t, tt.expectErr, err != nil)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
