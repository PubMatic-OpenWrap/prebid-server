package openrtb2

import (
	"context"
	"encoding/json"
	"fmt"
	"header-bidding/openrtb"

	"github.com/prebid/prebid-server/openrtb_ext"

	"github.com/julienschmidt/httprouter"
	"github.com/prebid/openrtb/v17/openrtb2"
	analyticsConf "github.com/prebid/prebid-server/analytics/config"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/exchange"
	metricsConfig "github.com/prebid/prebid-server/metrics/config"
	"github.com/prebid/prebid-server/stored_requests/backends/empty_fetcher"
)

func formORtbV25Request(formatFlag bool, videoFlag bool) *openrtb.BidRequest {
	request := new(openrtb.BidRequest)
	banner := new(openrtb.Banner)
	if formatFlag == true {
		formatObj1 := new(openrtb.Format) // openrtb.Format{728, 90, nil}
		formatObj1.W = new(int)
		*formatObj1.W = 728
		formatObj1.H = new(int)
		*formatObj1.H = 90

		formatObj2 := new(openrtb.Format) // openrtb.Format{728, 90, nil}
		formatObj2.W = new(int)
		*formatObj2.W = 300
		formatObj2.H = new(int)
		*formatObj2.H = 250

		formatArray := []*openrtb.Format{formatObj1, formatObj2}
		banner.Format = formatArray

		banner.W = new(int)
		*banner.W = 700
		banner.H = new(int)
		*banner.H = 900

	} else {
		banner.W = new(int)
		*banner.W = 728
		banner.H = new(int)
		*banner.H = 90
	}

	imp := new(openrtb.Imp)
	if videoFlag == true {
		video := formVideoObject()
		imp.Video = video
	}

	imp.Id = new(string)
	*imp.Id = "abcdefgh"
	imp.Banner = banner
	imp.TagId = new(string)
	*imp.TagId = "adunit"

	impWrapExt := new(openrtb.ExtImpWrapper)
	impWrapExt.Div = new(string)
	*impWrapExt.Div = "div"

	inImpExt := new(openrtb.ImpExtension)

	imp.Ext = inImpExt
	impArr := make([]*openrtb.Imp, 0)
	impArr = append(impArr, imp)
	request.Id = new(string)
	*request.Id = "123-456-789"
	request.Imp = impArr

	inImpExt.Prebid = new(openrtb_ext.ExtImpPrebid)
	inImpExt.Prebid.Bidder = map[string]json.RawMessage{
		"pubmatic": json.RawMessage(`""`),
	}

	len := 2
	request.Wseat = make([]string, len)
	for i := 0; i < len; i++ {
		request.Wseat[i] = fmt.Sprintf("Wseat_%d", i)
	}

	request.Cur = make([]string, len)
	for i := 0; i < len; i++ {
		request.Cur[i] = fmt.Sprintf("cur_%d", i)
	}

	request.Badv = make([]string, len)
	for i := 0; i < len; i++ {
		request.Badv[i] = fmt.Sprintf("badv_%d", i)
	}

	request.Bapp = make([]string, len)
	for i := 0; i < len; i++ {
		request.Bapp[i] = fmt.Sprintf("bapp_%d", i)
	}

	request.Bcat = make([]string, len)
	for i := 0; i < len; i++ {
		request.Bcat[i] = fmt.Sprintf("bcat_%d", i)
	}

	request.Wlang = make([]string, len)
	for i := 0; i < len; i++ {
		request.Wlang[i] = fmt.Sprintf("Wlang_%d", i)
	}

	request.Bseat = make([]string, len)
	for i := 0; i < len; i++ {
		request.Bseat[i] = fmt.Sprintf("Bseat_%d", i)
	}

	site := new(openrtb.Site)
	publisher := new(openrtb.Publisher)
	publisher.Id = new(string)
	*publisher.Id = "5890"
	site.Publisher = publisher
	site.Page = new(string)
	*site.Page = "www.test.com"

	site.Domain = new(string)
	*site.Domain = "test.com"

	request.Site = site

	request.Device = new(openrtb.Device)
	request.Device.IP = new(string)
	*request.Device.IP = "123.145.167.10"
	request.Device.Ua = new(string)
	*request.Device.Ua = "Mozilla/5.0(X11;Linuxx86_64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/52.0.2743.82Safari/537.36"

	request.User = new(openrtb.User)
	request.User.ID = new(string)
	*request.User.ID = "119208432"

	request.User.BuyerUID = new(string)
	*request.User.BuyerUID = "1rwe432"

	request.User.Yob = new(int)
	*request.User.Yob = 1980

	request.User.Gender = new(string)
	*request.User.Gender = "F"

	request.User.Geo = new(openrtb.Geo)
	request.User.Geo.Country = new(string)
	*request.User.Geo.Country = "US"

	request.User.Geo.Region = new(string)
	*request.User.Geo.Region = "CA"

	request.User.Geo.Metro = new(string)
	*request.User.Geo.Metro = "90001"

	request.User.Geo.City = new(string)
	*request.User.Geo.City = "Alamo"

	request.Source = new(openrtb.Source)
	request.Source.Ext = map[string]interface{}{
		"omidpn": "MyIntegrationPartner",
		"omidpv": "7.1",
	}

	wExt := new(openrtb.ExtRequest)
	dmExt := new(openrtb.ExtRequestWrapper)
	dmExt.ProfileId = new(int)
	*dmExt.ProfileId = 123
	dmExt.VersionId = new(int)
	*dmExt.VersionId = 1
	dmExt.LoggerImpressionID = new(string)
	*dmExt.LoggerImpressionID = "test_display_wiid"
	wExt.Wrapper = dmExt

	request.Ext = wExt

	request.Test = new(int)
	*request.Test = 0
	return request

}

func formVideoObject() *openrtb.Video {
	video := new(openrtb.Video)
	video.Mimes = []string{"video/mp4", "video/mpeg"}
	video.W = new(int)
	*video.W = 640
	video.H = new(int)
	*video.H = 480

	video.Ext = map[string]interface{}{
		"adpod": map[string]int{
			"minads":        1,
			"adminduration": 5,
			"excladv":       50,
			"maxads":        3,
			"excliabcat":    100,
			"admaxduration": 100,
		},
		"offset": 20,
	}
	video.MaxDuration = new(int)
	video.MinDuration = new(int)
	*video.MaxDuration = 50
	*video.MinDuration = 5

	return video
}

type mockExchangeCTV struct {
}

func (m *mockExchangeCTV) HoldAuction(ctx context.Context, auctionRequest exchange.AuctionRequest, debugLog *exchange.DebugLog) (*openrtb2.BidResponse, error) {

	ext := json.RawMessage(`{"video":{"duration":30}, "prebid":{"video":{"duration":30}}}`)
	return &openrtb2.BidResponse{
		SeatBid: []openrtb2.SeatBid{
			{
				Seat: "pubmatic",
				Bid: []openrtb2.Bid{
					{ID: "VIDEO12-89A1-41F1-8708-978FD3C0912A", ImpID: "abcdefgh_1", Price: 0, AdM: "<VAST><![CDATA[XYZ]]></VAST>", Dur: 30, Ext: ext},
					{ID: "VIDEO12-89A1-41F1-8708-978FD3C0912A", ImpID: "abcdefgh_2", Price: 10, AdM: "<VAST><![CDATA[XYZ]]></VAST>", Dur: 30, Ext: ext},
				},
			},
		},
	}, nil
}

func GetCTVHandler() httprouter.Handle {
	mockExchange := mockExchangeCTV{}
	endpoint, _ := NewCTVEndpoint(
		&mockExchange,
		mockBidderParamValidator{},
		&mockVideoStoredReqFetcher{},
		&mockVideoStoredReqFetcher{},
		empty_fetcher.EmptyFetcher{},
		&config.Configuration{MaxRequestSize: maxSize, GenerateBidID: true},
		&metricsConfig.NilMetricsEngine{},
		analyticsConf.NewPBSAnalytics(&config.Analytics{}),
		map[string]string{},
		[]byte{},
		openrtb_ext.BuildBidderMap(),
	)
	return endpoint
}
