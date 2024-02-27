package openrtb_ext

import "github.com/prebid/openrtb/v19/openrtb2"

// NonBidCollection contains the map of seat with list of nonBids
type NonBidCollection struct {
	seatNonBidsMap map[string][]NonBid
}

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

// NewNonBid creates the NonBid object from NonBidParams and return it
func NewNonBid(bidParams NonBidParams) NonBid {
	if bidParams.Bid == nil {
		bidParams.Bid = &openrtb2.Bid{}
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
				ID:                bidParams.Bid.ID,
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
func (snb *NonBidCollection) AddBid(nonBid NonBid, seat string) {
	if snb.seatNonBidsMap == nil {
		snb.seatNonBidsMap = make(map[string][]NonBid)
	}
	snb.seatNonBidsMap[seat] = append(snb.seatNonBidsMap[seat], nonBid)
}

// Append functions appends the NonBids from the input instance into the current instance's seatNonBidsMap, creating the map if needed.
// Note: This function is not a thread safe.
func (snb *NonBidCollection) Append(nonbid NonBidCollection) {
	if snb == nil || len(nonbid.seatNonBidsMap) == 0 {
		return
	}
	if snb.seatNonBidsMap == nil {
		snb.seatNonBidsMap = make(map[string][]NonBid, len(nonbid.seatNonBidsMap))
	}
	for seat, nonBids := range nonbid.seatNonBidsMap {
		snb.seatNonBidsMap[seat] = append(snb.seatNonBidsMap[seat], nonBids...)
	}
}

// Get function converts the internal seatNonBidsMap to standard openrtb seatNonBid structure and returns it
func (snb *NonBidCollection) Get() []SeatNonBid {
	if snb == nil {
		return nil
	}

	// seatNonBid := make([]SeatNonBid, len(snb.seatNonBidsMap))
	var seatNonBid []SeatNonBid
	for seat, nonBids := range snb.seatNonBidsMap {
		seatNonBid = append(seatNonBid, SeatNonBid{
			Seat:   seat,
			NonBid: nonBids,
		})
	}
	return seatNonBid
}

func (snb *NonBidCollection) GetSeatNonBidMap() map[string][]NonBid {
	if snb == nil {
		return nil
	}
	return snb.seatNonBidsMap
}
