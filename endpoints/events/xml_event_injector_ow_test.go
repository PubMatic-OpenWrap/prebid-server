package events

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func search(arr []int, value int) bool {
	idx := sort.SearchInts(arr, value)
	return idx < len(arr) && arr[idx] == value
}

func quoteUnescape[T []byte | string](s T) string {
	buf := bytes.Buffer{}
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '\\' {
			if i+1 < len(s) {
				nextCh := s[i+1]
				if nextCh == '\\' || nextCh == '"' || nextCh == '\'' {
					i++
					ch = nextCh
				}
			}
		}
		buf.WriteByte(ch)
	}
	return buf.String()
}

func TestETreeBehaviour(t *testing.T) {
	// vast1 := `<?xml version="1.0" encoding="UTF-8" standalone="no"?><VAST version="2.0"><Ad id="4650_86226f7b2a982e9cadfd8dc58d6965d0"><InLine><AdSystem version="1.0.0">Appreciate</AdSystem><Impression><![CDATA[https://ets-us-east-1.track.smaato.net/v1/view?sessionId=5251ca0e-bbf6-0e29-ae23-b7f26a5afde1&adSourceId=3b574e75-bf30-58a8-dd58-e1150fc75c7a&originalRequestTime=1728487967713&e=prebid&c=ortb26&expires=1728489767713&dpid=XXf9QjPbrtRrwIB0Nwyjfg%3D%3D%7CN7ChHoSOYimw_5CVEwXUIA%3D%3D&winurl=9YmCuFWdQfG4XONgkFmrin5Z3eiObza_044Wg3fHLUXiDz3TMFktd0VlAqWfhhkLZrF9JHy0zzYCFSJCL0jzgsAoBbaDy_rRj-RP0eCTyasl0oHJUs3BQgHWmJRaFA1hnc3WNrIf3Rsh8jCyDv5u9xl7PzMTXsBws4Mrd6vgGRytdqp-BuzQvJbWVEaRGkus9UMusCAjjlg1LDEIYthN8FQnCXX_3gT5DWlnnAXC2S6FBdJymSbfrkgsVkY_-g_PPx1ceGLgX6q6WxnL7Oof3pJ56SIjTszQ9xEOIk1mRmrNVgVumfu1LsFWgv0SRFMXyGKlYbTHjv_7cEcDmrjgky__uRyyqc5-ZUsF_9S1BgFfHMq9vcy7KQXAmRac0mRR8Psrnd3346wT15YyBSwkyg%3D%3D%7CixC2LMzFYaCtkl4MdiJPAA%3D%3D]]></Impression><AdTitle><![CDATA[ ]]></AdTitle><Description><![CDATA[ ]]></Description><Error><![CDATA[https://gotu.tpbid.com/vast/2?bid_id=4124e8b2e5c34d0b02021b8e50dca0d05bfaec52df969b6d6706a21f&cid=4650_4154bbf6a600a80b63c9171e94701ad2&crid=4650_86226f7b2a982e9cadfd8dc58d6965d0&lid=27f313dcd213beb73fb51378aeff34b7&g=l&a=error&code=[ERRORCODE]]]></Error><Impression><![CDATA[https://gotu.tpbid.com/vast/2?bid_id=4124e8b2e5c34d0b02021b8e50dca0d05bfaec52df969b6d6706a21f&cid=4650_4154bbf6a600a80b63c9171e94701ad2&crid=4650_86226f7b2a982e9cadfd8dc58d6965d0&lid=27f313dcd213beb73fb51378aeff34b7&g=l&t=0]]></Impression><Impression><![CDATA[https://gotu.tpbid.com/tsi?bid_id=4124e8b2e5c34d0b02021b8e50dca0d05bfaec52df969b6d6706a21f&cid=4650_4154bbf6a600a80b63c9171e94701ad2&pt=4rUVxgZ4Xk13tX1v5zDrNhCRAsZlHo4MmockDHvuO4p425Ov4Y_BchAgD-4ZBKZzv2t9LGDBgZm2_ytNH1AglytvY0bPqId8nwsksCRX6vqb1-GqVwUkk3ZIPcUtx8INSl..]]></Impression><Creatives><Creative id="1" sequence="1"><Linear><Duration>00:00:30</Duration><TrackingEvents><Tracking event="firstQuartile"><![CDATA[https://gotu.tpbid.com/vast/2?bid_id=4124e8b2e5c34d0b02021b8e50dca0d05bfaec52df969b6d6706a21f&cid=4650_4154bbf6a600a80b63c9171e94701ad2&crid=4650_86226f7b2a982e9cadfd8dc58d6965d0&lid=27f313dcd213beb73fb51378aeff34b7&g=l&t=25]]></Tracking><Tracking event="midpoint"><![CDATA[https://gotu.tpbid.com/vast/2?bid_id=4124e8b2e5c34d0b02021b8e50dca0d05bfaec52df969b6d6706a21f&cid=4650_4154bbf6a600a80b63c9171e94701ad2&crid=4650_86226f7b2a982e9cadfd8dc58d6965d0&lid=27f313dcd213beb73fb51378aeff34b7&g=l&t=50]]></Tracking><Tracking event="thirdQuartile"><![CDATA[https://gotu.tpbid.com/vast/2?bid_id=4124e8b2e5c34d0b02021b8e50dca0d05bfaec52df969b6d6706a21f&cid=4650_4154bbf6a600a80b63c9171e94701ad2&crid=4650_86226f7b2a982e9cadfd8dc58d6965d0&lid=27f313dcd213beb73fb51378aeff34b7&g=l&t=75]]></Tracking><Tracking event="complete"><![CDATA[https://gotu.tpbid.com/vast/2?bid_id=4124e8b2e5c34d0b02021b8e50dca0d05bfaec52df969b6d6706a21f&cid=4650_4154bbf6a600a80b63c9171e94701ad2&crid=4650_86226f7b2a982e9cadfd8dc58d6965d0&lid=27f313dcd213beb73fb51378aeff34b7&g=l&t=100]]></Tracking><Tracking event="start"><![CDATA[https://vet-us-east-1.track.smaato.net/start?sessionId=5251ca0e-bbf6-0e29-ae23-b7f26a5afde1&adSourceId=3b574e75-bf30-58a8-dd58-e1150fc75c7a&originalRequestTime=1728487967713&e=prebid]]></Tracking><Tracking event="firstQuartile"><![CDATA[https://vet-us-east-1.track.smaato.net/firstQuartile?sessionId=5251ca0e-bbf6-0e29-ae23-b7f26a5afde1&adSourceId=3b574e75-bf30-58a8-dd58-e1150fc75c7a&originalRequestTime=1728487967713&e=prebid]]></Tracking><Tracking event="midpoint"><![CDATA[https://vet-us-east-1.track.smaato.net/midpoint?sessionId=5251ca0e-bbf6-0e29-ae23-b7f26a5afde1&adSourceId=3b574e75-bf30-58a8-dd58-e1150fc75c7a&originalRequestTime=1728487967713&e=prebid]]></Tracking><Tracking event="thirdQuartile"><![CDATA[https://vet-us-east-1.track.smaato.net/thirdQuartile?sessionId=5251ca0e-bbf6-0e29-ae23-b7f26a5afde1&adSourceId=3b574e75-bf30-58a8-dd58-e1150fc75c7a&originalRequestTime=1728487967713&e=prebid]]></Tracking><Tracking event="complete"><![CDATA[https://vet-us-east-1.track.smaato.net/complete?sessionId=5251ca0e-bbf6-0e29-ae23-b7f26a5afde1&adSourceId=3b574e75-bf30-58a8-dd58-e1150fc75c7a&originalRequestTime=1728487967713&e=prebid]]></Tracking></TrackingEvents><VideoClicks><ClickThrough><![CDATA[https://gotu.tpbid.com/click?bid_id=4124e8b2e5c34d0b02021b8e50dca0d05bfaec52df969b6d6706a21f&ep0=eTw&ep1=uxWDmWxcmM&cf=1&ifap=1&esb=Jwd.2hwhSwMkC&cid=4650_4154bbf6a600a80b63c9171e94701ad2&crid=4650_86226f7b2a982e9cadfd8dc58d6965d0&lid=27f313dcd213beb73fb51378aeff34b7&xid=2&whoc=1&rt=4rUmY7Z4xh13tX1vTSIVDTo3As8Ft9F3h0ZZ9aWafHV-QPaaDpGhZK5YPiFpYamFv4tUnN6A5gOT3d-7zextn8P_-kd6C_daF2I9QjbeV3zhHl4Lt5A9R0H4tDEKzJB78WlR3W7QkqcEFAcEPNyqhl..&dmv=1]]></ClickThrough><ClickTracking><![CDATA[https://ets-us-east-1.track.smaato.net/v1/click?sessionId=5251ca0e-bbf6-0e29-ae23-b7f26a5afde1&adSourceId=3b574e75-bf30-58a8-dd58-e1150fc75c7a&originalRequestTime=1728487967713&e=prebid]]></ClickTracking></VideoClicks><MediaFiles><MediaFile bitrate="353" delivery="progressive" height="360" type="video/mp4" width="640"><![CDATA[https://c.tpbid.com/ssb/4650/videos/41e1c3e2b8873a195cbbcc524319e6bc.mp4]]></MediaFile></MediaFiles></Linear></Creative><Creative id="2" sequence="1"><CompanionAds><Companion height="320" width="480"><StaticResource creativeType="image/png"><![CDATA[https://c.tpbid.com/ssb/4650/images/e70b73e9da1c2c3dc6fdbafe9f96c494.jpg]]></StaticResource><CompanionClickThrough><![CDATA[https://gotu.tpbid.com/click?bid_id=4124e8b2e5c34d0b02021b8e50dca0d05bfaec52df969b6d6706a21f&ep0=eTw&ep1=uxWDmWxcmM&cf=1&ifap=1&esb=Jwd.2hwhSwMkC&cid=4650_4154bbf6a600a80b63c9171e94701ad2&crid=4650_86226f7b2a982e9cadfd8dc58d6965d0&lid=27f313dcd213beb73fb51378aeff34b7&xid=2&whoc=1&rt=4rUmY7Z4xh13tX1vTSIVDTo3As8Ft9F3h0ZZ9aWafHV-QPaaDpGhZK5YPiFpYamFv4tUnN6A5gOT3d-7zextn8P_-kd6C_daF2I9QjbeV3zhHl4Lt5A9R0H4tDEKzJB78WlR3W7QkqcEFAcEPNyqhl..&dmv=1]]></CompanionClickThrough><TrackingEvents><Tracking event="creativeView"><![CDATA[https://vet-us-east-1.track.smaato.net/companion/creativeView?sessionId=5251ca0e-bbf6-0e29-ae23-b7f26a5afde1&adSourceId=3b574e75-bf30-58a8-dd58-e1150fc75c7a&originalRequestTime=1728487967713&e=prebid]]></Tracking></TrackingEvents></Companion></CompanionAds></Creative></Creatives><Extensions><Extension type="Pricing"><Price currency="USD" model="CPM" source="smaato"><![CDATA[0.14087]]></Price></Extension></Extensions></InLine></Ad></VAST>`
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "test",
			in:   "<AdTitle>&#xA;        [ini:PDC][fmt:Video][crs:3682][csz:15s]&#xA;      </AdTitle>",
			out:  "<AdTitle><![CDATA[[ini:PDC][fmt:Video][crs:3682][csz:15s]]]></AdTitle>",
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := etree.NewDocument()
			doc.WriteSettings.CanonicalEndTags = true

			err := doc.ReadFromString(tt.in)
			assert.Nil(t, err)

			out, err := doc.WriteToString()
			assert.Nil(t, err)
			assert.Equal(t, tt.out, out)
		})
	}
}

func TestCompareXMLParsers(t *testing.T) {
	//$ cat *-prod.txt | sed -n 's/.*creative:\[\(.*\)\].*/\1/p' > $GOPATH/src/github.com/PubMatic-OpenWrap/prebid-server/endpoints/events/test/base64_vast.txt
	type stats struct {
		valid           []int
		generalMismatch []int
		singleQuote     []int
	}

	var (
		//fileName              = `./test/base64_vast.txt`
		//fileName              = `./test/base64_quoted_vast.txt`
		fileName              = `./test/raw_vast.txt`
		quoted                = strings.Contains(fileName, "quoted") //xml files retrived from prod vast unwrapper
		base64Decode          = strings.Contains(fileName, "base64")
		debugLines            = []int{}
		st                    = stats{}
		currentLine, xmlCount = 0, 0
	)

	file, err := os.Open(fileName)
	if !assert.Nil(t, err) {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	sort.Ints(debugLines)

	for scanner.Scan() {
		currentLine++
		vast := scanner.Text()

		//presetup
		{
			//debug
			if len(debugLines) > 0 {
				if found := search(debugLines, currentLine); !found {
					continue
				}
			}

			//base64decode
			if base64Decode {
				data, err := base64.StdEncoding.DecodeString(vast)
				if !assert.Nil(t, err) {
					continue
				}
				vast = string(data)
				if quoted {
					vast = quoteUnescape(data)
				}
			}
		}

		t.Run(fmt.Sprintf("vast_%d", currentLine), func(t *testing.T) {
			xmlCount++

			etreeXML, _ := etreeEJ.Inject(vast, eventURLMap)
			fastXML, _ := fastxmlEJ.Inject(vast, eventURLMap)

			if vast != fastXML {
				fastXML, etreeXML = openrtb_ext.NormalizeXML(fastXML), openrtb_ext.NormalizeXML(etreeXML)
			}

			if len(debugLines) > 0 {
				assert.Equal(t, etreeXML, fastXML, vast)
			}

			if etreeXML != fastXML {
				if idx := strings.Index(etreeXML, "&apos;"); idx != -1 &&
					(strings.HasPrefix(fastXML[idx:], "&#39;") || strings.HasPrefix(fastXML[idx:], "\"")) {
					st.singleQuote = append(st.singleQuote, currentLine)
				} else {
					st.generalMismatch = append(st.generalMismatch, currentLine)
				}
				return
			}
			st.valid = append(st.valid, currentLine)
		})
	}

	t.Logf("\nTotal:[%v] validCount:[%v] generalMismatch:[%v] singleQuote:[%v]", xmlCount, st.valid, st.generalMismatch, st.singleQuote)
	assert.NotZero(t, xmlCount)
	assert.Equal(t, xmlCount, len(st.valid), "validXMLCount")
	assert.Equal(t, 0, len(st.generalMismatch), "generalMismatch")
	assert.Equal(t, 0, len(st.singleQuote), "singleQuote")
	assert.Nil(t, scanner.Err())
}
