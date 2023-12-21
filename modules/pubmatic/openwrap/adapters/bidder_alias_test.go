package adapters

import (
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestAlias(t *testing.T) {
	expected := map[string]string{
		models.BidderAdGenerationAlias:      string(openrtb_ext.BidderAdgeneration),
		models.BidderDistrictmDMXAlias:      string(openrtb_ext.BidderDmx),
		models.BidderPubMaticSecondaryAlias: string(openrtb_ext.BidderPubmatic),
		models.BidderDistrictmAlias:         string(openrtb_ext.BidderAppnexus),
		models.BidderAndBeyondAlias:         string(openrtb_ext.BidderAdkernel),
		models.BidderMediaFuseAlias:         string(openrtb_ext.BidderAppnexus),
	}
	assert.Equal(t, expected, Alias())
}

func TestResolveOWBidder(t *testing.T) {
	assert.Equal(t, "", ResolveOWBidder(""))
	assert.Equal(t, models.BidderPubMatic, ResolveOWBidder(models.BidderPubMatic))
	for alias, coreBidder := range Alias() {
		assert.Equal(t, coreBidder, ResolveOWBidder(alias))
	}
}
