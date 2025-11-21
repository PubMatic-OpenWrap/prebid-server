package ctvutils

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

/*
setIncludeBrandCategory sets PBS's  bidrequest.ext.prebid.Targeting object
 1. If pReqExt.supportDeals  = true then sets IncludeBrandCategory of targeting as follows
    WithCategory        = false
    TranslateCategories = false
*/
func SetIncludeBrandCategory(rCtx models.RequestCtx) {
	includeBrandCategory := &openrtb_ext.ExtIncludeBrandCategory{
		SkipDedup:           true,
		TranslateCategories: ptrutil.ToPtr(false),
	}

	if rCtx.NewReqExt.Wrapper != nil && rCtx.NewReqExt.Wrapper.IncludeBrandCategory != nil &&
		(models.IncludeIABBranchCategory == *rCtx.NewReqExt.Wrapper.IncludeBrandCategory ||
			models.IncludeAdServerBrandCategory == *rCtx.NewReqExt.Wrapper.IncludeBrandCategory) {

		includeBrandCategory.WithCategory = true

		if models.IncludeAdServerBrandCategory == *rCtx.NewReqExt.Wrapper.IncludeBrandCategory {
			adserver := models.GetVersionLevelPropertyFromPartnerConfig(rCtx.PartnerConfigMap, models.AdserverKey)
			prebidAdServer := getPrebidPrimaryAdServer(adserver)
			if prebidAdServer > 0 {
				includeBrandCategory.PrimaryAdServer = prebidAdServer
				includeBrandCategory.Publisher = getPrebidPublisher(adserver)
				*includeBrandCategory.TranslateCategories = true
			} else {
				includeBrandCategory.WithCategory = false
			}
		}
	}
	rCtx.NewReqExt.Prebid.Targeting.IncludeBrandCategory = includeBrandCategory
}

func getPrebidPrimaryAdServer(adserver string) int {
	//TODO: Make it map[OWPrimaryAdServer]PrebidPrimaryAdServer
	//1-Freewheel 2-DFP
	if models.OWPrimaryAdServerDFP == adserver {
		return models.PrebidPrimaryAdServerDFPID
	}
	return 0
}

func getPrebidPublisher(adserver string) string {
	//TODO: Make it map[OWPrimaryAdServer]PrebidPrimaryAdServer
	if models.OWPrimaryAdServerDFP == adserver {
		return models.PrebidPrimaryAdServerDFP
	}
	return ""
}

func ValidateVideoImpressions(request *openrtb2.BidRequest) error {
	if len(request.Imp) == 0 {
		return errors.New("recieved request with no impressions")
	}

	var validImpCount int
	for _, imp := range request.Imp {
		if imp.Video != nil {
			validImpCount++
		}
	}

	if validImpCount == 0 {
		return errors.New("video object is missing in the request")
	}

	return nil
}

// isValidSchain validated the schain object
func IsValidSchain(schain *openrtb2.SupplyChain) error {
	if schain.Ver != openrtb_ext.SChainVersion1 {
		return fmt.Errorf("invalid schain version, version should be %s", openrtb_ext.SChainVersion1)
	}

	if (int(schain.Complete) != openrtb_ext.SChainCompleteYes) && (schain.Complete != openrtb_ext.SChainCompleteNo) {
		return errors.New("invalid schain.complete value should be 0 or 1")
	}

	if len(schain.Nodes) == 0 {
		return errors.New("invalid schain node fields, Node can't be empty")
	}

	for _, schainNode := range schain.Nodes {
		if schainNode.ASI == "" {
			return errors.New("invalid schain node fields, ASI can't be empty")
		}

		if schainNode.SID == "" {
			return errors.New("invalid schain node fields, SID can't be empty")
		}

		if len([]rune(schainNode.SID)) > openrtb_ext.SIDLength {
			return errors.New("invalid schain node fields, sid can have maximum 64 characters")
		}

		// for schain version 1.0 hp must be 1
		if schainNode.HP == nil || *schainNode.HP != openrtb_ext.HPOne {
			return errors.New("invalid schain node fields, HP must be one")
		}
	}
	return nil
}

func FilterNonVideoImpressions(request *openrtb2.BidRequest) error {
	if request != nil && len(request.Imp) > 0 {
		j := 0
		for index, imp := range request.Imp {
			//Validate Native Impressions
			if imp.Video == nil {
				continue
			}

			//Banner Request Not Supported
			imp.Banner = nil

			//Native Request Not Supported
			imp.Native = nil

			if index != j {
				request.Imp[j] = imp
			}
			j++
		}
		request.Imp = request.Imp[:j]
		if len(request.Imp) == 0 {
			return fmt.Errorf("video object is missing for ctv request")
		}
	}
	return nil
}

// GetTargeting returns the value of targeting key associated with bidder
// it is expected that bid.Ext contains prebid.targeting map
// if value not present or any error occured empty value will be returned
// along with error.
func GetTargeting(key openrtb_ext.TargetingKey, bidder openrtb_ext.BidderName, bidCtx models.BidCtx, seq int) string {
	bidderKey := string(bidder)
	if seq > 1 {
		bidderKey = bidderKey + strconv.Itoa(seq)
	}

	bidderSpecificKey := key.BidderKey(models.DefaultTargetingKeyPrefix, openrtb_ext.BidderName(bidderKey), 20)
	return bidCtx.BidExt.Prebid.Targeting[bidderSpecificKey]
}

func AddTargetingKey(bidCtx models.BidCtx, key openrtb_ext.TargetingKey, value string) {
	key = models.DefaultTargetingKeyPrefix + key
	bidCtx.BidExt.Prebid.Targeting[string(key)] = value
}

func RemoveAdpodDataFromExt(bidrequest *openrtb2.BidRequest) {
	for i := range bidrequest.Imp {
		if bidrequest.Imp[i].Video == nil {
			continue
		}

		if bidrequest.Imp[i].Video.Ext == nil {
			continue
		}

		bidrequest.Imp[i].Video.Ext = jsonparser.Delete(bidrequest.Imp[i].Video.Ext, "adpod")
		bidrequest.Imp[i].Video.Ext = jsonparser.Delete(bidrequest.Imp[i].Video.Ext, "offset")
		if len(bidrequest.Imp[i].Video.Ext) == 0 || string(bidrequest.Imp[i].Video.Ext) == "{}" {
			bidrequest.Imp[i].Video.Ext = nil
		}
	}

	bidrequest.Ext = jsonparser.Delete(bidrequest.Ext, "adpod")
	bidrequest.Ext = jsonparser.Delete(bidrequest.Ext, "offset")
}

func AddPWTTargetingKeysForAdpod(rCtx models.RequestCtx, bid *openrtb2.Bid, seat string) {
	impCtx, ok := rCtx.ImpBidCtx[bid.ImpID]
	if !ok {
		return
	}

	bidCtx, ok := impCtx.BidCtx[bid.ID]
	if !ok {
		return
	}

	if bidCtx.Prebid == nil {
		bidCtx.Prebid = &openrtb_ext.ExtBidPrebid{}
	}

	if bidCtx.Prebid.Targeting == nil {
		bidCtx.Prebid.Targeting = make(map[string]string)
	}

	bidCtx.Prebid.Targeting[models.PWT_PARTNERID] = seat

	if bidCtx.Prebid != nil {
		if bidCtx.Prebid.Video != nil && bidCtx.Prebid.Video.Duration > 0 {
			bidCtx.Prebid.Targeting[models.PWT_DURATION] = strconv.Itoa(bidCtx.Prebid.Video.Duration)
		}

		partnerConfig := rCtx.PartnerConfigMap[models.VersionLevelConfigID]

		prefix, _, _, err := jsonparser.Get(impCtx.NewExt, "prebid", "bidder", seat, "dealtier", "prefix")
		if bidCtx.Prebid.DealTierSatisfied && partnerConfig[models.DealTierLineItemSetup] == "1" && err == nil && len(prefix) > 0 {
			bidCtx.Prebid.Targeting[models.PwtDT] = fmt.Sprintf("%s%d", string(prefix), bidCtx.Prebid.DealPriority)
		} else if len(bid.DealID) > 0 && partnerConfig[models.DealIDLineItemSetup] == "1" {
			bidCtx.Prebid.Targeting[models.PWT_DEALID] = bid.DealID
		} else {
			priceBucket, ok := bidCtx.Prebid.Targeting[models.PwtPb]
			if ok {
				bidCtx.Prebid.Targeting[models.PwtPb] = priceBucket
			}
		}

		catDur, ok := bidCtx.Prebid.Targeting[models.PwtPbCatDur]
		if ok {
			cat, dur := getCatAndDurFromPwtCatDur(catDur)
			if len(cat) > 0 {
				bidCtx.Prebid.Targeting[models.PwtCat] = cat
			}

			if len(dur) > 0 && bidCtx.Prebid.Targeting[models.PWT_DURATION] == "" {
				bidCtx.Prebid.Targeting[models.PWT_DURATION] = dur
			}

			// Remove support deals from targeting when support deals is false
			if !rCtx.SupportDeals {
				delete(bidCtx.Prebid.Targeting, models.PwtPbCatDur)
			}
		}
	}

	// Add targetting keys in case of debug
	if rCtx.Debug {
		bidCtx.Prebid.Targeting[models.PwtBidID] = utils.GetOriginalBidId(bid.ID)
		bidCtx.Prebid.Targeting[models.PWT_CACHE_PATH] = models.AMP_CACHE_PATH
		bidCtx.Prebid.Targeting[models.PWT_ECPM] = fmt.Sprintf("%.2f", bidCtx.NetECPM)
		bidCtx.Prebid.Targeting[models.PWT_PUBID] = rCtx.PubIDStr
		bidCtx.Prebid.Targeting[models.PWT_SLOTID] = impCtx.TagID
		bidCtx.Prebid.Targeting[models.PWT_PROFILEID] = rCtx.ProfileIDStr

		if bidCtx.Prebid.Targeting[models.PWT_ECPM] == "" {
			bidCtx.Prebid.Targeting[models.PWT_ECPM] = "0"
		}

		versionID := fmt.Sprint(rCtx.DisplayID)
		if versionID != "0" {
			bidCtx.Prebid.Targeting[models.PWT_VERSIONID] = versionID
		}

		if !rCtx.SupportDeals {
			delete(bidCtx.Prebid.Targeting, models.PwtPbCatDur)
		}
	}

	impCtx.BidCtx[bid.ID] = bidCtx
	rCtx.ImpBidCtx[bid.ImpID] = impCtx
}

func getCatAndDurFromPwtCatDur(pwtCatDur string) (string, string) {
	arr := strings.Split(pwtCatDur, "_")
	if len(arr) == 2 {
		return "", TrimRightByte(arr[1], 's')
	}
	if len(arr) == 3 {
		return arr[1], TrimRightByte(arr[2], 's')
	}
	return "", ""
}

func TrimRightByte(s string, b byte) string {
	if s[len(s)-1] == b {
		return s[:len(s)-1]
	}
	return s
}

// SetCORSHeaders sets CORS headers in response
func SetCORSHeaders(w http.ResponseWriter, header http.Header) {
	origin := header.Get("Origin")
	if len(origin) == 0 {
		origin = "*"
	} else {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
}
