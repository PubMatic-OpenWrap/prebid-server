package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

type CountryPartnerFilterDB struct {
	db                  *sql.DB
	refreshInterval     time.Duration
	cache               atomic.Value
	query               string
	MaxDbContextTimeout time.Duration
}

func NewCountryPartnerFilterDB(db *sql.DB, refreshInterval time.Duration, query string, maxDbContextTimeout time.Duration) (*CountryPartnerFilterDB, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}

	filter := &CountryPartnerFilterDB{
		db:                  db,
		refreshInterval:     refreshInterval * time.Hour,
		query:               query,
		MaxDbContextTimeout: maxDbContextTimeout * time.Millisecond,
	}

	if err := filter.RefreshCache(); err != nil {
		return nil, fmt.Errorf("error initializing filter cache: %w", err)
	}

	filter.ScheduleRefresh()
	return filter, nil
}

func (cpf *CountryPartnerFilterDB) RefreshCache() error {
	var (
		data map[string]map[string]struct{}
		err  error
	)

	for i := 0; i < models.MaxRetryAttempts; i++ {
		data, err = cpf.getCountryPartnerFilteringData()
		if err == nil {
			break
		}
		glog.V(models.LogLevelDebug).Infof("Retry %d: Failed to load cache: %v", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		glog.Errorf("Failed to load cache for country filtering: %v", err)
		return fmt.Errorf("cache load failed after retries: %w", err)
	}

	cpf.cache.Store(data)
	return nil
}

func (cpf *CountryPartnerFilterDB) getCountryPartnerFilteringData() (map[string]map[string]struct{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cpf.MaxDbContextTimeout)
	defer cancel()

	rows, err := cpf.db.QueryContext(ctx, cpf.query)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	result := make(map[string]map[string]struct{})
	for rows.Next() {
		var country, featureValue string

		if err := rows.Scan(&country, &featureValue); err != nil {
			glog.Errorf("Scan error getThrottledPartnersByCountry: %v", err)
			continue
		}

		if result[country] == nil {
			result[country] = make(map[string]struct{})
		}
		result[country][featureValue] = struct{}{}
	}

	return result, rows.Err()
}

func (cpf *CountryPartnerFilterDB) ScheduleRefresh() {
	go func() {
		ticker := time.NewTicker(cpf.refreshInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := cpf.RefreshCache(); err != nil {
				glog.Errorf("Scheduled cache refresh failed: %v", err)
			}
		}
	}()
}

func (cpf *CountryPartnerFilterDB) GetLatestCountryPartnerFilter() map[string]map[string]struct{} {
	val := cpf.cache.Load()
	if val == nil {
		return nil
	}
	return val.(map[string]map[string]struct{})
}

func (db *mySqlDB) GetLatestCountryPartnerFilter() map[string]map[string]struct{} {
	if db.countryPartnerFilterDB == nil {
		return nil
	}
	return db.countryPartnerFilterDB.GetLatestCountryPartnerFilter()
}
