package openwrap

import (
	"testing"

	"github.com/PubMatic-OpenWrap/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/magiconair/properties/assert"
)

func TestUpdateAliasGVLIds(t *testing.T) {
	type args struct {
		aliasgvlids   map[string]uint16
		bidderCode    string
		partnerConfig map[string]string
	}
	tests := []struct {
		name            string
		args            args
		wantAliasgvlids map[string]uint16
	}{
		{
			name: "vendorId not present in partner config",
			args: args{
				aliasgvlids:   map[string]uint16{},
				bidderCode:    "pubmatic",
				partnerConfig: map[string]string{},
			},
			wantAliasgvlids: map[string]uint16{},
		},
		{
			name: "empty vendorId present in partner config",
			args: args{
				aliasgvlids: map[string]uint16{},
				bidderCode:  "pubmatic",
				partnerConfig: map[string]string{
					models.VENDORID: "",
				},
			},
			wantAliasgvlids: map[string]uint16{},
		},
		{
			name: "invalid vendorId present in partner config",
			args: args{
				aliasgvlids: map[string]uint16{},
				bidderCode:  "pubmatic",
				partnerConfig: map[string]string{
					models.VENDORID: "abc",
				},
			},
			wantAliasgvlids: map[string]uint16{},
		},
		{
			name: "vendorId=0 present in partner config",
			args: args{
				aliasgvlids: map[string]uint16{},
				bidderCode:  "pubmatic",
				partnerConfig: map[string]string{
					models.VENDORID: "abc",
				},
			},
			wantAliasgvlids: map[string]uint16{},
		},
		{
			name: "vendorId=0 present in partner config",
			args: args{
				aliasgvlids: map[string]uint16{},
				bidderCode:  "pubmatic",
				partnerConfig: map[string]string{
					models.VENDORID: "0",
				},
			},
			wantAliasgvlids: map[string]uint16{},
		},
		{
			name: "valid vendorId present in partner config",
			args: args{
				aliasgvlids: map[string]uint16{},
				bidderCode:  "pubmatic",
				partnerConfig: map[string]string{
					models.VENDORID: "72",
				},
			},
			wantAliasgvlids: map[string]uint16{"pubmatic": 72},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateAliasGVLIds(tt.args.aliasgvlids, tt.args.bidderCode, tt.args.partnerConfig)
			assert.Equal(t, tt.args.aliasgvlids, tt.wantAliasgvlids, tt.name)
		})
	}
}
