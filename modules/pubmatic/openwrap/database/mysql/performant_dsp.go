package mysql

import (
	"context"
	"strconv"
	"time"

	"github.com/golang/glog"
)

func (db *mySqlDB) GetPerformantDSPs() (map[int]struct{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetPerformantDSPQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	performantDSPs := make(map[int]struct{})
	for rows.Next() {
		var dspId int
		var value string
		if err := rows.Scan(&dspId, &value); err != nil {
			glog.Error("Error in getting performant-dsp details from DB:", err.Error())
			continue
		}
		// convert threshold string to int
		isEnable, err := strconv.Atoi(value)
		if err != nil {
			glog.Errorf("Invalid enable value for dspId:%d, value:%v", dspId, value)
			continue
		}

		if isEnable == 1 {
			performantDSPs[dspId] = struct{}{}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return performantDSPs, nil
}
