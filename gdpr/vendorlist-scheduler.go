package gdpr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type vendorListScheduler struct {
	ticker    *time.Ticker
	interval  time.Duration
	done      chan bool
	isRunning bool
	lastRun   time.Time

	httpClient *http.Client
	timeout    time.Duration
}

//Only single instance must be created
var _instance *vendorListScheduler

func GetVendorListScheduler(interval, timeout string, httpClient *http.Client) (*vendorListScheduler, error) {
	if _instance != nil {
		return _instance, nil
	}

	intervalDuration, err := time.ParseDuration(interval)
	if err != nil {
		return nil, errors.New("error parsing vendor list scheduler interval: " + err.Error())
	}

	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, errors.New("error parsing vendor list scheduler timeout: " + err.Error())
	}

	_instance := &vendorListScheduler{
		ticker:     nil,
		interval:   intervalDuration,
		done:       make(chan bool),
		httpClient: httpClient,
		timeout:    timeoutDuration,
	}
	return _instance, nil
}

func (scheduler *vendorListScheduler) Start() {
	scheduler.ticker = time.NewTicker(scheduler.interval)
	go func() {
		for {
			select {
			case <-scheduler.done:
				scheduler.isRunning = false
				return
			case t := <-scheduler.ticker.C:
				if !scheduler.isRunning {
					scheduler.isRunning = true

					fmt.Println("Tick at", t)
					scheduler.runLoadCache()

					scheduler.lastRun = t
					scheduler.isRunning = false
				}
			}
		}
	}()
}

func (scheduler *vendorListScheduler) Stop() {
	if scheduler.isRunning {
		scheduler.ticker.Stop()
		scheduler.done <- true
	}
}

func (scheduler *vendorListScheduler) runLoadCache() {
	preloadContext, cancel := context.WithTimeout(context.Background(), scheduler.timeout)
	defer cancel()
	//loadCache(preloadContext, scheduler.httpClient, vendorListURLMaker, cacheSave)

	latestVersion := saveOne(preloadContext, scheduler.httpClient, vendorListURLMaker(0), cacheSave)

	// The GVL for TCF2 has no vendors defined in its first version. It's very unlikely to be used, so don't preload it.
	firstVersionToLoad := uint16(2)

	for i := latestVersion; i > firstVersionToLoad; i-- {
		// Check if version is present in cache
		if list := cacheLoad(i); list != nil {
			continue
		}
		saveOne(preloadContext, scheduler.httpClient, vendorListURLMaker(i), cacheSave)
	}
}

// loadCache saves newly available versions of the vendor list for future use.
func loadCache(ctx context.Context, client *http.Client, urlMaker func(uint16) string, saver saveVendors) {
	latestVersion := saveOne(ctx, client, urlMaker(0), saver)

	// The GVL for TCF2 has no vendors defined in its first version. It's very unlikely to be used, so don't preload it.
	firstVersionToLoad := uint16(2)

	for i := latestVersion; i > firstVersionToLoad; i-- {
		// Check if version is present in cache
		if list := cacheLoad(i); list != nil {
			continue
		}
		saveOne(ctx, client, urlMaker(i), saver)
	}
}
