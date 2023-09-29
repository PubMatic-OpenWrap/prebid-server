package xmlparser

import (
	"testing"
)

var (
	vastXMLString                        string
	xmlTokenizer                         *XMLTokenizer
	xmlTokenizerWithXPath                *XMLTokenizer
	xmlReader                            *XMLReader
	xmlReaderWithXPath                   *XMLReader
	mockTokenHandler1, mockTokenHandler2 mockTokenHandler
)

func init() {
	xpath := GetXPath([][]string{
		{"VAST", "Ad", "InLine", "Impression"},
		{"VAST", "Ad", "InLine", "Error"},
		{"VAST", "Ad", "InLine", "Creatives", "Creative", "NonLinearAds", "TrackingEvents", "Tracking"},
		{"VAST", "Ad", "Wrapper", "Impression"},
		{"VAST", "Ad", "Wrapper", "Error"},
		{"VAST", "Ad", "Wrapper", "Creatives", "Creative", "Linear", "TrackingEvents", "Tracking"},
		{"VAST", "Ad", "Wrapper", "Creatives", "Creative", "Linear", "VideoClicks"},
	})

	vastXMLString = `<VAST  version="3.0" xmlns:xs="http://www.w3.org/2001/XMLSchema">
    <Ad  id="20001">
        <Wrapper>
            <Error>http://example.com/error</Error>
            <Impression  id="Impression-ID">http://example.com/track/impression</Impression>
            <Creatives>
                <Creative  id="5480" sequence="1">
                    <Linear>
                        <Duration>00:00:16</Duration>
                        <TrackingEvents>
                            <Tracking  event="start">http://example.com/tracking/start</Tracking>
                            <Tracking  event="firstQuartile">http://example.com/tracking/firstQuartile</Tracking>
                            <Tracking  event="midpoint">http://example.com/tracking/midpoint</Tracking>
                            <Tracking  event="thirdQuartile">http://example.com/tracking/thirdQuartile</Tracking>
                            <Tracking  event="complete">http://example.com/tracking/complete</Tracking>
                            <Tracking  event="progress" offset="00:00:10">http://example.com/tracking/progress-10</Tracking>
                        </TrackingEvents>
                        <VideoClicks>
                            <ClickThrough  id="blog">
                                <![CDATA[https://iabtechlab.com]]>
                            </ClickThrough>
                        </VideoClicks>
                        <MediaFiles>
                            <MediaFile  id="5241" delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" minBitrate="360" maxBitrate="1080" scalable="1" maintainAspectRatio="1" codec="0" apiFramework="VAST">
                                <![CDATA[https://iab-publicfiles.s3.amazonaws.com/vast/VAST-4.0-Short-Intro.mp4]]>
                            </MediaFile>
                        </MediaFiles>
                    </Linear>
                </Creative>
            </Creatives>
        </Wrapper>
    </Ad>
</VAST>`
	xmlTokenizer = NewXMLTokenizer(nil)
	xmlTokenizerWithXPath = NewXMLTokenizer(xpath)
	xmlReader = NewXMLReader(nil)
	xmlReaderWithXPath = NewXMLReader(xpath)
}

func BenchmarkXMLTokenizer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mockTokenHandler1.tokens = mockTokenHandler1.tokens[:0]
		xmlTokenizer.Parse([]byte(vastXMLString), mockTokenHandler1.append)
	}
}

func BenchmarkXMLReader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		xmlReader.Parse([]byte(vastXMLString))
	}
}

func BenchmarkXMLTokenizerWithXPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mockTokenHandler2.tokens = mockTokenHandler2.tokens[:0]
		xmlTokenizerWithXPath.Parse([]byte(vastXMLString), mockTokenHandler2.append)
	}
}

func BenchmarkXMLReaderWithXPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		xmlReaderWithXPath.Parse([]byte(vastXMLString))
	}
}

/*
Running tool: /usr/local/src/go/bin/go test -benchmem -run=^$ -coverprofile=/var/folders/12/tnwntpbn5h3gjx_5gb30ngzm0000gn/T/vscode-go6QV7PF/go-code-cover -bench . vastevents/xmlparser

goos: darwin
goarch: amd64
pkg: vastevents/xmlparser
cpu: Intel(R) Core(TM) i5-8279U CPU @ 2.40GHz
BenchmarkXMLTokenizer-8            	  316545	      3542 ns/op	    3584 B/op	      25 allocs/op
BenchmarkXMLReader-8               	  307064	      4082 ns/op	    3584 B/op	      25 allocs/op
BenchmarkXMLTokenizerWithXPath-8   	  254358	      4467 ns/op	    3640 B/op	      25 allocs/op
BenchmarkXMLReaderWithXPath-8      	  267708	      4533 ns/op	    3640 B/op	      25 allocs/op
PASS
coverage: 48.6% of statements
ok  	vastevents/xmlparser	5.204s

*/
