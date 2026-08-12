package mysql

import (
	"context"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (db *mySqlDB) GetGDPRCountryCodes() (models.HashSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetGDPRCountryCodes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		countryCode string
		// map to store the country codes
		countryCodes = make(models.HashSet)
	)
	for rows.Next() {
		if err := rows.Scan(&countryCode); err != nil {
			glog.Errorf(models.ErrDBRowScanFailed, models.GDPRCountryCodesQuery, "", "", err.Error())
			continue
		}
		countryCodes[countryCode] = struct{}{}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return countryCodes, nil
}
