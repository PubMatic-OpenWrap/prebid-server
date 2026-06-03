package sdkutils

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/stretchr/testify/assert"
)

func TestMergeDeviceExtFromSignal(t *testing.T) {
	signalExt := json.RawMessage(`{
		"boottime":1710000000000,
		"pbtime":1710000001000,
		"diskspace":18201.5,
		"totaldisk":64000,
		"inputlaunguage":["en","fr"],
		"charging":1,
		"batterylevel":0.85,
		"totalmem":8589934592,
		"dnh":"abc123",
		"sua":{"browsers":[{"brand":"Chrome","version":["120"]}]},
		"ringmute":0,
		"darkmode":1,
		"bluetooth":0,
		"airplane":-1,
		"dnd":0,
		"headset":1,
		"screenbright":0.75,
		"atts":3,
		"ifv":"SIGNAL-IFV"
	}`)

	got := MergeDeviceExtFromSignal(signalExt, nil)

	assert.JSONEq(t, string(signalExt), string(got))
}

func TestMergeDeviceExtFromSignal_PrefersSignalOverRequest(t *testing.T) {
	requestExt := json.RawMessage(`{"boottime":1,"ifv":"REQUEST-IFV"}`)
	signalExt := json.RawMessage(`{"boottime":1710000000000,"ifv":"SIGNAL-IFV"}`)

	got := MergeDeviceExtFromSignal(signalExt, requestExt)

	assert.Contains(t, string(got), `"boottime":1710000000000`)
	assert.Contains(t, string(got), `"ifv":"SIGNAL-IFV"`)
}

func TestMergeAppExtFromSignal(t *testing.T) {
	signalExt := json.RawMessage(`{"install_time":1710000000000,"first_launch_time":1710000002000}`)
	got := MergeAppExtFromSignal(signalExt, json.RawMessage(`{"token":"remove-me"}`))

	assert.Contains(t, string(got), `"install_time":1710000000000`)
	assert.Contains(t, string(got), `"first_launch_time":1710000002000`)
	assert.Contains(t, string(got), `"token":"remove-me"`)
}

func TestMergeImpLTVFieldsFromSignal(t *testing.T) {
	dst := &openrtb2.Imp{
		Rwdd: 0,
		Banner: &openrtb2.Banner{
			MIMEs: []string{"image/jpeg"},
		},
	}
	src := &openrtb2.Imp{
		Rwdd: 1,
		Banner: &openrtb2.Banner{
			MIMEs: []string{"image/png", "image/gif"},
		},
	}

	MergeImpLTVFieldsFromSignal(dst, src)

	assert.Equal(t, int8(1), dst.Rwdd)
	assert.Equal(t, []string{"image/png", "image/gif"}, dst.Banner.MIMEs)
}

func TestMergeDeviceCopiesPPI(t *testing.T) {
	dst := &openrtb2.Device{}
	src := &openrtb2.Device{PPI: 326}

	got := MergeDevice(dst, src)

	assert.Equal(t, int64(326), got.PPI)
}
