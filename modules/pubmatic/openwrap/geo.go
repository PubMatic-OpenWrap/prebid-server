package openwrap

import (
	"net/http"
	"time"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb/netacuity"
)

// geo provides geo metadata from ip
type geo struct {
	countryCode string `json:"cc,omitempty"`
	stateCode   string `json:"sc,omitempty"`
	compliance  string `json:"compliance,omitempty"`
	sectionID   string `json:"sectionId,omitempty"`
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

// Handler for /geo endpoint
func (handler *geoHandler) Handle(w http.ResponseWriter, r *http.Request) {

}
