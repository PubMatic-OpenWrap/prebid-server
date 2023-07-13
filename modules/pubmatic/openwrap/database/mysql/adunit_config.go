package mysql

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

// GetAdunitConfig - Method to get adunit config for a given profile and display version from giym DB
func (db *mySqlDB) GetAdunitConfig(profileID, displayVersionID int) (*adunitconfig.AdUnitConfig, error) {
	var adunitConfigJSON string
	adunitConfig := new(adunitconfig.AdUnitConfig)
	adunitConfigQuery := ""
	if displayVersionID == 0 {
		adunitConfigQuery = db.cfg.Queries.GetAdunitConfigForLiveVersion
	} else {
		adunitConfigQuery = db.cfg.Queries.GetAdunitConfigQuery
	}
	adunitConfigQuery = strings.Replace(adunitConfigQuery, profileIdKey, strconv.Itoa(profileID), -1)
	adunitConfigQuery = strings.Replace(adunitConfigQuery, displayVersionKey, strconv.Itoa(displayVersionID), -1)
	err := db.conn.QueryRow(adunitConfigQuery).Scan(&adunitConfigJSON)
	if err != nil {
		err = fmt.Errorf("[QUERY_FAILED] Name:[%v] Error:[%v]", "GetAdunitConfig", err.Error())
		return nil, err
	}

	err = json.Unmarshal([]byte(adunitConfigJSON), &adunitConfig)
	if err != nil {
		return nil, err
	}

	for k, v := range adunitConfig.Config {
		adunitConfig.Config[strings.ToLower(k)] = v
		// shall we delete the orignal key-val?
	}

	if adunitConfig.ConfigPattern == "" {
		//Default configPattern value is "_AU_" if not present in db config
		adunitConfig.ConfigPattern = models.MACRO_AD_UNIT_ID
	}
	return adunitConfig, err
}
