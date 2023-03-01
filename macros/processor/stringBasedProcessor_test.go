package processor

import (
	"testing"

	"github.com/prebid/prebid-server/config"
)

func Test_stringIndexCachedProcessor_Replace(t *testing.T) {

	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "string index cached replace",
			args: args{
				url: "http://tracker.com?macro1=##PBS_BIDID##&macro2=##PBS_APPBUNDLE##&macro3=##PBS_APPBUNDLE##&macro4=##PBS_PUBDOMAIN##&macro5=##PBS_PAGEURL##&macro6=##PBS_ACCOUNTID##&macro6=##PBS_LIMITADTRACKING##&macro7=##PBS_GDPRCONSENT##&macro8=##PBS_GDPRCONSENT##&macro9=##PBS_MACRO_CUSTOMMACRO1##&macro10=##PBS_MACRO_CUSTOMMACRO2##",
			},
			want:    "http://tracker.com?macro1=bidId123&macro2=testdomain&macro3=testdomain&macro4=publishertestdomain&macro5=pageurltest&macro6=testpublisherID&macro6=10&macro7=yes&macro8=yes&macro9=&macro10=",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor(config.MacroProcessorConfig{
				ProcessorType: config.StringBasedProcessor,
			})
			macroProvider := NewProvider(req)
			macroProvider.SetContext(bid, nil)
			got, err := processor.Replace(tt.args.url, macroProvider)
			if (err != nil) != tt.wantErr {
				t.Errorf("stringIndexCachedProcessor.Replace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("stringIndexCachedProcessor.Replace() = %v, want %v", got, tt.want)
			}
		})
	}
}
