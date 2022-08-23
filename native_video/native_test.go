package native_video

import "testing"

func TestGetVideoFilePathFromVAST(t *testing.T) {
	type args struct {
		vastBody string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				vastBody: `&lt;VAST version=&#34;3.0&#34; xmlns:xs=&#34;http://www.w3.org/2001/XMLSchema&#34;&gt;&lt;Ad id=&#34;20001&#34;&gt;&lt;InLine&gt;&lt;AdSystem version=&#34;4.0&#34;&gt;iabtechlab&lt;/AdSystem&gt;&lt;AdTitle&gt;iabtechlab video ad&lt;/AdTitle&gt;&lt;Pricing model=&#34;cpm&#34; currency=&#34;USD&#34;&gt;&lt;![CDATA[25]]&gt;&lt;/Pricing&gt;&lt;Error&gt;https://example.com/error&lt;/Error&gt;&lt;Impression id=&#34;Impression-ID&#34;&gt;https://example.com/track/impression&lt;/Impression&gt;&lt;Creatives&gt;&lt;Creative id=&#34;b292addd-0293-4ff9-ab09-5f100a6b6183&#34; sequence=&#34;1&#34;&gt;&lt;Linear &gt;&lt;Duration&gt;00:00:15&lt;/Duration&gt; &lt;VideoClicks&gt;&lt;ClickTracking id=&#34;blog&#34;&gt;&lt;![CDATA[https://iabtechlab.com]]&gt;&lt;/ClickTracking&gt;&lt;CustomClick&gt;http://iabtechlab.com&lt;/CustomClick&gt;&lt;/VideoClicks&gt;&lt;MediaFiles&gt;&lt;MediaFile id=&#34;5241&#34; delivery=&#34;progressive&#34; type=&#34;video/mp4&#34; bitrate=&#34;500&#34; width=&#34;400&#34; height=&#34;300&#34; minBitrate=&#34;360&#34; maxBitrate=&#34;1080&#34; scalable=&#34;1&#34; maintainAspectRatio=&#34;1&#34; codec=&#34;0&#34;&gt;&lt;![CDATA[http://localhost:8080/files?name=india_background.mp4]]&gt;&lt;/MediaFile&gt;&lt;/MediaFiles&gt;&lt;/Linear&gt;&lt;/Creative&gt;&lt;/Creatives&gt;&lt;Extensions&gt;&lt;Extension type=&#34;iab-Count&#34;&gt;&lt;total_available&gt;&lt;![CDATA[ 2 ]]&gt;&lt;/total_available&gt;&lt;/Extension&gt;&lt;/Extensions&gt;&lt;/InLine&gt;&lt;/Ad&gt;&lt;/VAST&gt;`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetVideoFilePathFromVAST(tt.args.vastBody)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVideoFilePathFromVAST() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetVideoFilePathFromVAST() = %v, want %v", got, tt.want)
			}
		})
	}
}
