package mysql

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adpodconfig"
)

func (db *mySqlDB) GetAdpodConfigs(profileID, displayVersion int) (*adpodconfig.AdpodConfig, error) {
	adpodConfigQuery := db.cfg.Queries.GetAdpodConfig
	if displayVersion == 0 {
		adpodConfigQuery = db.cfg.Queries.GetAdpodConfigForLiveVersion
	}

	adpodConfigQuery = strings.Replace(adpodConfigQuery, profileIdKey, strconv.Itoa(profileID), -1)
	adpodConfigQuery = strings.Replace(adpodConfigQuery, displayVersionKey, strconv.Itoa(displayVersion), -1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, adpodConfigQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var config *adpodconfig.AdpodConfig
	for rows.Next() {
		var podType, podConfig string
		var err error

		if err = rows.Scan(&podType, &podConfig); err != nil {
			continue
		}

		if len(podConfig) > 0 && config == nil {
			config = &adpodconfig.AdpodConfig{}
		}

		switch strings.ToLower(podType) {
		case models.AdPodTypeDynamic:
			err = json.Unmarshal([]byte(podConfig), &config.Dynamic)
		case models.AdPodTypeStructured:
			err = json.Unmarshal([]byte(podConfig), &config.Structured)
		case models.AdPodTypeHybrid:
			err = json.Unmarshal([]byte(podConfig), &config.Hybrid)
		}

		if err != nil {
			return nil, err
		}
	}

	if err = rows.Err(); err != nil {
		glog.Errorf("adpod config row scan failed for profile %d with versionID %d", profileID, displayVersion)
	}

	return config, nil
}
