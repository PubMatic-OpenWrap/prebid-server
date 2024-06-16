package mysql

import (
	"context"
	"time"

	"github.com/golang/glog"
)

func (db *mySqlDB) GetAppSubIntegrationPath() (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetAppSubIntegrationPathQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	AppSubIntegrationPath := make(map[string]int)
	for rows.Next() {
		var asipKey string
		var asipValue int
		if err := rows.Scan(&asipKey, &asipValue); err != nil {
			glog.Error("Error in getting AppSubIntegrationPath details from DB:", err.Error())
			continue
		}
		AppSubIntegrationPath[asipKey] = asipValue
	}
	return AppSubIntegrationPath, nil
}
