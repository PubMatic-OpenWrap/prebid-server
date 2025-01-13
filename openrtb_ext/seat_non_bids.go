package openrtb_ext

import (
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v2/util/uuidutil"
)

type SeatNonBidBuilder map[string][]NonBid

// NonBidParams contains the fields that are required to form the nonBid object
type NonBidParams struct {
	Bid               *openrtb2.Bid
	NonBidReason      int
	OriginalBidCPM    float64
	OriginalBidCur    string
	DealPriority      int
	DealTierSatisfied bool
	GeneratedBidID    string
	TargetBidderCode  string
	OriginalBidCPMUSD float64
	BidMeta           *ExtBidPrebidMeta
	BidType           BidType
	BidTargets        map[string]string
	BidVideo          *ExtBidPrebidVideo
	BidEvents         *ExtBidPrebidEvents
	BidFloors         *ExtBidPrebidFloors
}

// mock uuid instance
const fakeUuid = "30470a14-2949-4110-abce-b62d57304ad5"

type testUUIDGenerator struct{}

func (testUUIDGenerator) Generate() (string, error) {
	return fakeUuid, nil
}

func TestUuidGeneratorInstance() uuidutil.UUIDGenerator {
	uuidGenerator = testUUIDGenerator{}
	return uuidGenerator
}

var uuidGenerator uuidutil.UUIDGenerator = uuidutil.UUIDRandomGenerator{}

// NewNonBid creates the NonBid object from NonBidParams and return it
func NewNonBid(bidParams NonBidParams) NonBid {
	if bidParams.Bid == nil {
		bidParams.Bid = &openrtb2.Bid{}
	}

	var bidId string
	if bidParams.Bid.ID == "" {
		uuid, _ := uuidGenerator.Generate()
		bidId = uuid
	} else {
		bidId = bidParams.Bid.ID
	}

	return NonBid{
		ImpId:      bidParams.Bid.ImpID,
		StatusCode: bidParams.NonBidReason,
		Ext: ExtNonBid{
			Prebid: ExtNonBidPrebid{Bid: ExtNonBidPrebidBid{
				Price:          bidParams.Bid.Price,
				ADomain:        bidParams.Bid.ADomain,
				CatTax:         bidParams.Bid.CatTax,
				Cat:            bidParams.Bid.Cat,
				DealID:         bidParams.Bid.DealID,
				W:              bidParams.Bid.W,
				H:              bidParams.Bid.H,
				Dur:            bidParams.Bid.Dur,
				MType:          bidParams.Bid.MType,
				OriginalBidCPM: bidParams.OriginalBidCPM,
				OriginalBidCur: bidParams.OriginalBidCur,

				//OW specific
				ID:                bidId,
				DealPriority:      bidParams.DealPriority,
				DealTierSatisfied: bidParams.DealTierSatisfied,
				Meta:              bidParams.BidMeta,
				Targeting:         bidParams.BidTargets,
				Type:              bidParams.BidType,
				Video:             bidParams.BidVideo,
				BidId:             bidParams.GeneratedBidID,
				Floors:            bidParams.BidFloors,
				OriginalBidCPMUSD: bidParams.OriginalBidCPMUSD,
			}},
		},
	}
}

// AddBid adds the nonBid into the map against the respective seat.
// Note: This function is not a thread safe.
func (snb *SeatNonBidBuilder) AddBid(nonBid NonBid, seat string) {
	if *snb == nil {
		*snb = make(map[string][]NonBid)
	}
	(*snb)[seat] = append((*snb)[seat], nonBid)
}

// append adds the nonBids from the input nonBids to the current nonBids.
// This method is not thread safe as we are initializing and writing to map
func (snb *SeatNonBidBuilder) Append(nonBids ...SeatNonBidBuilder) {
	if *snb == nil {
		return
	}
	for _, nonBid := range nonBids {
		for seat, nonBids := range nonBid {
			(*snb)[seat] = append((*snb)[seat], nonBids...)
		}
	}
}

// Get function converts the internal seatNonBidsMap to standard openrtb seatNonBid structure and returns it
func (snb *SeatNonBidBuilder) Get() []SeatNonBid {
	if *snb == nil {
		return nil
	}
	var seatNonBid []SeatNonBid
	for seat, nonBids := range *snb {
		seatNonBid = append(seatNonBid, SeatNonBid{
			Seat:   seat,
			NonBid: nonBids,
		})
	}
	return seatNonBid
}

// rejectImps appends a non bid object to the builder for every specified imp
func (snb *SeatNonBidBuilder) RejectImps(impIds []string, nonBidReason openrtb3.NoBidReason, seat string) {
	nonBids := []NonBid{}
	for _, impId := range impIds {
		nonBid := NonBid{
			ImpId:      impId,
			StatusCode: int(nonBidReason),
		}
		nonBids = append(nonBids, nonBid)
	}

	if len(nonBids) > 0 {
		(*snb)[seat] = append((*snb)[seat], nonBids...)
	}
}
