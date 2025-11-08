package ctv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"

	v26 "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/endpoints/legacy/openrtb/v26"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	uuid "github.com/satori/go.uuid"
)

type OpenRTB struct {
	request *http.Request
	values  URLValues
	ortb    map[string]interface{}
}

// NewOpenRTB Returns New ORTB Object of Version 2.5
func NewOpenRTB(request *http.Request) Parser {
	request.ParseForm()

	obj := &OpenRTB{
		request: request,
		values:  URLValues{Values: request.Form},
		ortb: map[string]interface{}{
			"imp": []map[string]interface{}{
				{
					"video": map[string]interface{}{},
					"ext":   map[string]interface{}{},
				},
			},
			"ext": map[string]interface{}{},
		},
	}

	return obj
}

/********************** Helper Functions **********************/

// ParseORTBRequest this will parse ortb request by reading parserMap and calling respective function for mapped parameter
func (o *OpenRTB) ParseORTBRequest(parserMap *ParserMap) (map[string]interface{}, error) {
	var errs []error
	for k, value := range o.values.Values {
		if len(value) > 0 && len(value[0]) > 0 {
			if parser, ok := parserMap.KeyMapping[k]; ok {
				if err := parser(o); err != nil {
					errs = append(errs, err)
				}
			} else {
				//Check for Ext
				extIndex := strings.Index(k, Ext)
				if extIndex != -1 {
					parentKey := k[:extIndex+ExtLen-1]
					childKey := k[extIndex+ExtLen:]
					if len(childKey) > 0 {
						if parser, ok := parserMap.ExtMapping[parentKey]; ok {
							if err := parser(o, childKey, o.values.GetStringPtr(k)); err != nil {
								errs = append(errs, err)
							}
						}
					}
				} else if _, ok = parserMap.IgnoreList[k]; !ok {
					glog.Warningf("Key Not Present : Key:[%v] Value:[%v]", k, value)
				}
			}
		}
	}

	if len(errs) > 0 {
		return o.ortb, fmt.Errorf("%v", errs)
	}

	o.formORTBRequest()
	return o.ortb, nil
}

// formORTBRequest this will generate bidrequestID or impressionID if not present
func (o *OpenRTB) formORTBRequest() {
	if o.ortb == nil {
		return
	}

	reqId, ok := o.ortb["id"].(string)
	if !ok || len(reqId) == 0 {
		o.ortb["id"] = uuid.NewV4().String()
	}

	imps, ok := o.ortb["imp"].([]interface{})
	if len(imps) == 0 {
		return
	}

	if len(imps[0].(map[string]interface{})["id"].(string)) == 0 {
		imps[0].(map[string]interface{})["id"] = uuid.NewV4().String()
		o.ortb["imp"] = imps
	}
}

/*********************** BidRequest ***********************/

// ORTBBidRequestID will read and set ortb BidRequest.ID parameter
func (o *OpenRTB) ORTBBidRequestID() (err error) {
	val := o.values.Get(ORTBBidRequestID)
	if len(val) == 0 {
		o.ortb["id"] = uuid.NewV4().String()
	} else {
		o.ortb["id"] = val
	}
	return
}

// ORTBBidRequestTest will read and set ortb BidRequest.Test parameter
func (o *OpenRTB) ORTBBidRequestTest() (err error) {
	val := o.values.Get(ORTBBidRequestTest)
	if len(val) > 0 {
		o.ortb["test"] = val
	}
	return
}

// ORTBBidRequestAt will read and set ortb BidRequest.At parameter
func (o *OpenRTB) ORTBBidRequestAt() (err error) {
	val := o.values.Get(ORTBBidRequestAt)
	if len(val) > 0 {
		o.ortb["at"] = val
	}
	return
}

// ORTBBidRequestTmax will read and set ortb BidRequest.Tmax parameter
func (o *OpenRTB) ORTBBidRequestTmax() (err error) {
	val := o.values.Get(ORTBBidRequestTmax)
	if len(val) > 0 {
		o.ortb["tmax"] = val
	}
	return
}

// ORTBBidRequestWseat will read and set ortb BidRequest.Wseat parameter
func (o *OpenRTB) ORTBBidRequestWseat() (err error) {
	o.ortb["wseat"] = o.values.GetStringArray(ORTBBidRequestWseat, ArraySeparator)
	return
}

// ORTBBidRequestWlang will read and set ortb BidRequest.Wlang Parameter
func (o *OpenRTB) ORTBBidRequestWlang() (err error) {
	o.ortb["wlang"] = o.values.GetStringArray(ORTBBidRequestWlang, ArraySeparator)
	return
}

// ORTBBidRequestBseat will read and set ortb BidRequest.Bseat Parameter
func (o *OpenRTB) ORTBBidRequestBseat() (err error) {
	o.ortb["bseat"] = o.values.GetStringArray(ORTBBidRequestBseat, ArraySeparator)
	return
}

// ORTBBidRequestAllImps will read and set ortb BidRequest.AllImps parameter
func (o *OpenRTB) ORTBBidRequestAllImps() (err error) {
	val := o.values.Get(ORTBBidRequestAllImps)
	if len(val) > 0 {
		o.ortb["allimps"] = val
	}
	return
}

// ORTBBidRequestCur will read and set ortb BidRequest.Cur parameter
func (o *OpenRTB) ORTBBidRequestCur() (err error) {
	o.ortb["cur"] = o.values.GetStringArray(ORTBBidRequestCur, ArraySeparator)
	return
}

// ORTBBidRequestBcat will read and set ortb BidRequest.Bcat parameter
func (o *OpenRTB) ORTBBidRequestBcat() (err error) {
	o.ortb["bcat"] = o.values.GetStringArray(ORTBBidRequestBcat, ArraySeparator)
	return
}

// ORTBBidRequestBadv will read and set ortb BidRequest.Badv parameter
func (o *OpenRTB) ORTBBidRequestBadv() (err error) {
	o.ortb["badv"] = o.values.GetStringArray(ORTBBidRequestBadv, ArraySeparator)
	return
}

// ORTBBidRequestBapp will read and set ortb BidRequest.Bapp parameter
func (o *OpenRTB) ORTBBidRequestBapp() (err error) {
	o.ortb["bapp"] = o.values.GetStringArray(ORTBBidRequestBapp, ArraySeparator)
	return
}

/*********************** Source ***********************/

// ORTBSourceFD will read and set ortb Source.FD parameter
func (o *OpenRTB) ORTBSourceFD() (err error) {
	val := o.values.Get(ORTBSourceFD)
	if len(val) == 0 {
		return
	}
	source, ok := o.ortb["source"].(map[string]interface{})
	if !ok {
		source = map[string]interface{}{}
	}
	source["fd"] = val
	o.ortb["source"] = source
	return
}

// ORTBSourceTID will read and set ortb Source.TID parameter
func (o *OpenRTB) ORTBSourceTID() (err error) {
	val := o.values.Get(ORTBSourceTID)
	if len(val) == 0 {
		return
	}
	source, ok := o.ortb["source"].(map[string]interface{})
	if !ok {
		source = map[string]interface{}{}
	}
	source["tid"] = val
	o.ortb["source"] = source
	return
}

// ORTBSourcePChain will read and set ortb Source.PChain parameter
func (o *OpenRTB) ORTBSourcePChain() (err error) {
	val := o.values.Get(ORTBSourcePChain)
	if len(val) == 0 {
		return
	}
	source, ok := o.ortb["source"].(map[string]interface{})
	if !ok {
		source = map[string]interface{}{}
	}
	source["pchain"] = val
	o.ortb["source"] = source
	return
}

// ORTBSourceSChain will read and set ortb Source.Ext.SChain parameter
func (o *OpenRTB) ORTBSourceSChain() (err error) {
	sChainString := o.values.Get(ORTBSourceSChain)
	if len(sChainString) == 0 {
		return
	}

	sChain, err := openrtb_ext.DeserializeSupplyChain(sChainString)
	if err != nil {
		pubId := ""
		if v, ok := o.values.GetString(ORTBAppPublisherID); ok {
			pubId = v
		} else if v, ok := o.values.GetString(ORTBSitePublisherID); ok {
			pubId = v
		}
		glog.Errorf(ErrDeserializationFailed, ORTBSourceSChain, err, pubId, sChainString)
		return nil
	}

	source, ok := o.ortb["source"].(map[string]interface{})
	if !ok {
		source = map[string]interface{}{}
	}
	source["schain"] = sChain
	o.ortb["source"] = source

	return
}

/*********************** Site ***********************/

// ORTBSiteID will read and set ortb Site.ID parameter
func (o *OpenRTB) ORTBSiteID() (err error) {
	val := o.values.Get(ORTBSiteID)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["id"] = val
	o.ortb["site"] = site
	return
}

// ORTBSiteName will read and set ortb Site.Name parameter
func (o *OpenRTB) ORTBSiteName() (err error) {
	val := o.values.Get(ORTBSiteName)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["name"] = val
	o.ortb["site"] = site
	return
}

// ORTBSiteDomain will read and set ortb Site.Domain parameter
func (o *OpenRTB) ORTBSiteDomain() (err error) {
	val := o.values.Get(ORTBSiteDomain)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["domain"] = val
	o.ortb["site"] = site
	return
}

// ORTBSitePage will read and set ortb Site.Page parameter
func (o *OpenRTB) ORTBSitePage() (err error) {
	val := o.values.Get(ORTBSitePage)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["page"] = val
	o.ortb["site"] = site
	return
}

// ORTBSiteRef will read and set ortb Site.Ref parameter
func (o *OpenRTB) ORTBSiteRef() (err error) {
	val := o.values.Get(ORTBSiteRef)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["ref"] = val
	o.ortb["site"] = site
	return
}

// ORTBSiteSearch will read and set ortb Site.Search parameter
func (o *OpenRTB) ORTBSiteSearch() (err error) {
	val := o.values.Get(ORTBSiteSearch)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["search"] = val
	o.ortb["site"] = site
	return
}

// ORTBSiteMobile will read and set ortb Site.Mobile parameter
func (o *OpenRTB) ORTBSiteMobile() (err error) {
	val := o.values.Get(ORTBSiteMobile)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["mobile"] = val
	o.ortb["site"] = site
	return
}

// ORTBSiteCat will read and set ortb Site.Cat parameter
func (o *OpenRTB) ORTBSiteCat() (err error) {
	val := o.values.GetStringArray(ORTBSiteCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["cat"] = val
	o.ortb["site"] = site
	return
}

// ORTBSiteSectionCat will read and set ortb Site.SectionCat parameter
func (o *OpenRTB) ORTBSiteSectionCat() (err error) {
	val := o.values.GetStringArray(ORTBSiteSectionCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["sectioncat"] = val
	o.ortb["site"] = site
	return
}

// ORTBSitePageCat will read and set ortb Site.PageCat parameter
func (o *OpenRTB) ORTBSitePageCat() (err error) {
	val := o.values.GetStringArray(ORTBSitePageCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["pagecat"] = val
	o.ortb["site"] = site
	return
}

// ORTBSitePrivacyPolicy will read and set ortb Site.PrivacyPolicy parameter
func (o *OpenRTB) ORTBSitePrivacyPolicy() (err error) {
	val := o.values.Get(ORTBSitePrivacyPolicy)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["privacypolicy"] = val
	o.ortb["site"] = site
	return
}

// ORTBSiteKeywords will read and set ortb Site.Keywords parameter
func (o *OpenRTB) ORTBSiteKeywords() (err error) {
	val := o.values.Get(ORTBSiteKeywords)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	site["keywords"] = val
	o.ortb["site"] = site
	return
}

/*********************** Site.Publisher ***********************/

// ORTBSitePublisherID will read and set ortb Site.Publisher.ID parameter
func (o *OpenRTB) ORTBSitePublisherID() (err error) {
	val := o.values.Get(ORTBSitePublisherID)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	publisher, ok := site["publisher"].(map[string]interface{})
	if !ok {
		publisher = map[string]interface{}{}
	}
	publisher["id"] = val
	site["publisher"] = publisher
	o.ortb["site"] = site
	return
}

// ORTBSitePublisherName will read and set ortb Site.Publisher.Name parameter
func (o *OpenRTB) ORTBSitePublisherName() (err error) {
	val := o.values.Get(ORTBSitePublisherName)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	publisher, ok := site["publisher"].(map[string]interface{})
	if !ok {
		publisher = map[string]interface{}{}
	}
	publisher["name"] = val
	site["publisher"] = publisher
	o.ortb["site"] = site
	return
}

// ORTBSitePublisherCat will read and set ortb Site.Publisher.Cat parameter
func (o *OpenRTB) ORTBSitePublisherCat() (err error) {
	val := o.values.GetStringArray(ORTBSitePublisherCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	publisher, ok := site["publisher"].(map[string]interface{})
	if !ok {
		publisher = map[string]interface{}{}
	}
	publisher["cat"] = val
	site["publisher"] = publisher
	o.ortb["site"] = site
	return
}

// ORTBSitePublisherDomain will read and set ortb Site.Publisher.Domain parameter
func (o *OpenRTB) ORTBSitePublisherDomain() (err error) {
	val := o.values.Get(ORTBSitePublisherDomain)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	publisher, ok := site["publisher"].(map[string]interface{})
	if !ok {
		publisher = map[string]interface{}{}
	}
	publisher["domain"] = val
	site["publisher"] = publisher
	o.ortb["site"] = site
	return
}

/********************** Site.Content **********************/

// ORTBSiteContentID will read and set ortb Site.Content.ID parameter
func (o *OpenRTB) ORTBSiteContentID() (err error) {
	val := o.values.Get(ORTBSiteContentID)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["id"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentEpisode will read and set ortb Site.Content.Episode parameter
func (o *OpenRTB) ORTBSiteContentEpisode() (err error) {
	val := o.values.Get(ORTBSiteContentEpisode)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["episode"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentTitle will read and set ortb Site.Content.Title parameter
func (o *OpenRTB) ORTBSiteContentTitle() (err error) {
	val := o.values.Get(ORTBSiteContentTitle)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["title"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentSeries will read and set ortb Site.Content.Series parameter
func (o *OpenRTB) ORTBSiteContentSeries() (err error) {
	val := o.values.Get(ORTBSiteContentSeries)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["series"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentSeason will read and set ortb Site.Content.Season parameter
func (o *OpenRTB) ORTBSiteContentSeason() (err error) {
	val := o.values.Get(ORTBSiteContentSeason)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["season"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentArtist will read and set ortb Site.Content.Artist parameter
func (o *OpenRTB) ORTBSiteContentArtist() (err error) {
	val := o.values.Get(ORTBSiteContentArtist)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["artist"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentGenre will read and set ortb Site.Content.Genre parameter
func (o *OpenRTB) ORTBSiteContentGenre() (err error) {
	val := o.values.Get(ORTBSiteContentGenre)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["genre"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentAlbum will read and set ortb Site.Content.Album parameter
func (o *OpenRTB) ORTBSiteContentAlbum() (err error) {
	val := o.values.Get(ORTBSiteContentAlbum)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["album"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentIsRc will read and set ortb Site.Content.IsRc parameter
func (o *OpenRTB) ORTBSiteContentIsRc() (err error) {
	val := o.values.Get(ORTBSiteContentIsRc)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["isrc"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentURL will read and set ortb Site.Content.URL parameter
func (o *OpenRTB) ORTBSiteContentURL() (err error) {
	val := o.values.Get(ORTBSiteContentURL)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["url"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentCat will read and set ortb Site.Content.Cat parameter
func (o *OpenRTB) ORTBSiteContentCat() (err error) {
	val := o.values.GetStringArray(ORTBSiteContentCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["cat"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentProdQ will read and set ortb Site.Content.ProdQ parameter
func (o *OpenRTB) ORTBSiteContentProdQ() (err error) {
	val := o.values.Get(ORTBSiteContentProdQ)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["prodq"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentVideoQuality will read and set ortb Site.Content.VideoQuality parameter
func (o *OpenRTB) ORTBSiteContentVideoQuality() (err error) {
	val := o.values.Get(ORTBSiteContentVideoQuality)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["videoquality"] = val
	site["content"] = content
	o.ortb["site"] = site
	return

}

// ORTBSiteContentContext will read and set ortb Site.Content.Context parameter
func (o *OpenRTB) ORTBSiteContentContext() (err error) {
	val := o.values.Get(ORTBSiteContentContext)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["context"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentContentRating will read and set ortb Site.Content.ContentRating parameter
func (o *OpenRTB) ORTBSiteContentContentRating() (err error) {
	val := o.values.Get(ORTBSiteContentContentRating)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["contentrating"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentUserRating will read and set ortb Site.Content.UserRating parameter
func (o *OpenRTB) ORTBSiteContentUserRating() (err error) {
	val := o.values.Get(ORTBSiteContentUserRating)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["userrating"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentQaGmeDiarating will read and set ortb Site.Content.QaGmeDiarating parameter
func (o *OpenRTB) ORTBSiteContentQaGmeDiarating() (err error) {
	val := o.values.Get(ORTBSiteContentQaGmeDiarating)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["qagmediarating"] = val
	site["content"] = content
	o.ortb["site"] = site
	return

}

// ORTBSiteContentKeywords will read and set ortb Site.Content.Keywords parameter
func (o *OpenRTB) ORTBSiteContentKeywords() (err error) {
	val := o.values.Get(ORTBSiteContentKeywords)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["keywords"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentLiveStream will read and set ortb Site.Content.LiveStream parameter
func (o *OpenRTB) ORTBSiteContentLiveStream() (err error) {
	val := o.values.Get(ORTBSiteContentLiveStream)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["livestream"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentSourceRelationship will read and set ortb Site.Content.SourceRelationship parameter
func (o *OpenRTB) ORTBSiteContentSourceRelationship() (err error) {
	val := o.values.Get(ORTBSiteContentSourceRelationship)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["sourcerelationship"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentLen will read and set ortb Site.Content.Len parameter
func (o *OpenRTB) ORTBSiteContentLen() (err error) {
	val := o.values.Get(ORTBSiteContentLen)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["len"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentLanguage will read and set ortb Site.Content.Language parameter
func (o *OpenRTB) ORTBSiteContentLanguage() (err error) {
	val := o.values.Get(ORTBSiteContentLanguage)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["language"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentEmbeddable will read and set ortb Site.Content.Embeddable parameter
func (o *OpenRTB) ORTBSiteContentEmbeddable() (err error) {
	val := o.values.Get(ORTBSiteContentEmbeddable)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["embeddable"] = val
	site["content"] = content
	o.ortb["site"] = site
	return
}

/********************** Site.Content.Network **********************/

// ORTBSiteContentNetworkID will read and set ortb Site.Content.Network.Id parameter
func (o *OpenRTB) ORTBSiteContentNetworkID() (err error) {
	val := o.values.Get(ORTBSiteContentNetworkID)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	network, ok := content["network"].(map[string]interface{})
	if !ok {
		network = map[string]interface{}{}
	}
	network["id"] = val
	content["network"] = network
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentNetworkName will read and set ortb Site.Content.Network.Name parameter
func (o *OpenRTB) ORTBSiteContentNetworkName() (err error) {
	val := o.values.Get(ORTBSiteContentNetworkName)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	network, ok := content["network"].(map[string]interface{})
	if !ok {
		network = map[string]interface{}{}
	}
	network["name"] = val
	content["network"] = network
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentNetworkDomain will read and set ortb Site.Content.Network.Domain parameter
func (o *OpenRTB) ORTBSiteContentNetworkDomain() (err error) {
	val := o.values.Get(ORTBSiteContentNetworkDomain)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	network, ok := content["network"].(map[string]interface{})
	if !ok {
		network = map[string]interface{}{}
	}
	network["domain"] = val
	content["network"] = network
	site["content"] = content
	o.ortb["site"] = site
	return
}

/********************** Site.Content.Channel **********************/

// ORTBSiteContentChannelID will read and set ortb Site.Content.Channel.Id parameter
func (o *OpenRTB) ORTBSiteContentChannelID() (err error) {
	val := o.values.Get(ORTBSiteContentChannelID)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	channel, ok := content["channel"].(map[string]interface{})
	if !ok {
		channel = map[string]interface{}{}
	}
	channel["id"] = val
	content["channel"] = channel
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentChannelName will read and set ortb Site.Content.Channel.Name parameter
func (o *OpenRTB) ORTBSiteContentChannelName() (err error) {
	val := o.values.Get(ORTBSiteContentChannelName)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	channel, ok := content["channel"].(map[string]interface{})
	if !ok {
		channel = map[string]interface{}{}
	}
	channel["name"] = val
	content["channel"] = channel
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentChannelDomain will read and set ortb Site.Content.Channel.Domain parameter
func (o *OpenRTB) ORTBSiteContentChannelDomain() (err error) {
	val := o.values.Get(ORTBSiteContentChannelDomain)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	channel, ok := content["channel"].(map[string]interface{})
	if !ok {
		channel = map[string]interface{}{}
	}
	channel["domain"] = val
	content["channel"] = channel
	site["content"] = content
	o.ortb["site"] = site
	return
}

/********************** Site.Content.Producer **********************/

// ORTBSiteContentProducerID will read and set ortb Site.Content.Producer.ID parameter
func (o *OpenRTB) ORTBSiteContentProducerID() (err error) {
	val := o.values.Get(ORTBSiteContentProducerID)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	producer, ok := content["producer"].(map[string]interface{})
	if !ok {
		producer = map[string]interface{}{}
	}
	producer["id"] = val
	content["producer"] = producer
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentProducerName will read and set ortb Site.Content.Producer.Name parameter
func (o *OpenRTB) ORTBSiteContentProducerName() (err error) {
	val := o.values.Get(ORTBSiteContentProducerName)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	producer, ok := content["producer"].(map[string]interface{})
	if !ok {
		producer = map[string]interface{}{}
	}
	producer["name"] = val
	content["producer"] = producer
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentProducerCat will read and set ortb Site.Content.Producer.Cat parameter
func (o *OpenRTB) ORTBSiteContentProducerCat() (err error) {
	val := o.values.GetStringArray(ORTBSiteContentProducerCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	producer, ok := content["producer"].(map[string]interface{})
	if !ok {
		producer = map[string]interface{}{}
	}
	producer["cat"] = val
	content["producer"] = producer
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentProducerDomain will read and set ortb Site.Content.Producer.Domain parameter
func (o *OpenRTB) ORTBSiteContentProducerDomain() (err error) {
	val := o.values.Get(ORTBSiteContentProducerDomain)
	if len(val) == 0 {
		return
	}
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	producer, ok := content["producer"].(map[string]interface{})
	if !ok {
		producer = map[string]interface{}{}
	}
	producer["domain"] = val
	content["producer"] = producer
	site["content"] = content
	o.ortb["site"] = site
	return
}

/*********************** App ***********************/

// ORTBAppID will read and set ortb App.ID parameter
func (o *OpenRTB) ORTBAppID() (err error) {
	val := o.values.Get(ORTBAppID)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["id"] = val
	o.ortb["app"] = app
	return
}

// ORTBAppName will read and set ortb App.Name parameter
func (o *OpenRTB) ORTBAppName() (err error) {
	val := o.values.Get(ORTBAppName)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["name"] = val
	o.ortb["app"] = app
	return
}

// ORTBAppBundle will read and set ortb App.Bundle parameter
func (o *OpenRTB) ORTBAppBundle() (err error) {
	val := o.values.Get(ORTBAppBundle)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["bundle"] = val
	o.ortb["app"] = app
	return
}

// ORTBAppDomain will read and set ortb App.Domain parameter
func (o *OpenRTB) ORTBAppDomain() (err error) {
	val := o.values.Get(ORTBAppDomain)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["domain"] = val
	o.ortb["app"] = app
	return
}

// ORTBAppStoreURL will read and set ortb App.StoreURL parameter
func (o *OpenRTB) ORTBAppStoreURL() (err error) {
	val := o.values.Get(ORTBAppStoreURL)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["storeurl"] = val
	o.ortb["app"] = app
	return
}

// ORTBAppVer will read and set ortb App.Ver parameter
func (o *OpenRTB) ORTBAppVer() (err error) {
	val := o.values.Get(ORTBAppVer)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["ver"] = val
	o.ortb["app"] = app
	return
}

// ORTBAppPaid will read and set ortb App.Paid parameter
func (o *OpenRTB) ORTBAppPaid() (err error) {
	val := o.values.Get(ORTBAppPaid)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["paid"] = val
	o.ortb["app"] = app
	return
}

// ORTBAppCat will read and set ortb App.Cat parameter
func (o *OpenRTB) ORTBAppCat() (err error) {
	val := o.values.GetStringArray(ORTBAppCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["cat"] = val
	o.ortb["app"] = app
	return
}

// ORTBAppSectionCat will read and set ortb App.SectionCat parameter
func (o *OpenRTB) ORTBAppSectionCat() (err error) {
	val := o.values.GetStringArray(ORTBAppSectionCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["sectioncat"] = val
	o.ortb["app"] = app
	return
}

// ORTBAppPageCat will read and set ortb App.PageCat parameter
func (o *OpenRTB) ORTBAppPageCat() (err error) {
	val := o.values.GetStringArray(ORTBAppPageCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["pagecat"] = val
	o.ortb["app"] = app
	return
}

// ORTBAppPrivacyPolicy will read and set ortb App.PrivacyPolicy parameter
func (o *OpenRTB) ORTBAppPrivacyPolicy() (err error) {
	val := o.values.Get(ORTBAppPrivacyPolicy)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["privacypolicy"] = val
	o.ortb["app"] = app
	return
}

// ORTBAppKeywords will read and set ortb App.Keywords parameter
func (o *OpenRTB) ORTBAppKeywords() (err error) {
	val := o.values.Get(ORTBAppKeywords)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	app["keywords"] = val
	o.ortb["app"] = app
	return
}

/*********************** App.Publisher ***********************/

// ORTBAppPublisherID will read and set ortb App.Publisher.ID parameter
func (o *OpenRTB) ORTBAppPublisherID() (err error) {
	val := o.values.Get(ORTBAppPublisherID)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	publisher, ok := app["publisher"].(map[string]interface{})
	if !ok {
		publisher = map[string]interface{}{}
	}
	publisher["id"] = val
	app["publisher"] = publisher
	o.ortb["app"] = app
	return
}

// ORTBAppPublisherName will read and set ortb App.Publisher.Name parameter
func (o *OpenRTB) ORTBAppPublisherName() (err error) {
	val := o.values.Get(ORTBAppPublisherName)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	publisher, ok := app["publisher"].(map[string]interface{})
	if !ok {
		publisher = map[string]interface{}{}
	}
	publisher["name"] = val
	app["publisher"] = publisher
	o.ortb["app"] = app
	return
}

// ORTBAppPublisherCat will read and set ortb App.Publisher.Cat parameter
func (o *OpenRTB) ORTBAppPublisherCat() (err error) {
	val := o.values.GetStringArray(ORTBAppPublisherCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	publisher, ok := app["publisher"].(map[string]interface{})
	if !ok {
		publisher = map[string]interface{}{}
	}
	publisher["cat"] = val
	app["publisher"] = publisher
	o.ortb["app"] = app
	return
}

// ORTBAppPublisherDomain will read and set ortb App.Publisher.Domain parameter
func (o *OpenRTB) ORTBAppPublisherDomain() (err error) {
	val := o.values.Get(ORTBAppPublisherDomain)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	publisher, ok := app["publisher"].(map[string]interface{})
	if !ok {
		publisher = map[string]interface{}{}
	}
	publisher["domain"] = val
	app["publisher"] = publisher
	o.ortb["app"] = app
	return
}

/********************** App.Content **********************/

// ORTBAppContentID will read and set ortb App.Content.ID parameter
func (o *OpenRTB) ORTBAppContentID() (err error) {
	val := o.values.Get(ORTBAppContentID)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["id"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentEpisode will read and set ortb App.Content.Episode parameter
func (o *OpenRTB) ORTBAppContentEpisode() (err error) {
	val := o.values.Get(ORTBAppContentEpisode)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["episode"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentTitle will read and set ortb App.Content.Title parameter
func (o *OpenRTB) ORTBAppContentTitle() (err error) {
	val := o.values.Get(ORTBAppContentTitle)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["title"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentSeries will read and set ortb App.Content.Series parameter
func (o *OpenRTB) ORTBAppContentSeries() (err error) {
	val := o.values.Get(ORTBAppContentSeries)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["series"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentSeason will read and set ortb App.Content.Season parameter
func (o *OpenRTB) ORTBAppContentSeason() (err error) {
	val := o.values.Get(ORTBAppContentSeason)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["season"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentArtist will read and set ortb App.Content.Artist parameter
func (o *OpenRTB) ORTBAppContentArtist() (err error) {
	val := o.values.Get(ORTBAppContentArtist)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["artist"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentGenre will read and set ortb App.Content.Genre parameter
func (o *OpenRTB) ORTBAppContentGenre() (err error) {
	val := o.values.Get(ORTBAppContentGenre)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["genre"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentAlbum will read and set ortb App.Content.Album parameter
func (o *OpenRTB) ORTBAppContentAlbum() (err error) {
	val := o.values.Get(ORTBAppContentAlbum)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["album"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentIsRc will read and set ortb App.Content.IsRc parameter
func (o *OpenRTB) ORTBAppContentIsRc() (err error) {
	val := o.values.Get(ORTBAppContentIsRc)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["isrc"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentURL will read and set ortb App.Content.URL parameter
func (o *OpenRTB) ORTBAppContentURL() (err error) {
	val := o.values.Get(ORTBAppContentURL)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["url"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentCat will read and set ortb App.Content.Cat parameter
func (o *OpenRTB) ORTBAppContentCat() (err error) {
	val := o.values.GetStringArray(ORTBAppContentCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["cat"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentProdQ will read and set ortb App.Content.ProdQ parameter
func (o *OpenRTB) ORTBAppContentProdQ() (err error) {
	val := o.values.Get(ORTBAppContentProdQ)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["prodq"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentVideoQuality will read and set ortb App.Content.VideoQuality parameter
func (o *OpenRTB) ORTBAppContentVideoQuality() (err error) {
	val := o.values.Get(ORTBAppContentVideoQuality)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["videoquality"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentContext will read and set ortb App.Content.Context parameter
func (o *OpenRTB) ORTBAppContentContext() (err error) {
	val := o.values.Get(ORTBAppContentContext)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["context"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentContentRating will read and set ortb App.Content.ContentRating parameter
func (o *OpenRTB) ORTBAppContentContentRating() (err error) {
	val := o.values.Get(ORTBAppContentContentRating)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["contentrating"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentUserRating will read and set ortb App.Content.UserRating parameter
func (o *OpenRTB) ORTBAppContentUserRating() (err error) {
	val := o.values.Get(ORTBAppContentUserRating)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["userrating"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentQaGmeDiarating will read and set ortb App.Content.QaGmeDiarating parameter
func (o *OpenRTB) ORTBAppContentQaGmeDiarating() (err error) {
	val := o.values.Get(ORTBAppContentQaGmeDiarating)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["qagmediarating"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentKeywords will read and set ortb App.Content.Keywords parameter
func (o *OpenRTB) ORTBAppContentKeywords() (err error) {
	val := o.values.Get(ORTBAppContentKeywords)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["keywords"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentLiveStream will read and set ortb App.Content.LiveStream parameter
func (o *OpenRTB) ORTBAppContentLiveStream() (err error) {
	val := o.values.Get(ORTBAppContentLiveStream)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["livestream"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentSourceRelationship will read and set ortb App.Content.SourceRelationship parameter
func (o *OpenRTB) ORTBAppContentSourceRelationship() (err error) {
	val := o.values.Get(ORTBAppContentSourceRelationship)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["sourcerelationship"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentLen will read and set ortb App.Content.Len parameter
func (o *OpenRTB) ORTBAppContentLen() (err error) {
	val := o.values.Get(ORTBAppContentLen)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["len"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentLanguage will read and set ortb App.Content.Language parameter
func (o *OpenRTB) ORTBAppContentLanguage() (err error) {
	val := o.values.Get(ORTBAppContentLanguage)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["language"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentEmbeddable will read and set ortb App.Content.Embeddable parameter
func (o *OpenRTB) ORTBAppContentEmbeddable() (err error) {
	val := o.values.Get(ORTBAppContentEmbeddable)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	content["embeddable"] = val
	app["content"] = content
	o.ortb["app"] = app
	return
}

/********************** App.Content.Network **********************/

// ORTBAppContentNetworkID will read and set ortb App.Content.Network.Id parameter
func (o *OpenRTB) ORTBAppContentNetworkID() (err error) {
	val := o.values.Get(ORTBAppContentNetworkID)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	network, ok := content["network"].(map[string]interface{})
	if !ok {
		network = map[string]interface{}{}
	}
	network["id"] = val
	content["network"] = network
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentNetworkName will read and set ortb App.Content.Network.Name parameter
func (o *OpenRTB) ORTBAppContentNetworkName() (err error) {
	val := o.values.Get(ORTBAppContentNetworkName)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	network, ok := content["network"].(map[string]interface{})
	if !ok {
		network = map[string]interface{}{}
	}
	network["name"] = val
	content["network"] = network
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentNetworkDomain will read and set ortb App.Content.Network.Domain parameter
func (o *OpenRTB) ORTBAppContentNetworkDomain() (err error) {
	val := o.values.Get(ORTBAppContentNetworkDomain)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	network, ok := content["network"].(map[string]interface{})
	if !ok {
		network = map[string]interface{}{}
	}
	network["domain"] = val
	content["network"] = network
	app["content"] = content
	o.ortb["app"] = app
	return
}

/********************** App.Content.Channel **********************/

// ORTBAppContentChannelID will read and set ortb App.Content.Channel.Id parameter
func (o *OpenRTB) ORTBAppContentChannelID() (err error) {
	val := o.values.Get(ORTBAppContentChannelID)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	channel, ok := content["channel"].(map[string]interface{})
	if !ok {
		channel = map[string]interface{}{}
	}
	channel["id"] = val
	content["channel"] = channel
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentChannelName will read and set ortb App.Content.Channel.Name parameter
func (o *OpenRTB) ORTBAppContentChannelName() (err error) {
	val := o.values.Get(ORTBAppContentChannelName)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	channel, ok := content["channel"].(map[string]interface{})
	if !ok {
		channel = map[string]interface{}{}
	}
	channel["name"] = val
	content["channel"] = channel
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentChannelDomain will read and set ortb App.Content.Channel.Domain parameter
func (o *OpenRTB) ORTBAppContentChannelDomain() (err error) {
	val := o.values.Get(ORTBAppContentChannelDomain)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	channel, ok := content["channel"].(map[string]interface{})
	if !ok {
		channel = map[string]interface{}{}
	}
	channel["domain"] = val
	content["channel"] = channel
	app["content"] = content
	o.ortb["app"] = app
	return
}

/********************** App.Content.Producer **********************/

// ORTBAppContentProducerID will read and set ortb App.Content.Producer.ID parameter
func (o *OpenRTB) ORTBAppContentProducerID() (err error) {
	val := o.values.Get(ORTBAppContentProducerID)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	producer, ok := content["producer"].(map[string]interface{})
	if !ok {
		producer = map[string]interface{}{}
	}
	producer["id"] = val
	content["producer"] = producer
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentProducerName will read and set ortb App.Content.Producer.Name parameter
func (o *OpenRTB) ORTBAppContentProducerName() (err error) {
	val := o.values.Get(ORTBAppContentProducerName)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	producer, ok := content["producer"].(map[string]interface{})
	if !ok {
		producer = map[string]interface{}{}
	}
	producer["name"] = val
	content["producer"] = producer
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentProducerCat will read and set ortb App.Content.Producer.Cat parameter
func (o *OpenRTB) ORTBAppContentProducerCat() (err error) {
	val := o.values.GetStringArray(ORTBAppContentProducerCat, ArraySeparator)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	producer, ok := content["producer"].(map[string]interface{})
	if !ok {
		producer = map[string]interface{}{}
	}
	producer["cat"] = val
	content["producer"] = producer
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentProducerDomain will read and set ortb App.Content.Producer.Domain parameter
func (o *OpenRTB) ORTBAppContentProducerDomain() (err error) {
	val := o.values.Get(ORTBAppContentProducerDomain)
	if len(val) == 0 {
		return
	}
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	producer, ok := content["producer"].(map[string]interface{})
	if !ok {
		producer = map[string]interface{}{}
	}
	producer["domain"] = val
	content["producer"] = producer
	app["content"] = content
	o.ortb["app"] = app
	return
}

/********************** Video **********************/

// ORTBImpVideoMimes will read and set ortb Imp.Video.Mimes parameter
func (o *OpenRTB) ORTBImpVideoMimes() (err error) {
	val := o.values.GetStringArray(ORTBImpVideoMimes, ArraySeparator)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["mimes"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoMinDuration will read and set ortb Imp.Video.MinDuration parameter
func (o *OpenRTB) ORTBImpVideoMinDuration() (err error) {
	val := o.values.Get(ORTBImpVideoMinDuration)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["minduration"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoMaxDuration will read and set ortb Imp.Video.MaxDuration parameter
func (o *OpenRTB) ORTBImpVideoMaxDuration() (err error) {
	val := o.values.Get(ORTBImpVideoMaxDuration)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["maxduration"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoProtocols will read and set ortb Imp.Video.Protocols parameter
func (o *OpenRTB) ORTBImpVideoProtocols() (err error) {
	protocols := o.values.GetStringArray(ORTBImpVideoProtocols, ArraySeparator)
	if len(protocols) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["protocols"] = protocols
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoPlayerWidth will read and set ortb Imp.Video.PlayerWidth parameter
func (o *OpenRTB) ORTBImpVideoPlayerWidth() (err error) {
	val := o.values.Get(ORTBImpVideoPlayerWidth)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["w"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoPlayerHeight will read and set ortb Imp.Video.PlayerHeight parameter
func (o *OpenRTB) ORTBImpVideoPlayerHeight() (err error) {
	val := o.values.Get(ORTBImpVideoPlayerHeight)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["h"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoStartDelay will read and set ortb Imp.Video.StartDelay parameter
func (o *OpenRTB) ORTBImpVideoStartDelay() (err error) {
	val := o.values.Get(ORTBImpVideoStartDelay)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["startdelay"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoPlacement will read and set ortb Imp.Video.Placement parameter
func (o *OpenRTB) ORTBImpVideoPlacement() (err error) {
	val := o.values.Get(ORTBImpVideoPlacement)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["placement"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

func (o *OpenRTB) ORTBImpVideoPlcmt() (err error) {
	val := o.values.Get(ORTBImpVideoPlcmt)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["plcmt"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoLinearity will read and set ortb Imp.Video.Linearity parameter
func (o *OpenRTB) ORTBImpVideoLinearity() (err error) {
	val := o.values.Get(ORTBImpVideoLinearity)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["linearity"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoSkip will read and set ortb Imp.Video.Skip parameter
func (o *OpenRTB) ORTBImpVideoSkip() (err error) {
	val := o.values.Get(ORTBImpVideoSkip)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["skip"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoSkipMin will read and set ortb Imp.Video.SkipMin parameter
func (o *OpenRTB) ORTBImpVideoSkipMin() (err error) {
	val := o.values.Get(ORTBImpVideoSkipMin)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["skipmin"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoSkipAfter will read and set ortb Imp.Video.SkipAfter parameter
func (o *OpenRTB) ORTBImpVideoSkipAfter() (err error) {
	val := o.values.Get(ORTBImpVideoSkipAfter)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["skipafter"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoSequence will read and set ortb Imp.Video.Sequence parameter
func (o *OpenRTB) ORTBImpVideoSequence() (err error) {
	val := o.values.Get(ORTBImpVideoSequence)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["sequence"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoBAttr will read and set ortb Imp.Video.BAttr parameter
func (o *OpenRTB) ORTBImpVideoBAttr() (err error) {
	bAttr, err := o.values.GetIntArray(ORTBImpVideoBAttr, ArraySeparator)
	if err != nil {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["battr"] = bAttr
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoMaxExtended will read and set ortb Imp.Video.MaxExtended parameter
func (o *OpenRTB) ORTBImpVideoMaxExtended() (err error) {
	val := o.values.Get(ORTBImpVideoMaxExtended)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["maxextended"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoMinBitrate will read and set ortb Imp.Video.MinBitrate parameter
func (o *OpenRTB) ORTBImpVideoMinBitrate() (err error) {
	val := o.values.Get(ORTBImpVideoMinBitrate)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["minbitrate"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoMaxBitrate will read and set ortb Imp.Video.MaxBitrate parameter
func (o *OpenRTB) ORTBImpVideoMaxBitrate() (err error) {
	val := o.values.Get(ORTBImpVideoMaxBitrate)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["maxbitrate"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoBoxingAllowed will read and set ortb Imp.Video.BoxingAllowed parameter
func (o *OpenRTB) ORTBImpVideoBoxingAllowed() (err error) {
	val := o.values.Get(ORTBImpVideoBoxingAllowed)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["boxingallowed"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoPlaybackMethod will read and set ortb Imp.Video.PlaybackMethod parameter
func (o *OpenRTB) ORTBImpVideoPlaybackMethod() (err error) {
	playbackMethod, err := o.values.GetIntArray(ORTBImpVideoPlaybackMethod, ArraySeparator)
	if err != nil {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["playbackmethod"] = v26.GetPlaybackMethod(playbackMethod)
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoDelivery will read and set ortb Imp.Video.Delivery parameter
func (o *OpenRTB) ORTBImpVideoDelivery() (err error) {
	delivery, err := o.values.GetIntArray(ORTBImpVideoDelivery, ArraySeparator)
	if err != nil {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["delivery"] = delivery
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoPos will read and set ortb Imp.Video.Pos parameter
func (o *OpenRTB) ORTBImpVideoPos() (err error) {
	val := o.values.Get(ORTBImpVideoPos)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["pos"] = val
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoAPI will read and set ortb Imp.Video.API parameter
func (o *OpenRTB) ORTBImpVideoAPI() (err error) {
	api, err := o.values.GetIntArray(ORTBImpVideoAPI, ArraySeparator)
	if err != nil {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["api"] = api
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoCompanionType will read and set ortb Imp.Video.CompanionType parameter
func (o *OpenRTB) ORTBImpVideoCompanionType() (err error) {
	companionType, err := o.values.GetIntArray(ORTBImpVideoCompanionType, ArraySeparator)
	if err != nil {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	video["companiontype"] = companionType
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

/*********************** Regs ***********************/

// ORTBRegsCoppa will read and set ortb Regs.Coppa parameter
func (o *OpenRTB) ORTBRegsCoppa() (err error) {
	val := o.values.Get(ORTBRegsCoppa)
	if len(val) == 0 {
		return
	}
	regs, ok := o.ortb["regs"].(map[string]interface{})
	if !ok {
		regs = map[string]interface{}{}
	}
	regs["coppa"] = val
	o.ortb["regs"] = regs
	return
}

/*********************** Imp ***********************/

// ORTBImpID will read and set ortb Imp.ID parameter
func (o *OpenRTB) ORTBImpID() (err error) {
	val := o.values.Get(ORTBImpID)
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}

	if len(val) == 0 {
		imp[0]["id"] = uuid.NewV4().String()
	} else {
		imp[0]["id"] = val
	}
	o.ortb["imp"] = imp
	return
}

// ORTBImpDisplayManager will read and set ortb Imp.DisplayManager parameter
func (o *OpenRTB) ORTBImpDisplayManager() (err error) {
	val := o.values.Get(ORTBImpDisplayManager)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	imp[0]["displaymanager"] = val
	o.ortb["imp"] = imp
	return
}

// ORTBImpDisplayManagerVer will read and set ortb Imp.DisplayManagerVer parameter
func (o *OpenRTB) ORTBImpDisplayManagerVer() (err error) {
	val := o.values.Get(ORTBImpDisplayManagerVer)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	imp[0]["displaymanagerver"] = val
	o.ortb["imp"] = imp
	return
}

// ORTBImpInstl will read and set ortb Imp.Instl parameter
func (o *OpenRTB) ORTBImpInstl() (err error) {
	val := o.values.Get(ORTBImpInstl)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	imp[0]["instl"] = val
	o.ortb["imp"] = imp
	return
}

// ORTBImpTagID will read and set ortb Imp.TagId parameter
func (o *OpenRTB) ORTBImpTagID() (err error) {
	val := o.values.Get(ORTBImpTagID)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	imp[0]["tagid"] = val
	o.ortb["imp"] = imp
	return
}

// ORTBImpBidFloor will read and set ortb Imp.BidFloor parameter
func (o *OpenRTB) ORTBImpBidFloor() (err error) {
	bidFloor := o.values.Get(ORTBImpBidFloor)
	if len(bidFloor) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	imp[0]["bidfloor"] = bidFloor
	o.ortb["imp"] = imp
	return
}

// ORTBImpBidFloorCur will read and set ortb Imp.BidFloorCur parameter
func (o *OpenRTB) ORTBImpBidFloorCur() (err error) {
	bidFloor := o.values.Get(ORTBImpBidFloor)
	if len(bidFloor) == 0 {
		return
	}
	bidFloorCur := o.values.Get(ORTBImpBidFloorCur)
	if len(bidFloorCur) == 0 {
		bidFloorCur = USD
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	imp[0]["bidfloorcur"] = bidFloorCur
	o.ortb["imp"] = imp
	return
}

// ORTBImpClickBrowser will read and set ortb Imp.ClickBrowser parameter
func (o *OpenRTB) ORTBImpClickBrowser() (err error) {
	val := o.values.Get(ORTBImpClickBrowser)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	imp[0]["clickbrowser"] = val
	o.ortb["imp"] = imp
	return
}

// ORTBImpSecure will read and set ortb Imp.Secure parameter
func (o *OpenRTB) ORTBImpSecure() (err error) {
	val := o.values.Get(ORTBImpSecure)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	imp[0]["secure"] = val
	o.ortb["imp"] = imp
	return
}

// ORTBImpIframeBuster will read and set ortb Imp.IframeBuster parameter
func (o *OpenRTB) ORTBImpIframeBuster() (err error) {
	val := o.values.GetStringArray(ORTBImpIframeBuster, ArraySeparator)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	imp[0]["iframebuster"] = val
	o.ortb["imp"] = imp
	return
}

// ORTBImpExp will read and set ortb Imp.Exp parameter
func (o *OpenRTB) ORTBImpExp() (err error) {
	val := o.values.Get(ORTBImpExp)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	imp[0]["exp"] = val
	o.ortb["imp"] = imp
	return
}

// ORTBImpPmp will read and set ortb Imp.Pmp parameter
func (o *OpenRTB) ORTBImpPmp() (err error) {
	pmp := o.values.Get(ORTBImpPmp)
	if len(pmp) == 0 {
		return
	}
	ortbPmp := map[string]interface{}{}
	err = json.Unmarshal([]byte(pmp), &ortbPmp)
	if err != nil {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	imp[0]["pmp"] = ortbPmp
	o.ortb["imp"] = imp
	return
}

// ORTBImpExtBidder will read and set ortb Imp.Ext.Bidder parameter
func (o *OpenRTB) ORTBImpExtBidder() (err error) {
	val := o.values.Get(ORTBImpExtBidder)
	if len(val) == 0 {
		return
	}

	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}

	impExt, ok := imp[0]["ext"].(map[string]interface{})
	if !ok {
		impExt = map[string]interface{}{}
	}

	impExtBidder := map[string]interface{}{}
	err = json.Unmarshal([]byte(val), &impExtBidder)
	if err != nil {
		return
	}

	impExt[BIDDER_KEY] = impExtBidder
	imp[0]["ext"] = impExt
	o.ortb["imp"] = imp
	return
}

// ORTBImpExtPrebid will read and set ortb Imp.Ext.Prebid parameter
func (o *OpenRTB) ORTBImpExtPrebid() (err error) {
	str, ok := o.values.GetString(ORTBImpExtPrebid)
	if !ok {
		return
	}

	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}

	impExt, ok := imp[0]["ext"].(map[string]interface{})
	if !ok {
		impExt = map[string]interface{}{}
	}

	impExtPrebid := map[string]interface{}{}
	err = json.Unmarshal([]byte(str), &impExtPrebid)
	if err != nil {
		return
	}

	impExt[PrebidKey] = impExtPrebid
	imp[0]["ext"] = impExt
	o.ortb["imp"] = imp
	return
}

/********************** Device **********************/

// ORTBDeviceUserAgent will read and set ortb Device.UserAgent parameter
func (o *OpenRTB) ORTBDeviceUserAgent() (err error) {
	val := o.values.Get(ORTBDeviceUserAgent)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["ua"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceIP will read and set ortb Device.IP parameter
func (o *OpenRTB) ORTBDeviceIP() (err error) {
	val := o.values.Get(ORTBDeviceIP)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["ip"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceIpv6 will read and set ortb Device.Ipv6 parameter
func (o *OpenRTB) ORTBDeviceIpv6() (err error) {
	val := o.values.Get(ORTBDeviceIpv6)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["ipv6"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceDnt will read and set ortb Device.Dnt parameter
func (o *OpenRTB) ORTBDeviceDnt() (err error) {
	val := o.values.Get(ORTBDeviceDnt)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["dnt"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceLmt will read and set ortb Device.Lmt parameter
func (o *OpenRTB) ORTBDeviceLmt() (err error) {
	val := o.values.Get(ORTBDeviceLmt)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["lmt"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceDeviceType will read and set ortb Device.DeviceType parameter
func (o *OpenRTB) ORTBDeviceDeviceType() (err error) {
	val := o.values.Get(ORTBDeviceDeviceType)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["devicetype"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceMake will read and set ortb Device.Make parameter
func (o *OpenRTB) ORTBDeviceMake() (err error) {
	val := o.values.Get(ORTBDeviceMake)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["make"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceModel will read and set ortb Device.Model parameter
func (o *OpenRTB) ORTBDeviceModel() (err error) {
	val := o.values.Get(ORTBDeviceModel)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["model"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceOs will read and set ortb Device.Os parameter
func (o *OpenRTB) ORTBDeviceOs() (err error) {
	val := o.values.Get(ORTBDeviceOs)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["os"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceOsv will read and set ortb Device.Osv parameter
func (o *OpenRTB) ORTBDeviceOsv() (err error) {
	val := o.values.Get(ORTBDeviceOsv)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["osv"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceHwv will read and set ortb Device.Hwv parameter
func (o *OpenRTB) ORTBDeviceHwv() (err error) {
	val := o.values.Get(ORTBDeviceHwv)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["hwv"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceWidth will read and set ortb Device.Width parameter
func (o *OpenRTB) ORTBDeviceWidth() (err error) {
	val := o.values.Get(ORTBDeviceWidth)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["w"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceHeight will read and set ortb Device.Height parameter
func (o *OpenRTB) ORTBDeviceHeight() (err error) {
	val := o.values.Get(ORTBDeviceHeight)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["h"] = val
	o.ortb["device"] = device
	return
}

// ORTBDevicePpi will read and set ortb Device.Ppi parameter
func (o *OpenRTB) ORTBDevicePpi() (err error) {
	val := o.values.Get(ORTBDevicePpi)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["ppi"] = val
	o.ortb["device"] = device
	return
}

// ORTBDevicePxRatio will read and set ortb Device.PxRatio parameter
func (o *OpenRTB) ORTBDevicePxRatio() (err error) {
	val := o.values.Get(ORTBDevicePxRatio)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["pxratio"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceJS will read and set ortb Device.JS parameter
func (o *OpenRTB) ORTBDeviceJS() (err error) {
	val := o.values.Get(ORTBDeviceJS)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["js"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoFetch will read and set ortb Device.Geo.Fetch parameter
func (o *OpenRTB) ORTBDeviceGeoFetch() (err error) {
	val := o.values.Get(ORTBDeviceGeoFetch)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["geofetch"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceFlashVer will read and set ortb Device.FlashVer parameter
func (o *OpenRTB) ORTBDeviceFlashVer() (err error) {
	val := o.values.Get(ORTBDeviceFlashVer)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["flashver"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceLanguage will read and set ortb Device.Language parameter
func (o *OpenRTB) ORTBDeviceLanguage() (err error) {
	val := o.values.Get(ORTBDeviceLanguage)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["language"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceCarrier will read and set ortb Device.Carrier parameter
func (o *OpenRTB) ORTBDeviceCarrier() (err error) {
	val := o.values.Get(ORTBDeviceCarrier)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["carrier"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceMccmnc will read and set ortb Device.Mccmnc parameter
func (o *OpenRTB) ORTBDeviceMccmnc() (err error) {
	val := o.values.Get(ORTBDeviceMccmnc)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["mccmnc"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceConnectionType will read and set ortb Device.ConnectionType parameter
func (o *OpenRTB) ORTBDeviceConnectionType() (err error) {
	val := o.values.Get(ORTBDeviceConnectionType)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["connectiontype"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceIfa will read and set ortb Device.Ifa parameter
func (o *OpenRTB) ORTBDeviceIfa() (err error) {
	val := o.values.Get(ORTBDeviceIfa)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["ifa"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceDidSha1 will read and set ortb Device.DidSha1 parameter
func (o *OpenRTB) ORTBDeviceDidSha1() (err error) {
	val := o.values.Get(ORTBDeviceDidSha1)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["didsha1"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceDidMd5 will read and set ortb Device.DidMd5 parameter
func (o *OpenRTB) ORTBDeviceDidMd5() (err error) {
	val := o.values.Get(ORTBDeviceDidMd5)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["didmd5"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceDpidSha1 will read and set ortb Device.DpidSha1 parameter
func (o *OpenRTB) ORTBDeviceDpidSha1() (err error) {
	val := o.values.Get(ORTBDeviceDpidSha1)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["dpidsha1"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceDpidMd5 will read and set ortb Device.DpidMd5 parameter
func (o *OpenRTB) ORTBDeviceDpidMd5() (err error) {
	val := o.values.Get(ORTBDeviceDpidMd5)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["dpidmd5"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceMacSha1 will read and set ortb Device.MacSha1 parameter
func (o *OpenRTB) ORTBDeviceMacSha1() (err error) {
	val := o.values.Get(ORTBDeviceMacSha1)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["macsha1"] = val
	o.ortb["device"] = device
	return
}

// ORTBDeviceMacMd5 will read and set ortb Device.MacMd5 parameter
func (o *OpenRTB) ORTBDeviceMacMd5() (err error) {
	val := o.values.Get(ORTBDeviceMacMd5)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	device["macmd5"] = val
	o.ortb["device"] = device
	return
}

/*********************** Device.Geo ***********************/

// ORTBDeviceGeoLat will read and set ortb Device.Geo.Lat parameter
func (o *OpenRTB) ORTBDeviceGeoLat() (err error) {
	val := o.values.Get(ORTBDeviceGeoLat)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["lat"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoLon will read and set ortb Device.Geo.Lon parameter
func (o *OpenRTB) ORTBDeviceGeoLon() (err error) {
	val := o.values.Get(ORTBDeviceGeoLon)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["lon"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoType will read and set ortb Device.Geo.Type parameter
func (o *OpenRTB) ORTBDeviceGeoType() (err error) {
	val := o.values.Get(ORTBDeviceGeoType)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["type"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoAccuracy will read and set ortb Device.Geo.Accuracy parameter
func (o *OpenRTB) ORTBDeviceGeoAccuracy() (err error) {
	val := o.values.Get(ORTBDeviceGeoAccuracy)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["accuracy"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoLastFix will read and set ortb Device.Geo.LastFix parameter
func (o *OpenRTB) ORTBDeviceGeoLastFix() (err error) {
	val := o.values.Get(ORTBDeviceGeoLastFix)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["lastfix"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoIPService will read and set ortb Device.Geo.IPService parameter
func (o *OpenRTB) ORTBDeviceGeoIPService() (err error) {
	val := o.values.Get(ORTBDeviceGeoIPService)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["ipservice"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoCountry will read and set ortb Device.Geo.Country parameter
func (o *OpenRTB) ORTBDeviceGeoCountry() (err error) {
	val := o.values.Get(ORTBDeviceGeoCountry)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["country"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoRegion will read and set ortb Device.Geo.Region parameter
func (o *OpenRTB) ORTBDeviceGeoRegion() (err error) {
	val := o.values.Get(ORTBDeviceGeoRegion)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["region"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoRegionFips104 will read and set ortb Device.Geo.RegionFips104 parameter
func (o *OpenRTB) ORTBDeviceGeoRegionFips104() (err error) {
	val := o.values.Get(ORTBDeviceGeoRegionFips104)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["regionfips104"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoMetro will read and set ortb Device.Geo.Metro parameter
func (o *OpenRTB) ORTBDeviceGeoMetro() (err error) {
	val := o.values.Get(ORTBDeviceGeoMetro)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["metro"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoCity will read and set ortb Device.Geo.City parameter
func (o *OpenRTB) ORTBDeviceGeoCity() (err error) {
	val := o.values.Get(ORTBDeviceGeoCity)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["city"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoZip will read and set ortb Device.Geo.Zip parameter
func (o *OpenRTB) ORTBDeviceGeoZip() (err error) {
	val := o.values.Get(ORTBDeviceGeoZip)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["zip"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoUtcOffset will read and set ortb Device.Geo.UtcOffset parameter
func (o *OpenRTB) ORTBDeviceGeoUtcOffset() (err error) {
	val := o.values.Get(ORTBDeviceGeoUtcOffset)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["utcoffset"] = val
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

/*********************** User ***********************/

// ORTBUserID will read and set ortb UserID parameter
func (o *OpenRTB) ORTBUserID() (err error) {
	val := o.values.Get(ORTBUserID)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	user["id"] = val
	o.ortb["user"] = user
	return
}

// ORTBUserBuyerUID will read and set ortb UserBuyerUID parameter
func (o *OpenRTB) ORTBUserBuyerUID() (err error) {
	val := o.values.Get(ORTBUserBuyerUID)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	user["buyeruid"] = val
	o.ortb["user"] = user
	return
}

// ORTBUserYob will read and set ortb UserYob parameter
func (o *OpenRTB) ORTBUserYob() (err error) {
	val := o.values.Get(ORTBUserYob)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	user["yob"] = val
	o.ortb["user"] = user
	return
}

// ORTBUserGender will read and set ortb UserGender parameter
func (o *OpenRTB) ORTBUserGender() (err error) {
	val := o.values.Get(ORTBUserGender)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	user["gender"] = val
	o.ortb["user"] = user
	return
}

// ORTBUserKeywords will read and set ortb UserKeywords parameter
func (o *OpenRTB) ORTBUserKeywords() (err error) {
	val := o.values.Get(ORTBUserKeywords)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	user["keywords"] = val
	o.ortb["user"] = user
	return
}

// ORTBUserCustomData will read and set ortb UserCustomData parameter
func (o *OpenRTB) ORTBUserCustomData() (err error) {
	val := o.values.Get(ORTBUserCustomData)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	user["customdata"] = val
	o.ortb["user"] = user
	return
}

/*********************** User.Geo ***********************/

// ORTBUserGeoLat will read and set ortb UserGeo.Lat parameter
func (o *OpenRTB) ORTBUserGeoLat() (err error) {
	val := o.values.Get(ORTBUserGeoLat)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["lat"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoLon will read and set ortb UserGeo.Lon parameter
func (o *OpenRTB) ORTBUserGeoLon() (err error) {
	val := o.values.Get(ORTBUserGeoLon)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["lon"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoType will read and set ortb UserGeo.Type parameter
func (o *OpenRTB) ORTBUserGeoType() (err error) {
	val := o.values.Get(ORTBUserGeoType)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["type"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoAccuracy will read and set ortb UserGeo.Accuracy parameter
func (o *OpenRTB) ORTBUserGeoAccuracy() (err error) {
	val := o.values.Get(ORTBUserGeoAccuracy)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["accuracy"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoLastFix will read and set ortb UserGeo.LastFix parameter
func (o *OpenRTB) ORTBUserGeoLastFix() (err error) {
	val := o.values.Get(ORTBUserGeoLastFix)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["lastfix"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoIPService will read and set ortb UserGeo.IPService parameter
func (o *OpenRTB) ORTBUserGeoIPService() (err error) {
	val := o.values.Get(ORTBUserGeoIPService)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["ipservice"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoCountry will read and set ortb UserGeo.Country parameter
func (o *OpenRTB) ORTBUserGeoCountry() (err error) {
	val := o.values.Get(ORTBUserGeoCountry)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["country"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoRegion will read and set ortb UserGeo.Region parameter
func (o *OpenRTB) ORTBUserGeoRegion() (err error) {
	val := o.values.Get(ORTBUserGeoRegion)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["region"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoRegionFips104 will read and set ortb UserGeo.RegionFips104 parameter
func (o *OpenRTB) ORTBUserGeoRegionFips104() (err error) {
	val := o.values.Get(ORTBUserGeoRegionFips104)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["regionfips104"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoMetro will read and set ortb UserGeo.Metro parameter
func (o *OpenRTB) ORTBUserGeoMetro() (err error) {
	val := o.values.Get(ORTBUserGeoMetro)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["metro"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoCity will read and set ortb UserGeo.City parameter
func (o *OpenRTB) ORTBUserGeoCity() (err error) {
	val := o.values.Get(ORTBUserGeoCity)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["city"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoZip will read and set ortb UserGeo.Zip parameter
func (o *OpenRTB) ORTBUserGeoZip() (err error) {
	val := o.values.Get(ORTBUserGeoZip)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["zip"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserGeoUtcOffset will read and set ortb UserGeo.UtcOffset parameter
func (o *OpenRTB) ORTBUserGeoUtcOffset() (err error) {
	val := o.values.Get(ORTBUserGeoUtcOffset)
	if len(val) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geo["utcoffset"] = val
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

/*********************** Request.Ext.Parameters ***********************/

// ORTBProfileID will read and set ortb ProfileId parameter
func (o *OpenRTB) ORTBProfileID() (err error) {
	val := o.values.Get(ORTBProfileID)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtProfileId] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	o.ortb["ext"] = reqExt
	return
}

// ORTBVersionID will read and set ortb VersionId parameter
func (o *OpenRTB) ORTBVersionID() (err error) {
	val := o.values.Get(ORTBVersionID)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtVersionId] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	o.ortb["ext"] = reqExt
	return
}

// ORTBSSAuctionFlag will read and set ortb SSAuctionFlag parameter
func (o *OpenRTB) ORTBSSAuctionFlag() (err error) {
	val := o.values.Get(ORTBSSAuctionFlag)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtSSAuctionFlag] = val
	reqExt[ORTBExtWrapper] = wrapperExt
	o.ortb["ext"] = reqExt
	return
}

// ORTBSumryDisableFlag will read and set ortb SumryDisableFlag parameter
func (o *OpenRTB) ORTBSumryDisableFlag() (err error) {
	val := o.values.Get(ORTBSumryDisableFlag)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtSumryDisableFlag] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	o.ortb["ext"] = reqExt
	return
}

// ORTBClientConfigFlag will read and set ortb ClientConfigFlag parameter
func (o *OpenRTB) ORTBClientConfigFlag() (err error) {
	val := o.values.Get(ORTBClientConfigFlag)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtClientConfigFlag] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	o.ortb["ext"] = reqExt
	return
}

// ORTBSupportDeals will read and set ortb ClientConfigFlag parameter
func (o *OpenRTB) ORTBSupportDeals() (err error) {
	val := o.values.Get(ORTBSupportDeals)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtSupportDeals] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	o.ortb["ext"] = reqExt
	return
}

// ORTBIncludeBrandCategory will read and set ortb ORTBIncludeBrandCategory parameter
func (o *OpenRTB) ORTBIncludeBrandCategory() (err error) {
	val := o.values.Get(ORTBIncludeBrandCategory)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtIncludeBrandCategory] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	o.ortb["ext"] = reqExt
	return
}

// ORTBSSAI will read and set ortb ssai parameter
func (o *OpenRTB) ORTBSSAI() (err error) {
	val := o.values.Get(ORTBSSAI)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtSsai] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	o.ortb["ext"] = reqExt
	return
}

// ORTBKeyValues read and set keyval parameter
func (o *OpenRTB) ORTBKeyValues() (err error) {
	val, err := o.values.GetQueryParams(ORTBKeyValues)
	if val == nil {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtKV] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	o.ortb["ext"] = reqExt
	return
}

// ORTBKeyValuesMap read and set keyval parameter
func (o *OpenRTB) ORTBKeyValuesMap() (err error) {
	val, err := o.values.GetJSON(ORTBKeyValuesMap)
	if val == nil {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtKV] = val
	reqExt[ORTBExtWrapper] = wrapperExt
	o.ortb["ext"] = reqExt
	return
}

/*********************** User.Ext.Consent ***********************/

// ORTBUserExtConsent will read and set ortb User.Ext.Consent parameter
func (o *OpenRTB) ORTBUserExtConsent() (err error) {
	val := o.values.Get(ORTBUserExtConsent)
	if len(val) == 0 {
		return
	}

	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}

	userExt, ok := user["ext"].(map[string]interface{})
	if !ok {
		userExt = map[string]interface{}{}
	}
	userExt[ORTBExtConsent] = val
	user["ext"] = userExt
	o.ortb["user"] = user
	return
}

/*********************** Regs.Ext.Gdpr ***********************/

// ORTBRegsExtGdpr will read and set ortb Regs.Ext.Gdpr parameter
func (o *OpenRTB) ORTBRegsExtGdpr() (err error) {
	val := o.values.Get(ORTBRegsExtGdpr)
	if len(val) == 0 {
		return
	}
	regs, ok := o.ortb["regs"].(map[string]interface{})
	if !ok {
		regs = map[string]interface{}{}
	}
	regs["gdpr"] = val

	regsExt, ok := regs["ext"].(map[string]interface{})
	if !ok {
		regsExt = map[string]interface{}{}
	}
	regsExt[ORTBExtGDPR] = val
	regs["ext"] = regsExt
	o.ortb["regs"] = regs
	return
}

// ORTBRegsExtUSPrivacy will read and set ortb Regs.Ext.USPrivacy parameter
func (o *OpenRTB) ORTBRegsExtUSPrivacy() (err error) {
	val := o.values.Get(ORTBRegsExtUSPrivacy)
	if len(val) == 0 {
		return
	}
	regs, ok := o.ortb["regs"].(map[string]interface{})
	if !ok {
		regs = map[string]interface{}{}
	}
	regs["us_privacy"] = val

	regsExt, ok := regs["ext"].(map[string]interface{})
	if !ok {
		regsExt = map[string]interface{}{}
	}
	regsExt[ORTBExtUSPrivacy] = val
	regs["ext"] = regsExt
	o.ortb["regs"] = regs
	return
}

/*********************** Imp.Video.Ext ***********************/

// ORTBImpVideoExtOffset will read and set ortb Imp.Vid.Ext.Offset parameter
func (o *OpenRTB) ORTBImpVideoExtOffset() (err error) {
	val := o.values.Get(ORTBImpVideoExtOffset)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	videoExt, ok := video["ext"].(map[string]interface{})
	if !ok {
		videoExt = map[string]interface{}{}
	}
	videoExt[ORTBExtAdPodOffset] = val
	video["ext"] = videoExt
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoExtAdPodMinAds will read and set ortb Imp.Vid.Ext.AdPod.MinAds parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodMinAds() (err error) {
	val := o.values.Get(ORTBImpVideoExtAdPodMinAds)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	videoExt, ok := video["ext"].(map[string]interface{})
	if !ok {
		videoExt = map[string]interface{}{}
	}
	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMinAds] = val
	videoExt[ORTBExtAdPod] = adpod
	video["ext"] = videoExt
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoExtAdPodMaxAds will read and set ortb Imp.Vid.Ext.AdPod.MaxAds parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodMaxAds() (err error) {
	val := o.values.Get(ORTBImpVideoExtAdPodMaxAds)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	videoExt, ok := video["ext"].(map[string]interface{})
	if !ok {
		videoExt = map[string]interface{}{}
	}
	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMaxAds] = val
	videoExt[ORTBExtAdPod] = adpod
	video["ext"] = videoExt
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoExtAdPodMinDuration will read and set ortb Imp.Vid.Ext.AdPod.MinDuration parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodMinDuration() (err error) {
	val := o.values.Get(ORTBImpVideoExtAdPodMinDuration)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	videoExt, ok := video["ext"].(map[string]interface{})
	if !ok {
		videoExt = map[string]interface{}{}
	}
	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMinDuration] = val
	videoExt[ORTBExtAdPod] = adpod
	video["ext"] = videoExt
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoExtAdPodMaxDuration will read and set ortb Imp.Vid.Ext.AdPod.MaxDuration parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodMaxDuration() (err error) {
	val := o.values.Get(ORTBImpVideoExtAdPodMaxDuration)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	videoExt, ok := video["ext"].(map[string]interface{})
	if !ok {
		videoExt = map[string]interface{}{}
	}
	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMaxDuration] = val
	videoExt[ORTBExtAdPod] = adpod
	video["ext"] = videoExt
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoExtAdPodAdvertiserExclusionPercent will read and set ortb Imp.Vid.Ext.AdPod.AdvertiserExclusionPercent parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodAdvertiserExclusionPercent() (err error) {
	val := o.values.Get(ORTBImpVideoExtAdPodAdvertiserExclusionPercent)
	if len(val) == 0 {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	videoExt, ok := video["ext"].(map[string]interface{})
	if !ok {
		videoExt = map[string]interface{}{}
	}
	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodAdvertiserExclusionPercent] = val
	videoExt[ORTBExtAdPod] = adpod
	video["ext"] = videoExt
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoExtAdPodIABCategoryExclusionPercent will read and set ortb Imp.Vid.Ext.AdPod.IABCategoryExclusionPercent parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodIABCategoryExclusionPercent() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoExtAdPodIABCategoryExclusionPercent)
	if !ok || err != nil {
		return
	}
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	videoExt, ok := video["ext"].(map[string]interface{})
	if !ok {
		videoExt = map[string]interface{}{}
	}
	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodIABCategoryExclusionPercent] = val
	videoExt[ORTBExtAdPod] = adpod
	video["ext"] = videoExt
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

/*********************** Req.Ext ***********************/

// ORTBRequestExtAdPodMinAds will read and set ortb Request.Ext.AdPod.MinAds parameter
func (o *OpenRTB) ORTBRequestExtAdPodMinAds() (err error) {
	val := o.values.Get(ORTBRequestExtAdPodMinAds)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMinAds] = val

	reqExt[ORTBExtAdPod] = adpod
	o.ortb["ext"] = reqExt
	return
}

// ORTBRequestExtAdPodMaxAds will read and set ortb Request.Ext.AdPod.MaxAds parameter
func (o *OpenRTB) ORTBRequestExtAdPodMaxAds() (err error) {
	val := o.values.Get(ORTBRequestExtAdPodMaxAds)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMaxAds] = val

	reqExt[ORTBExtAdPod] = adpod
	o.ortb["ext"] = reqExt
	return
}

// ORTBRequestExtAdPodMinDuration will read and set ortb Request.Ext.AdPod.MinDuration parameter
func (o *OpenRTB) ORTBRequestExtAdPodMinDuration() (err error) {
	val := o.values.Get(ORTBRequestExtAdPodMinDuration)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMinDuration] = val

	reqExt[ORTBExtAdPod] = adpod
	o.ortb["ext"] = reqExt
	return
}

// ORTBRequestExtAdPodMaxDuration will read and set ortb Request.Ext.AdPod.MaxDuration parameter
func (o *OpenRTB) ORTBRequestExtAdPodMaxDuration() (err error) {
	val := o.values.Get(ORTBRequestExtAdPodMaxDuration)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMaxDuration] = val

	reqExt[ORTBExtAdPod] = adpod
	o.ortb["ext"] = reqExt
	return
}

// ORTBRequestExtAdPodAdvertiserExclusionPercent will read and set ortb Request.Ext.AdPod.AdvertiserExclusionPercent parameter
func (o *OpenRTB) ORTBRequestExtAdPodAdvertiserExclusionPercent() (err error) {
	val := o.values.Get(ORTBRequestExtAdPodAdvertiserExclusionPercent)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodAdvertiserExclusionPercent] = val

	reqExt[ORTBExtAdPod] = adpod
	o.ortb["ext"] = reqExt
	return
}

// ORTBRequestExtAdPodIABCategoryExclusionPercent will read and set ortb Request.Ext.AdPod.IABCategoryExclusionPercent parameter
func (o *OpenRTB) ORTBRequestExtAdPodIABCategoryExclusionPercent() (err error) {
	val := o.values.Get(ORTBRequestExtAdPodIABCategoryExclusionPercent)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodIABCategoryExclusionPercent] = val

	reqExt[ORTBExtAdPod] = adpod
	o.ortb["ext"] = reqExt
	return
}

// ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent will read and set ortb Request.Ext.AdPod.CrossPodAdvertiserExclusionPercent parameter
func (o *OpenRTB) ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent() (err error) {
	val := o.values.Get(ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodCrossPodAdvertiserExclusionPercent] = val

	reqExt[ORTBExtAdPod] = adpod
	o.ortb["ext"] = reqExt
	return
}

// ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent will read and set ortb Request.Ext.AdPod.CrossPodIABCategoryExclusionPercent parameter
func (o *OpenRTB) ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent() (err error) {
	val := o.values.Get(ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodCrossPodIABCategoryExclusionPercent] = val

	reqExt[ORTBExtAdPod] = adpod
	o.ortb["ext"] = reqExt
	return
}

// ORTBRequestExtAdPodIABCategoryExclusionWindow will read and set ortb Request.Ext.AdPod.IABCategoryExclusionWindow parameter
func (o *OpenRTB) ORTBRequestExtAdPodIABCategoryExclusionWindow() (err error) {
	val := o.values.Get(ORTBRequestExtAdPodIABCategoryExclusionWindow)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodIABCategoryExclusionWindow] = val

	reqExt[ORTBExtAdPod] = adpod
	o.ortb["ext"] = reqExt
	return
}

// ORTBRequestExtAdPodAdvertiserExclusionWindow will read and set ortb Request.Ext.AdPod.AdvertiserExclusionWindow parameter
func (o *OpenRTB) ORTBRequestExtAdPodAdvertiserExclusionWindow() (err error) {
	val := o.values.Get(ORTBRequestExtAdPodAdvertiserExclusionWindow)
	if len(val) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodAdvertiserExclusionWindow] = val

	reqExt[ORTBExtAdPod] = adpod
	o.ortb["ext"] = reqExt
	return
}

/*********************** Ext ***********************/

// ORTBBidRequestExt will read and set ortb BidRequest.Ext parameter
func (o *OpenRTB) ORTBBidRequestExt(key string, value *string) (err error) {
	ext, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		ext = map[string]interface{}{}
	}
	SetValue(ext, key, value)

	o.ortb["ext"] = ext
	return
}

// ORTBSourceExt will read and set ortb Source.Ext parameter
func (o *OpenRTB) ORTBSourceExt(key string, value *string) (err error) {
	source, ok := o.ortb["source"].(map[string]interface{})
	if !ok {
		source = map[string]interface{}{}
	}

	sourceExt, ok := source["ext"].(map[string]interface{})
	if !ok {
		sourceExt = map[string]interface{}{}
	}

	SetValue(sourceExt, key, value)
	source["ext"] = sourceExt
	o.ortb["source"] = source
	return
}

// ORTBRegsExt will read and set ortb Regs.Ext parameter
func (o *OpenRTB) ORTBRegsExt(key string, value *string) (err error) {
	regs, ok := o.ortb["regs"].(map[string]interface{})
	if !ok {
		regs = map[string]interface{}{}
	}

	regsExt, ok := regs["ext"].(map[string]interface{})
	if !ok {
		regsExt = map[string]interface{}{}
	}
	SetValue(regsExt, key, value)
	regs["ext"] = regsExt
	o.ortb["regs"] = regs
	return
}

// ORTBImpExt will read and set ortb Imp.Ext parameter
func (o *OpenRTB) ORTBImpExt(key string, value *string) (err error) {
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	impExt, ok := imp[0]["ext"].(map[string]interface{})
	if !ok {
		impExt = map[string]interface{}{}
	}
	SetValue(impExt, key, value)

	imp[0]["ext"] = impExt
	o.ortb["imp"] = imp
	return
}

// ORTBImpVideoExt will read and set ortb Imp.Video.Ext parameter
func (o *OpenRTB) ORTBImpVideoExt(key string, value *string) (err error) {
	imp, ok := o.ortb["imp"].([]map[string]interface{})
	if !ok {
		imp = []map[string]interface{}{}
	}
	video, ok := imp[0]["video"].(map[string]interface{})
	if !ok {
		video = map[string]interface{}{}
	}
	videoExt, ok := video["ext"].(map[string]interface{})
	if !ok {
		videoExt = map[string]interface{}{}
	}

	SetValue(videoExt, key, value)
	video["ext"] = videoExt
	imp[0]["video"] = video
	o.ortb["imp"] = imp
	return
}

// ORTBSiteExt will read and set ortb Site.Ext parameter
func (o *OpenRTB) ORTBSiteExt(key string, value *string) (err error) {
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	siteExt, ok := site["ext"].(map[string]interface{})
	if !ok {
		siteExt = map[string]interface{}{}
	}
	SetValue(siteExt, key, value)
	site["ext"] = siteExt
	o.ortb["site"] = site
	return
}

// ORTBSiteContentNetworkExt will read and set ortb Site.Content.Network.Ext parameter
func (o *OpenRTB) ORTBSiteContentNetworkExt(key string, value *string) (err error) {
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	network, ok := content["network"].(map[string]interface{})
	if !ok {
		network = map[string]interface{}{}
	}
	networkExt, ok := network["ext"].(map[string]interface{})
	if !ok {
		networkExt = map[string]interface{}{}
	}
	SetValue(networkExt, key, value)
	network["ext"] = networkExt
	content["network"] = network
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBSiteContentChannelExt will read and set ortb Site.Content.Channel.Ext parameter
func (o *OpenRTB) ORTBSiteContentChannelExt(key string, value *string) (err error) {
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	channel, ok := content["channel"].(map[string]interface{})
	if !ok {
		channel = map[string]interface{}{}
	}
	channelExt, ok := channel["ext"].(map[string]interface{})
	if !ok {
		channelExt = map[string]interface{}{}
	}
	SetValue(channelExt, key, value)
	channel["ext"] = channelExt
	content["channel"] = channel
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBAppExt will read and set ortb App.Ext parameter
func (o *OpenRTB) ORTBAppExt(key string, value *string) (err error) {
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	appExt, ok := app["ext"].(map[string]interface{})
	if !ok {
		appExt = map[string]interface{}{}
	}
	SetValue(appExt, key, value)
	app["ext"] = appExt
	o.ortb["app"] = app
	return
}

// ORTBAppContentNetworkExt will read and set ortb App.Content.Network.Ext parameter
func (o *OpenRTB) ORTBAppContentNetworkExt(key string, value *string) (err error) {
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	network, ok := content["network"].(map[string]interface{})
	if !ok {
		network = map[string]interface{}{}
	}
	networkExt, ok := network["ext"].(map[string]interface{})
	if !ok {
		networkExt = map[string]interface{}{}
	}
	SetValue(networkExt, key, value)
	network["ext"] = networkExt
	content["network"] = network
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentChannelExt will read and set ortb App.Content.Channel.Ext parameter
func (o *OpenRTB) ORTBAppContentChannelExt(key string, value *string) (err error) {
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	channel, ok := content["channel"].(map[string]interface{})
	if !ok {
		channel = map[string]interface{}{}
	}
	channelExt, ok := channel["ext"].(map[string]interface{})
	if !ok {
		channelExt = map[string]interface{}{}
	}
	SetValue(channelExt, key, value)
	channel["ext"] = channelExt
	content["channel"] = channel
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBSitePublisherExt will read and set ortb Site.Publisher.Ext parameter
func (o *OpenRTB) ORTBSitePublisherExt(key string, value *string) (err error) {
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	publisher, ok := site["publisher"].(map[string]interface{})
	if !ok {
		publisher = map[string]interface{}{}
	}
	publisherExt, ok := publisher["ext"].(map[string]interface{})
	if !ok {
		publisherExt = map[string]interface{}{}
	}
	SetValue(publisherExt, key, value)
	publisher["ext"] = publisherExt
	site["publisher"] = publisher
	o.ortb["site"] = site
	return
}

// ORTBSiteContentExt will read and set ortb Site.Content.Ext parameter
func (o *OpenRTB) ORTBSiteContentExt(key string, value *string) (err error) {
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	contentExt, ok := content["ext"].(map[string]interface{})
	if !ok {
		contentExt = map[string]interface{}{}
	}
	SetValue(contentExt, key, value)
	content["ext"] = contentExt
	o.ortb["site"] = site
	return
}

// ORTBSiteContentProducerExt will read and set ortb Site.Content.Producer.Ext parameter
func (o *OpenRTB) ORTBSiteContentProducerExt(key string, value *string) (err error) {
	site, ok := o.ortb["site"].(map[string]interface{})
	if !ok {
		site = map[string]interface{}{}
	}
	content, ok := site["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	producer, ok := content["producer"].(map[string]interface{})
	if !ok {
		producer = map[string]interface{}{}
	}
	producerExt, ok := producer["ext"].(map[string]interface{})
	if !ok {
		producerExt = map[string]interface{}{}
	}
	SetValue(producerExt, key, value)
	producer["ext"] = producerExt
	content["producer"] = producer
	site["content"] = content
	o.ortb["site"] = site
	return
}

// ORTBAppPublisherExt will read and set ortb App.Publisher.Ext parameter
func (o *OpenRTB) ORTBAppPublisherExt(key string, value *string) (err error) {
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	publisher, ok := app["publisher"].(map[string]interface{})
	if !ok {
		publisher = map[string]interface{}{}
	}
	pubExt, ok := publisher["ext"].(map[string]interface{})
	if !ok {
		pubExt = map[string]interface{}{}
	}
	SetValue(pubExt, key, value)
	publisher["ext"] = pubExt
	app["publisher"] = publisher
	o.ortb["app"] = app
	return
}

// ORTBAppContentExt will read and set ortb App.Content.Ext parameter
func (o *OpenRTB) ORTBAppContentExt(key string, value *string) (err error) {
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	cntExt, ok := content["ext"].(map[string]interface{})
	if !ok {
		cntExt = map[string]interface{}{}
	}
	SetValue(cntExt, key, value)
	content["ext"] = cntExt
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBAppContentProducerExt will read and set ortb App.Content.Producer.Ext parameter
func (o *OpenRTB) ORTBAppContentProducerExt(key string, value *string) (err error) {
	app, ok := o.ortb["app"].(map[string]interface{})
	if !ok {
		app = map[string]interface{}{}
	}
	content, ok := app["content"].(map[string]interface{})
	if !ok {
		content = map[string]interface{}{}
	}
	producer, ok := content["producer"].(map[string]interface{})
	if !ok {
		producer = map[string]interface{}{}
	}
	pdcExt, ok := producer["ext"].(map[string]interface{})
	if !ok {
		pdcExt = map[string]interface{}{}
	}
	SetValue(pdcExt, key, value)
	producer["ext"] = pdcExt
	content["producer"] = producer
	app["content"] = content
	o.ortb["app"] = app
	return
}

// ORTBDeviceExt will read and set ortb Device.Ext parameter
func (o *OpenRTB) ORTBDeviceExt(key string, value *string) (err error) {
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	deviceExt, ok := device["ext"].(map[string]interface{})
	if !ok {
		deviceExt = map[string]interface{}{}
	}
	SetValue(deviceExt, key, value)
	device["ext"] = deviceExt
	o.ortb["device"] = device
	return
}

// ORTBDeviceGeoExt will read and set ortb Device.Geo.Ext parameter
func (o *OpenRTB) ORTBDeviceGeoExt(key string, value *string) (err error) {
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	geo, ok := device["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geoExt, ok := geo["ext"].(map[string]interface{})
	if !ok {
		geoExt = map[string]interface{}{}
	}
	SetValue(geoExt, key, value)
	geo["ext"] = geoExt
	device["geo"] = geo
	o.ortb["device"] = device
	return
}

// ORTBUserExt will read and set ortb User.Ext parameter
func (o *OpenRTB) ORTBUserExt(key string, value *string) (err error) {
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	userExt, ok := user["ext"].(map[string]interface{})
	if !ok {
		userExt = map[string]interface{}{}
	}
	userExt[key] = value
	user["ext"] = userExt
	o.ortb["user"] = user
	return
}

// ORTBUserGeoExt will read and set ortb User.Geo.Ext parameter
func (o *OpenRTB) ORTBUserGeoExt(key string, value *string) (err error) {
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	geo, ok := user["geo"].(map[string]interface{})
	if !ok {
		geo = map[string]interface{}{}
	}
	geoExt, ok := geo["ext"].(map[string]interface{})
	if !ok {
		geoExt = map[string]interface{}{}
	}
	SetValue(geoExt, key, value)
	geo["ext"] = geoExt
	user["geo"] = geo
	o.ortb["user"] = user
	return
}

// ORTBUserExtConsent will read and set ortb User.Ext.Consent parameter
func (o *OpenRTB) ORTBDeviceExtIfaType() (err error) {
	val := o.values.Get(ORTBDeviceExtIfaType)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	deviceExt, ok := device["ext"].(map[string]interface{})
	if !ok {
		deviceExt = map[string]interface{}{}
	}
	deviceExt[ORTBExtIfaType] = val
	device["ext"] = deviceExt
	o.ortb["device"] = device
	return
}

// ORTBDeviceExtSessionID will read and set ortb device.Ext.SessionID parameter
func (o *OpenRTB) ORTBDeviceExtSessionID() (err error) {
	val := o.values.Get(ORTBDeviceExtSessionID)
	if len(val) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}

	deviceExt, ok := device["ext"].(map[string]interface{})
	if !ok {
		deviceExt = map[string]interface{}{}
	}
	deviceExt[ORTBExtSessionID] = val
	device["ext"] = deviceExt
	o.ortb["device"] = device
	return
}

// ORTBDeviceExtATTS will read and set ortb device.ext.atts parameter
func (o *OpenRTB) ORTBDeviceExtATTS() (err error) {
	value := o.values.Get(ORTBDeviceExtATTS)
	if len(value) == 0 {
		return
	}
	device, ok := o.ortb["device"].(map[string]interface{})
	if !ok {
		device = map[string]interface{}{}
	}
	deviceExt, ok := device["ext"].(map[string]interface{})
	if !ok {
		deviceExt = map[string]interface{}{}
	}
	deviceExt[ORTBExtATTS] = value
	device["ext"] = deviceExt
	o.ortb["device"] = device
	return
}

// ORTBRequestExtPrebidTransparencyContent will read and set ortb Request.Ext.Prebid.Transparency.Content parameter
func (o *OpenRTB) ORTBRequestExtPrebidTransparencyContent() (err error) {
	contentString := o.values.Get(ORTBRequestExtPrebidTransparencyContent)
	if len(contentString) == 0 {
		return
	}

	requestExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		requestExt = map[string]interface{}{}
	}

	prebidExt, ok := requestExt[ORTBExtPrebid].(map[string]interface{})
	if !ok {
		prebidExt = map[string]interface{}{}
	}

	transparancy, ok := prebidExt[ORTBExtPrebidTransparency].(map[string]interface{})
	if !ok {
		transparancy = map[string]interface{}{}
	}
	transparancy[ORTBExtPrebidTransparencyContent] = contentString
	prebidExt[ORTBExtPrebidTransparency] = transparancy
	requestExt[ORTBExtPrebid] = prebidExt
	o.ortb["ext"] = requestExt
	return
}

// ORTBUserExtEIDS will read and set ortb user.ext.eids parameter
func (o *OpenRTB) ORTBUserExtEIDS() (err error) {
	eidsValue := o.values.Get(ORTBUserExtEIDS)
	if len(eidsValue) == 0 {
		return
	}

	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}

	userExt, ok := user["ext"].(map[string]interface{})
	if !ok {
		userExt = map[string]interface{}{}
	}
	userExt[ORTBExtEIDS] = eidsValue
	user["ext"] = userExt
	o.ortb["user"] = user
	return
}

// ORTBUserExtSessionDuration will read and set ortb User.Ext.sessionduration parameter
func (o *OpenRTB) ORTBUserExtSessionDuration() (err error) {
	valStr := o.values.Get(ORTBUserExtSessionDuration)
	if len(valStr) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	userExt, ok := user["ext"].(map[string]interface{})
	if !ok {
		userExt = map[string]interface{}{}
	}
	userExt[ORTBExtSessionDuration] = valStr
	user["ext"] = userExt
	o.ortb["user"] = user
	return
}

// ORTBUserExtImpDepth will read and set ortb User.Ext.impdepth parameter
func (o *OpenRTB) ORTBUserExtImpDepth() (err error) {
	valStr := o.values.Get(ORTBUserExtImpDepth)
	if len(valStr) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	userExt, ok := user["ext"].(map[string]interface{})
	if !ok {
		userExt = map[string]interface{}{}
	}
	userExt[ORTBExtImpDepth] = valStr
	user["ext"] = userExt
	o.ortb["user"] = user
	return
}

// ORTBUserData will read and set ortb user.data parameter
func (o *OpenRTB) ORTBUserData() (err error) {
	dataValue := o.values.Get(ORTBUserData)
	if len(dataValue) == 0 {
		return
	}
	user, ok := o.ortb["user"].(map[string]interface{})
	if !ok {
		user = map[string]interface{}{}
	}
	user["data"] = dataValue
	o.ortb["user"] = user
	return
}

func (o *OpenRTB) ORTBExtPrebidFloorsEnforceFloorDeals() (err error) {
	enforcementString := o.values.Get(ORTBExtPrebidFloorsEnforcement)
	if len(enforcementString) == 0 {
		return
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	prebidExt, ok := reqExt[ORTBExtPrebid].(map[string]interface{})
	if !ok {
		prebidExt = map[string]interface{}{}
	}

	floors, ok := prebidExt[ORTBExtPrebidFloors].(map[string]interface{})
	if !ok {
		floors = map[string]interface{}{}
	}

	decodedString, err := url.QueryUnescape(enforcementString)
	if err != nil {
		return err
	}

	var enforcement map[string]interface{}
	err = json.Unmarshal([]byte(decodedString), &enforcement)
	if err != nil {
		return err
	}

	floors[ORTBExtFloorEnforcement] = enforcement
	prebidExt[ORTBExtPrebidFloors] = floors
	reqExt[ORTBExtPrebid] = prebidExt
	o.ortb["ext"] = reqExt
	return
}

// ORTBExtPrebidReturnAllBidStatus sets returnallbidstatus
func (o *OpenRTB) ORTBExtPrebidReturnAllBidStatus() (err error) {
	returnAllbidStatus := o.values.Get(ORTBExtPrebidReturnAllBidStatus)
	if len(returnAllbidStatus) == 0 {
		return
	}
	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}
	prebidExt, ok := reqExt[ORTBExtPrebid].(map[string]interface{})
	if !ok {
		prebidExt = map[string]interface{}{}
	}

	if returnAllbidStatus == "1" {
		prebidExt[ReturnAllBidStatus] = true
	} else {
		prebidExt[ReturnAllBidStatus] = false
	}

	reqExt[ORTBExtPrebid] = prebidExt
	o.ortb["ext"] = reqExt
	return nil
}

// ORTBExtPrebidBidderParamsPubmaticCDS sets cds in req.ext.prebid.bidderparams.pubmatic
func (o *OpenRTB) ORTBExtPrebidBidderParamsPubmaticCDS() (err error) {
	cdsData := o.values.Get(ORTBExtPrebidBidderParamsPubmaticCDS)
	if len(cdsData) == 0 {
		return
	}

	decodedString, err := url.QueryUnescape(cdsData)
	if err != nil {
		return err
	}

	var cds map[string]interface{}
	err = json.Unmarshal([]byte(decodedString), &cds)
	if err != nil {
		return err
	}

	reqExt, ok := o.ortb["ext"].(map[string]interface{})
	if !ok {
		reqExt = map[string]interface{}{}
	}

	prebidExt, ok := reqExt[ORTBExtPrebid].(map[string]interface{})
	if !ok {
		prebidExt = map[string]interface{}{}
	}

	bidderParams, ok := prebidExt[ORTBExtPrebidBidderParams].(map[string]interface{})
	if !ok {
		bidderParams = map[string]interface{}{}
	}

	pubmaticBidderParams, ok := bidderParams[models.BidderPubMatic].(map[string]interface{})
	if !ok {
		pubmaticBidderParams = map[string]interface{}{}
	}
	pubmaticBidderParams[models.CustomDimensions] = cds

	bidderParams[models.BidderPubMatic] = pubmaticBidderParams
	prebidExt[ORTBExtPrebidBidderParams] = bidderParams
	reqExt[ORTBExtPrebid] = prebidExt
	o.ortb["ext"] = reqExt
	return
}

/*********************** Regs.Gpp And Regs.GppSid***********************/

// ORTBRegsGpp will read and set ortb Regs.gpp parameter
func (o *OpenRTB) ORTBRegsGpp() (err error) {
	val := o.values.Get(ORTBRegsGpp)
	if len(val) == 0 {
		return
	}
	regs, ok := o.ortb["regs"].(map[string]interface{})
	if !ok {
		regs = map[string]interface{}{}
	}
	regs["gpp"] = val
	o.ortb["regs"] = regs
	return
}

// ORTBRegsGpp will read and set ortb Regs.gpp_sid parameter
func (o *OpenRTB) ORTBRegsGppSid() error {
	val, err := o.values.GetInt8Array(ORTBRegsGppSid, ArraySeparator)
	if len(val) == 0 {
		return err
	}

	regs, ok := o.ortb["regs"].(map[string]interface{})
	if !ok {
		regs = map[string]interface{}{}
	}
	regs["gpp_sid"] = val
	o.ortb["regs"] = regs
	return nil
}
