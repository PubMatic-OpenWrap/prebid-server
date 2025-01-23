package openwrap

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestSetGlobalSchain(t *testing.T) {
	type args struct {
		source           *openrtb2.Source
		partnerConfigMap map[int]map[string]string
	}
	type want struct {
		source *openrtb2.Source
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "schain present in request",
			args: args{
				source: &openrtb2.Source{
					SChain: &openrtb2.SupplyChain{
						Ver:      "1.0",
						Complete: 1,
						Nodes: []openrtb2.SupplyChainNode{
							{
								ASI:    "ASI1",
								SID:    "SID1",
								HP:     ptrutil.ToPtr(int8(1)),
								RID:    "RID1",
								Name:   "Name1",
								Domain: "Domain1",
							},
						},
					},
				},
			},
			want: want{
				source: &openrtb2.Source{
					Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"ASI1","sid":"SID1","rid":"RID1","name":"Name1","domain":"Domain1","hp":1}],"ver":"1.0"}}`),
				},
			},
		},
		{
			name: "schain not present in request and profile level schain object not present",
			args: args{
				source:           &openrtb2.Source{},
				partnerConfigMap: map[int]map[string]string{-1: {models.SChainDBKey: "1"}},
			},
			want: want{
				source: &openrtb2.Source{},
			},
		},
		{
			name: "schain not present in request and profile level schain object empty",
			args: args{
				source:           &openrtb2.Source{},
				partnerConfigMap: map[int]map[string]string{-1: {models.SChainDBKey: "1", models.SChainObjectDBKey: ""}},
			},
			want: want{
				source: &openrtb2.Source{},
			},
		},
		{
			name: "schain not present in request and invalid db schain",
			args: args{
				source: &openrtb2.Source{},
				partnerConfigMap: map[int]map[string]string{-1: {models.SChainDBKey: "1", models.SChainObjectDBKey: `{"validation": "strict", "config": {"ver": "1.0", "complete": 1, "nodes": [
					{"asi": "indirectseller.com", "sid":{}, "hp": 1}]}}`}},
			},
			want: want{
				source: &openrtb2.Source{},
			},
		},
		{
			name: "schain not present in request but source.ext is invalid, set correct source.ext.schain",
			args: args{
				source: &openrtb2.Source{
					Ext: json.RawMessage(`{`),
				},
				partnerConfigMap: map[int]map[string]string{-1: {models.SChainDBKey: "1", models.SChainObjectDBKey: `{"validation": "strict", "config": {"complete":1,"nodes":[{"asi":"ASI1","sid":"SID1","rid":"RID1","name":"Name1","domain":"Domain1","hp":1}],"ver":"1.0"}}`}},
			},
			want: want{
				source: &openrtb2.Source{
					Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"ASI1","sid":"SID1","rid":"RID1","name":"Name1","domain":"Domain1","hp":1}],"ver":"1.0"}}`),
				},
			},
		},
		{
			name: "schain not present in request and valid profile level schain",
			args: args{
				source:           &openrtb2.Source{},
				partnerConfigMap: map[int]map[string]string{-1: {models.SChainDBKey: "1", models.SChainObjectDBKey: `{"validation": "strict", "config": {"complete":1,"nodes":[{"asi":"ASI1","sid":"SID1","rid":"RID1","name":"Name1","domain":"Domain1","hp":1}],"ver":"1.0"}} `}},
			},
			want: want{
				source: &openrtb2.Source{
					Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"ASI1","sid":"SID1","rid":"RID1","name":"Name1","domain":"Domain1","hp":1}],"ver":"1.0"}}`),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setGlobalSchain(tt.args.source, tt.args.partnerConfigMap)
			assert.Equal(t, tt.want.source, tt.args.source, tt.name)
		})
	}
}

func TestSetAllBidderSchain(t *testing.T) {
	type args struct {
		requestExt       *models.RequestExt
		partnerConfigMap map[int]map[string]string
	}
	type want struct {
		requestExt *models.RequestExt
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty models.AllBidderSchainObject",
			args: args{
				requestExt: &models.RequestExt{},
				partnerConfigMap: map[int]map[string]string{
					-1: {
						models.AllBidderSchainObj: ``,
					},
				},
			},
			want: want{
				requestExt: &models.RequestExt{},
			},
		},
		{
			name: "valid models.AllBidderSchainObject",
			args: args{
				requestExt: &models.RequestExt{},
				partnerConfigMap: map[int]map[string]string{
					-1: {
						models.AllBidderSchainObj: `[{"bidders":["bidderA"],"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"example.com"}]}}]`,
					},
				},
			},
			want: want{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							SChains: []*openrtb_ext.ExtRequestPrebidSChain{
								{
									Bidders: []string{"bidderA"},
									SChain:  openrtb2.SupplyChain{Ver: "1.0", Complete: 1, Nodes: []openrtb2.SupplyChainNode{{ASI: "example.com"}}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "invalid models.AllBidderSchainObject",
			args: args{
				requestExt: &models.RequestExt{},
				partnerConfigMap: map[int]map[string]string{
					-1: {
						models.AllBidderSchainObj: `invalid-json`,
					},
				},
			},
			want: want{
				requestExt: &models.RequestExt{},
			},
		},
		{
			name: "no object present in models.AllBidderSchainObject",
			args: args{
				requestExt: &models.RequestExt{},
				partnerConfigMap: map[int]map[string]string{
					-1: {
						models.AllBidderSchainObj: `[]`,
					},
				},
			},
			want: want{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							SChains: []*openrtb_ext.ExtRequestPrebidSChain{},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setAllBidderSchain(tt.args.requestExt, tt.args.partnerConfigMap)
			assert.Equal(t, tt.want.requestExt, tt.args.requestExt, tt.name)
		})
	}
}
