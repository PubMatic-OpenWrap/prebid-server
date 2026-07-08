package rawext

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/stretchr/testify/assert"
)

func TestCloneBidderParamsMapOwnsBytes(t *testing.T) {
	parent := json.RawMessage(`{"pubmatic":{"eds":{"device":{"boottime":1}}}}`)
	var params map[string]map[string]json.RawMessage
	assert.NoError(t, json.Unmarshal(parent, &params))

	cloned := CloneBidderParamsMap(params)
	eds := cloned["pubmatic"]["eds"]
	assert.True(t, json.Valid(eds))

	parent[0] = '{'
	assert.True(t, json.Valid(eds))
}

func TestRepairExtDropsInvalidBytes(t *testing.T) {
	assert.Nil(t, RepairExt(json.RawMessage(`}`)))
	assert.NotNil(t, RepairExt(json.RawMessage(`{"a":1}`)))
}

func TestNormalizeBidRequestExtsRepairsImpExt(t *testing.T) {
	req := &openrtb2.BidRequest{
		Imp: []openrtb2.Imp{{
			ID:  "imp1",
			Ext: json.RawMessage(`{"prebid":{"bidder":{"pubmatic":{"eds":{"device":{"boottime":1}}}}}}`),
		}},
	}
	NormalizeBidRequestExts(req)
	assert.True(t, json.Valid(req.Imp[0].Ext))
}
