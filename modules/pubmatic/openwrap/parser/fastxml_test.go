package parser

import (
	"testing"

	"github.com/PubMatic-OpenWrap/fastxml"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func getFastXMLTreeNode(tag string) *fastxml.XMLReader {
	doc := fastxml.NewXMLReader()
	err := doc.Parse([]byte(tag))
	if err != nil {
		return nil
	}
	return doc
}

func TestAddCategoryTagFastXML(t *testing.T) {
	tests := []struct {
		name        string
		inputVAST   string
		adCat       []string
		expectError bool
		wantCat     string
		wantAdM     string
	}{
		{
			name:      "Add_category_to_VAST_Inline",
			inputVAST: `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
			adCat:     []string{"IAB1-1", "IAB1-2"},
			wantCat:   "IAB1-1,IAB1-2",
			wantAdM:   `<VAST version="2.0"><Ad><InLine><Category><![CDATA[IAB1-1,IAB1-2]]></Category></InLine></Ad></VAST>`,
		},
		{
			name:      "Add_category_to_VAST_Wrapper",
			inputVAST: `<VAST version="2.0"><Ad><Wrapper></Wrapper></Ad></VAST>`,
			adCat:     []string{"IAB1-3"},
			wantCat:   "IAB1-3",
			wantAdM:   `<VAST version="2.0"><Ad><Wrapper><Category><![CDATA[IAB1-3]]></Category></Wrapper></Ad></VAST>`,
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
			wantCat:   "IAB-1",
			wantAdM:   `<VAST version="2.0"><Ad><InLine><Category><![CDATA[IAB-1]]></Category></InLine></Ad></VAST>`,
		},
		{
			name:      "Category_already_present_should_not_overwrite_with_wrapper",
			inputVAST: `<VAST version="2.0"><Ad><Wrapper><Category><![CDATA[IAB-1]]></Category></Wrapper></Ad></VAST>`,
			adCat:     []string{"IAB2"},
			wantCat:   "IAB-1",
			wantAdM:   `<VAST version="2.0"><Ad><Wrapper><Category><![CDATA[IAB-1]]></Category></Wrapper></Ad></VAST>`,
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
			doc := getFastXMLTreeNode(tt.inputVAST)
			fastxmlInjector := &FastXMLHandler{
				doc:     doc,
				xu:      fastxml.NewXMLUpdater(doc, fastxml.WriteSettings{CDATAWrap: true, CompressWhitespace: true}),
				vastTag: doc.SelectElement(nil, "VAST"),
			}

			adm, err := fastxmlInjector.AddCategoryTag(tt.adCat)

			assert.Equal(t, tt.expectError, err != nil)
			if tt.wantAdM != "" {
				assert.Equal(t, tt.wantAdM, adm)
			}
			if tt.wantCat != "" {
				assertFastXMLXMLField(t, adm, models.VideoAdCatTag, tt.wantCat)
			}
		})
	}
}

func TestAddAdvertiserTagFastXML(t *testing.T) {
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
			wantAdM:     `<VAST version="2.0"><Ad><InLine></InLine></Ad></VAST>`,
		},
		{
			name:       "Domain_already_present_should_not_overwrite_with_inline",
			inputVAST:  `<VAST version="2.0"><Ad><InLine><Advertiser>example.com</Advertiser></InLine></Ad></VAST>`,
			adDomain:   "test.com",
			wantDomain: "example.com",
			wantAdM:    `<VAST version="2.0"><Ad><InLine><Advertiser>example.com</Advertiser></InLine></Ad></VAST>`,
		},
		{
			name:       "Domain_already_present_should_not_overwrite_with_wrapper",
			inputVAST:  `<VAST version="2.0"><Ad><Wrapper><Advertiser>example.com</Advertiser></Wrapper></Ad></VAST>`,
			adDomain:   "test.com",
			wantDomain: "example.com",
			wantAdM:    `<VAST version="2.0"><Ad><Wrapper><Advertiser>example.com</Advertiser></Wrapper></Ad></VAST>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := getFastXMLTreeNode(tt.inputVAST)
			fastxmlInjector := &FastXMLHandler{
				doc:     doc,
				xu:      fastxml.NewXMLUpdater(doc, fastxml.WriteSettings{CDATAWrap: true, CompressWhitespace: true}),
				vastTag: doc.SelectElement(nil, "VAST"),
			}

			adm, err := fastxmlInjector.AddAdvertiserTag(tt.adDomain)
			assert.Equal(t, tt.expectError, err != nil)
			if tt.wantDomain != "" {
				assertFastXMLXMLField(t, adm, models.VideoAdvertiserTag, tt.wantDomain)
			}
		})
	}
}

func assertFastXMLXMLField(t *testing.T, adm string, keyName string, expectedValue string) bool {
	newDoc := getFastXMLTreeNode(adm)
	vast := newDoc.SelectElement(nil, "VAST")
	assert.NotNil(t, vast)

	ad := newDoc.SelectElement(vast, "Ad")
	assert.NotNil(t, ad)

	adType := newDoc.SelectElement(ad, "InLine")
	if adType == nil {
		adType = newDoc.SelectElement(ad, "Wrapper")
	}
	assert.NotNil(t, adType)

	actualValue := newDoc.SelectElement(adType, keyName)
	return assert.Equal(t, expectedValue, newDoc.Text(actualValue))
}

func TestExtractCTAOverlayFromVAST(t *testing.T) {
	tests := []struct {
		name     string
		adm      string
		wantStr  string
		parseErr bool
	}{
		{
			name:     "empty_adm_parse_fails",
			adm:      "",
			parseErr: true,
		},
		{
			name:    "VAST_2.0_returns_empty_version_unsupported",
			adm:     `<VAST version="2.0"><Ad><InLine><AdSystem>Test</AdSystem><Creatives><Creative><CreativeExtensions><CreativeExtension name="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"delay":0}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantStr: "",
		},
		{
			name:    "VAST_no_version_returns_empty",
			adm:     `<VAST><Ad><InLine><Creatives><Creative><CreativeExtensions><CreativeExtension name="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantStr: "",
		},
		{
			name:    "VAST_3.0_InLine_name_PubMatic_returns_first_inner_text",
			adm:     `<VAST version="3.0"><Ad><InLine><AdSystem>Test</AdSystem><Creatives><Creative><CreativeExtensions><CreativeExtension name="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"delay":0,"pos":1}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantStr: `{"ctaoverlay":{"delay":0,"pos":1}}`,
		},
		{
			name:    "VAST_3.0_no_CreativeExtensions_returns_empty",
			adm:     `<VAST version="3.0"><Ad><InLine><AdSystem>Test</AdSystem><Creatives><Creative></Creative></Creatives></InLine></Ad></VAST>`,
			wantStr: "",
		},
		{
			name:    "VAST_3.0_CreativeExtension_without_name_PubMatic_returns_empty",
			adm:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><CreativeExtensions><CreativeExtension type="application/json"><![CDATA[{"ctaoverlay":{"x":1}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantStr: "",
		},
		{
			name:    "VAST_3.0_Wrapper_only_returns_empty_InLine_required",
			adm:     `<VAST version="3.0"><Ad><Wrapper><AdSystem>Test</AdSystem><VASTAdTagURI><![CDATA[https://example.com]]></VASTAdTagURI><Creatives><Creative><CreativeExtensions><CreativeExtension name="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantStr: "",
		},
		{
			name:    "VAST_3.0_multiple_PubMatic_extensions_returns_first_only",
			adm:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><CreativeExtensions><CreativeExtension name="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"first":1}}]]></CreativeExtension><CreativeExtension name="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"second":2}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantStr: `{"ctaoverlay":{"first":1}}`,
		},
		{
			name:    "VAST_4.1_InLine_name_PubMatic_returns_inner_text",
			adm:     `<VAST version="4.1"><Ad><InLine><Creatives><Creative><CreativeExtensions><CreativeExtension name="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"delay":0}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantStr: `{"ctaoverlay":{"delay":0}}`,
		},
		{
			name:    "VAST_3.0_name_pubmatic_lowercase_matched",
			adm:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><CreativeExtensions><CreativeExtension name="pubmatic" type="application/json"><![CDATA[{"ctaoverlay":{"x":1}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantStr: `{"ctaoverlay":{"x":1}}`,
		},
		{
			name:    "VAST_3.0_name_PUBMATIC_uppercase_matched",
			adm:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><CreativeExtensions><CreativeExtension name="PUBMATIC" type="application/json"><![CDATA[{"ctaoverlay":{"y":2}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantStr: `{"ctaoverlay":{"y":2}}`,
		},
		{
			name:    "VAST_3.0_name_PubMatic_matched",
			adm:     `<VAST version="3.0"><Ad><InLine><Creatives><Creative><CreativeExtensions><CreativeExtension name="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"n":4}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantStr: `{"ctaoverlay":{"n":4}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &FastXMLHandler{}
			err := h.Parse(tt.adm)
			assert.Equal(t, tt.parseErr, err != nil, "parse error")
			if err != nil {
				return
			}
			gotStr := h.ExtractCTAOverlayFromVAST()
			assert.Equal(t, tt.wantStr, gotStr)
			if gotStr != "" {
				assert.NotContains(t, gotStr, "<![CDATA[", "returned content must be CDATA-trimmed")
				assert.NotContains(t, gotStr, "]]>", "returned content must be CDATA-trimmed")
			}
		})
	}
}

func TestVastVersionSupportsCreativeExtensions(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"3.0", true},
		{"3.1", true},
		{"4.0", true},
		{"4.1", true},
		{"2.0", false},
		{"2.1", false},
		{"1.0", false},
		{"", false},
		{" 3.0 ", true},
		{"  2.0  ", false},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := VastVersionSupportsCreativeExtensions(tt.version)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Parser-level CTA overlay benchmarks (Parse + ExtractCTAOverlayFromVAST; name=PubMatic).
var (
	parserBenchVASTHit  = `<VAST version="3.0"><Ad><InLine><AdSystem>Test</AdSystem><Creatives><Creative><CreativeExtensions><CreativeExtension name="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"delay":0,"pos":1}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`
	parserBenchVASTMiss = `<VAST version="2.0"><Ad><InLine><AdSystem>Test</AdSystem><Creatives><Creative><CreativeExtensions><CreativeExtension name="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"delay":0}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`
)

// BenchmarkExtractCTAOverlayFromVAST_Parser_Hit measures Parse + ExtractCTAOverlayFromVAST when overlay is present.
func BenchmarkExtractCTAOverlayFromVAST_Parser_Hit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		h := &FastXMLHandler{}
		_ = h.Parse(parserBenchVASTHit)
		_ = h.ExtractCTAOverlayFromVAST()
	}
}

// BenchmarkExtractCTAOverlayFromVAST_Parser_MissVersion measures Parse + Extract when version is unsupported (early return).
func BenchmarkExtractCTAOverlayFromVAST_Parser_MissVersion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		h := &FastXMLHandler{}
		_ = h.Parse(parserBenchVASTMiss)
		_ = h.ExtractCTAOverlayFromVAST()
	}
}
