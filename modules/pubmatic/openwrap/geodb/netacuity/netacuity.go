package netacuity

import (
	"git.pubmatic.com/PubMatic/go-netacuity-client"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/geodb"
)

type NetAcuity struct{}

// LookUp function performs the ip-to-geo lookup
func (geo NetAcuity) LookUp(ip string) (*geodb.GeoInfo, error) {
	geoInfo, err := netacuity.LookUp(ip)
	if err != nil {
		return nil, err
	}
	return &geodb.GeoInfo{
		CountryCode:    geoInfo.CountryCode,
		ISOCountryCode: geoInfo.ISOCountryCode,
		RegionCode:     geoInfo.RegionCode,
		City:           geoInfo.City,
		PostalCode:     geoInfo.PostalCode,
		DmaCode:        geoInfo.DmaCode,
		Latitude:       geoInfo.Latitude,
		Longitude:      geoInfo.Longitude,
		AreaCode:       geoInfo.AreaCode,
	}, nil
}

// InitGeoDBClient initialises the NetAcuity client
func (geo NetAcuity) InitGeoDBClient(dbPath string) error {
	return netacuity.InitNetacuityClient(dbPath)
}
