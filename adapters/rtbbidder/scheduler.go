package rtbbidder

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/openrtb_ext"
)

const (
	filePermission = 0755
	fileTypeYAML   = "yaml"
	fileTypeJSON   = "json"
)

// OWSyncer represents the OWSyncer that manages tasks.
type OWSyncer struct {
	tasks       []func() bool // Slice to hold tasks.
	serviceStop chan bool
	cache       cache.Cache

	syncedCoreBidders      map[string]struct{}
	syncedBidderInfos      RTBBidderInfos
	syncedInfoAwareBidders RTBInfoAwareBidders

	bidderInfoPath   string // TODO:configurable
	bidderParamsPath string
	syncInterval     time.Duration //TODO:configurable
}

var syncer *OWSyncer

// we need to access the syncer at multiple locations in PBS core
// don't see any better option than exposing it
func GetOWSyncer() *OWSyncer {
	return syncer
}

// NewOWSyncer initializes a new OWSyncer.
func NewOWSyncer(cache cache.Cache) *OWSyncer {
	syncer = &OWSyncer{
		serviceStop:            make(chan bool),
		syncedCoreBidders:      make(map[string]struct{}, 0),
		syncedBidderInfos:      RTBBidderInfos{},
		syncedInfoAwareBidders: RTBInfoAwareBidders{},

		cache:            cache,
		bidderInfoPath:   "./static/bidder-info-rtb/",   // TODO:configurable
		bidderParamsPath: "./static/bidder-params-rtb/", // TODO:configurable
		syncInterval:     15 * time.Second,              // TODO:configurable
	}
	syncer.tasks = []func() bool{
		syncer.syncRTBBidders,
		syncer.syncBidderInfosAndInfoAwareBidders,
	}
	// syncer.Run()
	return syncer
}

type RTBInfoAwareBidders struct {
	sync.RWMutex
	infoAwareBidders map[string]adapters.Bidder
}

func (r *RTBInfoAwareBidders) Get(bidder string) adapters.Bidder {
	defer r.RUnlock()
	r.RLock()
	return r.infoAwareBidders[bidder]
}

func (r *RTBInfoAwareBidders) Set(infoAwareBidders map[string]adapters.Bidder) {
	defer r.Unlock()
	r.Lock()
	r.infoAwareBidders = infoAwareBidders
}

type RTBBidderInfos struct {
	sync.RWMutex
	bidderInfos config.BidderInfos
}

func (r *RTBBidderInfos) Get(bidder string) config.BidderInfo {
	defer r.RUnlock()
	r.RLock()
	return r.bidderInfos[bidder]
}

func (r *RTBBidderInfos) Set(bidderInfos config.BidderInfos) {
	defer r.Unlock()
	r.Lock()
	r.bidderInfos = bidderInfos
}

// Run executes all tasks in the OWSyncer.
func (s *OWSyncer) Run() {
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		for _, task := range s.tasks {
			task()
		}
		// block function until the next occurrence of syncInterval or until the syncer stops
		select {
		case <-s.serviceStop:
			return
		case t := <-ticker.C:
			fmt.Printf("Sync RTBBidder @%v", t)
		}
	}
}

// syncRTBBidders fetch the list of RTB Bidders from database and register them as PBS coreBidder.
func (s *OWSyncer) syncRTBBidders() bool {
	rtbBidders, err := s.cache.GetRTBBidders()
	if err != nil {
		fmt.Printf("Failed to get RTB bidders from database = %v\n", err)
		return false
	}
	if len(rtbBidders) == 0 {
		return false
	}

	// remove all data from bidder-info and bidder-param directory
	os.RemoveAll(s.bidderInfoPath)
	os.RemoveAll(s.bidderParamsPath)

	// create files
	os.MkdirAll(s.bidderInfoPath, filePermission)
	os.MkdirAll(s.bidderParamsPath, filePermission)

	for rtbBidder, info := range rtbBidders {
		if _, present := s.syncedCoreBidders[rtbBidder]; !present {
			err = openrtb_ext.SetAliasBidderName(string(rtbBidder), openrtb_ext.BidderName(rtbBidder))
			if err != nil {
				fmt.Printf("Failed to register RTB bidder:[%s] as coreBidder, err:[%v]", rtbBidder, err)
				continue
			}
			s.syncedCoreBidders[rtbBidder] = struct{}{}
		}
		if len(info.BidderInfo) > 0 {
			os.WriteFile(s.bidderInfoPath+"/"+rtbBidder+"."+fileTypeYAML, []byte(info.BidderInfo), filePermission)
		}
		if len(info.BidderParams) > 0 {
			os.WriteFile(s.bidderParamsPath+"/"+rtbBidder+"."+fileTypeJSON, []byte(info.BidderParams), filePermission)
		}
	}

	return false
}

// syncBidderInfosAndInfoAwareBidders fetch bidder-infos of all RTB bidders from database,
// creates the yaml files and stores in the filesystem so that we can parse the bidder-info data using
// existing 'LoadBidderInfoFromDisk' function of PBS core. It also prepares infoAwareBidder for RTBBidders
func (s *OWSyncer) syncBidderInfosAndInfoAwareBidders() bool {
	if len(s.syncedCoreBidders) == 0 {
		return false
	}

	rtbBidderInfos, errs := config.LoadBidderInfoFromDisk(s.bidderInfoPath)
	fmt.Println("rtbBidderInfos:[%v],errs:[%v]", rtbBidderInfos, errs)
	if errs == nil {
		infoAwareBidders := make(map[string]adapters.Bidder, len(rtbBidderInfos))
		for bidder, info := range rtbBidderInfos {
			infoAwareBidders[bidder] = adapters.BuildInfoAwareBidder(getInstance(), info)
		}
		s.syncedInfoAwareBidders.Set(infoAwareBidders)
		s.syncedBidderInfos.Set(rtbBidderInfos)
		fmt.Println("infoAwareBidders:[%v]", infoAwareBidders)
		return true
	}
	return false
}
