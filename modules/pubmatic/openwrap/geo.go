package openwrap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"

	"git.pubmatic.com/PubMatic/go-common/util"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/prometheus"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// geo provides geo metadata from ip
type geo struct {
	CountryCode string `json:"cc,omitempty"`
	StateCode   string `json:"sc,omitempty"`
	Compliance  string `json:"compliance,omitempty"`
	SectionID   int    `json:"sectionId,omitempty"`
}

var gppSectionIDs = map[string]int{
	"ca": 8,
	"va": 9,
	"co": 10,
	"ut": 11,
	"ct": 12,
}

// Handler for /geo endpoint
func Handler(w http.ResponseWriter, r *http.Request) {
	var pubIdStr string
	metricEngine := ow.GetMetricEngine()
	metricLabels := metrics.Labels{RType: models.EndpointGeo, RequestStatus: prometheus.RequestStatusOK}
	defer func() {
		metricEngine.RecordRequest(metricLabels)
		if r := recover(); r != nil {
			metricEngine.RecordOpenWrapServerPanicStats(ow.cfg.Server.HostName, "HandleGeoEndpoint")
			glog.Errorf("stacktrace:[%s], error:[%v], pubid:[%s]", string(debug.Stack()), r, pubIdStr)
			return
		}
	}()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		metricLabels.RequestStatus = prometheus.RequestStatusBadInput
		return
	}
	pubIdStr = r.URL.Query().Get(models.PublisherID)
	_, err := strconv.Atoi(pubIdStr)
	if err != nil {
		glog.Errorf("[geo] error:[invalid pubid passed:%s], [requestType]:%v [url]:%v [origin]:%v [referer]:%v", err.Error(), models.EndpointGeo,
			r.URL.RequestURI(), r.Header.Get(models.HeaderOriginKey), r.Header.Get(models.HeaderRefererKey))
		w.WriteHeader(http.StatusBadRequest)
		metricLabels.RequestStatus = prometheus.RequestStatusBadInput
		return
	}

	ip := util.GetIP(r)
	w.Header().Set(models.HeaderContentType, models.HeaderContentTypeValue)
	w.Header().Set(models.HeaderAccessControlAllowOrigin, "*")

	geoInfo, _ := ow.geoInfoFetcher.LookUp(ip)
	if geoInfo == nil || geoInfo.ISOCountryCode == "" {
		metricEngine.RecordGeoLookupFailure(models.EndpointGeo)
		return
	}

	geo := geo{
		CountryCode: geoInfo.ISOCountryCode,
		StateCode:   geoInfo.RegionCode,
	}
	if ow.GetFeature().IsCountryGDPREnabled(geo.CountryCode) {
		geo.Compliance = models.GDPRCompliance
	} else if geo.CountryCode == models.CountryCodeUS && geo.StateCode == models.StateCodeCalifornia {
		geo.Compliance = models.USPCompliance
	} else if sectionid, ok := gppSectionIDs[geo.StateCode]; ok {
		geo.Compliance = models.GPPCompliance
		geo.SectionID = sectionid
	}

	w.Header().Set(models.HeaderCacheControl, "max-age="+fmt.Sprint(models.CacheTimeout.Seconds()))
	json.NewEncoder(w).Encode(geo)
}
