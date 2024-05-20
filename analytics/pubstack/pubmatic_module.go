package pubstack

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/benbjohnson/clock"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/analytics"
	"github.com/prebid/prebid-server/v2/analytics/pubstack/eventchannel"
)

func NewModulePubmatic(client *http.Client, scope, endpoint, configRefreshDelay string, maxEventCount int, maxByteSize, maxTime string, clock clock.Clock) (analytics.Module, error) {

	return NewModuleWithConfigTaskPubmatic(client, scope, endpoint, maxEventCount, maxByteSize, maxTime, clock)
}

func NewModuleWithConfigTaskPubmatic(client *http.Client, scope, endpoint string, maxEventCount int, maxByteSize, maxTime string, clock clock.Clock) (analytics.Module, error) {
	glog.Infof("[pubstack] Initializing module scope=%s endpoint=%s\n", scope, endpoint)

	// parse args
	bufferCfg, err := newBufferConfig(maxEventCount, maxByteSize, maxTime)
	if err != nil {
		return nil, fmt.Errorf("fail to parse the module args, arg=analytics.pubstack.buffers, :%v", err)
	}

	defaultFeatures := map[string]bool{
		auction:    true,
		video:      true,
		amp:        true,
		cookieSync: true,
		setUID:     true,
	}

	defaultConfig := &Configuration{
		ScopeID:  scope,
		Endpoint: endpoint,
		Features: defaultFeatures,
	}

	pb := PubstackModule{
		scope:         scope,
		httpClient:    client,
		cfg:           defaultConfig,
		buffsCfg:      bufferCfg,
		sigTermCh:     make(chan os.Signal),
		stopCh:        make(chan struct{}),
		eventChannels: make(map[string]*eventchannel.EventChannel),
		muxConfig:     sync.RWMutex{},
		clock:         clock,
	}

	pb.registerChannel(auction)

	return &pb, nil
}
