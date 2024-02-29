package rtbbidder

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/prebid/prebid-server/config"

	"github.com/prebid/prebid-server/adapters"
	// "github.com/prebid/prebid-server/config"

	"github.com/prebid/prebid-server/openrtb_ext"
)

type Syncer struct {
	syncedBidders    []string
	syncedBiddersMap map[string]struct{}
	BidderInfos      config.BidderInfos
	InfoAwareBidders map[string]adapters.Bidder

	tasks        []func() bool
	syncPath     string
	syncInfoPath string
	AliasMap     map[string]string
	HostCfg      *config.Configuration
}

type steps int

const (
	SYNC_CORE_BIDDERS steps = iota
	SYNC_BIDDER_PARAMS
	SYNC_BIDDER_INFO
)

func StoreHostConfig(cfg *config.Configuration) {
	GetSyncer().HostCfg = cfg
}

func (s *Syncer) sync() {
	// sync core bidders, sync bidder-params, sync bidder-info
	s.tasks = make([]func() bool, 3)
	schedule := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	s.tasks[SYNC_CORE_BIDDERS] = s.syncCoreBidders // core bidders task
	s.tasks[SYNC_BIDDER_INFO] = syncBiddersInfos() // core bidders task
	startScheduler(schedule, quit, s.tasks)
}

// syncCoreBidders obtains the list of RTBBidders (non-prebid) from wrapper_partner
// and uppdates the prebid coreBidderNames
func (s *Syncer) syncCoreBidders() bool {
	// list of rtb bidders from wrapper_partner
	var rtbBidders []openrtb_ext.BidderName = make([]openrtb_ext.BidderName, 0)
	/* temporary code to get list of bidders from text file, to be replaced with database-query output */
	file, err := os.Open("./corebidder.txt")
	if err != nil {
		fmt.Printf("fail to read corebidder.txt-", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rtbBidders = append(rtbBidders, openrtb_ext.BidderName(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("scanner-error: %v", err)
	}
	/* ----------------*/
	syncDone := false
	var newRTBBidders []openrtb_ext.BidderName
	for _, rtbBidder := range rtbBidders {
		if _, present := s.syncedBiddersMap[string(rtbBidder)]; !present {
			openrtb_ext.SetAliasBidderName(string(rtbBidder), rtbBidder)
			// s.syncedBidders = append(s.syncedBidders, string(rtbBidder))
			newRTBBidders = append(newRTBBidders, rtbBidder)
			s.syncedBiddersMap[string(rtbBidder)] = struct{}{}
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

// syncBiddersInfos will load bidderInfos from /static/bidder_info_rtb/
func syncBiddersInfos() func() bool {
	return func() bool {
		fmt.Printf("Synching Bidder Info")
		rtbBidderInfos, errs := config.LoadBidderInfoFromDisk("./static/bidder_info_rtb/")
		fmt.Printf("syncBiddersInfos [%v] - err-[%v]", rtbBidderInfos, errs)
		if errs == nil {
			// will need mutex here since auction pkg refers the same
			GetSyncer().BidderInfos = rtbBidderInfos

			for bidder, info := range rtbBidderInfos {
				GetSyncer().InfoAwareBidders[bidder] = adapters.BuildInfoAwareBidder(getInstance(), info)
			}
		}
		return true
	}
}

func (r *RTBBidder) SyncBidderInfos() {
	syncBiddersInfos()
}

func startScheduler(ticker *time.Ticker, quit chan struct{}, tasks []func() bool) {
	go func() {
		for {
			select {
			case <-ticker.C:
				tasks := GetSyncer().tasks
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
