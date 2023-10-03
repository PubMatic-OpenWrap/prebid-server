package openwrap

import (
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestCheckABTestEnabled(t *testing.T) {
	type args struct {
		rctx models.RequestCtx
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "AbTest_enabled_in_partner_config_abTestEnabled=1",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.AbTestEnabled: "1",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "AbTest_is_not_enabled_in_partner_config_abTestEnabled_is_other_than_1",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.AbTestEnabled: "0",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "AbTest_is_not_enabled_in_partner_config_abTestEnabled_is_empty",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.AbTestEnabled: "",
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckABTestEnabled(tt.args.rctx); got != tt.want {
				t.Errorf("CheckABTestEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestABTestProcessing(t *testing.T) {
	type args struct {
		rctx models.RequestCtx
		val  int
	}
	tests := []struct {
		name  string
		args  args
		want  map[int]map[string]string
		want1 bool
	}{
		{
			name: "AbTest_enabled_but_random_no_return_do_not_apply_Abtest",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.AbTestEnabled:           "1",
							models.TestGroupSize + "_test": "0",
						},
					},
				},
			},
			want:  nil,
			want1: false,
		},
		{
			name: "AbTest_is_disabled",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.AbTestEnabled:           "0",
							models.TestGroupSize + "_test": "0",
						},
					},
				},
			},
			want:  nil,
			want1: false,
		},
		{
			name: "AbTest_is_enabled_and_random_no_return_apply_AbTest",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							models.AbTestEnabled:           "1",
							models.TestType + "_test":      models.TestTypeAuctionTimeout,
							models.SSTimeoutKey + "_test":  "350",
							models.TestGroupSize + "_test": "90",
							models.SSTimeoutKey:            "100",
						},
					},
				},
				val: 50,
			},
			want: map[int]map[string]string{
				-1: {
					models.AbTestEnabled:           "1",
					models.TestType + "_test":      models.TestTypeAuctionTimeout,
					models.SSTimeoutKey + "_test":  "350",
					models.TestGroupSize + "_test": "90",
					models.SSTimeoutKey:            "350",
				},
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetRandomNumberIn1To100 = func() int {
				return tt.args.val
			}
			got, got1 := ABTestProcessing(tt.args.rctx)
			assert.Equal(t, tt.want, got)
			if got1 != tt.want1 {
				t.Errorf("ABTestProcessing() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestApplyTestConfig(t *testing.T) {
	type args struct {
		rctx models.RequestCtx
		val  int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "testGroupSize_is_zero_in_partner_config",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"testGroupSize_test": "0",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "testGroupSize_in_partner_config_is_greater_than_random_number_generated",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"testGroupSize_test": "60",
						},
					},
				},
				val: 20,
			},
			want: true,
		},
		{
			name: "testGroupSize_in_partner_config_is_equal_to_random_number_generated",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"testGroupSize_test": "60",
						},
					},
				},
				val: 60,
			},
			want: true,
		},
		{
			name: "testGroupSize_in_partner_config_is_less_than_random_number_generated",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							"testGroupSize_test": "20",
						},
					},
				},
				val: 60,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetRandomNumberIn1To100 = func() int {
				return tt.args.val
			}
			if got := ApplyTestConfig(tt.args.rctx); got != tt.want {
				t.Errorf("ApplyTestConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppendTest(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				key: models.AbTestEnabled,
			},
			want: models.AbTestEnabled + "_test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AppendTest(tt.args.key); got != tt.want {
				t.Errorf("AppendTest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateTestConfig(t *testing.T) {
	type args struct {
		rctx models.RequestCtx
	}
	tests := []struct {
		name string
		args args
		want map[int]map[string]string
	}{
		{
			name: "testype_is_Auction_Timeout",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							AppendTest(models.TestType):     models.TestTypeAuctionTimeout,
							AppendTest(models.SSTimeoutKey): "350",
							models.SSTimeoutKey:             "100",
						},
					},
				},
			},

			want: map[int]map[string]string{
				-1: {
					AppendTest(models.TestType):     models.TestTypeAuctionTimeout,
					AppendTest(models.SSTimeoutKey): "350",
					models.SSTimeoutKey:             "350",
				},
			},
		},
		{
			name: "testype_is_partners",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							AppendTest(models.TestType): models.TestTypePartners,
							models.SSTimeoutKey:         "300",
						},
						123: {
							"adapterId":                  "201",
							"adapterName":                "testAdapter",
							"partnerId":                  "101",
							"partnerName":                "testPartner",
							"prebidPartnerName":          "testPartner",
							models.PartnerTestEnabledKey: "1",
							"accountId_test":             "1234",
							"pubId_test":                 "8888",
							"rev_share_test":             "10",
							"throttle_test":              "100",
							"serverSideEnabled_test":     "1",
						},
						234: {
							"adapterId":         "202",
							"adapterName":       "SecondAdapter",
							"partnerId":         "102",
							"partnerName":       "controlPartner",
							"prebidPartnerName": "controlPartner",
							"rev_share":         "10",
							"throttle":          "100",
							"serverSideEnabled": "1",
						},
					},
				},
			},

			want: map[int]map[string]string{
				-1: {
					AppendTest(models.TestType): models.TestTypePartners,
					models.SSTimeoutKey:         "300",
				},
				123: {
					"adapterId":                  "201",
					"adapterName":                "testAdapter",
					"partnerId":                  "101",
					"partnerName":                "testPartner",
					"prebidPartnerName":          "testPartner",
					models.PartnerTestEnabledKey: "1",
					"accountId_test":             "1234",
					"pubId_test":                 "8888",
					"rev_share_test":             "10",
					"throttle_test":              "100",
					"serverSideEnabled_test":     "1",
					"rev_share":                  "10",
					"throttle":                   "100",
					"serverSideEnabled":          "1",
					"accountId":                  "1234",
					"pubId":                      "8888",
				},
			},
		},

		{
			name: "testype_is_client_side_vs._server_side_path",
			args: args{
				rctx: models.RequestCtx{
					PartnerConfigMap: map[int]map[string]string{
						-1: {
							AppendTest(models.TestType): models.TestTypeClientVsServerPath,
							models.SSTimeoutKey:         "300",
						},
						123: {
							"adapterId":                  "201",
							"adapterName":                "testAdapter",
							"partnerId":                  "101",
							"partnerName":                "testPartner",
							"prebidPartnerName":          "testPartner",
							models.PartnerTestEnabledKey: "1",
							"pubId_test":                 "8888",
							"rev_share":                  "10",
							"throttle":                   "100",
							"serverSideEnabled":          "0",
							"rev_share_test":             "10",
							"throttle_test":              "100",
							"serverSideEnabled_test":     "1",
						},
						234: {
							"adapterId":              "202",
							"adapterName":            "SecondAdapter",
							"partnerId":              "102",
							"partnerName":            "controlPartner",
							"prebidPartnerName":      "controlPartner",
							"rev_share":              "10",
							"throttle":               "100",
							"serverSideEnabled":      "1",
							"rev_share_test":         "10",
							"throttle_test":          "100",
							"serverSideEnabled_test": "0",
						},
					},
				},
			},

			want: map[int]map[string]string{
				-1: {
					AppendTest(models.TestType): models.TestTypeClientVsServerPath,
					models.SSTimeoutKey:         "300",
				},
				123: {
					"adapterId":                  "201",
					"adapterName":                "testAdapter",
					"partnerId":                  "101",
					"partnerName":                "testPartner",
					"prebidPartnerName":          "testPartner",
					models.PartnerTestEnabledKey: "1",
					"pubId_test":                 "8888",
					"rev_share":                  "10",
					"throttle":                   "100",
					"serverSideEnabled":          "1",
					"rev_share_test":             "10",
					"throttle_test":              "100",
					"serverSideEnabled_test":     "1",
				},
				234: {
					"adapterId":              "202",
					"adapterName":            "SecondAdapter",
					"partnerId":              "102",
					"partnerName":            "controlPartner",
					"prebidPartnerName":      "controlPartner",
					"rev_share":              "10",
					"throttle":               "100",
					"serverSideEnabled":      "0",
					"rev_share_test":         "10",
					"throttle_test":          "100",
					"serverSideEnabled_test": "0",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpdateTestConfig(tt.args.rctx)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_copyPartnerConfigMap(t *testing.T) {
	type args struct {
		m map[int]map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[int]map[string]string
	}{
		{
			name: "test",
			args: args{
				m: map[int]map[string]string{
					123: {
						"adapterId":         "201",
						"adapterName":       "testAdapter",
						"partnerId":         "101",
						"partnerName":       "testPartner",
						"prebidPartnerName": "testPartner",
						"accountId":         "1234",
						"pubId":             "8888",
					},
				},
			},
			want: map[int]map[string]string{
				123: {
					"adapterId":         "201",
					"adapterName":       "testAdapter",
					"partnerId":         "101",
					"partnerName":       "testPartner",
					"prebidPartnerName": "testPartner",
					"accountId":         "1234",
					"pubId":             "8888",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := copyPartnerConfigMap(tt.args.m)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_replaceControlConfig(t *testing.T) {
	type args struct {
		partnerConfig map[int]map[string]string
		partnerID     int
		key           string
	}
	tests := []struct {
		name string
		args args
		want map[int]map[string]string
	}{
		{
			name: "testValue_is_present",
			args: args{
				partnerConfig: map[int]map[string]string{
					-1: {
						AppendTest(models.TestType):     models.TestTypeAuctionTimeout,
						AppendTest(models.SSTimeoutKey): "350",
						models.SSTimeoutKey:             "100",
					},
				},
				partnerID: models.VersionLevelConfigID,
				key:       models.SSTimeoutKey,
			},
			want: map[int]map[string]string{
				-1: {
					AppendTest(models.TestType):     models.TestTypeAuctionTimeout,
					AppendTest(models.SSTimeoutKey): "350",
					models.SSTimeoutKey:             "350",
				},
			},
		},
		{
			name: "testValue_is_not_present",
			args: args{
				partnerConfig: map[int]map[string]string{
					-1: {
						AppendTest(models.TestType): models.TestTypeAuctionTimeout,
						models.SSTimeoutKey:         "100",
					},
				},
				partnerID: models.VersionLevelConfigID,
				key:       models.SSTimeoutKey,
			},
			want: map[int]map[string]string{
				-1: {
					AppendTest(models.TestType): models.TestTypeAuctionTimeout,
					models.SSTimeoutKey:         "100",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replaceControlConfig(tt.args.partnerConfig, tt.args.partnerID, tt.args.key)
			assert.Equal(t, tt.want, tt.args.partnerConfig)
		})
	}
}

func Test_copyTestConfig(t *testing.T) {
	type args struct {
		partnerConfig map[int]map[string]string
		partnerID     int
		key           string
	}
	tests := []struct {
		name string
		args args
		want map[int]map[string]string
	}{
		{
			name: "key_has__test_suffix",
			args: args{
				partnerConfig: map[int]map[string]string{
					123: {
						"accountId_test": "1234",
					},
				},
				key:       "accountId_test",
				partnerID: 123,
			},
			want: map[int]map[string]string{
				123: {
					"accountId_test": "1234",
					"accountId":      "1234",
				},
			},
		},
		{
			name: "key_do_not_have__test_suffix",
			args: args{
				partnerConfig: map[int]map[string]string{
					123: {
						"accountId": "1234",
					},
				},
				key:       "accountId",
				partnerID: 123,
			},
			want: map[int]map[string]string{
				123: {
					"accountId": "1234",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copyTestConfig(tt.args.partnerConfig, tt.args.partnerID, tt.args.key)
			assert.Equal(t, tt.want, tt.args.partnerConfig)
		})
	}
}
