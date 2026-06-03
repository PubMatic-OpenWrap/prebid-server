package models

import (
	"testing"

	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

func TestOmitTrackerBidExp(t *testing.T) {
	zero := 0
	one := 1
	admobASIP := AppSubIntegrationPathIDAdMobSDKBidding
	otherASIP := 99

	if !OmitTrackerBidExp(RequestCtx{AppSubIntegrationPath: ptrutil.ToPtr(admobASIP)}, &zero) {
		t.Fatal("expected omit when bidexp_enf is 0 and AppSubIntegrationPath is AdMob SDK Bidding")
	}
	if OmitTrackerBidExp(RequestCtx{AppSubIntegrationPath: ptrutil.ToPtr(admobASIP)}, nil) {
		t.Fatal("did not expect omit when bidexp_enf is absent")
	}
	if OmitTrackerBidExp(RequestCtx{AppSubIntegrationPath: ptrutil.ToPtr(admobASIP)}, &one) {
		t.Fatal("did not expect omit when bidexp_enf is non-zero")
	}
	if OmitTrackerBidExp(RequestCtx{AppSubIntegrationPath: ptrutil.ToPtr(otherASIP)}, &zero) {
		t.Fatal("did not expect omit when AppSubIntegrationPath is not AdMob")
	}
	if OmitTrackerBidExp(RequestCtx{}, &zero) {
		t.Fatal("did not expect omit when AppSubIntegrationPath is nil")
	}
	if OmitTrackerBidExp(RequestCtx{AppSubIntegrationPath: ptrutil.ToPtr(-1)}, &zero) {
		t.Fatal("did not expect omit when AppSubIntegrationPath is unset/invalid (-1)")
	}
}
