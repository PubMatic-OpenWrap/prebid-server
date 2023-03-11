package processor

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

func Test_macroProvider_GetAllMacros(t *testing.T) {

	type args struct {
		keys []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "get all macros success",
			args: args{keys: []string{BidIDKey, AccountIDKey, AppBundleKey, PubDomainkey,
				PageURLKey, AccountIDKey, LmtTrackingKey, ConsentKey, VastCRTIDKey, LineIDKey, TimestampKey, AuctionIDKey, ChannelKey}},
			want: map[string]string{"PBS_ACCOUNTID": "testpublisherID", "PBS_APPBUNDLE": "testdomain", "PBS_BIDID": "bidId123", "PBS_GDPRCONSENT": "yes", "PBS_LIMITADTRACKING": "10", "PBS_PAGEURL": "pageurltest", "PBS_PUBDOMAIN": "publishertestdomain",
				"PBS-VASTCRTID": "creative_1", "PBS-LINEID": "campaign_1", "PBS-TIMESTAMP": strconv.Itoa(int(time.Now().Unix())), "PBS-AUCTIONID": "123", "PBS-CHANNEL": ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			macroProvider := NewProvider(req)
			macroProvider.SetContext(bid, nil, "test")
			if got := macroProvider.GetAllMacros(tt.args.keys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("macroProvider.GetAllMacros() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_macroProvider_GetMacro(t *testing.T) {
	type args struct {
		keys []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "get all macros success",
			args: args{keys: []string{BidIDKey, AccountIDKey, AppBundleKey, PubDomainkey,
				PageURLKey, AccountIDKey, LmtTrackingKey, ConsentKey}},
			want: map[string]string{"PBS_ACCOUNTID": "testpublisherID", "PBS_APPBUNDLE": "testdomain", "PBS_BIDID": "bidId123", "PBS_GDPRCONSENT": "yes", "PBS_LIMITADTRACKING": "10", "PBS_PAGEURL": "pageurltest", "PBS_PUBDOMAIN": "publishertestdomain"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			macroProvider := NewProvider(req)
			macroProvider.SetContext(bid, nil, "test")
			for _, key := range tt.args.keys {
				if got := macroProvider.GetMacro(key); got != tt.want[key] {
					t.Errorf("macroProvider.GetMacro() = %v, want %v", got, tt.want[key])
				}
			}
		})
	}
}
