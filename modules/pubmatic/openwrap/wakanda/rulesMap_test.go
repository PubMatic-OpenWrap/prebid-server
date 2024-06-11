package wakanda

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddRule(t *testing.T) {
	rm := getNewRulesMap(Wakanda{})

	assert.Equal(t, true, rm.AddIfNotPresent("FIRST_RULE", 2, "local"), "A non-existing rule should be added. Failed to add a non-existing rule.")
	assert.Equal(t, false, rm.AddIfNotPresent("FIRST_RULE", 2, "local"), "An existing rule should not be added.")
	assert.Equal(t, true, rm.IsRulePresent("FIRST_RULE"), "Should have returned true as rule is present")
	assert.Equal(t, false, rm.IsRulePresent("SECOND_RULE"), "Should have returned false as rule is NOT present")

	wr := rm.Incr("FIRST_RULE")
	assert.Equal(t, 1, wr.TraceCount, "TraceCount Should have been 0 TraceCount=%v", wr.TraceCount)
	assert.Equal(t, 1, wr.TraceCount, "DebugLevel should have been 2")
	assert.Equal(t, "local"+"__FIRST_RULE", wr.FolderPath, "FolderPath formation is not as expected")
}

func TestRulesMapCleanRules(t *testing.T) {
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

func TestRulesMapIncr(t *testing.T) {
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
						TraceCount: cMaxTraceCount,
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
