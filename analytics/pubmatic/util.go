package pubmatic

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

func GetValueFromRequest(request *openrtb2.BidRequest, key string) interface{} {
	switch key {

	case models.UserAgent:
		if nil != request.Device {
			return request.Device.UA
		}

	case models.StoreURL:
		if request.App != nil {
			return request.App.StoreURL
		} else if request.Site != nil {
			return request.Site.Page
		}

	case models.Origin:
		if request.App != nil {
			return request.App.Bundle
		} else if request.Site != nil {

			if request.Site.Domain != "" {
				return request.Site.Domain
			}

			if request.Site.Page != "" {

				hostname := ""
				pageURL, err := url.Parse(request.Site.Page)
				if err == nil && pageURL != nil {
					hostname = pageURL.Host
				}

				return hostname
			}
		}
	case models.IP:
		if nil != request.Device && "" != request.Device.IP {
			return request.Device.IP
		}

	case models.Consent:
		if IsUserExtPresent(request.User) {
			var userExt openrtb_ext.ExtUser
			err := json.Unmarshal(request.User.Ext, &userExt)
			// userExt, ok := request.User.Ext.(*openrtb.ExtUser)
			if err != nil && userExt.Consent != "" {
				return userExt.Consent
			}
			//AAA : verify this
		}

	case models.GDPR:
		if IsRegsExtPresent(request.Regs) {

			var regExt openrtb_ext.ExtRegs
			err := json.Unmarshal(request.Regs.Ext, &regExt)
			// userExt, ok := request.User.Ext.(*openrtb.ExtUser)
			if err != nil && regExt.GDPR != nil {
				return *regExt.GDPR
			}
		}

	case models.PublisherID:
		if request.App != nil && request.App.Publisher != nil {
			return request.App.Publisher.ID
		} else if request.Site != nil && request.Site.Publisher != nil {
			return request.Site.Publisher.ID
		}
	}
	return nil

}

func IsUserExtPresent(user *openrtb2.User) bool {
	if user != nil && user.Ext != nil {
		return true
	}
	return false
}

func IsRegsExtPresent(reg *openrtb2.Regs) bool {
	if reg != nil && reg.Ext != nil {
		return true
	}
	return false
}

func getValue(oRTBParamName string, values url.Values, redirectQueryParams url.Values, DFPParamName string, OWParamName string) interface{} {
	paramArr := models.ORTBToDFPOWMap[oRTBParamName]
	if paramArr == nil {
		return nil
	}

	if values.Get(OWParamName) != "" {
		return values.Get(OWParamName)
	} else if paramArr[1] != "" && DFPParamName != "" && redirectQueryParams.Get(DFPParamName) != "" {
		return redirectQueryParams.Get(DFPParamName)
	}

	return nil
}

func getValueInArray(oRTBParamName string, values url.Values, redirectQueryParams url.Values, DFPParamName string, OWParamName string) interface{} {
	valStr := GetString(getValue(oRTBParamName, values, redirectQueryParams, DFPParamName, OWParamName))
	if valStr != "" {
		valIntArr := make([]int, 0)
		for _, val := range strings.Split(valStr, ",") {
			valInt, _ := strconv.Atoi(val)
			valIntArr = append(valIntArr, valInt)
		}
		return valIntArr
	}
	return nil
}

func GetString(val interface{}) string {
	var result string
	if val != nil {
		result, ok := val.(string)
		if ok {
			return result
		}
	}
	return result
}

func GetInt(val interface{}) int {
	var result int
	if val != nil {
		result, ok := val.(int)
		if ok {
			return result
		}
	}
	return result
}

// ExtRequest Request Extension
type ExtRequest struct {
	Wrapper *ExtRequestWrapper                `json:"wrapper,omitempty"`
	Bidder  map[string]map[string]interface{} `json:"bidder,omitempty"`
	AdPod   *ExtRequestAdPod                  `json:"adpod,omitempty"`
	Prebid  *ExtRequestPrebid                 `json:"prebid"`
}

// ExtRequestWrapper holds wrapper specific extension parameters
type ExtRequestWrapper struct {
	ProfileId            *int    `json:"profileid,omitempty"`
	VersionId            *int    `json:"versionid,omitempty"`
	SSAuctionFlag        *int    `json:"ssauction,omitempty"`
	SumryDisableFlag     *int    `json:"sumry_disable,omitempty"`
	ClientConfigFlag     *int    `json:"clientconfig,omitempty"`
	LogInfoFlag          *int    `json:"loginfo,omitempty"`
	SupportDeals         bool    `json:"supportdeals,omitempty"`
	IncludeBrandCategory *int    `json:"includebrandcategory,omitempty"`
	ABTestConfig         *int    `json:"abtest,omitempty"`
	LoggerImpressionID   *string `json:"wiid,omitempty"`
	SSAI                 *string `json:"ssai,omitempty"`
}

// ExtRequestPrebid defines the contract for bidrequest.ext.prebid
type ExtRequestPrebid struct {
	Aliases              interface{} `json:"aliases,omitempty"`
	BidAdjustmentFactors interface{} `json:"bidadjustmentfactors,omitempty"`
	Cache                interface{} `json:"cache,omitempty"`
	Data                 interface{} `json:"data,omitempty"`
	Debug                bool        `json:"debug,omitempty"`
	Events               interface{} `json:"events,omitempty"`
	SChains              interface{} `json:"schains,omitempty"`
	StoredRequest        interface{} `json:"storedrequest,omitempty"`
	SupportDeals         bool        `json:"supportdeals,omitempty"`
	Targeting            interface{} `json:"targeting,omitempty"`

	// NoSale specifies bidders with whom the publisher has a legal relationship where the
	// passing of personally identifiable information doesn't constitute a sale per CCPA law.
	// The array may contain a single sstar ('*') entry to represent all bidders.
	NoSale       []string                     `json:"nosale,omitempty"`
	Transparency *ExtTransparency             `json:"transparency,omitempty"`
	Floors       *openrtb_ext.PriceFloorRules `json:"floors,omitempty"`

	AlternateBidderCodes *openrtb_ext.ExtAlternateBidderCodes `json:"alternatebiddercodes,omitempty"`
	Channel              *openrtb_ext.ExtRequestPrebidChannel `json:"channel,omitempty"`
	ReturnAllBidStatus   bool                                 `json:"returnallbidstatus,omitempty"`
}

// ExtTransparency holds bidder level content transparency rules
type ExtTransparency struct {
	Content map[string]TransparencyRule `json:"content,omitempty"`
}

// TransparencyRule contains transperancy rule for a single bidder
type TransparencyRule struct {
	Include bool     `json:"include,omitempty"`
	Keys    []string `json:"keys,omitempty"`
}

// ExtRequestAdPod holds AdPod specific extension parameters at request level
type ExtRequestAdPod struct {
	AdPod
	CrossPodAdvertiserExclusionPercent  *int `json:"crosspodexcladv,omitempty"`    //Percent Value - Across multiple impression there will be no ads from same advertiser. Note: These cross pod rule % values can not be more restrictive than per pod
	CrossPodIABCategoryExclusionPercent *int `json:"crosspodexcliabcat,omitempty"` //Percent Value - Across multiple impression there will be no ads from same advertiser
	IABCategoryExclusionWindow          *int `json:"excliabcatwindow,omitempty"`   //Duration in minute between pods where exclusive IAB rule needs to be applied
	AdvertiserExclusionWindow           *int `json:"excladvwindow,omitempty"`      //Duration in minute between pods where exclusive advertiser rule needs to be applied
}

// AdPod holds Video AdPod specific extension parameters at impression level
type AdPod struct {
	MinAds                      *int `json:"minads,omitempty"`        //Default 1 if not specified
	MaxAds                      *int `json:"maxads,omitempty"`        //Default 1 if not specified
	MinDuration                 *int `json:"adminduration,omitempty"` // (adpod.adminduration * adpod.minads) should be greater than or equal to video.minduration
	MaxDuration                 *int `json:"admaxduration,omitempty"` // (adpod.admaxduration * adpod.maxads) should be less than or equal to video.maxduration + video.maxextended
	AdvertiserExclusionPercent  *int `json:"excladv,omitempty"`       // Percent value 0 means none of the ads can be from same advertiser 100 means can have all same advertisers
	IABCategoryExclusionPercent *int `json:"excliabcat,omitempty"`    // Percent value 0 means all ads should be of different IAB categories.
}

// BidResponseAdPodExt defines the prebid adpod response in bidresponse.ext.adpod parameter
type BidResponseAdPodExt struct {
	Response *openrtb2.BidResponse      `json:"bidresponse,omitempty"`
	Config   map[string]*AdPodImpConfig `json:"config,omitempty"`
}

// AdPodImpConfig example
type AdPodImpConfig struct {
	//AdPodGenerator
	VideoExt        *ExtVideo           `json:"vidext,omitempty"`
	Config          []*ImpAdPodConfig   `json:"imp,omitempty"`
	BlockedVASTTags map[string][]string `json:"blockedtags,omitempty"`
	Error           *ExtBidderMessage   `json:"ec,omitempty"`
}

// ExtVideo structure to accept video specific more parameters like adpod
type ExtVideo struct {
	Offset *int   `json:"offset,omitempty"` // Minutes from start where this ad is intended to show
	AdPod  *AdPod `json:"adpod,omitempty"`
}

// ImpAdPodConfig configuration for creating ads in adpod
type ImpAdPodConfig struct {
	ImpID          string `json:"id,omitempty"`
	SequenceNumber int8   `json:"seq,omitempty"`
	MinDuration    int64  `json:"minduration,omitempty"`
	MaxDuration    int64  `json:"maxduration,omitempty"`
}

// ExtBidderMessage defines an error object to be returned, consiting of a machine readable error code, and a human readable error message string.
type ExtBidderMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
