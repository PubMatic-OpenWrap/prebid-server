package openrtb2

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/PubMatic-OpenWrap/etree"
	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func TestGetAdDuration(t *testing.T) {
	var tests = []struct {
		scenario      string
		adDuration    string // actual ad duration. 0 value will be assumed as no ad duration
		maxAdDuration int    // requested max ad duration
		expect        int
	}{
		{"0sec ad duration", "0", 200, 200},
		{"30sec ad duration", "30", 100, 30},
		{"negative ad duration", "-30", 100, 100},
		{"invalid ad duration", "invalid", 80, 80},
		{"ad duration breaking bid.Ext json", `""quote""`, 50, 50},
	}
	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			bid := openrtb.Bid{
				Ext: []byte(`{"prebid" : {"video" : {"duration" : ` + test.adDuration + `}}}`),
			}
			assert.Equal(t, test.expect, getAdDuration(bid, int64(test.maxAdDuration)))
		})
	}
}

func TestAddTargetingKeys(t *testing.T) {
	var tests = []struct {
		scenario string // Testcase scenario
		key      string
		value    string
		bidExt   string
		expect   map[string]string
	}{
		{scenario: "key_not_exists", key: "hb_pb_cat_dur", value: "some_value", bidExt: `{"prebid":{"targeting":{}}}`, expect: map[string]string{"hb_pb_cat_dur": "some_value"}},
		{scenario: "key_already_exists", key: "hb_pb_cat_dur", value: "new_value", bidExt: `{"prebid":{"targeting":{"hb_pb_cat_dur":"old_value"}}}`, expect: map[string]string{"hb_pb_cat_dur": "new_value"}},
	}
	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			bid := new(openrtb.Bid)
			bid.Ext = []byte(test.bidExt)
			key := openrtb_ext.TargetingKey(test.key)
			assert.Nil(t, addTargetingKey(bid, key, test.value))
			extBid := openrtb_ext.ExtBid{}
			json.Unmarshal(bid.Ext, &extBid)
			assert.Equal(t, test.expect, extBid.Prebid.Targeting)
		})
	}
	assert.Equal(t, "Invalid bid", addTargetingKey(nil, openrtb_ext.HbCategoryDurationKey, "some value").Error())
}

func TestAdjustBidIDInVideoEventTrackers(t *testing.T) {
	type args struct {
		modifiedBid *openrtb.Bid
	}
	type want struct {
		eventURLMap map[string]string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "replace_with_custom_ctv_bid_id",
			want: want{
				eventURLMap: map[string]string{
					"thirdQuartile": "https://thirdQuartile.com?key1=value1&bidid=ctv_bid_123",
					"complete":      "https://complete.com?key1=value1&bidid=ctv_bid_123&key2=value2",
					"firstQuartile": "https://firstQuartile.com?key1=value1&bidid=ctv_bid_123&key2=value2",
					"midpoint":      "https://midpoint.com?key1=value1&bidid=ctv_bid_123&key2=value2",
				},
			},
			args: args{
				modifiedBid: &openrtb.Bid{
					ID: "ctv_bid_123",
					AdM: `<VAST  version="3.0">
					<Ad>
						<Wrapper>
							<AdSystem>
								<![CDATA[prebid.org wrapper]]>
							</AdSystem>
							<VASTAdTagURI>
								<![CDATA[https://search.spotxchange.com/vast/2.00/85394?VPI=MP4]]>
							</VASTAdTagURI>
							<Impression>
								<![CDATA[https://imptracker.url]]>
							</Impression>
							<Impression/>
							<Creatives>
								<Creative>
									<Linear>
										<TrackingEvents>
											<Tracking  event="thirdQuartile"><![CDATA[https://thirdQuartile.com?key1=value1&bidid=ctv_bid_123]]></Tracking>
											<Tracking  event="complete"><![CDATA[https://complete.com?key1=value1&bidid=ctv_bid_123&key2=value2]]></Tracking>
											<Tracking  event="firstQuartile"><![CDATA[https://firstQuartile.com?key1=value1&bidid=ctv_bid_123&key2=value2]]></Tracking>
											<Tracking  event="midpoint"><![CDATA[https://midpoint.com?key1=value1&bidid=ctv_bid_123&key2=value2]]></Tracking>
										</TrackingEvents>
									</Linear>
								</Creative>
							</Creatives>
							<Error>
								<![CDATA[https://error.com]]>
							</Error>
						</Wrapper>
					</Ad>
				</VAST>`,
				},
			},
		},
	}
	for _, test := range tests {
		doc := etree.NewDocument()
		doc.ReadFromString(test.args.modifiedBid.AdM)
		adjustBidIDInVideoEventTrackers(doc, test.args.modifiedBid)
		events := doc.FindElements("VAST/Ad/Wrapper/Creatives/Creative/Linear/TrackingEvents/Tracking")
		for _, event := range events {
			evntName := event.SelectAttr("event").Value
			expectedURL, _ := url.Parse(test.want.eventURLMap[evntName])
			expectedValues := expectedURL.Query()
			actualURL, _ := url.Parse(event.Text())
			actualValues := actualURL.Query()
			for k, ev := range expectedValues {
				av := actualValues[k]
				for i := 0; i < len(ev); i++ {
					assert.Equal(t, ev[i], av[i], fmt.Sprintf("Expected '%v' for '%v'. but found %v", ev[i], k, av[i]))
				}
			}
		}
	}
}
