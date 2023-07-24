package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

type Payload struct {
	IR json.RawMessage
	AC json.RawMessage
}

func main() {
	http.HandleFunc("/verifyauc", verifyauc)
	http.ListenAndServe(":8181", nil)
}

func verifyauc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("failed to read " + err.Error()))
		w.WriteHeader(500)
		return
	}

	payload := &Payload{}
	if err := json.Unmarshal(body, payload); err != nil {
		w.Write([]byte("failed to unmarshal request " + err.Error()))
		w.WriteHeader(400)
		return
	}

	newAdunitConfigObj := new(adunitconfig.AdUnitConfig)
	if err := json.Unmarshal(payload.AC, &newAdunitConfigObj); err != nil || newAdunitConfigObj == nil || len(newAdunitConfigObj.Config) == 0 {
		if err != nil {
			w.Write([]byte("failed to unmarshal auc " + err.Error()))
		} else {
			w.Write([]byte("failed to unmarshal auc."))
		}

		w.WriteHeader(400)
		return
	}

	defaultAUC, ok := newAdunitConfigObj.Config["default"]
	if !ok || defaultAUC == nil || defaultAUC.Video == nil {
		json.NewEncoder(w).Encode(newAdunitConfigObj)
		return
	}

	// default disable is disable for all
	if defaultAUC.Video.Enabled != nil && !*defaultAUC.Video.Enabled {
		for tag, auc := range newAdunitConfigObj.Config {
			auc.Video = nil
			newAdunitConfigObj.Config[tag] = auc
		}
	}

	if defaultAUC.Video.Config != nil {
		for tag, auc := range newAdunitConfigObj.Config {
			if tag == "default" {
				continue
			}

			if auc.Video == nil {
				auc.Video = defaultAUC.Video
				newAdunitConfigObj.Config[tag] = auc
				continue
			}

			if auc.Video.Enabled != nil && !*auc.Video.Enabled {
				auc.Video = nil
				newAdunitConfigObj.Config[tag] = auc
				continue
			}

			if auc.Video.Config == nil {
				auc.Video.Config = defaultAUC.Video.Config
				newAdunitConfigObj.Config[tag] = auc
				continue
			}

			if auc.Video.Config.MinDuration == nil {
				auc.Video.Config.MinDuration = defaultAUC.Video.Config.MinDuration
			}

			newAdunitConfigObj.Config[tag] = auc
		}
	}

	json.NewEncoder(w).Encode(newAdunitConfigObj)
}
