package eds

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

const ulpDebugPrefix = "[EDS-DEBUG-ULP]"

// TODO(EDS): remove temporary Unity LevelPlay marshal investigation logs.

func LogRawBytes(wiid, stage, field string, b []byte) {
	valid := len(b) == 0 || json.Valid(b)
	glog.Warningf("%s wiid=%s stage=%s field=%s len=%d valid=%v", ulpDebugPrefix, wiid, stage, field, len(b), valid)
	if len(b) > 0 && !valid {
		glog.Warningf("%s wiid=%s stage=%s field=%s invalid_bytes=%q", ulpDebugPrefix, wiid, stage, field, truncateForLog(b, 160))
	}
}

func LogResolvedEds(wiid, stage string, resolved models.ResolvedEds) {
	LogRawBytes(wiid, stage, "resolved.device", resolved.Device)
	LogRawBytes(wiid, stage, "resolved.app", resolved.App)
}

func LogMarshalError(wiid, stage, label string, err error) {
	if err == nil {
		return
	}
	glog.Warningf("%s wiid=%s stage=%s marshal=%s err=%v", ulpDebugPrefix, wiid, stage, label, err)
}

func ProbeMarshal(wiid, stage, label string, v interface{}) {
	_, err := json.Marshal(v)
	LogMarshalError(wiid, stage, label, err)
}

func LogBidderParamsRaw(wiid, stage string, bidderParams json.RawMessage) {
	LogRawBytes(wiid, stage, "bidderparams", bidderParams)
	if len(bidderParams) == 0 {
		return
	}

	var paramsMap map[string]map[string]json.RawMessage
	if err := json.Unmarshal(bidderParams, &paramsMap); err != nil {
		glog.Warningf("%s wiid=%s stage=%s bidderparams_unmarshal_err=%v", ulpDebugPrefix, wiid, stage, err)
		return
	}

	for bidder, params := range paramsMap {
		for key, val := range params {
			LogRawBytes(wiid, stage, fmt.Sprintf("bidderparams.%s.%s", bidder, key), val)
			if len(val) == 0 {
				glog.Warningf("%s wiid=%s stage=%s empty_rawmessage_key=%s.%s", ulpDebugPrefix, wiid, stage, bidder, key)
			}
		}
	}
}

func LogNestedEdsSubslice(wiid, stage string, parentExt, nested []byte) {
	if len(parentExt) == 0 || len(nested) == 0 {
		return
	}

	idx := bytes.Index(parentExt, nested)
	sharesBacking := idx >= 0 && idx+len(nested) <= len(parentExt)
	glog.Warningf("%s wiid=%s stage=%s nested_eds_subslice_of_parent=%v parent_len=%d nested_len=%d nested_offset=%d",
		ulpDebugPrefix, wiid, stage, sharesBacking, len(parentExt), len(nested), idx)
}

func LogNote(wiid, stage, msg string) {
	glog.Warningf("%s wiid=%s stage=%s note=%s", ulpDebugPrefix, wiid, stage, msg)
}

func glogWarningEmptyKey(wiid, bidder, key string, valLen int) {
	glog.Warningf("%s wiid=%s stage=inject invalid_param_key=%s.%s len=%d", ulpDebugPrefix, wiid, bidder, key, valLen)
}

func truncateForLog(b []byte, max int) string {
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "..."
}
