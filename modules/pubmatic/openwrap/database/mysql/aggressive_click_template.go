package mysql

import (
	"context"
	"strconv"
	"time"

	"github.com/golang/glog"
)

func (db *mySqlDB) GetACTThresholdPerDSP() (map[int]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetAllDspActPcntQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actDspThresholds := make(map[int]int)
	for rows.Next() {
		var dspId int
		var actThreshold string
		if err := rows.Scan(&dspId, &actThreshold); err != nil {
			glog.Error("Error in getting dsp-thresholds details from DB:", err.Error())
			continue
		}
		// convert threshold string to int
		pcnt, err := strconv.Atoi(actThreshold)
		if err != nil {
			glog.Errorf("Invalid act_pcnt value for dspId:%d, threshold:%v", dspId, actThreshold)
			continue
		}
		actDspThresholds[dspId] = pcnt
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return actDspThresholds, nil
}
