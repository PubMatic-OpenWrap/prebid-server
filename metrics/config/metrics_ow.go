package config

import (
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prometheus/client_golang/prometheus"
	gometrics "github.com/rcrowley/go-metrics"
)

type RegistryType = string
type MetricsRegistry map[RegistryType]interface{}

const (
	PrometheusRegistry RegistryType = "prometheus"
	InfluxRegistry     RegistryType = "influx"
)

// NewMetricsRegistry returns the map of metrics-engine-name and its respective registry
func NewMetricsRegistry() MetricsRegistry {
	return MetricsRegistry{
		PrometheusRegistry: prometheus.NewRegistry(),
		InfluxRegistry:     gometrics.NewPrefixedRegistry("prebidserver."),
	}
}

func (me *MultiMetricsEngine) RecordVASTTagType(biddder, vastTag string) {
	for _, thisME := range *me {
		thisME.RecordVASTTagType(biddder, vastTag)
	}
}

func (me *MultiMetricsEngine) RecordRejectedBids(pubid, bidder, code string) {
	for _, thisME := range *me {
		thisME.RecordRejectedBids(pubid, bidder, code)
	}
}

func (me *MultiMetricsEngine) RecordBids(pubid, profileid, biddder, deal string) {
	for _, thisME := range *me {
		thisME.RecordBids(pubid, profileid, biddder, deal)
	}
}

func (me *MultiMetricsEngine) RecordVastVersion(biddder, vastVersion string) {
	for _, thisME := range *me {
		thisME.RecordVastVersion(biddder, vastVersion)
	}
}

// RecordRejectedBidsForBidder across all engines
func (me *MultiMetricsEngine) RecordRejectedBidsForBidder(bidder openrtb_ext.BidderName) {
	for _, thisME := range *me {
		thisME.RecordRejectedBidsForBidder(bidder)
	}
}

// RecordFloorStatus across all engines
func (me *MultiMetricsEngine) RecordFloorStatus(pubId, source, code string) {
	for _, thisME := range *me {
		thisME.RecordFloorStatus(pubId, source, code)
	}
}

// RecordVASTTagType as a noop
func (me *NilMetricsEngine) RecordVASTTagType(biddder, vastTag string) {
}

// RecordFloorStatus as a noop
func (me *NilMetricsEngine) RecordFloorStatus(pubId, source, code string) {
}

// RecordRejectedBids as a noop
func (me *NilMetricsEngine) RecordRejectedBids(pubid, bidder, code string) {
}

// RecordBids as a noop
func (me *NilMetricsEngine) RecordBids(pubid, profileid, biddder, deal string) {
}

// RecordVastVersion as a noop
func (me *NilMetricsEngine) RecordVastVersion(biddder, vastVersion string) {
}

// RecordRejectedBidsForBidder as a noop
func (me *NilMetricsEngine) RecordRejectedBidsForBidder(bidder openrtb_ext.BidderName) {
}
