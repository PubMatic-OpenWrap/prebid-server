package gdpr

import (
	"fmt"
	"time"
)

type VendorListScheduler struct {
	ticker   *time.Ticker
	interval time.Duration
	done     chan bool
}

func NewVendorListScheduler(interval string) (*VendorListScheduler, error) {
	d, err := time.ParseDuration(interval)
	if err != nil {
		return nil, err
	}

	scheduler := &VendorListScheduler{
		ticker:   nil,
		interval: d,
		done:     make(chan bool),
	}
	return scheduler, nil
}

func (scheduler *VendorListScheduler) Start() {
	scheduler.ticker = time.NewTicker(scheduler.interval)
	go func() {
		for {
			select {
			case <-scheduler.done:
				return
			case t := <-scheduler.ticker.C:
				fmt.Println("Tick at", t)
			}
		}
	}()
}

func (scheduler *VendorListScheduler) Stop() {
	scheduler.ticker.Stop()
	scheduler.done <- true
}
