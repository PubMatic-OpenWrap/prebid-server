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
	"github.com/pkg/errors"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/usersync"
)

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

// GetAdFormat gets adformat from creative(adm) of the bid
func GetAdFormat(adm string) string {
	adFormat := Banner
	videoRegex, _ := regexp.Compile("<VAST\\s+")

	if videoRegex.MatchString(adm) {
		adFormat = Video
	} else {
		var admJSON map[string]interface{}
		err := json.Unmarshal([]byte(strings.Replace(adm, "/\\/g", "", -1)), &admJSON)
		if err == nil && admJSON != nil && admJSON["native"] != nil {
			adFormat = Native
		}
	}
	return adFormat
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
