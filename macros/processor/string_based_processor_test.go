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
			name: "success",
			args: args{
				url: "http://tracker.com?macro1=##PBS-BIDID##&macro2=##PBS-APPBUNDLE##&macro3=##PBS-DOMAIN##&macro4=##PBS-PUBDOMAIN##&macro5=##PBS-PAGEURL##&macro6=##PBS-ACCOUNTID##&macro7=##PBS-LIMITADTRACKING##&macro8=##PBS-GDPRCONSENT##&macro9=##PBS-MACRO_##&macro10=##PBS-BIDDER##&macro11=##PBS-INTEGRATION##&macro12=##PBS-VASTCRTID##&macro13=##PBS-LINEID##&macro15=##PBS-AUCTIONID##&macro16=##PBS-CHANNEL##&macro17=##PBS-EVENTTYPE##&macro18=##PBS-VASTEVENT##",
				getMacroProvider: func() Provider {
					macroProvider := NewProvider(req)

					macroProvider.SetContext(MacroContext{
						Bid:            bid,
						Imp:            nil,
						Seat:           "test",
						VastCreativeID: "123",
						VastEventType:  config.FirstQuartile,
						EventElement:   config.TrackingVASTElement,
					})
					return macroProvider
				},
			},
			want:    "http://tracker.com?macro1=bidId123&macro2=testbundle&macro3=testdomain&macro4=publishertestdomain&macro5=pageurltest&macro6=testpublisherID&macro7=10&macro8=yes&macro9=&macro10=test&macro11=&macro12=123&macro13=campaign_1&macro15=123&macro16=&macro17=firstQuartile&macro18=tracking",
			wantErr: false,
		},
		{
			name: "url does not have macro",
			args: args{
				url: "http://tracker.com",
				getMacroProvider: func() Provider {
					macroProvider := NewProvider(req)

					macroProvider.SetContext(MacroContext{
						Bid:            bid,
						Imp:            nil,
						Seat:           "test",
						VastCreativeID: "123",
						VastEventType:  config.FirstQuartile,
						EventElement:   config.TrackingVASTElement,
					})
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

					macroProvider.SetContext(MacroContext{
						Bid:            bid,
						Imp:            nil,
						Seat:           "test",
						VastCreativeID: "123",
						VastEventType:  config.FirstQuartile,
						EventElement:   config.TrackingVASTElement,
					})
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

					macroProvider.SetContext(MacroContext{
						Bid:            bid,
						Imp:            nil,
						Seat:           "test",
						VastCreativeID: "123",
						VastEventType:  config.FirstQuartile,
						EventElement:   config.TrackingVASTElement,
					})
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
