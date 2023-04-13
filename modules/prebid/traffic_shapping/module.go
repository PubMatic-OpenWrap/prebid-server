package trafficshapping

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func Builder(cfg json.RawMessage, _ moduledeps.ModuleDeps) (interface{}, error) {
	//Accept bidder targeting rules here from config
	fmt.Println(cfg)
	return Module{}, nil
}

type Module struct {
}

func (m Module) HandleProcessedAuctionHook(
	context context.Context,
	moduleContext hookstage.ModuleInvocationContext,
	payload hookstage.ProcessedAuctionRequestPayload,
) (hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error) {
	result := hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{}
	result.ChangeSet = hookstage.ChangeSet[hookstage.ProcessedAuctionRequestPayload]{}

	// request := openrtb_ext.RequestWrapper{
	// 	BidRequest: payload.BidRequest,
	// }
	rw := payload.BidRequest
	requestRules, err := getRequestLevelRules(rw.Ext)
	rtbRequest := RTBRequest{rw.BidRequest}
	bidderTargeting := getBidderTargeting(rtbRequest)
	result.Warnings = make([]string, 0)
	if err == nil {
		// iterate over request.impressions
		newImps := make([]*openrtb_ext.ImpWrapper, 0)
		for _, imp := range rw.GetImp() {
			impExt, err := imp.GetImpExt()
			if err == nil {
				biddersToRemove := []string{}
				impExtPrebid := impExt.GetPrebid()
				// verify filter rule for each bidder
				for bidder := range impExtPrebid.Bidder {
					targeting := bidderTargeting[bidder]
					targetBidder, err := m.bidderTargetingMatches(targeting, requestRules)
					if err == nil && !targetBidder {
						biddersToRemove = append(biddersToRemove, bidder)
					}
				}
				// remove bidder from impExt.prebid.bidder using biddersToRemove
				for _, bidder := range biddersToRemove {
					delete(impExtPrebid.Bidder, bidder)
					warning := fmt.Sprintf("Removed bidder '%s' from Impression Id = '%s' targeting rule [%s] not satisfied", bidder, imp.ID, bidderTargeting[bidder].GetName())
					fmt.Println(warning)
					result.Warnings = append(result.Warnings, warning)
				}
				if len(biddersToRemove) > 0 {
					impExt.SetPrebid(impExtPrebid)
					imp.RebuildImp()
				}
				newImps = append(newImps, imp)
			}
		}
		rw.SetImp(newImps)
		rw.RebuildRequest()
	}

	result.ChangeSet.AddMutation(func(parp hookstage.ProcessedAuctionRequestPayload) (hookstage.ProcessedAuctionRequestPayload, error) {
		parp.BidRequest = rw
		return parp, err
	}, hookstage.MutationUpdate, "filter-bidders-traffic-shaping")

	return result, err
}

// keywords and values : User Input  (req.ext.rules)
type rules map[string]string

func getRequestLevelRules(reqExt json.RawMessage) (rules, error) {
	var Ext struct {
		Rules rules `json:"rules"`
	}
	return Ext.Rules, json.Unmarshal(reqExt, &Ext)
}

// goval.Evaluator based - https://github.com/maja42/goval
func (m Module) bidderTargetingMatches(rule Expression, variables map[string]string) (bool, error) {
	if rule == nil {
		return true, nil
	}

	return rule.Evaluate(variables), nil
}

// keywords whitelisting (freeform type)  : publisher level
// var keywords = []string{"series", "country"}
