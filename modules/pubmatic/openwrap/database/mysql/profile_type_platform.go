package mysql

import (
	"context"
	"time"

	"github.com/golang/glog"
)

func (db *mySqlDB) GetProfileTypePlatform() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetProfileTypePlatformQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profileTypePlatform := make(map[string]int)
	for rows.Next() {
		var ptpKey string
		var ptpValue int
		if err := rows.Scan(&ptpKey, &ptpValue); err != nil {
			glog.Error("Error in getting profileTypePlatform details from DB:", err.Error())
			continue
		}
		profileTypePlatform[ptpKey] = ptpValue
	}
	return profileTypePlatform, nil
}
