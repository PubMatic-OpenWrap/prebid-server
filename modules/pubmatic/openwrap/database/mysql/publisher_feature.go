package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
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
		var pubId, featureId, enabled int
		var value sql.NullString
		if err := rows.Scan(&pubId, &featureId, &enabled, &value); err != nil {
			glog.Error("ErrRowScanFailed GetPublisherFeatureMap pubid: ", pubId, " err: ", err.Error())
			continue
		}
		if _, ok := publisherFeatureMap[pubId]; !ok {
			publisherFeatureMap[pubId] = make(map[int]models.FeatureData)
		}
		publisherFeatureMap[pubId][featureId] = models.FeatureData{
			Enabled: enabled,
			Value:   value.String,
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return publisherFeatureMap, nil
}
