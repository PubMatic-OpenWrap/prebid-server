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
		return &geodb.GeoInfo{}, err
	}
	return &geodb.GeoInfo{
		CountryCode:    data.CountryCode,
		ISOCountryCode: data.ISOCountryCode,
	}, nil
}

// InitGeoDBClient initialises the geoDB client
func (geo NetAcuity) InitGeoDBClient(dbPath string) error {
	return netacuity.InitNetacuityClient(dbPath)
}
