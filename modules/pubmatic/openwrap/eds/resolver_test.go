package eds

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/sdk/signal"
	"github.com/stretchr/testify/assert"
)

func TestResolveSignalEdsOnly(t *testing.T) {
	signalReq := &openrtb2.BidRequest{
		Device: &openrtb2.Device{
			Ext: json.RawMessage(`{"eds":{"boottime":1710000000000,"totalmem":8589934592}}`),
		},
		App: &openrtb2.App{
			Ext: json.RawMessage(`{"eds":{"install_time":1710000000001}}`),
		},
	}

	resolved := Resolve(Sources{Signal: signalReq})

	assert.JSONEq(t, `{"boottime":1710000000000,"totalmem":8589934592}`, string(resolved.Device))
	assert.JSONEq(t, `{"install_time":1710000000001}`, string(resolved.App))
}

func TestResolveRequestEdsOnly(t *testing.T) {
	request := &openrtb2.BidRequest{
		Device: &openrtb2.Device{
			Ext: json.RawMessage(`{"eds":{"boottime":1710000000000}}`),
		},
		App: &openrtb2.App{
			Ext: json.RawMessage(`{"eds":{"first_launch_time":1710000000002}}`),
		},
	}

	resolved := Resolve(Sources{Request: request})

	assert.JSONEq(t, `{"boottime":1710000000000}`, string(resolved.Device))
	assert.JSONEq(t, `{"first_launch_time":1710000000002}`, string(resolved.App))
}

func TestResolveSignalEdsIgnoresDirectExtKeys(t *testing.T) {
	signalReq := &openrtb2.BidRequest{
		Device: &openrtb2.Device{
			Ext: json.RawMessage(`{"eds":{"boottime":1710000000000},"boottime":999,"totalmem":8589934592}`),
		},
	}
	request := &openrtb2.BidRequest{
		Device: &openrtb2.Device{
			Ext: json.RawMessage(`{"eds":{"totalmem":8589934592},"boottime":999}`),
		},
	}

	resolved := Resolve(Sources{Signal: signalReq, Request: request})

	assert.JSONEq(t, `{"boottime":1710000000000,"totalmem":8589934592}`, string(resolved.Device))
}

func TestMergeGapFill(t *testing.T) {
	base := models.ResolvedEds{
		Device: json.RawMessage(`{"boottime":1710000000000}`),
	}
	overlay := models.ResolvedEds{
		Device: json.RawMessage(`{"boottime":999,"totalmem":8589934592}`),
		App:    json.RawMessage(`{"install_time":1710000000001}`),
	}

	merged := MergeGapFill(base, overlay)

	assert.JSONEq(t, `{"boottime":1710000000000,"totalmem":8589934592}`, string(merged.Device))
	assert.JSONEq(t, `{"install_time":1710000000001}`, string(merged.App))
}

func TestStripFromRequest(t *testing.T) {
	req := &openrtb2.BidRequest{
		Device: &openrtb2.Device{
			Ext: json.RawMessage(`{"eds":{"boottime":1710000000000},"boottime":1710000000000,"atts":1}`),
		},
		App: &openrtb2.App{
			Ext: json.RawMessage(`{"eds":{"install_time":1710000000001},"install_time":1710000000001,"orientation":1}`),
		},
	}

	StripFromRequest(req, models.ResolvedEds{
		Device: json.RawMessage(`{"boottime":1710000000000}`),
		App:    json.RawMessage(`{"install_time":1710000000001}`),
	})

	assert.JSONEq(t, `{"atts":1}`, string(req.Device.Ext))
	assert.JSONEq(t, `{"orientation":1}`, string(req.App.Ext))
}

func TestStripFromRequestRemovesEmptyExt(t *testing.T) {
	req := &openrtb2.BidRequest{
		App: &openrtb2.App{
			Ext: json.RawMessage(`{"eds":{"install_time":1710000000001},"install_time":1710000000001}`),
		},
	}

	StripFromRequest(req, models.ResolvedEds{
		App: json.RawMessage(`{"install_time":1710000000001}`),
	})

	assert.Nil(t, req.App.Ext)
}

func TestApplyToRequest(t *testing.T) {
	req := &openrtb2.BidRequest{
		Device: &openrtb2.Device{Ext: json.RawMessage(`{"atts":1}`)},
		App:    &openrtb2.App{Ext: json.RawMessage(`{"orientation":1}`)},
	}
	resolved := models.ResolvedEds{
		Device: json.RawMessage(`{"boottime":1710000000000}`),
		App:    json.RawMessage(`{"install_time":1710000000001}`),
	}

	ApplyToRequest(req, resolved)

	assert.JSONEq(t, `{"atts":1,"boottime":1710000000000}`, string(req.Device.Ext))
	assert.JSONEq(t, `{"orientation":1,"install_time":1710000000001}`, string(req.App.Ext))
}

func TestInjectAndExtractBidderParamsEds(t *testing.T) {
	resolved := models.ResolvedEds{
		Device: json.RawMessage(`{"boottime":1710000000000}`),
		App:    json.RawMessage(`{"install_time":1710000000001}`),
	}

	injected, err := InjectIntoBidderParams(nil, resolved, "pubmatic")
	assert.NoError(t, err)

	var params map[string]map[string]json.RawMessage
	assert.NoError(t, json.Unmarshal(injected, &params))
	assert.NotNil(t, params["pubmatic"]["eds"])

	flatParams, err := json.Marshal(params["pubmatic"])
	assert.NoError(t, err)
	extracted := ExtractFromBidderParams(flatParams)
	assert.JSONEq(t, string(resolved.Device), string(extracted.Device))
	assert.JSONEq(t, string(resolved.App), string(extracted.App))
}

func TestParseAPSSignal(t *testing.T) {
	innerSignal := `{"device":{"ext":{"eds":{"boottime":1710000000000}}}}`
	body, err := json.Marshal(map[string]map[string]string{
		"user": {"buyeruid": innerSignal},
	})
	assert.NoError(t, err)

	signalReq := signal.ParseAPS(body)
	assert.NotNil(t, signalReq)
	assert.NotNil(t, signalReq.Device)

	resolved := Resolve(Sources{Signal: signalReq})
	assert.JSONEq(t, `{"boottime":1710000000000}`, string(resolved.Device))
}
