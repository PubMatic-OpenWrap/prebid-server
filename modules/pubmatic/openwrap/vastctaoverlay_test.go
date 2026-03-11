package openwrap

import (
	"encoding/json"
	"testing"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/parser"
	"github.com/stretchr/testify/assert"
)

func TestExtractCTAOverlayFromVAST(t *testing.T) {
	tests := []struct {
		name    string
		adm     string
		wantVal interface{}
		wantOk  bool
	}{
		{
			name:    "empty_adm",
			adm:     "",
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "invalid_XML",
			adm:     "not xml",
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "VAST_2.0_with_CreativeExtension_id=PubMatic_is_skipped_(CreativeExtensions_only_in_3.0+)",
			adm:     `<VAST version="2.0"><Ad><InLine><AdSystem>Test</AdSystem><AdTitle></AdTitle><Impression></Impression><Creatives><Creative><CreativeExtensions><CreativeExtension id="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"delay":0}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "VAST_with_no_version_attribute_is_skipped",
			adm:     `<VAST><Ad><InLine><AdSystem>Test</AdSystem><AdTitle></AdTitle><Impression></Impression><Creatives><Creative><CreativeExtensions><CreativeExtension id="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"delay":0}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "VAST_InLine_with_no_Creatives",
			adm:     `<VAST version="3.0"><Ad><InLine><AdSystem>Test</AdSystem><AdTitle></AdTitle><Impression></Impression><Creatives></Creatives></InLine></Ad></VAST>`,
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "VAST_Wrapper_with_no_CreativeExtensions_in_creatives",
			adm:     `<VAST version="3.0"><Ad><Wrapper><AdSystem>Test</AdSystem><VASTAdTagURI><![CDATA[https://example.com]]></VASTAdTagURI><Impression></Impression><Creatives></Creatives></Wrapper></Ad></VAST>`,
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "Wrapper_Creative_with_CreativeExtensions_(rs/vast_CreativeWrapper_has_no_CreativeExtensions;_only_InLine_supported)",
			adm:     `<VAST version="3.0"><Ad><Wrapper><AdSystem>Test</AdSystem><VASTAdTagURI><![CDATA[https://example.com]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><CreativeExtensions><CreativeExtension type="application/json"><![CDATA[{"ctaoverlay":{"delay":0,"pos":2,"ctacopy":"Learn More"}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "InLine_CreativeExtension_without_id=PubMatic_is_ignored",
			adm:     `<VAST version="3.0"><Ad><InLine><AdSystem>Test</AdSystem><AdTitle></AdTitle><Impression></Impression><Creatives><Creative><CreativeExtensions><CreativeExtension type="application/json"><![CDATA[{"ctaoverlay":{"delay":0,"pos":1,"ctacopy":"Learn More"}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantVal: nil,
			wantOk:  false,
		},
		{
			name: "InLine_CreativeExtension_id=PubMatic_returns_first_ctaoverlay",
			adm:  `<VAST version="3.0"><Ad><InLine><AdSystem>Test</AdSystem><AdTitle></AdTitle><Impression></Impression><Creatives><Creative><CreativeExtensions><CreativeExtension id="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"delay":0,"pos":1,"ctacopy":"Learn More"}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantVal: map[string]interface{}{
				"delay":   float64(0),
				"pos":     float64(1),
				"ctacopy": "Learn More",
			},
			wantOk: true,
		},
		{
			name: "InLine_CreativeExtension_id=PubMatic_with_multi-line_JSON_in_CDATA_(example_format)",
			adm: `<VAST version="3.0"><Ad><InLine><AdSystem>Test</AdSystem><AdTitle></AdTitle><Impression></Impression><Creatives><Creative><CreativeExtensions><CreativeExtension id="PubMatic" type="application/json"><![CDATA[
      {
        "ctaoverlay": {
          "delay": 0,
          "pos": 1,
          "ctacopy": "Learn More"
        }
      }
      ]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantVal: map[string]interface{}{
				"delay":   float64(0),
				"pos":     float64(1),
				"ctacopy": "Learn More",
			},
			wantOk: true,
		},
		{
			name: "InLine_multiple_CreativeExtensions_with_id=PubMatic_returns_first",
			adm:  `<VAST version="3.0"><Ad><InLine><AdSystem>Test</AdSystem><AdTitle></AdTitle><Impression></Impression><Creatives><Creative><CreativeExtensions><CreativeExtension id="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"first":1}}]]></CreativeExtension><CreativeExtension id="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"second":2}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantVal: map[string]interface{}{
				"first": float64(1),
			},
			wantOk: true,
		},
		{
			// Same format/location as production: Creative→CreativeExtensions→CreativeExtension id=PubMatic;
			// CDATA with JSON object containing "ctaoverlay" (spaces in JSON and full field set supported).
			name: "InLine_CreativeExtension_id=PubMatic_full_ctaoverlay_format_(delay_endcarddelay_pos_ctacopy_etc)",
			adm: `<VAST version="3.0"><Ad><InLine><AdSystem>Test</AdSystem><AdTitle></AdTitle><Impression></Impression><Creatives><Creative><CreativeExtensions><CreativeExtension id ="PubMatic" type="application/json"><![CDATA[
{"ctaoverlay" : {"delay" : 0,"endcarddelay" : 0,"dismissible" : 0,"pos" : 1,"ctacopy" : "Add To Cart","ctabuttonbgcolor" : "#ffa41d","ctacopycolor" : "#000000","iconimageurl" : "abc","header" : "App Store","title" : "Amazon Shopping","description" : "Grab Prime Deals","clickurl" : "clickurl","clicktrackers" : ["click1"]}}
]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantVal: map[string]interface{}{
				"delay": float64(0), "endcarddelay": float64(0), "dismissible": float64(0), "pos": float64(1),
				"ctacopy": "Add To Cart", "ctabuttonbgcolor": "#ffa41d", "ctacopycolor": "#000000",
				"iconimageurl": "abc", "header": "App Store", "title": "Amazon Shopping",
				"description": "Grab Prime Deals", "clickurl": "clickurl",
				"clicktrackers": []interface{}{"click1"},
			},
			wantOk: true,
		},
		{
			name:    "Wrapper_CreativeExtension_id=PubMatic_is_ignored_(InLine_only)",
			adm:     `<VAST version="3.0"><Ad><Wrapper><AdSystem>Test</AdSystem><VASTAdTagURI><![CDATA[https://example.com]]></VASTAdTagURI><Impression></Impression><Creatives><Creative><CreativeExtensions><CreativeExtension id="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"delay":0}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></Wrapper></Ad></VAST>`,
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "Invalid_JSON_in_CreativeExtension_id=PubMatic_is_ignored_(no_ctaoverlay)",
			adm:     `<VAST version="3.0"><Ad><InLine><AdSystem>Test</AdSystem><AdTitle></AdTitle><Impression></Impression><Creatives><Creative><CreativeExtensions><CreativeExtension id="PubMatic" type="application/json"><![CDATA[not valid json]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "Invalid_JSON_in_first_PubMatic_extension_ignored,_second_valid_returned",
			adm:     `<VAST version="3.0"><Ad><InLine><AdSystem>Test</AdSystem><AdTitle></AdTitle><Impression></Impression><Creatives><Creative><CreativeExtensions><CreativeExtension id="PubMatic" type="application/json"><![CDATA[not json]]></CreativeExtension><CreativeExtension id="PubMatic" type="application/json"><![CDATA[{"ctaoverlay":{"ok":1}}]]></CreativeExtension></CreativeExtensions></Creative></Creatives></InLine></Ad></VAST>`,
			wantVal: map[string]interface{}{"ok": float64(1)},
			wantOk:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRaw, gotOk := ExtractCTAOverlayFromVASTFastXML(tt.adm)
			assert.Equal(t, tt.wantOk, gotOk, "ExtractCTAOverlayFromVASTFastXML ok")
			if tt.wantOk {
				var got interface{}
				assert.NoError(t, json.Unmarshal(gotRaw, &got), "ctaoverlay JSON must be valid")
				assert.Equal(t, tt.wantVal, got, "ExtractCTAOverlayFromVASTFastXML value")
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
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := parser.VastVersionSupportsCreativeExtensions(tt.version)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsVideoBidEligibleForCTAOverlay(t *testing.T) {
	tests := []struct {
		name                string
		bidExt              *models.BidExt
		ctaOverlayRequested bool
		displayManagerVer   string
		want                bool
	}{
		{
			name: "eligible:_video,_imp_owsdk.ctaoverlay=1,_bid.owsdk.ctaoverlay_absent,_sdk_4.10.0",
			bidExt: &models.BidExt{
				CreativeType: models.MediaTypeVideo,
				OWSDK:        nil,
			},
			ctaOverlayRequested: true,
			displayManagerVer:   "4.10.0",
			want:                true,
		},
		{
			name: "eligible:_video,_imp_ctaoverlay=1,_sdk_4.9.0",
			bidExt: &models.BidExt{
				CreativeType: models.MediaTypeVideo,
			},
			ctaOverlayRequested: true,
			displayManagerVer:   "4.9.0",
			want:                true,
		},
		{
			name: "eligible:_video,_imp_ctaoverlay=1,_sdk_4.9.1_(Android)",
			bidExt: &models.BidExt{
				CreativeType: models.MediaTypeVideo,
			},
			ctaOverlayRequested: true,
			displayManagerVer:   "4.9.1",
			want:                true,
		},
		{
			name: "eligible:_video,_imp_ctaoverlay=1,_sdk_4.11.0",
			bidExt: &models.BidExt{
				CreativeType: models.MediaTypeVideo,
			},
			ctaOverlayRequested: true,
			displayManagerVer:   "4.11.0",
			want:                true,
		},
		{
			name: "not_eligible:_bid.ext.owsdk.ctaoverlay_already_present",
			bidExt: &models.BidExt{
				CreativeType: models.MediaTypeVideo,
				OWSDK:        map[string]any{models.CTAOVERLAY: map[string]any{"delay": 0}},
			},
			ctaOverlayRequested: true,
			displayManagerVer:   "4.10.0",
			want:                false,
		},
		{
			name: "not_eligible:_imp_owsdk.ctaoverlay_not_1",
			bidExt: &models.BidExt{
				CreativeType: models.MediaTypeVideo,
			},
			ctaOverlayRequested: false,
			displayManagerVer:   "4.10.0",
			want:                false,
		},
		{
			name: "not_eligible:_imp_owsdk.ctaoverlay_missing",
			bidExt: &models.BidExt{
				CreativeType: models.MediaTypeVideo,
			},
			ctaOverlayRequested: false,
			displayManagerVer:   "4.10.0",
			want:                false,
		},
		{
			name: "not_eligible:_sdk_4.8.0",
			bidExt: &models.BidExt{
				CreativeType: models.MediaTypeVideo,
			},
			ctaOverlayRequested: true,
			displayManagerVer:   "4.8.0",
			want:                false,
		},
		{
			name: "not_eligible:_sdk_4.9.2_(not_in_allowed_list)",
			bidExt: &models.BidExt{
				CreativeType: models.MediaTypeVideo,
			},
			ctaOverlayRequested: true,
			displayManagerVer:   "4.9.2",
			want:                false,
		},
		{
			name: "not_eligible:_sdk_4.12.0",
			bidExt: &models.BidExt{
				CreativeType: models.MediaTypeVideo,
			},
			ctaOverlayRequested: true,
			displayManagerVer:   "4.12.0",
			want:                false,
		},
		{
			name: "not_eligible:_banner",
			bidExt: &models.BidExt{
				CreativeType: "banner",
			},
			ctaOverlayRequested: true,
			displayManagerVer:   "4.10.0",
			want:                false,
		},
		{
			name:                "not_eligible:_nil_bidExt",
			bidExt:              nil,
			ctaOverlayRequested: true,
			displayManagerVer:   "4.10.0",
			want:                false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsVideoBidEligibleForCTAOverlay(tt.bidExt, tt.ctaOverlayRequested, tt.displayManagerVer)
			assert.Equal(t, tt.want, got)
		})
	}
}
