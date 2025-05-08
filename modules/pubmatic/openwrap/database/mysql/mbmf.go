package mysql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

func (db *mySqlDB) GetProfileAdUnitMultiFloors() (models.ProfileAdUnitMultiFloors, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetProfileAdUnitMultiFloors)
	if err != nil {
		return models.ProfileAdUnitMultiFloors{}, err
	}
	defer rows.Close()

	profileAdUnitMultiFloors := make(models.ProfileAdUnitMultiFloors)
	for rows.Next() {
		var (
			adunitName, value string
			profileID         int
		)
		if err := rows.Scan(&adunitName, &profileID, &value); err != nil {
			glog.Errorf(models.ErrDBRowScanFailed, models.ProfileAdUnitMultiFloorsQuery, "", profileID, err.Error())
			continue
		}

		var adUnitMultiFloors models.MultiFloors
		if err := json.Unmarshal([]byte(value), &adUnitMultiFloors); err != nil {
			glog.Errorf(models.ErrMBMFFloorsUnmarshal, "", profileID, err.Error())
			continue
		}
		// Ensure nested map exists
		if _, ok := profileAdUnitMultiFloors[profileID]; !ok {
			profileAdUnitMultiFloors[profileID] = make(map[string]models.MultiFloors)
		}
		profileAdUnitMultiFloors[profileID][adunitName] = adUnitMultiFloors
	}

	if err := rows.Err(); err != nil {
		return profileAdUnitMultiFloors, err
	}
	return profileAdUnitMultiFloors, nil
}
