package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"sync/atomic"
	"time"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type CountryPartnerFilterDB struct {
	db              *sql.DB
	refreshInterval time.Duration
	cache           atomic.Value
	query           string
}

func NewCountryPartnerFilterDB(db *sql.DB, refreshInterval time.Duration, query string) (*CountryPartnerFilterDB, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}

	filter := &CountryPartnerFilterDB{
		db:              db,
		refreshInterval: refreshInterval * time.Hour,
		query:           query,
	}

	if err := filter.RefreshCache(); err != nil {
		return nil, fmt.Errorf("error initializing filter cache: %w", err)
	}

	filter.ScheduleRefresh()
	return filter, nil
}

func (cpf *CountryPartnerFilterDB) RefreshCache() error {
	var (
		data map[string][]models.PartnerFeatureRecord
		err  error
	)

	for i := 0; i < 3; i++ {
		data, err = cpf.getCountryPartnerFilteringData()
		if err == nil {
			break
		}
		log.Printf("Retry %d: Failed to load cache: %v", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return fmt.Errorf("cache load failed after retries: %w", err)
	}

	cpf.cache.Store(data)
	return nil
}

func (cpf *CountryPartnerFilterDB) getCountryPartnerFilteringData() (map[string][]models.PartnerFeatureRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := cpf.db.QueryContext(ctx, cpf.query)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]models.PartnerFeatureRecord)
	for rows.Next() {
		var record models.PartnerFeatureRecord
		var threshold int64
		if err := rows.Scan(&record.Country, &record.FeatureValue, &record.Criteria, &threshold); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		if threshold > math.MaxInt32 || threshold < math.MinInt32 {
			log.Printf("Threshold out of bounds: %d for country %s", threshold, record.Country)
			continue
		}
		record.CriteriaThreshold = int(threshold)
		result[record.Country] = append(result[record.Country], record)
	}
	return result, rows.Err()
}

func (cpf *CountryPartnerFilterDB) ScheduleRefresh() {
	go func() {
		ticker := time.NewTicker(cpf.refreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := cpf.RefreshCache(); err != nil {
				log.Printf("Scheduled cache refresh failed: %v", err)
			}
		}
	}()
}

func (cpf *CountryPartnerFilterDB) GetLatestCountryPartnerFilter() map[string][]models.PartnerFeatureRecord {
	val := cpf.cache.Load()
	if val == nil {
		return nil
	}
	return val.(map[string][]models.PartnerFeatureRecord)
}

func (db *mySqlDB) GetLatestCountryPartnerFilter() map[string][]models.PartnerFeatureRecord {
	if db.countryPartnerFilterDB == nil {
		return nil
	}
	return db.countryPartnerFilterDB.GetLatestCountryPartnerFilter()
}
