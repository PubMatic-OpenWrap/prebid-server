package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
)

func (db *mySqlDB) GetAppIntegrationPaths() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	getAppIntegrationPathMapQuery := fmt.Sprintf(db.cfg.Queries.GetAppIntegrationPathMapQuery, db.cfg.MaxQueryExecution)
	rows, err := db.conn.QueryContext(ctx, getAppIntegrationPathMapQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appIntegrationPathMap := make(map[string]int)
	for rows.Next() {
		var aipKey string
		var aipValue int
		if err := rows.Scan(&aipKey, &aipValue); err != nil {
			glog.Error("Error in getting AppIntegrationPath details from DB:", err.Error())
			continue
		}
		appIntegrationPathMap[aipKey] = aipValue
	}
	return appIntegrationPathMap, nil
}
