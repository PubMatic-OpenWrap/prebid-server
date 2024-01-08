package rtbbidder

import (
	"fmt"
	"strings"
	"time"

	"github.com/prebid/prebid-server/openrtb_ext"
)

type Syncer struct {
	syncedBidders []string
	tasks         []func() bool
	syncPath      string
}

type steps int

const (
	SYNC_CORE_BIDDERS steps = iota
	SYNC_BIDDER_PARAMS
)

func (s *Syncer) sync() {
	// sync core bidders, sync bidder-params, sync bidder-info
	s.tasks = make([]func() bool, 3)
	schedule := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	s.tasks[SYNC_CORE_BIDDERS] = s.syncCoreBidders // core bidders task
	startScheduler(schedule, quit, s.tasks)
}

// syncCoreBidders obtains the list of RTBBidders (non-prebid) from wrapper_partner
// and uppdates the prebid coreBidderNames
func (s *Syncer) syncCoreBidders() bool {
	// list of rtb bidders from wrapper_partner
	var rtbBidders []openrtb_ext.BidderName
	rtbBidders = append(rtbBidders, openrtb_ext.BidderName("myrtbbidder"))
	syncDone := false
	var newRTBBidders []openrtb_ext.BidderName
	for _, rtbBidder := range rtbBidders {
		if !contains(s.syncedBidders, string(rtbBidder)) {
			openrtb_ext.SetAliasBidderName(string(rtbBidder), rtbBidder)
			s.syncedBidders = append(s.syncedBidders, string(rtbBidder))
			newRTBBidders = append(newRTBBidders, rtbBidder)
			syncDone = true
		}
	}

	fmt.Printf("Synced New Bidders = %v\n", newRTBBidders)

	return syncDone
}

// func (s *Syncer) BuildAndSyncRTBAdapters(builder func() bool) []error {
// 	s.tasks = append(s.tasks, builder) // build RTB adapters task
// 	return nil
// }

/*
SyncBiddersParameters collects the RTBBidders specific bidder-params.json
file from static directory and updated the OpenWrap spcific paramsValidator
*/
func (s *Syncer) syncBiddersParameters(paramsValidator *OpenWrapbidderParamValidator) bool {
	// fmt.Println(paramsValidator)
	s.tasks[SYNC_BIDDER_PARAMS] = func() bool {
		syncedBidders := paramsValidator.LoadSchema()
		fmt.Printf("Synced Bidders Parameters for bidders = %v\n", strings.Join(syncedBidders, ","))
		return true
	}
	return false
}

func syncBiddersInfos(s *Syncer) bool {
	// rtbBidderInfos, errs := config.LoadBidderInfoFromDisk(main_ow.InfoDirectory + s.syncPath)
	fmt.Printf("Synching Bidder Info")
	return false
}

func (r *RTBBidder) SyncBidderInfos() {
	syncBiddersInfos(&r.syncher)
}

func startScheduler(ticker *time.Ticker, quit chan struct{}, tasks []func() bool) {
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, task := range tasks {
					if task != nil {
						task()
					}
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
