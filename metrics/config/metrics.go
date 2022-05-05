package config

import (
	"time"

	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/metrics"
	prometheusmetrics "github.com/prebid/prebid-server/metrics/prometheus"
	"github.com/prebid/prebid-server/openrtb_ext"
	gometrics "github.com/rcrowley/go-metrics"
	influxdb "github.com/vrischmann/go-metrics-influxdb"
)

// NewMetricsEngine reads the configuration and returns the appropriate metrics engine
// for this instance.
func NewMetricsEngine(cfg *config.Configuration, adapterList []openrtb_ext.BidderName, syncerKeys []string) *DetailedMetricsEngine {
	// Create a list of metrics engines to use.
	// Capacity of 2, as unlikely to have more than 2 metrics backends, and in the case
	// of 1 we won't use the list so it will be garbage collected.
	engineList := make(MultiMetricsEngine, 0, 2)
	returnEngine := DetailedMetricsEngine{}

	if cfg.Metrics.Influxdb.Host != "" {
		// Currently use go-metrics as the metrics piece for influx
		returnEngine.GoMetrics = metrics.NewMetrics(gometrics.NewPrefixedRegistry("prebidserver."), adapterList, cfg.Metrics.Disabled, syncerKeys)
		engineList = append(engineList, returnEngine.GoMetrics)

		// Set up the Influx logger
		go influxdb.InfluxDB(
			returnEngine.GoMetrics.MetricsRegistry,                             // metrics registry
			time.Second*time.Duration(cfg.Metrics.Influxdb.MetricSendInterval), // Configurable interval
			cfg.Metrics.Influxdb.Host,                                          // the InfluxDB url
			cfg.Metrics.Influxdb.Database,                                      // your InfluxDB database
			cfg.Metrics.Influxdb.Measurement,                                   // your measurement
			cfg.Metrics.Influxdb.Username,                                      // your InfluxDB user
			cfg.Metrics.Influxdb.Password,                                      // your InfluxDB password,
			cfg.Metrics.Influxdb.AlignTimestamps,                               // align timestamps
		)
		// Influx is not added to the engine list as goMetrics takes care of it already.
	}
	if cfg.Metrics.Prometheus.Port != 0 {
		// Set up the Prometheus metrics.
		returnEngine.PrometheusMetrics = prometheusmetrics.NewMetrics(cfg.Metrics.Prometheus, cfg.Metrics.Disabled, syncerKeys)
		engineList = append(engineList, returnEngine.PrometheusMetrics)
	}

	// Now return the proper metrics engine
	if len(engineList) > 1 {
		returnEngine.MetricsEngine = &engineList
	} else if len(engineList) == 1 {
		returnEngine.MetricsEngine = engineList[0]
	} else {
		returnEngine.MetricsEngine = &NilMetricsEngine{}
	}

	return &returnEngine
}

// DetailedMetricsEngine is a MultiMetricsEngine that preserves links to underlying metrics engines.
type DetailedMetricsEngine struct {
	metrics.MetricsEngine
	GoMetrics         *metrics.Metrics
	PrometheusMetrics *prometheusmetrics.Metrics
}

// MultiMetricsEngine logs metrics to multiple metrics databases The can be useful in transitioning
// an instance from one engine to another, you can run both in parallel to verify stats match up.
type MultiMetricsEngine []metrics.MetricsEngine

// RecordRequest across all engines
func (me *MultiMetricsEngine) RecordRequest(labels metrics.Labels) {
	for _, thisME := range *me {
		thisME.RecordRequest(labels)
	}
}

func (me *MultiMetricsEngine) RecordConnectionAccept(success bool) {
	for _, thisME := range *me {
		thisME.RecordConnectionAccept(success)
	}
}

func (me *MultiMetricsEngine) RecordConnectionClose(success bool) {
	for _, thisME := range *me {
		thisME.RecordConnectionClose(success)
	}
}

//RecordsImps records imps with imp types across all metric engines
func (me *MultiMetricsEngine) RecordImps(implabels metrics.ImpLabels) {
	for _, thisME := range *me {
		thisME.RecordImps(implabels)
	}
}

// RecordRequestTime across all engines
func (me *MultiMetricsEngine) RecordRequestTime(labels metrics.Labels, length time.Duration) {
	for _, thisME := range *me {
		thisME.RecordRequestTime(labels, length)
	}
}

// RecordStoredDataFetchTime across all engines
func (me *MultiMetricsEngine) RecordStoredDataFetchTime(labels metrics.StoredDataLabels, length time.Duration) {
	for _, thisME := range *me {
		thisME.RecordStoredDataFetchTime(labels, length)
	}
}

// RecordStoredDataError across all engines
func (me *MultiMetricsEngine) RecordStoredDataError(labels metrics.StoredDataLabels) {
	for _, thisME := range *me {
		thisME.RecordStoredDataError(labels)
	}
}

// RecordAdapterPanic across all engines
func (me *MultiMetricsEngine) RecordAdapterPanic(labels metrics.AdapterLabels) {
	for _, thisME := range *me {
		thisME.RecordAdapterPanic(labels)
	}
}

// RecordAdapterRequest across all engines
func (me *MultiMetricsEngine) RecordAdapterRequest(labels metrics.AdapterLabels) {
	for _, thisME := range *me {
		thisME.RecordAdapterRequest(labels)
	}
}

// Keeps track of created and reused connections to adapter bidders and the time from the
// connection request, to the connection creation, or reuse from the pool across all engines
func (me *MultiMetricsEngine) RecordAdapterConnections(bidderName openrtb_ext.BidderName, connWasReused bool, connWaitTime time.Duration) {
	for _, thisME := range *me {
		thisME.RecordAdapterConnections(bidderName, connWasReused, connWaitTime)
	}
}

// Times the DNS resolution process
func (me *MultiMetricsEngine) RecordDNSTime(dnsLookupTime time.Duration) {
	for _, thisME := range *me {
		thisME.RecordDNSTime(dnsLookupTime)
	}
}

func (me *MultiMetricsEngine) RecordTLSHandshakeTime(adapterName openrtb_ext.BidderName, tlsHandshakeTime time.Duration) {
	for _, thisME := range *me {
		thisME.RecordTLSHandshakeTime(adapterName, tlsHandshakeTime)
	}
}

// RecordAdapterBidReceived across all engines
func (me *MultiMetricsEngine) RecordAdapterBidReceived(labels metrics.AdapterLabels, bidType openrtb_ext.BidType, hasAdm bool) {
	for _, thisME := range *me {
		thisME.RecordAdapterBidReceived(labels, bidType, hasAdm)
	}
}

// RecordAdapterPrice across all engines
func (me *MultiMetricsEngine) RecordAdapterPrice(labels metrics.AdapterLabels, cpm float64) {
	for _, thisME := range *me {
		thisME.RecordAdapterPrice(labels, cpm)
	}
}

// RecordAdapterTime across all engines
func (me *MultiMetricsEngine) RecordAdapterTime(labels metrics.AdapterLabels, length time.Duration) {
	for _, thisME := range *me {
		thisME.RecordAdapterTime(labels, length)
	}
}

// RecordCookieSync across all engines
func (me *MultiMetricsEngine) RecordCookieSync(status metrics.CookieSyncStatus) {
	for _, thisME := range *me {
		thisME.RecordCookieSync(status)
	}
}

// RecordSyncerRequest across all engines
func (me *MultiMetricsEngine) RecordSyncerRequest(key string, status metrics.SyncerCookieSyncStatus) {
	for _, thisME := range *me {
		thisME.RecordSyncerRequest(key, status)
	}
}

// RecordSetUid across all engines
func (me *MultiMetricsEngine) RecordSetUid(status metrics.SetUidStatus) {
	for _, thisME := range *me {
		thisME.RecordSetUid(status)
	}
}

// RecordSyncerSet across all engines
func (me *MultiMetricsEngine) RecordSyncerSet(key string, status metrics.SyncerSetUidStatus) {
	for _, thisME := range *me {
		thisME.RecordSyncerSet(key, status)
	}
}

// RecordStoredReqCacheResult across all engines
func (me *MultiMetricsEngine) RecordStoredReqCacheResult(cacheResult metrics.CacheResult, inc int) {
	for _, thisME := range *me {
		thisME.RecordStoredReqCacheResult(cacheResult, inc)
	}
}

// RecordStoredImpCacheResult across all engines
func (me *MultiMetricsEngine) RecordStoredImpCacheResult(cacheResult metrics.CacheResult, inc int) {
	for _, thisME := range *me {
		thisME.RecordStoredImpCacheResult(cacheResult, inc)
	}
}

// RecordAccountCacheResult across all engines
func (me *MultiMetricsEngine) RecordAccountCacheResult(cacheResult metrics.CacheResult, inc int) {
	for _, thisME := range *me {
		thisME.RecordAccountCacheResult(cacheResult, inc)
	}
}

// RecordPrebidCacheRequestTime across all engines
func (me *MultiMetricsEngine) RecordPrebidCacheRequestTime(success bool, length time.Duration) {
	for _, thisME := range *me {
		thisME.RecordPrebidCacheRequestTime(success, length)
	}
}

// RecordRequestQueueTime across all engines
func (me *MultiMetricsEngine) RecordRequestQueueTime(success bool, requestType metrics.RequestType, length time.Duration) {
	for _, thisME := range *me {
		thisME.RecordRequestQueueTime(success, requestType, length)
	}
}

// RecordTimeoutNotice across all engines
func (me *MultiMetricsEngine) RecordTimeoutNotice(success bool) {
	for _, thisME := range *me {
		thisME.RecordTimeoutNotice(success)
	}
}

// RecordRequestPrivacy across all engines
func (me *MultiMetricsEngine) RecordRequestPrivacy(privacy metrics.PrivacyLabels) {
	for _, thisME := range *me {
		thisME.RecordRequestPrivacy(privacy)
	}
}

// RecordAdapterDuplicateBidID across all engines
func (me *MultiMetricsEngine) RecordAdapterDuplicateBidID(adaptor string, collisions int) {
	for _, thisME := range *me {
		thisME.RecordAdapterDuplicateBidID(adaptor, collisions)
	}
}

// RecordRequestHavingDuplicateBidID across all engines
func (me *MultiMetricsEngine) RecordRequestHavingDuplicateBidID() {
	for _, thisME := range *me {
		thisME.RecordRequestHavingDuplicateBidID()
	}
}

// RecordPodImpGenTime across all engines
func (me *MultiMetricsEngine) RecordPodImpGenTime(labels metrics.PodLabels, startTime time.Time) {
	for _, thisME := range *me {
		thisME.RecordPodImpGenTime(labels, startTime)
	}
}

// RecordPodCombGenTime as a noop
func (me *MultiMetricsEngine) RecordPodCombGenTime(labels metrics.PodLabels, elapsedTime time.Duration) {
	for _, thisME := range *me {
		thisME.RecordPodCombGenTime(labels, elapsedTime)
	}
}

// RecordPodCompititveExclusionTime as a noop
func (me *MultiMetricsEngine) RecordPodCompititveExclusionTime(labels metrics.PodLabels, elapsedTime time.Duration) {
	for _, thisME := range *me {
		thisME.RecordPodCompititveExclusionTime(labels, elapsedTime)
	}
}

// RecordAdapterVideoBidDuration as a noop
func (me *MultiMetricsEngine) RecordAdapterVideoBidDuration(labels metrics.AdapterLabels, videoBidDuration int) {
	for _, thisME := range *me {
		thisME.RecordAdapterVideoBidDuration(labels, videoBidDuration)
	}
}

// RecordAdapterGDPRRequestBlocked across all engines
func (me *MultiMetricsEngine) RecordAdapterGDPRRequestBlocked(adapter openrtb_ext.BidderName) {
	for _, thisME := range *me {
		thisME.RecordAdapterGDPRRequestBlocked(adapter)
	}
}

// NilMetricsEngine implements the MetricsEngine interface where no metrics are actually captured. This is
// used if no metric backend is configured and also for tests.
type NilMetricsEngine struct{}

func (me *NilMetricsEngine) RecordAdapterDuplicateBidID(adaptor string, collisions int) {
}

func (me *NilMetricsEngine) RecordRequestHavingDuplicateBidID() {
}

func (me *NilMetricsEngine) RecordPodImpGenTime(labels metrics.PodLabels, startTime time.Time) {
}

func (me *NilMetricsEngine) RecordPodCombGenTime(labels metrics.PodLabels, elapsedTime time.Duration) {
}

func (me *NilMetricsEngine) RecordPodCompititveExclusionTime(labels metrics.PodLabels, elapsedTime time.Duration) {
}

func (me *NilMetricsEngine) RecordAdapterVideoBidDuration(labels metrics.AdapterLabels, videoBidDuration int) {
}

// RecordRequest as a noop
func (me *NilMetricsEngine) RecordRequest(labels metrics.Labels) {
}

// RecordConnectionAccept as a noop
func (me *NilMetricsEngine) RecordConnectionAccept(success bool) {
}

// RecordConnectionClose as a noop
func (me *NilMetricsEngine) RecordConnectionClose(success bool) {
}

// RecordImps as a noop
func (me *NilMetricsEngine) RecordImps(implabels metrics.ImpLabels) {
}

// RecordRequestTime as a noop
func (me *NilMetricsEngine) RecordRequestTime(labels metrics.Labels, length time.Duration) {
}

// RecordStoredDataFetchTime as a noop
func (me *NilMetricsEngine) RecordStoredDataFetchTime(labels metrics.StoredDataLabels, length time.Duration) {
}

// RecordStoredDataError as a noop
func (me *NilMetricsEngine) RecordStoredDataError(labels metrics.StoredDataLabels) {
}

// RecordAdapterPanic as a noop
func (me *NilMetricsEngine) RecordAdapterPanic(labels metrics.AdapterLabels) {
}

// RecordAdapterRequest as a noop
func (me *NilMetricsEngine) RecordAdapterRequest(labels metrics.AdapterLabels) {
}

// RecordAdapterConnections as a noop
func (me *NilMetricsEngine) RecordAdapterConnections(bidderName openrtb_ext.BidderName, connWasReused bool, connWaitTime time.Duration) {
}

// RecordDNSTime as a noop
func (me *NilMetricsEngine) RecordDNSTime(dnsLookupTime time.Duration) {
}

// RecordTLSHandshakeTime as a noop
func (me *NilMetricsEngine) RecordTLSHandshakeTime(adapterName openrtb_ext.BidderName, tlsHandshakeTime time.Duration) {
}

// RecordAdapterBidReceived as a noop
func (me *NilMetricsEngine) RecordAdapterBidReceived(labels metrics.AdapterLabels, bidType openrtb_ext.BidType, hasAdm bool) {
}

// RecordAdapterPrice as a noop
func (me *NilMetricsEngine) RecordAdapterPrice(labels metrics.AdapterLabels, cpm float64) {
}

// RecordAdapterTime as a noop
func (me *NilMetricsEngine) RecordAdapterTime(labels metrics.AdapterLabels, length time.Duration) {
}

// RecordCookieSync as a noop
func (me *NilMetricsEngine) RecordCookieSync(status metrics.CookieSyncStatus) {
}

// RecordSyncerRequest as a noop
func (me *NilMetricsEngine) RecordSyncerRequest(key string, status metrics.SyncerCookieSyncStatus) {
}

// RecordSetUid as a noop
func (me *NilMetricsEngine) RecordSetUid(status metrics.SetUidStatus) {
}

// RecordSyncerSet as a noop
func (me *NilMetricsEngine) RecordSyncerSet(key string, status metrics.SyncerSetUidStatus) {
}

// RecordStoredReqCacheResult as a noop
func (me *NilMetricsEngine) RecordStoredReqCacheResult(cacheResult metrics.CacheResult, inc int) {
}

// RecordStoredImpCacheResult as a noop
func (me *NilMetricsEngine) RecordStoredImpCacheResult(cacheResult metrics.CacheResult, inc int) {
}

// RecordAccountCacheResult as a noop
func (me *NilMetricsEngine) RecordAccountCacheResult(cacheResult metrics.CacheResult, inc int) {
}

// RecordPrebidCacheRequestTime as a noop
func (me *NilMetricsEngine) RecordPrebidCacheRequestTime(success bool, length time.Duration) {
}

// RecordRequestQueueTime as a noop
func (me *NilMetricsEngine) RecordRequestQueueTime(success bool, requestType metrics.RequestType, length time.Duration) {
}

// RecordTimeoutNotice as a noop
func (me *NilMetricsEngine) RecordTimeoutNotice(success bool) {
}

// RecordRequestPrivacy as a noop
func (me *NilMetricsEngine) RecordRequestPrivacy(privacy metrics.PrivacyLabels) {
}

// RecordAdapterGDPRRequestBlocked as a noop
func (me *NilMetricsEngine) RecordAdapterGDPRRequestBlocked(adapter openrtb_ext.BidderName) {
}
