package openwrap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVastUnwrapperEnable(t *testing.T) {
	type args struct {
		ctx   context.Context
		field string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "given field present in context",
			args: args{
				ctx:   context.WithValue(context.Background(), "abc", "1"),
				field: "abc",
			},
			want: true,
		},
		{
			name: "given field is not present in context",
			args: args{
				ctx:   context.WithValue(context.Background(), "abc", "1"),
				field: "xyz",
			},
			want: false,
		},
		{
			name: "No field is not present in context",
			args: args{
				ctx:   context.Background(),
				field: "xyz",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getVastUnwrapperEnable(tt.args.ctx, tt.args.field)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
