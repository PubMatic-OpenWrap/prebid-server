package mysql

import (
	"context"
	"time"

	"github.com/golang/glog"
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
			glog.Error("Error in getting AppSubIntegrationPath details from DB:", err.Error())
			continue
		}
		appSubIntegrationPathMap[asipKey] = asipValue
	}
	return appSubIntegrationPathMap, nil
}
