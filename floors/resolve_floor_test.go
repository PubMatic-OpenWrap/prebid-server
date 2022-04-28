package floors

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/mxmCherry/openrtb/v15/openrtb2"
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
				"A",
				"*",
			},
		},
		{
			name: "Schema items, n = 2",
			in:   []string{"A", "B"},
			n:    2,
			del:  "|",
			out: []string{
				"A|B",
				"A|*",
				"*|B",
				"*|*",
			},
		},
		{
			name: "Schema items, n = 3",
			in:   []string{"A", "B", "C"},
			n:    3,
			del:  "|",
			out: []string{
				"A|B|C",
				"A|B|*",
				"A|*|C",
				"*|B|C",
				"A|*|*",
				"*|B|*",
				"*|*|C",
				"*|*|*",
			},
		},
		{
			name: "Schema items, n = 4",
			in:   []string{"A", "B", "C", "D"},
			n:    4,
			del:  "|",
			out: []string{
				"A|B|C|D",
				"A|B|C|*",
				"A|B|*|D",
				"A|*|C|D",
				"*|B|C|D",
				"A|B|*|*",
				"A|*|C|*",
				"A|*|*|D",
				"*|B|C|*",
				"*|B|*|D",
				"*|*|C|D",
				"A|*|*|*",
				"*|B|*|*",
				"*|*|C|*",
				"*|*|*|D",
				"*|*|*|*",
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := PrepareRuleCombinations(tc.in, tc.n, tc.del)
			if !reflect.DeepEqual(out, tc.out) {
				t.Errorf("error: \nreturn:\t%v\nwant:\t%v", out, tc.out)
			}
		})
	}
}

func TestUpdateImpsWithFloors1(t *testing.T) {
	floorExt := &openrtb_ext.PriceFloorRules{Data: openrtb_ext.PriceFloorData{ModelGroups: []openrtb_ext.PriceFloorModelGroup{{Schema: openrtb_ext.PriceFloorSchema{Fields: []string{"mediaType", "size", "domain"}, Delimiter: "|"},
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
		}, Default: 0.01}}}}

	floorExt2 := &openrtb_ext.PriceFloorRules{Data: openrtb_ext.PriceFloorData{ModelGroups: []openrtb_ext.PriceFloorModelGroup{{Schema: openrtb_ext.PriceFloorSchema{Fields: []string{"mediaType", "size", "domain"}, Delimiter: "|"},
		Values: map[string]float64{
			"banner|300x250|www.publisher.com": 1.01,
			"banner|300x250|*":                 2.01,
			"banner|300x600|www.publisher.com": 3.01,
			"banner|300x600|*":                 4.01,
			"banner|728x90|www.website.com":    5.01,
			"banner|728x90|*":                  6.01,
			"banner|*|www.website.com":         7.01,
			"banner|*|*":                       8.01,
			"video|*|*":                        9.01,
			"*|300x250|www.website.com":        10.01,
			"*|300x250|*":                      10.11,
			"*|300x600|www.website.com":        11.01,
			"*|300x600|*":                      12.01,
			"*|728x90|www.website.com":         13.01,
			"*|728x90|*":                       14.01,
			"*|*|www.website.com":              15.01,
			"*|*|*":                            16.01,
		}, Default: 0.01}}}}

	floorExt3 := &openrtb_ext.PriceFloorRules{Data: openrtb_ext.PriceFloorData{ModelGroups: []openrtb_ext.PriceFloorModelGroup{{Schema: openrtb_ext.PriceFloorSchema{Fields: []string{"mediaType", "size", "domain"}, Delimiter: "|"},
		Values: map[string]float64{
			"banner|300x250|www.publisher.com": 1.01,
			"banner|300x250|*":                 2.01,
			"banner|300x600|www.publisher.com": 3.01,
			"banner|300x600|*":                 4.01,
			"banner|728x90|www.website.com":    5.01,
			"banner|728x90|*":                  6.01,
			"banner|*|www.website.com":         7.01,
			"banner|*|*":                       8.01,
		}, Default: 0.01}}}}

	tt := []struct {
		name     string
		floorExt *openrtb_ext.PriceFloorRules
		request  *openrtb2.BidRequest
		floorVal float64
		floorCur string
	}{
		{
			name: "banner|300x250|www.website.com",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt,
			floorVal: 1.01,
			floorCur: "USD",
		},
		{
			name: "banner|300x600|www.website.com",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 600}}}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt,
			floorVal: 3.01,
			floorCur: "USD",
		},
		{
			name: "*|*|www.website.com",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 640, H: 480}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt,

			floorVal: 15.01,
			floorCur: "USD",
		},
		{
			name: "*|300x250|www.website.com",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt,

			floorVal: 9.01,
			floorCur: "USD",
		},
		{
			name: "banner|300x600|*",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 600}}}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt2,
			floorVal: 4.01,
			floorCur: "USD",
		},
		{
			name: "video|*|*",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 640, H: 480}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt2,

			floorVal: 9.01,
			floorCur: "USD",
		},
		{
			name: "*|300x250|www.website.com",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt2,

			floorVal: 10.01,
			floorCur: "USD",
		},
		{
			name: "Default Floor Value",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt3,

			floorVal: 0.01,
			floorCur: "USD",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_ = UpdateImpsWithFloors(tc.floorExt, tc.request)
			if !reflect.DeepEqual(tc.request.Imp[0].BidFloor, tc.floorVal) {
				t.Errorf("Floor Value error: \nreturn:\t%v\nwant:\t%v", tc.request.Imp[0].BidFloor, tc.floorVal)
			}
			if !reflect.DeepEqual(tc.request.Imp[0].BidFloorCur, tc.floorCur) {
				t.Errorf("Floor Currency error: \nreturn:\t%v\nwant:\t%v", tc.request.Imp[0].BidFloor, tc.floorCur)
			}
		})
	}
}

func TestCreateRuleKeys(t *testing.T) {
	tt := []struct {
		name        string
		floorSchema openrtb_ext.PriceFloorSchema
		request     *openrtb2.BidRequest
		imp         openrtb2.Imp
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
			imp:         openrtb2.Imp{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}},
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
			imp:         openrtb2.Imp{ID: "1234", Video: &openrtb2.Video{W: 640, H: 480}},
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
			imp:         openrtb2.Imp{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}},
			out:         []string{"video", "300x250", "www.test.com"},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := CreateRuleKey(tc.floorSchema, tc.request, tc.imp)
			if !reflect.DeepEqual(out, tc.out) {
				t.Errorf("error: \nreturn:\t%v\nwant:\t%v", out, tc.out)
			}
		})
	}
}
