package tagbidder

import "github.com/PubMatic-OpenWrap/openrtb"

type ITagBidder interface {
	Name() string
	URL() string
	LoadImpression(imp *openrtb.Imp) error
}
