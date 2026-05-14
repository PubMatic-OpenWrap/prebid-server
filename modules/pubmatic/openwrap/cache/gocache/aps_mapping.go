package gocache

import "reflect"

func (c *cache) GetApsOwMapping(slotUUID string) (string, int, bool) {
	if c.db == nil || reflect.ValueOf(c.db).IsNil() {
		return "", 0, false
	}
	return c.db.GetApsOwMapping(slotUUID)
}
