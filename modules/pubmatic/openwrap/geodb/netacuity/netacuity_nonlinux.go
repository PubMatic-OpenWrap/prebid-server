//go:build !linux || ignoreNetacuity
// +build !linux ignoreNetacuity

// Package netacuity offers methods for initializing a GeoIP database client and
// to perform the ip-to-geo lookup functionality.
// This file removes the compile time dependency of go-netacuity-client library to makes sure
// that the application compiles and run successfully on non-linux platforms (including macOS).
package netacuity

import (
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb"
)

type NetAcuity struct{}

// LookUp function returns empty values for GeoInfo for non-linux platform
func (geo NetAcuity) LookUp(ip string) (*geodb.GeoInfo, error) {
	return &geodb.GeoInfo{}, nil
}

// NewNetAcuity initialises the NetAcuity client
func NewNetacuity(dbPath string) (*NetAcuity, error) {
	return &NetAcuity{}, nil
}
