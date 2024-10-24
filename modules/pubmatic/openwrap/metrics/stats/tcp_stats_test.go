package stats

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/stretchr/testify/assert"
)

func TestInitTCPStatsClient(t *testing.T) {

	type args struct {
		endpoint string
		pubInterval, pubThreshold, retries, dialTimeout,
		keepAliveDur, maxIdleConn, maxIdleConnPerHost, respHeaderTimeout,
		maxChannelLength, poolMaxWorkers, poolMaxCapacity int
	}

	type want struct {
		client *StatsTCP
		err    error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "returns_error",
			args: args{
				endpoint:           "",
				pubInterval:        10,
				pubThreshold:       10,
				retries:            3,
				dialTimeout:        10,
				keepAliveDur:       10,
				maxIdleConn:        10,
				maxIdleConnPerHost: 10,
			},
			want: want{
				client: nil,
				err:    fmt.Errorf("invalid stats client configurations:stat server endpoint cannot be empty"),
			},
		},
		{
			name: "returns_valid_client",
			args: args{
				endpoint:           "http://10.10.10.10:8000/stat",
				pubInterval:        10,
				pubThreshold:       10,
				retries:            3,
				dialTimeout:        10,
				keepAliveDur:       10,
				maxIdleConn:        10,
				maxIdleConnPerHost: 10,
			},
			want: want{
				client: &StatsTCP{
					statsClient: &Client{
						endpoint: "http://10.10.10.10:8000/stat",
						httpClient: &http.Client{
							Transport: &http.Transport{
								DialContext: (&net.Dialer{
									Timeout:   10 * time.Second,
									KeepAlive: 10 * time.Minute,
								}).DialContext,
								MaxIdleConns:          10,
								MaxIdleConnsPerHost:   10,
								ResponseHeaderTimeout: 30 * time.Second,
							},
						},
						config: &config{
							Endpoint:              "http://10.10.10.10:8000/stat",
							PublishingInterval:    5,
							PublishingThreshold:   1000,
							Retries:               3,
							DialTimeout:           10,
							KeepAliveDuration:     15,
							MaxIdleConns:          10,
							MaxIdleConnsPerHost:   10,
							retryInterval:         100,
							MaxChannelLength:      1000,
							ResponseHeaderTimeout: 30,
							PoolMaxWorkers:        minPoolWorker,
							PoolMaxCapacity:       minPoolCapacity,
						},
						pubChan: make(chan stat, 1000),
						statMap: map[string]int{},
					},
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := initTCPStatsClient(tt.args.endpoint,
				tt.args.pubInterval, tt.args.pubThreshold, tt.args.retries, tt.args.dialTimeout, tt.args.keepAliveDur,
				tt.args.maxIdleConn, tt.args.maxIdleConnPerHost, tt.args.respHeaderTimeout, tt.args.maxChannelLength,
				tt.args.poolMaxWorkers, tt.args.poolMaxCapacity)

			assert.Equal(t, tt.want.err, err)
			if err == nil {
				compareClient(tt.want.client.statsClient, client.statsClient, t)
			}
		})
	}
}

func TestRecordFunctions(t *testing.T) {

	initStatKeys("N:P", "N:P")

	type args struct {
		statTCP *StatsTCP
	}

	type want struct {
		expectedkeyVal map[string]int
		channelSize    int
	}

	tests := []struct {
		name       string
		args       args
		want       want
		callRecord func(*StatsTCP)
	}{
		{
			name: "RecordOpenWrapServerPanicStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					statKeys[statsKeyOpenWrapServerPanic]: 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordOpenWrapServerPanicStats("", "")
			},
		},
		{
			name: "RecordPublisherPartnerNoCookieStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyPublisherPartnerNoCookieRequests], "5890", "pubmatic"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPublisherPartnerNoCookieStats("5890", "pubmatic")
			},
		},
		{
			name: "RecordPartnerTimeoutErrorStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyPartnerTimeoutErrorRequests], "5890", "pubmatic"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPartnerResponseErrors("5890", "pubmatic", models.PartnerErrTimeout)
			},
		},
		{
			name: "RecordNobidErrorStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyNobidErrorRequests], "5890", "pubmatic"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPartnerResponseErrors("5890", "pubmatic", models.PartnerErrNoBid)
			},
		},
		{
			name: "RecordUnkownPrebidErrorStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyUnknownPrebidErrorResponse], "5890", "pubmatic"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPartnerResponseErrors("5890", "pubmatic", models.PartnerErrUnknownPrebidError)
			},
		},
		{
			name: "RecordSlotNotMappedErrorStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeySlotunMappedErrorRequests], "5890", "pubmatic"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPartnerConfigErrors("5890", "1234", "pubmatic", models.PartnerErrSlotNotMapped)
			},
		},
		{
			name: "RecordMisConfigurationErrorStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyMisConfErrorRequests], "5890", "pubmatic"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPartnerConfigErrors("5890", "1234", "pubmatic", models.PartnerErrMisConfig)
			},
		},
		{
			name: "RecordPublisherProfileRequests",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyPublisherProfileRequests], "5890", "pubmatic"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPublisherProfileRequests("5890", "pubmatic")
			},
		},
		{
			name: "RecordPublisherInvalidProfileRequests",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 3),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyPublisherInvProfileVideoRequests], "5890", "pubmatic"): 1,
					fmt.Sprintf(statKeys[statsKeyPublisherInvProfileAMPRequests], "5890", "pubmatic"):   1,
					fmt.Sprintf(statKeys[statsKeyPublisherInvProfileRequests], "5890", "pubmatic"):      1,
				},
				channelSize: 3,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPublisherInvalidProfileRequests("video", "5890", "pubmatic")
				st.RecordPublisherInvalidProfileRequests("amp", "5890", "pubmatic")
				st.RecordPublisherInvalidProfileRequests("", "5890", "pubmatic")
			},
		},
		{
			name: "RecordNobidErrPrebidServerRequests",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyNobidErrPrebidServerRequests], "5890"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordNobidErrPrebidServerRequests("5890", 0)
			},
		},
		{
			name: "RecordNobidErrPrebidServerResponse",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyNobidErrPrebidServerResponse], "5890"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordNobidErrPrebidServerResponse("5890")
			},
		},
		{
			name: "RecordInvalidCreativeStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyInvalidCreatives], "5890", "pubmatic"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordInvalidCreativeStats("5890", "pubmatic")
			},
		},
		{
			name: "RecordPlatformPublisherPartnerReqStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyPlatformPublisherPartnerRequests], "web", "5890", "pubmatic"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPlatformPublisherPartnerReqStats("web", "5890", "pubmatic")
			},
		},
		{
			name: "RecordPlatformPublisherPartnerResponseStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyPlatformPublisherPartnerResponses], "web", "5890", "pubmatic"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPlatformPublisherPartnerResponseStats("web", "5890", "pubmatic")
			},
		},
		{
			name: "RecordPublisherResponseEncodingErrorStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyPublisherResponseEncodingErrors], "5890"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPublisherResponseEncodingErrorStats("5890")
			},
		},
		{
			name: "RecordPartnerResponseTimeStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 20),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyL50], "5890", "pubmatic"):   1,
					fmt.Sprintf(statKeys[statsKeyA50], "5890", "pubmatic"):   1,
					fmt.Sprintf(statKeys[statsKeyA100], "5890", "pubmatic"):  1,
					fmt.Sprintf(statKeys[statsKeyA200], "5890", "pubmatic"):  1,
					fmt.Sprintf(statKeys[statsKeyA300], "5890", "pubmatic"):  1,
					fmt.Sprintf(statKeys[statsKeyA400], "5890", "pubmatic"):  1,
					fmt.Sprintf(statKeys[statsKeyA500], "5890", "pubmatic"):  1,
					fmt.Sprintf(statKeys[statsKeyA600], "5890", "pubmatic"):  1,
					fmt.Sprintf(statKeys[statsKeyA700], "5890", "pubmatic"):  1,
					fmt.Sprintf(statKeys[statsKeyA800], "5890", "pubmatic"):  1,
					fmt.Sprintf(statKeys[statsKeyA900], "5890", "pubmatic"):  1,
					fmt.Sprintf(statKeys[statsKeyA1000], "5890", "pubmatic"): 1,
					fmt.Sprintf(statKeys[statsKeyA1500], "5890", "pubmatic"): 1,
					fmt.Sprintf(statKeys[statsKeyA2000], "5890", "pubmatic"): 1,
				},
				channelSize: 14,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 10)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 60)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 110)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 210)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 310)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 410)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 510)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 610)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 710)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 810)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 910)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 1010)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 1510)
				st.RecordPartnerResponseTimeStats("5890", "pubmatic", 2010)
			},
		},
		{
			name: "RecordPublisherResponseTimeStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 20),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyL50], "5890", "overall"):   1,
					fmt.Sprintf(statKeys[statsKeyA50], "5890", "overall"):   1,
					fmt.Sprintf(statKeys[statsKeyA100], "5890", "overall"):  1,
					fmt.Sprintf(statKeys[statsKeyA200], "5890", "overall"):  1,
					fmt.Sprintf(statKeys[statsKeyA300], "5890", "overall"):  1,
					fmt.Sprintf(statKeys[statsKeyA400], "5890", "overall"):  1,
					fmt.Sprintf(statKeys[statsKeyA500], "5890", "overall"):  1,
					fmt.Sprintf(statKeys[statsKeyA600], "5890", "overall"):  1,
					fmt.Sprintf(statKeys[statsKeyA700], "5890", "overall"):  1,
					fmt.Sprintf(statKeys[statsKeyA800], "5890", "overall"):  1,
					fmt.Sprintf(statKeys[statsKeyA900], "5890", "overall"):  1,
					fmt.Sprintf(statKeys[statsKeyA1000], "5890", "overall"): 1,
					fmt.Sprintf(statKeys[statsKeyA1500], "5890", "overall"): 1,
					fmt.Sprintf(statKeys[statsKeyA2000], "5890", "overall"): 1,
				},
				channelSize: 14,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPublisherResponseTimeStats("5890", 10)
				st.RecordPublisherResponseTimeStats("5890", 60)
				st.RecordPublisherResponseTimeStats("5890", 110)
				st.RecordPublisherResponseTimeStats("5890", 210)
				st.RecordPublisherResponseTimeStats("5890", 310)
				st.RecordPublisherResponseTimeStats("5890", 410)
				st.RecordPublisherResponseTimeStats("5890", 510)
				st.RecordPublisherResponseTimeStats("5890", 610)
				st.RecordPublisherResponseTimeStats("5890", 710)
				st.RecordPublisherResponseTimeStats("5890", 810)
				st.RecordPublisherResponseTimeStats("5890", 910)
				st.RecordPublisherResponseTimeStats("5890", 1010)
				st.RecordPublisherResponseTimeStats("5890", 1510)
				st.RecordPublisherResponseTimeStats("5890", 2010)
			},
		},
		{
			name: "RecordPublisherWrapperLoggerFailure",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyLoggerErrorRequests], "5890", "1234", "0"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPublisherWrapperLoggerFailure("5890", "1234", "0")
			},
		},
		{
			name: "RecordPrebidTimeoutRequests",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyPrebidTORequests], "5890", "1234"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPrebidTimeoutRequests("5890", "1234")
			},
		},
		{
			name: "RecordSSTimeoutRequests",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeySsTORequests], "5890", "1234"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordSSTimeoutRequests("5890", "1234")
			},
		},
		{
			name: "RecordUidsCookieNotPresentErrorStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyNoUIDSErrorRequest], "5890", "1234"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordUidsCookieNotPresentErrorStats("5890", "1234")
			},
		},
		{
			name: "RecordVideoInstlImpsStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyVideoInterstitialImpressions], "5890", "1234"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordVideoInstlImpsStats("5890", "1234")
			},
		},
		{
			name: "RecordImpDisabledViaConfigStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 2),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyVideoImpDisabledViaConfig], "5890", "1234"):  1,
					fmt.Sprintf(statKeys[statsKeyBannerImpDisabledViaConfig], "5890", "1234"): 1,
				},
				channelSize: 2,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordImpDisabledViaConfigStats("video", "5890", "1234")
				st.RecordImpDisabledViaConfigStats("banner", "5890", "1234")
			},
		},
		{
			name: "RecordVideoImpDisabledViaConnTypeStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 2),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyVideoImpDisabledViaConnType], "5890", "1234"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordVideoImpDisabledViaConnTypeStats("5890", "1234")
			},
		},
		{
			name: "RecordPreProcessingTimeStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 5),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyPrTimeAbv100], "5890"): 1,
					fmt.Sprintf(statKeys[statsKeyPrTimeAbv50], "5890"):  1,
					fmt.Sprintf(statKeys[statsKeyPrTimeAbv10], "5890"):  1,
					fmt.Sprintf(statKeys[statsKeyPrTimeAbv1], "5890"):   1,
					fmt.Sprintf(statKeys[statsKeyPrTimeBlw1], "5890"):   1,
				},
				channelSize: 5,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPreProcessingTimeStats("5890", 0)
				st.RecordPreProcessingTimeStats("5890", 5)
				st.RecordPreProcessingTimeStats("5890", 15)
				st.RecordPreProcessingTimeStats("5890", 75)
				st.RecordPreProcessingTimeStats("5890", 105)
			},
		},
		{
			name: "RecordStatsKeyCTVPrebidFailedImpression",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 5),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyCTVPrebidFailedImpression], 1, "5890", "1234"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordStatsKeyCTVPrebidFailedImpression(1, "5890", "1234")
			},
		},
		{
			name: "RecordCTVRequests",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 5),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyCTVRequests], "5890", "web"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordCTVRequests("5890", "web")
			},
		},
		{
			name: "RecordBadRequests",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 7),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyAMPBadRequests]):                  1,
					fmt.Sprintf(statKeys[statsKeyVideoBadRequests]):                1,
					fmt.Sprintf(statKeys[statsKey25BadRequests]):                   1,
					fmt.Sprintf(statKeys[statsKeyCTVBadRequests], "json", 100):     1,
					fmt.Sprintf(statKeys[statsKeyCTVBadRequests], "openwrap", 200): 1,
					fmt.Sprintf(statKeys[statsKeyCTVBadRequests], "ortb", 300):     1,
					fmt.Sprintf(statKeys[statsKeyCTVBadRequests], "vast", 400):     1,
				},
				channelSize: 7,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordBadRequests("amp", 1)
				st.RecordBadRequests("video", 1)
				st.RecordBadRequests("v25", 1)
				st.RecordBadRequests("json", 100)
				st.RecordBadRequests("openwrap", 200)
				st.RecordBadRequests("ortb", 300)
				st.RecordBadRequests("vast", 400)
			},
		},
		{
			name: "RecordCTVHTTPMethodRequests",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyCTVHTTPMethodRequests], "ortb", "5890", "GET"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordCTVHTTPMethodRequests("ortb", "5890", "GET")
			},
		},
		{
			name: "RecordCTVInvalidReasonCount",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 5),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyCTVValidationErr], 100, "5890"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordCTVInvalidReasonCount(100, "5890")
			},
		},
		{
			name: "RecordCTVReqImpsWithDbConfigCount",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 5),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyCTVReqImpstWithConfig], "db", "5890"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordCTVReqImpsWithDbConfigCount("5890")
			},
		},
		{
			name: "RecordCTVReqImpsWithReqConfigCount",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 5),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyCTVReqImpstWithConfig], "req", "5890"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordCTVReqImpsWithReqConfigCount("5890")
			},
		},
		{
			name: "RecordAdPodGeneratedImpressionsCount",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 4),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyTotalAdPodImpression], "1-3", "5890"): 1,
					fmt.Sprintf(statKeys[statsKeyTotalAdPodImpression], "4-6", "5890"): 1,
					fmt.Sprintf(statKeys[statsKeyTotalAdPodImpression], "7-9", "5890"): 1,
					fmt.Sprintf(statKeys[statsKeyTotalAdPodImpression], "9+", "5890"):  1,
				},
				channelSize: 4,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordAdPodGeneratedImpressionsCount(3, "5890")
				st.RecordAdPodGeneratedImpressionsCount(6, "5890")
				st.RecordAdPodGeneratedImpressionsCount(9, "5890")
				st.RecordAdPodGeneratedImpressionsCount(11, "5890")
			},
		},
		{
			name: "RecordRequestAdPodGeneratedImpressionsCount",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyReqTotalAdPodImpression], "5890"): 2,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordRequestAdPodGeneratedImpressionsCount(2, "5890")
			},
		},
		{
			name: "RecordReqImpsWithAppContentCount",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyContentObjectPresent], "app", "5890"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordReqImpsWithContentCount("5890", models.ContentTypeApp)
			},
		},
		{
			name: "RecordReqImpsWithSiteContentCount",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyContentObjectPresent], "site", "5890"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordReqImpsWithContentCount("5890", models.ContentTypeSite)
			},
		},
		{
			name: "RecordAdPodImpressionYield",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyReqImpDurationYield], 10, 1, "5890"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordAdPodImpressionYield(10, 1, "5890")
			},
		},
		{
			name: "RecordCTVReqCountWithAdPod",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyReqWithAdPodCount], "5890", "1234"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordCTVReqCountWithAdPod("5890", "1234")
			},
		},
		{
			name: "RecordPBSAuctionRequestsStats",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyPBSAuctionRequests]): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPBSAuctionRequestsStats()
			},
		},
		{
			name: "RecordInjectTrackerErrorCount",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyInjectTrackerErrorCount], "banner", "5890", "1234"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordInjectTrackerErrorCount("banner", "5890", "1234")
			},
		},
		{
			name: "RecordBidResponseByDealCountInPBS",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsBidResponsesByDealUsingPBS], "5890", "1234", "pubmatic", "pubdeal"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordBidResponseByDealCountInPBS("5890", "1234", "pubmatic", "pubdeal")
			},
		},
		{
			name: "RecordBidResponseByDealCountInHB",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsBidResponsesByDealUsingHB], "5890", "1234", "pubmatic", "pubdeal"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordBidResponseByDealCountInHB("5890", "1234", "pubmatic", "pubdeal")
			},
		},
		{
			name: "RecordPartnerTimeoutInPBS",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 1),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsPartnerTimeoutInPBS], "5890", "1234", "pubmatic"): 1,
				},
				channelSize: 1,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPartnerTimeoutInPBS("5890", "1234", "pubmatic")
			},
		},
		{
			name: "RecordPublisherRequests",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 6),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyAMPPublisherRequests], "5890"):                    1,
					fmt.Sprintf(statKeys[statsKeyVideoPublisherRequests], "5890"):                  1,
					fmt.Sprintf(statKeys[statsKey25PublisherRequests], "app", "5890"):              1,
					fmt.Sprintf(statKeys[statsKeyCTVPublisherRequests], "ortb", "app", "5890"):     1,
					fmt.Sprintf(statKeys[statsKeyCTVPublisherRequests], "json", "app", "5890"):     1,
					fmt.Sprintf(statKeys[statsKeyCTVPublisherRequests], "vast", "display", "5890"): 1,
				},
				channelSize: 6,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordPublisherRequests("amp", "5890", "")
				st.RecordPublisherRequests("video", "5890", "")
				st.RecordPublisherRequests("v25", "5890", "in-app")
				st.RecordPublisherRequests("ortb", "5890", "in-app")
				st.RecordPublisherRequests("json", "5890", "app")
				st.RecordPublisherRequests("vast", "5890", "display")
			},
		},
		{
			name: "RecordCacheErrorRequests",
			args: args{
				statTCP: &StatsTCP{
					&Client{
						pubChan: make(chan stat, 2),
					},
				},
			},
			want: want{
				expectedkeyVal: map[string]int{
					fmt.Sprintf(statKeys[statsKeyAMPCacheError], "5890", "1234"):   1,
					fmt.Sprintf(statKeys[statsKeyVideoCacheError], "5890", "1234"): 1,
				},
				channelSize: 2,
			},
			callRecord: func(st *StatsTCP) {
				st.RecordCacheErrorRequests("amp", "5890", "1234")
				st.RecordCacheErrorRequests("video", "5890", "1234")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.callRecord(tt.args.statTCP)

			close(tt.args.statTCP.statsClient.pubChan)
			assert.Equal(t, tt.want.channelSize, len(tt.args.statTCP.statsClient.pubChan))
			for stat := range tt.args.statTCP.statsClient.pubChan {
				assert.Equalf(t, tt.want.expectedkeyVal[stat.Key], stat.Value,
					"Mismatched value for key [%s]", stat.Key)
			}
		})
	}
}
