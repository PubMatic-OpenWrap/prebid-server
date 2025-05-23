package mysql

import (
	"context"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (db *mySqlDB) GetProfileTypePlatforms() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetProfileTypePlatformMapQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profileTypePlatformMap := make(map[string]int)
	for rows.Next() {
		var ptpKey string
		var ptpValue int
		if err := rows.Scan(&ptpKey, &ptpValue); err != nil {
			glog.Errorf(models.ErrDBRowScanFailed, models.ProfileTypePlatformMapQuery, "", "", err.Error())
			continue
		}
		profileTypePlatformMap[ptpKey] = ptpValue
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return profileTypePlatformMap, nil
}
