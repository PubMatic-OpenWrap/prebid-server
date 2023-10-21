package pubmatic

import (
	"net/url"
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestGetSizeForPlatform(t *testing.T) {
	type args struct {
		width, height int64
		platform      string
	}
	tests := []struct {
		name string
		args args
		size string
	}{
		{
			name: "in-app platform",
			args: args{
				width:    100,
				height:   10,
				platform: models.PLATFORM_APP,
			},
			size: "100x10",
		},
		{
			name: "video platform",
			args: args{
				width:    100,
				height:   10,
				platform: models.PLATFORM_VIDEO,
			},
			size: "100x10v",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := getSizeForPlatform(tt.args.width, tt.args.height, tt.args.platform)
			assert.Equal(t, tt.size, size, tt.name)
		})
	}
}

func TestPrepareLoggerURL(t *testing.T) {
	type args struct {
		wlog        *WloggerRecord
		loggerURL   string
		gdprEnabled int
	}
	tests := []struct {
		name     string
		args     args
		owlogger string
	}{
		{
			name: "gdprEnabled=1",
			args: args{
				wlog: &WloggerRecord{
					record: record{
						PubID:     10,
						ProfileID: "1",
						VersionID: "0",
					},
				},
				loggerURL:   "http://t.pubmatic.com/wl",
				gdprEnabled: 1,
			},
			owlogger: `http://t.pubmatic.com/wl?gdEn=1&json={"pubid":10,"pid":"1","pdvid":"0","dvc":{},"ft":0}&pubid=10`,
		},
		{
			name: "gdprEnabled=0",
			args: args{
				wlog: &WloggerRecord{
					record: record{
						PubID:     10,
						ProfileID: "1",
						VersionID: "0",
					},
				},
				loggerURL:   "http://t.pubmatic.com/wl",
				gdprEnabled: 0,
			},
			owlogger: `http://t.pubmatic.com/wl?json={"pubid":10,"pid":"1","pdvid":"0","dvc":{},"ft":0}&pubid=10`,
		},
		{
			name: "private endpoint",
			args: args{
				wlog: &WloggerRecord{
					record: record{
						PubID:     5,
						ProfileID: "5",
						VersionID: "1",
					},
				},
				loggerURL:   "http://10.172.141.11/wl",
				gdprEnabled: 0,
			},
			owlogger: `http://10.172.141.11/wl?json={"pubid":5,"pid":"5","pdvid":"1","dvc":{},"ft":0}&pubid=5`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owlogger := PrepareLoggerURL(tt.args.wlog, tt.args.loggerURL, tt.args.gdprEnabled)
			// assert.Equal(t, url.QueryEscape(owlogger), tt.owlogger, tt.name)
			// assert.Equal(t, owlogger, url.QueryEscape(tt.owlogger), tt.name)
			decodedOwlogger, _ := url.QueryUnescape(owlogger)
			assert.Equal(t, decodedOwlogger, tt.owlogger, tt.name)
		})
	}
}

// TODO - remove this
func TestGenerateSlotName(t *testing.T) {
	type args struct {
		h     int64
		w     int64
		kgp   string
		tagid string
		div   string
		src   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "_AU_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_AU_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "/15671365/Test_Adunit",
		},
		{
			name: "_DIV_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_DIV_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "Div1",
		},
		{
			name: "_AU_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_AU_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "/15671365/Test_Adunit",
		},
		{
			name: "_AU_@_W_x_H_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_AU_@_W_x_H_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "/15671365/Test_Adunit@200x100",
		},
		{
			name: "_DIV_@_W_x_H_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_DIV_@_W_x_H_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "Div1@200x100",
		},
		{
			name: "_W_x_H_@_W_x_H_",
			args: args{
				h:     100,
				w:     200,
				kgp:   "_W_x_H_@_W_x_H_",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "200x100@200x100",
		},
		// {
		// 	name: "_AU_@_DIV_@_W_x_H_",
		// 	args: args{
		// 		h:     100,
		// 		w:     200,
		// 		kgp:   "_AU_@_DIV_@_W_x_H_",
		// 		tagid: "/15671365/Test_Adunit",
		// 		div:   "Div1",
		// 		src:   "test.com",
		// 	},
		// 	want: "/15671365/Test_Adunit@Div1@200x100",
		// },
		// {
		// 	name: "_AU_@_SRC_@_VASTTAG_",
		// 	args: args{
		// 		h:     100,
		// 		w:     200,
		// 		kgp:   "_AU_@_SRC_@_VASTTAG_",
		// 		tagid: "/15671365/Test_Adunit",
		// 		div:   "Div1",
		// 		src:   "test.com",
		// 	},
		// 	want: "/15671365/Test_Adunit@test.com@_VASTTAG_",
		// },
		{
			name: "empty_kgp",
			args: args{
				h:     100,
				w:     200,
				kgp:   "",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "",
		},
		{
			name: "random_kgp",
			args: args{
				h:     100,
				w:     200,
				kgp:   "fjkdfhk",
				tagid: "/15671365/Test_Adunit",
				div:   "Div1",
				src:   "test.com",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlotName(tt.args.h, tt.args.w, tt.args.kgp, tt.args.tagid, tt.args.div, tt.args.src)
			assert.Equal(t, tt.want, got)
		})
	}
}
