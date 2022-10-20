package floors

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/mxmCherry/openrtb/v16/openrtb2"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/currency"
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
			out:  true,
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
			out := tc.in.Prebid.Floors.GetEnabled()
			if !reflect.DeepEqual(out, tc.out) {
				t.Errorf("error: \nreturn:\t%v\nwant:\t%v", out, tc.out)
			}
		})
	}
}

func getCurrencyRates(rates map[string]map[string]float64) currency.Conversions {
	return currency.NewRates(rates)
}

func TestEnrichWithPriceFloors(t *testing.T) {
	rates := map[string]map[string]float64{
		"USD": {
			"INR": 70,
			"EUR": 0.9,
			"JPY": 5.09,
		},
	}

	width := int64(300)
	height := int64(600)

	tt := []struct {
		name              string
		bidRequestWrapper *openrtb_ext.RequestWrapper
		account           config.Account
		conversions       currency.Conversions
		Skipped           bool
		err               string
		expFloorVal       float64
		expFloorCur       string
		expPriceFlrLoc    int
	}{
		{
			name: "Floors disbaled in account config",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","currency":"USD","values":{"banner|300x600|www.website5.com":5,"*|*|*":7},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled: false,
				},
			},
		},
		{
			name: "Floors disbaled in req.ext.prebid.floors.Enabled=false config",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","currency":"USD","values":{"banner|300x600|www.website.com":5,"*|*|*":7},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":false,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled:        true,
					UseDynamicData: false,
					Fetch: config.AccountFloorFetch{
						Enabled: false,
					},
				},
			},
		},
		{
			name: "Floors enabled in req.ext.prebid.floors.Enabled and account config",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"data":{"currency":"USD","modelgroups":[{"modelversion":"model 2 from req","currency":"USD","values":{"banner|300x250|www.website.com":5,"*|*|*":7},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}},{"modelversion":"model 1 from req","currency":"USD","values":{"banner|300x250|www.website.com":5,"*|*|*":7},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled:        true,
					UseDynamicData: false,
					Fetch: config.AccountFloorFetch{
						Enabled: false,
					},
				},
			},
			expFloorVal:    5,
			expFloorCur:    "USD",
			expPriceFlrLoc: openrtb_ext.Request,
		},
		{
			name: "Skiprate = 100, Floors enaabled in  req.ext.prebid.floors.Enabled and account config: Floors singalling skipped ",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","currency":"USD","values":{"banner|300x250|www.website.com":5,"*|*|*":7},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"skiprate": 100,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled:        true,
					UseDynamicData: false,
					Fetch: config.AccountFloorFetch{
						Enabled: false,
					},
				},
			},
			Skipped: true,
		},
		{
			name: "Invalid Skiprate = 110: Floors singalling skipped ",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","currency":"USD","values":{"banner|300x250|www.website.com":5,"*|*|*":7},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"skiprate": 110,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled:        true,
					UseDynamicData: false,
					Fetch: config.AccountFloorFetch{
						Enabled: false,
					},
				},
			},
			err: "Invalid SkipRate at root level = '110'",
		},
		{
			name: "Rule selection with Site object, banner|300x600|www.website.com",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{W: &width, H: &height}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","currency":"USD","values":{"banner|300x600|www.website.com":5,"*|*|*":7},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled:        true,
					UseDynamicData: false,
					Fetch: config.AccountFloorFetch{
						Enabled: false,
					},
				},
			},
			expFloorVal:    5,
			expFloorCur:    "USD",
			expPriceFlrLoc: openrtb_ext.Request,
		},
		{
			name: "Rule selection with App object, *|*|www.test.com",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					App: &openrtb2.App{
						Domain: "www.test.com",
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{W: &width, H: &height}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","currency":"USD","values":{"banner|300x600|www.website.com":5,"*|*|www.test.com":15,"*|*|*":7},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled:        true,
					UseDynamicData: false,
					Fetch: config.AccountFloorFetch{
						Enabled: false,
					},
				},
			},
			expFloorVal:    15,
			expFloorCur:    "USD",
			expPriceFlrLoc: openrtb_ext.Request,
		},
		{
			name: "Floors Signalling not done as req.ext.prebid.floors not provided",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					App: &openrtb2.App{
						Domain: "www.test.com",
					},
					Imp: []openrtb2.Imp{{ID: "1234", BidFloor: 10, BidFloorCur: "EUR", Banner: &openrtb2.Banner{W: &width, H: &height}}},
					Ext: json.RawMessage(`{"prebid":{}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled:        true,
					UseDynamicData: false,
					Fetch: config.AccountFloorFetch{
						Enabled: false,
					},
				},
			},
			expFloorVal:    10,
			expFloorCur:    "EUR",
			expPriceFlrLoc: openrtb_ext.NoData,
		},
		{
			name: "BidFloor(USD) Less than MinBidFloor(INR) with different currency",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"floormin":80,"floormincur":"INR","data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","currency":"USD","values":{"banner|300x250|www.website.com":1,"*|*|www.test.com":15,"*|*|*":7},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled: true,
				},
			},
			expFloorVal:    1.1429,
			expFloorCur:    "USD",
			expPriceFlrLoc: openrtb_ext.Request,
		},
		{
			name: "BidFloor(INR) Less than MinBidFloor(USD) with different currency",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"floormin":1,"floormincur":"USD","data":{"currency":"INR","modelgroups":[{"modelversion":"model 1 from req","values":{"banner|300x250|www.website.com":60,"*|*|www.test.com":65,"*|*|*":67},"Default":50,"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled: true,
				},
			},
			expFloorVal:    70,
			expFloorCur:    "INR",
			expPriceFlrLoc: openrtb_ext.Request,
		},
		{
			name: "BidFloor is greater than MinBidFloor with same currency",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"floormin":1,"floormincur":"USD","data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","values":{"banner|300x250|www.website.com":2,"*|*|www.test.com":1.5,"*|*|*":1.7},"Default":5,"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled: true,
				},
			},
			expFloorVal:    2,
			expFloorCur:    "USD",
			expPriceFlrLoc: openrtb_ext.Request,
		},
		{
			name: "BidFloor Less than MinBidFloor with same currency",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"floormin":3,"floormincur":"USD","data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","values":{"banner|300x250|www.website.com":2,"*|*|www.test.com":1.5,"*|*|*":1.7},"Default":5,"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled: true,
				},
			},
			expFloorVal:    3,
			expFloorCur:    "USD",
			expPriceFlrLoc: openrtb_ext.Request,
		},
		{
			name: "No rule matched, Default value  greater than MinBidFloor with same currency",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"floormin":3,"floormincur":"USD","data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","values":{"banner|300x250|www.website.com1":2,"*|*|www.test2.com":1.5},"Default":15,"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled: true,
				},
			},
			expFloorVal:    15,
			expFloorCur:    "USD",
			expPriceFlrLoc: openrtb_ext.Request,
		},
		{
			name: "No rule matched, Default value  less than MinBidFloor with same currency",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"floormin":5,"floormincur":"USD","data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","values":{"banner|300x250|www.website.com1":2,"*|*|www.test2.com":1.5},"Default":2.5,"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled: true,
				},
			},
			expFloorVal:    5,
			expFloorCur:    "USD",
			expPriceFlrLoc: openrtb_ext.Request,
		},
		{
			name: "imp.bidfloor provided, No Rule matching and MinBidFloor, default values not provided in floor JSON",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", BidFloor: 1.5, BidFloorCur: "INR", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{ "data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","values":{"banner|300x250|www.website.com1":2,"*|*|www.test2.com":1.5},"schema":{"fields":["mediaType","size","domain"],"delimiter":"|"}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled: true,
				},
			},
			expFloorVal:    1.5,
			expFloorCur:    "INR",
			expPriceFlrLoc: openrtb_ext.Request,
		},
		{
			name: "imp.bidfloor provided, No Rule matching, MinBidFloor provided and , default values not provided in floor JSON",
			bidRequestWrapper: &openrtb_ext.RequestWrapper{
				BidRequest: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{Domain: "www.website.com"},
					},
					Imp: []openrtb2.Imp{{ID: "1234", BidFloor: 100, BidFloorCur: "INR", Banner: &openrtb2.Banner{Format: []openrtb2.Format{{W: 300, H: 250}}}}},
					Ext: json.RawMessage(`{"prebid":{"floors":{"floormin":2,"floormincur":"USD","data":{"currency":"USD","modelgroups":[{"modelversion":"model 1 from req","values":{"banner|300x250|www.website.com1":2,"*|*|www.test2.com":1.5},"schema":{"fields":["mediaType","size","domain"]}}]},"enabled":true,"enforcement":{"enforcepbs":true,"floordeals":true,"enforcerate":100}}}}`),
				},
			},
			account: config.Account{
				PriceFloors: config.AccountPriceFloors{
					Enabled: true,
				},
			},
			expFloorVal:    2,
			expFloorCur:    "USD",
			expPriceFlrLoc: openrtb_ext.Request,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ErrList := EnrichWithPriceFloors(tc.bidRequestWrapper, tc.account, getCurrencyRates(rates))
			if !reflect.DeepEqual(tc.bidRequestWrapper.Imp[0].BidFloor, tc.expFloorVal) {
				t.Errorf("Floor Value error: \nreturn:\t%v\nwant:\t%v", tc.bidRequestWrapper.Imp[0].BidFloor, tc.expFloorVal)
			}
			if !reflect.DeepEqual(tc.bidRequestWrapper.Imp[0].BidFloorCur, tc.expFloorCur) {
				t.Errorf("Floor Currency error: \nreturn:\t%v\nwant:\t%v", tc.bidRequestWrapper.Imp[0].BidFloor, tc.expFloorCur)
			}

			if len(ErrList) > 0 && !reflect.DeepEqual(ErrList[0].Error(), tc.err) {
				t.Errorf("Incorrect Error: \nreturn:\t%v\nwant:\t%v", ErrList[0].Error(), tc.err)
			}
			requestExt, err := tc.bidRequestWrapper.GetRequestExt()
			if tc.Skipped {
				if err == nil {
					prebidExt := requestExt.GetPrebid()
					if !reflect.DeepEqual(*prebidExt.Floors.Skipped, tc.Skipped) {
						t.Errorf("Floor Skipped error: \nreturn:\t%v\nwant:\t%v", *prebidExt.Floors.Skipped, tc.Skipped)
					}
				}
			} else {
				if err == nil {
					prebidExt := requestExt.GetPrebid()
					if prebidExt != nil && prebidExt.Floors != nil && !reflect.DeepEqual(prebidExt.Floors.PriceFloorLocation, tc.expPriceFlrLoc) {
						t.Errorf("Floor Skipped error: \nreturn:\t%v\nwant:\t%v", prebidExt.Floors.PriceFloorLocation, tc.expPriceFlrLoc)
					}
				}
			}

		})
	}
}

func TestResolveFloorMin(t *testing.T) {
	rates := map[string]map[string]float64{
		"USD": {
			"INR": 70,
			"EUR": 0.9,
			"JPY": 5.09,
		},
	}

	tt := []struct {
		name        string
		reqFloors   openrtb_ext.PriceFloorRules
		fetchFloors openrtb_ext.PriceFloorRules
		conversions currency.Conversions
		expPrice    Price
	}{
		{
			name: "FloorsMin presnet in request Floors only",
			reqFloors: openrtb_ext.PriceFloorRules{
				FloorMin:    10,
				FloorMinCur: "JPY",
			},
			fetchFloors: openrtb_ext.PriceFloorRules{},
			expPrice:    Price{FloorMin: 10, FloorMinCur: "JPY"},
		},
		{
			name: "FloorsMin presnet in request Floors and data currency presnet",
			reqFloors: openrtb_ext.PriceFloorRules{
				FloorMin:    10,
				FloorMinCur: "JPY",
				Data: &openrtb_ext.PriceFloorData{
					Currency: "JPY",
				},
			},
			fetchFloors: openrtb_ext.PriceFloorRules{},
			expPrice:    Price{FloorMin: 10, FloorMinCur: "JPY"},
		},
		{
			name: "FloorsMin presnet in request Floors and fetched floors",
			reqFloors: openrtb_ext.PriceFloorRules{
				FloorMin:    10,
				FloorMinCur: "USD",
			},
			fetchFloors: openrtb_ext.PriceFloorRules{
				FloorMin:    15,
				FloorMinCur: "USD",
			},
			expPrice: Price{FloorMin: 10, FloorMinCur: "USD"},
		},
		{
			name:      "FloorsMin presnet fetched floors only",
			reqFloors: openrtb_ext.PriceFloorRules{},
			fetchFloors: openrtb_ext.PriceFloorRules{
				FloorMin:    15,
				FloorMinCur: "EUR",
			},
			expPrice: Price{FloorMin: 15, FloorMinCur: "EUR"},
		},
		{
			name: "FloorMinCur presnet in reqFloors And FloorsMin, FloorMinCur presnet fetched floors (Same Currency)",
			reqFloors: openrtb_ext.PriceFloorRules{
				FloorMinCur: "EUR",
			},
			fetchFloors: openrtb_ext.PriceFloorRules{
				FloorMin:    15,
				FloorMinCur: "EUR",
			},
			expPrice: Price{FloorMin: 15, FloorMinCur: "EUR"},
		},
		{
			name: "FloorMinCur presnet in reqFloors And FloorsMin, FloorMinCur presnet fetched floors (Different Currency)",
			reqFloors: openrtb_ext.PriceFloorRules{
				FloorMinCur: "USD",
			},
			fetchFloors: openrtb_ext.PriceFloorRules{
				FloorMin:    15,
				FloorMinCur: "EUR",
			},
			expPrice: Price{FloorMin: 16.6667, FloorMinCur: "USD"},
		},
		{
			name: "FloorMin presnet in reqFloors And FloorMinCur presnet fetched floors",
			reqFloors: openrtb_ext.PriceFloorRules{
				FloorMin: 11,
			},
			fetchFloors: openrtb_ext.PriceFloorRules{
				FloorMinCur: "EUR",
			},
			expPrice: Price{FloorMin: 11, FloorMinCur: "EUR"},
		},
		{
			name: "FloorMinCur presnet in reqFloors And FloorMin presnet fetched floors",
			reqFloors: openrtb_ext.PriceFloorRules{
				FloorMinCur: "INR",
			},
			fetchFloors: openrtb_ext.PriceFloorRules{
				FloorMin: 12,
			},
			expPrice: Price{FloorMin: 12, FloorMinCur: "INR"},
		},
		{
			name:        "Empty reqFloors And Empty fetched floors",
			reqFloors:   openrtb_ext.PriceFloorRules{},
			fetchFloors: openrtb_ext.PriceFloorRules{},
			expPrice:    Price{FloorMin: 0.0, FloorMinCur: ""},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			price := resolveFloorMin(&tc.reqFloors, tc.fetchFloors, getCurrencyRates(rates))
			if !reflect.DeepEqual(price.FloorMin, tc.expPrice.FloorMin) {
				t.Errorf("Floor Value error: \nreturn:\t%v\nwant:\t%v", price.FloorMin, tc.expPrice.FloorMin)
			}
			if !reflect.DeepEqual(price.FloorMinCur, tc.expPrice.FloorMinCur) {
				t.Errorf("Floor Currency error: \nreturn:\t%v\nwant:\t%v", price.FloorMinCur, tc.expPrice.FloorMinCur)
			}

		})
	}
}
