package mysql

import (
	"strconv"

	"github.com/golang/glog"
)

func (db *mySqlDB) GetFSCDisabledPublishers() (map[int]struct{}, error) {
	rows, err := db.conn.Query(db.cfg.Queries.GetAllFscDisabledPublishersQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fscDisabledPublishers := make(map[int]struct{})
	for rows.Next() {
		var pubid int
		if err := rows.Scan(&pubid); err == nil {
			fscDisabledPublishers[pubid] = struct{}{}
		}
	}
	return fscDisabledPublishers, nil
}

func (db *mySqlDB) GetFSCThresholdPerDSP() (map[int]int, error) {
	rows, err := db.conn.Query(db.cfg.Queries.GetAllDspFscPcntQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fscDspThresholds := make(map[int]int)
	for rows.Next() {
		var dspId int
		var fscThreshold string
		if err := rows.Scan(&dspId, &fscThreshold); err != nil {
			glog.Error("Error in getting dsp-thresholds details from DB:", err.Error())
			continue
		}
		// convert threshold string to int
		pcnt, err := strconv.Atoi(fscThreshold)
		if err != nil {
			glog.Errorf("Invalid fsc_pcnt value for dspId:%d, threshold:%v", dspId, fscThreshold)
			continue
		}
		fscDspThresholds[dspId] = pcnt
	}
	return fscDspThresholds, nil
}
