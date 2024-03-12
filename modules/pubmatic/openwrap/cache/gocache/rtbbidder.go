// Package gocache contains caching functionalities of database.
// This file provides caching functionalities for RTBBidder.
package gocache

import "github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"

// GetRTBBidders simply forwards the call to database.
// This will not set the data in cache since it maintains it's own map as cache
// Adding this function only because we are calling all database functions through cache
func (c *cache) GetRTBBidders() (map[string]models.RTBBidderData, error) {
	return c.db.GetRTBBidders()
}
