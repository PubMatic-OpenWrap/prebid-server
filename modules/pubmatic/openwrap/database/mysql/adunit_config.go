package mysql

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
)

// GetAdunitConfig - Method to get adunit config for a given profile and display version from giym DB
func (db *mySqlDB) GetAdunitConfig(profileID, displayVersion int) (*adunitconfig.AdUnitConfig, error) {
	adunitConfigQuery := db.cfg.Queries.GetAdunitConfigQuery
	if displayVersion == 0 {
		adunitConfigQuery = db.cfg.Queries.GetAdunitConfigForLiveVersion
	}
	adunitConfigQuery = strings.Replace(adunitConfigQuery, profileIdKey, strconv.Itoa(profileID), -1)
	adunitConfigQuery = strings.Replace(adunitConfigQuery, displayVersionKey, strconv.Itoa(displayVersion), -1)

	var adunitConfigJSON string
	err := db.conn.QueryRow(adunitConfigQuery).Scan(&adunitConfigJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	adunitConfig := &adunitconfig.AdUnitConfig{}
	err = json.Unmarshal([]byte(adunitConfigJSON), &adunitConfig)
	if err != nil {
		return nil, adunitconfig.ErrAdUnitUnmarshal
	}

	for k, v := range adunitConfig.Config {
		adunitConfig.Config[strings.ToLower(k)] = v
		// shall we delete the orignal key-val?
	}

	if adunitConfig.ConfigPattern == "" {
		//Default configPattern value is "_AU_" if not present in db config
		adunitConfig.ConfigPattern = models.MACRO_AD_UNIT_ID
	}

	// safe check for old legacy profiles
	// new profiles cannot be created as UI-API has config object validation
	if adunitConfig.Config == nil {
		adunitConfig.Config = make(map[string]*adunitconfig.AdConfig)
	}

	if _, ok := adunitConfig.Config["default"]; !ok {
		adunitConfig.Config["default"] = &adunitconfig.AdConfig{}
	}

	return adunitConfig, err
}
