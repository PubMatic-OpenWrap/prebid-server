package wakanda

import (
	"errors"
	"io"
	"os"
	"testing"

	mock_wakanda "github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/wakanda/mock"

	"github.com/golang/mock/gomock"
)

func Test_send(t *testing.T) {
	type args struct {
		mockSFTPfunc        func(*mock_wakanda.MockCommands)
		destinationFileName string
		pubProfDir          string
		cfg                 SFTP
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid_scp_path", args: args{
				mockSFTPfunc: func(mockSFTP *mock_wakanda.MockCommands) {
					mockSFTP.EXPECT().Start().Return(nil)
					mockSFTP.EXPECT().StdinPipe().Return(io.WriteCloser(os.Stdout), nil)
					mockSFTP.EXPECT().Wait().Return(nil).AnyTimes()
				},
				destinationFileName: "my_test_file",
				cfg: SFTP{
					User:        "user",
					Password:    "mypass",
					ServerIP:    "10.20.30.40",
					Destination: "/path",
				},
			}, want: true,
		},
		{
			name:    "invalid_destination_file",
			wantErr: true,
			args:    args{destinationFileName: "/"},
		},
		{
			name:    "command_input_error",
			wantErr: true,
			args: args{mockSFTPfunc: func(mockSFTP *mock_wakanda.MockCommands) {
				mockSFTP.EXPECT().StdinPipe().Return(nil, errors.New("some_error")).AnyTimes()

			}},
			want: false,
		},
		{
			name:    "command_start_error",
			wantErr: true,
			args: args{mockSFTPfunc: func(mockSFTP *mock_wakanda.MockCommands) {
				mockSFTP.EXPECT().StdinPipe().Return(io.WriteCloser(os.Stdout), nil).AnyTimes()
				mockSFTP.EXPECT().Start().Return(errors.New("some_error")).AnyTimes()
			}},
			want: false,
		},
		{
			name: "command_wait_error",
			want: true,
			// wantErr: true, // can't collect this from go routine
			args: args{mockSFTPfunc: func(mockSFTP *mock_wakanda.MockCommands) {
				mockSFTP.EXPECT().StdinPipe().Return(io.WriteCloser(os.Stdout), nil).AnyTimes()
				mockSFTP.EXPECT().Start().Return(nil).AnyTimes()
				mockSFTP.EXPECT().Wait().Return(errors.New("some_error")).AnyTimes()
			}},
		},
	}
	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		mockCommands := mock_wakanda.NewMockCommands(ctrl)
		if tt.args.mockSFTPfunc != nil {
			tt.args.mockSFTPfunc(mockCommands)
		}
		commandHandler.commandExecutor = &mockcommandExecutor{
			mockCommands,
		}
		got, err := send(tt.args.destinationFileName, tt.args.pubProfDir, []byte(`some_log`), tt.args.cfg)
		if (err != nil) != tt.wantErr {
			t.Errorf("SFTPLog() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if got != tt.want {
			t.Errorf("SFTPLog() = %v, want %v", got, tt.want)
		}
	}
}

type mockcommandExecutor struct {
	commandExecutor Commands
}

func (h *mockcommandExecutor) Command() Commands {
	return h.commandExecutor
}
