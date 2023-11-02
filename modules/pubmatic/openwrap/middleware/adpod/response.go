package middleware

import "github.com/prebid/openrtb/v19/openrtb2"

type responseBid struct {
	*openrtb2.Bid
	seat string
}
