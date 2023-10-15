package pubmatic

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// Send method
func Send(url string, headers http.Header) error {
	mhc := NewMultiHttpContext()
	hc, err := NewHttpCall(url, "")
	if err != nil {
		return err
	}

	for k, v := range headers {
		if len(v) != 0 {
			hc.AddHeader(k, v[0])
		}
	}

	mhc.AddHttpCall(hc)
	_, erc := mhc.Execute()
	if erc != 0 {
		return errors.New("error in sending logger pixel")
	}

	return nil
}

// PrepareLoggerURL returns the url for OW logger call
func PrepareLoggerURL(wlog *WloggerRecord, loggerURL string, gdprEnabled int) string {
	v := url.Values{}

	jsonString, err := json.Marshal(wlog.record)
	if err != nil {
		return ""
	}

	v.Set(models.WLJSON, string(jsonString))
	v.Set(models.WLPUBID, strconv.Itoa(wlog.PubID))
	if gdprEnabled == 1 {
		v.Set(models.WLGDPR, strconv.Itoa(gdprEnabled))
	}
	queryString := v.Encode()

	finalLoggerURL := loggerURL + "?" + queryString
	return finalLoggerURL
}

func getSizeForPlatform(width, height int64, platform string) string {
	s := models.GetSize(width, height)
	if platform == models.PLATFORM_VIDEO {
		s = s + models.VideoSizeSuffix
	}
	return s
}

// set partnerRecord MetaData
func (partnerRecord *PartnerRecord) setMetaDataObject(meta *openrtb_ext.ExtBidPrebidMeta) {

	if meta.NetworkID != 0 || meta.AdvertiserID != 0 || len(meta.SecondaryCategoryIDs) > 0 {
		partnerRecord.MetaData = &MetaData{
			NetworkID:            meta.NetworkID,
			AdvertiserID:         meta.AdvertiserID,
			PrimaryCategoryID:    meta.PrimaryCategoryID,
			AgencyID:             meta.AgencyID,
			DemandSource:         meta.DemandSource,
			SecondaryCategoryIDs: meta.SecondaryCategoryIDs,
		}
	}
	//NOTE : We Don't get following Data points in Response, whenever got from translator,
	//they can be populated.
	//partnerRecord.MetaData.NetworkName = meta.NetworkName
	//partnerRecord.MetaData.AdvertiserName = meta.AdvertiserName
	//partnerRecord.MetaData.AgencyName = meta.AgencyName
	//partnerRecord.MetaData.BrandName = meta.BrandName
	//partnerRecord.MetaData.BrandID = meta.BrandID
	//partnerRecord.MetaData.DChain = meta.DChain (type is json.RawMessage)
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
		if div == "" {
			return fmt.Sprintf("%s@%s@s%dx%d", tagid, div, w, h)
		}
		return fmt.Sprintf("%s@%s@s%dx%d", tagid, div, w, h)
	case "_AU_@_SRC_@_VASTTAG_":
		return fmt.Sprintf("%s@%s@s_VASTTAG_", tagid, src) //TODO check where/how _VASTTAG_ is updated
	default:
		// TODO: check if we need to fallback to old generic flow (below)
		// Add this cases in a map and read it from yaml file
	}
	return ""
}
