package floors

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func TestIsRequestEnabledWithFloor(t *testing.T) {
	FalseFlag := false
	TrueFlag := true

	tt := []struct {
		name string
		in   *openrtb_ext.ExtRequest
		out  bool
	}{
		{
			name: "Request With Nil Floors",
			in:   &openrtb_ext.ExtRequest{},
			out:  false,
		},
		{
			name: "Request With Floors Disabled",
			in:   &openrtb_ext.ExtRequest{Prebid: openrtb_ext.ExtRequestPrebid{Floors: &openrtb_ext.PriceFloorRules{Enabled: &FalseFlag}}},
			out:  false,
		},
		{
			name: "Request With Floors Enabled",
			in:   &openrtb_ext.ExtRequest{Prebid: openrtb_ext.ExtRequestPrebid{Floors: &openrtb_ext.PriceFloorRules{Enabled: &TrueFlag}}},
			out:  true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out := IsRequestEnabledWithFloor(tc.in)
			if !reflect.DeepEqual(out, tc.out) {
				t.Errorf("error: \nreturn:\t%v\nwant:\t%v", out, tc.out)
			}
		})
	}
}
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

	floorExt := &openrtb_ext.PriceFloorRules{Data: &openrtb_ext.PriceFloorData{ModelGroups: []openrtb_ext.PriceFloorModelGroup{{Schema: openrtb_ext.PriceFloorSchema{Fields: []string{"mediaType", "country", "deviceType"}},
		Values: map[string]float64{
			"audio|USA|phone": 1.01,
		}, Default: 0.01}}}}

	floorExt2 := &openrtb_ext.PriceFloorRules{Data: &openrtb_ext.PriceFloorData{ModelGroups: []openrtb_ext.PriceFloorModelGroup{{Schema: openrtb_ext.PriceFloorSchema{Fields: []string{"channel", "country", "deviceType"}},
		Values: map[string]float64{
			"chName|USA|tablet": 1.01,
			"*|USA|tablet":      2.01,
		}, Default: 0.01}}}}

	floorExt3 := &openrtb_ext.PriceFloorRules{FloorMin: 1.00, Data: &openrtb_ext.PriceFloorData{ModelGroups: []openrtb_ext.PriceFloorModelGroup{{Schema: openrtb_ext.PriceFloorSchema{Fields: []string{"mediaType", "gptSlot", "bundle"}},
		Values: map[string]float64{
			"native|adslot123|bundle1":   0.01,
			"native|pbadslot123|bundle1": 0.01,
		}, Default: 0.01}}}}

	floorExt4 := &openrtb_ext.PriceFloorRules{FloorMin: 1.00, Data: &openrtb_ext.PriceFloorData{ModelGroups: []openrtb_ext.PriceFloorModelGroup{{Schema: openrtb_ext.PriceFloorSchema{Fields: []string{"mediaType", "pbAdSlot", "bundle"}},
		Values: map[string]float64{
			"native|pbadslot123|bundle1": 0.01,
		}, Default: 0.01}}}}
	tt := []struct {
		name     string
		floorExt *openrtb_ext.PriceFloorRules
		request  *openrtb2.BidRequest
		floorVal float64
		floorCur string
	}{
		{
			name: "audio|USA|phone",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "USA"}, UA: "Phone"},
				Imp:    []openrtb2.Imp{{ID: "1234", Audio: &openrtb2.Audio{MaxDuration: 10}}},
				Ext:    json.RawMessage(`{"prebid": {"floors": {"data": {"currency": "USD","skipRate": 0, "schema": {"fields": ["channel","size","domain"]},"values": {"chName|USA|tablet": 1.01, "*|*|*": 16.01},"default": 1},"channel": {"name": "chName","version": "ver1"}}}}`),
			},
			floorExt: floorExt,
			floorVal: 1.01,
			floorCur: "USD",
		},
		{
			name: "chName|USA|tablet",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "USA"}, UA: "tablet"},
				Imp:    []openrtb2.Imp{{ID: "1234", Audio: &openrtb2.Audio{MaxDuration: 10}}},
				Ext:    json.RawMessage(`{"prebid": {"channel": {"name": "chName","version": "ver1"}}}`)},
			floorExt: floorExt2,
			floorVal: 1.01,
			floorCur: "USD",
		},
		{
			name: "*|USA|tablet",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "USA"}, UA: "tablet"},
				Imp:    []openrtb2.Imp{{ID: "1234", Audio: &openrtb2.Audio{MaxDuration: 10}}},
				Ext:    json.RawMessage(`{"prebid": }`)},
			floorExt: floorExt2,
			floorVal: 2.01,
			floorCur: "USD",
		},
		{
			name: "native|gptSlot|bundle1",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Bundle:    "bundle1",
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "USA"}, UA: "tablet"},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, Ext: json.RawMessage(`{"data": {"adserver": {"name": "gam","adslot": "adslot123"}, "pbadslot": "pbadslot123"}}`)}},
				Ext:    json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt3,
			floorVal: 1.00,
			floorCur: "USD",
		},
		{
			name: "native|pbAdSlot|bundle1",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Bundle:    "bundle1",
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "USA"}, UA: "tablet"},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, Ext: json.RawMessage(`{"data": {"adserver": {"name": "gam","adslot": "adslot123"}, "pbadslot": "pbadslot123"}}`)}},
				Ext:    json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt4,
			floorVal: 1.00,
			floorCur: "USD",
		},
		{
			name: "native|gptSlot|bundle1",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Bundle:    "bundle1",
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Device: &openrtb2.Device{Geo: &openrtb2.Geo{Country: "USA"}, UA: "tablet"},
				Imp:    []openrtb2.Imp{{ID: "1234", Native: &openrtb2.Native{}, Ext: json.RawMessage(`{"data": {"adserver": {"name": "ow","adslot": "adslot123"}, "pbadslot": "pbadslot123"}}`)}},
				Ext:    json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt3,
			floorVal: 1.00,
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
func TestUpdateImpsWithFloors(t *testing.T) {

	floorExt := &openrtb_ext.PriceFloorRules{Data: &openrtb_ext.PriceFloorData{ModelGroups: []openrtb_ext.PriceFloorModelGroup{{Schema: openrtb_ext.PriceFloorSchema{Fields: []string{"mediaType", "size", "domain"}},
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

	floorExt2 := &openrtb_ext.PriceFloorRules{Data: &openrtb_ext.PriceFloorData{ModelGroups: []openrtb_ext.PriceFloorModelGroup{{Schema: openrtb_ext.PriceFloorSchema{Fields: []string{"mediaType", "size", "siteDomain"}, Delimiter: "|"},
		Values: map[string]float64{
			"banner|300x250|www.publisher.com":   1.01,
			"banner|300x250|*":                   2.01,
			"banner|300x600|www.publisher.com":   3.01,
			"banner|300x600|*":                   4.01,
			"banner|728x90|www.website.com":      5.01,
			"banner|728x90|www.website.com|test": 5.01,
			"banner|728x90|*":                    6.01,
			"banner|*|www.website.com":           7.01,
			"banner|*|*":                         8.01,
			"video|*|*":                          9.01,
			"*|300x250|www.website.com":          10.01,
			"*|300x250|*":                        10.11,
			"*|300x600|www.website.com":          11.01,
			"*|300x600|*":                        12.01,
			"*|728x90|www.website.com":           13.01,
			"*|728x90|*":                         14.01,
			"*|*|www.website.com":                15.01,
			"*|*|*":                              16.01,
		}, Default: 0.01}}}}

	floorExt3 := &openrtb_ext.PriceFloorRules{Data: &openrtb_ext.PriceFloorData{ModelGroups: []openrtb_ext.PriceFloorModelGroup{{Schema: openrtb_ext.PriceFloorSchema{Fields: []string{"mediaType", "size", "pubDomain"}, Delimiter: "|"},
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

	width := int64(300)
	height := int64(600)
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
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
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
				Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{W: &width, H: &height}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt,
			floorVal: 3.01,
			floorCur: "USD",
		},
		{
			name: "*|*|www.website.com",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
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
				App: &openrtb2.App{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt,
			floorVal: 9.01,
			floorCur: "USD",
		},
		{
			name: "siteDomain, banner|300x600|*",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 600}}}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "siteDomain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt2,
			floorVal: 4.01,
			floorCur: "USD",
		},
		{
			name: "siteDomain, video|*|*",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 640, H: 480}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "siteDomain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt2,

			floorVal: 9.01,
			floorCur: "USD",
		},
		{
			name: "pubDomain, *|300x250|www.website.com",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "pubDomain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt2,

			floorVal: 9.01,
			floorCur: "USD",
		},
		{
			name: "pubDomain, Default Floor Value",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "pubDomain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt: floorExt3,

			floorVal: 0.01,
			floorCur: "USD",
		},

		{
			name: "pubDomain, Default Floor Value",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
				},
				Imp: []openrtb2.Imp{{ID: "1234", Video: &openrtb2.Video{W: 300, H: 250}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "pubDomain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
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

func TestUpdateImpsWithModelGroups(t *testing.T) {
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
				ModelWeight:  50,
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
		request      *openrtb2.BidRequest
		floorVal     float64
		floorCur     string
		ModelVersion string
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
			floorExt:     floorExt,
			floorVal:     1.01,
			floorCur:     "USD",
			ModelVersion: "Version 1",
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
			floorExt:     floorExt,
			floorVal:     3.01,
			floorCur:     "USD",
			ModelVersion: "Version 2",
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

			floorVal:     15.01,
			floorCur:     "USD",
			ModelVersion: "Version 2",
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

			floorVal:     9.01,
			floorCur:     "USD",
			ModelVersion: "Version 2",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_ = UpdateImpsWithFloors(tc.floorExt, tc.request)
			if tc.floorExt.Skipped != nil && *tc.floorExt.Skipped != true {
				if !reflect.DeepEqual(tc.request.Imp[0].BidFloor, tc.floorVal) {
					t.Errorf("Floor Value error: \nreturn:\t%v\nwant:\t%v", tc.request.Imp[0].BidFloor, tc.floorVal)
				}
				if !reflect.DeepEqual(tc.request.Imp[0].BidFloorCur, tc.floorCur) {
					t.Errorf("Floor Currency error: \nreturn:\t%v\nwant:\t%v", tc.request.Imp[0].BidFloor, tc.floorCur)
				}

				if !reflect.DeepEqual(tc.floorExt.Data.ModelGroups[0].ModelVersion, tc.ModelVersion) {
					t.Errorf("Floor Model Version mismatch error: \nreturn:\t%v\nwant:\t%v", tc.floorExt.Data.ModelGroups[0].ModelVersion, tc.ModelVersion)
				}
			}
		})
	}
}

func TestUpdateImpsWithInvalidModelGroups(t *testing.T) {
	floorExt := &openrtb_ext.PriceFloorRules{Data: &openrtb_ext.PriceFloorData{
		ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
			ModelWeight:  50,
			SkipRate:     110,
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
				ModelWeight:  50,
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

	floorExt2 := &openrtb_ext.PriceFloorRules{Data: &openrtb_ext.PriceFloorData{
		ModelGroups: []openrtb_ext.PriceFloorModelGroup{{
			ModelWeight:  -1,
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
				ModelWeight:  50,
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
		request      *openrtb2.BidRequest
		floorVal     float64
		floorCur     string
		ModelVersion string
		Err          string
	}{
		{
			name: "Invalid Skip Rate in model Group 1, with banner|300x250|www.website.com",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt:     floorExt,
			floorVal:     1.01,
			floorCur:     "USD",
			ModelVersion: "Version 2",
			Err:          "Invalid Floor Model = 'Version 1' due to SkipRate = '110'",
		},
		{
			name: "Invalid model weight Model Group 1, with banner|300x250|www.website.com",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					Domain: "www.website.com",
				},
				Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
				Ext: json.RawMessage(`{"prebid": { "floors": {"data": {"currency": "USD","skipRate": 0,"schema": {"fields": [ "mediaType", "size", "domain" ] },"values": {  "banner|300x250|www.website.com": 1.01, "banner|300x250|*": 2.01, "banner|300x600|www.website.com": 3.01,  "banner|300x600|*": 4.01, "banner|728x90|www.website.com": 5.01, "banner|728x90|*": 6.01, "banner|*|www.website.com": 7.01, "banner|*|*": 8.01, "*|300x250|www.website.com": 9.01, "*|300x250|*": 10.01, "*|300x600|www.website.com": 11.01,  "*|300x600|*": 12.01,  "*|728x90|www.website.com": 13.01, "*|728x90|*": 14.01,  "*|*|www.website.com": 15.01, "*|*|*": 16.01  }, "default": 1}}}}`),
			},
			floorExt:     floorExt2,
			floorVal:     1.01,
			floorCur:     "USD",
			ModelVersion: "Version 2",
			Err:          "Invalid Floor Model = 'Version 1' due to ModelWeight = '-1'",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ErrList := UpdateImpsWithFloors(tc.floorExt, tc.request)
			if tc.floorExt.Skipped != nil && *tc.floorExt.Skipped != true {
				if !reflect.DeepEqual(tc.request.Imp[0].BidFloor, tc.floorVal) {
					t.Errorf("Floor Value error: \nreturn:\t%v\nwant:\t%v", tc.request.Imp[0].BidFloor, tc.floorVal)
				}
				if !reflect.DeepEqual(tc.request.Imp[0].BidFloorCur, tc.floorCur) {
					t.Errorf("Floor Currency error: \nreturn:\t%v\nwant:\t%v", tc.request.Imp[0].BidFloor, tc.floorCur)
				}

				if !reflect.DeepEqual(tc.floorExt.Data.ModelGroups[0].ModelVersion, tc.ModelVersion) {
					t.Errorf("Floor Model Version mismatch error: \nreturn:\t%v\nwant:\t%v", tc.floorExt.Data.ModelGroups[0].ModelVersion, tc.ModelVersion)
				}
			}

			if !reflect.DeepEqual(ErrList[0].Error(), tc.Err) {
				t.Errorf("Incorrect Error: \nreturn:\t%v\nwant:\t%v", ErrList[0].Error(), tc.Err)
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
