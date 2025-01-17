package router

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/analytics"
	"github.com/prebid/prebid-server/v2/currency"
	"github.com/prebid/prebid-server/v2/hooks"
	"github.com/prebid/prebid-server/v2/ortb"

	analyticsBuild "github.com/prebid/prebid-server/v2/analytics/build"
	"github.com/prebid/prebid-server/v2/config"
	"github.com/prebid/prebid-server/v2/endpoints"
	"github.com/prebid/prebid-server/v2/endpoints/openrtb2"
	"github.com/prebid/prebid-server/v2/exchange"
	"github.com/prebid/prebid-server/v2/gdpr"
	"github.com/prebid/prebid-server/v2/metrics"
	metricsConf "github.com/prebid/prebid-server/v2/metrics/config"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	pbc "github.com/prebid/prebid-server/v2/prebid_cache_client"
	"github.com/prebid/prebid-server/v2/stored_requests"
	"github.com/prebid/prebid-server/v2/usersync"
	"github.com/prebid/prebid-server/v2/util/uuidutil"
	"github.com/prometheus/client_golang/prometheus"
)

// TODO: Delete router_sshb.go usage after PBS-OpenWrap module

var (
	g_syncers             map[string]usersync.Syncer
	g_cfg                 *config.Configuration
	g_ex                  *exchange.Exchange
	g_accounts            *stored_requests.AccountFetcher
	g_requestValidator    *ortb.RequestValidator
	g_storedReqFetcher    *stored_requests.Fetcher
	g_storedRespFetcher   *stored_requests.Fetcher
	g_metrics             metrics.MetricsEngine
	g_analytics           *analytics.Runner
	g_disabledBidders     map[string]string
	g_videoFetcher        *stored_requests.Fetcher
	g_activeBidders       map[string]openrtb_ext.BidderName
	g_defReqJSON          []byte
	g_cacheClient         *pbc.Client
	g_gdprPermsBuilder    gdpr.PermissionsBuilder
	g_tcf2CfgBuilder      gdpr.TCF2ConfigBuilder
	g_planBuilder         *hooks.ExecutionPlanBuilder
	g_currencyConversions currency.Conversions
	g_tmaxAdjustments     *exchange.TmaxAdjustmentsPreprocessed
)

func GetCacheClient() *pbc.Client {
	return g_cacheClient
}

func GetPrebidCacheURL() string {
	return g_cfg.ExternalURL
}

// RegisterAnalyticsModule function registers the PBSAnalyticsModule
func RegisterAnalyticsModule(anlt analytics.Module) error {
	if g_analytics == nil {
		return fmt.Errorf("g_analytics is nil")
	}
	modules, err := analyticsBuild.EnableAnalyticsModule(anlt, *g_analytics)
	if err != nil {
		return err
	}
	g_analytics = &modules
	return nil
}

// OrtbAuctionEndpointWrapper Openwrap wrapper method for calling /openrtb2/auction endpoint
func OrtbAuctionEndpointWrapper(w http.ResponseWriter, r *http.Request) error {
	ortbAuctionEndpoint, err := openrtb2.NewEndpoint(uuidutil.UUIDRandomGenerator{}, *g_ex, *g_requestValidator, *g_storedReqFetcher, *g_accounts, g_cfg, g_metrics, *g_analytics, g_disabledBidders, g_defReqJSON, g_activeBidders, *g_storedRespFetcher, *g_planBuilder, g_tmaxAdjustments)
	if err != nil {
		return err
	}
	ortbAuctionEndpoint(w, r, nil)
	return nil
}

// GetPBSCurrencyRate Openwrap wrapper method for currency conversion
func GetPBSCurrencyConversion(from, to string, value float64) (float64, error) {
	rate, err := g_currencyConversions.GetRate(from, to)
	if err == nil {
		return value * rate, nil
	}
	return 0, err
}

// VideoAuctionEndpointWrapper Openwrap wrapper method for calling /openrtb2/video endpoint
func VideoAuctionEndpointWrapper(w http.ResponseWriter, r *http.Request) error {
	videoAuctionEndpoint, err := openrtb2.NewCTVEndpoint(*g_ex, *g_requestValidator, *g_storedReqFetcher, *g_videoFetcher, *g_accounts, g_cfg, g_metrics, *g_analytics, g_disabledBidders, g_defReqJSON, g_activeBidders, *g_planBuilder, g_tmaxAdjustments)
	if err != nil {
		return err
	}
	videoAuctionEndpoint(w, r, nil)
	return nil
}

// GetUIDSWrapper Openwrap wrapper method for calling /getuids endpoint
func GetUIDSWrapper(w http.ResponseWriter, r *http.Request) {
	getUID := endpoints.NewGetUIDsEndpoint(g_cfg.HostCookie)
	getUID(w, r, nil)
}

// SetUIDSWrapper Openwrap wrapper method for calling /setuid endpoint
func SetUIDSWrapper(w http.ResponseWriter, r *http.Request) {
	setUID := endpoints.NewSetUIDEndpoint(g_cfg, g_syncers, g_gdprPermsBuilder, g_tcf2CfgBuilder, *g_analytics, *g_accounts, g_metrics)
	setUID(w, r, nil)
}

// CookieSync Openwrap wrapper method for calling /cookie_sync endpoint
func CookieSync(w http.ResponseWriter, r *http.Request) {
	cookiesync := endpoints.NewCookieSyncEndpoint(g_syncers, g_cfg, g_gdprPermsBuilder, g_tcf2CfgBuilder, g_metrics, *g_analytics, *g_accounts, g_activeBidders)
	cookiesync.Handle(w, r, nil)
}

// SyncerMap Returns map of bidder and its usersync info
func SyncerMap() map[string]usersync.Syncer {
	return g_syncers
}

func GetPrometheusGatherer() *prometheus.Registry {
	mEngine, ok := g_metrics.(*metricsConf.DetailedMetricsEngine)
	if !ok || mEngine == nil || mEngine.PrometheusMetrics == nil {
		return nil
	}

	return mEngine.PrometheusMetrics.Gatherer
}

// CallRecordNonBids calls RecordRejectedBids function on prebid's metric-engine
func CallRecordNonBids(pubId, bidder string, code openrtb3.NoBidReason) {
	if g_metrics != nil {
		codeStr := strconv.FormatInt(int64(code), 10)
		g_metrics.RecordRejectedBids(pubId, bidder, codeStr)
	}
}
