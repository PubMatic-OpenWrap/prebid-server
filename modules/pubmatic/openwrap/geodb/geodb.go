package geodb

import "git.pubmatic.com/PubMatic/go-netacuity-client"

type Geography interface {
	LookUp(ip string) (*netacuity.GeoInfo, error)
}

type GeoLookUp struct{}

func (geo GeoLookUp) LookUp(ip string) (*netacuity.GeoInfo, error) {
	return netacuity.LookUp(ip)
}

var InitNetacuityClient = func(dbPath string) error {
	return netacuity.InitNetacuityClient(dbPath)
}
