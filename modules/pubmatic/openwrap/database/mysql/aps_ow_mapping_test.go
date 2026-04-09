package mysql

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApsOwMappingDB_getApsOwMappingData(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT slot, adunit, prof").
		WillReturnRows(sqlmock.NewRows([]string{"slot", "adunit", "prof"}).
			AddRow("uuid-1", "ad-1", 10).
			AddRow("uuid-2", "ad-2", 20))

	a := &ApsOwMappingDB{
		db:                  db,
		query:               "SELECT slot, adunit, prof FROM aps_ow_mapping",
		MaxDbContextTimeout: 500 * time.Millisecond,
	}

	got, err := a.getApsOwMappingData()
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "ad-1", got["uuid-1"].AdUnitID)
	assert.Equal(t, 10, got["uuid-1"].ProfileID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApsOwMappingDB_Lookup(t *testing.T) {
	var a ApsOwMappingDB
	a.cache.Store(map[string]ApsOwMappingEntry{
		"slot": {AdUnitID: "au", ProfileID: 99},
	})
	ad, pid, ok := a.Lookup("slot")
	assert.True(t, ok)
	assert.Equal(t, "au", ad)
	assert.Equal(t, 99, pid)
	_, _, ok = a.Lookup("")
	assert.False(t, ok)
	_, _, ok = a.Lookup("nope")
	assert.False(t, ok)
}

func TestNew_ApsOwMappingWithCountry(t *testing.T) {
	dbOnceReset()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Two columns only: CountryPartnerFilterDB scans (country, featureValue) — matches getCountryPartnerFilteringData.
	mock.ExpectQuery("SELECT country, value FROM wrapper_metrics WHERE feature_id=1").
		WillReturnRows(sqlmock.NewRows([]string{"country", "value"}).
			AddRow("US", "partnerA"))
	mock.ExpectQuery("SELECT aps_slot_uuid, ad_unit_id, profile_id FROM wrapper_aps_adunit_mapping").
		WillReturnRows(sqlmock.NewRows([]string{"aps_slot_uuid", "ad_unit_id", "profile_id"}).
			AddRow("u1", "a1", 5))

	got := New(db, config.Database{
		CountryPartnerFilterMaxDbContextTimeout: 1,
		MaxDbContextTimeout:                     500, // ms; required for NewApsOwMappingDB (0 => immediate context deadline)
		Queries: config.Queries{
			GetCountryPartnerFilteringData: "SELECT country, value FROM wrapper_metrics WHERE feature_id=1",
			GetApsOwMapping:                "SELECT aps_slot_uuid, ad_unit_id, profile_id FROM wrapper_aps_adunit_mapping",
		},
	}, config.Cache{
		CountryPartnerFilterRefreshInterval: 1,
		ApsOwMappingRefreshInterval:         1,
	})

	assert.NotNil(t, got)
	assert.NotNil(t, got.countryPartnerFilterDB)
	assert.NotNil(t, got.apsOwMappingDB)
	ad, pid, ok := got.GetApsOwMapping("u1")
	assert.True(t, ok)
	assert.Equal(t, "a1", ad)
	assert.Equal(t, 5, pid)
	assert.NoError(t, mock.ExpectationsWereMet())
	_ = db.Close()
}

func TestMySqlDB_GetApsOwMapping_nil(t *testing.T) {
	var db *mySqlDB
	ad, pid, ok := db.GetApsOwMapping("x")
	assert.False(t, ok)
	assert.Equal(t, "", ad)
	assert.Equal(t, 0, pid)
}
