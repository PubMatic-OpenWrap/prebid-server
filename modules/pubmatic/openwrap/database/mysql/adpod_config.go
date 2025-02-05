package mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adpodconfig"
)

func (db *mySqlDB) GetAdpodConfig(pubID, profileID, displayVersion int) (*adpodconfig.AdpodConfig, error) {
	versionID, displayVersion, _, _, err := db.getVersionIdAndProfileDetails(profileID, displayVersion, pubID)
	if err != nil {
		return nil, fmt.Errorf("LiveVersionInnerQuery/DisplayVersionInnerQuery Failure Error: %w", err)
	}

	config, err := db.getAdpodConfig(versionID)
	if err != nil {
		return nil, fmt.Errorf("GetAdpodConfigQuery Failure Error: %w", err)
	}

	return config, nil
}

func (db *mySqlDB) getAdpodConfig(versionID int) (*adpodconfig.AdpodConfig, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(db.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, db.cfg.Queries.GetAdpodConfig, versionID)
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
		return nil, err
	}

	return config, nil
}
