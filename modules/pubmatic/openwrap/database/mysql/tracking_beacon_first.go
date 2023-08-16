// Package mysql provides functionalities to interact with the giym database.
// This file is used for retrieving and managing data related to the tracking-beacon-first (TBF) feature for publishers.
// This includes methods to fetch and process tracking-beacon-first traffic details associated
// with publisher IDs from the giym database.
package mysql

import (
	"encoding/json"

	"github.com/golang/glog"
)

// GetTBFTrafficForPublishers function fetches the publisher data for TBF (tracking-beacon-first) feature from database
func (db *mySqlDB) GetTBFTrafficForPublishers() (map[int]map[int]int, error) {
	rows, err := db.conn.Query(db.cfg.Queries.GetTBFRateQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pubProfileTrafficRate := make(map[int]map[int]int)
	for rows.Next() {
		var pubID int
		var trafficDetails string

		if err := rows.Scan(&pubID, &trafficDetails); err != nil {
			glog.Error("ErrRowScanFailed GetTBFRateQuery pubid: ", pubID, " err: ", err.Error())
			continue
		}

		// convert trafficDetails into map[profileId]traffic
		var profileTrafficRate map[int]int
		if err := json.Unmarshal([]byte(trafficDetails), &profileTrafficRate); err != nil {
			glog.Error("ErrJSONUnmarshalFailed TBFProfileTrafficRate pubid: ", pubID, " trafficDetails: ", trafficDetails, " err: ", err.Error())
			continue
		}
		pubProfileTrafficRate[pubID] = profileTrafficRate
	}
	return pubProfileTrafficRate, nil
}
