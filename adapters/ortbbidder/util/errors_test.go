package util

import (
	"testing"

	"github.com/prebid/prebid-server/v3/errortypes"
	"github.com/stretchr/testify/assert"
)

func TestNewBadInputError(t *testing.T) {
	type args struct {
		message string
		args    []any
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "bad input error with params",
			args: args{
				message: "bad input error [%s]",
				args:    []any{"field"},
			},
			wantErr: &errortypes.BadInput{
				Message: "bad input error [field]",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewBadInputError(tt.args.message, tt.args.args...)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestNewBadServerResponseError(t *testing.T) {
	type args struct {
		message string
		args    []any
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "bad serevr error with params",
			args: args{
				message: "bad input error [%s]",
				args:    []any{"field"},
			},
			wantErr: &errortypes.BadServerResponse{
				Message: "bad input error [field]",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewBadServerResponseError(tt.args.message, tt.args.args...)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestNewWarning(t *testing.T) {
	type args struct {
		message string
		args    []any
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "warning with params",
			args: args{
				message: "bad error [%s] : [%d]",
				args:    []any{"field", 10},
			},
			wantErr: &errortypes.Warning{
				Message: "bad error [field] : [10]",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewWarning(tt.args.message, tt.args.args...)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
