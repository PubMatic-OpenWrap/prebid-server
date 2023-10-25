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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetOriginalBidId(tt.args.bidId); got != tt.want {
				t.Errorf("GetOriginalBidId() = %v, want %v", got, tt.want)
			}
		})
	}
}
