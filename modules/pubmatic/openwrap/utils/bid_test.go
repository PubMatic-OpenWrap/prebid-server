package utils

import (
	"testing"
)

func TestGetOriginalBidId(t *testing.T) {
	type args struct {
		bidId string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Split the bid Id when get valid bidId",
			args: args{
				bidId: "original-id::gen-id",
			},
			want: "original-id",
		},
		{
			name: "Empty BidId",
			args: args{
				bidId: "",
			},
			want: "",
		},
		{
			name: "Partial BidId",
			args: args{
				bidId: "::gen-id",
			},
			want: "",
		},
		{
			name: "Partial BidId without generated and separator",
			args: args{
				bidId: "original-bid-1",
			},
			want: "original-bid-1",
		},
		{
			name: "Partial BidId without generated",
			args: args{
				bidId: "original-bid::",
			},
			want: "original-bid",
		},
		{
			name: "BidId with single colon in origin Id",
			args: args{
				bidId: "original-bid:2::generated-bid",
			},
			want: "original-bid:2",
		},
		{
			name: "BidId with single colon in generated Id",
			args: args{
				bidId: "original-bid:2::generated-bid:3",
			},
			want: "original-bid:2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetOriginalBidId(tt.args.bidId); got != tt.want {
				t.Errorf("GetOriginalBidId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetUniqueBidID(t *testing.T) {
	type args struct {
		originalBidID  string
		generatedBidID string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Unique bid id will be generated",
			args: args{
				originalBidID:  "orig-bid",
				generatedBidID: "gen-bid",
			},
			want: "orig-bid::gen-bid",
		},
		{
			name: "Original Bid Id empty",
			args: args{
				originalBidID:  "",
				generatedBidID: "gen-bid",
			},
			want: "::gen-bid",
		},
		{
			name: "generated BidId empty",
			args: args{
				originalBidID:  "orig-bid",
				generatedBidID: "",
			},
			want: "orig-bid::",
		},
		{
			name: "Both Id empty",
			args: args{
				originalBidID:  "",
				generatedBidID: "",
			},
			want: "::",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetUniqueBidID(tt.args.originalBidID, tt.args.generatedBidID); got != tt.want {
				t.Errorf("SetUniqueBidId() = %v, want %v", got, tt.want)
			}
		})
	}
}
