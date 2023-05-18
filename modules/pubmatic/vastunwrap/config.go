package vastunwrap

import (
	"encoding/json"

	unWrapCfg "git.pubmatic.com/vastunwrap/config"
)

// VastUnwrapModuleCfg contains the values read from the config file  for vast unwrapper module at boot time
type VastUnwrapModuleCfg struct {
	VastUnWrapCfg unWrapCfg.VastUnWrapCfg
}

func (cfg *VastUnwrapModuleCfg) String() string {
	jsonBytes, err := json.Marshal(cfg)

	if nil != err {
		return err.Error()
	}

	return string(jsonBytes[:])
}
