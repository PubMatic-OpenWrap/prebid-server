package pubmatic

import (
	"testing"

	"github.com/prebid/prebid-server/analytics"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/hooks/hookanalytics"
	"github.com/prebid/prebid-server/hooks/hookexecution"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPLogger(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.PubMaticWL
	}{
		{
			name: "check if NewHTTPLogger returns nil",
			cfg: config.PubMaticWL{
				MaxClients:     5,
				MaxConnections: 50,
				MaxCalls:       5,
				RespTimeout:    50,
			},
		},
	}
	for _, tt := range tests {
		module := NewHTTPLogger(tt.cfg)
		assert.NotNil(t, module, tt.name)
	}
}

// TestLogAuctionObject just increases code coverage, it does not validate anything
func TestLogAuctionObject(t *testing.T) {
	tests := []struct {
		name string
		ao   *analytics.AuctionObject
	}{
		{
			name: "rctx is nil",
			ao:   &analytics.AuctionObject{},
		},
		{
			name: "rctx is present",
			ao: &analytics.AuctionObject{
				HookExecutionOutcome: []hookexecution.StageOutcome{
					{
						Groups: []hookexecution.GroupOutcome{
							{
								InvocationResults: []hookexecution.HookOutcome{
									{
										AnalyticsTags: hookanalytics.Analytics{
											Activities: []hookanalytics.Activity{
												{
													Results: []hookanalytics.Result{
														{
															Values: map[string]interface{}{
																"request-ctx": &models.RequestCtx{},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		HTTPLogger{}.LogAuctionObject(tt.ao)
	}
}
