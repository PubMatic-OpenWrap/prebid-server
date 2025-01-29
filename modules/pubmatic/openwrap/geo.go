package openwrap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.pubmatic.com/PubMatic/go-common/util"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/metrics/prometheus"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// compliance consts
const (
	gdprCompliance      = "GDPR"
	uspCompliance       = "USP"
	gppCompliance       = "GPP"
	countryCodeUS       = "US"
	stateCodeCalifornia = "ca"
)

// headers const
const (
	cacheTimeout                   = time.Duration(48) * time.Hour
	headerAccessControlAllowOrigin = "Access-Control-Allow-Origin"
	headerCacheControl             = "Cache-Control"
	headerOriginKey                = "Origin"
	headerRefererKey               = "Referer"
)

var maxAgeHeaderValue = "max-age=" + fmt.Sprint(cacheTimeout.Seconds())

// geo provides geo metadata from ip
type geo struct {
	CountryCode string `json:"cc,omitempty"`
	StateCode   string `json:"sc,omitempty"`
	Compliance  string `json:"compliance,omitempty"`
	SectionID   int    `json:"sId,omitempty"`
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
	var (
		pubIDStr     string
		metricEngine = ow.GetMetricEngine()
		metricLabels = metrics.Labels{RType: models.EndpointGeo, RequestStatus: prometheus.RequestStatusBadInput}
	)
	defer func() {
		metricEngine.RecordRequest(metricLabels)
		panicHandler("HandleGeoEndpoint", pubIDStr)
	}()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		return
	}

	_, err := strconv.Atoi(r.FormValue(models.PublisherID))
	if err != nil {
		glog.Errorf("[geo] url:[%s] origin:[%s] referer:[%s] error:[%s]", r.URL.RawQuery,
			r.Header.Get(headerOriginKey), r.Header.Get(headerRefererKey), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	metricLabels.RequestStatus = prometheus.RequestStatusOK
	w.Header().Set(models.ContentType, models.ContentTypeApplicationJSON)
	w.Header().Set(headerAccessControlAllowOrigin, "*")

	ip := util.GetIP(r)
	geoInfo, err := ow.geoInfoFetcher.LookUp(ip)
	if err != nil {
		glog.Errorf("[geo] url:[%s] ip:[%s] error:[%s]", r.URL.RawQuery, ip, err.Error())
	}

	if geoInfo == nil || geoInfo.ISOCountryCode == "" {
		metricEngine.RecordGeoLookupFailure(models.EndpointGeo)
		return
	}

	geo := geo{
		CountryCode: geoInfo.ISOCountryCode,
		StateCode:   geoInfo.RegionCode,
	}
	updateGeoObject(&geo)

	w.Header().Set(headerCacheControl, maxAgeHeaderValue)
	json.NewEncoder(w).Encode(geo)
}

func updateGeoObject(geo *geo) {
	if ow.GetFeature().IsCountryGDPREnabled(geo.CountryCode) {
		geo.Compliance = gdprCompliance
	} else if geo.CountryCode == countryCodeUS && geo.StateCode == stateCodeCalifornia {
		geo.Compliance = uspCompliance
	} else if sectionid, ok := gppSectionIDs[geo.StateCode]; ok {
		geo.Compliance = gppCompliance
		geo.SectionID = sectionid
	}
}
