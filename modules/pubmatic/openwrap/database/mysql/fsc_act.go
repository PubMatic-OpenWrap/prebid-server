package mysql

import (
	"context"
	"strconv"
	"time"

	"github.com/golang/glog"
)

// GetFSCAndACTThresholdsPerDSP returns both FSC and ACT DSP thresholds using the single query
// cfg.Queries.GetAllDspFscAndActPcntQuery (expected columns: dsp_id, value, key_name with key_name 'fsc_pcnt' or 'act_pcnt').
// Returns empty maps when the query is not configured.
func (db *mySqlDB) GetFSCAndACTThresholdsPerDSP() (fscMap map[int]int, actMap map[int]int, err error) {
	if db.cfg.Queries.GetAllDspFscAndActPcntQuery == "" {
		return make(map[int]int), make(map[int]int), nil
	}
	return db.getFSCAndACTThresholdsPerDSP()
}

func (db *mySqlDB) getFSCAndACTThresholdsPerDSP() (fscMap map[int]int, actMap map[int]int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetAllDspFscAndActPcntQuery)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	fscMap = make(map[int]int)
	actMap = make(map[int]int)
	for rows.Next() {
		var dspId int
		var value string
		var keyName string
		if err := rows.Scan(&dspId, &value, &keyName); err != nil {
			glog.Error("Error scanning dsp feature thresholds from DB:", err.Error())
			continue
		}
		pcnt, err := strconv.Atoi(value)
		if err != nil {
			glog.Errorf("Invalid threshold value for dspId:%d key_name:%s value:%v", dspId, keyName, value)
			continue
		}
		switch keyName {
		case "fsc_pcnt":
			fscMap[dspId] = pcnt
		case "act_pcnt":
			actMap[dspId] = pcnt
		}
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	return fscMap, actMap, nil
}
