package openwrap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"git.pubmatic.com/PubMatic/go-common/util"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb/netacuity"
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

const (
	cacheTimeout                        = time.Duration(48) * time.Hour
	headerContentType                   = "Content-Type"
	headerAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	headerCacheControl                  = "Cache-Control"
	headerContentTypeValue              = "application/json"
	headerAccessControlAllowOriginValue = "*"
)

// geoHandler provides a handler for geo lookups.
type geoHandler struct {
	geoService netacuity.NetAcuity
}

// NewGeoHandler initializes and returns a new GeoHandler.
func NewGeoHandler() *geoHandler {
	return &geoHandler{
		geoService: netacuity.NetAcuity{},
	}
}

var gppSectionIDs = map[string]int{
	"ca": 8,
	"va": 9,
	"co": 10,
	"ut": 11,
	"ct": 12,
}

const (
	OriginHeaderKey     = "Origin"
	RefererKey          = "Referer"
	GDPRCompliance      = "GDPR"
	USPCompliance       = "USP"
	GPPCompliance       = "GPP"
	StateCodeCalifornia = "ca"
	CountryCodeUS       = "US"
)

// Handler for /geo endpoint
func (handler *geoHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var pubIdStr string
	metricEngine := ow.GetMetricEngine()
	defer func() {
		if r := recover(); r != nil {
			metricEngine.RecordOpenWrapServerPanicStats(ow.cfg.Server.HostName, "HandleGeoEndpoint")
			glog.Errorf("stacktrace:[%s], error:[%v], pubid:[%s]", string(debug.Stack()), r, pubIdStr)
			return
		}
	}()
	metricEngine.RecordRequest(metrics.Labels{RType: models.EndpointGeo, RequestStatus: prometheus.RequestStatusOK})

	pubIdStr = r.URL.Query().Get(models.PublisherID)
	_, err := strconv.Atoi(pubIdStr)
	if err != nil {
		glog.Errorf("[geo] error:[invalid pubid passed:%s], [requestType]:%v [url]:%v [origin]:%v [referer]:%v", err.Error(), models.EndpointGeo,
			r.URL.RequestURI(), r.Header.Get(OriginHeaderKey), r.Header.Get(RefererKey))

		//TO-Do keep this stat?
		metricEngine.RecordBadRequests(models.EndpointGeo, pubIdStr, -1)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	metricEngine.RecordPublisherRequests(models.EndpointGeo, pubIdStr, "")

	ip := util.GetIP(r)
	w.Header().Set(headerContentType, headerContentTypeValue)
	w.Header().Set(headerAccessControlAllowOrigin, "*")
	success := false
	geoInfo, _ := handler.geoService.LookUp(ip)
	if geoInfo != nil {
		if geoInfo.ISOCountryCode != "" {
			success = true
			geo := geo{
				CountryCode: geoInfo.ISOCountryCode,
				StateCode:   geoInfo.RegionCode,
			}

			if ow.GetFeature().IsCountryGDPREnabled(geo.CountryCode) {
				geo.Compliance = GDPRCompliance
			} else if geo.CountryCode == CountryCodeUS && geo.StateCode == StateCodeCalifornia {
				geo.Compliance = USPCompliance
			} else if sectionid, ok := gppSectionIDs[geo.StateCode]; ok {
				geo.Compliance = GPPCompliance
				geo.SectionID = sectionid
			}

			w.Header().Set(headerCacheControl, "max-age="+fmt.Sprint(cacheTimeout.Seconds()))
			json.NewEncoder(w).Encode(geo)
		}
	}
	if !success {
		metricEngine.RecordGeoLookupFailure(models.EndpointGeo)
	}
}
