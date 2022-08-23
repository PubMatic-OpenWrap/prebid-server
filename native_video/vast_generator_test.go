package native_video

import "testing"

func Test_generateVASTXml(t *testing.T) {
	type args struct {
		price        string
		httpFilePath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				price:        "25",
				httpFilePath: "https://tech-stack-mgmt.pubmatic.com/owtools/hackathon2k22/owtools/api/getbid?reqid=11",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateVASTXml(tt.args.price, tt.args.httpFilePath); got != tt.want {
				t.Errorf("generateVASTXml() = %v, want %v", got, tt.want)
			}
		})
	}
}
