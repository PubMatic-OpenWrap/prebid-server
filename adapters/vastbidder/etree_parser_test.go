package vastbidder

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getAdvertiserTestCases() []struct {
	name            string
	vastXML         string
	wantAdvertisers []string
} {
	return []struct {
		name            string
		vastXML         string
		wantAdvertisers []string
	}{
		{
			name: "vast_4_with_advertiser",
			vastXML: `<VAST version="4.2" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST">
					  <Ad id="20001" sequence="1" >
						<InLine>
							  <Advertiser>www.iabtechlab.com</Advertiser>
						</InLine>
					  </Ad>
				  </VAST>`,
			wantAdvertisers: []string{"www.iabtechlab.com"},
		},
		{
			name: "vast_4_without_advertiser",
			vastXML: `<VAST version="4.2" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST">
					  <Ad id="20001" sequence="1" >
						<InLine>
						</InLine>
					  </Ad>
				  </VAST>`,
			wantAdvertisers: nil,
		},
		{
			name: "vast_4_with_empty_advertiser",
			vastXML: `<VAST version="4.2" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST">
					  <Ad id="20001" sequence="1" >
						<InLine>
							<Advertiser></Advertiser>
						</InLine>
					  </Ad>
				  </VAST>`,
			wantAdvertisers: nil,
		},
		{
			name: "vast_2_with_single_advertiser",
			vastXML: `<VAST version="2.0" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST">
					  <Ad id="20001" sequence="1" >
						<InLine>
							<Extensions>
								<Extension type="advertiser">
									<Advertiser>google.com</Advertiser>
								</Extension>
							</Extensions>
						</InLine>
					  </Ad>
				  </VAST>`,
			wantAdvertisers: []string{"google.com"},
		},
		{
			name: "vast_2_with_two_advertiser",
			vastXML: `<VAST version="2.0" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST">
					  <Ad id="20001" sequence="1" >
						<InLine>
							<Extensions>
								<Extension type="advertiser">
									<Advertiser>google.com</Advertiser>
								</Extension>
								<Extension type="advertiser">
									<Advertiser>facebook.com</Advertiser>
								</Extension>
							</Extensions>
						</InLine>
					  </Ad>
				  </VAST>`,
			wantAdvertisers: []string{"google.com", "facebook.com"},
		},
		{
			name: "vast_2_with_no_advertiser",
			vastXML: `<VAST version="2.0" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST">
					  <Ad id="20001" sequence="1" >
						<InLine>
						</InLine>
					  </Ad>
				  </VAST>`,
			wantAdvertisers: nil,
		},
		{
			name: "vast_2_with_epmty_advertiser",
			vastXML: `<VAST version="2.0" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST">
					  <Ad id="20001" sequence="1" >
						<InLine>
							<Extensions>
								<Extension type="advertiser">
									<Advertiser></Advertiser>
								</Extension>
							</Extensions>
						</InLine>
					  </Ad>
				  </VAST>`,
			wantAdvertisers: nil,
		},
		{
			name: "vast_3_with_single_advertiser",
			vastXML: `<VAST version="3.1" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST">
					  <Ad id="20001" sequence="1" >
						<InLine>
							<Extensions>
								<Extension type="advertiser">
									<Advertiser>google.com</Advertiser>
								</Extension>
							</Extensions>
						</InLine>
					  </Ad>
				  </VAST>`,
			wantAdvertisers: []string{"google.com"},
		},
		{
			name: "vast_3_with_two_advertiser",
			vastXML: `<VAST version="3.2" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST">
					  <Ad id="20001" sequence="1" >
						<InLine>
							<Extensions>
								<Extension type="advertiser">
									<Advertiser>google.com</Advertiser>
								</Extension>
								<Extension type="advertiser">
									<Advertiser>facebook.com</Advertiser>
								</Extension>
							</Extensions>
						</InLine>
					  </Ad>
				  </VAST>`,
			wantAdvertisers: []string{"google.com", "facebook.com"},
		},
		{
			name: "vast_3_with_no_advertiser",
			vastXML: `<VAST version="3.0" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST">
					  <Ad id="20001" sequence="1" >
						<InLine>
						</InLine>
					  </Ad>
				  </VAST>`,
			wantAdvertisers: nil,
		},
		{
			name: "vast_3_with_epmty_advertiser",
			vastXML: `<VAST version="3.1" xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns="http://www.iab.com/VAST">
					  <Ad id="20001" sequence="1" >
						<InLine>
							<Extensions>
								<Extension type="advertiser">
									<Advertiser></Advertiser>
								</Extension>
							</Extensions>
						</InLine>
					  </Ad>
				  </VAST>`,
			wantAdvertisers: nil,
		},
	}
}

func getCreativeIDTestCases() []struct {
	name     string
	vastXML  string
	randomID string
	wantID   string
} {
	return []struct {
		name     string
		vastXML  string
		randomID string
		wantID   string
	}{
		{
			name:     "no creative tag",
			vastXML:  `<VAST><Ad><Wrapper></Wrapper></Ad></VAST>`,
			randomID: "1234",
			wantID:   "cr_1234",
		},
		{
			name:     "creative tag without id",
			vastXML:  `<VAST><Ad><InLine><Creatives><Creative></Creative></Creatives></InLine></Ad></VAST>`,
			randomID: "1234",
			wantID:   "cr_1234",
		},
		{
			name:     "creative tag with id",
			vastXML:  `<VAST><Ad><InLine><Creatives><Creative id="233ff44"></Creative></Creatives></InLine></Ad></VAST>`,
			randomID: "1234",
			wantID:   "233ff44",
		},
	}
}

func getCreativeDurationTestCases() []struct {
	name         string
	vastXML      string
	wantDuration int
	wantErr      error
} {
	return []struct {
		name         string
		vastXML      string
		wantDuration int
		wantErr      error
	}{
		{
			name:         "no_creative_tag",
			vastXML:      `<VAST><Ad><Wrapper></Wrapper></Ad></VAST>`,
			wantDuration: 0,
			wantErr:      errEmptyVideoCreative,
		},
		{
			name:         "creative_tag_without_linear_creative",
			vastXML:      `<VAST><Ad><InLine><Creatives><Creative></Creative></Creatives></InLine></Ad></VAST>`,
			wantDuration: 0,
			wantErr:      errEmptyVideoDuration,
		},
		{
			name:         "creative_tag_without_duration",
			vastXML:      `<VAST><Ad><InLine><Creatives><Creative><Linear></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			wantDuration: 0,
			wantErr:      errEmptyVideoDuration,
		},
		{
			name:         "case_sensitive",
			vastXML:      `<VAST><Ad><InLine><Creatives><Creative><Linear><DURATION>0:0:25</DURATION></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			wantDuration: 0,
			wantErr:      errEmptyVideoDuration,
		},
		{
			name:         "creative_tag_with_duration",
			vastXML:      `<VAST><Ad><InLine><Creatives><Creative><Linear><Duration>0:0:25</Duration></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			wantDuration: 25,
			wantErr:      nil,
		},
		{
			name:         "multiple_linear_tags",
			vastXML:      `<VAST><Ad><InLine><Creatives><Creative><Linear><Duration>0:0:30</Duration></Linear><Linear><Duration>0:0:25</Duration></Linear></Creative></Creatives></InLine></Ad></VAST>`,
			wantDuration: 30,
			wantErr:      nil,
		},
	}
}

func getPricingDetailsTestCases() []struct {
	name         string
	vastXML      string
	wantPrice    float64
	wantCurrency string
} {
	return []struct {
		name         string
		vastXML      string
		wantPrice    float64
		wantCurrency string
	}{
		{
			name:         "vast_2.0_without_extensions",
			vastXML:      `<VAST><Ad><Wrapper></Wrapper></Ad></VAST>`,
			wantPrice:    0,
			wantCurrency: "",
		},
		{
			name:         "vast_2.0_without_extension",
			vastXML:      `<VAST><Ad><Wrapper><Extensions></Extensions></Wrapper></Ad></VAST>`,
			wantPrice:    0,
			wantCurrency: "",
		},
		{
			name:         "vast_2.0_without_price",
			vastXML:      `<VAST><Ad><Wrapper><Extensions><Extension></Extension></Extensions></Wrapper></Ad></VAST>`,
			wantPrice:    0,
			wantCurrency: "",
		},
		{
			name:         "vast_2.0_empty_price",
			vastXML:      `<VAST><Ad><Wrapper><Extensions><Extension><Price></Price></Extension></Extensions></Wrapper></Ad></VAST>`,
			wantPrice:    0,
			wantCurrency: "",
		},
		{
			name:         "vast_2.0_cdata_price",
			vastXML:      `<VAST><Ad><Wrapper><Extensions><Extension><Price><![CDATA[ 12.05 ]]></Price></Extension></Extensions></Wrapper></Ad></VAST>`,
			wantPrice:    12.05,
			wantCurrency: "USD",
		},
		{
			name:         "vast_2.0_price",
			vastXML:      `<VAST><Ad><Wrapper><Extensions><Extension><Price>12.05</Price></Extension></Extensions></Wrapper></Ad></VAST>`,
			wantPrice:    12.05,
			wantCurrency: "USD",
		},
		{
			name:         "vast_2.0_empty_currency",
			vastXML:      `<VAST><Ad><Wrapper><Extensions><Extension><Price currency="">12.05</Price></Extension></Extensions></Wrapper></Ad></VAST>`,
			wantPrice:    12.05,
			wantCurrency: "USD",
		},
		{
			name:         "vast_2.0_inr_currency",
			vastXML:      `<VAST><Ad><Wrapper><Extensions><Extension><Price currency="INR">12.05</Price></Extension></Extensions></Wrapper></Ad></VAST>`,
			wantPrice:    12.05,
			wantCurrency: "INR",
		},
		{
			name:         "vast_gt_2.x_missing_pricing",
			vastXML:      `<VAST version="3.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
			wantPrice:    0,
			wantCurrency: "",
		},
		{
			name:         "vast_gt_2.x_empty_pricing",
			vastXML:      `<VAST version="3.0"><Ad><Wrapper><Pricing></Pricing></Wrapper></Ad></VAST>`,
			wantPrice:    0,
			wantCurrency: "",
		},
		{
			name:         "vast_gt_2.x_cdata_pricing",
			vastXML:      `<VAST version="3.0"><Ad><Wrapper><Pricing><![CDATA[ 12.05 ]]></Pricing></Wrapper></Ad></VAST>`,
			wantPrice:    12.05,
			wantCurrency: "USD",
		},
		{
			name:         "vast_gt_2.x_pricing",
			vastXML:      `<VAST version="3.0"><Ad><Wrapper><Pricing>12.05</Pricing></Wrapper></Ad></VAST>`,
			wantPrice:    12.05,
			wantCurrency: "USD",
		},
		{
			name:         "invalid_price",
			vastXML:      `<VAST version="3.0"><Ad><Wrapper><Pricing>abcd</Pricing></Wrapper></Ad></VAST>`,
			wantPrice:    0,
			wantCurrency: "",
		},
		// TODO: Add test cases.
	}
}

func Test_etreeXMLParser_GetAdvertiser(t *testing.T) {
	for _, tt := range getAdvertiserTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			parser := newETreeXMLParser()
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

func Test_etreeXMLParser_GetCreativeId(t *testing.T) {
	for _, tt := range getCreativeIDTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			generateRandomID = func() string { return tt.randomID }
			parser := newETreeXMLParser()
			err := parser.Parse([]byte(tt.vastXML))
			if !assert.NoError(t, err) {
				return
			}

			gotID := parser.GetCreativeID()
			assert.Equal(t, tt.wantID, gotID)
		})
	}
}

func Test_etreeXMLParser_GetDuration(t *testing.T) {
	for _, tt := range getCreativeDurationTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			parser := newETreeXMLParser()
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

func Test_etreeXMLParser_GetPricingDetails(t *testing.T) {
	for _, tt := range getPricingDetailsTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			parser := newETreeXMLParser()
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
