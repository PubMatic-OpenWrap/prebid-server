package mysql

import (
	"context"
	"time"

	"github.com/golang/glog"
)

func (db *mySqlDB) GetGDPRCountryCodes() (map[string]struct{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetGDPRCountryCodes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// map to store the country codes
	countryCodes := make(map[string]struct{})
	for rows.Next() {
		var countryCode string
		if err := rows.Scan(&countryCode); err != nil {
			glog.Error("ErrRowScanFailed GetGDPRCountryCodes Err: ", err.Error())
			continue
		}
		countryCodes[countryCode] = struct{}{}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return countryCodes, nil
}
