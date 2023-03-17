package processor

import (
	"testing"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func Test_stringBasedProcessor_Replace(t *testing.T) {

	type args struct {
		url              string
		getMacroProvider func() Provider
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
				url: "http://tracker.com?macro1=##PBS-BIDID##&macro2=##PBS-APPBUNDLE##&macro3=##PBS-APPBUNDLE##&macro4=##PBS-PUBDOMAIN##&macro5=##PBS-PAGEURL##&macro6=##PBS-ACCOUNTID##&macro6=##PBS-LIMITADTRACKING##&macro7=##PBS-GDPRCONSENT##&macro8=##PBS-GDPRCONSENT##&macro9=##PBS-MACRO_CUSTOMMACRO1##&macro10=##PBS-MACRO_CUSTOMMACRO2##",
				getMacroProvider: func() Provider {
					macroProvider := NewProvider(req)
					macroProvider.SetContext(bid, nil, "test", "123", config.FirstQuartile, config.TrackingVASTElement)
					return macroProvider
				},
			},
			want:    "http://tracker.com?macro1=bidId123&macro2=testdomain&macro3=testdomain&macro4=publishertestdomain&macro5=pageurltest&macro6=testpublisherID&macro6=10&macro7=yes&macro8=yes&macro9=&macro10=&macro11=&macro12=&macro13=&macro14=firstQuartile&macro15=&macro16=&macro17=123&macro18=campaign_1",
			wantErr: false,
		},
		{
			name: "url does not have macro",
			args: args{
				url: "http://tracker.com",
				getMacroProvider: func() Provider {
					macroProvider := NewProvider(req)
					macroProvider.SetContext(bid, nil, "test", "123", config.FirstQuartile, config.TrackingVASTElement)
					return macroProvider
				},
			},
			want:    "http://tracker.com",
			wantErr: false,
		},
		{
			name: "macro not found",
			args: args{
				url: "http://tracker.com?macro1=##PBS-test1##",
				getMacroProvider: func() Provider {
					macroProvider := NewProvider(&openrtb_ext.RequestWrapper{BidRequest: &openrtb2.BidRequest{}})
					macroProvider.SetContext(bid, nil, "test", "123", config.FirstQuartile, config.TrackingVASTElement)
					return macroProvider
				},
			},
			want:    "http://tracker.com?macro1=",
			wantErr: false,
		},
		{
			name: "tracker url is empty",
			args: args{
				url: "",
				getMacroProvider: func() Provider {
					macroProvider := NewProvider(&openrtb_ext.RequestWrapper{BidRequest: &openrtb2.BidRequest{}})
					macroProvider.SetContext(bid, nil, "test", "123", config.FirstQuartile, config.TrackingVASTElement)
					return macroProvider
				},
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor()
			got, err := processor.Replace(tt.args.url, tt.args.getMacroProvider())
			if (err != nil) != tt.wantErr {
				t.Errorf("stringBasedProcessor.Replace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("stringBasedProcessor.Replace() = %v, want %v", got, tt.want)
			}
		})
	}
}
