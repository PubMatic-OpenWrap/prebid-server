package pubmatic

import (
	"testing"

	"github.com/prebid/prebid-server/config"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPLogger(t *testing.T) {

	type want struct {
		MaxClients     int32
		MaxConnections int
		MaxCalls       int
	}

	tests := []struct {
		name string
		cfg  config.PubMaticWL
		want want
	}{
		{
			name: "test global values",
			cfg: config.PubMaticWL{
				MaxClients:     1,
				MaxConnections: 10,
				MaxCalls:       1,
				RespTimeout:    10,
			},
			want: want{
				MaxClients:     1,
				MaxConnections: 10,
				MaxCalls:       1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module := NewHTTPLogger(tt.cfg)
			assert.NotNil(t, module, tt.name)
			assert.Equal(t, maxHttpClients, tt.want.MaxClients, tt.name)
			assert.Equal(t, maxHttpConnections, tt.want.MaxConnections, tt.name)
			assert.Equal(t, maxHttpCalls, tt.want.MaxCalls, tt.name)
		})
	}
}
