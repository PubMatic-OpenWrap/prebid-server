package vastbidder

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fastXMLParser_GetAdvertiser(t *testing.T) {
	for _, tt := range getAdvertiserTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			parser := newFastXMLParser()
			err := parser.Parse([]byte(tt.vastXML))
			if !assert.NoError(t, err) {
				return
			}
			gotAdvertisers := parser.GetAdvertiser()
			sort.Strings(gotAdvertisers)
			sort.Strings(tt.wantAdvertisers)

			assert.Equal(t, tt.wantAdvertisers, gotAdvertisers)
			assert.Equal(t, len(tt.wantAdvertisers), len(gotAdvertisers))
		})
	}
}

func Test_fastXMLParser_GetCreativeId(t *testing.T) {
	for _, tt := range getCreativeIDTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			generateRandomID = func() string { return tt.randomID }
			parser := newFastXMLParser()
			err := parser.Parse([]byte(tt.vastXML))
			if !assert.NoError(t, err) {
				return
			}

			gotID := parser.GetCreativeID()
			assert.Equal(t, tt.wantID, gotID)
		})
	}
}

func Test_fastXMLParser_GetDuration(t *testing.T) {
	for _, tt := range getCreativeDurationTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			parser := newFastXMLParser()
			err := parser.Parse([]byte(tt.vastXML))
			if !assert.NoError(t, err) {
				return
			}

			gotID, gotErr := parser.GetDuration()
			assert.Equal(t, tt.wantDuration, gotID)
			assert.Equal(t, tt.wantErr, gotErr)
		})
	}
}

func Test_fastXMLParser_GetPricingDetails(t *testing.T) {
	for _, tt := range getPricingDetailsTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			parser := newFastXMLParser()
			err := parser.Parse([]byte(tt.vastXML))
			if !assert.NoError(t, err) {
				return
			}
			gotPrice, gotCurrency := parser.GetPricingDetails()
			assert.Equal(t, tt.wantPrice, gotPrice)
			assert.Equal(t, tt.wantCurrency, gotCurrency)
		})
	}
}

func Test_fastXMLParser_getPricingNode(t *testing.T) {
	for _, tt := range getPricingNodeTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			parser := newFastXMLParser()
			err := parser.Parse([]byte(tt.vastXML))
			if !assert.NoError(t, err) {
				return
			}
			node := parser.getPricingNode()
			if tt.wantNil {
				assert.Nil(t, node)
			} else {
				assert.NotNil(t, node)
				assert.Equal(t, tt.wantPrice, strings.TrimSpace(parser.reader.RawText(node)))
			}
		})
	}
}
