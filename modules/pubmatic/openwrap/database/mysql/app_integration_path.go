package mysql

import (
	"context"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (db *mySqlDB) GetAppIntegrationPaths() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetAppIntegrationPathMapQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appIntegrationPathMap := make(map[string]int)
	for rows.Next() {
		var aipKey string
		var aipValue int
		if err := rows.Scan(&aipKey, &aipValue); err != nil {
			glog.Errorf(models.ErrDBRowScanFailed, models.AppIntegrationPathMapQuery, "", "", err.Error())
			continue
		}
		appIntegrationPathMap[aipKey] = aipValue
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return appIntegrationPathMap, nil
}
