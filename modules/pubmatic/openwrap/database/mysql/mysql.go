package mysql

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
)

type mySqlDB struct {
	conn                   *sql.DB
	cfg                    config.Database
	countryPartnerFilterDB *CountryPartnerFilterDB
	apsOwMappingDB         *ApsOwMappingDB
}

var db *mySqlDB
var dbOnce sync.Once

func New(conn *sql.DB, cfg config.Database, cache config.Cache) *mySqlDB {
	dbOnce.Do(
		func() {
			var cpf *CountryPartnerFilterDB
			if c, err := NewCountryPartnerFilterDB(conn, time.Duration(cache.CountryPartnerFilterRefreshInterval), cfg.Queries.GetCountryPartnerFilteringData, time.Duration(cfg.CountryPartnerFilterMaxDbContextTimeout)); err != nil {
				glog.Errorf("country partner filter cache init failed: %v", err)
			} else {
				cpf = c
			}

			var aps *ApsOwMappingDB
			if q := strings.TrimSpace(cfg.Queries.GetApsOwMapping); q != "" {
				ri := cache.ApsOwMappingRefreshInterval
				if ri <= 0 {
					ri = 1
				}
				if a, err := NewApsOwMappingDB(conn, time.Duration(ri), q, time.Duration(cfg.MaxDbContextTimeout)); err != nil {
					glog.Errorf("APS OW mapping cache init failed: %v", err)
				} else {
					aps = a
				}
			}

			db = &mySqlDB{conn: conn, cfg: cfg, countryPartnerFilterDB: cpf, apsOwMappingDB: aps}
		})
	return db
}

// Shutdown stops background work owned by this DB handle (APS OW mapping refresh ticker).
func (db *mySqlDB) Shutdown() {
	if db == nil || db.apsOwMappingDB == nil {
		return
	}
	db.apsOwMappingDB.Stop()
}
