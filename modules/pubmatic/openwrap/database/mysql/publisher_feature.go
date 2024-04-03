package mysql

import (
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func (db *mySqlDB) GetPublisherFeatureMap() (map[int]map[int]models.FeatureData, error) {
	rows, err := db.conn.Query(db.cfg.Queries.GetPublisherFeatureMapQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	publisherFeatureMap := make(map[int]map[int]models.FeatureData)
	for rows.Next() {
		var pubId, featureId, enabled int
		var value string
		if err := rows.Scan(&pubId, &featureId, &enabled, &value); err != nil {
			glog.Error("ErrRowScanFailed GetPublisherFeatureMap pubid: ", pubId, " err: ", err.Error())
			continue
		}
		if _, ok := publisherFeatureMap[pubId]; !ok {
			publisherFeatureMap[pubId] = make(map[int]models.FeatureData)
		}
		publisherFeatureMap[pubId][featureId] = models.FeatureData{
			Enabled: enabled,
			Value:   value,
		}
	}
	return publisherFeatureMap, nil
}
