package mysql

import (
	"database/sql"
	"sync"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
)

type mySqlDB struct {
	conn *sql.DB
	cfg  config.Database
}

type PlatformStruct struct {
	// other fields
	Platform sql.NullString `json:"platform" db:"platform"`
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
