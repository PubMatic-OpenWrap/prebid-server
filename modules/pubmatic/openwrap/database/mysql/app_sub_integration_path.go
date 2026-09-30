package mysql

import (
	"context"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (db *mySqlDB) GetAppSubIntegrationPaths() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetAppSubIntegrationPathMapQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appSubIntegrationPathMap := make(map[string]int)
	for rows.Next() {
		var asipKey string
		var asipValue int
		if err := rows.Scan(&asipKey, &asipValue); err != nil {
			glog.Errorf(models.ErrDBRowScanFailed, models.AppSubIntegrationPathMapQuery, "", "", err.Error())
			continue
		}
		appSubIntegrationPathMap[asipKey] = asipValue
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return appSubIntegrationPathMap, nil
}
