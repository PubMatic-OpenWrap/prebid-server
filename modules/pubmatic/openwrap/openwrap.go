package openwrap

import (
	"encoding/json"
	"fmt"

	vastunwrap "git.pubmatic.com/vastunwrap"
	"github.com/prebid/prebid-server/modules/moduledeps"
	ow_config "github.com/prebid/prebid-server/modules/pubmatic/openwrap/config"
)

type OpenWrap struct {
	cfg ow_config.SSHB
}

func initOpenWrap(rawCfg json.RawMessage, _ moduledeps.ModuleDeps) (OpenWrap, error) {
	cfg := ow_config.SSHB{}

	err := json.Unmarshal(rawCfg, &cfg)
	if err != nil {
		return OpenWrap{}, fmt.Errorf("invalid openwrap config: %v", err)
	}

	vastunwrap.InitUnWrapperConfig(cfg.OpenWrap.Vastunwrap)

	return OpenWrap{
		cfg: cfg,
	}, nil

}
