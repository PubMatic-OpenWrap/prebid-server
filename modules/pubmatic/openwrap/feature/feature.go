package feature

import (
	"database/sql"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
)

const (
	// FeatureNameGoogleSDK is the name of the Google SDK feature
	FeatureNameGoogleSDK = "googlesdk"
)

type Features map[string][]Feature

type Feature struct {
	Name string
	Data any
}

type FeatureLoader struct {
	db  *sql.DB
	cfg config.Database
}

func NewFeatureLoader(db *sql.DB, config config.Database) *FeatureLoader {
	return &FeatureLoader{
		db:  db,
		cfg: config,
	}
}

func (fl *FeatureLoader) LoadFeatures() Features {
	features := make(Features)

	// load Google SDK features
	features[FeatureNameGoogleSDK] = fl.LoadGoogleSDKFeatures()

	return features
}
