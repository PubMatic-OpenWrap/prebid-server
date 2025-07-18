package mysql

import (
	"database/sql"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/stretchr/testify/assert"
)

var dbOnceReset = func() { dbOnce = sync.Once{} }

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func() (*sql.DB, sqlmock.Sqlmock, error)
		cacheCfg       config.Cache
		dbCfg          config.Database
		expectNilCPF   bool
		expectErrorLog bool
	}{
		{
			name: "success - NewCountryPartnerFilterDB_returns_valid_object",
			setupDB: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}
				mock.ExpectQuery("SELECT country, value , criteria, criteria_threshold").
					WillReturnRows(sqlmock.NewRows([]string{"country", "value", "criteria", "criteria_threshold"}).
						AddRow("US", "partnerA", "monetized_cpm", 0))
				return db, mock, nil
			},
			cacheCfg: config.Cache{
				CountryPartnerFilterRefreshInterval: 1,
			},
			dbCfg: config.Database{
				Queries: config.Queries{
					GetCountryPartnerFilteringData: "SELECT country, value , criteria, criteria_threshold FROM wrapper_metrics WHERE feature_id=1",
				},
			},
			expectNilCPF: false,
		},
		{
			name: "failure - NewCountryPartnerFilterDB_returns_error",
			setupDB: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, _, err := sqlmock.New()
				return db, nil, err
			},
			cacheCfg: config.Cache{
				CountryPartnerFilterRefreshInterval: 1,
			},
			dbCfg: config.Database{
				Queries: config.Queries{
					GetCountryPartnerFilteringData: "SELECT invalid",
				},
			},
			expectNilCPF: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbOnceReset()

			conn, _, err := tt.setupDB()
			assert.NoError(t, err)

			got := New(conn, tt.dbCfg, tt.cacheCfg)

			assert.NotNil(t, got)
			assert.Equal(t, tt.expectNilCPF, got.countryPartnerFilterDB == nil)
		})
	}
}
