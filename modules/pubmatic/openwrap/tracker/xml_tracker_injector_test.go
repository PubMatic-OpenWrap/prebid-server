package tracker

import (
	"testing"

	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/beevik/etree"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func getETreeNode(tag string) (*etree.Document, []*etree.Element) {
	doc := etree.NewDocument()
	err := doc.ReadFromString(tag)
	if err != nil {
		return nil, nil
	}
	return doc, doc.ChildElements()
}

func getFastXMLTreeNode(tag string) *fastxml.XMLReader {
	doc := fastxml.NewXMLReader()
	err := doc.Parse([]byte(tag))
	if err != nil {
		return nil
	}
	return doc
}

func Test_injectPricingNodeInExtension(t *testing.T) {
	type args struct {
		tag      string
		price    float64
		model    string
		currency string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "pricing_node_missing",
			args: args{
				tag:      `<Impressions/>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Impressions/><Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "extensions_present_pricing_node_missing",
			args: args{
				tag:      `<Extensions/>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "extension_present_pricing_node_missing",
			args: args{
				tag:      `<Extensions><Extension/></Extensions>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Extensions><Extension/><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "override_pricing_cpm",
			args: args{
				tag:      `<Impressions/><Extensions><Extension><Pricing model="CPM" currency="USD">1.23</Pricing></Extension></Extensions>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Impressions/><Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "override_pricing_cpm_add_currency",
			args: args{
				tag:      `<Impressions/><Extensions><Extension><Pricing model="CPM">1.23</Pricing></Extension></Extensions>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Impressions/><Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "override_pricing_cpm_add_attributes",
			args: args{
				tag:      `<Impressions/><Extensions><Extension><Pricing>1.23</Pricing></Extension></Extensions>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Impressions/><Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
		{
			name: "override_pricing_node",
			args: args{
				tag:      `<Impressions/><Extensions><Extension><Pricing model="CPC" currency="INR">1.23</Pricing></Extension></Extensions>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Impressions/><Extensions><Extension><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing></Extension></Extensions>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := getETreeNode(tt.args.tag)
			ej := etreeTrackerInjector{}
			ej.injectPricingNodeInExtension(&doc.Element, tt.args.price, tt.args.model, tt.args.currency)
			actual, _ := doc.WriteToString()
			assert.Equal(t, tt.want, actual)
		})
	}
}

func Test_injectPricingNodeInVAST(t *testing.T) {
	type args struct {
		tag      string
		price    float64
		model    string
		currency string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "override_cpm_pricing",
			args: args{
				tag:      `<Pricing model="CPM" currency="USD">1.23</Pricing>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[4.5]]></Pricing>`,
		},
		{
			name: "override_cpc_pricing",
			args: args{
				tag:      `<Pricing model="CPC" currency="INR">1.23</Pricing>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[4.5]]></Pricing>`,
		},
		{
			name: "add_currency",
			args: args{
				tag:      `<Pricing model="CPM">1.23</Pricing>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[4.5]]></Pricing>`,
		},
		{
			name: "add_all_attributes",
			args: args{
				tag:      `<Pricing>1.23</Pricing>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: "USD",
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[4.5]]></Pricing>`,
		},
		{
			name: "pricing_node_not_present",
			args: args{
				tag:      `<Impressions></Impressions>`,
				price:    4.5,
				model:    models.VideoPricingModelCPM,
				currency: models.VideoPricingCurrencyUSD,
			},
			want: `<Impressions/><Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[4.5]]></Pricing>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := getETreeNode(tt.args.tag)
			ej := etreeTrackerInjector{}
			ej.injectPricingNodeInVAST(&doc.Element, tt.args.price, tt.args.model, tt.args.currency)
			actual, _ := doc.WriteToString()
			assert.Equal(t, tt.want, actual)
		})
	}
}

func Test_updatePricingNode(t *testing.T) {
	type args struct {
		tag      string
		price    float64
		model    string
		currency string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "overwrite_existing_price",
			args: args{
				tag:      `<Pricing>4.5</Pricing>`,
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: models.VideoPricingCurrencyUSD,
			},
			want: `<Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "empty_attributes",
			args: args{
				tag:      `<Pricing>4.5</Pricing>`,
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: models.VideoPricingCurrencyUSD,
			},
			want: `<Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "overwrite_model",
			args: args{
				tag:      `<Pricing model="CPM">4.5</Pricing>`,
				price:    1.2,
				model:    "CPC",
				currency: models.VideoPricingCurrencyUSD,
			},
			want: `<Pricing model="CPC" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "overwrite_currency",
			args: args{
				tag:      `<Pricing currency="USD">4.5</Pricing>`,
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: "INR",
			},
			want: `<Pricing currency="INR" model="` + models.VideoPricingModelCPM + `"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "default_values_attribute",
			args: args{
				tag:      `<Pricing>4.5</Pricing>`,
				price:    1.2,
				model:    "",
				currency: "",
			},
			want: `<Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "adding_space_in_price",
			args: args{
				tag:      `<Pricing>  4.5  </Pricing>`,
				price:    1.2,
				model:    "",
				currency: "",
			},
			want: `<Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "adding_space_in_price_with_cdata",
			args: args{
				tag:      `<Pricing>  <![CDATA[4.5]]>  </Pricing>`,
				price:    1.2,
				model:    "",
				currency: "",
			},
			want: `<Pricing model="` + models.VideoPricingModelCPM + `" currency="` + models.VideoPricingCurrencyUSD + `"><![CDATA[1.2]]></Pricing>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, elements := getETreeNode(tt.args.tag)
			ej := etreeTrackerInjector{}
			ej.updatePricingNode(elements[0], tt.args.price, tt.args.model, tt.args.currency)
			actual, _ := doc.WriteToString()
			assert.Equal(t, tt.want, actual)
		})
	}
}

func Test_newPricingNode(t *testing.T) {
	type args struct {
		price    float64
		model    string
		currency string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "node",
			args: args{
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: models.VideoPricingCurrencyUSD,
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "empty_currency",
			args: args{
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: "",
			},
			want: `<Pricing model="CPM" currency="USD"><![CDATA[1.2]]></Pricing>`,
		},
		{
			name: "other_currency",
			args: args{
				price:    1.2,
				model:    models.VideoPricingModelCPM,
				currency: `INR`,
			},
			want: `<Pricing model="CPM" currency="INR"><![CDATA[1.2]]></Pricing>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ej := etreeTrackerInjector{}
			got := ej.newPricingNode(tt.args.price, tt.args.model, tt.args.currency)
			doc := etree.NewDocument()
			doc.InsertChild(nil, got)
			actual, _ := doc.WriteToString()
			assert.Equal(t, tt.want, actual)
		})
	}
}

func TestUpdateADMWithAdvCat(t *testing.T) {
	tests := []struct {
		name        string
		inputVAST   string
		adDomain    string
		adCat       []string
		expectError bool
		wantDomain  string
		wantCat     string
	}{
		{
			name:       "VAST_Inline_both_adv_and_cat_not_present_in_VAST_But_present_in_bidresp",
			inputVAST:  `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adDomain:   "example.com",
			adCat:      []string{"IAB1-1", "IAB1-2"},
			wantDomain: "example.com",
			wantCat:    "IAB1-1,IAB1-2",
		},
		{
			name:       "VAST_Wrapper_both_adv_and_cat_not_present_in_VAST_But_present_in_bidresp",
			inputVAST:  `<VAST version="2.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
			adDomain:   "advertiser.net",
			adCat:      []string{"IAB1-1", "IAB1-2"},
			wantDomain: "advertiser.net",
			wantCat:    "IAB1-1,IAB1-2",
		},
		{
			name:       "VAST_Wrapper_both_adv_and_cat_not_present_in_VAST_and_adv_present_in_bidresp",
			inputVAST:  `<VAST version="2.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
			adDomain:   "advertiser.net",
			adCat:      nil,
			wantDomain: "advertiser.net",
			wantCat:    "",
		},
		{
			name:       "VAST_Inline_only_category_present",
			inputVAST:  `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adDomain:   "",
			adCat:      []string{"IAB3-1"},
			wantDomain: "",
			wantCat:    "IAB3-1",
		},
		{
			name:        "VAST_Inline_both_domain_and_category_empty",
			inputVAST:   `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adDomain:    "",
			adCat:       nil,
			expectError: true,
		},
		{
			name:       "VAST_Inline_both_domain_and_category_present_in_VAST_and_bidResp",
			inputVAST:  `<VAST version="2.0"><Ad><InLine><Category>IAB1-1,IAB1-2</Category><Advertiser>example.com</Advertiser></InLine></Ad></VAST>`,
			adDomain:   "test.com",
			adCat:      []string{"IAB1-3", "IAB1-4"},
			wantDomain: "example.com",
			wantCat:    "IAB1-1,IAB1-2",
		},
		{
			name:       "VAST_Wrapper_both_domain_and_category_present_in_VAST_and_bidResp",
			inputVAST:  `<VAST version="2.0"><Ad><Wrapper><Category>IAB1-1,IAB1-2</Category><Advertiser>example.com</Advertiser></Wrapper></Ad></VAST>`,
			adDomain:   "test.com",
			adCat:      []string{"IAB1-3", "IAB1-4"},
			wantDomain: "example.com",
			wantCat:    "IAB1-1,IAB1-2",
		},
		{
			name:       "VAST_Wrapper_adm_is_url",
			inputVAST:  `<VAST version="3.0"><Ad id="1"><Wrapper><AdSystem>PubMatic Wrapper</AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error></Wrapper></Ad></VAST>`,
			adDomain:   "advertiser.net",
			adCat:      []string{"IAB1-1", "IAB1-2"},
			wantDomain: "advertiser.net",
			wantCat:    "IAB1-1,IAB1-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := getETreeNode(tt.inputVAST)

			ti := &etreeTrackerInjector{
				doc: doc,
			}

			output, err := ti.UpdateADMWithAdvCat(tt.adDomain, tt.adCat)
			assert.Equal(t, tt.expectError, err != nil)
			if tt.expectError {
				return
			}

			newDoc, _ := getETreeNode(output)

			ad := newDoc.FindElement("//Ad/InLine")
			if ad == nil {
				ad = newDoc.FindElement("//Ad/Wrapper")
			}

			if tt.wantDomain != "" {
				d := ad.FindElement(models.VideoAdDomainTag)
				assert.Equal(t, tt.wantDomain, d.Text())
			}

			if tt.wantCat != "" {
				c := ad.FindElement(models.VideoAdCatTag)
				assert.Equal(t, tt.wantCat, c.Text())
			}
		})
	}
}
func TestFastXMLTrackerInjector_UpdateADMWithAdvCat(t *testing.T) {
	tests := []struct {
		name        string
		inputVAST   string
		adDomain    string
		adCat       []string
		expectError bool
		wantDomain  string
		wantCat     string
	}{
		{
			name:       "VAST_Inline_both_adv_and_cat_not_present_in_VAST_But_present_in_bidresp",
			inputVAST:  `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adDomain:   "example.com",
			adCat:      []string{"IAB1-1", "IAB1-2"},
			wantDomain: "example.com",
			wantCat:    "IAB1-1,IAB1-2",
		},
		{
			name:       "VAST_Wrapper_both_adv_and_cat_not_present_in_VAST_But_present_in_bidresp",
			inputVAST:  `<VAST version="2.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
			adDomain:   "advertiser.net",
			adCat:      []string{"IAB1-1", "IAB1-2"},
			wantDomain: "advertiser.net",
			wantCat:    "IAB1-1,IAB1-2",
		},
		{
			name:       "VAST_Wrapper_both_adv_and_cat_not_present_in_VAST_and_adv_present_in_bidresp",
			inputVAST:  `<VAST version="2.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
			adDomain:   "advertiser.net",
			adCat:      nil,
			wantDomain: "advertiser.net",
			wantCat:    "",
		},
		{
			name:       "VAST_Inline_only_category_present",
			inputVAST:  `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adDomain:   "",
			adCat:      []string{"IAB3-1"},
			wantDomain: "",
			wantCat:    "IAB3-1",
		},
		{
			name:        "VAST_Inline_both_domain_and_category_empty",
			inputVAST:   `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adDomain:    "",
			adCat:       nil,
			expectError: true,
		},
		{
			name:       "VAST_Inline_both_domain_and_category_present_in_VAST_and_bidResp",
			inputVAST:  `<VAST version="2.0"><Ad><InLine><Category>IAB1-1,IAB1-2</Category><Advertiser>example.com</Advertiser></InLine></Ad></VAST>`,
			adDomain:   "test.com",
			adCat:      []string{"IAB1-3", "IAB1-4"},
			wantDomain: "example.com",
			wantCat:    "IAB1-1,IAB1-2",
		},
		{
			name:       "VAST_Wrapper_both_domain_and_category_present_in_VAST_and_bidResp",
			inputVAST:  `<VAST version="2.0"><Ad><Wrapper><Category>IAB1-1,IAB1-2</Category><Advertiser>example.com</Advertiser></Wrapper></Ad></VAST>`,
			adDomain:   "test.com",
			adCat:      []string{"IAB1-3", "IAB1-4"},
			wantDomain: "example.com",
			wantCat:    "IAB1-1,IAB1-2",
		},
		{
			name:       "VAST_Wrapper_adm_is_url",
			inputVAST:  `<VAST version="3.0"><Ad id="1"><Wrapper><AdSystem>PubMatic Wrapper</AdSystem><VASTAdTagURI><![CDATA[https://stagingnyc.pubmatic.com:8443/test/pub_vast.xml]]></VASTAdTagURI><Impression><![CDATA[Tracking URL]]></Impression><Error><![CDATA[Error URL]]></Error></Wrapper></Ad></VAST>`,
			adDomain:   "advertiser.net",
			adCat:      []string{"IAB1-1", "IAB1-2"},
			wantDomain: "advertiser.net",
			wantCat:    "IAB1-1,IAB1-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := getFastXMLTreeNode(tt.inputVAST)

			ti := &fastXMLTrackerInjector{
				doc: doc,
				xu:  fastxml.NewXMLUpdater(doc, fastxml.WriteSettings{CDATAWrap: true, CompressWhitespace: true}),
			}
			ti.vastTag = doc.SelectElement(nil, "VAST")

			output, err := ti.UpdateADMWithAdvCat(tt.adDomain, tt.adCat)
			assert.Equal(t, tt.expectError, err != nil)
			if tt.expectError {
				return
			}

			newDoc := getFastXMLTreeNode(output)

			vast := newDoc.SelectElement(nil, "VAST")
			assert.NotNil(t, vast)

			ad := newDoc.SelectElement(vast, "Ad")
			assert.NotNil(t, ad)

			adType := newDoc.SelectElement(ad, "InLine")
			if adType == nil {
				adType = newDoc.SelectElement(ad, "Wrapper")
			}
			assert.NotNil(t, adType)

			if tt.wantDomain != "" {
				dElem := newDoc.SelectElement(adType, models.VideoAdDomainTag)
				if assert.NotNil(t, dElem) {
					text := newDoc.Text(dElem)
					assert.Equal(t, tt.wantDomain, text)
				}
			}

			if tt.wantCat != "" {
				cElem := newDoc.SelectElement(adType, models.VideoAdCatTag)
				if assert.NotNil(t, cElem) {
					text := newDoc.Text(cElem)
					assert.Equal(t, tt.wantCat, text)
				}
			}
		})
	}
}
