package gocache

import (
	"bytes"
	"fmt"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func (c *cache) GetBidderFilterConditions(rCtx models.RequestCtx) map[string]*bytes.Reader {
	key := fmt.Sprintf("%d_%d_%d_bidding_conditions", rCtx.PubID, rCtx.ProfileID, rCtx.DisplayID)

	biddingConditions, ok := c.Get(key)
	if ok {
		return biddingConditions.(map[string]*bytes.Reader)
	}
	bidderToBiddingCondition := map[string]*bytes.Reader{}
	defaultAdUnitConfig, ok := rCtx.AdUnitConfig.Config[models.AdunitConfigDefaultKey]
	if ok && defaultAdUnitConfig != nil && defaultAdUnitConfig.BidderFilter != nil {
		for _, filterCfg := range defaultAdUnitConfig.BidderFilter.FilterConfig {
			reader := bytes.NewReader(filterCfg.BiddingConditions)
			for _, bidder := range filterCfg.Bidders {
				bidderToBiddingCondition[bidder] = reader
			}
		}
	}
	c.Set(key, bidderToBiddingCondition)

	return bidderToBiddingCondition
}
