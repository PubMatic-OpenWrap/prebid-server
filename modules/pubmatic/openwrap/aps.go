package openwrap

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/buger/jsonparser"
	gocache "github.com/patrickmn/go-cache"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/metrics"
)

// Prometheus label values for metrics.aps_slot_mapping_rejects (reason).
const (
	apsMetricReasonUnmappedUUID  = "unmapped_uuid"
	negativeCacheTimeout         = 15 * time.Minute
	negativeCacheCleanupInterval = 1 * time.Minute
)

var apsNegativeCache = gocache.New(negativeCacheTimeout, negativeCacheCleanupInterval)

// setApsRequestExtWrapperProfileID sets req.Ext JSON field ext.prebid.bidderparams.pubmatic.wrapper.profileid (numeric).
func setApsRequestExtWrapperProfileID(req *openrtb2.BidRequest, profileID int) error {
	ext := req.Ext
	if len(ext) == 0 {
		ext = []byte(`{}`)
	}
	newExt, err := jsonparser.Set(ext, []byte(strconv.Itoa(profileID)), "prebid", "bidderparams", "pubmatic", "wrapper", "profileid")
	if err != nil {
		return fmt.Errorf("aps: set request wrapper.profileid: %w", err)
	}
	req.Ext = newExt
	return nil
}

// enrichApsRequest replaces imp[0].tagid with the mapped OW ad unit id and sets
// ext.prebid.bidderparams.pubmatic.wrapper.profileid from that mapping.
// Only the first impression is modified; any additional imps are left unchanged.
// publisherID is used for metrics labels when me is non-nil.
func enrichApsRequest(body []byte, owCache cache.Cache, me metrics.MetricsEngine, publisherID string) ([]byte, error) {
	if owCache == nil {
		return nil, fmt.Errorf("aps: cache not configured")
	}
	req := &openrtb2.BidRequest{}
	if err := json.Unmarshal(body, req); err != nil {
		return nil, fmt.Errorf("aps: unmarshal bid request: %w", err)
	}

	if len(req.Imp) == 0 {
		return nil, fmt.Errorf("aps: no impressions")
	}

	imp := &req.Imp[0]
	slotUUID := imp.TagID
	if slotUUID == "" {
		return nil, fmt.Errorf("aps: empty or missing imp[0].tagid")
	}

	if _, hit := apsNegativeCache.Get(slotUUID); hit {
		if me != nil {
			me.RecordAPSSlotMappingReject(publisherID, slotUUID, apsMetricReasonUnmappedUUID)
		}
		return nil, fmt.Errorf("aps: slot uuid %q found in negative cache", slotUUID)
	}

	adUnitID, profileID, ok := owCache.GetApsOwMapping(slotUUID)
	if !ok {
		apsNegativeCache.Set(slotUUID, struct{}{}, negativeCacheTimeout)
		if me != nil {
			me.RecordAPSSlotMappingReject(publisherID, slotUUID, apsMetricReasonUnmappedUUID)
		}
		return nil, fmt.Errorf("aps: no mapping for slot uuid %q", slotUUID)
	}

	imp.TagID = adUnitID

	if err := setApsRequestExtWrapperProfileID(req, profileID); err != nil {
		return nil, err
	}

	modifiedRequestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("aps: marshal bid request: %w", err)
	}

	return modifiedRequestBody, nil
}
