//go:build linux && !ignoreNetacuity
// +build linux,!ignoreNetacuity

// Package netacuity offers methods for initializing a GeoIP database client and
// to perform the ip-to-geo lookup functionality.
// Build constraint flag makes sure that this file compiles only for linux platform
package netacuity

import (
	"strings"

	"git.pubmatic.com/PubMatic/go-netacuity-client"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb"
)

type NetAcuity struct{}

// LookUp function performs the ip-to-geo lookup
func (geo NetAcuity) LookUp(ip string) (*geodb.GeoInfo, error) {
	geoInfo, err := netacuity.LookUp(ip)
	if err != nil {
		return nil, err
	}
	return &geodb.GeoInfo{
		CountryCode:           geoInfo.CountryCode,
		ISOCountryCode:        geoInfo.ISOCountryCode,
		RegionCode:            geoInfo.RegionCode,
		City:                  geoInfo.City,
		PostalCode:            geoInfo.PostalCode,
		DmaCode:               geoInfo.DmaCode,
		Latitude:              geoInfo.Latitude,
		Longitude:             geoInfo.Longitude,
		AreaCode:              geoInfo.AreaCode,
		AlphaThreeCountryCode: strings.ToUpper(geoInfo.AlphaThreeCountryCode),
	}, nil
}

// InitGeoDBClient initialises the NetAcuity client
func (geo NetAcuity) InitGeoDBClient(dbPath string) error {
	return netacuity.InitNetacuityClient(dbPath)
}
