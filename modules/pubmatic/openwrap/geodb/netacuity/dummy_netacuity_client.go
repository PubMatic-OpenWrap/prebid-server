package netacuity

import (
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/geodb"
)

// DummyNetAcuity instance for netacuity
type DummyNetAcuity struct{}

// LookUp function returns empty values for GeoInfo for non-linux platform
func (geo DummyNetAcuity) LookUp(ip string) (*geodb.GeoInfo, error) {
	return &geodb.GeoInfo{}, nil
}

// InitGeoDBClient do nothing for non-linux platform
func (geo DummyNetAcuity) InitGeoDBClient(dbPath string) error {
	return nil
}
