package geodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoDBLookupFailureCases(t *testing.T) {
	// initialise with invalid path
	err := InitGeoDBClient("/invalid/path")
	assert.Errorf(t, err, "InitNetacuityClient should return an error")

	info, err := GeoLookUp{}.LookUp("10.10.10.10")
	assert.Errorf(t, err, "LookUp should return an error")
	assert.Nilf(t, info, "LookUp should return nil geoinfo")
}
