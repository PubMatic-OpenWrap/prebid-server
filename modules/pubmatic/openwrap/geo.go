package openwrap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.pubmatic.com/PubMatic/go-common/util"
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// compliance consts
const (
	gdprCompliance      = 1
	uspCompliance       = 2
	gppCompliance       = 3
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
	Compliance  int    `json:"gc,omitempty"`
	SectionID   int    `json:"gsId,omitempty"`
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
	)
	defer func() {
		metricEngine.RecordPublisherRequests(models.EndpointGeo, pubIDStr, "")
		panicHandler("HandleGeoEndpoint", pubIDStr)
	}()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		metricEngine.RecordBadRequests(models.EndpointGeo, pubIDStr, 0)
		return
	}

	if err := r.ParseForm(); err != nil {
		glog.Errorf("[geo] url:[%s] error:[%s]", r.URL.RawQuery, err.Error())
		metricEngine.RecordBadRequests(models.EndpointGeo, pubIDStr, 0)
		return
	}

	pubIDStr = r.FormValue(models.PublisherID)
	_, err := strconv.Atoi(pubIDStr)
	if err != nil {
		glog.Errorf("[geo] url:[%s] origin:[%s] referer:[%s] error:[%s]", r.URL.RawQuery,
			r.Header.Get(headerOriginKey), r.Header.Get(headerRefererKey), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		metricEngine.RecordBadRequests(models.EndpointGeo, pubIDStr, 0)
		return
	}
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
		return
	}

	if geo.CountryCode == countryCodeUS && geo.StateCode == stateCodeCalifornia {
		geo.Compliance = uspCompliance
		return
	}

	if sectionid, ok := gppSectionIDs[geo.StateCode]; ok {
		geo.Compliance = gppCompliance
		geo.SectionID = sectionid
	}
}
