package wakanda

import (
	"header-bidding/config"
	"testing"
)

func init() {
	config.ServerConfig = &config.DMHBConfig{}
	config.ServerConfig.OpenWrap.Server.DCName = "local"
}

func TestEnable(t *testing.T) {
	pubID := "31445"
	profID := "55"
	_ = wakandaRulesMap.AddIfNotPresent(generateKeyFromWakandaRequest(pubID, profID), 2)

	wd := &Debug{
		Enabled: false,
	}

	wd.EnableIfRequired(pubID, profID)
	if wd.Enabled != true {
		t.Errorf("Enabled expected to be true found false")
	}

	if wd.DebugLevel != 2 {
		t.Errorf("DebugLevel expected to be 2 found %d", wd.DebugLevel)
	}

	if wd.FolderPaths[0] != "local__PUB:31445__PROF:55" {
		t.Errorf("FolderPath[0] expected to be local__PUB:31445__PROF:55 found %s", wd.FolderPaths[0])
	}
}

func TestDebug_WriteLogToFiles(t *testing.T) {
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
		// {
		// 	name: "invalid_json",
		// 	fields: fields{
		// 		DebugLevel:  2,
		// 		FolderPaths: []string{`pub_1`, `pub_1_prof_1`},
		// 		DebugData: DebugData{
		// 			HTTPRequestBody: json.RawMessage(`{'invalid_json`),
		// 		},
		// 	},
		// },
		// {
		// 	name: "valid_json",
		// 	fields: fields{
		// 		DebugLevel:  2,
		// 		FolderPaths: []string{`pub_1`, `pub_1_prof_1`},
		// 		DebugData:   DebugData{},
		// 	},
		// },
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
