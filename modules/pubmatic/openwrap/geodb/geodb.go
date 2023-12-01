package geodb

import "git.pubmatic.com/PubMatic/go-netacuity-client"

// Geography interface contains ip-to-geo LookUp function
type Geography interface {
	LookUp(ip string) (*netacuity.GeoInfo, error)
}

type GeoLookUp struct{}

// LookUp function performs the ip-to-geo lookup
func (geo GeoLookUp) LookUp(ip string) (*netacuity.GeoInfo, error) {
	return netacuity.LookUp(ip)
}

// InitNetacuityClient initialises the netacuity client
func InitNetacuityClient(dbPath string) error {
	return netacuity.InitNetacuityClient(dbPath)
}
