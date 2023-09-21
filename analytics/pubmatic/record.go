package pubmatic

import (
	"encoding/json"
	"math"
	"strings"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// WloggerRecord structure for wrapper analytics logger object
type WloggerRecord struct {
	record
}

type record struct {
	Timeout           int              `json:"to,omitempty"`
	PubID             int              `json:"pubid,omitempty"`
	PageURL           string           `json:"purl,omitempty"`
	Timestamp         int64            `json:"tst,omitempty"`
	IID               string           `json:"iid,omitempty"`
	ProfileID         string           `json:"pid,omitempty"`
	VersionID         string           `json:"pdvid,omitempty"`
	IP                string           `json:"-"`
	UserAgent         string           `json:"-"`
	UID               string           `json:"-"`
	GDPR              int8             `json:"gdpr,omitempty"`
	ConsentString     string           `json:"cns,omitempty"`
	PubmaticConsent   int              `json:"pmc,omitempty"`
	UserID            string           `json:"uid,omitempty"`
	PageValue         float64          `json:"pv,omitempty"` //sum of all winning bids
	ServerLogger      int              `json:"sl,omitempty"`
	Slots             []SlotRecord     `json:"s,omitempty"`
	CachePutMiss      int              `json:"cm,omitempty"`
	Origin            string           `json:"orig,omitempty"`
	Device            Device           `json:"dvc,omitempty"`
	AdPodPercentage   *AdPodPercentage `json:"aps,omitempty"`
	Content           *Content         `json:"ct,omitempty"`
	TestConfigApplied int              `json:"tgid,omitempty"`
	//Geo             GeoRecord    `json:"geo,omitempty"`
}

// Device struct for storing device information
type Device struct {
	Platform models.DevicePlatform `json:"plt,omitempty"`
	IFAType  *models.DeviceIFAType `json:"ifty,omitempty"` //OTT-416, adding device.ext.ifa_type
}

/*
//GeoRecord structure for storing geo information
type GeoRecord struct {
	CountryCode string `json:"cc,omitempty"`
}
*/

// AdPodPercentage will store adpod percentage value comes in request
type AdPodPercentage struct {
	CrossPodAdvertiserExclusionPercent  *int `json:"cpexap,omitempty"` //Percent Value - Across multiple impression there will be no ads from same advertiser. Note: These cross pod rule % values can not be more restrictive than per pod
	CrossPodIABCategoryExclusionPercent *int `json:"cpexip,omitempty"` //Percent Value - Across multiple impression there will be no ads from same advertiser
	IABCategoryExclusionWindow          *int `json:"exapw,omitempty"`  //Duration in minute between pods where exclusive IAB rule needs to be applied
	AdvertiserExclusionWindow           *int `json:"exipw,omitempty"`  //Duration in minute between pods where exclusive advertiser rule needs to be applied
}

// Content of openrtb request object
type Content struct {
	ID      string   `json:"id,omitempty"`  // ID uniquely identifying the content
	Episode int      `json:"eps,omitempty"` // Episode number (typically applies to video content).
	Title   string   `json:"ttl,omitempty"` // Content title.
	Series  string   `json:"srs,omitempty"` // Content series
	Season  string   `json:"ssn,omitempty"` // Content season
	Cat     []string `json:"cat,omitempty"` // Array of IAB content categories that describe the content producer
}

// AdPodSlot of adpod object logging
type AdPodSlot struct {
	MinAds                      *int `json:"mnad,omitempty"` //Default 1 if not specified
	MaxAds                      *int `json:"mxad,omitempty"` //Default 1 if not specified
	MinDuration                 *int `json:"amnd,omitempty"` // (adpod.adminduration * adpod.minads) should be greater than or equal to video.minduration
	MaxDuration                 *int `json:"amxd,omitempty"` // (adpod.admaxduration * adpod.maxads) should be less than or equal to video.maxduration + video.maxextended
	AdvertiserExclusionPercent  *int `json:"exap,omitempty"` // Percent value 0 means none of the ads can be from same advertiser 100 means can have all same advertisers
	IABCategoryExclusionPercent *int `json:"exip,omitempty"` // Percent value 0 means all ads should be of different IAB categories.
}

// SlotRecord structure for storing slot level information
type SlotRecord struct {
	SlotName          string          `json:"sn,omitempty"`
	SlotSize          []string        `json:"sz,omitempty"`
	Adunit            string          `json:"au,omitempty"`
	AdPodSlot         *AdPodSlot      `json:"aps,omitempty"`
	PartnerData       []PartnerRecord `json:"ps"`
	RewardedInventory int             `json:"rwrd,omitempty"` // Indicates if the ad slot was enabled (rwrd=1) for rewarded or disabled (rwrd=0)
}

// PartnerRecord structure for storing partner information
type PartnerRecord struct {
	PartnerID            string  `json:"pn"`
	BidderCode           string  `json:"bc"`
	KGPV                 string  `json:"kgpv"`  // In case of Regex mapping, this will contain the regex string.
	KGPSV                string  `json:"kgpsv"` // In case of Regex mapping, this will contain the actual slot name that matched the regex.
	PartnerSize          string  `json:"psz"`   //wxh
	Adformat             string  `json:"af"`
	GrossECPM            float64 `json:"eg"`
	NetECPM              float64 `json:"en"`
	Latency1             int     `json:"l1"` //response time
	Latency2             int     `json:"l2"`
	PostTimeoutBidStatus int     `json:"t"`
	WinningBidStaus      int     `json:"wb"`
	BidID                string  `json:"bidid"`
	OrigBidID            string  `json:"origbidid"`
	DealID               string  `json:"di"`
	DealChannel          string  `json:"dc"`
	DealPriority         int     `json:"dp,omitempty"`
	DefaultBidStatus     int     `json:"db"`
	ServerSide           int     `json:"ss"`
	MatchedImpression    int     `json:"mi"`

	//AdPod Specific
	AdPodSequenceNumber *int     `json:"adsq,omitempty"`
	AdDuration          *int     `json:"dur,omitempty"`
	ADomain             string   `json:"adv,omitempty"`
	Cat                 []string `json:"cat,omitempty"`
	NoBidReason         *int     `json:"aprc,omitempty"`

	OriginalCPM float64 `json:"ocpm"`
	OriginalCur string  `json:"ocry"`

	MetaData *MetaData `json:"md,omitempty"`

	FloorValue     float64 `json:"fv,omitempty"`
	FloorRule      string  `json:"fr,omitempty"`
	FloorRuleValue float64 `json:"frv,omitempty"`
}

type MetaData struct {
	NetworkID            int             `json:"nwid,omitempty"`
	AdvertiserID         int             `json:"adid,omitempty"`
	NetworkName          string          `json:"nwnm,omitempty"`
	PrimaryCategoryID    string          `json:"pcid,omitempty"`
	AdvertiserName       string          `json:"adnm,omitempty"`
	AgencyID             int             `json:"agid,omitempty"`
	AgencyName           string          `json:"agnm,omitempty"`
	BrandID              int             `json:"brid,omitempty"`
	BrandName            string          `json:"brnm,omitempty"`
	DChain               json.RawMessage `json:"dc,omitempty"`
	DemandSource         string          `json:"ds,omitempty"`
	SecondaryCategoryIDs []string        `json:"secondaryCatIds,omitempty"`
}

// logDeviceObject will be used to log device specific parameters like platform and ifa_type
func (wlog *WloggerRecord) logDeviceObject(rctx models.RequestCtx, uaFromHTTPReq string, ortbBidRequest *openrtb2.BidRequest, platform string) {
	dvc := Device{
		Platform: rctx.DevicePlatform,
	}

	if ortbBidRequest != nil && ortbBidRequest.Device != nil && ortbBidRequest.Device.Ext != nil {
		ext := make(map[string]interface{})
		err := json.Unmarshal(ortbBidRequest.Device.Ext, &ext)
		if err != nil {
			return

		}
		// if ext, ok := ortbBidRequest.Device.Ext.(map[string]interface{}); ok {
		//use ext object for logging any other extension parameters

		//log device.ext.ifa_type parameter to ifty in logger record
		if value, ok := ext["ifa_type"].(string); ok {

			//ifa_type checkking is valid parameter and log its respective id
			ifaType := models.DeviceIFATypeID[strings.ToLower(value)]
			dvc.IFAType = &ifaType
		}
		// }
	}

	//settind device object
	wlog.Device = dvc
}

// Round value to 2 digit
func roundToTwoDigit(value float64) float64 {
	output := math.Pow(10, float64(2))
	return float64(math.Round(value*output)) / output
}
