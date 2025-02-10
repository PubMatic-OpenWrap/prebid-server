package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

func (db *mySqlDB) GetPublisherFeatureMap() (map[int]map[int]models.FeatureData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetPublisherFeatureMapQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	publisherFeatureMap := make(map[int]map[int]models.FeatureData)
	for rows.Next() {
		var pubID, featureID, enabled int
		var value sql.NullString
		if err := rows.Scan(&pubID, &featureID, &enabled, &value); err != nil {
			glog.Errorf(models.ErrDBRowScanFailed, models.PublisherFeatureMapQuery, pubID, "", err.Error())
			continue
		}
		if _, ok := publisherFeatureMap[pubID]; !ok {
			publisherFeatureMap[pubID] = make(map[int]models.FeatureData)
		}
		publisherFeatureMap[pubID][featureID] = models.FeatureData{
			Enabled: enabled,
			Value:   value.String,
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return publisherFeatureMap, nil
}
