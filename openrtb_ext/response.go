package openrtb_ext

import (
	"encoding/json"

	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
)

type NonBidStatusCode openrtb3.LossReason

// ExtBidResponse defines the contract for bidresponse.ext
type ExtBidResponse struct {
	Debug *ExtResponseDebug `json:"debug,omitempty"`
	// Errors defines the contract for bidresponse.ext.errors
	Errors   map[BidderName][]ExtBidderMessage `json:"errors,omitempty"`
	Warnings map[BidderName][]ExtBidderMessage `json:"warnings,omitempty"`
	// ResponseTimeMillis defines the contract for bidresponse.ext.responsetimemillis
	ResponseTimeMillis map[BidderName]int `json:"responsetimemillis,omitempty"`
	// RequestTimeoutMillis returns the timeout used in the auction.
	// This is useful if the timeout is saved in the Stored Request on the server.
	// Clients can run one auction, and then use this to set better connection timeouts on future auction requests.
	RequestTimeoutMillis int64 `json:"tmaxrequest,omitempty"`
	// ResponseUserSync defines the contract for bidresponse.ext.usersync
	Usersync map[BidderName]*ExtResponseSyncData `json:"usersync,omitempty"`
	// Prebid defines the contract for bidresponse.ext.prebid
	Prebid *ExtResponsePrebid `json:"prebid,omitempty"`

	OwMatchedImpression json.RawMessage `json:"matchedimpression,omitempty"`
	OwSendAllBids       int             `json:"sendallbids,omitempty"`
	OwLogger            string          `json:"owlogger,omitempty"`
	Wrapper             *ExtWrapper     `json:"wrapper,omitempty"`
}

type ExtWrapper struct {
	ResponseFormat   string            `json:"responseformat,omitempty"`
	RedirectURL      string            `json:"redirecturl,omitempty"`
	ImpToAdServerURL map[string]string `json:"imptoadserverurl,omitempty"`
}

// ExtResponseDebug defines the contract for bidresponse.ext.debug
type ExtResponseDebug struct {
	// HttpCalls defines the contract for bidresponse.ext.debug.httpcalls
	HttpCalls map[BidderName][]*ExtHttpCall `json:"httpcalls,omitempty"`
	// Request after resolution of stored requests and debug overrides
	ResolvedRequest json.RawMessage `json:"resolvedrequest,omitempty"`
}

// ExtResponseSyncData defines the contract for bidresponse.ext.usersync.{bidder}
type ExtResponseSyncData struct {
	Status CookieStatus `json:"status"`
	// Syncs must have length > 0
	Syncs []*ExtUserSync `json:"syncs"`
}

// ExtResponsePrebid defines the contract for bidresponse.ext.prebid
type ExtResponsePrebid struct {
	AuctionTimestamp int64             `json:"auctiontimestamp,omitempty"`
	Passthrough      json.RawMessage   `json:"passthrough,omitempty"`
	Modules          json.RawMessage   `json:"modules,omitempty"`
	Fledge           *Fledge           `json:"fledge,omitempty"`
	Targeting        map[string]string `json:"targeting,omitempty"`
	// SeatNonBid holds the array of Bids which are either rejected, no bids inside bidresponse.ext.prebid.seatnonbid
	SeatNonBid []SeatNonBid     `json:"seatnonbid,omitempty"`
	Floors     *PriceFloorRules `json:"floors,omitempty"`
}

// FledgeResponse defines the contract for bidresponse.ext.fledge
type Fledge struct {
	AuctionConfigs []*FledgeAuctionConfig `json:"auctionconfigs,omitempty"`
}

// FledgeAuctionConfig defines the container for bidresponse.ext.fledge.auctionconfigs[]
type FledgeAuctionConfig struct {
	ImpId   string          `json:"impid"`
	Bidder  string          `json:"bidder,omitempty"`
	Adapter string          `json:"adapter,omitempty"`
	Config  json.RawMessage `json:"config"`
}

// ExtUserSync defines the contract for bidresponse.ext.usersync.{bidder}.syncs[i]
type ExtUserSync struct {
	Url  string       `json:"url"`
	Type UserSyncType `json:"type"`
}

// ExtBidderMessage defines an error object to be returned, consiting of a machine readable error code, and a human readable error message string.
type ExtBidderMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ExtHttpCall defines the contract for a bidresponse.ext.debug.httpcalls.{bidder}[i]
type ExtHttpCall struct {
	Uri            string              `json:"uri"`
	RequestBody    string              `json:"requestbody"`
	RequestHeaders map[string][]string `json:"requestheaders"`
	ResponseBody   string              `json:"responsebody"`
	Status         int                 `json:"status"`
	Params         map[string]int      `json:"params,omitempty"`
}

// CookieStatus describes the allowed values for bidresponse.ext.usersync.{bidder}.status
type CookieStatus string

const (
	CookieNone      CookieStatus = "none"
	CookieExpired   CookieStatus = "expired"
	CookieAvailable CookieStatus = "available"
)

// UserSyncType describes the allowed values for bidresponse.ext.usersync.{bidder}.syncs[i].type
type UserSyncType string

const (
	UserSyncIframe UserSyncType = "iframe"
	UserSyncPixel  UserSyncType = "pixel"
)

// ExtNonBidPrebidBid is subset of Bid object with exact json signature
// defined at https://github.com/prebid/openrtb/blob/v19.0.0/openrtb2/bid.go
// It also contains the custom fields
type ExtNonBidPrebidBid struct {
	Price   float64                 `json:"price,omitempty"`
	ADomain []string                `json:"adomain,omitempty"`
	CatTax  adcom1.CategoryTaxonomy `json:"cattax,omitempty"`
	Cat     []string                `json:"cat,omitempty"`
	DealID  string                  `json:"dealid,omitempty"`
	W       int64                   `json:"w,omitempty"`
	H       int64                   `json:"h,omitempty"`
	Dur     int64                   `json:"dur,omitempty"`
	MType   openrtb2.MarkupType     `json:"mtype,omitempty"`

	// Custom Fields
	OriginalBidCPM float64 `json:"origbidcpm,omitempty"`
	OriginalBidCur string  `json:"origbidcur,omitempty"`

	//OW specific fields
	ID                string              `json:"id,omitempty"` // need to check
	DealPriority      int                 `json:"dealpriority,omitempty"`
	DealTierSatisfied bool                `json:"dealtiersatisfied,omitempty"`
	Meta              *ExtBidPrebidMeta   `json:"meta,omitempty"`
	Targeting         map[string]string   `json:"targeting,omitempty"`
	Type              BidType             `json:"type,omitempty"`
	Video             *ExtBidPrebidVideo  `json:"video,omitempty"`
	BidId             string              `json:"bidid,omitempty"`
	Floors            *ExtBidPrebidFloors `json:"floors,omitempty"`
	OriginalBidCPMUSD float64             `json:"origbidcpmusd,omitempty"`
	Bundle            string              `json:"bndl,omitempty"`
}

// ExtNonBidPrebid represents bidresponse.ext.prebid.seatnonbid[].nonbid[].ext
type ExtNonBidPrebid struct {
	Bid ExtNonBidPrebidBid `json:"bid"`
}

type ExtNonBid struct {
	Prebid  ExtNonBidPrebid `json:"prebid"`
	IsAdPod *bool           `json:"-"` // OW specific Flag to determine if it is Ad-Pod specific nonbid
}

// NonBid represnts the Non Bid Reason (statusCode) for given impression ID
type NonBid struct {
	ImpId      string    `json:"impid"`
	StatusCode int       `json:"statuscode"`
	Ext        ExtNonBid `json:"ext"`
}

// SeatNonBid is collection of NonBid objects with seat information
type SeatNonBid struct {
	NonBid []NonBid        `json:"nonbid"`
	Seat   string          `json:"seat"`
	Ext    json.RawMessage `json:"ext,omitempty"`
}
