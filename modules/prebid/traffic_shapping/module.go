package trafficshapping

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/moduledeps"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func Builder(_ json.RawMessage, _ moduledeps.ModuleDeps) (interface{}, error) {
	return Module{}, nil
}

type Module struct {
}

func (m Module) HandleProcessedAuctionHook(
	context context.Context,
	moduleContext hookstage.ModuleInvocationContext,
	payload hookstage.ProcessedAuctionRequestPayload,
) (hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload], error) {
	hookResult := hookstage.HookResult[hookstage.ProcessedAuctionRequestPayload]{}

	request := openrtb_ext.RequestWrapper{
		BidRequest: payload.BidRequest,
	}
	requestRules, err := getRequestLevelRules(payload.BidRequest.Ext)
	bidderTargeting := getBidderTargeting()
	if err == nil {
		// iterate over request.impressions
		for _, imp := range request.GetImp() {
			impExt, err := imp.GetImpExt()
			if err == nil {
				biddersToRemove := []string{}
				// verify filter rule for each bidder
				for bidder := range impExt.GetPrebid().Bidder {
					targeting := bidderTargeting[bidder]
					targetBidder, err := m.bidderTargetingMatches(targeting, requestRules)
					if err == nil && !targetBidder {
						biddersToRemove = append(biddersToRemove, bidder)
					}
				}
				// remove bidder from impExt.prebid.bidder using biddersToRemove
				for _, bidder := range biddersToRemove {
					delete(impExt.GetPrebid().Bidder, bidder)
					warning := fmt.Sprintf("Removed bidder '%s' from Impression Id = '%s' (targeting rule [%s] not satified)", bidder, imp.ID, bidderTargeting[bidder])
					fmt.Println(warning)
					hookResult.Warnings = append(hookResult.Warnings, warning)
				}
				if len(biddersToRemove) > 0 {
					request.RebuildRequest()
				}
			}
		}
	}

	return hookResult, err
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

// Bidder rules : Profile version level
// allow only whitelisted keywords

func getBidderTargeting() map[string]Expression {
	// consider appnexus if series = friends and country = india
	seriesExp := Eq{
		Key:   "series",
		Value: "friends",
	}
	countryExp := Eq{
		Key:   "country",
		Value: "India",
	}
	appNexusRule := And{
		Left:  seriesExp,
		Right: countryExp,
	}

	// consider rubicon if series = saregampa
	fSeriesExp := Eq{
		Key:   "series",
		Value: "saregampa",
	}
	freewheelSspRule := fSeriesExp

	return map[string]Expression{
		"appnexus":     appNexusRule,
		"freewheelssp": freewheelSspRule,
	}
}

// keywords whitelisting (freeform type)  : publisher level
// var keywords = []string{"series", "country"}
