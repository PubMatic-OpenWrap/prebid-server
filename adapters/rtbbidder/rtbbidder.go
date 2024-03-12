package rtbbidder

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"

	"github.com/prebid/openrtb/v19/openrtb2"
)

// reference: https://docs.prebid.org/prebid-server/developers/add-new-bidder-go.html
type RTBBidder struct {
	RequestMode RequestMode
	Uri         string
	syncher     Syncer
}

type RequestMode int

const (
	Multi RequestMode = iota // default is multi
	Single
)

// oRTB 2.6
func (r *RTBBidder) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	fmt.Println("Making RTB Requests")
	bidderInfo := GetSyncer().BidderInfos[reqInfo.BidderName]

	requestData := make([]*adapters.RequestData, 0)
	var errs []error
	var bidderUrl string = ""
	/* Iterate over each impression and determine this bidder specific param value */
	for _, imp := range request.Imp {
		impExt := adapters.ExtImpBidder{}
		if err := json.Unmarshal(imp.Ext, &impExt); err != nil {
			errs = append(errs, err)
		}
		fmt.Println(string(imp.Ext))
		paramsMap := map[string]string{}
		json.Unmarshal(impExt.Bidder, &paramsMap)
		if bidderUrl == "" {
			// assuming same url is present across all bidder params
			bidderUrl = paramsMap["uri"]
		}
		if r.RequestMode == Single {
			clonedRequest := request
			clonedRequest.Imp = []openrtb2.Imp{imp}
			body, err := json.Marshal(clonedRequest)
			if err == nil {
				requestData = append(requestData, &adapters.RequestData{
					Method: "POST",
					Uri:    bidderInfo.Endpoint,
					Body:   body,
				})
			}
		}
		if r.RequestMode == Multi {
			body, err := json.Marshal(request)
			if err == nil {
				requestData = append(requestData, &adapters.RequestData{
					Method: "POST",
					Uri:    bidderInfo.Endpoint,
					Body:   body,
				})
			}
			break
		}
	}

	return requestData, errs
}

func (r *RTBBidder) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	fmt.Println("Making RTB Response")
	var ortbResponse *openrtb2.BidResponse
	var errors []error
	var bidderResponse *adapters.BidderResponse
	if err := json.Unmarshal(response.Body, &ortbResponse); err != nil {
		errors = append(errors, err)
	} else {
		if ortbResponse != nil {
			bidderResponse = &adapters.BidderResponse{
				Currency: ortbResponse.Cur,
				Bids:     getTypeBids(*ortbResponse),
			}
		}
	}
	return bidderResponse, errors
}

// Builder builds a new instance of the Pubmatic adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	syncher := Syncer{}
	singleTonbidder := &RTBBidder{
		Uri:     config.Endpoint,
		syncher: syncher,
	}
	return singleTonbidder, nil
}

func getInstance() *RTBBidder {
	return singleTonbidder
}

func GetSyncer() *Syncer {
	return &singleTonbidder.syncher
}

var singleTonbidder *RTBBidder = &RTBBidder{
	syncher: Syncer{
		syncPath:         "/../rtb",
		syncedBiddersMap: make(map[string]struct{}),
		InfoAwareBidders: make(map[string]adapters.Bidder),
		// assume- we will get this 'AliasMap' from database query execution
		AliasMap: map[string]string{
			"magnite-1":         "rtb_magnite",
			"magnite_alias":     "magnite",
			"myrtbbidder-1":     "myrtbbidder",
			"ashish-1":          "ashish",
			"rtb_magnite_core":  "rtb_magnite",
			"rtb_magnite_alias": "rtb_magnite",
			"rtb_magnite_bc":    "rtb_magnite",
			"rtb_magnite_demo":  "rtb_magnite",
			"magnite_abcd":      "rtb_magnite",
		},
		UserSyncData:            &sync.Map{},
		RTBBidderToSyncerKey:    &sync.Map{},
		RTBBidderGVLVEndorIDMap: &sync.Map{},
	},
}

func init() {
	singleTonbidder.syncher.syncCoreBidders()
	singleTonbidder.syncher.sync()
}

func getTypeBids(bidResponse openrtb2.BidResponse) []*adapters.TypedBid {
	tBids := make([]*adapters.TypedBid, 0)
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			tBids = append(tBids, &adapters.TypedBid{
				Bid: &bid,
			})
		}
	}
	return tBids
}
