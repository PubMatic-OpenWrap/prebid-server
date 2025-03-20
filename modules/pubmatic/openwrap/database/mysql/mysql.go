package mysql

import (
	"database/sql"
	"sync"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
)

type mySqlDB struct {
	conn *sql.DB
	cfg  config.Database
}

var db *mySqlDB
var dbOnce sync.Once

func New(conn *sql.DB, cfg config.Database) *mySqlDB {
	dbOnce.Do(
		func() {
			db = &mySqlDB{conn: conn, cfg: cfg}
		})
	return db
}
