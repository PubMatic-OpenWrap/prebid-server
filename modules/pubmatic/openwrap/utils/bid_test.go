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
			name: "Split the bid Id",
			args: args{
				bidId: "original-id::gen-id",
			},
			want: "original-id",
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
