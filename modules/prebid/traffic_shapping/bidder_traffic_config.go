package trafficshapping

type BidderTrafficConfig struct { // feature to pick and choose who and when to send this acros
	Bidder              string     // typically seat
	Targeting           Expression // targeing rules
	SupportedKeys       []string   // send all Key values for PGs and PMPs that Fox sources through a chosen SSP
	RTBRequestTargeting Expression
}
