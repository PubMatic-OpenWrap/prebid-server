package floors

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func TestPrepareRuleCombinations(t *testing.T) {
	tt := []struct {
		name string
		in   []string
		n    int
		del  string
		out  []string
	}{
		{
			name: "Schema items, n = 1",
			in:   []string{"A"},
			n:    1,
			del:  "|",
			out: []string{
				"a",
				"*",
			},
		},
		{
			name: "Schema items, n = 2",
			in:   []string{"A", "B"},
			n:    2,
			del:  "|",
			out: []string{
				"a|b",
				"a|*",
				"*|b",
				"*|*",
			},
		},
		{
			name: "Schema items, n = 3",
			in:   []string{"A", "B", "C"},
			n:    3,
			del:  "|",
			out: []string{
				"a|b|c",
				"a|b|*",
				"a|*|c",
				"*|b|c",
				"a|*|*",
				"*|b|*",
				"*|*|c",
				"*|*|*",
			},
		},
		{
			name: "Schema items, n = 4",
			in:   []string{"A", "B", "C", "D"},
			n:    4,
			del:  "|",
			out: []string{
				"a|b|c|d",
				"a|b|c|*",
				"a|b|*|d",
				"a|*|c|d",
				"*|b|c|d",
				"a|b|*|*",
				"a|*|c|*",
				"a|*|*|d",
				"*|b|c|*",
				"*|b|*|d",
				"*|*|c|d",
				"a|*|*|*",
				"*|b|*|*",
				"*|*|c|*",
				"*|*|*|d",
				"*|*|*|*",
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := prepareRuleCombinations(tc.in, tc.n, tc.del)
			if !reflect.DeepEqual(out, tc.out) {
				t.Errorf("error: \nreturn:\t%v\nwant:\t%v", out, tc.out)
			}
		})
	}
}

func TestUpdateImpExtWithFloorDetails(t *testing.T) {
	tt := []struct {
		name         string
		matchedRule  string
		floorRuleVal float64
		imp          openrtb2.Imp
		expected     json.RawMessage
	}{
		{
			name:         "Nil ImpExt",
			matchedRule:  "test|123|xyz",
			floorRuleVal: 5.5,
			imp:          openrtb2.Imp{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}},
			expected:     json.RawMessage(`{"prebid":{"floors":{"floorRule":"test|123|xyz","floorRuleValue":5.5}}}`),
		},
		{
			name:         "Empty ImpExt",
			matchedRule:  "test|123|xyz",
			floorRuleVal: 5.5,
			imp:          openrtb2.Imp{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}, Ext: json.RawMessage{}},
			expected:     json.RawMessage(`{"prebid":{"floors":{"floorRule":"test|123|xyz","floorRuleValue":5.5}}}`),
		},
		{
			name:         "With prebid Ext",
			matchedRule:  "banner|www.test.com|*",
			floorRuleVal: 5.500123,
			imp:          openrtb2.Imp{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}, Ext: []byte(`{"prebid": {"test": true}}`)},
			expected:     []byte(`{"prebid":{"floors":{"floorRule":"banner|www.test.com|*","floorRuleValue":5.5001}}}`),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			iw := &openrtb_ext.ImpWrapper{Imp: &tc.imp}
			updateImpExtWithFloorDetails(tc.matchedRule, iw, tc.floorRuleVal)
			if tc.imp.Ext != nil && !reflect.DeepEqual(tc.imp.Ext, tc.expected) {
				t.Errorf("error: \nreturn:\t%v\n want:\t%v", string(tc.imp.Ext), string(tc.expected))
			}
		})
	}
}

func TestCreateRuleKeys(t *testing.T) {
	tt := []struct {
		name        string
		floorSchema openrtb_ext.PriceFloorSchema
		request     *openrtb2.BidRequest
		out         []string
	}{
		{
			name: "CreateRule with banner mediatype, size and domain",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.test.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "size", "domain"}},
			out:         []string{"banner", "300x250", "www.test.com"},
		},
		{
			name: "CreateRule with video mediatype, size and domain",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.test.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 640, H: 480}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "size", "domain"}},
			out:         []string{"video", "640x480", "www.test.com"},
		},
		{
			name: "CreateRule with video mediatype, size and domain",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.test.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "size", "domain"}},
			out:         []string{"video", "300x250", "www.test.com"},
		},
		{
			name: "CreateRule with Audio mediatype, country and deviceType (Phone)",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.test.com",
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Audio: &openrtb2.Audio{MinDuration: 10}}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "USA"}, UA: "Phone"},
				Ext:    json.RawMessage(`{"prebid":{"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["mediaType","country","deviceType"]},"values":{"audio|USA|phone":1.01,"*|*|*":16.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "country", "deviceType"}},
			out:         []string{"audio", "USA", "phone"},
		},
		{
			name: "CreateRule with channel, country and deviceType",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.test.com",
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Audio: &openrtb2.Audio{MinDuration: 10}}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "GBR"}, UA: "tablet"},
				Ext:    json.RawMessage(`{"prebid":{"channel":{"name":"channel1","version":"ver1"},"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["channel","country","deviceType"]},"values":{"channel1|USA|tablet":10.01,"*|*|*":16.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"channel", "country", "deviceType"}},
			out:         []string{"channel1", "GBR", "tablet"},
		},
		{
			name: "CreateRule with Native mediaType, gptSlot and bundle",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Bundle:    "bundle1",
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, Ext: json.RawMessage(`{"data": {"adserver": {"name": "gam","adslot": "adslot123"}, "pbadslot": "pbadslot123"}}`)}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "GBR"}, UA: "tablet"},
				Ext:    json.RawMessage(`{"prebid":{"channel":{"name":"chName","version":"ver1"},"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["mediaType","gptSlot","bundle"]},"values":{"native|adslot123|bundle1":10.01,"native|pbadslot123|bundle1":11.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "gptSlot", "bundle"}},
			out:         []string{"native", "adslot123", "bundle1"},
		},
		{
			name: "CreateRule with Native mediaType, adUnitCode and bundle",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Bundle:    "bundle1",
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, Ext: json.RawMessage(`{"data": {"adserver": {"name": "gam","adslot": "adslot123"}, "pbadslot": "pbadslot123"}}`)}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "GBR"}, UA: "tablet"},
				Ext:    json.RawMessage(`{"prebid":{"channel":{"name":"chName","version":"ver1"},"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["mediaType","gptSlot","bundle"]},"values":{"native|adslot123|bundle1":10.01,"native|pbadslot123|bundle1":11.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "adUnitCode", "bundle"}},
			out:         []string{"native", "pbadslot123", "bundle1"},
		},
		{
			name: "CreateRule with Native mediaType, adUnitCode and siteDomain (App)",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Domain: "www.test.com",
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, Ext: json.RawMessage(`{"data": {"adserver": {"name": "gam","adslot": "adslot123"}, "pbadslot": "pbadslot123"}}`)}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "GBR"}, UA: "tablet"},
				Ext:    json.RawMessage(`{"prebid":{"channel":{"name":"chName","version":"ver1"},"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["mediaType","gptSlot","siteDomain"]},"values":{"native|adslot123|www.test.com":10.01,"native|pbadslot123|*":11.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "adUnitCode", "siteDomain"}},
			out:         []string{"native", "pbadslot123", "www.test.com"},
		},
		{
			name: "CreateRule with Native mediaType, siteDomain and adUnitCode (gpid)",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Domain: "www.test.com",
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, Ext: json.RawMessage(`{"gpid": "gpid_1"}`)}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "GBR"}, UA: "tablet"},
				Ext:    json.RawMessage(`{"prebid":{"channel":{"name":"chName","version":"ver1"},"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["mediaType","gptSlot","siteDomain"]},"values":{"native|gpid_1|www.test.com":10.01,"native|*|*":11.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "adUnitCode", "siteDomain"}},
			out:         []string{"native", "gpid_1", "www.test.com"},
		},
		{
			name: "CreateRule with Native mediaType, siteDomain and adUnitCode (tagId)",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Domain: "www.test.com",
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, TagID: "tag_123"}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "GBR"}, UA: "tablet"},
				Ext:    json.RawMessage(`{"prebid":{"channel":{"name":"chName","version":"ver1"},"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["mediaType","gptSlot","siteDomain"]},"values":{"native|tag_123|www.test.com":10.01,"native|tag_123|*":11.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "adUnitCode", "siteDomain"}},
			out:         []string{"native", "tag_123", "www.test.com"},
		},
		{
			name: "CreateRule with Native mediaType, siteDomain and adUnitCode (*)",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Domain: "www.test.com",
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "GBR"}, UA: "tablet"},
				Ext:    json.RawMessage(`{"prebid":{"channel":{"name":"chName","version":"ver1"},"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["mediaType","gptSlot","siteDomain"]},"values":{"native|*|www.test.com":10.01,"native|*|*":11.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "adUnitCode", "siteDomain"}},
			out:         []string{"native", "*", "www.test.com"},
		},
		{
			name: "CreateRule with Native mediaType, adUnitCode and pubDomain  (App)",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, Ext: json.RawMessage(`{"data": {"adserver": {"adslot": "adslot123"}, "pbadslot": "pbadslot123"}}`)}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "GBR"}, UA: "tablet"},
				Ext:    json.RawMessage(`{"prebid":{"channel":{"name":"chName","version":"ver1"},"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["mediaType","gptSlot","pubDomain"]},"values":{"native|pbadslot123|www.website.com":10.01,"native|*|*":11.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "gptSlot", "pubDomain"}},
			out:         []string{"native", "pbadslot123", "www.website.com"},
		},
		{
			name: "CreateRule with Native mediaType, adUnitCode and domain  (App)",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, Ext: json.RawMessage(`{"data": {"adserver": {"adslot": "adslot123"}, "pbadslot": "pbadslot123"}}`)}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "GBR"}, UA: "tablet"},
				Ext:    json.RawMessage(`{"prebid":{"channel":{"name":"chName","version":"ver1"},"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["mediaType","gptSlot","domain"]},"values":{"native|pbadslot123|www.website.com":10.01,"native|*|*":11.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"mediaType", "gptSlot", "domain"}},
			out:         []string{"native", "pbadslot123", "www.website.com"},
		},
		{
			name: "CreateRule with deviceType, adUnitCode and siteDomain (Site)",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.test.com",
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, Ext: json.RawMessage(`{"data": {"adserver": {"name": "gam","adslot": "adslot123"}, "pbadslot": "pbadslot123"}}`)}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "GBR"}},
				Ext:    json.RawMessage(`{"prebid":{"channel":{"name":"chName","version":"ver1"},"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["deviceType","gptSlot","siteDomain"]},"values":{"*|adslot123|www.test.com":10.01,"*|pbadslot123|*":11.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"deviceType", "adUnitCode", "siteDomain"}},
			out:         []string{"*", "pbadslot123", "www.test.com"},
		},
		{
			name: "CreateRule with channel, adUnitCode and pubDomain (Site)",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, Ext: json.RawMessage(`{"prebid": {"storedrequest": {"id": "123"}}}`)}},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "GBR"}, UA: "tablet"},
				Ext:    json.RawMessage(`{"prebid":{"floors":{"data":{"currency":"USD","skipRate":0,"schema":{"fields":["channel","adUnitCode","pubDomain"]},"values":{"*|123|www.website.com":10.01,"*|*|*":11.01},"default":1}}}}`),
			},
			floorSchema: openrtb_ext.PriceFloorSchema{Delimiter: "|", Fields: []string{"channel", "adUnitCode", "pubDomain"}},
			out:         []string{"*", "123", "www.website.com"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := createRuleKey(tc.floorSchema, tc.request, tc.request.Imp[0])
			if !reflect.DeepEqual(out, tc.out) {
				t.Errorf("error: \nreturn:\t%v\nwant:\t%v", out, tc.out)
			}
		})
	}
}

func TestShouldSkipFloors(t *testing.T) {

	tt := []struct {
		name                string
		ModelGroupsSkipRate int
		DataSkipRate        int
		RootSkipRate        int
		out                 bool
		randomGen           func(int) int
	}{
		{
			name:                "ModelGroupsSkipRate=10 with skip = true",
			ModelGroupsSkipRate: 10,
			DataSkipRate:        0,
			RootSkipRate:        0,
			randomGen:           func(i int) int { return 5 },
			out:                 true,
		},
		{
			name:                "ModelGroupsSkipRate=100 with skip = true",
			ModelGroupsSkipRate: 100,
			DataSkipRate:        0,
			RootSkipRate:        0,
			randomGen:           func(i int) int { return 5 },
			out:                 true,
		},
		{
			name:                "ModelGroupsSkipRate=0 with skip = false",
			ModelGroupsSkipRate: 0,
			DataSkipRate:        0,
			RootSkipRate:        0,
			randomGen:           func(i int) int { return 5 },
			out:                 false,
		},
		{
			name:                "DataSkipRate=50  with with skip = true",
			ModelGroupsSkipRate: 0,
			DataSkipRate:        50,
			RootSkipRate:        0,
			randomGen:           func(i int) int { return 40 },
			out:                 true,
		},
		{
			name:                "RootSkipRate=50  with with skip = true",
			ModelGroupsSkipRate: 0,
			DataSkipRate:        0,
			RootSkipRate:        60,
			randomGen:           func(i int) int { return 40 },
			out:                 true,
		},
		{
			name:                "RootSkipRate=50  with with skip = false",
			ModelGroupsSkipRate: 0,
			DataSkipRate:        0,
			RootSkipRate:        60,
			randomGen:           func(i int) int { return 70 },
			out:                 false,
		},
		{
			name:                "RootSkipRate=100  with with skip = true",
			ModelGroupsSkipRate: 0,
			DataSkipRate:        0,
			RootSkipRate:        100,
			randomGen:           func(i int) int { return 100 },
			out:                 true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := shouldSkipFloors(tc.ModelGroupsSkipRate, tc.DataSkipRate, tc.RootSkipRate, tc.randomGen)
			if !reflect.DeepEqual(out, tc.out) {
				t.Errorf("error: \nreturn:\t%v\nwant:\t%v", out, tc.out)
			}
		})
	}

}

func TestSelectFloorModelGroup(t *testing.T) {
	floorExt := &openrtb_ext.PriceFloorRules{Data: &openrtb_ext.PriceFloorData{
		SkipRate: 30,
		ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
			ModelWeight:  50,
			SkipRate:     10,
			ModelVersion: "Version 1",
			Schema:       openrtb_ext.PriceFloorSchema{Fields: []string{"mediaType", "size", "domain"}},
			Values: map[string]float64{
				"banner|300x250|www.website.com": 1.01,
				"banner|300x250|*":               2.01,
				"banner|300x600|www.website.com": 3.01,
				"banner|300x600|*":               4.01,
				"banner|728x90|www.website.com":  5.01,
				"banner|728x90|*":                6.01,
				"banner|*|www.website.com":       7.01,
				"banner|*|*":                     8.01,
				"*|300x250|www.website.com":      9.01,
				"*|300x250|*":                    10.01,
				"*|300x600|www.website.com":      11.01,
				"*|300x600|*":                    12.01,
				"*|728x90|www.website.com":       13.01,
				"*|728x90|*":                     14.01,
				"*|*|www.website.com":            15.01,
				"*|*|*":                          16.01,
			}, Default: 0.01},
			{
				ModelWeight:  25,
				SkipRate:     20,
				ModelVersion: "Version 2",
				Schema:       openrtb_ext.PriceFloorSchema{Fields: []string{"mediaType", "size", "domain"}},
				Values: map[string]float64{
					"banner|300x250|www.website.com": 1.01,
					"banner|300x250|*":               2.01,
					"banner|300x600|www.website.com": 3.01,
					"banner|300x600|*":               4.01,
					"banner|728x90|www.website.com":  5.01,
					"banner|728x90|*":                6.01,
					"banner|*|www.website.com":       7.01,
					"banner|*|*":                     8.01,
					"*|300x250|www.website.com":      9.01,
					"*|300x250|*":                    10.01,
					"*|300x600|www.website.com":      11.01,
					"*|300x600|*":                    12.01,
					"*|728x90|www.website.com":       13.01,
					"*|728x90|*":                     14.01,
					"*|*|www.website.com":            15.01,
					"*|*|*":                          16.01,
				}, Default: 0.01},
		}}}

	tt := []struct {
		name         string
		floorExt     *openrtb_ext.PriceFloorRules
		ModelVersion string
		fn           func(int) int
	}{
		{
			name:         "Version 2 Selection",
			floorExt:     floorExt,
			ModelVersion: "Version 2",
			fn:           func(i int) int { return 5 },
		},
		{
			name:         "Version 1 Selection",
			floorExt:     floorExt,
			ModelVersion: "Version 1",
			fn:           func(i int) int { return 55 },
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			selectFloorModelGroup(tc.floorExt.Data.ModelGroups, tc.fn)

			if !reflect.DeepEqual(tc.floorExt.Data.ModelGroups[0].ModelVersion, tc.ModelVersion) {
				t.Errorf("Floor Model Version mismatch error: \nreturn:\t%v\nwant:\t%v", tc.floorExt.Data.ModelGroups[0].ModelVersion, tc.ModelVersion)
			}

		})
	}
}
