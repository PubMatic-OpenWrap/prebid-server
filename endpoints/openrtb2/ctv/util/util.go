package util

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/mxmCherry/openrtb/v15/openrtb2"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/constant"
	"github.com/prebid/prebid-server/endpoints/openrtb2/ctv/types"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func GetDurationWiseBidsBucket(bids []*types.Bid) types.BidsBuckets {
	result := types.BidsBuckets{}

	for i, bid := range bids {
		result[bid.Duration] = append(result[bid.Duration], bids[i])
	}

	for k, v := range result {
		//sort.Slice(v[:], func(i, j int) bool { return v[i].Price > v[j].Price })
		sortBids(v[:])
		result[k] = v
	}

	return result
}

func sortBids(bids []*types.Bid) {
	sort.Slice(bids, func(i, j int) bool {
		if bids[i].DealTierSatisfied == bids[j].DealTierSatisfied {
			return bids[i].Price > bids[j].Price
		}
		return bids[i].DealTierSatisfied
	})
}

// GetDealTierSatisfied ...
func GetDealTierSatisfied(ext *openrtb_ext.ExtBid) bool {
	return ext != nil && ext.Prebid != nil && ext.Prebid.DealTierSatisfied
}

func DecodeImpressionID(id string) (string, int) {
	index := strings.LastIndex(id, constant.CTVImpressionIDSeparator)
	if index == -1 {
		return id, 0
	}

	sequence, err := strconv.Atoi(id[index+1:])
	if nil != err || 0 == sequence {
		return id, 0
	}

	return id[:index], sequence
}

func GetCTVImpressionID(impID string, seqNo int) string {
	return fmt.Sprintf(constant.CTVImpressionIDFormat, impID, seqNo)
}

func GetUniqueBidID(bidID string, id int) string {
	return fmt.Sprintf(constant.CTVUniqueBidIDFormat, id, bidID)
}

var Logf = func(msg string, args ...interface{}) {
	if glog.V(3) {
		glog.Infof(msg, args...)
	}
	//fmt.Printf(msg+"\n", args...)
}

func JLogf(msg string, obj interface{}) {
	if glog.V(3) {
		data, _ := json.Marshal(obj)
		glog.Infof("[OPENWRAP] %v:%v", msg, string(data))
	}
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	Logf("[TIMETRACK] %s took %s", name, elapsed)
	//eg: defer TimeTrack(time.Now(), "factorial")
}

// GetTargeting returns the value of targeting key associated with bidder
// it is expected that bid.Ext contains prebid.targeting map
// if value not present or any error occured empty value will be returned
// along with error.
func GetTargeting(key openrtb_ext.TargetingKey, bidder openrtb_ext.BidderName, bid openrtb2.Bid) (string, error) {
	bidderSpecificKey := key.BidderKey(openrtb_ext.BidderName(bidder), 20)
	return jsonparser.GetString(bid.Ext, "prebid", "targeting", bidderSpecificKey)
}
