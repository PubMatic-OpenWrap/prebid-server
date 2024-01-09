package models

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/usersync"
	"github.com/prebid/prebid-server/util/ptrutil"
)

var videoRegex *regexp.Regexp

func init() {
	videoRegex, _ = regexp.Compile("<VAST\\s+")
}

var SyncerMap map[string]usersync.Syncer

func SetSyncerMap(s map[string]usersync.Syncer) {
	SyncerMap = s
}

// IsCTVAPIRequest will return true if reqAPI is from CTV EndPoint
func IsCTVAPIRequest(api string) bool {
	return api == "/video/json" || api == "/video/vast" || api == "/video/openrtb"
}

func GetRequestExtWrapper(request []byte, wrapperLocation ...string) (RequestExtWrapper, error) {
	extWrapper := RequestExtWrapper{
		SSAuctionFlag: -1,
	}

	if len(wrapperLocation) == 0 {
		wrapperLocation = []string{"ext", "prebid", "bidderparams", "pubmatic", "wrapper"}
	}

	extWrapperBytes, _, _, err := jsonparser.Get(request, wrapperLocation...)
	if err != nil {
		return extWrapper, fmt.Errorf("request.ext.wrapper not found: %v", err)
	}

	err = json.Unmarshal(extWrapperBytes, &extWrapper)
	if err != nil {
		return extWrapper, fmt.Errorf("failed to decode request.ext.wrapper : %v", err)
	}

	return extWrapper, nil
}

func GetTest(request []byte) (int64, error) {
	test, err := jsonparser.GetInt(request, "test")
	if err != nil {
		return test, fmt.Errorf("request.test not found: %v", err)
	}
	return test, nil
}

func GetSize(width, height int64) string {
	return fmt.Sprintf("%dx%d", width, height)
}

// CreatePartnerKey returns key with partner appended
func CreatePartnerKey(partner, key string) string {
	if partner == "" {
		return key
	}
	return key + "_" + partner
}

// GetCreativeType gets adformat from creative(adm) of the bid
func GetCreativeType(bid *openrtb2.Bid, bidExt *BidExt, impCtx *ImpCtx) string {
	if bidExt.Prebid != nil && len(bidExt.Prebid.Type) > 0 {
		return string(bidExt.Prebid.Type)
	}
	if bid.AdM == "" {
		return ""
	}
	if videoRegex.MatchString(bid.AdM) {
		return Video
	}
	if impCtx.Native != nil {
		var admJSON map[string]interface{}
		err := json.Unmarshal([]byte(strings.Replace(bid.AdM, "/\\/g", "", -1)), &admJSON)
		if err == nil && admJSON != nil && admJSON["native"] != nil {
			return Native
		}
	}
	return Banner
}

func IsDefaultBid(bid *openrtb2.Bid) bool {
	return bid.Price == 0 && bid.DealID == "" && bid.W == 0 && bid.H == 0
}

// GetAdFormat returns adformat of the bid.
// for default bid it refers to impression object
// for non-default bids it uses creative(adm) of the bid
func GetAdFormat(bid *openrtb2.Bid, bidExt *BidExt, impCtx *ImpCtx) string {
	if bid == nil || impCtx == nil {
		return ""
	}
	if IsDefaultBid(bid) {
		if impCtx.Banner {
			return Banner
		}
		if impCtx.Video != nil {
			return Video
		}
		if impCtx.Native != nil {
			return Native
		}
		return ""
	}
	if bidExt == nil {
		return ""
	}
	return GetCreativeType(bid, bidExt, impCtx)
}

func GetRevenueShare(partnerConfig map[string]string) float64 {
	var revShare float64

	if val, ok := partnerConfig[REVSHARE]; ok {
		revShare, _ = strconv.ParseFloat(val, 64)
	}
	return revShare
}

func GetNetEcpm(price float64, revShare float64) float64 {
	if revShare == 0 {
		return toFixed(price, BID_PRECISION)
	}
	price = price * (1 - revShare/100)
	return toFixed(price, BID_PRECISION)
}

func GetGrossEcpm(price float64) float64 {
	return toFixed(price, BID_PRECISION)
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ExtractDomain(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "http://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	return u.Host, nil
}

// hybrid/web request would have bidder params prepopulated.
// TODO: refer request.ext.prebid.channel.name = pbjs instead?
func IsHybrid(body []byte) bool {
	_, _, _, err := jsonparser.Get(body, "imp", "[0]", "ext", "prebid", "bidder", "pubmatic")
	return err == nil
}

// GetVersionLevelPropertyFromPartnerConfig returns a Version level property from the partner config map
func GetVersionLevelPropertyFromPartnerConfig(partnerConfigMap map[int]map[string]string, propertyName string) string {
	if versionLevelConfig, ok := partnerConfigMap[VersionLevelConfigID]; ok && versionLevelConfig != nil {
		return versionLevelConfig[propertyName]
	}
	return ""
}

const (
	//The following are the headerds related to IP address
	XForwarded      = "X-FORWARDED-FOR"
	SourceIP        = "SOURCE_IP"
	ClusterClientIP = "X_CLUSTER_CLIENT_IP"
	RemoteAddr      = "REMOTE_ADDR"
	RlnClientIP     = "RLNCLIENTIPADDR"
)

func GetIP(in *http.Request) string {
	//The IP address priority is as follows:
	//0. HTTP_RLNCLIENTIPADDR  //For API
	//1. HTTP_X_FORWARDED_IP
	//2. HTTP_X_CLUSTER_CLIENT_IP
	//3. HTTP_SOURCE_IP
	//4. REMOTE_ADDR
	glog.Info("logging... request header OTT-908:")
	glog.Info(RlnClientIP+"="+in.Header.Get(RlnClientIP), SourceIP+"="+in.Header.Get(SourceIP), ClusterClientIP+"="+in.Header.Get(ClusterClientIP), XForwarded+"="+in.Header.Get(XForwarded), RemoteAddr+"="+in.Header.Get(RemoteAddr), "in.remote"+"="+in.Header.Get(in.RemoteAddr))
	ip := in.Header.Get(RlnClientIP)
	if ip == "" {
		ip = in.Header.Get(SourceIP)
		if ip == "" {
			ip = in.Header.Get(ClusterClientIP)
			if ip == "" {
				ip = in.Header.Get(XForwarded)
				if ip == "" {
					//RemoteAddr parses the header REMOTE_ADDR
					ip = in.Header.Get(RemoteAddr)
					if ip == "" {
						ip, _, _ = net.SplitHostPort(in.RemoteAddr)
					}
				}
			}
		}
	}
	ips := strings.Split(ip, ",")
	if len(ips) != 0 {
		return ips[0]
	}

	return ""
}

func Atof(value string, decimalplaces int) (float64, error) {

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	value = fmt.Sprintf("%."+strconv.Itoa(decimalplaces)+"f", floatValue)
	floatValue, err = strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	return floatValue, nil
}

// IsPubmaticCorePartner returns true when the partner is pubmatic or internally an alias of pubmatic
func IsPubmaticCorePartner(partnerName string) bool {
	if partnerName == string(openrtb_ext.BidderPubmatic) || partnerName == BidderPubMaticSecondaryAlias {
		return true
	}
	return false
}

// wraps error with error msg
func ErrorWrap(cErr, nErr error) error {
	if cErr == nil {
		return nErr
	}

	return errors.Wrap(cErr, nErr.Error())
}

func GetSizeForPlatform(width, height int64, platform string) string {
	s := fmt.Sprintf("%dx%d", width, height)
	if platform == PLATFORM_VIDEO {
		s = s + VideoSizeSuffix
	}
	return s
}

func GetKGPSV(bid openrtb2.Bid, bidderMeta PartnerData, adformat string, tagId string, div string, source string) (string, string) {
	kgpv := bidderMeta.KGPV
	kgpsv := bidderMeta.MatchedSlot
	isRegex := bidderMeta.IsRegex
	// 1. nobid
	if IsDefaultBid(&bid) {
		//NOTE: kgpsv = bidderMeta.MatchedSlot above. Use the same
		if !isRegex && kgpv != "" { // unmapped pubmatic's slot
			kgpsv = kgpv
		} else if !isRegex {
			kgpv = kgpsv
		}
	} else if !isRegex {
		if kgpv != "" { // unmapped pubmatic's slot
			kgpsv = kgpv
		} else if adformat == Video { // Check when adformat is video, bid.W and bid.H has to be zero with Price !=0. Ex: UOE-9222(0x0 default kgpv and kgpsv for video bid)
			// 2. valid video bid
			// kgpv has regex, do not generate slotName again
			// kgpsv could be unmapped or mapped slot, generate slotName with bid.W = bid.H = 0
			kgpsv = GenerateSlotName(0, 0, bidderMeta.KGP, tagId, div, source)
			kgpv = kgpsv // original /43743431/DMDemo1234@300x250 but new could be /43743431/DMDemo1234@0x0
		} else if bid.H != 0 && bid.W != 0 { // Check when bid.H and bid.W will be zero with Price !=0. Ex: MobileInApp-MultiFormat-OnlyBannerMapping_Criteo_Partner_Validaton
			// 3. valid bid
			// kgpv has regex, do not generate slotName again
			// kgpsv could be unmapped or mapped slot, generate slotName again based on bid.H and bid.W
			kgpsv = GenerateSlotName(bid.H, bid.W, bidderMeta.KGP, tagId, div, source)
			kgpv = kgpsv
		}
	}
	if kgpv == "" {
		kgpv = kgpsv
	}
	return kgpv, kgpsv
}

// Harcode would be the optimal. We could make it configurable like _AU_@_W_x_H_:%s@%dx%d entries in pbs.yaml
// mysql> SELECT DISTINCT key_gen_pattern FROM wrapper_mapping_template;
// +----------------------+
// | key_gen_pattern      |
// +----------------------+
// | _AU_@_W_x_H_         |
// | _DIV_@_W_x_H_        |
// | _W_x_H_@_W_x_H_      |
// | _DIV_                |
// | _AU_@_DIV_@_W_x_H_   |
// | _AU_@_SRC_@_VASTTAG_ |
// +----------------------+
// 6 rows in set (0.21 sec)
func GenerateSlotName(h, w int64, kgp, tagid, div, src string) string {
	// func (H, W, Div), no need to validate, will always be non-nil
	switch kgp {
	case "_AU_": // adunitconfig
		return tagid
	case "_DIV_":
		return div
	case "_AU_@_W_x_H_":
		return fmt.Sprintf("%s@%dx%d", tagid, w, h)
	case "_DIV_@_W_x_H_":
		return fmt.Sprintf("%s@%dx%d", div, w, h)
	case "_W_x_H_@_W_x_H_":
		return fmt.Sprintf("%dx%d@%dx%d", w, h, w, h)
	case "_AU_@_DIV_@_W_x_H_":
		return fmt.Sprintf("%s@%s@%dx%d", tagid, div, w, h)
	case "_AU_@_SRC_@_VASTTAG_":
		return fmt.Sprintf("%s@%s@_VASTTAG_", tagid, src) //TODO check where/how _VASTTAG_ is updated
	default:
		// TODO: check if we need to fallback to old generic flow (below)
		// Add this cases in a map and read it from yaml file
	}
	return ""
}

func RoundToTwoDigit(value float64) float64 {
	output := math.Pow(10, float64(2))
	return float64(math.Round(value*output)) / output
}

// GetBidLevelFloorsDetails return floorvalue and floorrulevalue
func GetBidLevelFloorsDetails(bidExt BidExt, impCtx ImpCtx,
	currencyConversion func(from, to string, value float64) (float64, error)) (fv, frv float64) {
	var floorCurrency string
	frv = NotSet

	if bidExt.Prebid != nil && bidExt.Prebid.Floors != nil {
		floorCurrency = bidExt.Prebid.Floors.FloorCurrency
		fv = RoundToTwoDigit(bidExt.Prebid.Floors.FloorValue)
		frv = fv
		if bidExt.Prebid.Floors.FloorRuleValue > 0 {
			frv = RoundToTwoDigit(bidExt.Prebid.Floors.FloorRuleValue)
		}
	}

	// if floor values are not set from bid.ext then fall back to imp.bidfloor
	if frv == NotSet && impCtx.BidFloor != 0 {
		fv = RoundToTwoDigit(impCtx.BidFloor)
		frv = fv
		floorCurrency = impCtx.BidFloorCur
	}

	// convert the floor values in USD currency
	if floorCurrency != "" && floorCurrency != USD {
		value, _ := currencyConversion(floorCurrency, USD, fv)
		fv = RoundToTwoDigit(value)
		value, _ = currencyConversion(floorCurrency, USD, frv)
		frv = RoundToTwoDigit(value)
	}

	if frv == NotSet {
		frv = 0 // set it back to 0
	}

	return
}

// GetFloorsDetails returns floors details from response.ext.prebid
func GetFloorsDetails(responseExt openrtb_ext.ExtBidResponse) (floorDetails FloorsDetails) {
	if responseExt.Prebid != nil && responseExt.Prebid.Floors != nil {
		floors := responseExt.Prebid.Floors
		if floors.Skipped != nil {
			floorDetails.Skipfloors = ptrutil.ToPtr(0)
			if *floors.Skipped {
				floorDetails.Skipfloors = ptrutil.ToPtr(1)
			}
		}
		if floors.Data != nil && len(floors.Data.ModelGroups) > 0 {
			floorDetails.FloorModelVersion = floors.Data.ModelGroups[0].ModelVersion
		}
		if len(floors.PriceFloorLocation) > 0 {
			if source, ok := FloorSourceMap[floors.PriceFloorLocation]; ok {
				floorDetails.FloorSource = &source
			}
		}
		if status, ok := FetchStatusMap[floors.FetchStatus]; ok {
			floorDetails.FloorFetchStatus = &status
		}
		floorDetails.FloorProvider = floors.FloorProvider
		if floors.Data != nil && len(floors.Data.FloorProvider) > 0 {
			floorDetails.FloorProvider = floors.Data.FloorProvider
		}
		if floors.Enforcement != nil && floors.Enforcement.EnforcePBS != nil && *floors.Enforcement.EnforcePBS {
			floorDetails.FloorType = HardFloor
		}
	}
	return floorDetails
}
