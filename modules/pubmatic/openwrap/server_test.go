package openwrap

import (
	"os"
	"testing"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/wakanda"
	"github.com/stretchr/testify/assert"
)

func TestInitOpenWrapServer(t *testing.T) {
	type args struct {
		cfg *config.Config
	}
	tests := []struct {
		name  string
		args  args
		want  wakanda.Wakanda
		setup func()
	}{
		{
			name: "check config",
			args: args{
				cfg: &config.Config{
					Server: config.Server{
						HostName: "localhost",
						DCName:   "abcd",
						Endpoint: "http://localhost:18012",
					},
				},
			},
			want: wakanda.Wakanda{
				HostName: "localhost",
				DCName:   "abcd",
				PodName:  "Default_Pod",
			},
			setup: func() {
				os.Setenv("MY_POD_NAME", "Default_Pod")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got := initOpenWrapServer(tt.args.cfg)
			assert.Equal(t, tt.args.cfg.Wakanda, tt.want)
			assert.NotNil(t, got)
		})
	}
}
