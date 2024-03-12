// Package mysql provides functionalities to interact with the giym database.
// This file is used for retrieving and managing data related to the RTBBidders.
package mysql

import (
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

	"github.com/golang/glog"
)

// GetRTBBidders function fetches the RTB bidder, bidder_info and bidder_params from wrapper_partner table
func (db *mySqlDB) GetRTBBidders() (map[string]models.RTBBidderData, error) {
	rows, err := db.conn.Query(db.cfg.Queries.GetRTBBidders)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rtbBidders := make(map[string]models.RTBBidderData, 0)
	for rows.Next() {
		var bidderName string
		var bidderData models.RTBBidderData

		if err := rows.Scan(&bidderName, &bidderData.BidderInfo, &bidderData.BidderParams); err != nil {
			glog.Error("ErrRowScanFailed GetRTBBidders err:", err.Error())
			continue
		}
		rtbBidders[bidderName] = bidderData
	}
	return rtbBidders, nil
}
