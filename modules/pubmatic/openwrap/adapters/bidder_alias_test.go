package adapters

import (
	"testing"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
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
		models.BidderAppStockAlias:          string(openrtb_ext.BidderLimelightDigital),
	}
	assert.Equal(t, expected, Alias())
}

func TestResolveOWBidder(t *testing.T) {
	assert.Equal(t, "", ResolveOWBidder(""))
	assert.Equal(t, models.BidderPubMatic, ResolveOWBidder(models.BidderPubMatic))
	assert.Equal(t, string(openrtb_ext.BidderAdf), ResolveOWBidder("adform_deprecated")) // deprecated custom alias
	assert.Equal(t, "tpmn_deprecated", ResolveOWBidder("tpmn_deprecated"))               // any other deprecated bidder
	for alias, coreBidder := range Alias() {
		assert.Equal(t, coreBidder, ResolveOWBidder(alias))
	}
}
