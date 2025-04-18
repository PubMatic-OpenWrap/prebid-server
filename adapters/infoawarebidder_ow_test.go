package adapters

import (
	"errors"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/errortypes"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestAppNotSupportedOW(t *testing.T) {
	bidder := &mockBidder{}
	info := config.BidderInfo{
		Capabilities: &config.CapabilitiesInfo{
			Site: &config.PlatformInfo{
				MediaTypes: []openrtb_ext.BidType{openrtb_ext.BidTypeBanner},
			},
		},
	}
	constrained := BuildInfoAwareBidder(bidder, info)
	bids, errs := constrained.MakeRequests(&openrtb2.BidRequest{
		Imp: []openrtb2.Imp{{ID: "imp-1", Banner: &openrtb2.Banner{}}},
		App: &openrtb2.App{},
	}, &ExtraRequestInfo{})
	if !assert.Len(t, errs, 1) {
		return
	}
	assert.EqualError(t, errs[0], "this bidder does not support app requests")
	assert.IsType(t, &errortypes.Warning{}, errs[0])
	assert.Len(t, bids, 0)
}

func TestSiteNotSupported(t *testing.T) {
	bidder := &mockBidder{}
	info := config.BidderInfo{
		Capabilities: &config.CapabilitiesInfo{
			App: &config.PlatformInfo{
				MediaTypes: []openrtb_ext.BidType{openrtb_ext.BidTypeBanner},
			},
		},
	}
	constrained := BuildInfoAwareBidder(bidder, info)
	bids, errs := constrained.MakeRequests(&openrtb2.BidRequest{
		Imp:  []openrtb2.Imp{{ID: "imp-1", Banner: &openrtb2.Banner{}}},
		Site: &openrtb2.Site{},
	}, &ExtraRequestInfo{})
	if !assert.Len(t, errs, 1) {
		return
	}
	assert.EqualError(t, errs[0], "this bidder does not support site requests")
	assert.IsType(t, &errortypes.Warning{}, errs[0])
	assert.Len(t, bids, 0)
}

func TestImpFiltering(t *testing.T) {
	bidder := &mockBidder{}
	info := config.BidderInfo{
		Capabilities: &config.CapabilitiesInfo{
			Site: &config.PlatformInfo{
				MediaTypes: []openrtb_ext.BidType{openrtb_ext.BidTypeVideo},
			},
			App: &config.PlatformInfo{
				MediaTypes: []openrtb_ext.BidType{openrtb_ext.BidTypeBanner},
			},
		},
	}

	constrained := BuildInfoAwareBidder(bidder, info)

	testCases := []struct {
		description    string
		inBidRequest   *openrtb2.BidRequest
		expectedErrors []error
		expectedImpLen int
	}{
		{
			description: "Empty Imp array. MakeRequest() call not expected",
			inBidRequest: &openrtb2.BidRequest{
				Imp:  []openrtb2.Imp{},
				Site: &openrtb2.Site{},
			},
			expectedErrors: []error{
				&errortypes.BadInput{Message: "Bid request didn't contain media types supported by the bidder"},
			},
			expectedImpLen: 0,
		},
		{
			description: "Sole imp in bid request is of wrong media type. MakeRequest() call not expected",
			inBidRequest: &openrtb2.BidRequest{
				Imp: []openrtb2.Imp{{ID: "imp-1", Video: &openrtb2.Video{}}},
				App: &openrtb2.App{},
			},
			expectedErrors: []error{
				&errortypes.BadInput{Message: "request.imp[0] uses video, but this bidder doesn't support it"},
				&errortypes.BadInput{Message: "Bid request didn't contain media types supported by the bidder"},
			},
			expectedImpLen: 0,
		},
		{
			description: "All imps in bid request of wrong media type, MakeRequest() call not expected",
			inBidRequest: &openrtb2.BidRequest{
				Imp: []openrtb2.Imp{
					{ID: "imp-1", Video: &openrtb2.Video{}},
					{ID: "imp-2", Native: &openrtb2.Native{}},
					{ID: "imp-3", Audio: &openrtb2.Audio{}},
				},
				App: &openrtb2.App{},
			},
			expectedErrors: []error{
				&errortypes.BadInput{Message: "request.imp[0] uses video, but this bidder doesn't support it"},
				&errortypes.BadInput{Message: "request.imp[1] uses native, but this bidder doesn't support it"},
				&errortypes.BadInput{Message: "request.imp[2] uses audio, but this bidder doesn't support it"},
				&errortypes.BadInput{Message: "Bid request didn't contain media types supported by the bidder"},
			},
			expectedImpLen: 0,
		},
		{
			description: "Some imps with correct media type, MakeRequest() call expected",
			inBidRequest: &openrtb2.BidRequest{
				Imp: []openrtb2.Imp{
					{
						ID:    "imp-1",
						Video: &openrtb2.Video{},
					},
					{
						Native: &openrtb2.Native{},
					},
					{
						ID:     "imp-2",
						Video:  &openrtb2.Video{},
						Native: &openrtb2.Native{},
					},
					{
						Banner: &openrtb2.Banner{},
					},
				},
				Site: &openrtb2.Site{},
			},
			expectedErrors: []error{
				&errortypes.BadInput{Message: "request.imp[1] uses native, but this bidder doesn't support it"},
				&errortypes.BadInput{Message: "request.imp[2] uses native, but this bidder doesn't support it"},
				&errortypes.BadInput{Message: "request.imp[3] uses banner, but this bidder doesn't support it"},
				&errortypes.BadInput{Message: "request.imp[1] has no supported MediaTypes. It will be ignored"},
				&errortypes.BadInput{Message: "request.imp[3] has no supported MediaTypes. It will be ignored"},
			},
			expectedImpLen: 2,
		},
		{
			description: "All imps with correct media type, MakeRequest() call expected",
			inBidRequest: &openrtb2.BidRequest{
				Imp: []openrtb2.Imp{
					{ID: "imp-1", Video: &openrtb2.Video{}},
					{ID: "imp-2", Video: &openrtb2.Video{}},
				},
				Site: &openrtb2.Site{},
			},
			expectedErrors: nil,
			expectedImpLen: 2,
		},
	}

	for _, test := range testCases {
		actualAdapterRequests, actualErrs := constrained.MakeRequests(test.inBidRequest, &ExtraRequestInfo{})

		// Assert the request.Imp slice was correctly filtered and if MakeRequest() was called by asserting
		// the corresponding error messages were returned
		for i, expectedErr := range test.expectedErrors {
			assert.EqualError(t, expectedErr, actualErrs[i].Error(), "Test failed. Error[%d] in error list mismatch: %s", i, test.description)
		}

		// Extra MakeRequests() call check: our mockBidder returns an adapter request for every imp
		assert.Len(t, actualAdapterRequests, test.expectedImpLen, "Test failed. Incorrect length of filtered imps: %s", test.description)
	}
}

type mockBidder struct {
}

func (m *mockBidder) MakeRequests(request *openrtb2.BidRequest, reqInfo *ExtraRequestInfo) ([]*RequestData, []error) {
	var adapterRequests []*RequestData

	for i := 0; i < len(request.Imp); i++ {
		adapterRequests = append(adapterRequests, &RequestData{})
	}

	return adapterRequests, nil
}

func (m *mockBidder) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *RequestData, response *ResponseData) (*BidderResponse, []error) {
	return nil, []error{errors.New("mock MakeBids error")}
}
