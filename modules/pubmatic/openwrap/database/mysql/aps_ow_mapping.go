package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// apsOwMappingSelectBySlot is the cache-miss single-row query; only the UUID argument varies per request.
// Must match the table/columns used by Queries.GetApsOwMapping (e.g. wrapper_aps_adunit_mapping).
const apsOwMappingSelectBySlot = `SELECT aps_slot_uuid, ad_unit_id, profile_id FROM wrapper_aps_adunit_mapping WHERE aps_slot_uuid = ?`

// ApsOwMappingEntry maps an APS slot UUID to OpenWrap ad unit and profile identifiers.
type ApsOwMappingEntry struct {
	AdUnitID  string
	ProfileID int
}

type ApsOwMappingDB struct {
	db                  *sql.DB
	refreshInterval     time.Duration
	cache               atomic.Value
	query               string
	MaxDbContextTimeout time.Duration
	// reloadMu serializes full refresh (query+Store) with per-slot merge so a stale full snapshot cannot
	// overwrite a merge that completed while the full query was in flight. Also serializes cache-miss loads.
	reloadMu sync.Mutex
	stopCh   chan struct{}
	stopOnce sync.Once
}

func NewApsOwMappingDB(db *sql.DB, refreshInterval time.Duration, query string, maxDbContextTimeout time.Duration) (*ApsOwMappingDB, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, errors.New("aps OW mapping query is required")
	}

	m := &ApsOwMappingDB{
		db:                  db,
		refreshInterval:     refreshInterval * time.Hour,
		query:               q,
		MaxDbContextTimeout: maxDbContextTimeout * time.Millisecond,
		stopCh:              make(chan struct{}),
	}

	if err := m.RefreshCache(); err != nil {
		return nil, fmt.Errorf("error initializing APS OW mapping cache: %w", err)
	}

	m.ScheduleRefresh()
	return m, nil
}

func (a *ApsOwMappingDB) RefreshCache() error {
	var (
		data map[string]ApsOwMappingEntry
		err  error
	)

	for i := 0; i < models.MaxRetryAttempts; i++ {
		a.reloadMu.Lock()
		data, err = a.getApsOwMappingData()
		if err != nil {
			a.reloadMu.Unlock()
			glog.V(models.LogLevelDebug).Infof("Retry %d: failed to load APS OW mapping cache: %v", i+1, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		a.cache.Store(data)
		a.reloadMu.Unlock()
		if glog.V(models.LogLevelDebug) {
			glog.Infof("APS-OW mapping cache refreshed: %d entries", len(data))
		}
		return nil
	}

	glog.Errorf("failed to load APS OW mapping cache: %v", err)
	return fmt.Errorf("APS OW mapping cache load failed after retries: %w", err)
}

// getApsOwMappingData loads all mappings using the configured query (same as CountryPartnerFilterDB.getCountryPartnerFilteringData).
func (a *ApsOwMappingDB) getApsOwMappingData() (map[string]ApsOwMappingEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.MaxDbContextTimeout)
	defer cancel()

	rows, err := a.db.QueryContext(ctx, a.query)
	if err != nil {
		return nil, fmt.Errorf("query execution error: %w", err)
	}
	defer rows.Close()

	out := make(map[string]ApsOwMappingEntry)
	for rows.Next() {
		var slotUUID, adUnitID string
		var profileID int64
		if err := rows.Scan(&slotUUID, &adUnitID, &profileID); err != nil {
			glog.Errorf("APS OW mapping row scan error: %v", err)
			continue
		}
		slotUUID = strings.TrimSpace(slotUUID)
		if slotUUID == "" {
			continue
		}
		out[slotUUID] = ApsOwMappingEntry{
			AdUnitID:  strings.TrimSpace(adUnitID),
			ProfileID: int(profileID),
		}
	}

	return out, rows.Err()
}

func validApsOwMappingEntry(e ApsOwMappingEntry) bool {
	return e.AdUnitID != "" && e.ProfileID > 0
}

// getApsOwMappingSingle loads one row for slotUUID using apsOwMappingSelectBySlot.
func (a *ApsOwMappingDB) getApsOwMappingSingle(slotUUID string) (ApsOwMappingEntry, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.MaxDbContextTimeout)
	defer cancel()

	var slot, adUnitID string
	var profileID int64
	err := a.db.QueryRowContext(ctx, apsOwMappingSelectBySlot, slotUUID).Scan(&slot, &adUnitID, &profileID)
	if err == sql.ErrNoRows {
		return ApsOwMappingEntry{}, false, nil
	}
	if err != nil {
		return ApsOwMappingEntry{}, false, err
	}
	slot = strings.TrimSpace(slot)
	if slot == "" {
		return ApsOwMappingEntry{}, false, nil
	}
	return ApsOwMappingEntry{
		AdUnitID:  strings.TrimSpace(adUnitID),
		ProfileID: int(profileID),
	}, true, nil
}

func (a *ApsOwMappingDB) mergeEntryIntoCache(slotUUID string, entry ApsOwMappingEntry) {
	val := a.cache.Load()
	var old map[string]ApsOwMappingEntry
	if val != nil {
		old = val.(map[string]ApsOwMappingEntry)
	}
	newMap := make(map[string]ApsOwMappingEntry, len(old)+1)
	for k, v := range old {
		newMap[k] = v
	}
	newMap[slotUUID] = entry
	a.cache.Store(newMap)
}

func (a *ApsOwMappingDB) ScheduleRefresh() {
	if a == nil || a.stopCh == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(a.refreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := a.RefreshCache(); err != nil {
					glog.Errorf("scheduled APS OW mapping cache refresh failed: %v", err)
				}
			case <-a.stopCh:
				return
			}
		}
	}()
}

// Stop ends the background refresh goroutine started by ScheduleRefresh (safe to call more than once).
func (a *ApsOwMappingDB) Stop() {
	if a == nil {
		return
	}
	a.stopOnce.Do(func() {
		if a.stopCh != nil {
			close(a.stopCh)
		}
	})
}

func (a *ApsOwMappingDB) Lookup(slotUUID string) (adUnitID string, profileID int, found bool) {
	if slotUUID == "" {
		return "", 0, false
	}
	val := a.cache.Load()
	if val == nil {
		return "", 0, false
	}
	m := val.(map[string]ApsOwMappingEntry)
	e, ok := m[slotUUID]
	if !ok || !validApsOwMappingEntry(e) {
		return "", 0, false
	}
	return e.AdUnitID, e.ProfileID, true
}

// lookupOrLoadSingleRow checks the in-memory cache first. Full table load happens only in RefreshCache (startup
// and scheduled ticker). On miss, apsOwMappingSelectBySlot loads that UUID from the DB and merges into the map;
// on error or unknown UUID the map is unchanged.
func (a *ApsOwMappingDB) lookupOrLoadSingleRow(slotUUID string) (adUnitID string, profileID int, found bool) {
	if adUnitID, profileID, found = a.Lookup(slotUUID); found {
		return adUnitID, profileID, true
	}
	if slotUUID == "" {
		return "", 0, false
	}

	a.reloadMu.Lock()
	defer a.reloadMu.Unlock()

	if adUnitID, profileID, found = a.Lookup(slotUUID); found {
		return adUnitID, profileID, true
	}

	entry, rowOk, err := a.getApsOwMappingSingle(slotUUID)
	if err != nil {
		glog.Errorf("APS OW mapping single-row load failed: %v", err)
		return "", 0, false
	}
	if !rowOk || !validApsOwMappingEntry(entry) {
		return "", 0, false
	}
	a.mergeEntryIntoCache(slotUUID, entry)
	return entry.AdUnitID, entry.ProfileID, true
}

func (db *mySqlDB) GetApsOwMapping(slotUUID string) (adUnitID string, profileID int, found bool) {
	if db == nil || db.apsOwMappingDB == nil {
		return "", 0, false
	}
	return db.apsOwMappingDB.lookupOrLoadSingleRow(slotUUID)
}
