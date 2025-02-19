package publisherfeature

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

const (
	errInvalidAnalyticsThrottlingKeyFormat = `[AnalyticsThrottling] status:[INVALID] part:[%v] key:[%v]`
	keySeparator                           = ","
	keyPartSeparator                       = ":"
	allProfiles                            = 0
)

const (
	keyPartPubID             = 0
	keyPartProfileID         = 1
	keyPartLoggerPercentage  = 2
	keyPartTrackerPercentage = 3
)

// percentageValue type defines throttling percentage for analytics logger and tracker
type percentageValue struct {
	logger  int
	tracker int
}

func (p *percentageValue) isThrottled() (loggerThrottled, trackerThrottled bool) {
	loggerThrottled = p.logger < 0 || predictAnalyticsThrottle(p.logger)

	if loggerThrottled && p.logger >= 0 {
		return true, true
	}

	if p.tracker > 0 {
		trackerThrottled = predictAnalyticsThrottle(p.tracker)
	}
	return
}

type pubThrottling map[int]map[int]*percentageValue

func newPubThrottling(defaultValue string) pubThrottling {
	ant := make(pubThrottling)
	ant.add(defaultValue)
	return ant
}

func (ant pubThrottling) add(value string) {
	// config format: <pubid>:<profileid>:<logger_%>:<tracker_%>,<pubid>:<profileid>:<logger_%>:<tracker_%>,...
	keys := strings.Split(value, keySeparator)
	for _, key := range keys {
		if key == "" {
			continue
		}
		pubID, profileID, loggerPercentage, trackerPercentage, err := getKeyParts(key)
		if err != nil {
			continue
		}

		if ant[pubID] == nil {
			ant[pubID] = map[int]*percentageValue{}
		}

		ant[pubID][profileID] = &percentageValue{
			logger:  loggerPercentage,
			tracker: trackerPercentage,
		}
	}
}

func (ant pubThrottling) merge(source pubThrottling, replaceExisting bool) {
	for pubID, profiles := range source {
		for profileID, percentage := range profiles {
			if ant[pubID] == nil {
				ant[pubID] = map[int]*percentageValue{}
			}
			if replaceExisting || ant[pubID][profileID] == nil {
				ant[pubID][profileID] = percentage
			}
		}
	}
}

func getKeyParts(key string) (int, int, int, int, error) {
	var err error
	var pubID, profileID, loggerPercentage, trackerPercentage int

	parts := strings.Split(key, keyPartSeparator)
	if len(parts) != 4 {
		return 0, 0, 0, 0, fmt.Errorf(errInvalidAnalyticsThrottlingKeyFormat, "LENGTH", key)
	}

	if pubID, err = strconv.Atoi(parts[keyPartPubID]); err != nil || pubID <= 0 {
		return 0, 0, 0, 0, fmt.Errorf(errInvalidAnalyticsThrottlingKeyFormat, "PUBID", key)
	}

	if profileID, err = strconv.Atoi(parts[keyPartProfileID]); err != nil || profileID < 0 {
		return 0, 0, 0, 0, fmt.Errorf(errInvalidAnalyticsThrottlingKeyFormat, "PROFILEID", key)
	}

	if loggerPercentage, err = strconv.Atoi(parts[keyPartLoggerPercentage]); err != nil || loggerPercentage < -1 {
		return 0, 0, 0, 0, fmt.Errorf(errInvalidAnalyticsThrottlingKeyFormat, "LOGGER_PERCENTAGE", key)
	}

	if trackerPercentage, err = strconv.Atoi(parts[keyPartTrackerPercentage]); err != nil || trackerPercentage < 0 {
		return 0, 0, 0, 0, fmt.Errorf(errInvalidAnalyticsThrottlingKeyFormat, "TRACKER_PERCENTAGE", key)
	}

	return pubID, profileID, loggerPercentage, trackerPercentage, err
}

type analyticsThrottle struct {
	vault, db pubThrottling
}

func (at *analyticsThrottle) getThrottlingPercentage(pubID, profileID int) *percentageValue {
	if pubThrottling, ok := at.db[pubID]; ok {
		if value, ok := pubThrottling[profileID]; ok {
			return value
		}
		if value, ok := pubThrottling[allProfiles]; ok {
			return value
		}
	}
	return nil
}

func (fe *feature) updateAnalyticsThrottling() {
	if fe.publisherFeature == nil {
		return
	}

	ant := make(pubThrottling)
	for _, feature := range fe.publisherFeature {
		if val, ok := feature[models.FeatureAnalyticsThrottle]; ok && val.Enabled == 1 {
			ant.add(val.Value)
		}
	}

	if len(ant) == 0 {
		return
	}

	//add vault settings to db
	ant.merge(fe.ant.db, false)

	fe.Lock()
	fe.ant.db = ant
	fe.Unlock()
}

var randInt = rand.Intn

func predictAnalyticsThrottle(threshold int) bool {
	return randInt(100) < threshold
}

// IsAnalyticsTrackingThrottled returns throttling logger,tracker or not
func (fe *feature) IsAnalyticsTrackingThrottled(pubID, profileID int) (bool, bool) {
	fe.RLocker().Lock()
	throttlingPercentage := fe.ant.getThrottlingPercentage(pubID, profileID)
	fe.RLocker().Unlock()

	if throttlingPercentage == nil {
		return false, false
	}
	return throttlingPercentage.isThrottled()
}
