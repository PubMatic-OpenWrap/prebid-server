package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// return the list of active server side header bidding partners
// with their configurations at publisher-profile-version level
func (db *mySqlDB) GetActivePartnerConfigurations(pubID, profileID int, displayVersion int) (map[int]map[string]string, error) {
	versionID, displayVersionID, platform, err := db.getVersionID(profileID, displayVersion, pubID)
	if err != nil {
		return nil, err
	}

	partnerConfigMap, err := db.getActivePartnerConfigurations(versionID)
	if err == nil && partnerConfigMap[-1] != nil {
		partnerConfigMap[-1][models.DisplayVersionID] = strconv.Itoa(displayVersionID)
		if platform != "" {
			partnerConfigMap[-1][models.PLATFORM_KEY] = platform
		}
	}
	return partnerConfigMap, err
}

func (db *mySqlDB) getActivePartnerConfigurations(versionID int) (map[int]map[string]string, error) {
	getActivePartnersQuery := fmt.Sprintf(db.cfg.Queries.GetParterConfig, db.cfg.MaxDbContextTimeout, versionID, versionID, versionID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()
	rows, err := db.conn.QueryContext(ctx, getActivePartnersQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	partnerConfigMap := make(map[int]map[string]string, 0)
	for rows.Next() {
		var (
			keyName, value, prebidPartnerName, bidderCode          string
			partnerID, entityTypeID, testConfig, isAlias, vendorID int
		)
		if err := rows.Scan(&partnerID, &prebidPartnerName, &bidderCode, &isAlias, &entityTypeID, &testConfig, &vendorID, &keyName, &value); err != nil {
			continue
		}

		_, ok := partnerConfigMap[partnerID]
		//below logic will take care of overriding account level partner keys with version level partner keys
		//if key name is same for a given partnerID (Ref ticket: UOE-5647)
		if !ok {
			partnerConfigMap[partnerID] = map[string]string{models.PARTNER_ID: strconv.Itoa(partnerID)}
		}

		if testConfig == 1 {
			keyName = keyName + "_test"
			partnerConfigMap[partnerID][models.PartnerTestEnabledKey] = "1"
		}

		partnerConfigMap[partnerID][keyName] = value

		if _, ok := partnerConfigMap[partnerID][models.PREBID_PARTNER_NAME]; !ok && prebidPartnerName != "-" {
			partnerConfigMap[partnerID][models.PREBID_PARTNER_NAME] = prebidPartnerName
			partnerConfigMap[partnerID][models.BidderCode] = bidderCode
			partnerConfigMap[partnerID][models.IsAlias] = strconv.Itoa(isAlias)
			if prebidPartnerName != models.BidderVASTBidder {
				partnerConfigMap[partnerID][models.VENDORID] = strconv.Itoa(vendorID)
			}
		}
	}

	// NYC_TODO: ignore close error
	if err = rows.Err(); err != nil {
		glog.Errorf("partner config row scan failed for versionID %d", versionID)
	}
	return partnerConfigMap, nil
}

func (db *mySqlDB) getVersionID(profileID, displayVersion, pubID int) (int, int, string, error) {
	var row *sql.Row
	if displayVersion == 0 {
		row = db.conn.QueryRow(db.cfg.Queries.LiveVersionInnerQuery, profileID, pubID)
	} else {
		row = db.conn.QueryRow(db.cfg.Queries.DisplayVersionInnerQuery, profileID, displayVersion, pubID)
	}

	var platformStruct PlatformStruct
	var versionID, displayVersionIDFromDB int
	err := row.Scan(&versionID, &displayVersionIDFromDB, &platformStruct.Platform)
	if err != nil {
		return versionID, displayVersionIDFromDB, platformStruct.Platform.String, err
	}
	return versionID, displayVersionIDFromDB, platformStruct.Platform.String, nil
}
