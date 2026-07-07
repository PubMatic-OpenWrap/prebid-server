package eds

import (
	"encoding/json"
	"strconv"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/util/ulpdebug"
)

// TODO(EDS): remove temporary Unity LevelPlay marshal investigation logs.

func LogRawBytes(wiid, stage, field string, b []byte) {
	ulpdebug.LogRawBytes(wiid, stage, field, b)
}

func LogResolvedEds(wiid, stage string, resolved models.ResolvedEds) {
	LogRawBytes(wiid, stage, "resolved.device", resolved.Device)
	LogRawBytes(wiid, stage, "resolved.app", resolved.App)
}

func LogMarshalError(wiid, stage, label string, err error) {
	ulpdebug.LogMarshalError(wiid, stage, label, err)
}

func ProbeMarshal(wiid, stage, label string, v interface{}) {
	ulpdebug.ProbeMarshal(wiid, stage, label, v)
}

func LogBidderParamsRaw(wiid, stage string, bidderParams json.RawMessage) {
	ulpdebug.LogBidderParamsRaw(wiid, stage, bidderParams)
}

func LogNestedEdsSubslice(wiid, stage string, parentExt, nested []byte) {
	ulpdebug.LogNestedSubslice(wiid, stage, "eds", parentExt, nested)
}

func LogNote(wiid, stage, msg string) {
	ulpdebug.LogNote(wiid, stage, msg)
}

func glogWarningEmptyKey(wiid, bidder, key string, valLen int) {
	ulpdebug.LogNote(wiid, "inject", "invalid_param_key="+bidder+"."+key+" len="+strconv.Itoa(valLen))
}
