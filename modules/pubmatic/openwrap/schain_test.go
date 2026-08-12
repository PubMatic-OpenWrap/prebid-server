package openwrap

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestSetGlobalSChain(t *testing.T) {
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
		{
			name: "schain present in both request and DB, give preference to request",
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
								Name:   "RequestName1",
								Domain: "Domain1",
							},
						},
					},
				},
				partnerConfigMap: map[int]map[string]string{-1: {models.SChainObjectDBKey: `{"validation": "strict", "config": {"complete":1,"nodes":[{"asi":"ASI1","sid":"SID1","rid":"RID1","name":"DBName1","domain":"Domain1","hp":1}],"ver":"1.0"}}`}},
			},
			want: want{
				source: &openrtb2.Source{
					Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"ASI1","sid":"SID1","rid":"RID1","name":"RequestName1","domain":"Domain1","hp":1}],"ver":"1.0"}}`),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setGlobalSChain(tt.args.source, tt.args.partnerConfigMap)
			assert.Equal(t, tt.want.source, tt.args.source, tt.name)
		})
	}
}

func TestSetAllBidderSChain(t *testing.T) {
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
			name: "empty schains in DB",
			args: args{
				requestExt: &models.RequestExt{},
				partnerConfigMap: map[int]map[string]string{
					-1: {
						models.AllBidderSChainObj: ``,
					},
				},
			},
			want: want{
				requestExt: &models.RequestExt{},
			},
		},
		{
			name: "valid schains in DB",
			args: args{
				requestExt: &models.RequestExt{},
				partnerConfigMap: map[int]map[string]string{
					-1: {
						models.AllBidderSChainObj: `[{"bidders":["bidderA"],"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"example.com"}]}}]`,
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
			name: "initialized schains obj in request and valid schains present in DB",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							SChains: []*openrtb_ext.ExtRequestPrebidSChain{},
						},
					},
				},
				partnerConfigMap: map[int]map[string]string{
					-1: {
						models.AllBidderSChainObj: `[{"bidders":["bidderA"],"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"example.com"}]}}]`,
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
			name: "invalid schains in DB",
			args: args{
				requestExt: &models.RequestExt{},
				partnerConfigMap: map[int]map[string]string{
					-1: {
						models.AllBidderSChainObj: `invalid-json`,
					},
				},
			},
			want: want{
				requestExt: &models.RequestExt{},
			},
		},
		{
			name: "no schains object present in DB",
			args: args{
				requestExt: &models.RequestExt{},
				partnerConfigMap: map[int]map[string]string{
					-1: {
						models.AllBidderSChainObj: `[]`,
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
		{
			name: "valid schains present only in request",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							SChains: []*openrtb_ext.ExtRequestPrebidSChain{
								{
									Bidders: []string{"bidderA"},
									SChain:  openrtb2.SupplyChain{Ver: "1.0", Complete: 1, Nodes: []openrtb2.SupplyChainNode{{ASI: "request.com"}}},
								},
							},
						},
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
									SChain:  openrtb2.SupplyChain{Ver: "1.0", Complete: 1, Nodes: []openrtb2.SupplyChainNode{{ASI: "request.com"}}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid schains present in DB and request, give preference to request",
			args: args{
				requestExt: &models.RequestExt{
					ExtRequest: openrtb_ext.ExtRequest{
						Prebid: openrtb_ext.ExtRequestPrebid{
							SChains: []*openrtb_ext.ExtRequestPrebidSChain{
								{
									Bidders: []string{"bidderA"},
									SChain:  openrtb2.SupplyChain{Ver: "1.0", Complete: 1, Nodes: []openrtb2.SupplyChainNode{{ASI: "request.com"}}},
								},
							},
						},
					},
				},
				partnerConfigMap: map[int]map[string]string{
					-1: {
						models.AllBidderSChainObj: `[{"bidders":["bidderA"],"schain":{"ver":"1.0","complete":1,"nodes":[{"asi":"database.com"}]}}]`,
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
									SChain:  openrtb2.SupplyChain{Ver: "1.0", Complete: 1, Nodes: []openrtb2.SupplyChainNode{{ASI: "request.com"}}},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setAllBidderSChain(tt.args.requestExt, tt.args.partnerConfigMap)
			assert.Equal(t, tt.want.requestExt, tt.args.requestExt, tt.name)
		})
	}
}

func TestRemoveApplovinNode(t *testing.T) {
	type args struct {
		src *openrtb2.Source
	}
	type want struct {
		removed bool
		src     *openrtb2.Source
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "src_nil",
			args: args{src: nil},
			want: want{removed: false, src: nil},
		},
		{
			name: "no_schain_anywhere",
			args: args{src: &openrtb2.Source{}},
			want: want{removed: false, src: &openrtb2.Source{}},
		},
		{
			name: "source_schain_only_applovin_deleted",
			args: args{src: &openrtb2.Source{SChain: &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{{ASI: "applovin.com", SID: "s1"}}}}},
			want: want{removed: true, src: &openrtb2.Source{SChain: &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{}}}},
		},
		{
			name: "source_schain_mixed_only_applovin_removed",
			args: args{src: &openrtb2.Source{SChain: &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{{ASI: "applovin.com", SID: "s1"}, {ASI: "example.com", SID: "s2"}}}}},
			want: want{removed: true, src: &openrtb2.Source{SChain: &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{{ASI: "example.com", SID: "s2"}}}}},
		},
		{
			name: "source_ext_schain_no_nodes_noop",
			args: args{src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":0,"nodes":null,"ver":"1"}}`)}},
			want: want{removed: false, src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":0,"nodes":null,"ver":"1"}}`)}},
		},
		{
			name: "source_ext_schain_only_applovin_deleted",
			args: args{src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"applovin.com","sid":"s1"}],"ver":"1.0"}}`)}},
			want: want{removed: true, src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[],"ver":"1.0"}}`)}},
		},
		{
			name: "source_ext_schain_mixed_only_applovin_removed",
			args: args{src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"applovin.com","sid":"s1"},{"asi":"example.com","sid":"s2"}],"ver":"1.0"}}`)}},
			want: want{removed: true, src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"example.com","sid":"s2"}],"ver":"1.0"}}`)}},
		},
		{
			name: "both_source_and_ext_only_applovin_removed",
			args: args{src: &openrtb2.Source{SChain: &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{{ASI: "applovin.com", SID: "s1"}, {ASI: "example.com", SID: "s2"}}}, Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"applovin.com","sid":"s3"},{"asi":"example2.com","sid":"s4"}],"ver":"1.0"}}`)}},
			want: want{removed: true, src: &openrtb2.Source{SChain: &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{{ASI: "example.com", SID: "s2"}}}, Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"example2.com","sid":"s4"}],"ver":"1.0"}}`)}},
		},
		{
			name: "ext_invalid_json_still_removes_source_schain",
			args: args{src: &openrtb2.Source{SChain: &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{{ASI: "applovin.com", SID: "s1"}}}, Ext: json.RawMessage(`{`)}},
			want: want{removed: true, src: &openrtb2.Source{SChain: &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{}}, Ext: json.RawMessage(`{`)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removed := removeApplovinNode(tt.args.src)
			got := want{removed: removed, src: tt.args.src}
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}

func TestRemoveNode(t *testing.T) {
	type args struct {
		schain *openrtb2.SupplyChain
	}
	type want struct {
		removed bool
		schain  *openrtb2.SupplyChain
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "nil_schain",
			args: args{schain: nil},
			want: want{removed: false, schain: nil},
		},
		{
			name: "empty_nodes",
			args: args{schain: &openrtb2.SupplyChain{Nodes: nil}},
			want: want{removed: false, schain: &openrtb2.SupplyChain{Nodes: nil}},
		},
		{
			name: "no_match",
			args: args{schain: &openrtb2.SupplyChain{Nodes: []openrtb2.SupplyChainNode{{ASI: "example.com"}}}},
			want: want{removed: false, schain: &openrtb2.SupplyChain{Nodes: []openrtb2.SupplyChainNode{{ASI: "example.com"}}}},
		},
		{
			name: "case_insensitive_match",
			args: args{schain: &openrtb2.SupplyChain{Nodes: []openrtb2.SupplyChainNode{{ASI: "ApPlOvIn.CoM"}, {ASI: "example.com"}}}},
			want: want{removed: false, schain: &openrtb2.SupplyChain{Nodes: []openrtb2.SupplyChainNode{{ASI: "ApPlOvIn.CoM"}, {ASI: "example.com"}}}},
		},
		{
			name: "all_removed",
			args: args{schain: &openrtb2.SupplyChain{Nodes: []openrtb2.SupplyChainNode{{ASI: "applovin.com"}}}},
			want: want{removed: true, schain: &openrtb2.SupplyChain{Nodes: []openrtb2.SupplyChainNode{}}},
		},
		{
			name: "partial_removed_not_empty",
			args: args{schain: &openrtb2.SupplyChain{Nodes: []openrtb2.SupplyChainNode{{ASI: "applovin.com"}, {ASI: "example.com"}}}},
			want: want{removed: true, schain: &openrtb2.SupplyChain{Nodes: []openrtb2.SupplyChainNode{{ASI: "example.com"}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removed := removeNode(tt.args.schain)
			assert.Equal(t, tt.want, want{removed: removed, schain: tt.args.schain}, tt.name)
		})
	}
}

func TestRemoveNodeFromSourceExt(t *testing.T) {
	type args struct {
		src *openrtb2.Source
	}
	type want struct {
		removed bool
		src     *openrtb2.Source
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty_ext",
			args: args{src: &openrtb2.Source{Ext: nil}},
			want: want{removed: false, src: &openrtb2.Source{Ext: nil}},
		},
		{
			name: "no_schain_key",
			args: args{src: &openrtb2.Source{Ext: json.RawMessage(`{"foo":1}`)}},
			want: want{removed: false, src: &openrtb2.Source{Ext: json.RawMessage(`{"foo":1}`)}},
		},
		{
			name: "invalid_schain_json",
			args: args{src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{`)}},
			want: want{removed: false, src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{`)}},
		},
		{
			name: "schain_present_no_applovin_noop",
			args: args{src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"example.com","sid":"s1"}],"ver":"1.0"}}`)}},
			want: want{removed: false, src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"example.com","sid":"s1"}],"ver":"1.0"}}`)}},
		},
		{
			name: "schain_present_partial_removed_updated",
			args: args{src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"applovin.com","sid":"s1"},{"asi":"example.com","sid":"s2"}],"ver":"1.0"}}`)}},
			want: want{removed: true, src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"example.com","sid":"s2"}],"ver":"1.0"}}`)}},
		},
		{
			name: "schain_present_all_removed_deleted",
			args: args{src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[{"asi":"applovin.com","sid":"s1"}],"ver":"1.0"}}`)}},
			want: want{removed: true, src: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[],"ver":"1.0"}}`)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removed := removeNodeFromSourceExt(tt.args.src)
			assert.Equal(t, tt.want, want{removed: removed, src: tt.args.src}, tt.name)
		})
	}
}
