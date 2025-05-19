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

// Tests
func TestLoadGoogleSDKFeatures_Success(t *testing.T) {
	mock := newMockDB(t)
	defer mock.Close()

	rows := sqlmock.NewRows([]string{"slot_size"}).
		AddRow("300x250").
		AddRow("728x90")

	mock.ExpectQuery("SELECT .* FROM banner_sizes").WillReturnRows(rows)

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

	assert.Len(t, features, 1)
	assert.Equal(t, FeatureFlexSlot, features[0].Name)
	assert.ElementsMatch(t, []string{"300x250", "728x90"}, features[0].Data)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadGoogleSDKFeatures_QueryError(t *testing.T) {
	mock := newMockDB(t)
	defer mock.Close()

	mock.ExpectQuery("SELECT slot_size FROM banner_sizes").WillReturnError(errors.New("db error"))

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
	assert.Nil(t, features)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadGoogleSDKFeatures_ScanError(t *testing.T) {
	mock := newMockDB(t)
	defer mock.Close()

	rows := sqlmock.NewRows([]string{"slot_size"}).
		AddRow(nil) // will cause Scan error for string

	mock.ExpectQuery("SELECT slot_size FROM banner_sizes").WillReturnRows(rows)

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
	assert.Nil(t, features)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadGoogleSDKFeatures_RowsError(t *testing.T) {
	mock := newMockDB(t)
	defer mock.Close()

	rows := sqlmock.NewRows([]string{"slot_size"}).
		AddRow("300x250")
	rows.RowError(0, errors.New("row error"))

	mock.ExpectQuery("SELECT slot_size FROM banner_sizes").WillReturnRows(rows)

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
	assert.Nil(t, features)
	assert.NoError(t, mock.ExpectationsWereMet())
}
