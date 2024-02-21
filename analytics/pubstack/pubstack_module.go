package pubstack

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/analytics/pubmatic"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/analytics/pubstack/eventchannel"
	"github.com/prebid/prebid-server/analytics/pubstack/helpers"
)

type Configuration struct {
	ScopeID  string          `json:"scopeId"`
	Endpoint string          `json:"endpoint"`
	Features map[string]bool `json:"features"`
}

// routes for events
const (
	auction    = "auction"
	cookieSync = "cookiesync"
	amp        = "amp"
	setUID     = "setuid"
	video      = "video"
)

type bufferConfig struct {
	timeout time.Duration
	count   int64
	size    int64
}

type PubstackModule struct {
	eventChannels map[string]*eventchannel.EventChannel
	httpClient    *http.Client
	sigTermCh     chan os.Signal
	stopCh        chan struct{}
	scope         string
	cfg           *Configuration
	buffsCfg      *bufferConfig
	muxConfig     sync.RWMutex
	clock         clock.Clock
}

func NewModule(client *http.Client, scope, endpoint, configRefreshDelay string, maxEventCount int, maxByteSize, maxTime string, clock clock.Clock) (analytics.PBSAnalyticsModule, error) {
	configUpdateTask, err := NewConfigUpdateHttpTask(
		client,
		scope,
		endpoint,
		configRefreshDelay)
	if err != nil {
		return nil, err
	}

	return NewModuleWithConfigTask(client, scope, endpoint, maxEventCount, maxByteSize, maxTime, configUpdateTask, clock)
}

func NewModuleWithConfigTask(client *http.Client, scope, endpoint string, maxEventCount int, maxByteSize, maxTime string, configTask ConfigUpdateTask, clock clock.Clock) (analytics.PBSAnalyticsModule, error) {
	glog.Infof("[pubstack] Initializing module scope=%s endpoint=%s\n", scope, endpoint)

	// parse args
	bufferCfg, err := newBufferConfig(maxEventCount, maxByteSize, maxTime)
	if err != nil {
		return nil, fmt.Errorf("fail to parse the module args, arg=analytics.pubstack.buffers, :%v", err)
	}

	defaultFeatures := map[string]bool{
		auction:    false,
		video:      false,
		amp:        false,
		cookieSync: false,
		setUID:     false,
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

	signal.Notify(pb.sigTermCh, os.Interrupt, syscall.SIGTERM)

	configChannel := configTask.Start(pb.stopCh)
	go pb.start(configChannel)

	glog.Info("[pubstack] Pubstack analytics configured and ready")
	return &pb, nil
}

func (p *PubstackModule) LogAuctionObject(ao *analytics.AuctionObject) {
	p.muxConfig.RLock()
	defer p.muxConfig.RUnlock()

	if !p.isFeatureEnable(auction) {
		return
	}

	var rCtx *models.RequestCtx
	defer func() {
		if r := recover(); r != nil {
			if rCtx != nil {
				glog.Errorf("stacktrace:[%s], error:[%v], pubid:[%d], profid:[%d], ver:[%d]", string(debug.Stack()), r, rCtx.PubID, rCtx.ProfileID, rCtx.VersionID)
				return
			}
			glog.Errorf("stacktrace:[%s], error:[%v]", string(debug.Stack()), r)
		}
	}()

	rCtx = pubmatic.GetRequestCtx(ao.HookExecutionOutcome)
	if rCtx == nil {
		// glog.Errorf("Failed to get the request context for AuctionObject - [%v]", ao)
		// add this log once complete header-bidding code is migrated to modules
		return
	}

	url, _ := pubmatic.GetLogAuctionObjectAsURL(*ao, rCtx, false, false)
	if url == "" {
		glog.Errorf("Failed to prepare the owlogger for pub:[%d], profile:[%d], version:[%d].",
			rCtx.PubID, rCtx.ProfileID, rCtx.VersionID)
		return
	}

	url = strings.TrimPrefix(url, "http://10.172.141.11/wl?")

	payload := []byte(url)

	payload = append(payload, byte('\n'))

	p.eventChannels[auction].Push(payload)
}

func (p *PubstackModule) LogNotificationEventObject(ne *analytics.NotificationEvent) {
}

func (p *PubstackModule) LogVideoObject(vo *analytics.VideoObject) {
	p.muxConfig.RLock()
	defer p.muxConfig.RUnlock()

	if !p.isFeatureEnable(video) {
		return
	}

	// serialize event
	payload, err := helpers.JsonifyVideoObject(vo, p.scope)
	if err != nil {
		glog.Warning("[pubstack] Cannot serialize video")
		return
	}

	p.eventChannels[video].Push(payload)
}

func (p *PubstackModule) LogSetUIDObject(so *analytics.SetUIDObject) {
	p.muxConfig.RLock()
	defer p.muxConfig.RUnlock()

	if !p.isFeatureEnable(setUID) {
		return
	}

	// serialize event
	payload, err := helpers.JsonifySetUIDObject(so, p.scope)
	if err != nil {
		glog.Warning("[pubstack] Cannot serialize video")
		return
	}

	p.eventChannels[setUID].Push(payload)
}

func (p *PubstackModule) LogCookieSyncObject(cso *analytics.CookieSyncObject) {
	p.muxConfig.RLock()
	defer p.muxConfig.RUnlock()

	if !p.isFeatureEnable(cookieSync) {
		return
	}

	// serialize event
	payload, err := helpers.JsonifyCookieSync(cso, p.scope)
	if err != nil {
		glog.Warning("[pubstack] Cannot serialize video")
		return
	}

	p.eventChannels[cookieSync].Push(payload)
}

func (p *PubstackModule) LogAmpObject(ao *analytics.AmpObject) {
	p.muxConfig.RLock()
	defer p.muxConfig.RUnlock()

	if !p.isFeatureEnable(amp) {
		return
	}

	// serialize event
	payload, err := helpers.JsonifyAmpObject(ao, p.scope)
	if err != nil {
		glog.Warning("[pubstack] Cannot serialize video")
		return
	}

	p.eventChannels[amp].Push(payload)
}

func (p *PubstackModule) start(c <-chan *Configuration) {
	for {
		select {
		case <-p.sigTermCh:
			close(p.stopCh)
			cfg := p.cfg.clone().disableAllFeatures()
			p.updateConfig(cfg)
			return
		case config := <-c:
			p.updateConfig(config)
			glog.Infof("[pubstack] Updating config: %v", p.cfg)
		}
	}
}

func (p *PubstackModule) updateConfig(config *Configuration) {
	p.muxConfig.Lock()
	defer p.muxConfig.Unlock()

	if p.cfg.isSameAs(config) {
		return
	}

	p.cfg = config
	p.closeAllEventChannels()

	p.registerChannel(amp)
	p.registerChannel(auction)
	p.registerChannel(cookieSync)
	p.registerChannel(video)
	p.registerChannel(setUID)
}

func (p *PubstackModule) isFeatureEnable(feature string) bool {
	val, ok := p.cfg.Features[feature]
	return ok && val
}

func (p *PubstackModule) registerChannel(feature string) {
	if p.isFeatureEnable(feature) {
		sender := eventchannel.BuildEndpointSender(p.httpClient, p.cfg.Endpoint, feature)
		p.eventChannels[feature] = eventchannel.NewEventChannel(sender, p.clock, p.buffsCfg.size, p.buffsCfg.count, p.buffsCfg.timeout)
	}
}

func (p *PubstackModule) closeAllEventChannels() {
	for key, ch := range p.eventChannels {
		ch.Close()
		delete(p.eventChannels, key)
	}
}
