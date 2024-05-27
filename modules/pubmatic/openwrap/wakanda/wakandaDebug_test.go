package wakanda

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnable(t *testing.T) {
	pubID := "31445"
	profID := "55"
	config := Wakanda{HostName: "", DCName: "DC1"}
	InitWakanda(config)
	_ = wakandaRulesMap.AddIfNotPresent(generateKeyFromWakandaRequest(pubID, profID), 2, "local")

	wd := &Debug{
		Enabled: false,
	}

	wd.EnableIfRequired(pubID, profID)
	assert.Equal(t, true, wd.Enabled, "Enabled expected to be true found false")
	assert.Equal(t, 2, wd.DebugLevel, "DebugLevel expected to be 2 found %d", wd.DebugLevel)
	assert.Equal(t, "local__PUB:31445__PROF:55", wd.FolderPaths[0], "FolderPath[0] expected to be local__PUB:31445__PROF:55 found %s", wd.FolderPaths[0])
}

func TestDebugWriteLogToFiles(t *testing.T) {
	type fields struct {
		Enabled     bool
		FolderPaths []string
		DebugLevel  int
		DebugData   DebugData
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "disabled",
			fields: fields{
				DebugLevel: 1,
			},
		},
		{
			name: "invalid_json",
			fields: fields{
				DebugLevel:  2,
				FolderPaths: []string{`pub_1`, `pub_1_prof_1`},
				DebugData: DebugData{
					HTTPRequestBody: json.RawMessage(`{'invalid_json`),
				},
			},
		},
		{
			name: "valid_json",
			fields: fields{
				DebugLevel:  2,
				FolderPaths: []string{`pub_1`, `pub_1_prof_1`},
				DebugData: DebugData{
					HTTPRequestBody: json.RawMessage(`{key:"value"}`),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wD := &Debug{
				Enabled:     tt.fields.Enabled,
				FolderPaths: tt.fields.FolderPaths,
				DebugLevel:  tt.fields.DebugLevel,
				DebugData:   tt.fields.DebugData,
			}
			wD.WriteLogToFiles()
		})
	}
}
