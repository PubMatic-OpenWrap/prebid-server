package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitStatKeys(t *testing.T) {

	type args struct {
		defaultServerName, actualServerName string
	}

	type want struct {
		testKeys [maxNumOfStats]string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "test_init_stat_keys",
			args: args{
				defaultServerName: "sv3:N:P",
				actualServerName:  "sv3:node123.sv3:ssheaderbidding",
			},
			want: want{
				testKeys: [maxNumOfStats]string{
					"hb:panic:sv3:node123.sv3:ssheaderbidding",
					"hb:pubnbreq:%s:sv3:N:P",
					"hb:pubnbres:%s:sv3:N:P",
					"hb:cnt:%s:%s:sv3:N:P",
					"hb:pprofrq:%s:%s:sv3:N:P",
					"hb:pubinp:%s:%s:sv3:N:P",
					"hb:pubinpimp:%s:%s:sv3:N:P",
					"hb:prebidto:%s:%s:sv3:N:P",
					"hb:ssto:%s:%s:sv3:N:P",
					"hb:nouids:%s:%s:sv3:N:P",
					"hb:ppvidinstlimps:%s:%s:sv3:N:P",
					"hb:ppdisimpcfg:%s:%s:sv3:N:P",
					"hb:ppdisimpct:%s:%s:sv3:N:P",
					"hb:ppnc:%s:%s:sv3:N:P",
					"hb:sler:%s:%s:sv3:N:P",
					"hb:cfer:%s:%s:sv3:N:P",
					"hb:toer:%s:%s:sv3:N:P",
					"hb:uner:%s:%s:sv3:N:P",
					"hb:nber:%s:%s:sv3:N:P",
					"hb:wle:%s:%s:%s:sv3:N:P",
					"hb:2.4:%s:pbrq:%s:sv3:N:P",
					"hb:2.5:badreq:sv3:N:P",
					"hb:2.5:%s:pbrq:%s:sv3:N:P",
					"hb:amp:badreq:sv3:N:P",
					"hb:amp:pbrq:%s:sv3:N:P",
					"hb:amp:ce:%s:%s:sv3:N:P",
					"hb:amp:pubinp:%s:%s:sv3:N:P",
					"hb:vid:badreq:sv3:N:P",
					"hb:vid:pbrq:%s:sv3:N:P",
					"hb:vid:ce:%s:%s:sv3:N:P",
					"hb:vid:pubinp:%s:%s:sv3:N:P",
					"hb:invcr:%s:%s:sv3:N:P",
					"hb:pppreq:%s:%s:%s:sv3:N:P",
					"hb:pppres:%s:%s:%s:sv3:N:P",
					"hb:encerr:%s:sv3:N:P",
					"hb:latabv_2000:%s:%s:sv3:N:P",
					"hb:latabv_1500:%s:%s:sv3:N:P",
					"hb:latabv_1000:%s:%s:sv3:N:P",
					"hb:latabv_900:%s:%s:sv3:N:P",
					"hb:latabv_800:%s:%s:sv3:N:P",
					"hb:latabv_700:%s:%s:sv3:N:P",
					"hb:latabv_600:%s:%s:sv3:N:P",
					"hb:latabv_500:%s:%s:sv3:N:P",
					"hb:latabv_400:%s:%s:sv3:N:P",
					"hb:latabv_300:%s:%s:sv3:N:P",
					"hb:latabv_200:%s:%s:sv3:N:P",
					"hb:latabv_100:%s:%s:sv3:N:P",
					"hb:latabv_50:%s:%s:sv3:N:P",
					"hb:latblw_50:%s:%s:sv3:N:P",
					"hb:ptabv_100:%s:sv3:N:P",
					"hb:ptabv_50:%s:sv3:N:P",
					"hb:ptabv_10:%s:sv3:N:P",
					"hb:ptabv_1:%s:sv3:N:P",
					"hb:ptblw_1:%s:sv3:N:P",
					"hb:bnrdiscfg:%s:%s:sv3:N:P",
					"hb:lfv:badimp:%v:%v:%v:sv3:N:P",
					"hb:lfv:%v:%v:req:sv3:N:P",
					"hb:lfv:%v:badreq:%d:sv3:N:P",
					"hb:lfv:%v:%v:pbrq:%v:sv3:N:P",
					"hb:lfv:%v:mtd:%v:%v:sv3:N:P",
					"hb:lfv:ivr:%d:%s:sv3:N:P",
					"hb:lfv:rwc:%s:%s:sv3:N:P",
					"hb:lfv:tpi:%s:%s:sv3:N:P",
					"hb:lfv:rtpi:%s:sv3:N:P",
					"hb:lfv:impy:%d:%d:%s:sv3:N:P",
					"hb:lfv:rwap:%s:%s:sv3:N:P",
					"hb:pbs:auc:sv3:N:P",
					"hb:mistrack:%s:%s:%s:sv3:N:P",
					"hb:pbs:dbc:%s:%s:%s:%s:sv3:N:P",
					"hb:dbc:%s:%s:%s:%s:sv3:N:P",
					"hb:pbs:pto:%s:%s:%s:sv3:N:P",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initStatKeys(tt.args.defaultServerName, tt.args.actualServerName)
			assert.Equal(t, statKeys, tt.want.testKeys)
		})
	}
}

func TestInitStat(t *testing.T) {

	type args struct {
		endpoint, defaultHost, actualHost, dcName string
		pubInterval, pubThreshold, retries, dialTimeout, keepAliveDuration,
		maxIdleConnes, maxIdleConnesPerHost, respHeaderTimeout, maxChannelLength,
		poolMaxWorkers, poolMaxCapacity int
	}

	type want struct {
		client *StatsTCP
		err    error
	}

	tests := []struct {
		name  string
		args  args
		want  want
		setup func() want
	}{
		{
			name: "singleton_instance",
			args: args{
				endpoint:             "10.10.10.10",
				defaultHost:          "N:P",
				actualHost:           "node1.sv3:ssheader",
				dcName:               "sv3",
				pubInterval:          10,
				pubThreshold:         10,
				retries:              3,
				dialTimeout:          10,
				keepAliveDuration:    10,
				maxIdleConnes:        10,
				maxIdleConnesPerHost: 10,
				respHeaderTimeout:    10,
				maxChannelLength:     10,
				poolMaxWorkers:       10,
				poolMaxCapacity:      10,
			},
			setup: func() want {
				st, err := InitStatsClient("10.10.10.10/stats", "N:P", "node1.sv3:ssheader", "sv3", 10, 10, 3, 10, 10, 10, 10, 10, 10, 10, 10)
				return want{client: st, err: err}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want = tt.setup()

			InitStatsClient(tt.args.endpoint, tt.args.defaultHost, tt.args.actualHost, tt.args.dcName,
				tt.args.pubInterval, tt.args.pubThreshold, tt.args.retries, tt.args.dialTimeout, tt.args.keepAliveDuration,
				tt.args.maxIdleConnes, tt.args.maxIdleConnesPerHost, tt.args.respHeaderTimeout,
				tt.args.maxChannelLength, tt.args.poolMaxWorkers, tt.args.poolMaxCapacity)

			assert.Equal(t, tt.want.client, owStats)
			assert.Equal(t, tt.want.err, owStatsErr)
		})
	}
}
