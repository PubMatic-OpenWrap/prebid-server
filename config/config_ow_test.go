package config

import "testing"

func testFullConfigOW(t *testing.T, cfg *Configuration) {
	cmpInts(t, "price_floor_fetcher.worker", 10, cfg.PriceFloorFetcher.Worker)
	cmpInts(t, "price_floor_fetcher.capacity", 20, cfg.PriceFloorFetcher.Capacity)
}
