package ulpdebug

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
)

const Prefix = "[EDS-DEBUG-ULP]"

// TODO(EDS): remove temporary Unity LevelPlay marshal investigation logs.

// ShouldTrace returns true when request.ext.prebid.bidderparams.pubmatic.eds is present.
func ShouldTrace(req *openrtb2.BidRequest) bool {
	if req == nil || len(req.Ext) == 0 {
		return false
	}
	eds, dataType, _, err := jsonparser.Get(req.Ext, "prebid", "bidderparams", "pubmatic", "eds")
	return err == nil && dataType == jsonparser.Object && len(eds) > 2
}

// Wiid reads ext.prebid.bidderparams.pubmatic.wiid when present.
func Wiid(req *openrtb2.BidRequest) string {
	if req == nil || len(req.Ext) == 0 {
		return ""
	}
	raw, dataType, _, err := jsonparser.Get(req.Ext, "prebid", "bidderparams", "pubmatic", "wiid")
	if err != nil || dataType != jsonparser.String || len(raw) == 0 {
		return ""
	}
	var wiid string
	if err := json.Unmarshal(raw, &wiid); err != nil {
		return string(raw)
	}
	return wiid
}

func LogNote(wiid, stage, note string) {
	glog.Warningf("%s wiid=%s stage=%s note=%s", Prefix, wiid, stage, note)
}

func LogStageOK(wiid, stage string) {
	glog.Warningf("%s wiid=%s stage=%s result=ok", Prefix, wiid, stage)
}

func LogStageErr(wiid, stage string, err error) {
	if err == nil {
		return
	}
	glog.Warningf("%s wiid=%s stage=%s result=err err=%v", Prefix, wiid, stage, err)
}

func LogRawBytes(wiid, stage, field string, b []byte) {
	valid := len(b) == 0 || json.Valid(b)
	glog.Warningf("%s wiid=%s stage=%s field=%s len=%d valid=%v", Prefix, wiid, stage, field, len(b), valid)
	if len(b) > 0 && !valid {
		glog.Warningf("%s wiid=%s stage=%s field=%s invalid_bytes=%q", Prefix, wiid, stage, field, truncateForLog(b, 160))
	}
}

func LogMarshalError(wiid, stage, label string, err error) {
	if err == nil {
		return
	}
	glog.Warningf("%s wiid=%s stage=%s marshal=%s err=%v", Prefix, wiid, stage, label, err)
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
		glog.Warningf("%s wiid=%s stage=%s bidderparams_unmarshal_err=%v", Prefix, wiid, stage, err)
		return
	}

	for bidder, params := range paramsMap {
		for key, val := range params {
			LogRawBytes(wiid, stage, fmt.Sprintf("bidderparams.%s.%s", bidder, key), val)
			if len(val) == 0 {
				glog.Warningf("%s wiid=%s stage=%s empty_rawmessage_key=%s.%s", Prefix, wiid, stage, bidder, key)
			}
		}
	}
}

func LogMergeCopy(wiid, stage, impID, bidder, key string, value json.RawMessage) {
	LogRawBytes(wiid, stage, fmt.Sprintf("merge_imp=%s bidder=%s key=%s", impID, bidder, key), value)
}

func LogNestedSubslice(wiid, stage, field string, parentExt, nested []byte) {
	if len(parentExt) == 0 || len(nested) == 0 {
		return
	}
	idx := bytes.Index(parentExt, nested)
	sharesBacking := idx >= 0 && idx+len(nested) <= len(parentExt)
	glog.Warningf("%s wiid=%s stage=%s field=%s subslice_of_parent=%v parent_len=%d nested_len=%d nested_offset=%d",
		Prefix, wiid, stage, field, sharesBacking, len(parentExt), len(nested), idx)
}

// CloneRawMessage returns an owned copy of b, or nil when empty.
func CloneRawMessage(b json.RawMessage) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	copied := make(json.RawMessage, len(b))
	copy(copied, b)
	return copied
}

// CloneBidderParamsMap returns a deep copy of bidder params so map values are not subslices of request.ext.
func CloneBidderParamsMap(in map[string]map[string]json.RawMessage) map[string]map[string]json.RawMessage {
	if in == nil {
		return nil
	}
	out := make(map[string]map[string]json.RawMessage, len(in))
	for bidder, params := range in {
		cloned := make(map[string]json.RawMessage, len(params))
		for k, v := range params {
			cloned[k] = CloneRawMessage(v)
		}
		out[bidder] = cloned
	}
	return out
}

// RepairExt returns an owned copy when b is valid JSON, or nil when empty or unrecoverable.
func RepairExt(b json.RawMessage) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	if json.Valid(b) {
		return CloneRawMessage(b)
	}
	compacted := bytes.NewBuffer(make([]byte, 0, len(b)))
	if err := json.Compact(compacted, b); err == nil && json.Valid(compacted.Bytes()) {
		return CloneRawMessage(compacted.Bytes())
	}
	return nil
}

// ProbeBidRequestExts logs json validity for every ext field on the bid request.
func ProbeBidRequestExts(wiid, stage string, req *openrtb2.BidRequest) {
	if req == nil {
		return
	}
	LogRawBytes(wiid, stage, "request.ext", req.Ext)
	if req.App != nil {
		LogRawBytes(wiid, stage, "app.ext", req.App.Ext)
	}
	if req.Device != nil {
		LogRawBytes(wiid, stage, "device.ext", req.Device.Ext)
	}
	if req.User != nil {
		LogRawBytes(wiid, stage, "user.ext", req.User.Ext)
	}
	if req.Source != nil {
		LogRawBytes(wiid, stage, "source.ext", req.Source.Ext)
	}
	if req.Regs != nil {
		LogRawBytes(wiid, stage, "regs.ext", req.Regs.Ext)
	}
	for i := range req.Imp {
		field := fmt.Sprintf("imp[%d].ext id=%s", i, req.Imp[i].ID)
		LogRawBytes(wiid, stage, field, req.Imp[i].Ext)
	}
}

// NormalizeBidRequestExts replaces ext fields with owned copies of valid JSON before marshaling.
func NormalizeBidRequestExts(req *openrtb2.BidRequest) {
	if req == nil {
		return
	}
	req.Ext = RepairExt(req.Ext)
	if req.App != nil {
		req.App.Ext = RepairExt(req.App.Ext)
	}
	if req.Device != nil {
		req.Device.Ext = RepairExt(req.Device.Ext)
	}
	if req.User != nil {
		req.User.Ext = RepairExt(req.User.Ext)
	}
	if req.Source != nil {
		req.Source.Ext = RepairExt(req.Source.Ext)
	}
	if req.Regs != nil {
		req.Regs.Ext = RepairExt(req.Regs.Ext)
	}
	for i := range req.Imp {
		req.Imp[i].Ext = RepairExt(req.Imp[i].Ext)
	}
}

func truncateForLog(b []byte, max int) string {
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "..."
}
