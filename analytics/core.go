package analytics

import (
	"context"
	"time"

	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
)

/*
  	PBSAnalyticsModule must be implemented by any analytics module that does transactional logging.

	New modules can use the /analytics/endpoint_data_objects, extract the
	information required and are responsible for handling all their logging activities inside LogAuctionObject, LogAmpObject
	LogCookieSyncObject and LogSetUIDObject method implementations.
*/

type PBSAnalyticsModule interface {
	LogAuctionObject(*AuctionObject)
	LogVideoObject(*VideoObject)
	LogCookieSyncObject(*CookieSyncObject)
	LogSetUIDObject(*SetUIDObject)
	LogAmpObject(*AmpObject)
	LogNotificationEventObject(*NotificationEvent)
}

// LoggableAuctionObject contains common attributes between AuctionObject, AmpObject, VideoObject
type LoggableAuctionObject struct {
	Context      context.Context
	Status       int
	Errors       []error
	Request      *openrtb2.BidRequest
	Response     *openrtb2.BidResponse
	RejectedBids []RejectedBid
}

//Loggable object of a transaction at /openrtb2/auction endpoint
type AuctionObject struct {
	LoggableAuctionObject
	Account   *config.Account
	StartTime time.Time
}

//Loggable object of a transaction at /openrtb2/amp endpoint
type AmpObject struct {
	LoggableAuctionObject
	AmpTargetingValues map[string]string
	Origin             string
	StartTime          time.Time
}

//Loggable object of a transaction at /openrtb2/video endpoint
type VideoObject struct {
	LoggableAuctionObject
	VideoRequest  *openrtb_ext.BidRequestVideo
	VideoResponse *openrtb_ext.BidResponseVideo
	StartTime     time.Time
}

//Loggable object of a transaction at /setuid
type SetUIDObject struct {
	Status  int
	Bidder  string
	UID     string
	Errors  []error
	Success bool
}

//Loggable object of a transaction at /cookie_sync
type CookieSyncObject struct {
	Status       int
	Errors       []error
	BidderStatus []*CookieSyncBidder
}

type CookieSyncBidder struct {
	BidderCode   string        `json:"bidder"`
	NoCookie     bool          `json:"no_cookie,omitempty"`
	UsersyncInfo *UsersyncInfo `json:"usersync,omitempty"`
}

type UsersyncInfo struct {
	URL         string `json:"url,omitempty"`
	Type        string `json:"type,omitempty"`
	SupportCORS bool   `json:"supportCORS,omitempty"`
}

// NotificationEvent is a loggable object
type NotificationEvent struct {
	Request *EventRequest   `json:"request"`
	Account *config.Account `json:"account"`
}
