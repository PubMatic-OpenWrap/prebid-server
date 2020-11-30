package openrtb2

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestAddTargetingKeys(t *testing.T) {
	var tests = []struct {
		scenario string // Testcase scenario
		key      string
		value    string
		bidExt   string
	}{
		{scenario: "key_not_exists", key: "hb_pb_cat_dur", value: "some_value", bidExt: `{ "prebid" : { "targeting" : {} } }`},
		{scenario: "key_already_exists", key: "hb_pb_cat_dur", value: "new_value", bidExt: `{ "prebid" : { "targeting" : { "hb_pb_cat_dur" : "old_value" } } }`},
	}

	for _, test := range tests {
		bid := new(openrtb.Bid)
		bid.Ext = []byte(test.bidExt)
		key := openrtb_ext.TargetingKey(test.key)
		assert.Nil(t, addTargetingKey(bid, key, test.value))
		assertTargetingKeyExists(t, key, test.value, *bid)
	}
	assert.Equal(t, "Invalid bid", addTargetingKey(nil, openrtb_ext.HbCategoryDurationKey, "some value").Error())
}

func assertTargetingKeyExists(t *testing.T, expecteKey openrtb_ext.TargetingKey, expectedValue string, bid openrtb.Bid) {
	t.Helper()
	bidExt := make(map[string]map[string]map[string]string)
	err := json.Unmarshal(bid.Ext, &bidExt)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	prebid := bidExt["prebid"]
	for k, v := range prebid["targeting"] {
		if k == string(expecteKey) && fmt.Sprintf("%v", v) == expectedValue {
			assert.True(t, true)
			return
		}
	}

	assert.Fail(t, "key not found")
}
