package native_video

import (
	"bytes"
	"text/template"
)

var vast = `<VAST version="3.0" xmlns:xs="http://www.w3.org/2001/XMLSchema">
<Ad id="20001">
	<InLine>
		<AdSystem version="4.0">iabtechlab</AdSystem>
		<AdTitle>iabtechlab video ad</AdTitle>
		<Pricing model="cpm" currency="USD"><![CDATA[{{.Price}}]]></Pricing>
		<Error>https://example.com/error</Error>
		<Impression id="Impression-ID">https://example.com/track/impression</Impression>
		<Creatives>
			<Creative id="{{.CreativeID}}" sequence="1">
				<Linear {{.Skip}}>
					<Duration>00:00:{{.Duration}}</Duration>
					 <VideoClicks>
						<ClickTracking id="blog">
							<![CDATA[https://iabtechlab.com]]>
						</ClickTracking>
						<CustomClick>http://iabtechlab.com</CustomClick>
					</VideoClicks>
					<MediaFiles>
						<MediaFile id="5241" delivery="progressive" type="video/mp4" bitrate="500" width="400" height="300" minBitrate="360" maxBitrate="1080" scalable="1" maintainAspectRatio="1" codec="0">
							<![CDATA[{{.MediaFileURL}}]]>
						</MediaFile>
					</MediaFiles>
				</Linear>
			</Creative>
		</Creatives>
		<Extensions>
			<Extension type="iab-Count">
				<total_available>
					<![CDATA[ 2 ]]>
				</total_available>
			</Extension>
		</Extensions>
	</InLine>
</Ad>
</VAST>
`

type content struct {
	MediaFileURL string
	Duration     string
	Price        string
	Skip         string
	CreativeID   string
}

func generateVASTXml(price, mediaPath string) string {

	t, _ := template.New("test").Parse(vast)
	c := content{}
	c.MediaFileURL = mediaPath
	c.Duration = "15"
	c.Price = "25"
	c.Skip = "" // not skippable default
	// generate random uniq id
	c.CreativeID = "111112233"
	var f = new(bytes.Buffer)
	t.Execute(f, c)
	var vast []byte
	f.Write(vast)

	return f.String()
}
