// Package geodb provides an interface for performing IP-to-geography lookups using a GeoIP database
package geodb

type GeoInfo struct {
	CountryCode           string
	ISOCountryCode        string
	RegionCode            string
	City                  string
	PostalCode            string
	DmaCode               int
	Latitude              float64
	Longitude             float64
	AreaCode              string
	AlphaThreeCountryCode string
}

// Geography interface defines methods for initializing a GeoIP database client and performing
// IP-to-geography lookups. Implement this interface to create custom GeoIP database clients.
type Geography interface {
	LookUp(ip string) (*GeoInfo, error)
	InitGeoDBClient(dbPath string) error
}
