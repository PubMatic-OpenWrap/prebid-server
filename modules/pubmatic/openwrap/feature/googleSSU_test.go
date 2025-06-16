package feature

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockVASTXMLHandler is a mock implementation of VASTXMLHandler interface for testing
type MockVASTXMLHandler struct {
	mock.Mock
}

func (m *MockVASTXMLHandler) Parse(vast string) error {
	args := m.Called(vast)
	return args.Error(0)
}

func (m *MockVASTXMLHandler) Inject(videoParams []models.OWTracker, skipTracker bool) (string, error) {
	args := m.Called(videoParams, skipTracker)
	return args.String(0), args.Error(1)
}

func (m *MockVASTXMLHandler) AddCategoryTag(adCat []string) (string, error) {
	args := m.Called(adCat)
	return args.String(0), args.Error(1)
}

func (m *MockVASTXMLHandler) AddAdvertiserTag(adDomain string) (string, error) {
	args := m.Called(adDomain)
	return args.String(0), args.Error(1)
}

func TestEnrichVASTForSSUFeature(t *testing.T) {
	mockHandler := new(MockVASTXMLHandler)
	tests := []struct {
		name        string
		bidResponse *openrtb2.BidResponse
		mockSetup   func()
		expectedAdM string
	}{
		{
			name: "Valid_VAST_enrichment_with_category_and_advertiser",
			bidResponse: &openrtb2.BidResponse{
				ID: "resp1",
				SeatBid: []openrtb2.SeatBid{{
					Bid: []openrtb2.Bid{{
						ID:      "bid1",
						AdM:     "<VAST></VAST>",
						Cat:     []string{"IAB1", "IAB2"},
						ADomain: []string{"example.com"},
					}},
				}},
			},
			mockSetup: func() {
				mockHandler.On("Parse", "<VAST></VAST>").Return(nil).Once()
				mockHandler.On("AddCategoryTag", []string{"IAB1", "IAB2"}).Return("<VAST><Category><![CDATA[IAB1-1,IAB1-2]]></Category></VAST>", nil).Once()
				mockHandler.On("AddAdvertiserTag", "example.com").Return("<VAST><Category><![CDATA[IAB1-1,IAB1-2]]></Category><Advertiser><CDATA[example.com]]></Advertiser></VAST>", nil).Once()
			},
			expectedAdM: "<VAST><Category><![CDATA[IAB1-1,IAB1-2]]></Category><Advertiser><CDATA[example.com]]></Advertiser></VAST>",
		},
		{
			name: "Valid_VAST_enrichment_with_category_and_multiple_advertiser",
			bidResponse: &openrtb2.BidResponse{
				ID: "resp1",
				SeatBid: []openrtb2.SeatBid{{
					Bid: []openrtb2.Bid{{
						ID:      "bid1",
						AdM:     "<VAST></VAST>",
						Cat:     []string{"IAB1", "IAB2"},
						ADomain: []string{"example.com", "test.com"},
					}},
				}},
			},
			mockSetup: func() {
				mockHandler.On("Parse", "<VAST></VAST>").Return(nil).Once()
				mockHandler.On("AddCategoryTag", []string{"IAB1", "IAB2"}).Return("<VAST><Category><![CDATA[IAB1-1,IAB1-2]]></Category></VAST>", nil).Once()
				mockHandler.On("AddAdvertiserTag", "example.com").Return("<VAST><Category><![CDATA[IAB1-1,IAB1-2]]></Category><Advertiser><CDATA[example.com]]></Advertiser></VAST>", nil).Once()
			},
			expectedAdM: "<VAST><Category><![CDATA[IAB1-1,IAB1-2]]></Category><Advertiser><CDATA[example.com]]></Advertiser></VAST>",
		},
		{
			name: "Skip_bid_with_empty_AdM",
			bidResponse: &openrtb2.BidResponse{
				ID: "resp2",
				SeatBid: []openrtb2.SeatBid{{
					Bid: []openrtb2.Bid{{ID: "bid2", AdM: ""}},
				}},
			},
			mockSetup:   func() {},
			expectedAdM: "",
		},
		{
			name: "Parse_error_skips_bid",
			bidResponse: &openrtb2.BidResponse{
				ID: "resp3",
				SeatBid: []openrtb2.SeatBid{{
					Bid: []openrtb2.Bid{{ID: "bid3", AdM: "<invalid/>"}},
				}},
			},
			mockSetup: func() {
				mockHandler.On("Parse", "<invalid/>").Return(assert.AnError).Once()
			},
			expectedAdM: "<invalid/>",
		},
		{
			name: "Valid_VAST_enrichment_with_category",
			bidResponse: &openrtb2.BidResponse{
				ID: "resp4",
				SeatBid: []openrtb2.SeatBid{{
					Bid: []openrtb2.Bid{{
						ID:  "bid4",
						AdM: "<VAST><Ad><InLine></InLine></Ad></VAST>",
						Cat: []string{"IAB1-1", "IAB1-2"},
					}},
				}},
			},
			mockSetup: func() {
				mockHandler.On("Parse", "<VAST><Ad><InLine></InLine></Ad></VAST>").Return(nil).Once()
				mockHandler.On("AddCategoryTag", []string{"IAB1-1", "IAB1-2"}).Return("<VAST><Ad><InLine><Category><![CDATA[IAB1-1,IAB1-2]]></Category></InLine></Ad></VAST>", nil).Once()
			},
			expectedAdM: "<VAST><Ad><InLine><Category><![CDATA[IAB1-1,IAB1-2]]></Category></InLine></Ad></VAST>",
		},
		{
			name: "Valid_VAST_enrichment_with_advertiser",
			bidResponse: &openrtb2.BidResponse{
				ID: "resp5",
				SeatBid: []openrtb2.SeatBid{{
					Bid: []openrtb2.Bid{{
						ID:      "bid5",
						AdM:     "<VAST><Ad><InLine></InLine></Ad></VAST>",
						ADomain: []string{"example.com"},
					}},
				}},
			},
			mockSetup: func() {
				mockHandler.On("Parse", "<VAST><Ad><InLine></InLine></Ad></VAST>").Return(nil).Once()
				mockHandler.On("AddAdvertiserTag", "example.com").Return("<VAST><Ad><InLine><Advertiser><CDATA[example.com]]></Advertiser></InLine></Ad></VAST>", nil).Once()
			},
			expectedAdM: "<VAST><Ad><InLine><Advertiser><CDATA[example.com]]></Advertiser></InLine></Ad></VAST>",
		},
		{
			name: "Valid_VAST_enrichment_dont_override_category_and_advertiser",
			bidResponse: &openrtb2.BidResponse{
				ID: "resp6",
				SeatBid: []openrtb2.SeatBid{{
					Bid: []openrtb2.Bid{{
						ID:      "bid6",
						AdM:     "<VAST><Ad><InLine><Category><CDATA[IAB1-1,IAB1-2]]></Category><Advertiser><CDATA[example.com]]></Advertiser></InLine></Ad></VAST>",
						Cat:     []string{"IAB1-3", "IAB1-4"},
						ADomain: []string{"test.com"},
					}},
				}},
			},
			mockSetup: func() {
				mockHandler.On("Parse", "<VAST><Ad><InLine><Category><CDATA[IAB1-1,IAB1-2]]></Category><Advertiser><CDATA[example.com]]></Advertiser></InLine></Ad></VAST>").Return(assert.AnError).Once()
			},
			expectedAdM: "<VAST><Ad><InLine><Category><CDATA[IAB1-1,IAB1-2]]></Category><Advertiser><CDATA[example.com]]></Advertiser></InLine></Ad></VAST>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler.ExpectedCalls = nil
			mockHandler.Calls = nil
			tt.mockSetup()

			EnrichVASTForSSUFeature(tt.bidResponse, mockHandler)

			actualAdM := tt.bidResponse.SeatBid[0].Bid[0].AdM
			assert.Equal(t, tt.expectedAdM, actualAdM)

			mockHandler.AssertExpectations(t)
		})
	}
}
