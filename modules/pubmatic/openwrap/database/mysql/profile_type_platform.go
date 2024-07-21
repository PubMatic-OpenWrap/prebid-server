package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
)

func (db *mySqlDB) GetProfileTypePlatforms() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	getProfileTypePlatformMapQuery := fmt.Sprintf(db.cfg.Queries.GetProfileTypePlatformMapQuery, db.cfg.MaxQueryExecution)
	rows, err := db.conn.QueryContext(ctx, getProfileTypePlatformMapQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profileTypePlatformMap := make(map[string]int)
	for rows.Next() {
		var ptpKey string
		var ptpValue int
		if err := rows.Scan(&ptpKey, &ptpValue); err != nil {
			glog.Error("Error in getting profileTypePlatform details from DB:", err.Error())
			continue
		}
		profileTypePlatformMap[ptpKey] = ptpValue
	}
	return profileTypePlatformMap, nil
}
