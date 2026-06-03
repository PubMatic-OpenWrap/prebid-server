package models

// OmitTrackerBidExp is true when bid.ext bidexp_enf is explicitly 0 and AppSubIntegrationPath
// resolves to AdMob SDK Bidding (see AppSubIntegrationPathIDAdMobSDKBidding / app_sub_integration_path).
func OmitTrackerBidExp(rctx RequestCtx, bidExpEnf *int) bool {
	if bidExpEnf == nil || *bidExpEnf != 0 {
		return false
	}
	if rctx.AppSubIntegrationPath == nil || *rctx.AppSubIntegrationPath < 0 {
		return false
	}
	return *rctx.AppSubIntegrationPath == AppSubIntegrationPathIDAdMobSDKBidding || *rctx.AppSubIntegrationPath == AppSubIntegrationPathIDGoogleAdManagerSDKBidding
}
