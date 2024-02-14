package pubstack

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prebid/prebid-server/util/task"
)

func NewConfigUpdateHttpTaskPubmatic(httpClient *http.Client, scope, endpoint, refreshInterval string) (*ConfigUpdateHttpTask, error) {
	refreshDuration, err := time.ParseDuration(refreshInterval)
	if err != nil {
		return nil, fmt.Errorf("fail to parse the module args, arg=analytics.pubstack.configuration_refresh_delay: %v", err)
	}

	configChan := make(chan *Configuration)

	tr := task.NewTickerTaskFromFunc(refreshDuration, func() error {
		config := &Configuration{
			ScopeID:  scope,
			Endpoint: endpoint,
			Features: map[string]bool{
				"auction":    true,
				"cookiesync": true,
				"amp":        true,
				"setuid":     true,
				"video":      true,
			},
		}
		configChan <- config
		return nil
	})

	return &ConfigUpdateHttpTask{
		task:       tr,
		configChan: configChan,
	}, nil
}
