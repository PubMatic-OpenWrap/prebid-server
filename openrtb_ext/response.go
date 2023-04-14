package openrtb_ext

import (
	"encoding/json"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/openrtb/v19/openrtb3"
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

	MatchedImpression json.RawMessage `json:"matchedimpression,omitempty"`
	SendAllBids       int             `json:"sendallbids,omitempty"`
	LogInfo           LogInfo         `json:"loginfo,omitempty"`
	Logger            string          `json:"logger,omitempty"`
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
	Floors           *PriceFloorRules  `json:"floors,omitempty"`
	SeatNonBid       []SeatNonBid      `json:"seatnonbid,omitempty"`
}

// Bid is Wrapper around original/proxy bid object
type Bid struct {
	openrtb2.Bid
	ID    string `json:"id,omitempty"`    // added omitempty
	ImpID string `json:"impid,omitempty"` // added omitempty

	OriginalBidCPM float64 `json:"originalbidcpm,omitempty"`
	OriginalBidCur string  `json:"originalbidcur,omitempty"`
}

// ExtResponseNonBidPrebid represents bidresponse.ext.prebid.seatnonbid[].nonbid[].ext
type ExtResponseNonBidPrebid struct {
	Bid Bid `json:"bid"`
}

type NonBidExt struct {
	Prebid ExtResponseNonBidPrebid `json:"prebid"`
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

// SeatNonBidResponse defines the contract for bidresponse.ext.debug.seatnonbid
type SeatNonBidResponse struct {
	SeatNonBids []SeatNonBid `json:"seatnonbid,omitempty"`
}

// SeatNonBid defines the contract to hold all elements of single seatnonbid
type SeatNonBid struct {
	NonBids []NonBid `json:"nonbid,omitempty"`
	Seat    string   `json:"seat,omitempty"`
}

// NonBid defines the contract for bidresponse.ext.debug.seatnonbid.nonbid
type NonBid struct {
	ImpId      string                    `json:"impid,omitempty"`
	StatusCode openrtb3.NonBidStatusCode `json:"statuscode,omitempty"`
	Ext        NonBidExt                 `json:"ext,omitempty"`
}

// ExtNonBid defines the contract for bidresponse.ext.debug.seatnonbid.nonbid.ext
type ExtNonBid struct {
	Prebid  *ExtNonBidPrebid `json:"prebid,omitempty"`
	IsAdPod *bool            `json:"-"` // OW specific Flag to determine if it is Ad-Pod specific nonbid
}

// ExtNonBidPrebid defines the contract for bidresponse.ext.debug.seatnonbid.nonbid.ext.prebid
type ExtNonBidPrebid struct {
	Bid interface{} `json:"bid,omitempty"` // To be removed once we start using single "Bid" data-type (unlike V25.Bid and openrtb2.Bid)
}

// LogInfo contains the logger, tracker calls to be sent in response
type LogInfo struct {
	Logger  string `json:"logger,omitempty"`
	Tracker string `json:"tracker,omitempty"`
}
