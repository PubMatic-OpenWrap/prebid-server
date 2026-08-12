package mysql

import (
	"context"
	"time"

	"github.com/golang/glog"
)

func (db *mySqlDB) GetPerformanceDSPs() (map[int]struct{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetPerformanceDSPQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	performanceDSPs := make(map[int]struct{})
	for rows.Next() {
		var dspId int
		if err := rows.Scan(&dspId); err != nil {
			glog.Error("Error in getting performance-dsp details from DB:", err.Error())
			continue
		}
		performanceDSPs[dspId] = struct{}{}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return performanceDSPs, nil
}
