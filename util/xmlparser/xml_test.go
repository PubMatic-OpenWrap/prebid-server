package xmlparser

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	xml = `<VAST  version="3.0" xmlns:xs="http://www.w3.org/2001/XMLSchema">
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
	minixml = `
<a>
    <b>b-data</b>
    <c>c-data</c>
    <d>
        <e>e-data</e>
		<c>c-data</c>
    </d>
	<f>
		<g>g-data</g>
	</f>
</a>`
)

var xpaths = map[string]*xpath{
	"vast": GetXPath([][]string{
		{"VAST", "Ad", "InLine", "Impression"},
		{"VAST", "Ad", "InLine", "Error"},
		{"VAST", "Ad", "InLine", "Creatives", "Creative", "NonLinearAds", "TrackingEvents", "Tracking"},
		{"VAST", "Ad", "Wrapper", "Impression"},
		{"VAST", "Ad", "Wrapper", "Error"},
		{"VAST", "Ad", "Wrapper", "Creatives", "Creative", "Linear", "TrackingEvents", "Tracking"},
		{"VAST", "Ad", "Wrapper", "Creatives", "Creative", "Linear", "VideoClicks"},
	}),
	"mini": GetXPath([][]string{
		{"a", "b"},
		{"a", "d", "e"},
		{"a", "f"},
	}),
}

type mockTokenHandler struct {
	tokens []XMLToken
}

func (r *mockTokenHandler) append(_ string, parent *Element, child Element) {
	r.tokens = append(r.tokens, child.data)
}

func printTokens(in []byte, tokens []XMLToken) string {
	out := bytes.Buffer{}
	for i, token := range tokens {
		out.WriteString(fmt.Sprintf("%d:%s:end(%d:%d)\n", i, token.Name(in[:]), token.end.si, token.end.ei))
	}
	return out.String()
}

func getXML(in []byte, nodes []XMLToken) string {
	buf := bytes.Buffer{}
	start := 0
	for _, token := range nodes {
		buf.Write(in[start:token.end.ei])
		start = token.end.ei
	}
	return buf.String()
}

func TestXMLTokenizer(t *testing.T) {
	parser := XMLTokenizer{}
	in := []byte(minixml)
	tokenHandler := mockTokenHandler{}

	//parsing
	parser.Parse(in[:], tokenHandler.append)

	actual := getXML(in[:], tokenHandler.tokens[:])
	t.Logf("Raw Tags: \n%v\n", printTokens(in[:], tokenHandler.tokens[:]))
	t.Logf("XML: %v\n", getXML(in[:], tokenHandler.tokens[:]))
	assert.Equal(t, string(in), actual)
}

func TestXPathXMLTokenizer(t *testing.T) {
	parser := XMLTokenizer{
		path: xpaths["mini"],
	}
	in := []byte(minixml)
	tokenHandler := mockTokenHandler{}

	//parsing
	parser.Parse(in[:], tokenHandler.append)

	actual := getXML(in[:], tokenHandler.tokens[:])
	t.Logf("Raw Tags: \n%v\n", printTokens(in[:], tokenHandler.tokens[:]))
	t.Logf("XML: %v\n", getXML(in[:], tokenHandler.tokens[:]))
	assert.Equal(t, string(in), actual)
}

func TestXMLReader(t *testing.T) {
	in := []byte(minixml)

	xmlReader := NewXMLReader(nil)
	xmlReader.Parse(in[:])

	actual := xmlReader.getXML(in[:])
	t.Logf("Raw Tags: \n%v\n", xmlReader.tree.printRaw(func(t XMLToken) string {
		return fmt.Sprintf("%s:end(%d:%d)", t.Name(in[:]), t.end.si, t.end.ei)
	}))
	t.Logf("XML: %v\n", actual)
	assert.Equal(t, string(in), actual)
}

/*
type rawToken struct {
	index int //XMLTokenIndex
	data  []byte
}

func splitTag(in []byte, tokens []XMLToken, index int, raw rawToken) {
	for i, token := range tokens {
		fmt.Printf("%d:%s:start(%d,%d):end(%d,%d)\n", i, token.Name(in[:]), token.start.si, token.start.ei, token.end.si, token.end.ei)
	}
	var buf bytes.Buffer

	var offset int = 0
	for i, token := range tokens {
		if i != raw.index {
			fmt.Printf("%d: %s\n", i, in[offset:token.end.ei])
			buf.Write(in[offset:token.end.ei])
		} else {
			fmt.Printf("%d.1: %s\n", i, in[offset:token.start.si])
			buf.Write(in[offset:token.start.si])

			fmt.Printf("%d.2: %s\n", i, raw.data)
			buf.Write(raw.data[:])
		}
		offset = token.end.ei
	}
	fmt.Printf("\n\n%s", buf.String())
}

func TestSplitTag(t *testing.T) {
	parser := XMLTokenizer{}
	in := []byte(minixml)
	tokens := []XMLToken{}
	parser.Parse(in[:], func(_ string, parent *tnode[XMLToken], child tnode[XMLToken]) {
		tokens = append(tokens, child.Data())
	})

	actual := parser.GetXML(in[:], tokens[:])
	//t.Logf("Raw Tags: \n%v\n", printTokens(in[:], tokens[:]))
	//t.Logf("XML: %v\n", parser.GetXML(in[:], tokens[:]))
	assert.Equal(t, string(in), actual)
	splitTag(in, tokens, 2, rawToken{index: 2, data: []byte(`<raw>rawdata</raw>`)})
}
*/
