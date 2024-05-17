//go:build !test

package openwrap

import (
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/geodb/netacuity"
)

func foo() geodb.Geography {
	return netacuity.NetAcuity{}
}
