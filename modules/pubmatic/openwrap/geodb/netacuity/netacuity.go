package netacuity

import (
	"git.pubmatic.com/PubMatic/go-netacuity-client"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/geodb"
)

type NetAcuity struct{}

// LookUp function performs the ip-to-geo lookup
func (geo NetAcuity) LookUp(ip string) (*geodb.GeoInfo, error) {
	data, err := netacuity.LookUp(ip)
	if err != nil {
		return nil, err
	}
	return &geodb.GeoInfo{
		CountryCode:    data.CountryCode,
		ISOCountryCode: data.ISOCountryCode,
		RegionCode:     data.RegionCode,
		City:           data.City,
		PostalCode:     data.PostalCode,
		DmaCode:        data.DmaCode,
		Latitude:       data.Latitude,
		Longitude:      data.Longitude,
		AreaCode:       data.AreaCode,
	}, nil
}

// InitGeoDBClient initialises the NetAcuity client
func (geo NetAcuity) InitGeoDBClient(dbPath string) error {
	return netacuity.InitNetacuityClient(dbPath)
}
