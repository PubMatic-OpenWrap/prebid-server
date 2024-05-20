package wakanda

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddRule(t *testing.T) {
	rm := getNewRulesMap(Wakanda{})

	if rm.AddIfNotPresent("FIRST_RULE", 2, "local") == false {
		t.Error("A non-existing rule should be added. Failed to add a non-existing rule.")
	}

	if rm.AddIfNotPresent("FIRST_RULE", 2, "local") == true {
		t.Error("An existing rule should not be added.")
	}

	if rm.IsRulePresent("FIRST_RULE") == false {
		t.Error("Should have returned true as rule is present")
	}

	if rm.IsRulePresent("SECOND_RULE") == true {
		t.Error("Should have returned false as rule is NOT present")
	}

	wr := rm.Incr("FIRST_RULE")
	if wr.TraceCount != 1 {
		t.Errorf("TraceCount Should have been 0 TraceCount=%v", wr.TraceCount)
	}
	if wr.DebugLevel != 2 {
		t.Error("DebugLevel should have been 2")
	}

	if wr.FolderPath != "local"+"__FIRST_RULE" {
		t.Error("FolderPath formation is not as expected")
	}
}

// func init() {
// 	config.ServerConfig = &config.DMHBConfig{}
// 	config.ServerConfig.OpenWrap.Server.DCName = "local"
// }

func Test_rulesMap_cleanRules(t *testing.T) {
	now := time.Now()

	maxDuration := time.Minute * 10
	tests := []struct {
		name string
		args map[string]*wakandaRule
		want map[string]*wakandaRule
	}{
		{
			name: "EmptyMap",
			args: map[string]*wakandaRule{},
			want: map[string]*wakandaRule{},
		},
		{
			name: "Mixed",
			args: map[string]*wakandaRule{
				"just_added": {
					StartTime: now,
				},
				"in_between": {
					StartTime: now.Add(-maxDuration + maxDuration/2),
				},
				"stale": {
					StartTime: now.Add(-maxDuration - 1),
				},
			},
			want: map[string]*wakandaRule{
				"just_added": {
					StartTime: now,
				},
				"in_between": {
					StartTime: now.Add(-maxDuration + maxDuration/2),
				},
			},
		},
		{
			name: "AllStale",
			args: map[string]*wakandaRule{
				"stale1": {
					StartTime: now.Add(-maxDuration - 1),
				},
				"stale2": {
					StartTime: now.Add(-maxDuration - 2),
				},
				"stale3": {
					StartTime: now.Add(-maxDuration - 3),
				},
			},
			want: map[string]*wakandaRule{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := &rulesMap{
				rules: tt.args,
			}
			rm.cleanRules(maxDuration)
			assert.Equal(t, tt.want, rm.rules)
		})
	}
}

func Test_rulesMap_Incr(t *testing.T) {
	type fields struct {
		rules map[string]*wakandaRule
		lock  sync.RWMutex
	}
	type args struct {
		key string
	}
	type want struct {
		rules    map[string]*wakandaRule
		wantRule *wakandaRule
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name:   "key_not_present",
			fields: fields{},
			args: args{
				key: `key1`,
			},
			want: want{},
		},
		{
			name: "key_present",
			fields: fields{
				rules: map[string]*wakandaRule{
					`key1`: {
						TraceCount: 1,
					},
				},
			},
			args: args{
				key: `key1`,
			},
			want: want{
				rules: map[string]*wakandaRule{
					`key1`: {
						TraceCount: 2,
					},
				},
				wantRule: &wakandaRule{
					TraceCount: 2,
				},
			},
		},
		{
			name: "key_expired",
			fields: fields{
				rules: map[string]*wakandaRule{
					`key1`: {
						TraceCount: CMaxTraceCount,
					},
				},
			},
			args: args{
				key: `key1`,
			},
			want: want{
				rules: map[string]*wakandaRule{},
			},
		},
	}
	for ind := range tests {
		tt := &tests[ind]
		t.Run(tt.name, func(t *testing.T) {
			rm := &rulesMap{
				rules: tt.fields.rules,
				lock:  sync.RWMutex{},
			}
			got := rm.Incr(tt.args.key)
			assert.Equal(t, tt.want.wantRule, got)
			assert.Equal(t, tt.want.rules, rm.rules)
		})
	}
}
