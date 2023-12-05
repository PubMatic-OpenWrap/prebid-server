package geodb

type GeoInfo struct {
	CountryCode    string
	ISOCountryCode string
	RegionCode     string
	City           string
	PostalCode     string
	DmaCode        int
	Latitude       float64
	Longitude      float64
	AreaCode       string
}

// Geography interface contains ip-to-geo LookUp function
type Geography interface {
	LookUp(ip string) (*GeoInfo, error)
	InitGeoDBClient(dbPath string) error
}
