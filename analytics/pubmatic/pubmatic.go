package pubmatic

import (
	"runtime/debug"
	"sync"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

type RequestType string

const (
	COOKIE_SYNC        RequestType = "/cookie_sync"
	AUCTION            RequestType = "/openrtb2/auction"
	VIDEO              RequestType = "/openrtb2/video"
	SETUID             RequestType = "/set_uid"
	AMP                RequestType = "/openrtb2/amp"
	NOTIFICATION_EVENT RequestType = "/event"
)

var ow HTTPLogger
var once sync.Once

// Module that can perform transactional logging
type HTTPLogger struct {
	cfg config.PubMaticWL
}

// LogAuctionObject prepares the owlogger url and send it to logger endpoint
func (ow HTTPLogger) LogAuctionObject(ao *analytics.AuctionObject) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error(string(debug.Stack()))
		}
	}()

	rCtx := GetRequestCtx(ao.HookExecutionOutcome)
	if rCtx == nil {
		// glog.Errorf("Failed to get the request context for AuctionObject - [%v]", ao)
		// add this log once complete header-bidding code is migrated to modules
		return
	}

	var err error
	url, headers := GetLogAuctionObjectAsURL(*ao, rCtx, false, false)
	if url != "" {
		cookies := make(map[string]string)
		if rCtx.KADUSERCookie != nil {
			cookies[models.KADUSERCOOKIE] = rCtx.KADUSERCookie.Value
		}
		err = Send(url, headers, cookies)
	}

	// record the logger failure
	if url == "" || err != nil {
		glog.Errorf("Failed to send the owlogger for pub:[%d], profile:[%d], version:[%d].",
			rCtx.PubID, rCtx.ProfileID, rCtx.VersionID)
	}
}

// Writes VideoObject to file
func (ow HTTPLogger) LogVideoObject(vo *analytics.VideoObject) {
}

// Logs SetUIDObject to file
func (ow HTTPLogger) LogSetUIDObject(so *analytics.SetUIDObject) {
}

// Logs CookieSyncObject to file
func (ow HTTPLogger) LogCookieSyncObject(cso *analytics.CookieSyncObject) {
}

// Logs AmpObject to file
func (ow HTTPLogger) LogAmpObject(ao *analytics.AmpObject) {
}

// Logs NotificationEvent to file
func (ow HTTPLogger) LogNotificationEventObject(ne *analytics.NotificationEvent) {
}

// Method to initialize the analytic module
func NewHTTPLogger(cfg config.PubMaticWL) analytics.PBSAnalyticsModule {
	once.Do(func() {
		Init(cfg.MaxClients, cfg.MaxConnections, cfg.MaxCalls, cfg.RespTimeout)

		ow = HTTPLogger{
			cfg: cfg,
		}
	})

	return ow
}
