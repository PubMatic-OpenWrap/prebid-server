package usersync

import (
	"sync"
)

type AdapterSyncerMap struct {
	PrebidBidder map[string]Syncer
	RTBBidder    *sync.Map
}

func (a AdapterSyncerMap) Get(key string) (Syncer, bool) {
	value, present := a.PrebidBidder[key]
	if !present {
		var ivalue any
		ivalue, present = a.RTBBidder.Load(key)
		value, _ = (ivalue).(Syncer)
	}
	return value.(Syncer), present
}

type BidderToSyncerKey struct {
	PrebidBidderToSyncerKey map[string]string
	RTBBidderToSyncerKey    *sync.Map
}

func (b BidderToSyncerKey) Get(key string) (string, bool) {
	value, present := b.PrebidBidderToSyncerKey[key]
	// if !present {
	// 	var ivalue any
	// 	ivalue, present = b.RTBBidderToSyncerKey.Load(key)
	// 	value, _ = (ivalue).(string)
	// }
	return value, present
}
