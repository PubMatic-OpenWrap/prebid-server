package openrtb_ext

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
	"github.com/stretchr/testify/assert"
)

func TestCloneSupplyChain(t *testing.T) {
	testCases := []struct {
		name       string
		schain     *openrtb2.SupplyChain
		schainCopy *openrtb2.SupplyChain                            // manual copy of above prebid object to verify against
		mutator    func(t *testing.T, schain *openrtb2.SupplyChain) // function to modify the prebid object
	}{
		{
			name:       "Nil", // Verify the nil case
			schain:     nil,
			schainCopy: nil,
			mutator:    func(t *testing.T, schain *openrtb2.SupplyChain) {},
		},
		{
			name: "General",
			schain: &openrtb2.SupplyChain{
				Complete: 2,
				Nodes: []openrtb2.SupplyChainNode{
					{
						SID:  "alpha",
						Name: "Johnny",
						HP:   ptrutil.ToPtr[int8](5),
						Ext:  json.RawMessage(`{}`),
					},
					{
						ASI:  "Oh my",
						Name: "Johnny",
						HP:   ptrutil.ToPtr[int8](5),
						Ext:  json.RawMessage(`{"samson"}`),
					},
				},
				Ver: "v2.5",
				Ext: json.RawMessage(`{"foo": "bar"}`),
			},
			schainCopy: &openrtb2.SupplyChain{
				Complete: 2,
				Nodes: []openrtb2.SupplyChainNode{
					{
						SID:  "alpha",
						Name: "Johnny",
						HP:   ptrutil.ToPtr[int8](5),
						Ext:  json.RawMessage(`{}`),
					},
					{
						ASI:  "Oh my",
						Name: "Johnny",
						HP:   ptrutil.ToPtr[int8](5),
						Ext:  json.RawMessage(`{"samson"}`),
					},
				},
				Ver: "v2.5",
				Ext: json.RawMessage(`{"foo": "bar"}`),
			},
			mutator: func(t *testing.T, schain *openrtb2.SupplyChain) {
				schain.Nodes[0].SID = "beta"
				schain.Nodes[1].HP = nil
				schain.Nodes[0].Ext = nil
				schain.Nodes = append(schain.Nodes, openrtb2.SupplyChainNode{SID: "Gamma"})
				schain.Complete = 0
				schain.Ext = json.RawMessage(`{}`)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			clone := cloneSupplyChain(test.schain)
			test.mutator(t, test.schain)
			assert.Equal(t, test.schainCopy, clone)
		})
	}
}

func TestSerializeSupplyChain(t *testing.T) {
	type args struct {
		schain *openrtb2.SupplyChain
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "single hop - chain complete",
			args: args{schain: &openrtb2.SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []openrtb2.SupplyChainNode{
					{
						ASI:    "exchange1.com",
						SID:    "1234",
						RID:    "bid-request-1",
						Name:   "publisher",
						Domain: "publisher.com",
						HP:     openrtb2.Int8Ptr(1),
					},
				}}},
			want: "1.0,1!exchange1.com,1234,1,bid-request-1,publisher,publisher.com",
		},
		{
			name: "multiple hops - with all properties supplied",
			args: args{schain: &openrtb2.SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []openrtb2.SupplyChainNode{
					{
						ASI:    "exchange1.com",
						SID:    "1234",
						HP:     openrtb2.Int8Ptr(1),
						RID:    "bid-request-1",
						Name:   "publisher",
						Domain: "publisher.com",
					},
					{
						ASI:    "exchange2.com",
						SID:    "abcd",
						HP:     openrtb2.Int8Ptr(1),
						RID:    "bid-request-2",
						Name:   "intermediary",
						Domain: "intermediary.com",
					},
				}}},
			want: "1.0,1!exchange1.com,1234,1,bid-request-1,publisher,publisher.com!exchange2.com,abcd,1,bid-request-2,intermediary,intermediary.com",
		},
		{
			name: "single hop - chain incomplete",
			args: args{schain: &openrtb2.SupplyChain{
				Complete: 0,
				Ver:      "1.0",
				Nodes: []openrtb2.SupplyChainNode{
					{
						ASI:  "exchange1.com",
						SID:  "1234",
						RID:  "bid-request-1",
						Name: "publisher",
						HP:   openrtb2.Int8Ptr(1),
					},
				}}},
			want: "1.0,0!exchange1.com,1234,1,bid-request-1,publisher,",
		},
		{
			name: "single hop - chain complete, encoded values",
			args: args{schain: &openrtb2.SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []openrtb2.SupplyChainNode{
					{
						ASI:    "exchange1.com",
						SID:    "1234!abcd",
						HP:     openrtb2.Int8Ptr(1),
						RID:    "bid-request-1",
						Name:   "publisher, Inc.",
						Domain: "publisher.com",
					},
				}}},
			want: "1.0,1!exchange1.com,1234%21abcd,1,bid-request-1,publisher%2C+Inc.,publisher.com",
		},
		{
			name: "multiple hops - with all properties supplied,encoded values",
			args: args{schain: &openrtb2.SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []openrtb2.SupplyChainNode{
					{
						ASI:    "exchange1.com",
						SID:    "1234&abcd",
						HP:     openrtb2.Int8Ptr(1),
						RID:    "bid-request-1",
						Name:   "publisher?name",
						Domain: "publisher.com",
					},
					{
						ASI:    "exchange2.com",
						SID:    "abcd?12345",
						HP:     openrtb2.Int8Ptr(1),
						RID:    "bid-request-2",
						Name:   "intermediary",
						Domain: "intermediary.com",
					},
				}}},
			want: "1.0,1!exchange1.com,1234%26abcd,1,bid-request-1,publisher%3Fname,publisher.com!exchange2.com,abcd%3F12345,1,bid-request-2,intermediary,intermediary.com",
		},
		{
			name: "single hop - chain complete, missing optional field - encoded values",
			args: args{schain: &openrtb2.SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []openrtb2.SupplyChainNode{
					{
						ASI:  "exchange1.com",
						SID:  "1234&&abcd",
						HP:   openrtb2.Int8Ptr(1),
						RID:  "bid-request-1",
						Name: "publisher, Inc.",
					},
				}}},
			want: "1.0,1!exchange1.com,1234%26%26abcd,1,bid-request-1,publisher%2C+Inc.,",
		},
		{
			name: "single hop - chain complete, missing domain field - encoded values",
			args: args{schain: &openrtb2.SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []openrtb2.SupplyChainNode{
					{
						ASI:  "exchange1.com",
						SID:  "1234!abcd",
						HP:   openrtb2.Int8Ptr(1),
						RID:  "bid-request-1",
						Name: "publisher, Inc.",
					},
				}}},
			want: "1.0,1!exchange1.com,1234%21abcd,1,bid-request-1,publisher%2C+Inc.,",
		},
		{
			name: "single hop with extension - chain complete",
			args: args{schain: &openrtb2.SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []openrtb2.SupplyChainNode{
					{
						ASI:    "exchange1.com",
						SID:    "1234",
						RID:    "bid-request-1",
						Name:   "publisher",
						Domain: "publisher.com",
						HP:     openrtb2.Int8Ptr(1),
						Ext:    []byte(`{"test":1,"url":"https://testuser.com?k1=v1&k2=v2"}`),
					},
				}}},
			want: "1.0,1!exchange1.com,1234,1,bid-request-1,publisher,publisher.com,%7B%22test%22%3A1%2C%22url%22%3A%22https%3A%2F%2Ftestuser.com%3Fk1%3Dv1%26k2%3Dv2%22%7D",
		},
		{
			name: "single hop with encoded extension - chain complete",
			args: args{schain: &openrtb2.SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []openrtb2.SupplyChainNode{
					{
						ASI:    "exchange1.com",
						SID:    "1234",
						RID:    "bid-request-1",
						Name:   "publisher",
						Domain: "publisher.com",
						HP:     openrtb2.Int8Ptr(1),
						Ext:    []byte(`%7B%22test%22%3A1%7D`),
					},
				}}},
			want: "1.0,1!exchange1.com,1234,1,bid-request-1,publisher,publisher.com,%257B%2522test%2522%253A1%257D",
		},
		{
			name: "multiple hops with extension - some optional properties not supplied",
			args: args{schain: &openrtb2.SupplyChain{
				Complete: 1,
				Ver:      "1.0",
				Nodes: []openrtb2.SupplyChainNode{
					{
						ASI:    "exchange1.com",
						SID:    "1234",
						HP:     openrtb2.Int8Ptr(1),
						RID:    "bid-request-1",
						Domain: "publisher.com",
						Ext:    []byte(`{"test1":1,"name":"test","age":22}`),
					},
					{
						ASI:  "exchange2.com",
						SID:  "abcd",
						HP:   openrtb2.Int8Ptr(1),
						RID:  "bid-request-2",
						Name: "intermediary",
						Ext:  []byte(`{"test"}`),
					},
				}}},
			want: "1.0,1!exchange1.com,1234,1,bid-request-1,,publisher.com,%7B%22test1%22%3A1%2C%22name%22%3A%22test%22%2C%22age%22%3A22%7D!exchange2.com,abcd,1,bid-request-2,intermediary,,%7B%22test%22%7D",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SerializeSupplyChain(tt.args.schain)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
