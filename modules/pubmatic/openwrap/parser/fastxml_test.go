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
				assertFastXMLXMLField(t, adm, models.VideoAdDomainTag, tt.wantDomain)
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
