package tagbidder

var bidderMapper map[string]Mapper

//RegisterBidderMapper will be used by each bidder to set its respective macro Mapper
func RegisterBidderMapper(bidder string, bidderMap Mapper) {
	bidderMapper[bidder] = bidderMap
}

//GetBidderMapper will return Mapper of specific bidder
func GetBidderMapper(bidder string) Mapper {
	return bidderMapper[bidder]
}
