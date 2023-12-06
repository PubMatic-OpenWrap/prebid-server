package netacuity

import (
	"fmt"
	"testing"

	"git.pubmatic.com/PubMatic/go-netacuity-client"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/geodb"
	"github.com/stretchr/testify/assert"
)

func TestLookUp(t *testing.T) {

	currentLookup := lookupFunc
	defer func() {
		lookupFunc = currentLookup
	}()

	type want struct {
		err     error
		geoinfo *geodb.GeoInfo
	}

	tests := []struct {
		name          string
		setupLookFunc func()
		want          want
	}{
		{
			name: "lookup_success",
			setupLookFunc: func() {
				lookupFunc = func(ip string) (*netacuity.GeoInfo, error) {
					return &netacuity.GeoInfo{
						CountryCode:    "IN",
						ISOCountryCode: "IN",
						City:           "Pune",
					}, nil
				}
			},
			want: want{
				err: nil,
				geoinfo: &geodb.GeoInfo{
					CountryCode:    "IN",
					ISOCountryCode: "IN",
					City:           "Pune",
				},
			},
		},
		{
			name: "lookup_fail",
			setupLookFunc: func() {
				lookupFunc = func(ip string) (*netacuity.GeoInfo, error) {
					return nil, fmt.Errorf("error")
				}
			},
			want: want{
				err:     fmt.Errorf("error"),
				geoinfo: nil,
			},
		},
	}
	for _, tt := range tests {
		tt.setupLookFunc()
		geoinfo, err := NetAcuity{}.LookUp("10.10.10.10")
		assert.Equalf(t, tt.want.err, err, "mismatched error")
		assert.Equalf(t, tt.want.geoinfo, geoinfo, "mismatched geoinfo")
	}
}
