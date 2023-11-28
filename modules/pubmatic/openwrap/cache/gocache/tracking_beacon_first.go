// Package gocache contains caching functionalities of header-bidding repostiry
// This file provides caching functionalities for tracking-beacon-first (TBF) feature details
// associated with publishers. It includes methods to interact with the underlying database package
// for retrieving and caching publisher level TBF data.
package gocache

// GetTBFTrafficForPublishers simply forwards the call to database
// This will not set the data in cache since TBF feature maintains its own map as cache
// Adding this function only because we are calling all database functions through cache
func (c *cache) GetTBFTrafficForPublishers() (pubProfileTraffic map[int]map[int]int, err error) {
	return c.db.GetTBFTrafficForPublishers()
}
