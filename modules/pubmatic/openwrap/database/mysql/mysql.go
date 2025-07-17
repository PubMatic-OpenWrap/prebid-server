package mysql

import (
	"database/sql"
	"sync"
	"time"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
)

type mySqlDB struct {
	conn                   *sql.DB
	cfg                    config.Database
	countryPartnerFilterDB *CountryPartnerFilterDB
}

var db *mySqlDB
var dbOnce sync.Once

func New(conn *sql.DB, cfg config.Database, cache config.Cache) *mySqlDB {
	dbOnce.Do(
		func() {
			cpf, err := NewCountryPartnerFilterDB(conn, time.Duration(cache.CountryPartnerFilterRefreshInterval), cfg.Queries.GetCountryPartnerFilteringData)
			if err != nil {
				db = &mySqlDB{conn: conn, cfg: cfg}
				return
			}
			db = &mySqlDB{conn: conn, cfg: cfg, countryPartnerFilterDB: cpf}
		})
	return db
}
