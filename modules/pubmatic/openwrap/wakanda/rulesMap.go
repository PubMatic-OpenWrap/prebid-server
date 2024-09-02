package wakanda

import (
	"sync"
	"time"

	"git.pubmatic.com/PubMatic/go-common/logger"
)

type wakandaRule struct {
	TraceCount int // how many request are logged
	FolderPath string
	DebugLevel int
	StartTime  time.Time
	Filters    string
}

type rulesMap struct {
	rules map[string]*wakandaRule
	lock  sync.RWMutex
}

// Incr function will increment trace count for each wakanda rule
func (rm *rulesMap) Incr(key string) *wakandaRule {
	rm.lock.Lock()
	aWakandaRule := rm.rules[key]
	defer rm.lock.Unlock()
	if aWakandaRule == nil {
		// other goroutine deleted the entry
		return nil
	}
	//below line can be moved outside of function
	aWakandaRule.TraceCount++
	if aWakandaRule.TraceCount > cMaxTraceCount {
		// this rule has got enough traces so we can delete this active rule
		delete(rm.rules, key)
		return nil
	}
	return aWakandaRule
}

// IsRulePresent function will check if rule is present in or not
func (rm *rulesMap) IsRulePresent(key string) bool {
	rm.lock.RLock()
	_, ok := rm.rules[key]
	rm.lock.RUnlock()
	return ok
}

// IsEmpty function will check if any rule is present in or not
func (rm *rulesMap) IsEmpty() bool {
	rm.lock.RLock()
	ok := len(rm.rules) == 0
	rm.lock.RUnlock()
	return ok
}

// AddIfNotPresent returns true if added; returns false if already present
func (rm *rulesMap) AddIfNotPresent(key string, debugLevel int, dcName string, filters string) bool {
	rm.lock.Lock()
	defer rm.lock.Unlock()
	aWakandaRule := rm.rules[key]
	if aWakandaRule != nil {
		//Already Present

		//Resetting time for particular rule
		aWakandaRule.StartTime = time.Now()
		return false
	}

	aWakandaRule = &wakandaRule{
		TraceCount: 0,
		FolderPath: dcName + "__" + key, // this should be in sync with UI
		DebugLevel: debugLevel,
		StartTime:  time.Now(),
		Filters:    filters,
	}
	rm.rules[key] = aWakandaRule
	return true
}

func (rm *rulesMap) clean(cleanupFrequencyInMin, MaxDurationInMin time.Duration) {
	c := time.Tick(cleanupFrequencyInMin)
	for range c {
		if !rm.IsEmpty() {
			rm.cleanRules(MaxDurationInMin)
		}
	}
}

func (rm *rulesMap) cleanRules(MaxDurationInMin time.Duration) {
	rm.lock.Lock()
	defer rm.lock.Unlock()
	now := time.Now()
	for key, rule := range rm.rules {
		if now.Sub(rule.StartTime) > MaxDurationInMin {
			logger.Debug("[Wakanda] Status:Cleanup Message:DeleteStale Key:%v KeyTime:%v CurrentTime:%v\n", key, rule.StartTime, now)
			delete(rm.rules, key)
		}
	}
}

// getNewRulesMap returns new RuleMap object
func getNewRulesMap(config Wakanda) *rulesMap {
	obj := &rulesMap{
		rules: make(map[string]*wakandaRule),
	}
	cleanup := time.Duration(config.CleanupFrequencyInMin) * time.Minute
	maxDur := time.Duration(config.MaxDurationInMin) * time.Minute
	go obj.clean(cleanup, maxDur)
	return obj
}
