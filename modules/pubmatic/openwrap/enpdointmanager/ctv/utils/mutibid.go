package ctvutils

import (
	"strconv"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

func AddMultiBidConfigurations(rCtx *models.RequestCtx) {
	var multibid []*openrtb_ext.ExtMultiBid
	for _, partnerConfig := range rCtx.PartnerConfigMap {
		if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
			continue
		}

		partneridstr, ok := partnerConfig[models.PARTNER_ID]
		if !ok {
			continue
		}
		partnerID, err := strconv.Atoi(partneridstr)
		if err != nil || partnerID == models.VersionLevelConfigID {
			continue
		}

		// bidderCode is in context with pubmatic. Ex. it could be appnexus-1, appnexus-2, etc.
		bidderCode := partnerConfig[models.BidderCode]

		// prebidBidderCode is equivalent of PBS-Core's bidderCode
		prebidBidderCode := partnerConfig[models.PREBID_PARTNER_NAME]

		multibidConfig := &openrtb_ext.ExtMultiBid{
			Bidder:                 prebidBidderCode,
			Alias:                  bidderCode,
			MaxBids:                ptrutil.ToPtr(int(openrtb_ext.MaxBidLimit)),
			TargetBidderCodePrefix: bidderCode,
		}
		multibid = append(multibid, multibidConfig)
	}

	rCtx.NewReqExt.Prebid.MultiBid = multibid
}
