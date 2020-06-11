package ctv

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

func GetDurationWiseBidsBucket(bids []*Bid) BidsBuckets {
	result := BidsBuckets{}

	for i, bid := range bids {
		result[bid.Duration] = append(result[bid.Duration], bids[i])
	}

	for k, v := range result {
		sort.Slice(v[:], func(i, j int) bool { return v[i].Price > v[j].Price })
		result[k] = v
	}

	return result
}

func DecodeImpressionID(id string) (string, int) {
	index := strings.LastIndex(id, CTVImpressionIDSeparator)
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
	return fmt.Sprintf(CTVImpressionIDFormat, impID, seqNo)
}

func GetUniqueBidID(bidID string, id int) string {
	return fmt.Sprintf(CTVUniqueBidIDFormat, id, bidID)
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
