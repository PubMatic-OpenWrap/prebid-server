package vastbidder

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/config"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

// TestSetDefaultHeaders verifies SetDefaultHeaders
func TestSetDefaultHeaders(t *testing.T) {
	type args struct {
		req *openrtb2.BidRequest
	}
	type want struct {
		headers http.Header
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "check all default headers",
			args: args{req: &openrtb2.BidRequest{
				Device: &openrtb2.Device{
					IP:       "1.1.1.1",
					UA:       "user-agent",
					Language: "en",
				},
				Site: &openrtb2.Site{
					Page: "http://test.com/",
				},
			}},
			want: want{
				headers: http.Header{
					"X-Device-Ip":              []string{"1.1.1.1"},
					"X-Forwarded-For":          []string{"1.1.1.1"},
					"X-Device-User-Agent":      []string{"user-agent"},
					"User-Agent":               []string{"user-agent"},
					"X-Device-Referer":         []string{"http://test.com/"},
					"X-Device-Accept-Language": []string{"en"},
				},
			},
		},
		{
			name: "nil bid request",
			args: args{req: nil},
			want: want{
				headers: http.Header{},
			},
		},
		{
			name: "no headers set",
			args: args{req: &openrtb2.BidRequest{}},
			want: want{
				headers: http.Header{},
			},
		}, {
			name: "vast 4 protocol",
			args: args{
				req: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						IP:       "1.1.1.1",
						UA:       "user-agent",
						Language: "en",
					},
					Site: &openrtb2.Site{
						Page: "http://test.com/",
					},
					Imp: []openrtb2.Imp{
						{
							Video: &openrtb2.Video{
								Protocols: []adcom1.MediaCreativeSubtype{
									adcom1.CreativeVAST40,
									adcom1.CreativeDAAST10,
								},
							},
						},
					},
				},
			},
			want: want{
				headers: http.Header{
					"X-Device-Ip":              []string{"1.1.1.1"},
					"X-Device-User-Agent":      []string{"user-agent"},
					"X-Device-Referer":         []string{"http://test.com/"},
					"X-Device-Accept-Language": []string{"en"},
				},
			},
		}, {
			name: "< vast 4",
			args: args{
				req: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						IP:       "1.1.1.1",
						UA:       "user-agent",
						Language: "en",
					},
					Site: &openrtb2.Site{
						Page: "http://test.com/",
					},
					Imp: []openrtb2.Imp{
						{
							Video: &openrtb2.Video{
								Protocols: []adcom1.MediaCreativeSubtype{
									adcom1.CreativeVAST20,
									adcom1.CreativeDAAST10,
								},
							},
						},
					},
				},
			},
			want: want{
				headers: http.Header{
					"X-Forwarded-For": []string{"1.1.1.1"},
					"User-Agent":      []string{"user-agent"},
				},
			},
		}, {
			name: "vast 4.0 and 4.0 wrapper",
			args: args{
				req: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						IP:       "1.1.1.1",
						UA:       "user-agent",
						Language: "en",
					},
					Site: &openrtb2.Site{
						Page: "http://test.com/",
					},
					Imp: []openrtb2.Imp{
						{
							Video: &openrtb2.Video{
								Protocols: []adcom1.MediaCreativeSubtype{
									adcom1.CreativeVAST40,
									adcom1.CreativeVAST40Wrapper,
								},
							},
						},
					},
				},
			},
			want: want{
				headers: http.Header{
					"X-Device-Ip":              []string{"1.1.1.1"},
					"X-Device-User-Agent":      []string{"user-agent"},
					"X-Device-Referer":         []string{"http://test.com/"},
					"X-Device-Accept-Language": []string{"en"},
				},
			},
		},
		{
			name: "vast 2.0 and 4.0",
			args: args{
				req: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						IP:       "1.1.1.1",
						UA:       "user-agent",
						Language: "en",
					},
					Site: &openrtb2.Site{
						Page: "http://test.com/",
					},
					Imp: []openrtb2.Imp{
						{
							Video: &openrtb2.Video{
								Protocols: []adcom1.MediaCreativeSubtype{
									adcom1.CreativeVAST40,
									adcom1.CreativeVAST20Wrapper,
								},
							},
						},
					},
				},
			},
			want: want{
				headers: http.Header{
					"X-Device-Ip":              []string{"1.1.1.1"},
					"X-Forwarded-For":          []string{"1.1.1.1"},
					"X-Device-User-Agent":      []string{"user-agent"},
					"User-Agent":               []string{"user-agent"},
					"X-Device-Referer":         []string{"http://test.com/"},
					"X-Device-Accept-Language": []string{"en"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := new(BidderMacro)
			tag.IBidderMacro = tag
			tag.IsApp = false
			tag.Request = tt.args.req
			if nil != tt.args.req && nil != tt.args.req.Imp && len(tt.args.req.Imp) > 0 {
				tag.Imp = &tt.args.req.Imp[0]
			}
			setDefaultHeaders(tag)
			assert.Equal(t, tt.want.headers, tag.ImpReqHeaders)
		})
	}
}

// TestGetAllHeaders verifies default and custom headers are returned
func TestGetAllHeaders(t *testing.T) {
	type args struct {
		req      *openrtb2.BidRequest
		myBidder IBidderMacro
	}
	type want struct {
		headers http.Header
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Default and custom headers check",
			args: args{
				req: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						IP:       "1.1.1.1",
						UA:       "user-agent",
						Language: "en",
					},
					Site: &openrtb2.Site{
						Page: "http://test.com/",
					},
				},
				myBidder: newMyVastBidderMacro(map[string]string{
					"my-custom-header": "some-value",
				}),
			},
			want: want{
				headers: http.Header{
					"X-Device-Ip":              []string{"1.1.1.1"},
					"X-Forwarded-For":          []string{"1.1.1.1"},
					"X-Device-User-Agent":      []string{"user-agent"},
					"User-Agent":               []string{"user-agent"},
					"X-Device-Referer":         []string{"http://test.com/"},
					"X-Device-Accept-Language": []string{"en"},
					"My-Custom-Header":         []string{"some-value"},
				},
			},
		},
		{
			name: "override default header value",
			args: args{
				req: &openrtb2.BidRequest{
					Site: &openrtb2.Site{
						Page: "http://test.com/", // default header value
					},
				},
				myBidder: newMyVastBidderMacro(map[string]string{
					"X-Device-Referer": "my-custom-value",
				}),
			},
			want: want{
				headers: http.Header{
					// http://test.com/ is not expected here as value
					"X-Device-Referer": []string{"my-custom-value"},
				},
			},
		},
		{
			name: "no custom headers",
			args: args{
				req: &openrtb2.BidRequest{
					Device: &openrtb2.Device{
						IP:       "1.1.1.1",
						UA:       "user-agent",
						Language: "en",
					},
					Site: &openrtb2.Site{
						Page: "http://test.com/",
					},
				},
				myBidder: newMyVastBidderMacro(nil), // nil - no custom headers
			},
			want: want{
				headers: http.Header{ // expect default headers
					"X-Device-Ip":              []string{"1.1.1.1"},
					"X-Forwarded-For":          []string{"1.1.1.1"},
					"X-Device-User-Agent":      []string{"user-agent"},
					"User-Agent":               []string{"user-agent"},
					"X-Device-Referer":         []string{"http://test.com/"},
					"X-Device-Accept-Language": []string{"en"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := tt.args.myBidder
			tag.(*myVastBidderMacro).Request = tt.args.req
			allHeaders := tag.getAllHeaders()
			assert.Equal(t, tt.want.headers, allHeaders)
		})
	}
}

type myVastBidderMacro struct {
	*BidderMacro
	customHeaders map[string]string
}

func newMyVastBidderMacro(customHeaders map[string]string) IBidderMacro {
	obj := &myVastBidderMacro{
		BidderMacro:   &BidderMacro{},
		customHeaders: customHeaders,
	}
	obj.IBidderMacro = obj
	return obj
}

func (tag *myVastBidderMacro) GetHeaders() http.Header {
	if nil == tag.customHeaders {
		return nil
	}
	h := http.Header{}
	for k, v := range tag.customHeaders {
		h.Set(k, v)
	}
	return h
}

type testBidderMacro struct {
	*BidderMacro
}

func (tag *testBidderMacro) MacroCacheBuster(key string) string {
	return `cachebuster`
}

func newTestBidderMacro() IBidderMacro {
	obj := &testBidderMacro{
		BidderMacro: &BidderMacro{},
	}
	obj.IBidderMacro = obj
	return obj
}

func TestBidderMacro_MacroTest(t *testing.T) {
	type args struct {
		tag        IBidderMacro
		conf       *config.Adapter
		bidRequest *openrtb2.BidRequest
	}
	tests := []struct {
		name   string
		args   args
		macros map[string]string
	}{
		{
			name: `App:EmptyBasicRequest`,
			args: args{
				tag:  newTestBidderMacro(),
				conf: &config.Adapter{},
				bidRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{
							Video: &openrtb2.Video{},
						},
					},
					App: &openrtb2.App{
						Publisher: &openrtb2.Publisher{},
					},
				},
			},
			macros: map[string]string{
				MacroTest:                      ``,
				MacroTimeout:                   ``,
				MacroWhitelistSeat:             ``,
				MacroWhitelistLang:             ``,
				MacroBlockedSeat:               ``,
				MacroCurrency:                  ``,
				MacroBlockedCategory:           ``,
				MacroBlockedAdvertiser:         ``,
				MacroBlockedApp:                ``,
				MacroFD:                        ``,
				MacroTransactionID:             ``,
				MacroPaymentIDChain:            ``,
				MacroCoppa:                     ``,
				MacroDisplayManager:            ``,
				MacroDisplayManagerVersion:     ``,
				MacroInterstitial:              ``,
				MacroTagID:                     ``,
				MacroBidFloor:                  ``,
				MacroBidFloorCurrency:          ``,
				MacroSecure:                    ``,
				MacroPMP:                       ``,
				MacroVideoMIMES:                ``,
				MacroVideoMinimumDuration:      ``,
				MacroVideoMaximumDuration:      ``,
				MacroVideoProtocols:            ``,
				MacroVideoPlayerWidth:          ``,
				MacroVideoPlayerHeight:         ``,
				MacroVideoStartDelay:           ``,
				MacroVideoPlacement:            ``,
				MacroVideoLinearity:            ``,
				MacroVideoSkip:                 ``,
				MacroVideoSkipMinimum:          ``,
				MacroVideoSkipAfter:            ``,
				MacroVideoSequence:             ``,
				MacroVideoBlockedAttribute:     ``,
				MacroVideoMaximumExtended:      ``,
				MacroVideoMinimumBitRate:       ``,
				MacroVideoMaximumBitRate:       ``,
				MacroVideoBoxing:               ``,
				MacroVideoPlaybackMethod:       ``,
				MacroVideoDelivery:             ``,
				MacroVideoPosition:             ``,
				MacroVideoAPI:                  ``,
				MacroSiteID:                    ``,
				MacroSiteName:                  ``,
				MacroSitePage:                  ``,
				MacroSiteReferrer:              ``,
				MacroSiteSearch:                ``,
				MacroSiteMobile:                ``,
				MacroAppID:                     ``,
				MacroAppName:                   ``,
				MacroAppBundle:                 ``,
				MacroAppStoreURL:               ``,
				MacroAppVersion:                ``,
				MacroAppPaid:                   ``,
				MacroCategory:                  ``,
				MacroDomain:                    ``,
				MacroSectionCategory:           ``,
				MacroPageCategory:              ``,
				MacroPrivacyPolicy:             ``,
				MacroKeywords:                  ``,
				MacroPubID:                     ``,
				MacroPubName:                   ``,
				MacroPubDomain:                 ``,
				MacroContentID:                 ``,
				MacroContentEpisode:            ``,
				MacroContentTitle:              ``,
				MacroContentSeries:             ``,
				MacroContentSeason:             ``,
				MacroContentArtist:             ``,
				MacroContentGenre:              ``,
				MacroContentAlbum:              ``,
				MacroContentISrc:               ``,
				MacroContentURL:                ``,
				MacroContentCategory:           ``,
				MacroContentProductionQuality:  ``,
				MacroContentVideoQuality:       ``,
				MacroContentContext:            ``,
				MacroContentContentRating:      ``,
				MacroContentUserRating:         ``,
				MacroContentQAGMediaRating:     ``,
				MacroContentKeywords:           ``,
				MacroContentLiveStream:         ``,
				MacroContentSourceRelationship: ``,
				MacroContentLength:             ``,
				MacroContentLanguage:           ``,
				MacroContentEmbeddable:         ``,
				MacroProducerID:                ``,
				MacroProducerName:              ``,
				MacroUserAgent:                 ``,
				MacroDNT:                       ``,
				MacroLMT:                       ``,
				MacroIP:                        ``,
				MacroDeviceType:                ``,
				MacroMake:                      ``,
				MacroModel:                     ``,
				MacroDeviceOS:                  ``,
				MacroDeviceOSVersion:           ``,
				MacroDeviceWidth:               ``,
				MacroDeviceHeight:              ``,
				MacroDeviceJS:                  ``,
				MacroDeviceLanguage:            ``,
				MacroDeviceIFA:                 ``,
				MacroDeviceIFAType:             ``,
				MacroDeviceDIDSHA1:             ``,
				MacroDeviceDIDMD5:              ``,
				MacroDeviceDPIDSHA1:            ``,
				MacroDeviceDPIDMD5:             ``,
				MacroDeviceMACSHA1:             ``,
				MacroDeviceMACMD5:              ``,
				MacroLatitude:                  ``,
				MacroLongitude:                 ``,
				MacroCountry:                   ``,
				MacroRegion:                    ``,
				MacroCity:                      ``,
				MacroZip:                       ``,
				MacroUTCOffset:                 ``,
				MacroUserID:                    ``,
				MacroYearOfBirth:               ``,
				MacroGender:                    ``,
				MacroGDPRConsent:               ``,
				MacroGDPR:                      ``,
				MacroUSPrivacy:                 ``,
				MacroCacheBuster:               `cachebuster`,
			},
		},
		{
			name: `Site:EmptyBasicRequest`,
			args: args{
				tag:  newTestBidderMacro(),
				conf: &config.Adapter{},
				bidRequest: &openrtb2.BidRequest{
					Imp: []openrtb2.Imp{
						{
							Video: &openrtb2.Video{},
						},
					},
					Site: &openrtb2.Site{
						Publisher: &openrtb2.Publisher{},
					},
				},
			},
			macros: map[string]string{
				MacroTest:                      ``,
				MacroTimeout:                   ``,
				MacroWhitelistSeat:             ``,
				MacroWhitelistLang:             ``,
				MacroBlockedSeat:               ``,
				MacroCurrency:                  ``,
				MacroBlockedCategory:           ``,
				MacroBlockedAdvertiser:         ``,
				MacroBlockedApp:                ``,
				MacroFD:                        ``,
				MacroTransactionID:             ``,
				MacroPaymentIDChain:            ``,
				MacroCoppa:                     ``,
				MacroDisplayManager:            ``,
				MacroDisplayManagerVersion:     ``,
				MacroInterstitial:              ``,
				MacroTagID:                     ``,
				MacroBidFloor:                  ``,
				MacroBidFloorCurrency:          ``,
				MacroSecure:                    ``,
				MacroPMP:                       ``,
				MacroVideoMIMES:                ``,
				MacroVideoMinimumDuration:      ``,
				MacroVideoMaximumDuration:      ``,
				MacroVideoProtocols:            ``,
				MacroVideoPlayerWidth:          ``,
				MacroVideoPlayerHeight:         ``,
				MacroVideoStartDelay:           ``,
				MacroVideoPlacement:            ``,
				MacroVideoLinearity:            ``,
				MacroVideoSkip:                 ``,
				MacroVideoSkipMinimum:          ``,
				MacroVideoSkipAfter:            ``,
				MacroVideoSequence:             ``,
				MacroVideoBlockedAttribute:     ``,
				MacroVideoMaximumExtended:      ``,
				MacroVideoMinimumBitRate:       ``,
				MacroVideoMaximumBitRate:       ``,
				MacroVideoBoxing:               ``,
				MacroVideoPlaybackMethod:       ``,
				MacroVideoDelivery:             ``,
				MacroVideoPosition:             ``,
				MacroVideoAPI:                  ``,
				MacroSiteID:                    ``,
				MacroSiteName:                  ``,
				MacroSitePage:                  ``,
				MacroSiteReferrer:              ``,
				MacroSiteSearch:                ``,
				MacroSiteMobile:                ``,
				MacroAppID:                     ``,
				MacroAppName:                   ``,
				MacroAppBundle:                 ``,
				MacroAppStoreURL:               ``,
				MacroAppVersion:                ``,
				MacroAppPaid:                   ``,
				MacroCategory:                  ``,
				MacroDomain:                    ``,
				MacroSectionCategory:           ``,
				MacroPageCategory:              ``,
				MacroPrivacyPolicy:             ``,
				MacroKeywords:                  ``,
				MacroPubID:                     ``,
				MacroPubName:                   ``,
				MacroPubDomain:                 ``,
				MacroContentID:                 ``,
				MacroContentEpisode:            ``,
				MacroContentTitle:              ``,
				MacroContentSeries:             ``,
				MacroContentSeason:             ``,
				MacroContentArtist:             ``,
				MacroContentGenre:              ``,
				MacroContentAlbum:              ``,
				MacroContentISrc:               ``,
				MacroContentURL:                ``,
				MacroContentCategory:           ``,
				MacroContentProductionQuality:  ``,
				MacroContentVideoQuality:       ``,
				MacroContentContext:            ``,
				MacroContentContentRating:      ``,
				MacroContentUserRating:         ``,
				MacroContentQAGMediaRating:     ``,
				MacroContentKeywords:           ``,
				MacroContentLiveStream:         ``,
				MacroContentSourceRelationship: ``,
				MacroContentLength:             ``,
				MacroContentLanguage:           ``,
				MacroContentEmbeddable:         ``,
				MacroProducerID:                ``,
				MacroProducerName:              ``,
				MacroUserAgent:                 ``,
				MacroDNT:                       ``,
				MacroLMT:                       ``,
				MacroIP:                        ``,
				MacroDeviceType:                ``,
				MacroMake:                      ``,
				MacroModel:                     ``,
				MacroDeviceOS:                  ``,
				MacroDeviceOSVersion:           ``,
				MacroDeviceWidth:               ``,
				MacroDeviceHeight:              ``,
				MacroDeviceJS:                  ``,
				MacroDeviceLanguage:            ``,
				MacroDeviceIFA:                 ``,
				MacroDeviceIFAType:             ``,
				MacroDeviceDIDSHA1:             ``,
				MacroDeviceDIDMD5:              ``,
				MacroDeviceDPIDSHA1:            ``,
				MacroDeviceDPIDMD5:             ``,
				MacroDeviceMACSHA1:             ``,
				MacroDeviceMACMD5:              ``,
				MacroLatitude:                  ``,
				MacroLongitude:                 ``,
				MacroCountry:                   ``,
				MacroRegion:                    ``,
				MacroCity:                      ``,
				MacroZip:                       ``,
				MacroUTCOffset:                 ``,
				MacroUserID:                    ``,
				MacroYearOfBirth:               ``,
				MacroGender:                    ``,
				MacroGDPRConsent:               ``,
				MacroGDPR:                      ``,
				MacroUSPrivacy:                 ``,
				MacroCacheBuster:               `cachebuster`,
			},
		},
		{
			name: `Site:RequestLevelMacros`,
			args: args{
				tag:  newTestBidderMacro(),
				conf: &config.Adapter{},
				bidRequest: &openrtb2.BidRequest{
					Test:  1,
					TMax:  1000,
					WSeat: []string{`wseat-1`, `wseat-2`},
					WLang: []string{`wlang-1`, `wlang-2`},
					BSeat: []string{`bseat-1`, `bseat-2`},
					Cur:   []string{`usd`, `inr`},
					BCat:  []string{`bcat-1`, `bcat-2`},
					BAdv:  []string{`badv-1`, `badv-2`},
					BApp:  []string{`bapp-1`, `bapp-2`},
					Source: &openrtb2.Source{
						FD:     ptrutil.ToPtr[int8](1),
						TID:    `source-tid`,
						PChain: `source-pchain`,
					},
					Regs: &openrtb2.Regs{
						COPPA: 1,
						Ext:   []byte(`{"gdpr":1,"us_privacy":"user-privacy"}`),
					},
					Imp: []openrtb2.Imp{
						{
							DisplayManager:    `disp-mgr`,
							DisplayManagerVer: `1.2`,
							Instl:             1,
							TagID:             `tag-id`,
							BidFloor:          3.0,
							BidFloorCur:       `usd`,
							Secure:            new(int8),
							PMP: &openrtb2.PMP{
								PrivateAuction: 1,
								Deals: []openrtb2.Deal{
									{
										ID:          `deal-1`,
										BidFloor:    4.0,
										BidFloorCur: `usd`,
										AT:          1,
										WSeat:       []string{`wseat-11`, `wseat-12`},
										WADomain:    []string{`wdomain-11`, `wdomain-12`},
									},
									{
										ID:          `deal-2`,
										BidFloor:    5.0,
										BidFloorCur: `inr`,
										AT:          1,
										WSeat:       []string{`wseat-21`, `wseat-22`},
										WADomain:    []string{`wdomain-21`, `wdomain-22`},
									},
								},
							},
							Video: &openrtb2.Video{
								MIMEs:          []string{`mp4`, `flv`},
								MinDuration:    30,
								MaxDuration:    60,
								Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST30, adcom1.CreativeVAST40Wrapper},
								Protocol:       adcom1.CreativeVAST40Wrapper,
								W:              ptrutil.ToPtr[int64](640),
								H:              ptrutil.ToPtr[int64](480),
								StartDelay:     new(adcom1.StartDelay),
								Placement:      adcom1.VideoPlacementInStream,
								Linearity:      adcom1.LinearityLinear,
								Skip:           new(int8),
								SkipMin:        10,
								SkipAfter:      5,
								Sequence:       1,
								BAttr:          []adcom1.CreativeAttribute{adcom1.AttrAudioAuto, adcom1.AttrAudioUser},
								MaxExtended:    10,
								MinBitRate:     360,
								MaxBitRate:     1080,
								BoxingAllowed:  ptrutil.ToPtr[int8](1),
								PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackPageLoadSoundOn, adcom1.PlaybackClickSoundOn},
								PlaybackEnd:    adcom1.PlaybackCompletion,
								Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryStreaming, adcom1.DeliveryDownload},
								Pos:            new(adcom1.PlacementPosition),
								API:            []adcom1.APIFramework{adcom1.APIVPAID10, adcom1.APIVPAID20},
							},
						},
					},
					Site: &openrtb2.Site{
						ID:            `site-id`,
						Name:          `site-name`,
						Domain:        `site-domain`,
						Cat:           []string{`site-cat1`, `site-cat2`},
						SectionCat:    []string{`site-sec-cat1`, `site-sec-cat2`},
						PageCat:       []string{`site-page-cat1`, `site-page-cat2`},
						Page:          `site-page-url`,
						Ref:           `site-referer-url`,
						Search:        `site-search-keywords`,
						Mobile:        ptrutil.ToPtr[int8](1),
						PrivacyPolicy: ptrutil.ToPtr[int8](2),
						Keywords:      `site-keywords`,
						Publisher: &openrtb2.Publisher{
							ID:     `site-pub-id`,
							Name:   `site-pub-name`,
							Domain: `site-pub-domain`,
						},
						Content: &openrtb2.Content{
							ID:                 `site-cnt-id`,
							Episode:            2,
							Title:              `site-cnt-title`,
							Series:             `site-cnt-series`,
							Season:             `site-cnt-season`,
							Artist:             `site-cnt-artist`,
							Genre:              `site-cnt-genre`,
							Album:              `site-cnt-album`,
							ISRC:               `site-cnt-isrc`,
							URL:                `site-cnt-url`,
							Cat:                []string{`site-cnt-cat1`, `site-cnt-cat2`},
							ProdQ:              new(adcom1.ProductionQuality),
							VideoQuality:       new(adcom1.ProductionQuality),
							Context:            adcom1.ContentVideo,
							ContentRating:      `1.2`,
							UserRating:         `2.2`,
							QAGMediaRating:     adcom1.MediaRatingAll,
							Keywords:           `site-cnt-keywords`,
							LiveStream:         ptrutil.ToPtr[int8](1),
							SourceRelationship: ptrutil.ToPtr[int8](1),
							Len:                100,
							Language:           `english`,
							Embeddable:         ptrutil.ToPtr[int8](1),
							Producer: &openrtb2.Producer{
								ID:   `site-cnt-prod-id`,
								Name: `site-cnt-prod-name`,
							},
						},
					},
					Device: &openrtb2.Device{
						UA:             `user-agent`,
						DNT:            new(int8),
						Lmt:            new(int8),
						IP:             `ipv4`,
						IPv6:           `ipv6`,
						DeviceType:     adcom1.DeviceTV,
						Make:           `device-make`,
						Model:          `device-model`,
						OS:             `os`,
						OSV:            `os-version`,
						H:              1024,
						W:              2048,
						JS:             ptrutil.ToPtr[int8](1),
						Language:       `device-lang`,
						ConnectionType: new(adcom1.ConnectionType),
						IFA:            `ifa`,
						DIDSHA1:        `didsha1`,
						DIDMD5:         `didmd5`,
						DPIDSHA1:       `dpidsha1`,
						DPIDMD5:        `dpidmd5`,
						MACSHA1:        `macsha1`,
						MACMD5:         `macmd5`,
						Geo: &openrtb2.Geo{
							Lat:       ptrutil.ToPtr[float64](1.1),
							Lon:       ptrutil.ToPtr[float64](2.2),
							Country:   `country`,
							Region:    `region`,
							City:      `city`,
							ZIP:       `zip`,
							UTCOffset: 1000,
						},
						Ext: []byte(`{"ifa_type":"idfa"}`),
					},
					User: &openrtb2.User{
						ID:     `user-id`,
						Yob:    1990,
						Gender: `M`,
						Ext:    []byte(`{"consent":"user-gdpr-consent"}`),
					},
				},
			},
			macros: map[string]string{
				MacroTest:                      `1`,
				MacroTimeout:                   `1000`,
				MacroWhitelistSeat:             `wseat-1,wseat-2`,
				MacroWhitelistLang:             `wlang-1,wlang-2`,
				MacroBlockedSeat:               `bseat-1,bseat-2`,
				MacroCurrency:                  `usd,inr`,
				MacroBlockedCategory:           `bcat-1,bcat-2`,
				MacroBlockedAdvertiser:         `badv-1,badv-2`,
				MacroBlockedApp:                `bapp-1,bapp-2`,
				MacroFD:                        `1`,
				MacroTransactionID:             `source-tid`,
				MacroPaymentIDChain:            `source-pchain`,
				MacroCoppa:                     `1`,
				MacroDisplayManager:            `disp-mgr`,
				MacroDisplayManagerVersion:     `1.2`,
				MacroInterstitial:              `1`,
				MacroTagID:                     `tag-id`,
				MacroBidFloor:                  `3`,
				MacroBidFloorCurrency:          `usd`,
				MacroSecure:                    `0`,
				MacroPMP:                       `{"private_auction":1,"deals":[{"id":"deal-1","bidfloor":4,"bidfloorcur":"usd","at":1,"wseat":["wseat-11","wseat-12"],"wadomain":["wdomain-11","wdomain-12"]},{"id":"deal-2","bidfloor":5,"bidfloorcur":"inr","at":1,"wseat":["wseat-21","wseat-22"],"wadomain":["wdomain-21","wdomain-22"]}]}`,
				MacroVideoMIMES:                `mp4,flv`,
				MacroVideoMinimumDuration:      `30`,
				MacroVideoMaximumDuration:      `60`,
				MacroVideoProtocols:            `3,8`,
				MacroVideoPlayerWidth:          `640`,
				MacroVideoPlayerHeight:         `480`,
				MacroVideoStartDelay:           `0`,
				MacroVideoPlacement:            `1`,
				MacroVideoLinearity:            `1`,
				MacroVideoSkip:                 `0`,
				MacroVideoSkipMinimum:          `10`,
				MacroVideoSkipAfter:            `5`,
				MacroVideoSequence:             `1`,
				MacroVideoBlockedAttribute:     `1,2`,
				MacroVideoMaximumExtended:      `10`,
				MacroVideoMinimumBitRate:       `360`,
				MacroVideoMaximumBitRate:       `1080`,
				MacroVideoBoxing:               `1`,
				MacroVideoPlaybackMethod:       `1,3`,
				MacroVideoDelivery:             `1,3`,
				MacroVideoPosition:             `0`,
				MacroVideoAPI:                  `1,2`,
				MacroSiteID:                    `site-id`,
				MacroSiteName:                  `site-name`,
				MacroSitePage:                  `site-page-url`,
				MacroSiteReferrer:              `site-referer-url`,
				MacroSiteSearch:                `site-search-keywords`,
				MacroSiteMobile:                `1`,
				MacroAppID:                     ``,
				MacroAppName:                   ``,
				MacroAppBundle:                 ``,
				MacroAppStoreURL:               ``,
				MacroAppVersion:                ``,
				MacroAppPaid:                   ``,
				MacroCategory:                  `site-cat1,site-cat2`,
				MacroDomain:                    `site-domain`,
				MacroSectionCategory:           `site-sec-cat1,site-sec-cat2`,
				MacroPageCategory:              `site-page-cat1,site-page-cat2`,
				MacroPrivacyPolicy:             `2`,
				MacroKeywords:                  `site-keywords`,
				MacroPubID:                     `site-pub-id`,
				MacroPubName:                   `site-pub-name`,
				MacroPubDomain:                 `site-pub-domain`,
				MacroContentID:                 `site-cnt-id`,
				MacroContentEpisode:            `2`,
				MacroContentTitle:              `site-cnt-title`,
				MacroContentSeries:             `site-cnt-series`,
				MacroContentSeason:             `site-cnt-season`,
				MacroContentArtist:             `site-cnt-artist`,
				MacroContentGenre:              `site-cnt-genre`,
				MacroContentAlbum:              `site-cnt-album`,
				MacroContentISrc:               `site-cnt-isrc`,
				MacroContentURL:                `site-cnt-url`,
				MacroContentCategory:           `site-cnt-cat1,site-cnt-cat2`,
				MacroContentProductionQuality:  `0`,
				MacroContentVideoQuality:       `0`,
				MacroContentContext:            `1`,
				MacroContentContentRating:      `1.2`,
				MacroContentUserRating:         `2.2`,
				MacroContentQAGMediaRating:     `1`,
				MacroContentKeywords:           `site-cnt-keywords`,
				MacroContentLiveStream:         `1`,
				MacroContentSourceRelationship: `1`,
				MacroContentLength:             `100`,
				MacroContentLanguage:           `english`,
				MacroContentEmbeddable:         `1`,
				MacroProducerID:                `site-cnt-prod-id`,
				MacroProducerName:              `site-cnt-prod-name`,
				MacroUserAgent:                 `user-agent`,
				MacroDNT:                       `0`,
				MacroLMT:                       `0`,
				MacroIP:                        `ipv4`,
				MacroDeviceType:                `3`,
				MacroMake:                      `device-make`,
				MacroModel:                     `device-model`,
				MacroDeviceOS:                  `os`,
				MacroDeviceOSVersion:           `os-version`,
				MacroDeviceWidth:               `2048`,
				MacroDeviceHeight:              `1024`,
				MacroDeviceJS:                  `1`,
				MacroDeviceLanguage:            `device-lang`,
				MacroDeviceIFA:                 `ifa`,
				MacroDeviceIFAType:             `idfa`,
				MacroDeviceDIDSHA1:             `didsha1`,
				MacroDeviceDIDMD5:              `didmd5`,
				MacroDeviceDPIDSHA1:            `dpidsha1`,
				MacroDeviceDPIDMD5:             `dpidmd5`,
				MacroDeviceMACSHA1:             `macsha1`,
				MacroDeviceMACMD5:              `macmd5`,
				MacroLatitude:                  `1.1`,
				MacroLongitude:                 `2.2`,
				MacroCountry:                   `country`,
				MacroRegion:                    `region`,
				MacroCity:                      `city`,
				MacroZip:                       `zip`,
				MacroUTCOffset:                 `1000`,
				MacroUserID:                    `user-id`,
				MacroYearOfBirth:               `1990`,
				MacroGender:                    `M`,
				MacroGDPRConsent:               `user-gdpr-consent`,
				MacroGDPR:                      `1`,
				MacroUSPrivacy:                 `user-privacy`,
				MacroCacheBuster:               `cachebuster`,
			},
		},
		{
			name: `App:RequestLevelMacros`,
			args: args{
				tag:  newTestBidderMacro(),
				conf: &config.Adapter{},
				bidRequest: &openrtb2.BidRequest{
					Test:  1,
					TMax:  1000,
					WSeat: []string{`wseat-1`, `wseat-2`},
					WLang: []string{`wlang-1`, `wlang-2`},
					BSeat: []string{`bseat-1`, `bseat-2`},
					Cur:   []string{`usd`, `inr`},
					BCat:  []string{`bcat-1`, `bcat-2`},
					BAdv:  []string{`badv-1`, `badv-2`},
					BApp:  []string{`bapp-1`, `bapp-2`},
					Source: &openrtb2.Source{
						FD:     ptrutil.ToPtr[int8](1),
						TID:    `source-tid`,
						PChain: `source-pchain`,
					},
					Regs: &openrtb2.Regs{
						COPPA: 1,
						Ext:   []byte(`{"gdpr":1,"us_privacy":"user-privacy"}`),
					},
					Imp: []openrtb2.Imp{
						{
							DisplayManager:    `disp-mgr`,
							DisplayManagerVer: `1.2`,
							Instl:             1,
							TagID:             `tag-id`,
							BidFloor:          3.0,
							BidFloorCur:       `usd`,
							Secure:            new(int8),
							PMP: &openrtb2.PMP{
								PrivateAuction: 1,
								Deals: []openrtb2.Deal{
									{
										ID:          `deal-1`,
										BidFloor:    4.0,
										BidFloorCur: `usd`,
										AT:          1,
										WSeat:       []string{`wseat-11`, `wseat-12`},
										WADomain:    []string{`wdomain-11`, `wdomain-12`},
									},
									{
										ID:          `deal-2`,
										BidFloor:    5.0,
										BidFloorCur: `inr`,
										AT:          1,
										WSeat:       []string{`wseat-21`, `wseat-22`},
										WADomain:    []string{`wdomain-21`, `wdomain-22`},
									},
								},
							},
							Video: &openrtb2.Video{
								MIMEs:          []string{`mp4`, `flv`},
								MinDuration:    30,
								MaxDuration:    60,
								Protocols:      []adcom1.MediaCreativeSubtype{adcom1.CreativeVAST30, adcom1.CreativeVAST40Wrapper},
								Protocol:       adcom1.CreativeVAST40Wrapper,
								W:              ptrutil.ToPtr[int64](640),
								H:              ptrutil.ToPtr[int64](480),
								StartDelay:     new(adcom1.StartDelay),
								Placement:      adcom1.VideoPlacementInStream,
								Linearity:      adcom1.LinearityLinear,
								Skip:           new(int8),
								SkipMin:        10,
								SkipAfter:      5,
								Sequence:       1,
								BAttr:          []adcom1.CreativeAttribute{adcom1.AttrAudioAuto, adcom1.AttrAudioUser},
								MaxExtended:    10,
								MinBitRate:     360,
								MaxBitRate:     1080,
								BoxingAllowed:  ptrutil.ToPtr[int8](1),
								PlaybackMethod: []adcom1.PlaybackMethod{adcom1.PlaybackPageLoadSoundOn, adcom1.PlaybackClickSoundOn},
								PlaybackEnd:    adcom1.PlaybackCompletion,
								Delivery:       []adcom1.DeliveryMethod{adcom1.DeliveryStreaming, adcom1.DeliveryDownload},
								Pos:            new(adcom1.PlacementPosition),
								API:            []adcom1.APIFramework{adcom1.APIVPAID10, adcom1.APIVPAID20},
							},
						},
					},
					App: &openrtb2.App{
						ID:            `app-id`,
						Bundle:        `app-bundle`,
						StoreURL:      `app-store-url`,
						Ver:           `app-version`,
						Paid:          ptrutil.ToPtr[int8](1),
						Name:          `app-name`,
						Domain:        `app-domain`,
						Cat:           []string{`app-cat1`, `app-cat2`},
						SectionCat:    []string{`app-sec-cat1`, `app-sec-cat2`},
						PageCat:       []string{`app-page-cat1`, `app-page-cat2`},
						PrivacyPolicy: ptrutil.ToPtr[int8](2),
						Keywords:      `app-keywords`,
						Publisher: &openrtb2.Publisher{
							ID:     `app-pub-id`,
							Name:   `app-pub-name`,
							Domain: `app-pub-domain`,
						},
						Content: &openrtb2.Content{
							ID:                 `app-cnt-id`,
							Episode:            2,
							Title:              `app-cnt-title`,
							Series:             `app-cnt-series`,
							Season:             `app-cnt-season`,
							Artist:             `app-cnt-artist`,
							Genre:              `app-cnt-genre`,
							Album:              `app-cnt-album`,
							ISRC:               `app-cnt-isrc`,
							URL:                `app-cnt-url`,
							Cat:                []string{`app-cnt-cat1`, `app-cnt-cat2`},
							ProdQ:              new(adcom1.ProductionQuality),
							VideoQuality:       new(adcom1.ProductionQuality),
							Context:            adcom1.ContentVideo,
							ContentRating:      `1.2`,
							UserRating:         `2.2`,
							QAGMediaRating:     adcom1.MediaRatingAll,
							Keywords:           `app-cnt-keywords`,
							LiveStream:         ptrutil.ToPtr[int8](1),
							SourceRelationship: ptrutil.ToPtr[int8](1),
							Len:                100,
							Language:           `english`,
							Embeddable:         ptrutil.ToPtr[int8](1),
							Producer: &openrtb2.Producer{
								ID:   `app-cnt-prod-id`,
								Name: `app-cnt-prod-name`,
							},
						},
					},
					Device: &openrtb2.Device{
						UA:             `user-agent`,
						DNT:            new(int8),
						Lmt:            new(int8),
						IPv6:           `ipv6`,
						DeviceType:     adcom1.DeviceTV,
						Make:           `device-make`,
						Model:          `device-model`,
						OS:             `os`,
						OSV:            `os-version`,
						H:              1024,
						W:              2048,
						JS:             ptrutil.ToPtr[int8](1),
						Language:       `device-lang`,
						ConnectionType: new(adcom1.ConnectionType),
						IFA:            `ifa`,
						DIDSHA1:        `didsha1`,
						DIDMD5:         `didmd5`,
						DPIDSHA1:       `dpidsha1`,
						DPIDMD5:        `dpidmd5`,
						MACSHA1:        `macsha1`,
						MACMD5:         `macmd5`,
						Geo: &openrtb2.Geo{
							Lat:       ptrutil.ToPtr[float64](1.1),
							Lon:       ptrutil.ToPtr[float64](2.2),
							Country:   `country`,
							Region:    `region`,
							City:      `city`,
							ZIP:       `zip`,
							UTCOffset: 1000,
						},
						Ext: []byte(`{"ifa_type":"idfa"}`),
					},
					User: &openrtb2.User{
						ID:     `user-id`,
						Yob:    1990,
						Gender: `M`,
						Ext:    []byte(`{"consent":"user-gdpr-consent"}`),
					},
				},
			},
			macros: map[string]string{
				MacroTest:                      `1`,
				MacroTimeout:                   `1000`,
				MacroWhitelistSeat:             `wseat-1,wseat-2`,
				MacroWhitelistLang:             `wlang-1,wlang-2`,
				MacroBlockedSeat:               `bseat-1,bseat-2`,
				MacroCurrency:                  `usd,inr`,
				MacroBlockedCategory:           `bcat-1,bcat-2`,
				MacroBlockedAdvertiser:         `badv-1,badv-2`,
				MacroBlockedApp:                `bapp-1,bapp-2`,
				MacroFD:                        `1`,
				MacroTransactionID:             `source-tid`,
				MacroPaymentIDChain:            `source-pchain`,
				MacroCoppa:                     `1`,
				MacroDisplayManager:            `disp-mgr`,
				MacroDisplayManagerVersion:     `1.2`,
				MacroInterstitial:              `1`,
				MacroTagID:                     `tag-id`,
				MacroBidFloor:                  `3`,
				MacroBidFloorCurrency:          `usd`,
				MacroSecure:                    `0`,
				MacroPMP:                       `{"private_auction":1,"deals":[{"id":"deal-1","bidfloor":4,"bidfloorcur":"usd","at":1,"wseat":["wseat-11","wseat-12"],"wadomain":["wdomain-11","wdomain-12"]},{"id":"deal-2","bidfloor":5,"bidfloorcur":"inr","at":1,"wseat":["wseat-21","wseat-22"],"wadomain":["wdomain-21","wdomain-22"]}]}`,
				MacroVideoMIMES:                `mp4,flv`,
				MacroVideoMinimumDuration:      `30`,
				MacroVideoMaximumDuration:      `60`,
				MacroVideoProtocols:            `3,8`,
				MacroVideoPlayerWidth:          `640`,
				MacroVideoPlayerHeight:         `480`,
				MacroVideoStartDelay:           `0`,
				MacroVideoPlacement:            `1`,
				MacroVideoLinearity:            `1`,
				MacroVideoSkip:                 `0`,
				MacroVideoSkipMinimum:          `10`,
				MacroVideoSkipAfter:            `5`,
				MacroVideoSequence:             `1`,
				MacroVideoBlockedAttribute:     `1,2`,
				MacroVideoMaximumExtended:      `10`,
				MacroVideoMinimumBitRate:       `360`,
				MacroVideoMaximumBitRate:       `1080`,
				MacroVideoBoxing:               `1`,
				MacroVideoPlaybackMethod:       `1,3`,
				MacroVideoDelivery:             `1,3`,
				MacroVideoPosition:             `0`,
				MacroVideoAPI:                  `1,2`,
				MacroSiteID:                    ``,
				MacroSiteName:                  ``,
				MacroSitePage:                  ``,
				MacroSiteReferrer:              ``,
				MacroSiteSearch:                ``,
				MacroSiteMobile:                ``,
				MacroAppID:                     `app-id`,
				MacroAppName:                   `app-name`,
				MacroAppBundle:                 `app-bundle`,
				MacroAppStoreURL:               `app-store-url`,
				MacroAppVersion:                `app-version`,
				MacroAppPaid:                   `1`,
				MacroCategory:                  `app-cat1,app-cat2`,
				MacroDomain:                    `app-domain`,
				MacroSectionCategory:           `app-sec-cat1,app-sec-cat2`,
				MacroPageCategory:              `app-page-cat1,app-page-cat2`,
				MacroPrivacyPolicy:             `2`,
				MacroKeywords:                  `app-keywords`,
				MacroPubID:                     `app-pub-id`,
				MacroPubName:                   `app-pub-name`,
				MacroPubDomain:                 `app-pub-domain`,
				MacroContentID:                 `app-cnt-id`,
				MacroContentEpisode:            `2`,
				MacroContentTitle:              `app-cnt-title`,
				MacroContentSeries:             `app-cnt-series`,
				MacroContentSeason:             `app-cnt-season`,
				MacroContentArtist:             `app-cnt-artist`,
				MacroContentGenre:              `app-cnt-genre`,
				MacroContentAlbum:              `app-cnt-album`,
				MacroContentISrc:               `app-cnt-isrc`,
				MacroContentURL:                `app-cnt-url`,
				MacroContentCategory:           `app-cnt-cat1,app-cnt-cat2`,
				MacroContentProductionQuality:  `0`,
				MacroContentVideoQuality:       `0`,
				MacroContentContext:            `1`,
				MacroContentContentRating:      `1.2`,
				MacroContentUserRating:         `2.2`,
				MacroContentQAGMediaRating:     `1`,
				MacroContentKeywords:           `app-cnt-keywords`,
				MacroContentLiveStream:         `1`,
				MacroContentSourceRelationship: `1`,
				MacroContentLength:             `100`,
				MacroContentLanguage:           `english`,
				MacroContentEmbeddable:         `1`,
				MacroProducerID:                `app-cnt-prod-id`,
				MacroProducerName:              `app-cnt-prod-name`,
				MacroUserAgent:                 `user-agent`,
				MacroDNT:                       `0`,
				MacroLMT:                       `0`,
				MacroIP:                        `ipv6`,
				MacroDeviceType:                `3`,
				MacroMake:                      `device-make`,
				MacroModel:                     `device-model`,
				MacroDeviceOS:                  `os`,
				MacroDeviceOSVersion:           `os-version`,
				MacroDeviceWidth:               `2048`,
				MacroDeviceHeight:              `1024`,
				MacroDeviceJS:                  `1`,
				MacroDeviceLanguage:            `device-lang`,
				MacroDeviceIFA:                 `ifa`,
				MacroDeviceIFAType:             `idfa`,
				MacroDeviceDIDSHA1:             `didsha1`,
				MacroDeviceDIDMD5:              `didmd5`,
				MacroDeviceDPIDSHA1:            `dpidsha1`,
				MacroDeviceDPIDMD5:             `dpidmd5`,
				MacroDeviceMACSHA1:             `macsha1`,
				MacroDeviceMACMD5:              `macmd5`,
				MacroLatitude:                  `1.1`,
				MacroLongitude:                 `2.2`,
				MacroCountry:                   `country`,
				MacroRegion:                    `region`,
				MacroCity:                      `city`,
				MacroZip:                       `zip`,
				MacroUTCOffset:                 `1000`,
				MacroUserID:                    `user-id`,
				MacroYearOfBirth:               `1990`,
				MacroGender:                    `M`,
				MacroGDPRConsent:               `user-gdpr-consent`,
				MacroGDPR:                      `1`,
				MacroUSPrivacy:                 `user-privacy`,
				MacroCacheBuster:               `cachebuster`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			macroMappings := GetDefaultMapper()

			tag := tt.args.tag
			tag.InitBidRequest(tt.args.bidRequest)
			tag.SetAdapterConfig(tt.args.conf)
			tag.LoadImpression(&tt.args.bidRequest.Imp[0])

			for key, result := range tt.macros {
				cb, ok := macroMappings[key]
				if !ok {
					assert.NotEmpty(t, result)
				} else {
					actual := cb.callback(tag, key)
					assert.Equal(t, result, actual, fmt.Sprintf("MacroFunction: %v", key))
				}
			}
		})
	}
}

func TestBidderGetValue(t *testing.T) {
	type fields struct {
		KV map[string]interface{}
	}
	type args struct {
		key string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       string
		isKeyFound bool // if key has the prefix kv/kvm then it should return thr isKeyFound true
	}{
		{
			name: "valid_Key",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"age":  22,
			}},
			args:       args{key: "kv.name"},
			want:       "test",
			isKeyFound: true,
		},
		{
			name: "invalid_Key",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"age":  22,
			}},
			args:       args{key: "kv.anykey"},
			want:       "",
			isKeyFound: true,
		},
		{
			name:       "empty_kv_map",
			fields:     fields{KV: nil},
			args:       args{key: "kv.anykey"},
			want:       "",
			isKeyFound: true,
		},
		{
			name:       "kv_map_with_no_key_val_pair",
			fields:     fields{KV: map[string]interface{}{}},
			args:       args{key: "kv.anykey"},
			want:       "",
			isKeyFound: true,
		},
		{
			name: "key_with_value_as_url",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"country": map[string]interface{}{
					"state":   "MH",
					"pincode": 411041,
					"url":     "http://example.com?k1=v1&k2=v2",
				},
			}},
			args:       args{key: "kvm.country.url"},
			want:       "http://example.com?k1=v1&k2=v2",
			isKeyFound: true,
		},
		{
			name: "kvm_prefix_key_with_value_as_nested_map",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"country": map[string]interface{}{
					"state":   "MH",
					"pincode": 411041,
					"url":     "http//example.com?k1=v1&k2=v2",
					"metadata": map[string]interface{}{
						"k1": "v1",
						"k2": "v2",
					},
				},
			}},
			args:       args{key: "kvm.country"},
			want:       "{\"metadata\":{\"k1\":\"v1\",\"k2\":\"v2\"},\"pincode\":411041,\"state\":\"MH\",\"url\":\"http//example.com?k1=v1&k2=v2\"}",
			isKeyFound: true,
		},
		{
			name: "kv_prefix_key_with_value_as_nested_map",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"country": map[string]interface{}{
					"state":   "MH",
					"pincode": 411041,
					"url":     "http://example.com?k1=v1&k2=v2",
					"metadata": map[string]interface{}{
						"k1": "v1",
						"k2": "v2",
					},
				},
			}},
			args:       args{key: "kv.country"},
			want:       "metadata=k1%3Dv1%26k2%3Dv2&pincode=411041&state=MH&url=http%3A%2F%2Fexample.com%3Fk1%3Dv1%26k2%3Dv2",
			isKeyFound: true,
		},
		{
			name: "key_without_kv_kvm_prefix",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"country": map[string]interface{}{
					"state":   "MH",
					"pincode": 411041,
					"url":     "http//example.com?k1=v1&k2=v2",
					"metadata": map[string]interface{}{
						"k1": "v1",
						"k2": "v2",
					},
				},
			}},
			args:       args{key: "someprefix.kv"},
			want:       "",
			isKeyFound: false, // hence this key is not starting with kv/kvm prefix we return isKeyFound as false
		},
		{
			name: "multi-level_key",
			fields: fields{KV: map[string]interface{}{
				"k1": map[string]interface{}{
					"k2": map[string]interface{}{
						"k3": map[string]interface{}{
							"k4": map[string]interface{}{
								"name": "test",
							},
						},
					},
				},
			}},
			args:       args{key: "kv.k1.k2.k3.k4.name"},
			want:       "test",
			isKeyFound: true,
		},
		{
			name: "key_not_matched",
			fields: fields{KV: map[string]interface{}{
				"k1": map[string]interface{}{
					"k2": map[string]interface{}{
						"k3": map[string]interface{}{
							"k4": map[string]interface{}{
								"name": "test",
							},
						},
					},
				},
			}},
			args:       args{key: "kv.k1.k2.k3.name"},
			want:       "",
			isKeyFound: true,
		},
		{
			name: "key_wihtout_any_prefix",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"age":  22,
			}},
			args:       args{key: "kv"},
			want:       "",
			isKeyFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := &BidderMacro{
				KV: tt.fields.KV,
			}
			value, isKeyFound := tag.GetValue(tt.args.key)
			assert.Equal(t, tt.want, value, tt.name)
			assert.Equal(t, tt.isKeyFound, isKeyFound)
		})
	}
}

func TestBidderMacroKV(t *testing.T) {
	type fields struct {
		KV map[string]interface{}
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "valid_test",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"age":  "22",
			}},
			args: args{key: "kv"},
			want: "age=22&name=test",
		},
		{
			name: "valid_test_with_url",
			fields: fields{KV: map[string]interface{}{
				"age": "22",
				"url": "http://example.com?k1=v1&k2=v2",
			}},
			args: args{key: "kv"},
			want: "age=22&url=http%3A%2F%2Fexample.com%3Fk1%3Dv1%26k2%3Dv2",
		},
		{
			name: "valid_test_with_encoded_url",
			fields: fields{KV: map[string]interface{}{
				"age": "22",
				"url": "http%3A%2F%2Fexample.com%3Fk1%3Dv1%26k2%3Dv2",
			}},
			args: args{key: "kv"},
			want: "age=22&url=http%3A%2F%2Fexample.com%3Fk1%3Dv1%26k2%3Dv2",
		},
		{
			name:   "empty_KV_map",
			fields: fields{KV: nil},
			args:   args{key: "kv"},
			want:   "",
		},
		{
			name:   "kv_map_with_no_key_val_pair",
			fields: fields{KV: map[string]interface{}{}},
			args:   args{key: "kv"},
			want:   "",
		},
		{
			name: "key_with_value_as_map",
			fields: fields{KV: map[string]interface{}{
				"age": 22,
				"country": map[string]interface{}{
					"state":   "MH",
					"pincode": 411041,
				},
			}},
			args: args{key: "kv"},
			want: "age=22&country=pincode%3D411041%26state%3DMH",
		},
		{
			name: "key_with_value_as_nested_map",
			fields: fields{KV: map[string]interface{}{
				"age": 22,
				"country": map[string]interface{}{
					"state":   "MH",
					"pincode": 411041,
					"metadata": map[string]interface{}{
						"k1": 223,
						"k2": "v2",
					},
				},
			}},
			args: args{key: "kv"},
			want: "age=22&country=metadata%3Dk1%253D223%2526k2%253Dv2%26pincode%3D411041%26state%3DMH",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := &BidderMacro{
				KV: tt.fields.KV,
			}
			got := tag.MacroKV(tt.args.key)

			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestBidderMacroKVM(t *testing.T) {
	type fields struct {
		KV map[string]interface{}
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "valid_test",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"age":  "22",
			}},
			args: args{key: "kvm"},
			want: "{\"age\":\"22\",\"name\":\"test\"}",
		},
		{
			name:   "empty_kv_map",
			fields: fields{KV: nil},
			args:   args{key: "kvm"},
			want:   "",
		},
		{
			name: "value_as_int_data_type",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"age":  22,
			}},
			args: args{key: "kvm"},
			want: "{\"age\":22,\"name\":\"test\"}",
		},
		{
			name:   "kv_map_with_no_key_val_pair",
			fields: fields{KV: map[string]interface{}{}},
			args:   args{key: "kvm"},
			want:   "{}",
		},
		{
			name: "marshal_error",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"age":  make(chan int),
			}},
			args: args{key: "kvm"},
			want: "",
		},
		{
			name: "test_with_url",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"url":  "http://example.com?k1=v1&k2=v2",
			}},
			args: args{key: "kvm"},
			want: "{\"name\":\"test\",\"url\":\"http://example.com?k1=v1&k2=v2\"}",
		},
		{
			name: "key_with_value_as_nested_map",
			fields: fields{KV: map[string]interface{}{
				"name": "test",
				"age":  22,
				"country": map[string]interface{}{
					"state":   "MH",
					"pincode": 411041,
					"metadata": map[string]interface{}{
						"k1": "v1",
						"k2": "v2",
					},
				},
			}},
			args: args{key: "kvm"},
			want: "{\"age\":22,\"country\":{\"metadata\":{\"k1\":\"v1\",\"k2\":\"v2\"},\"pincode\":411041,\"state\":\"MH\"},\"name\":\"test\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := &BidderMacro{
				KV: tt.fields.KV,
			}
			got := tag.MacroKVM(tt.args.key)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestMacroSchain(t *testing.T) {

	type fields struct {
		Request *openrtb2.BidRequest
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "source_object_with_both_source.schain_and_source.ext.schain",
			fields: fields{&openrtb2.BidRequest{Source: &openrtb2.Source{
				SChain: &openrtb2.SupplyChain{},
				Ext: []byte(`{
					"schain":{
						"complete":1,
						"nodes":[
							{
								"asi":"exchange1.com",
								"sid":"1234&abcd",
								"hp":1,
								"name":"publisher name"
							}
						],
						"ver":"1.0"
					}
				}`),
			}}},
			args: args{key: "schain"},
			want: "", // here we have given priority to source.schain object hence source.schain is not nil it return empty string
		},
		{
			name: "nil_source.schain_object",
			fields: fields{&openrtb2.BidRequest{
				Source: &openrtb2.Source{
					SChain: nil,
					Ext: []byte(`{
						"schain":{
							"complete":0,
							"nodes":[
								{
									"asi":"exchange2.com",
									"sid":"abcd",
									"hp":1
								}
							],
							"ver":"1.0"
						}
					}`),
				},
			}},
			args: args{key: "schain"},
			want: "1.0,0!exchange2.com,abcd,1,,,",
		},
		{
			name: "missing_schain_object",
			fields: fields{&openrtb2.BidRequest{Source: &openrtb2.Source{
				Ext: []byte(`{
					"somechain":{
						"complete":1,
						"nodes":[
							{
								"asi":"exchange1.com",
								"sid":"1234&abcd",
								"hp":1,
								"ext":{"k1":"v1"}
							}
						],
						"ver":"1.0"
					}
				}`),
			}}},
			args: args{key: "schain"},
			want: "",
		},
		{
			name:   "missing_both_source.schain_and_source.ext",
			fields: fields{&openrtb2.BidRequest{Source: nil}},
			args:   args{key: "schain"},
			want:   "",
		},
		{
			name: "source.schain_is_present",
			fields: fields{&openrtb2.BidRequest{Source: &openrtb2.Source{
				SChain: &openrtb2.SupplyChain{
					Complete: 1,
					Ver:      "1.0",
					Nodes: []openrtb2.SupplyChainNode{
						{
							ASI:    "asi",
							SID:    "sid",
							RID:    "rid",
							Name:   "name",
							Domain: "domain",
							HP:     openrtb2.Int8Ptr(1),
						},
					}},
			}}},
			args: args{key: "schain"},
			want: "1.0,1!asi,sid,1,rid,name,domain",
		},
		{
			name: "unmarshaling_error",
			fields: fields{&openrtb2.BidRequest{Source: &openrtb2.Source{
				Ext: []byte(`{
					"schain":{
						"complete":"1",
						"nodes":[
							{
								"asi":"exchange1.com",
								"sid":"1234&abcd",
								"rid":"bid-request-1",
								"name":"publisher%20name",
								"domain":"publisher.com",
								"hp":1
							}
						],
						"ver":"1.0"
					}
				}`),
			}}},
			args: args{key: "schain"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := &BidderMacro{
				Request: tt.fields.Request,
			}
			got := tag.MacroSchain(tt.args.key)
			assert.Equal(t, got, tt.want, tt.name)
		})
	}
}
