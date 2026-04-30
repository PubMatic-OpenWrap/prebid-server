package openwrap

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
)

// Prometheus label values for metrics.aps_slot_mapping_rejects (reason).
const (
	apsMetricReasonUnmappedUUID  = "unmapped_uuid"
	negativeCacheTimeout         = 15 * time.Minute
	negativeCacheCleanupInterval = 1 * time.Minute
)

var apsNegativeCache = gocache.New(negativeCacheTimeout, negativeCacheCleanupInterval)

// setApsWrapperProfileIDOnBody sets ext.prebid.bidderparams.pubmatic.wrapper.profileid on raw request JSON (numeric).
func setApsWrapperProfileIDOnBody(body []byte, profileID int) ([]byte, error) {
	out, err := jsonparser.Set(body, []byte(strconv.Itoa(profileID)), "ext", "prebid", "bidderparams", "pubmatic", "wrapper", "profileid")
	if err != nil {
		return nil, fmt.Errorf("aps: set request wrapper.profileid: %w", err)
	}
	return out, nil
}

// resolveApsSlotMapping resolves imp[0] APS slot UUID (tagid) to OW ad unit id and profile id via owCache.
// It applies negative caching and reject metrics consistent with enrichApsRequest.
func resolveApsSlotMapping(owCache cache.Cache, me metrics.MetricsEngine, publisherID, slotUUID string) (adUnitID string, profileID int, err error) {
	if _, hit := apsNegativeCache.Get(slotUUID); hit {
		if me != nil {
			me.RecordAPSSlotMappingReject(publisherID, slotUUID, apsMetricReasonUnmappedUUID)
		}
		return "", 0, fmt.Errorf("aps: slot uuid %q found in negative cache", slotUUID)
	}

	adUnitID, profileID, ok := owCache.GetApsOwMapping(slotUUID)
	if !ok {
		apsNegativeCache.Set(slotUUID, struct{}{}, negativeCacheTimeout)
		if me != nil {
			me.RecordAPSSlotMappingReject(publisherID, slotUUID, apsMetricReasonUnmappedUUID)
		}
		return "", 0, fmt.Errorf("aps: no mapping for slot uuid %q", slotUUID)
	}
	return adUnitID, profileID, nil
}

// enrichApsRequest replaces imp[0].tagid with the mapped OW ad unit id and sets
// ext.prebid.bidderparams.pubmatic.wrapper.profileid from that mapping.
// Only the first impression is modified; any additional imps are left unchanged.
// publisherID is used for metrics labels when me is non-nil.
func enrichApsRequest(body []byte, owCache cache.Cache, me metrics.MetricsEngine, publisherID string) ([]byte, error, openrtb3.NoBidReason) {
	if owCache == nil {
		return nil, fmt.Errorf("aps: cache not configured"), openrtb3.NoBidInvalidRequest
	}
	if !json.Valid(body) {
		return nil, fmt.Errorf("aps: unmarshal bid request: invalid JSON"), openrtb3.NoBidInvalidRequest
	}
	if _, _, _, err := jsonparser.Get(body, "imp", "[0]"); err != nil {
		return nil, fmt.Errorf("aps: no impressions"), openrtb3.NoBidInvalidRequest
	}
	slotUUID, err := jsonparser.GetString(body, "imp", "[0]", "tagid")
	if err != nil || slotUUID == "" {
		return nil, fmt.Errorf("aps: empty or missing imp[0].tagid"), nbr.InvalidImpressionTagID
	}

	adUnitID, profileID, err := resolveApsSlotMapping(owCache, me, publisherID, slotUUID)
	if err != nil {
		return nil, err, nbr.APSSlotUUIDNotMapped
	}

	out, err := jsonparser.Set(body, []byte(strconv.Quote(adUnitID)), "imp", "[0]", "tagid")
	if err != nil {
		return nil, fmt.Errorf("aps: set imp[0].tagid: %w", err), openrtb3.NoBidInvalidRequest
	}

	out, err = setApsWrapperProfileIDOnBody(out, profileID)
	if err != nil {
		return nil, err, openrtb3.NoBidInvalidRequest
	}
	return out, nil, 0
}
