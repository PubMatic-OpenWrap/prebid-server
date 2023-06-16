package bidderparams

import (
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/util/ptrutil"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestPreparePubMaticParamsV25(t *testing.T) {
	type args struct {
		rctx       models.RequestCtx
		cache      cache.Cache
		bidRequest openrtb2.BidRequest
		imp        openrtb2.Imp
		impExt     models.ImpExtension
		partnerID  int
	}
	tests := []struct {
		name               string
		args               args
		wanMatchedSlot     string
		wantMatchedPattern string
		wantIsRegexSlot    bool
		wantParams         []byte
	}{
		{
			name: "Test request",
			args: args{
				partnerID: 123,
				imp: openrtb2.Imp{
					Banner: &openrtb2.Banner{
						W: ptrutil.ToPtr(int64(10)),
						H: ptrutil.ToPtr(int64(10)),
					},
					TagID: "/adunit1",
				},
				rctx: models.RequestCtx{
					IsTestRequest: 1,
					PubID:         123,
					PartnerConfigMap: map[int]map[string]string{
						123: {models.BidderCode: "pubm",
							models.KEY_GEN_PATTERN: "_AU_",
						},
					},
					DisplayID: 1,
					ProfileID: 10,
				},
			},
			wanMatchedSlot:  "/adunit1",
			wantIsRegexSlot: false,
			wantParams:      []byte(`{"publisherId":"123","adSlot":"/adunit1","wrapper":{"version":1,"profile":10}}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3, err := PreparePubMaticParamsV25(tt.args.rctx, tt.args.cache, tt.args.bidRequest, tt.args.imp, tt.args.impExt, tt.args.partnerID)
			assert.Equal(t, got, tt.wanMatchedSlot, tt.name)
			assert.Equal(t, got1, tt.wantMatchedPattern, tt.name)
			assert.Equal(t, got2, tt.wantIsRegexSlot, tt.name)
			assert.Equal(t, got3, tt.wantParams, tt.name)
			assert.NoError(t, err, tt.name)
		})
	}
}
