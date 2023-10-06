package models

import (
	"fmt"
	"testing"
)

func TestErrorWrap(t *testing.T) {
	type args struct {
		cErr error
		nErr error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "current error as nil",
			args: args{
				cErr: nil,
				nErr: fmt.Errorf("error found for %d", 1234),
			},
			wantErr: true,
		},
		{
			name: "wrap error",
			args: args{
				cErr: fmt.Errorf("current error found for %d", 1234),
				nErr: fmt.Errorf("new error found for %d", 1234),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ErrorWrap(tt.args.cErr, tt.args.nErr); (err != nil) != tt.wantErr {
				t.Errorf("ErrorWrap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
