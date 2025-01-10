package openrtb_ext

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/stretchr/testify/assert"
)

func TestNewNonBid(t *testing.T) {
	tests := []struct {
		name           string
		bidParams      NonBidParams
		expectedNonBid NonBid
	}{
		{
			name:           "nil-bid-present-in-bidparams",
			bidParams:      NonBidParams{Bid: nil},
			expectedNonBid: NonBid{ImpId: "", StatusCode: 0, Ext: ExtNonBid{Prebid: ExtNonBidPrebid{Bid: ExtNonBidPrebidBid{Price: 0, ADomain: []string(nil), CatTax: 0, Cat: []string(nil), DealID: "", W: 0, H: 0, Dur: 0, MType: 0, OriginalBidCPM: 0, OriginalBidCur: "", ID: fakeUuid, DealPriority: 0, DealTierSatisfied: false, Meta: (*ExtBidPrebidMeta)(nil), Targeting: map[string]string(nil), Type: "", Video: (*ExtBidPrebidVideo)(nil), BidId: "", Floors: (*ExtBidPrebidFloors)(nil), OriginalBidCPMUSD: 0}}, IsAdPod: (*bool)(nil)}},
		},
		{
			name:           "non-nil-bid-present-in-bidparams",
			bidParams:      NonBidParams{Bid: &openrtb2.Bid{ImpID: "imp1"}, NonBidReason: 100},
			expectedNonBid: NonBid{ImpId: "imp1", StatusCode: 100, Ext: ExtNonBid{Prebid: ExtNonBidPrebid{Bid: ExtNonBidPrebidBid{Price: 0, ADomain: []string(nil), CatTax: 0, Cat: []string(nil), DealID: "", W: 0, H: 0, Dur: 0, MType: 0, OriginalBidCPM: 0, OriginalBidCur: "", ID: fakeUuid, DealPriority: 0, DealTierSatisfied: false, Meta: (*ExtBidPrebidMeta)(nil), Targeting: map[string]string(nil), Type: "", Video: (*ExtBidPrebidVideo)(nil), BidId: "", Floors: (*ExtBidPrebidFloors)(nil), OriginalBidCPMUSD: 0}}, IsAdPod: (*bool)(nil)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuidGenerator = TestUuidGeneratorInstance()
			nonBid := NewNonBid(tt.bidParams)
			nonBid.Ext.Prebid.Bid.ID, _ = uuidGenerator.Generate()
			assert.Equal(t, tt.expectedNonBid, nonBid, "found incorrect nonBid")
		})
	}
}

func TestSeatNonBidsAdd(t *testing.T) {
	type fields struct {
		seatNonBidsMap map[string][]NonBid
	}
	type args struct {
		nonbid NonBid
		seat   string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string][]NonBid
	}{
		{
			name:   "nil-seatNonBidsMap",
			fields: fields{seatNonBidsMap: nil},
			args: args{
				nonbid: NonBid{},
				seat:   "bidder1",
			},
			want: sampleSeatNonBidMap("bidder1", 1),
		},
		{
			name:   "non-nil-seatNonBidsMap",
			fields: fields{seatNonBidsMap: nil},
			args: args{

				nonbid: NonBid{},
				seat:   "bidder1",
			},
			want: sampleSeatNonBidMap("bidder1", 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snb := &NonBidCollection{
				seatNonBidsMap: tt.fields.seatNonBidsMap,
			}
			snb.AddBid(tt.args.nonbid, tt.args.seat)
			assert.Equalf(t, tt.want, snb.seatNonBidsMap, "found incorrect seatNonBidsMap")
		})
	}
}

func TestSeatNonBidsGet(t *testing.T) {
	type fields struct {
		snb *NonBidCollection
	}
	tests := []struct {
		name   string
		fields fields
		want   []SeatNonBid
	}{
		{
			name:   "get-seat-nonbids",
			fields: fields{&NonBidCollection{sampleSeatNonBidMap("bidder1", 2)}},
			want:   sampleSeatBids("bidder1", 2),
		},
		{
			name:   "nil-seat-nonbids",
			fields: fields{nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.snb.Get(); !assert.Equal(t, tt.want, got) {
				t.Errorf("seatNonBids.get() = %v, want %v", got, tt.want)
			}
		})
	}
}

var sampleSeatNonBidMap = func(seat string, nonBidCount int) map[string][]NonBid {
	nonBids := make([]NonBid, 0)
	for i := 0; i < nonBidCount; i++ {
		nonBids = append(nonBids, NonBid{
			Ext: ExtNonBid{Prebid: ExtNonBidPrebid{Bid: ExtNonBidPrebidBid{}}},
		})
	}
	return map[string][]NonBid{
		seat: nonBids,
	}
}

var sampleSeatBids = func(seat string, nonBidCount int) []SeatNonBid {
	seatNonBids := make([]SeatNonBid, 0)
	seatNonBid := SeatNonBid{
		Seat:   seat,
		NonBid: make([]NonBid, 0),
	}
	for i := 0; i < nonBidCount; i++ {
		seatNonBid.NonBid = append(seatNonBid.NonBid, NonBid{
			Ext: ExtNonBid{Prebid: ExtNonBidPrebid{Bid: ExtNonBidPrebidBid{}}},
		})
	}
	seatNonBids = append(seatNonBids, seatNonBid)
	return seatNonBids
}

func TestSeatNonBidsMerge(t *testing.T) {
	type target struct {
		snb *NonBidCollection
	}
	tests := []struct {
		name   string
		fields target
		input  NonBidCollection
		want   *NonBidCollection
	}{
		{
			name:   "target-NonBidCollection-is-nil",
			fields: target{nil},
			want:   nil,
		},
		{
			name:   "input-NonBidCollection-contains-nil-map",
			fields: target{&NonBidCollection{}},
			input:  NonBidCollection{seatNonBidsMap: nil},
			want:   &NonBidCollection{},
		},
		{
			name:   "input-NonBidCollection-contains-empty-nonBids",
			fields: target{&NonBidCollection{}},
			input:  NonBidCollection{seatNonBidsMap: make(map[string][]NonBid)},
			want:   &NonBidCollection{},
		},
		{
			name:   "append-nonbids-in-empty-target-NonBidCollection",
			fields: target{&NonBidCollection{}},
			input: NonBidCollection{
				seatNonBidsMap: sampleSeatNonBidMap("pubmatic", 1),
			},
			want: &NonBidCollection{
				seatNonBidsMap: sampleSeatNonBidMap("pubmatic", 1),
			},
		},
		{
			name: "merge-multiple-nonbids-in-non-empty-target-NonBidCollection",
			fields: target{&NonBidCollection{
				seatNonBidsMap: sampleSeatNonBidMap("pubmatic", 1),
			}},
			input: NonBidCollection{
				seatNonBidsMap: sampleSeatNonBidMap("pubmatic", 1),
			},
			want: &NonBidCollection{
				seatNonBidsMap: sampleSeatNonBidMap("pubmatic", 2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.snb.Append(tt.input)
			assert.Equal(t, tt.want, tt.fields.snb, "incorrect NonBidCollection generated by Append")
		})
	}
}
