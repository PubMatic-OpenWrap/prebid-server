package parser

import (
	"testing"

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
			ej := etreeHandler{}
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
			ej := etreeHandler{}
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
			ej := etreeHandler{}
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
			ej := etreeHandler{}
			got := ej.newPricingNode(tt.args.price, tt.args.model, tt.args.currency)
			doc := etree.NewDocument()
			doc.InsertChild(nil, got)
			actual, _ := doc.WriteToString()
			assert.Equal(t, tt.want, actual)
		})
	}
}

func TestAddCategoryTag(t *testing.T) {
	tests := []struct {
		name        string
		inputVAST   string
		adCat       []string
		expectError bool
		wantAdM     string
		wantCat     string
	}{
		{
			name:      "Add_category_to_VAST_Inline",
			inputVAST: `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adCat:     []string{"IAB1-1", "IAB1-2"},
			wantAdM:   `<VAST version="2.0"><Ad><InLine><Category><![CDATA[IAB1-1,IAB1-2]]></Category></InLine></Ad></VAST>`,
			wantCat:   "IAB1-1,IAB1-2",
		},
		{
			name:      "Add_category_to_VAST_Wrapper",
			inputVAST: `<VAST version="2.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
			adCat:     []string{"IAB1-3"},
			wantAdM:   `<VAST version="2.0"><Ad><Wrapper><Category><![CDATA[IAB1-3]]></Category></Wrapper></Ad></VAST>`,
			wantCat:   "IAB1-3",
		},
		{
			name:        "No_category_and_no_VAST_block",
			inputVAST:   `<VAST version="2.0"></VAST>`,
			adCat:       nil,
			expectError: true,
		},
		{
			name:      "Category_already_present_should_not_overwrite_with_inline",
			inputVAST: `<VAST version="2.0"><Ad><InLine><Category><![CDATA[IAB-1]]></Category></InLine></Ad></VAST>`,
			adCat:     []string{"IAB2"},
			wantAdM:   `<VAST version="2.0"><Ad><InLine><Category><![CDATA[IAB-1]]></Category></InLine></Ad></VAST>`,
			wantCat:   "IAB-1",
		},
		{
			name:      "Category_already_present_should_not_overwrite_with_wrapper",
			inputVAST: `<VAST version="2.0"><Ad><Wrapper><Category><![CDATA[IAB-1]]></Category></Wrapper></Ad></VAST>`,
			adCat:     []string{"IAB2"},
			wantAdM:   `<VAST version="2.0"><Ad><Wrapper><Category><![CDATA[IAB-1]]></Category></Wrapper></Ad></VAST>`,
			wantCat:   "IAB-1",
		},
		{
			name:        "No_adCat_passed_no_change",
			inputVAST:   `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adCat:       nil,
			wantCat:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etreeTrackerInjector := &etreeHandler{}
			doc, _ := getETreeNode(tt.inputVAST)
			etreeTrackerInjector.doc = doc
			adm, err := etreeTrackerInjector.AddCategoryTag(tt.adCat)
			assert.Equal(t, tt.expectError, err != nil)
			if tt.expectError {
				return
			}

			var ad *etree.Element
			ad = doc.FindElement("//Ad/InLine")
			if ad == nil {
				ad = doc.FindElement("//Ad/Wrapper")
			}
			if tt.wantAdM != "" {
				assert.Equal(t, tt.wantAdM, adm)
			}
			if tt.wantCat != "" {
				category := ad.FindElement(models.VideoAdCatTag)
				assert.Equal(t, tt.wantCat, category.Text())
			}
		})
	}
}

func TestAddAdvertiserTag(t *testing.T) {
	tests := []struct {
		name        string
		inputVAST   string
		adDomain    string
		expectError bool
		wantDomain  string
		wantAdM     string
	}{
		{
			name:       "Add_domain_to_VAST_Inline",
			inputVAST:  `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adDomain:   "example.com",
			wantDomain: "example.com",
			wantAdM:    `<VAST version="2.0"><Ad><InLine><Advertiser><![CDATA[example.com]]></Advertiser></InLine></Ad></VAST>`,
		},
		{
			name:       "Add_domain_to_VAST_Wrapper",
			inputVAST:  `<VAST version="2.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
			adDomain:   "example.com",
			wantDomain: "example.com",
			wantAdM:    `<VAST version="2.0"><Ad><Wrapper><Advertiser><![CDATA[example.com]]></Advertiser></Wrapper></Ad></VAST>`,
		},
		{
			name:        "No_domain_passed_no_change",
			inputVAST:   `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adDomain:    "",
			expectError: true,
		},
		{
			name:       "Domain_already_present_should_not_overwrite_with_inline",
			inputVAST:  `<VAST version="2.0"><Ad><InLine><Advertiser><![CDATA[example.com]]></Advertiser></InLine></Ad></VAST>`,
			adDomain:   "test.com",
			wantDomain: "example.com",
			wantAdM:    `<VAST version="2.0"><Ad><InLine><Advertiser><![CDATA[example.com]]></Advertiser></InLine></Ad></VAST>`,
		},
		{
			name:       "Domain_already_present_should_not_overwrite_with_wrapper",
			inputVAST:  `<VAST version="2.0"><Ad><Wrapper><Advertiser><![CDATA[example.com]]></Advertiser></Wrapper></Ad></VAST>`,
			adDomain:   "test.com",
			wantDomain: "example.com",
			wantAdM:    `<VAST version="2.0"><Ad><Wrapper><Advertiser><![CDATA[example.com]]></Advertiser></Wrapper></Ad></VAST>`,
		},
		{
			name:        "No_domain_passed_no_change",
			inputVAST:   `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adDomain:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := getETreeNode(tt.inputVAST)
			etreeTrackerInjector := &etreeHandler{doc: doc}
			_, err := etreeTrackerInjector.AddAdvertiserTag(tt.adDomain)
			assert.Equal(t, tt.expectError, err != nil)
			if tt.expectError {
				return
			}

			var ad *etree.Element
			ad = etreeTrackerInjector.doc.FindElement("//Ad/InLine")
			if ad == nil {
				ad = etreeTrackerInjector.doc.FindElement("//Ad/Wrapper")
			}
			if tt.wantAdM != "" {
				actual, _ := etreeTrackerInjector.doc.WriteToString()
				assert.Equal(t, tt.wantAdM, actual)
			}
			if tt.wantDomain != "" {
				advertiser := ad.FindElement(models.VideoAdDomainTag)
				assert.Equal(t, tt.wantDomain, advertiser.Text())
			}
		})
	}
}
