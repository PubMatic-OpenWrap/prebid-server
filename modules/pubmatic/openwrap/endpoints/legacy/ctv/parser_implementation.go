package ctv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"git.pubmatic.com/PubMatic/go-common/logger"
	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/adcom1"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/util/ptrutil"

	v26 "github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/endpoints/legacy/openrtb/v26"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	uuid "github.com/satori/go.uuid"
)

type OpenRTB struct {
	request *http.Request
	values  URLValues
	ortb    *openrtb2.BidRequest
}

// NewOpenRTB Returns New ORTB Object of Version 2.5
func NewOpenRTB(request *http.Request) Parser {
	request.ParseForm()

	obj := &OpenRTB{
		request: request,
		values:  URLValues{Values: request.Form},
		ortb: &openrtb2.BidRequest{
			Imp: []openrtb2.Imp{
				{},
			},
		},
	}

	return obj
}

/********************** Helper Functions **********************/

// ParseORTBRequest this will parse ortb request by reading parserMap and calling respective function for mapped parameter
func (o *OpenRTB) ParseORTBRequest(parserMap *ParserMap) (*openrtb2.BidRequest, error) {
	for k, value := range o.values.Values {
		if len(value) > 0 && len(value[0]) > 0 {
			if parser, ok := parserMap.KeyMapping[k]; ok {
				if err := parser(o); err != nil {
					return nil, err
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
								return nil, err
							}
						}
					}
				} else if _, ok = parserMap.IgnoreList[k]; !ok {
					glog.Warningf("Key Not Present : Key:[%v] Value:[%v]", k, value)
				}
			}
		}
	}

	o.formORTBRequest()
	return o.ortb, nil
}

// formORTBRequest this will generate bidrequestID or impressionID if not present
func (o *OpenRTB) formORTBRequest() {
	if len(o.ortb.ID) == 0 {
		o.ortb.ID = uuid.NewV4().String()
	}

	if len(o.ortb.Imp[0].ID) == 0 {
		o.ortb.Imp[0].ID = uuid.NewV4().String()
	}
}

/*********************** BidRequest ***********************/

// ORTBBidRequestID will read and set ortb BidRequest.ID parameter
func (o *OpenRTB) ORTBBidRequestID() (err error) {
	val, ok := o.values.GetString(ORTBBidRequestID)
	if !ok {
		o.ortb.ID = uuid.NewV4().String()
	} else {
		o.ortb.ID = val
	}
	return
}

// ORTBBidRequestTest will read and set ortb BidRequest.Test parameter
func (o *OpenRTB) ORTBBidRequestTest() (err error) {
	val, ok, err := o.values.GetInt(ORTBBidRequestTest)
	if ok {
		o.ortb.Test = int8(val)
	}
	return
}

// ORTBBidRequestAt will read and set ortb BidRequest.At parameter
func (o *OpenRTB) ORTBBidRequestAt() (err error) {
	val, ok, err := o.values.GetInt(ORTBBidRequestAt)
	if ok {
		o.ortb.AT = int64(val)
	}
	return
}

// ORTBBidRequestTmax will read and set ortb BidRequest.Tmax parameter
func (o *OpenRTB) ORTBBidRequestTmax() (err error) {
	val, ok, err := o.values.GetInt(ORTBBidRequestTmax)
	if ok {
		o.ortb.TMax = int64(val)
	}
	return
}

// ORTBBidRequestWseat will read and set ortb BidRequest.Wseat parameter
func (o *OpenRTB) ORTBBidRequestWseat() (err error) {
	o.ortb.WSeat = o.values.GetStringArray(ORTBBidRequestWseat, ArraySeparator)
	return
}

// ORTBBidRequestWlang will read and set ortb BidRequest.Wlang Parameter
func (o *OpenRTB) ORTBBidRequestWlang() (err error) {
	o.ortb.WLang = o.values.GetStringArray(ORTBBidRequestWlang, ArraySeparator)
	return
}

// ORTBBidRequestBseat will read and set ortb BidRequest.Bseat Parameter
func (o *OpenRTB) ORTBBidRequestBseat() (err error) {
	o.ortb.BSeat = o.values.GetStringArray(ORTBBidRequestBseat, ArraySeparator)
	return
}

// ORTBBidRequestAllImps will read and set ortb BidRequest.AllImps parameter
func (o *OpenRTB) ORTBBidRequestAllImps() (err error) {
	val, ok, err := o.values.GetInt(ORTBBidRequestAllImps)
	if ok {
		o.ortb.AllImps = int8(val)
	}
	return
}

// ORTBBidRequestCur will read and set ortb BidRequest.Cur parameter
func (o *OpenRTB) ORTBBidRequestCur() (err error) {
	o.ortb.Cur = o.values.GetStringArray(ORTBBidRequestCur, ArraySeparator)
	return
}

// ORTBBidRequestBcat will read and set ortb BidRequest.Bcat parameter
func (o *OpenRTB) ORTBBidRequestBcat() (err error) {
	o.ortb.BCat = o.values.GetStringArray(ORTBBidRequestBcat, ArraySeparator)
	return
}

// ORTBBidRequestBadv will read and set ortb BidRequest.Badv parameter
func (o *OpenRTB) ORTBBidRequestBadv() (err error) {
	o.ortb.BAdv = o.values.GetStringArray(ORTBBidRequestBadv, ArraySeparator)
	return
}

// ORTBBidRequestBapp will read and set ortb BidRequest.Bapp parameter
func (o *OpenRTB) ORTBBidRequestBapp() (err error) {
	o.ortb.BApp = o.values.GetStringArray(ORTBBidRequestBapp, ArraySeparator)
	return
}

/*********************** Source ***********************/

// ORTBSourceFD will read and set ortb Source.FD parameter
func (o *OpenRTB) ORTBSourceFD() (err error) {
	val, ok, err := o.values.GetInt(ORTBSourceFD)
	if !ok || err != nil {
		return
	}
	if o.ortb.Source == nil {
		o.ortb.Source = &openrtb2.Source{}
	}
	o.ortb.Source.FD = ptrutil.ToPtr(int8(val))
	return
}

// ORTBSourceTID will read and set ortb Source.TID parameter
func (o *OpenRTB) ORTBSourceTID() (err error) {
	val, ok := o.values.GetString(ORTBSourceTID)
	if !ok {
		return
	}
	if o.ortb.Source == nil {
		o.ortb.Source = &openrtb2.Source{}
	}
	o.ortb.Source.TID = val
	return
}

// ORTBSourcePChain will read and set ortb Source.PChain parameter
func (o *OpenRTB) ORTBSourcePChain() (err error) {
	val, ok := o.values.GetString(ORTBSourcePChain)
	if !ok {
		return
	}
	if o.ortb.Source == nil {
		o.ortb.Source = &openrtb2.Source{}
	}
	o.ortb.Source.PChain = val
	return
}

// ORTBSourceSChain will read and set ortb Source.Ext.SChain parameter
func (o *OpenRTB) ORTBSourceSChain() (err error) {
	sChainString, ok := o.values.GetString(ORTBSourceSChain)
	if !ok {
		return nil
	}
	var sChain *openrtb2.SupplyChain
	sChain, err = openrtb_ext.DeserializeSupplyChain(sChainString)
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

	if o.ortb.Source == nil {
		o.ortb.Source = &openrtb2.Source{}
	}

	o.ortb.Source.SChain = sChain

	return
}

/*********************** Site ***********************/

// ORTBSiteID will read and set ortb Site.ID parameter
func (o *OpenRTB) ORTBSiteID() (err error) {
	val, ok := o.values.GetString(ORTBSiteID)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.ID = val
	return
}

// ORTBSiteName will read and set ortb Site.Name parameter
func (o *OpenRTB) ORTBSiteName() (err error) {
	val, ok := o.values.GetString(ORTBSiteName)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.Name = val
	return
}

// ORTBSiteDomain will read and set ortb Site.Domain parameter
func (o *OpenRTB) ORTBSiteDomain() (err error) {
	val, ok := o.values.GetString(ORTBSiteDomain)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.Domain = val
	return
}

// ORTBSitePage will read and set ortb Site.Page parameter
func (o *OpenRTB) ORTBSitePage() (err error) {
	val, ok := o.values.GetString(ORTBSitePage)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.Page = val
	return
}

// ORTBSiteRef will read and set ortb Site.Ref parameter
func (o *OpenRTB) ORTBSiteRef() (err error) {
	val, ok := o.values.GetString(ORTBSiteRef)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.Ref = val
	return
}

// ORTBSiteSearch will read and set ortb Site.Search parameter
func (o *OpenRTB) ORTBSiteSearch() (err error) {
	val, ok := o.values.GetString(ORTBSiteSearch)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.Search = val
	return
}

// ORTBSiteMobile will read and set ortb Site.Mobile parameter
func (o *OpenRTB) ORTBSiteMobile() (err error) {
	val, ok, err := o.values.GetInt(ORTBSiteMobile)
	if !ok || err != nil {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.Mobile = ptrutil.ToPtr(int8(val))
	return
}

// ORTBSiteCat will read and set ortb Site.Cat parameter
func (o *OpenRTB) ORTBSiteCat() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.Cat = o.values.GetStringArray(ORTBSiteCat, ArraySeparator)
	return
}

// ORTBSiteSectionCat will read and set ortb Site.SectionCat parameter
func (o *OpenRTB) ORTBSiteSectionCat() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.SectionCat = o.values.GetStringArray(ORTBSiteSectionCat, ArraySeparator)
	return
}

// ORTBSitePageCat will read and set ortb Site.PageCat parameter
func (o *OpenRTB) ORTBSitePageCat() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.PageCat = o.values.GetStringArray(ORTBSitePageCat, ArraySeparator)
	return
}

// ORTBSitePrivacyPolicy will read and set ortb Site.PrivacyPolicy parameter
func (o *OpenRTB) ORTBSitePrivacyPolicy() (err error) {
	val, ok, err := o.values.GetInt(ORTBSitePrivacyPolicy)
	if !ok || err != nil {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.PrivacyPolicy = ptrutil.ToPtr(int8(val))
	return
}

// ORTBSiteKeywords will read and set ortb Site.Keywords parameter
func (o *OpenRTB) ORTBSiteKeywords() (err error) {
	val, ok := o.values.GetString(ORTBSiteKeywords)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	o.ortb.Site.Keywords = val
	return
}

/*********************** Site.Publisher ***********************/

// ORTBSitePublisherID will read and set ortb Site.Publisher.ID parameter
func (o *OpenRTB) ORTBSitePublisherID() (err error) {
	val, ok := o.values.GetString(ORTBSitePublisherID)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Publisher == nil {
		o.ortb.Site.Publisher = &openrtb2.Publisher{}
	}
	o.ortb.Site.Publisher.ID = val
	return
}

// ORTBSitePublisherName will read and set ortb Site.Publisher.Name parameter
func (o *OpenRTB) ORTBSitePublisherName() (err error) {
	val, ok := o.values.GetString(ORTBSitePublisherName)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Publisher == nil {
		o.ortb.Site.Publisher = &openrtb2.Publisher{}
	}
	o.ortb.Site.Publisher.Name = val
	return
}

// ORTBSitePublisherCat will read and set ortb Site.Publisher.Cat parameter
func (o *OpenRTB) ORTBSitePublisherCat() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Publisher == nil {
		o.ortb.Site.Publisher = &openrtb2.Publisher{}
	}
	o.ortb.Site.Publisher.Cat = o.values.GetStringArray(ORTBSitePublisherCat, ArraySeparator)
	return
}

// ORTBSitePublisherDomain will read and set ortb Site.Publisher.Domain parameter
func (o *OpenRTB) ORTBSitePublisherDomain() (err error) {
	val, ok := o.values.GetString(ORTBSitePublisherDomain)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Publisher == nil {
		o.ortb.Site.Publisher = &openrtb2.Publisher{}
	}
	o.ortb.Site.Publisher.Domain = val
	return
}

/********************** Site.Content **********************/

// ORTBSiteContentID will read and set ortb Site.Content.ID parameter
func (o *OpenRTB) ORTBSiteContentID() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentID)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.ID = val
	return
}

// ORTBSiteContentEpisode will read and set ortb Site.Content.Episode parameter
func (o *OpenRTB) ORTBSiteContentEpisode() (err error) {
	val, ok, err := o.values.GetInt(ORTBSiteContentEpisode)
	if !ok || err != nil {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Episode = int64(val)
	return
}

// ORTBSiteContentTitle will read and set ortb Site.Content.Title parameter
func (o *OpenRTB) ORTBSiteContentTitle() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentTitle)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Title = val
	return
}

// ORTBSiteContentSeries will read and set ortb Site.Content.Series parameter
func (o *OpenRTB) ORTBSiteContentSeries() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentSeries)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Series = val
	return
}

// ORTBSiteContentSeason will read and set ortb Site.Content.Season parameter
func (o *OpenRTB) ORTBSiteContentSeason() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentSeason)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Season = val
	return
}

// ORTBSiteContentArtist will read and set ortb Site.Content.Artist parameter
func (o *OpenRTB) ORTBSiteContentArtist() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentArtist)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Artist = val
	return
}

// ORTBSiteContentGenre will read and set ortb Site.Content.Genre parameter
func (o *OpenRTB) ORTBSiteContentGenre() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentGenre)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Genre = val
	return
}

// ORTBSiteContentAlbum will read and set ortb Site.Content.Album parameter
func (o *OpenRTB) ORTBSiteContentAlbum() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentAlbum)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Album = val
	return
}

// ORTBSiteContentIsRc will read and set ortb Site.Content.IsRc parameter
func (o *OpenRTB) ORTBSiteContentIsRc() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentIsRc)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.ISRC = val
	return
}

// ORTBSiteContentURL will read and set ortb Site.Content.URL parameter
func (o *OpenRTB) ORTBSiteContentURL() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentURL)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.URL = val
	return
}

// ORTBSiteContentCat will read and set ortb Site.Content.Cat parameter
func (o *OpenRTB) ORTBSiteContentCat() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Cat = o.values.GetStringArray(ORTBSiteContentCat, ArraySeparator)
	return
}

// ORTBSiteContentProdQ will read and set ortb Site.Content.ProdQ parameter
func (o *OpenRTB) ORTBSiteContentProdQ() (err error) {
	val, ok, err := o.values.GetInt(ORTBSiteContentProdQ)
	if !ok || err != nil {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	prodQ := adcom1.ProductionQuality(val)
	o.ortb.Site.Content.ProdQ = &prodQ
	return
}

// ORTBSiteContentVideoQuality will read and set ortb Site.Content.VideoQuality parameter
func (o *OpenRTB) ORTBSiteContentVideoQuality() (err error) {
	val, ok, err := o.values.GetInt(ORTBSiteContentVideoQuality)
	if !ok || err != nil {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	videoQuality := adcom1.ProductionQuality(val)
	o.ortb.Site.Content.VideoQuality = &videoQuality
	return

}

// ORTBSiteContentContext will read and set ortb Site.Content.Context parameter
func (o *OpenRTB) ORTBSiteContentContext() (err error) {
	val, ok, err := o.values.GetInt(ORTBSiteContentContext)
	if !ok || err != nil {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Context = adcom1.ContentContext(val)
	return
}

// ORTBSiteContentContentRating will read and set ortb Site.Content.ContentRating parameter
func (o *OpenRTB) ORTBSiteContentContentRating() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentContentRating)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.ContentRating = val
	return
}

// ORTBSiteContentUserRating will read and set ortb Site.Content.UserRating parameter
func (o *OpenRTB) ORTBSiteContentUserRating() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentUserRating)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.UserRating = val
	return
}

// ORTBSiteContentQaGmeDiarating will read and set ortb Site.Content.QaGmeDiarating parameter
func (o *OpenRTB) ORTBSiteContentQaGmeDiarating() (err error) {
	val, ok, err := o.values.GetInt(ORTBSiteContentQaGmeDiarating)
	if !ok || err != nil {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.QAGMediaRating = adcom1.MediaRating(val)
	return

}

// ORTBSiteContentKeywords will read and set ortb Site.Content.Keywords parameter
func (o *OpenRTB) ORTBSiteContentKeywords() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentKeywords)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Keywords = val
	return
}

// ORTBSiteContentLiveStream will read and set ortb Site.Content.LiveStream parameter
func (o *OpenRTB) ORTBSiteContentLiveStream() (err error) {
	val, ok, err := o.values.GetInt(ORTBSiteContentLiveStream)
	if !ok || err != nil {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.LiveStream = ptrutil.ToPtr(int8(val))
	return
}

// ORTBSiteContentSourceRelationship will read and set ortb Site.Content.SourceRelationship parameter
func (o *OpenRTB) ORTBSiteContentSourceRelationship() (err error) {
	val, ok, err := o.values.GetInt(ORTBSiteContentSourceRelationship)
	if !ok || err != nil {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.SourceRelationship = ptrutil.ToPtr(int8(val))
	return
}

// ORTBSiteContentLen will read and set ortb Site.Content.Len parameter
func (o *OpenRTB) ORTBSiteContentLen() (err error) {
	val, ok, err := o.values.GetInt(ORTBSiteContentLen)
	if !ok || err != nil {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Len = int64(val)
	return
}

// ORTBSiteContentLanguage will read and set ortb Site.Content.Language parameter
func (o *OpenRTB) ORTBSiteContentLanguage() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentLanguage)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Language = val
	return
}

// ORTBSiteContentEmbeddable will read and set ortb Site.Content.Embeddable parameter
func (o *OpenRTB) ORTBSiteContentEmbeddable() (err error) {
	val, ok, err := o.values.GetInt(ORTBSiteContentEmbeddable)
	if !ok || err != nil {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	o.ortb.Site.Content.Embeddable = ptrutil.ToPtr(int8(val))
	return
}

/********************** Site.Content.Network **********************/

// ORTBSiteContentNetworkID will read and set ortb Site.Content.Network.Id parameter
func (o *OpenRTB) ORTBSiteContentNetworkID() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Network == nil {
		o.ortb.Site.Content.Network = &openrtb2.Network{}
	}
	o.ortb.Site.Content.Network.ID = o.values.Get(ORTBSiteContentNetworkID)
	return
}

// ORTBSiteContentNetworkName will read and set ortb Site.Content.Network.Name parameter
func (o *OpenRTB) ORTBSiteContentNetworkName() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Network == nil {
		o.ortb.Site.Content.Network = &openrtb2.Network{}
	}
	o.ortb.Site.Content.Network.Name = o.values.Get(ORTBSiteContentNetworkName)
	return
}

// ORTBSiteContentNetworkDomain will read and set ortb Site.Content.Network.Domain parameter
func (o *OpenRTB) ORTBSiteContentNetworkDomain() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Network == nil {
		o.ortb.Site.Content.Network = &openrtb2.Network{}
	}
	o.ortb.Site.Content.Network.Domain = o.values.Get(ORTBSiteContentNetworkDomain)
	return
}

/********************** Site.Content.Channel **********************/

// ORTBSiteContentChannelID will read and set ortb Site.Content.Channel.Id parameter
func (o *OpenRTB) ORTBSiteContentChannelID() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Channel == nil {
		o.ortb.Site.Content.Channel = &openrtb2.Channel{}
	}
	o.ortb.Site.Content.Channel.ID = o.values.Get(ORTBSiteContentChannelID)
	return
}

// ORTBSiteContentChannelName will read and set ortb Site.Content.Channel.Name parameter
func (o *OpenRTB) ORTBSiteContentChannelName() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Channel == nil {
		o.ortb.Site.Content.Channel = &openrtb2.Channel{}
	}
	o.ortb.Site.Content.Channel.Name = o.values.Get(ORTBSiteContentChannelName)
	return
}

// ORTBSiteContentChannelDomain will read and set ortb Site.Content.Channel.Domain parameter
func (o *OpenRTB) ORTBSiteContentChannelDomain() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Channel == nil {
		o.ortb.Site.Content.Channel = &openrtb2.Channel{}
	}
	o.ortb.Site.Content.Channel.Domain = o.values.Get(ORTBSiteContentChannelDomain)
	return
}

/********************** Site.Content.Producer **********************/

// ORTBSiteContentProducerID will read and set ortb Site.Content.Producer.ID parameter
func (o *OpenRTB) ORTBSiteContentProducerID() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentProducerID)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Producer == nil {
		o.ortb.Site.Content.Producer = &openrtb2.Producer{}
	}
	o.ortb.Site.Content.Producer.ID = val
	return
}

// ORTBSiteContentProducerName will read and set ortb Site.Content.Producer.Name parameter
func (o *OpenRTB) ORTBSiteContentProducerName() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentProducerName)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Producer == nil {
		o.ortb.Site.Content.Producer = &openrtb2.Producer{}
	}
	o.ortb.Site.Content.Producer.Name = val
	return
}

// ORTBSiteContentProducerCat will read and set ortb Site.Content.Producer.Cat parameter
func (o *OpenRTB) ORTBSiteContentProducerCat() (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Producer == nil {
		o.ortb.Site.Content.Producer = &openrtb2.Producer{}
	}
	o.ortb.Site.Content.Producer.Cat = o.values.GetStringArray(ORTBSiteContentProducerCat, ArraySeparator)
	return
}

// ORTBSiteContentProducerDomain will read and set ortb Site.Content.Producer.Domain parameter
func (o *OpenRTB) ORTBSiteContentProducerDomain() (err error) {
	val, ok := o.values.GetString(ORTBSiteContentProducerDomain)
	if !ok {
		return
	}
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Producer == nil {
		o.ortb.Site.Content.Producer = &openrtb2.Producer{}
	}
	o.ortb.Site.Content.Producer.Domain = val
	return
}

/*********************** App ***********************/

// ORTBAppID will read and set ortb App.ID parameter
func (o *OpenRTB) ORTBAppID() (err error) {
	val, ok := o.values.GetString(ORTBAppID)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.ID = val
	return
}

// ORTBAppName will read and set ortb App.Name parameter
func (o *OpenRTB) ORTBAppName() (err error) {
	val, ok := o.values.GetString(ORTBAppName)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.Name = val
	return
}

// ORTBAppBundle will read and set ortb App.Bundle parameter
func (o *OpenRTB) ORTBAppBundle() (err error) {
	val, ok := o.values.GetString(ORTBAppBundle)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.Bundle = val
	return
}

// ORTBAppDomain will read and set ortb App.Domain parameter
func (o *OpenRTB) ORTBAppDomain() (err error) {
	val, ok := o.values.GetString(ORTBAppDomain)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.Domain = val
	return
}

// ORTBAppStoreURL will read and set ortb App.StoreURL parameter
func (o *OpenRTB) ORTBAppStoreURL() (err error) {
	val, ok := o.values.GetString(ORTBAppStoreURL)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.StoreURL = val
	return
}

// ORTBAppVer will read and set ortb App.Ver parameter
func (o *OpenRTB) ORTBAppVer() (err error) {
	val, ok := o.values.GetString(ORTBAppVer)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.Ver = val
	return
}

// ORTBAppPaid will read and set ortb App.Paid parameter
func (o *OpenRTB) ORTBAppPaid() (err error) {
	val, ok, err := o.values.GetInt(ORTBAppPaid)
	if !ok || err != nil {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.Paid = ptrutil.ToPtr(int8(val))
	return
}

// ORTBAppCat will read and set ortb App.Cat parameter
func (o *OpenRTB) ORTBAppCat() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.Cat = o.values.GetStringArray(ORTBAppCat, ArraySeparator)
	return
}

// ORTBAppSectionCat will read and set ortb App.SectionCat parameter
func (o *OpenRTB) ORTBAppSectionCat() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.SectionCat = o.values.GetStringArray(ORTBAppSectionCat, ArraySeparator)
	return
}

// ORTBAppPageCat will read and set ortb App.PageCat parameter
func (o *OpenRTB) ORTBAppPageCat() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.PageCat = o.values.GetStringArray(ORTBAppPageCat, ArraySeparator)
	return
}

// ORTBAppPrivacyPolicy will read and set ortb App.PrivacyPolicy parameter
func (o *OpenRTB) ORTBAppPrivacyPolicy() (err error) {
	val, ok, err := o.values.GetInt(ORTBAppPrivacyPolicy)
	if !ok || err != nil {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.PrivacyPolicy = ptrutil.ToPtr(int8(val))
	return
}

// ORTBAppKeywords will read and set ortb App.Keywords parameter
func (o *OpenRTB) ORTBAppKeywords() (err error) {
	val, ok := o.values.GetString(ORTBAppKeywords)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	o.ortb.App.Keywords = val
	return
}

/*********************** App.Publisher ***********************/

// ORTBAppPublisherID will read and set ortb App.Publisher.ID parameter
func (o *OpenRTB) ORTBAppPublisherID() (err error) {
	val, ok := o.values.GetString(ORTBAppPublisherID)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Publisher == nil {
		o.ortb.App.Publisher = &openrtb2.Publisher{}
	}
	o.ortb.App.Publisher.ID = val
	return
}

// ORTBAppPublisherName will read and set ortb App.Publisher.Name parameter
func (o *OpenRTB) ORTBAppPublisherName() (err error) {
	val, ok := o.values.GetString(ORTBAppPublisherName)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Publisher == nil {
		o.ortb.App.Publisher = &openrtb2.Publisher{}
	}
	o.ortb.App.Publisher.Name = val
	return
}

// ORTBAppPublisherCat will read and set ortb App.Publisher.Cat parameter
func (o *OpenRTB) ORTBAppPublisherCat() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Publisher == nil {
		o.ortb.App.Publisher = &openrtb2.Publisher{}
	}
	o.ortb.App.Publisher.Cat = o.values.GetStringArray(ORTBAppPublisherCat, ArraySeparator)
	return
}

// ORTBAppPublisherDomain will read and set ortb App.Publisher.Domain parameter
func (o *OpenRTB) ORTBAppPublisherDomain() (err error) {
	val, ok := o.values.GetString(ORTBAppPublisherDomain)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Publisher == nil {
		o.ortb.App.Publisher = &openrtb2.Publisher{}
	}
	o.ortb.App.Publisher.Domain = val
	return
}

/********************** App.Content **********************/

// ORTBAppContentID will read and set ortb App.Content.ID parameter
func (o *OpenRTB) ORTBAppContentID() (err error) {
	val, ok := o.values.GetString(ORTBAppContentID)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.ID = val
	return
}

// ORTBAppContentEpisode will read and set ortb App.Content.Episode parameter
func (o *OpenRTB) ORTBAppContentEpisode() (err error) {
	val, ok, err := o.values.GetInt(ORTBAppContentEpisode)
	if !ok || err != nil {
		return
	}

	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Episode = int64(val)
	return
}

// ORTBAppContentTitle will read and set ortb App.Content.Title parameter
func (o *OpenRTB) ORTBAppContentTitle() (err error) {
	val, ok := o.values.GetString(ORTBAppContentTitle)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Title = val
	return
}

// ORTBAppContentSeries will read and set ortb App.Content.Series parameter
func (o *OpenRTB) ORTBAppContentSeries() (err error) {
	val, ok := o.values.GetString(ORTBAppContentSeries)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Series = val
	return
}

// ORTBAppContentSeason will read and set ortb App.Content.Season parameter
func (o *OpenRTB) ORTBAppContentSeason() (err error) {
	val, ok := o.values.GetString(ORTBAppContentSeason)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Season = val
	return
}

// ORTBAppContentArtist will read and set ortb App.Content.Artist parameter
func (o *OpenRTB) ORTBAppContentArtist() (err error) {
	val, ok := o.values.GetString(ORTBAppContentArtist)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Artist = val
	return
}

// ORTBAppContentGenre will read and set ortb App.Content.Genre parameter
func (o *OpenRTB) ORTBAppContentGenre() (err error) {
	val, ok := o.values.GetString(ORTBAppContentGenre)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Genre = val
	return
}

// ORTBAppContentAlbum will read and set ortb App.Content.Album parameter
func (o *OpenRTB) ORTBAppContentAlbum() (err error) {
	val, ok := o.values.GetString(ORTBAppContentAlbum)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Album = val
	return
}

// ORTBAppContentIsRc will read and set ortb App.Content.IsRc parameter
func (o *OpenRTB) ORTBAppContentIsRc() (err error) {
	val, ok := o.values.GetString(ORTBAppContentIsRc)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.ISRC = val
	return
}

// ORTBAppContentURL will read and set ortb App.Content.URL parameter
func (o *OpenRTB) ORTBAppContentURL() (err error) {
	val, ok := o.values.GetString(ORTBAppContentURL)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.URL = val
	return
}

// ORTBAppContentCat will read and set ortb App.Content.Cat parameter
func (o *OpenRTB) ORTBAppContentCat() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Cat = o.values.GetStringArray(ORTBAppContentCat, ArraySeparator)
	return
}

// ORTBAppContentProdQ will read and set ortb App.Content.ProdQ parameter
func (o *OpenRTB) ORTBAppContentProdQ() (err error) {
	val, ok, err := o.values.GetInt(ORTBAppContentProdQ)
	if !ok || err != nil {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	prodQ := adcom1.ProductionQuality(val)
	o.ortb.App.Content.ProdQ = &prodQ

	return err
}

// ORTBAppContentVideoQuality will read and set ortb App.Content.VideoQuality parameter
func (o *OpenRTB) ORTBAppContentVideoQuality() (err error) {
	val, ok, err := o.values.GetInt(ORTBAppContentVideoQuality)
	if !ok || err != nil {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	videoQuality := adcom1.ProductionQuality(val)
	o.ortb.App.Content.VideoQuality = &videoQuality
	return
}

// ORTBAppContentContext will read and set ortb App.Content.Context parameter
func (o *OpenRTB) ORTBAppContentContext() (err error) {
	val, ok, err := o.values.GetInt(ORTBAppContentContext)
	if !ok || err != nil {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	context := adcom1.ContentContext(val)
	o.ortb.App.Content.Context = context
	return
}

// ORTBAppContentContentRating will read and set ortb App.Content.ContentRating parameter
func (o *OpenRTB) ORTBAppContentContentRating() (err error) {
	val, ok := o.values.GetString(ORTBAppContentContentRating)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.ContentRating = val
	return
}

// ORTBAppContentUserRating will read and set ortb App.Content.UserRating parameter
func (o *OpenRTB) ORTBAppContentUserRating() (err error) {
	val, ok := o.values.GetString(ORTBAppContentUserRating)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.UserRating = val
	return
}

// ORTBAppContentQaGmeDiarating will read and set ortb App.Content.QaGmeDiarating parameter
func (o *OpenRTB) ORTBAppContentQaGmeDiarating() (err error) {
	val, ok, err := o.values.GetInt(ORTBAppContentQaGmeDiarating)
	if !ok || err != nil {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	qagMediaRating := adcom1.MediaRating(val)
	o.ortb.App.Content.QAGMediaRating = qagMediaRating

	return err
}

// ORTBAppContentKeywords will read and set ortb App.Content.Keywords parameter
func (o *OpenRTB) ORTBAppContentKeywords() (err error) {
	val, ok := o.values.GetString(ORTBAppContentKeywords)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Keywords = val
	return
}

// ORTBAppContentLiveStream will read and set ortb App.Content.LiveStream parameter
func (o *OpenRTB) ORTBAppContentLiveStream() (err error) {
	val, ok, err := o.values.GetInt(ORTBAppContentLiveStream)
	if !ok || err != nil {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.LiveStream = ptrutil.ToPtr(int8(val))
	return
}

// ORTBAppContentSourceRelationship will read and set ortb App.Content.SourceRelationship parameter
func (o *OpenRTB) ORTBAppContentSourceRelationship() (err error) {
	val, ok, err := o.values.GetInt(ORTBAppContentSourceRelationship)
	if !ok || err != nil {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.SourceRelationship = ptrutil.ToPtr(int8(val))
	return
}

// ORTBAppContentLen will read and set ortb App.Content.Len parameter
func (o *OpenRTB) ORTBAppContentLen() (err error) {
	val, ok, err := o.values.GetInt(ORTBAppContentLen)
	if !ok || err != nil {
		return
	}

	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Len = int64(val)
	return
}

// ORTBAppContentLanguage will read and set ortb App.Content.Language parameter
func (o *OpenRTB) ORTBAppContentLanguage() (err error) {
	val, ok := o.values.GetString(ORTBAppContentLanguage)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Language = val
	return
}

// ORTBAppContentEmbeddable will read and set ortb App.Content.Embeddable parameter
func (o *OpenRTB) ORTBAppContentEmbeddable() (err error) {
	val, ok, err := o.values.GetInt(ORTBAppContentEmbeddable)
	if !ok || err != nil {
		return
	}

	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	o.ortb.App.Content.Embeddable = ptrutil.ToPtr(int8(val))
	return
}

/********************** App.Content.Network **********************/

// ORTBAppContentNetworkID will read and set ortb App.Content.Network.Id parameter
func (o *OpenRTB) ORTBAppContentNetworkID() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Network == nil {
		o.ortb.App.Content.Network = &openrtb2.Network{}
	}
	o.ortb.App.Content.Network.ID = o.values.Get(ORTBAppContentNetworkID)
	return
}

// ORTBAppContentNetworkName will read and set ortb App.Content.Network.Name parameter
func (o *OpenRTB) ORTBAppContentNetworkName() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Network == nil {
		o.ortb.App.Content.Network = &openrtb2.Network{}
	}
	o.ortb.App.Content.Network.Name = o.values.Get(ORTBAppContentNetworkName)
	return
}

// ORTBAppContentNetworkDomain will read and set ortb App.Content.Network.Domain parameter
func (o *OpenRTB) ORTBAppContentNetworkDomain() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Network == nil {
		o.ortb.App.Content.Network = &openrtb2.Network{}
	}
	o.ortb.App.Content.Network.Domain = o.values.Get(ORTBAppContentNetworkDomain)
	return
}

/********************** App.Content.Channel **********************/

// ORTBAppContentChannelID will read and set ortb App.Content.Channel.Id parameter
func (o *OpenRTB) ORTBAppContentChannelID() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Channel == nil {
		o.ortb.App.Content.Channel = &openrtb2.Channel{}
	}
	o.ortb.App.Content.Channel.ID = o.values.Get(ORTBAppContentChannelID)
	return
}

// ORTBAppContentChannelName will read and set ortb App.Content.Channel.Name parameter
func (o *OpenRTB) ORTBAppContentChannelName() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Channel == nil {
		o.ortb.App.Content.Channel = &openrtb2.Channel{}
	}
	o.ortb.App.Content.Channel.Name = o.values.Get(ORTBAppContentChannelName)
	return
}

// ORTBAppContentChannelDomain will read and set ortb App.Content.Channel.Domain parameter
func (o *OpenRTB) ORTBAppContentChannelDomain() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Channel == nil {
		o.ortb.App.Content.Channel = &openrtb2.Channel{}
	}
	o.ortb.App.Content.Channel.Domain = o.values.Get(ORTBAppContentChannelDomain)
	return
}

/********************** App.Content.Producer **********************/

// ORTBAppContentProducerID will read and set ortb App.Content.Producer.ID parameter
func (o *OpenRTB) ORTBAppContentProducerID() (err error) {
	val, ok := o.values.GetString(ORTBAppContentProducerID)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Producer == nil {
		o.ortb.App.Content.Producer = &openrtb2.Producer{}
	}
	o.ortb.App.Content.Producer.ID = val
	return
}

// ORTBAppContentProducerName will read and set ortb App.Content.Producer.Name parameter
func (o *OpenRTB) ORTBAppContentProducerName() (err error) {
	val, ok := o.values.GetString(ORTBAppContentProducerName)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Producer == nil {
		o.ortb.App.Content.Producer = &openrtb2.Producer{}
	}
	o.ortb.App.Content.Producer.Name = val
	return
}

// ORTBAppContentProducerCat will read and set ortb App.Content.Producer.Cat parameter
func (o *OpenRTB) ORTBAppContentProducerCat() (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Producer == nil {
		o.ortb.App.Content.Producer = &openrtb2.Producer{}
	}
	o.ortb.App.Content.Producer.Cat = o.values.GetStringArray(ORTBAppContentProducerCat, ArraySeparator)
	return
}

// ORTBAppContentProducerDomain will read and set ortb App.Content.Producer.Domain parameter
func (o *OpenRTB) ORTBAppContentProducerDomain() (err error) {
	val, ok := o.values.GetString(ORTBAppContentProducerDomain)
	if !ok {
		return
	}
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Producer == nil {
		o.ortb.App.Content.Producer = &openrtb2.Producer{}
	}
	o.ortb.App.Content.Producer.Domain = val
	return
}

/********************** Video **********************/

// ORTBImpVideoMimes will read and set ortb Imp.Video.Mimes parameter
func (o *OpenRTB) ORTBImpVideoMimes() (err error) {
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	o.ortb.Imp[0].Video.MIMEs = o.values.GetStringArray(ORTBImpVideoMimes, ArraySeparator)
	return
}

// ORTBImpVideoMinDuration will read and set ortb Imp.Video.MinDuration parameter
func (o *OpenRTB) ORTBImpVideoMinDuration() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoMinDuration)
	if !ok || err != nil {
		return
	}

	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	o.ortb.Imp[0].Video.MinDuration = int64(val)
	return
}

// ORTBImpVideoMaxDuration will read and set ortb Imp.Video.MaxDuration parameter
func (o *OpenRTB) ORTBImpVideoMaxDuration() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoMaxDuration)
	if !ok || err != nil {
		return
	}

	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	o.ortb.Imp[0].Video.MaxDuration = int64(val)
	return
}

// ORTBImpVideoProtocols will read and set ortb Imp.Video.Protocols parameter
func (o *OpenRTB) ORTBImpVideoProtocols() (err error) {
	protocols, err := o.values.GetIntArray(ORTBImpVideoProtocols, ArraySeparator)
	if err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	// o.ortb.Imp[0].Video.Protocols, err = o.values.GetIntArray(ORTBImpVideoProtocols, ArraySeparator)
	o.ortb.Imp[0].Video.Protocols = v26.GetProtocol(protocols)
	return
}

// ORTBImpVideoPlayerWidth will read and set ortb Imp.Video.PlayerWidth parameter
func (o *OpenRTB) ORTBImpVideoPlayerWidth() (err error) {
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	val, ok, err := o.values.GetInt(ORTBImpVideoPlayerWidth)
	if !ok || err != nil {
		return
	}
	o.ortb.Imp[0].Video.W = ptrutil.ToPtr(int64(val))
	return
}

// ORTBImpVideoPlayerHeight will read and set ortb Imp.Video.PlayerHeight parameter
func (o *OpenRTB) ORTBImpVideoPlayerHeight() (err error) {
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	val, ok, err := o.values.GetInt(ORTBImpVideoPlayerHeight)
	if !ok || err != nil {
		return
	}
	o.ortb.Imp[0].Video.H = ptrutil.ToPtr(int64(val))
	return
}

// ORTBImpVideoStartDelay will read and set ortb Imp.Video.StartDelay parameter
func (o *OpenRTB) ORTBImpVideoStartDelay() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoStartDelay)
	if !ok || err != nil {
		return
	}

	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	startDelay := adcom1.StartDelay(val)
	o.ortb.Imp[0].Video.StartDelay = &startDelay
	return
}

// ORTBImpVideoPlacement will read and set ortb Imp.Video.Placement parameter
func (o *OpenRTB) ORTBImpVideoPlacement() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoPlacement)
	if !ok || err != nil {
		return
	}

	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	placement := adcom1.VideoPlacementSubtype(val)
	o.ortb.Imp[0].Video.Placement = placement
	return
}

func (o *OpenRTB) ORTBImpVideoPlcmt() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoPlcmt)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	plcmt := adcom1.VideoPlcmtSubtype(val)
	o.ortb.Imp[0].Video.Plcmt = plcmt
	return
}

// ORTBImpVideoLinearity will read and set ortb Imp.Video.Linearity parameter
func (o *OpenRTB) ORTBImpVideoLinearity() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoLinearity)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	linearity := adcom1.LinearityMode(val)
	o.ortb.Imp[0].Video.Linearity = linearity
	return
}

// ORTBImpVideoSkip will read and set ortb Imp.Video.Skip parameter
func (o *OpenRTB) ORTBImpVideoSkip() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoSkip)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	val8 := int8(val)
	o.ortb.Imp[0].Video.Skip = &val8
	return
}

// ORTBImpVideoSkipMin will read and set ortb Imp.Video.SkipMin parameter
func (o *OpenRTB) ORTBImpVideoSkipMin() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoSkipMin)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	o.ortb.Imp[0].Video.SkipMin = int64(val)
	return
}

// ORTBImpVideoSkipAfter will read and set ortb Imp.Video.SkipAfter parameter
func (o *OpenRTB) ORTBImpVideoSkipAfter() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoSkipAfter)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	o.ortb.Imp[0].Video.SkipAfter = int64(val)
	return
}

// ORTBImpVideoSequence will read and set ortb Imp.Video.Sequence parameter
func (o *OpenRTB) ORTBImpVideoSequence() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoSequence)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	o.ortb.Imp[0].Video.Sequence = int8(val)
	return
}

// ORTBImpVideoBAttr will read and set ortb Imp.Video.BAttr parameter
func (o *OpenRTB) ORTBImpVideoBAttr() (err error) {
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	bAttr, err := o.values.GetIntArray(ORTBImpVideoBAttr, ArraySeparator)
	if bAttr != nil {
		o.ortb.Imp[0].Video.BAttr = v26.GetCreativeAttributes(bAttr)
	}
	return
}

// ORTBImpVideoMaxExtended will read and set ortb Imp.Video.MaxExtended parameter
func (o *OpenRTB) ORTBImpVideoMaxExtended() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoMaxExtended)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	o.ortb.Imp[0].Video.MaxExtended = int64(val)
	return
}

// ORTBImpVideoMinBitrate will read and set ortb Imp.Video.MinBitrate parameter
func (o *OpenRTB) ORTBImpVideoMinBitrate() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoMinBitrate)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	o.ortb.Imp[0].Video.MinBitRate = int64(val)
	return
}

// ORTBImpVideoMaxBitrate will read and set ortb Imp.Video.MaxBitrate parameter
func (o *OpenRTB) ORTBImpVideoMaxBitrate() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoMaxBitrate)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	o.ortb.Imp[0].Video.MaxBitRate = int64(val)
	return
}

// ORTBImpVideoBoxingAllowed will read and set ortb Imp.Video.BoxingAllowed parameter
func (o *OpenRTB) ORTBImpVideoBoxingAllowed() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoBoxingAllowed)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	o.ortb.Imp[0].Video.BoxingAllowed = ptrutil.ToPtr(int8(val))
	return
}

// ORTBImpVideoPlaybackMethod will read and set ortb Imp.Video.PlaybackMethod parameter
func (o *OpenRTB) ORTBImpVideoPlaybackMethod() (err error) {
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	playbackMethod, err := o.values.GetIntArray(ORTBImpVideoPlaybackMethod, ArraySeparator)
	o.ortb.Imp[0].Video.PlaybackMethod = v26.GetPlaybackMethod(playbackMethod)
	return
}

// ORTBImpVideoDelivery will read and set ortb Imp.Video.Delivery parameter
func (o *OpenRTB) ORTBImpVideoDelivery() (err error) {
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	delivery, err := o.values.GetIntArray(ORTBImpVideoDelivery, ArraySeparator)
	o.ortb.Imp[0].Video.Delivery = v26.GetDeliveryMethod(delivery)
	return
}

// ORTBImpVideoPos will read and set ortb Imp.Video.Pos parameter
func (o *OpenRTB) ORTBImpVideoPos() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoPos)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	position := adcom1.PlacementPosition(val)
	o.ortb.Imp[0].Video.Pos = &position
	return
}

// ORTBImpVideoAPI will read and set ortb Imp.Video.API parameter
func (o *OpenRTB) ORTBImpVideoAPI() (err error) {
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	api, err := o.values.GetIntArray(ORTBImpVideoAPI, ArraySeparator)
	if err == nil {
		o.ortb.Imp[0].Video.API = v26.GetAPIFramework(api)
	}
	return
}

// ORTBImpVideoCompanionType will read and set ortb Imp.Video.CompanionType parameter
func (o *OpenRTB) ORTBImpVideoCompanionType() (err error) {
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	companionType, err := o.values.GetIntArray(ORTBImpVideoCompanionType, ArraySeparator)
	if err == nil {
		o.ortb.Imp[0].Video.CompanionType = v26.GetCompanionType(companionType)
	}
	return
}

/*********************** Regs ***********************/

// ORTBRegsCoppa will read and set ortb Regs.Coppa parameter
func (o *OpenRTB) ORTBRegsCoppa() (err error) {
	val, ok, err := o.values.GetInt(ORTBRegsCoppa)
	if !ok || err != nil {
		return
	}
	if o.ortb.Regs == nil {
		o.ortb.Regs = &openrtb2.Regs{}
	}
	o.ortb.Regs.COPPA = int8(val)
	return
}

/*********************** Imp ***********************/

// ORTBImpID will read and set ortb Imp.ID parameter
func (o *OpenRTB) ORTBImpID() (err error) {
	val, ok := o.values.GetString(ORTBImpID)
	if !ok {
		o.ortb.Imp[0].ID = uuid.NewV4().String()
	} else {
		o.ortb.Imp[0].ID = val
	}
	return
}

// ORTBImpDisplayManager will read and set ortb Imp.DisplayManager parameter
func (o *OpenRTB) ORTBImpDisplayManager() (err error) {
	val, ok := o.values.GetString(ORTBImpDisplayManager)
	if !ok {
		return
	}
	o.ortb.Imp[0].DisplayManager = val
	return
}

// ORTBImpDisplayManagerVer will read and set ortb Imp.DisplayManagerVer parameter
func (o *OpenRTB) ORTBImpDisplayManagerVer() (err error) {
	val, ok := o.values.GetString(ORTBImpDisplayManagerVer)
	if !ok {
		return
	}
	o.ortb.Imp[0].DisplayManagerVer = val
	return
}

// ORTBImpInstl will read and set ortb Imp.Instl parameter
func (o *OpenRTB) ORTBImpInstl() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpInstl)
	if !ok || err != nil {
		return
	}
	o.ortb.Imp[0].Instl = int8(val)
	return
}

// ORTBImpTagID will read and set ortb Imp.TagId parameter
func (o *OpenRTB) ORTBImpTagID() (err error) {
	val, ok := o.values.GetString(ORTBImpTagID)
	if !ok {
		return
	}
	o.ortb.Imp[0].TagID = val
	return
}

// ORTBImpBidFloor will read and set ortb Imp.BidFloor parameter
func (o *OpenRTB) ORTBImpBidFloor() (err error) {
	bidFloor, ok, err := o.values.GetFloat64(ORTBImpBidFloor)
	if !ok || err != nil {
		return
	}
	if bidFloor > 0 {
		o.ortb.Imp[0].BidFloor = bidFloor
	}
	return err
}

// ORTBImpBidFloorCur will read and set ortb Imp.BidFloorCur parameter
func (o *OpenRTB) ORTBImpBidFloorCur() (err error) {
	bidFloor, ok, err := o.values.GetFloat64(ORTBImpBidFloor)
	if !ok || err != nil {
		return
	}
	if bidFloor > 0 {
		bidFloorCur, ok := o.values.GetString(ORTBImpBidFloorCur)
		if ok {
			o.ortb.Imp[0].BidFloorCur = bidFloorCur
		} else {
			o.ortb.Imp[0].BidFloorCur = USD
		}
	}
	return
}

// ORTBImpClickBrowser will read and set ortb Imp.ClickBrowser parameter
func (o *OpenRTB) ORTBImpClickBrowser() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpClickBrowser)
	if !ok || err != nil {
		return
	}
	o.ortb.Imp[0].ClickBrowser = ptrutil.ToPtr(int8(val))
	return
}

// ORTBImpSecure will read and set ortb Imp.Secure parameter
func (o *OpenRTB) ORTBImpSecure() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpSecure)
	if !ok || err != nil {
		return
	}
	val8 := int8(val)
	o.ortb.Imp[0].Secure = &val8
	return
}

// ORTBImpIframeBuster will read and set ortb Imp.IframeBuster parameter
func (o *OpenRTB) ORTBImpIframeBuster() (err error) {
	o.ortb.Imp[0].IframeBuster = o.values.GetStringArray(ORTBImpIframeBuster, ArraySeparator)
	return
}

// ORTBImpExp will read and set ortb Imp.Exp parameter
func (o *OpenRTB) ORTBImpExp() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpExp)
	if !ok || err != nil {
		return
	}
	o.ortb.Imp[0].Exp = int64(val)
	return
}

// ORTBImpPmp will read and set ortb Imp.Pmp parameter
func (o *OpenRTB) ORTBImpPmp() (err error) {
	pmp, ok := o.values.GetString(ORTBImpPmp)
	if !ok {
		return
	}
	ortbPmp := &openrtb2.PMP{}
	err = json.Unmarshal([]byte(pmp), ortbPmp)
	if err != nil {
		return
	}
	o.ortb.Imp[0].PMP = ortbPmp

	return
}

// ORTBImpExtBidder will read and set ortb Imp.Ext.Bidder parameter
func (o *OpenRTB) ORTBImpExtBidder() (err error) {
	str, ok := o.values.GetString(ORTBImpExtBidder)
	if !ok {
		return
	}

	impExt := map[string]interface{}{}
	if o.ortb.Imp[0].Ext != nil {
		err = json.Unmarshal(o.ortb.Imp[0].Ext, &impExt)
		if err != nil {
			return
		}
	}

	impExtBidder := map[string]interface{}{}
	err = json.Unmarshal([]byte(str), &impExtBidder)
	if err != nil {
		return
	}

	impExt[BIDDER_KEY] = impExtBidder
	data, err := json.Marshal(impExt)
	if err != nil {
		return
	}

	o.ortb.Imp[0].Ext = json.RawMessage(data)
	return
}

// ORTBImpExtPrebid will read and set ortb Imp.Ext.Prebid parameter
func (o *OpenRTB) ORTBImpExtPrebid() (err error) {
	str, ok := o.values.GetString(ORTBImpExtPrebid)
	if !ok {
		return
	}

	impExt := map[string]interface{}{}
	if o.ortb.Imp[0].Ext != nil {
		err = json.Unmarshal(o.ortb.Imp[0].Ext, &impExt)
		if err != nil {
			return
		}
	}

	impExtPrebid := map[string]interface{}{}
	err = json.Unmarshal([]byte(str), &impExtPrebid)
	if err != nil {
		return
	}

	impExt[PrebidKey] = impExtPrebid
	data, err := json.Marshal(impExt)
	if err != nil {
		return
	}

	o.ortb.Imp[0].Ext = data
	return
}

/********************** Device **********************/

// ORTBDeviceUserAgent will read and set ortb Device.UserAgent parameter
func (o *OpenRTB) ORTBDeviceUserAgent() (err error) {
	val, ok := o.values.GetString(ORTBDeviceUserAgent)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.UA = val
	return
}

// ORTBDeviceIP will read and set ortb Device.IP parameter
func (o *OpenRTB) ORTBDeviceIP() (err error) {
	val, ok := o.values.GetString(ORTBDeviceIP)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.IP = val
	return
}

// ORTBDeviceIpv6 will read and set ortb Device.Ipv6 parameter
func (o *OpenRTB) ORTBDeviceIpv6() (err error) {
	val, ok := o.values.GetString(ORTBDeviceIpv6)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.IPv6 = val
	return
}

// ORTBDeviceDnt will read and set ortb Device.Dnt parameter
func (o *OpenRTB) ORTBDeviceDnt() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceDnt)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	val8 := int8(val)
	o.ortb.Device.DNT = &val8
	return
}

// ORTBDeviceLmt will read and set ortb Device.Lmt parameter
func (o *OpenRTB) ORTBDeviceLmt() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceLmt)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	val8 := int8(val)
	o.ortb.Device.Lmt = &val8
	return
}

// ORTBDeviceDeviceType will read and set ortb Device.DeviceType parameter
func (o *OpenRTB) ORTBDeviceDeviceType() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceDeviceType)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	deviceType := adcom1.DeviceType(val)
	o.ortb.Device.DeviceType = deviceType

	return err
}

// ORTBDeviceMake will read and set ortb Device.Make parameter
func (o *OpenRTB) ORTBDeviceMake() (err error) {
	val, ok := o.values.GetString(ORTBDeviceMake)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.Make = val
	return
}

// ORTBDeviceModel will read and set ortb Device.Model parameter
func (o *OpenRTB) ORTBDeviceModel() (err error) {
	val, ok := o.values.GetString(ORTBDeviceModel)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.Model = val
	return
}

// ORTBDeviceOs will read and set ortb Device.Os parameter
func (o *OpenRTB) ORTBDeviceOs() (err error) {
	val, ok := o.values.GetString(ORTBDeviceOs)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.OS = val
	return
}

// ORTBDeviceOsv will read and set ortb Device.Osv parameter
func (o *OpenRTB) ORTBDeviceOsv() (err error) {
	val, ok := o.values.GetString(ORTBDeviceOsv)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.OSV = val
	return
}

// ORTBDeviceHwv will read and set ortb Device.Hwv parameter
func (o *OpenRTB) ORTBDeviceHwv() (err error) {
	val, ok := o.values.GetString(ORTBDeviceHwv)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.HWV = val
	return
}

// ORTBDeviceWidth will read and set ortb Device.Width parameter
func (o *OpenRTB) ORTBDeviceWidth() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceWidth)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.W = int64(val)
	return
}

// ORTBDeviceHeight will read and set ortb Device.Height parameter
func (o *OpenRTB) ORTBDeviceHeight() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceHeight)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.H = int64(val)
	return
}

// ORTBDevicePpi will read and set ortb Device.Ppi parameter
func (o *OpenRTB) ORTBDevicePpi() (err error) {
	val, ok, err := o.values.GetInt(ORTBDevicePpi)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.PPI = int64(val)
	return
}

// ORTBDevicePxRatio will read and set ortb Device.PxRatio parameter
func (o *OpenRTB) ORTBDevicePxRatio() (err error) {
	val, ok, err := o.values.GetFloat64(ORTBDevicePxRatio)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.PxRatio = val
	return
}

// ORTBDeviceJS will read and set ortb Device.JS parameter
func (o *OpenRTB) ORTBDeviceJS() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceJS)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.JS = ptrutil.ToPtr(int8(val))
	return
}

// ORTBDeviceGeoFetch will read and set ortb Device.Geo.Fetch parameter
func (o *OpenRTB) ORTBDeviceGeoFetch() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceGeoFetch)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.GeoFetch = ptrutil.ToPtr(int8(val))
	return
}

// ORTBDeviceFlashVer will read and set ortb Device.FlashVer parameter
func (o *OpenRTB) ORTBDeviceFlashVer() (err error) {
	val, ok := o.values.GetString(ORTBDeviceFlashVer)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.FlashVer = val
	return
}

// ORTBDeviceLanguage will read and set ortb Device.Language parameter
func (o *OpenRTB) ORTBDeviceLanguage() (err error) {
	val, ok := o.values.GetString(ORTBDeviceLanguage)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.Language = val
	return
}

// ORTBDeviceCarrier will read and set ortb Device.Carrier parameter
func (o *OpenRTB) ORTBDeviceCarrier() (err error) {
	val, ok := o.values.GetString(ORTBDeviceCarrier)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.Carrier = val
	return
}

// ORTBDeviceMccmnc will read and set ortb Device.Mccmnc parameter
func (o *OpenRTB) ORTBDeviceMccmnc() (err error) {
	val, ok := o.values.GetString(ORTBDeviceMccmnc)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.MCCMNC = val
	return
}

// ORTBDeviceConnectionType will read and set ortb Device.ConnectionType parameter
func (o *OpenRTB) ORTBDeviceConnectionType() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceConnectionType)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	connectionType := adcom1.ConnectionType(val)
	o.ortb.Device.ConnectionType = &connectionType

	return err
}

// ORTBDeviceIfa will read and set ortb Device.Ifa parameter
func (o *OpenRTB) ORTBDeviceIfa() (err error) {
	val, ok := o.values.GetString(ORTBDeviceIfa)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.IFA = val
	return
}

// ORTBDeviceDidSha1 will read and set ortb Device.DidSha1 parameter
func (o *OpenRTB) ORTBDeviceDidSha1() (err error) {
	val, ok := o.values.GetString(ORTBDeviceDidSha1)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.DIDSHA1 = val
	return
}

// ORTBDeviceDidMd5 will read and set ortb Device.DidMd5 parameter
func (o *OpenRTB) ORTBDeviceDidMd5() (err error) {
	val, ok := o.values.GetString(ORTBDeviceDidMd5)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.DIDMD5 = val
	return
}

// ORTBDeviceDpidSha1 will read and set ortb Device.DpidSha1 parameter
func (o *OpenRTB) ORTBDeviceDpidSha1() (err error) {
	val, ok := o.values.GetString(ORTBDeviceDpidSha1)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.DPIDSHA1 = val
	return
}

// ORTBDeviceDpidMd5 will read and set ortb Device.DpidMd5 parameter
func (o *OpenRTB) ORTBDeviceDpidMd5() (err error) {
	val, ok := o.values.GetString(ORTBDeviceDpidMd5)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.DPIDMD5 = val
	return
}

// ORTBDeviceMacSha1 will read and set ortb Device.MacSha1 parameter
func (o *OpenRTB) ORTBDeviceMacSha1() (err error) {
	val, ok := o.values.GetString(ORTBDeviceMacSha1)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.MACSHA1 = val
	return
}

// ORTBDeviceMacMd5 will read and set ortb Device.MacMd5 parameter
func (o *OpenRTB) ORTBDeviceMacMd5() (err error) {
	val, ok := o.values.GetString(ORTBDeviceMacMd5)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	o.ortb.Device.MACMD5 = val
	return
}

/*********************** Device.Geo ***********************/

// ORTBDeviceGeoLat will read and set ortb Device.Geo.Lat parameter
func (o *OpenRTB) ORTBDeviceGeoLat() (err error) {
	val, ok, err := o.values.GetFloat64(ORTBDeviceGeoLat)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	o.ortb.Device.Geo.Lat = ptrutil.ToPtr(val)
	return
}

// ORTBDeviceGeoLon will read and set ortb Device.Geo.Lon parameter
func (o *OpenRTB) ORTBDeviceGeoLon() (err error) {
	val, ok, err := o.values.GetFloat64(ORTBDeviceGeoLon)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	o.ortb.Device.Geo.Lon = ptrutil.ToPtr(val)
	return
}

// ORTBDeviceGeoType will read and set ortb Device.Geo.Type parameter
func (o *OpenRTB) ORTBDeviceGeoType() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceGeoType)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	geoType := adcom1.LocationType(val)
	o.ortb.Device.Geo.Type = geoType
	return
}

// ORTBDeviceGeoAccuracy will read and set ortb Device.Geo.Accuracy parameter
func (o *OpenRTB) ORTBDeviceGeoAccuracy() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceGeoAccuracy)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	o.ortb.Device.Geo.Accuracy = int64(val)
	return
}

// ORTBDeviceGeoLastFix will read and set ortb Device.Geo.LastFix parameter
func (o *OpenRTB) ORTBDeviceGeoLastFix() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceGeoLastFix)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	o.ortb.Device.Geo.LastFix = int64(val)
	return
}

// ORTBDeviceGeoIPService will read and set ortb Device.Geo.IPService parameter
func (o *OpenRTB) ORTBDeviceGeoIPService() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceGeoIPService)
	if !ok || err != nil {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	IPService := adcom1.IPLocationService(val)
	o.ortb.Device.Geo.IPService = IPService
	return
}

// ORTBDeviceGeoCountry will read and set ortb Device.Geo.Country parameter
func (o *OpenRTB) ORTBDeviceGeoCountry() (err error) {
	val, ok := o.values.GetString(ORTBDeviceGeoCountry)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	o.ortb.Device.Geo.Country = val
	return
}

// ORTBDeviceGeoRegion will read and set ortb Device.Geo.Region parameter
func (o *OpenRTB) ORTBDeviceGeoRegion() (err error) {
	val, ok := o.values.GetString(ORTBDeviceGeoRegion)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	o.ortb.Device.Geo.Region = val
	return
}

// ORTBDeviceGeoRegionFips104 will read and set ortb Device.Geo.RegionFips104 parameter
func (o *OpenRTB) ORTBDeviceGeoRegionFips104() (err error) {
	val, ok := o.values.GetString(ORTBDeviceGeoRegionFips104)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	o.ortb.Device.Geo.RegionFIPS104 = val
	return
}

// ORTBDeviceGeoMetro will read and set ortb Device.Geo.Metro parameter
func (o *OpenRTB) ORTBDeviceGeoMetro() (err error) {
	val, ok := o.values.GetString(ORTBDeviceGeoMetro)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	o.ortb.Device.Geo.Metro = val
	return
}

// ORTBDeviceGeoCity will read and set ortb Device.Geo.City parameter
func (o *OpenRTB) ORTBDeviceGeoCity() (err error) {
	val, ok := o.values.GetString(ORTBDeviceGeoCity)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	o.ortb.Device.Geo.City = val
	return
}

// ORTBDeviceGeoZip will read and set ortb Device.Geo.Zip parameter
func (o *OpenRTB) ORTBDeviceGeoZip() (err error) {
	val, ok := o.values.GetString(ORTBDeviceGeoZip)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	o.ortb.Device.Geo.ZIP = val
	return
}

// ORTBDeviceGeoUtcOffset will read and set ortb Device.Geo.UtcOffset parameter
func (o *OpenRTB) ORTBDeviceGeoUtcOffset() (err error) {
	val, ok, err := o.values.GetInt(ORTBDeviceGeoUtcOffset)
	if !ok || err != nil {
		return
	}

	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}
	o.ortb.Device.Geo.UTCOffset = int64(val)
	return
}

/*********************** User ***********************/

// ORTBUserID will read and set ortb UserID parameter
func (o *OpenRTB) ORTBUserID() (err error) {
	val, ok := o.values.GetString(ORTBUserID)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	o.ortb.User.ID = val
	return
}

// ORTBUserBuyerUID will read and set ortb UserBuyerUID parameter
func (o *OpenRTB) ORTBUserBuyerUID() (err error) {
	val, ok := o.values.GetString(ORTBUserBuyerUID)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	o.ortb.User.BuyerUID = val
	return
}

// ORTBUserYob will read and set ortb UserYob parameter
func (o *OpenRTB) ORTBUserYob() (err error) {
	val, ok, err := o.values.GetInt(ORTBUserYob)
	if !ok || err != nil {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	o.ortb.User.Yob = int64(val)
	return
}

// ORTBUserGender will read and set ortb UserGender parameter
func (o *OpenRTB) ORTBUserGender() (err error) {
	val, ok := o.values.GetString(ORTBUserGender)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	o.ortb.User.Gender = val
	return
}

// ORTBUserKeywords will read and set ortb UserKeywords parameter
func (o *OpenRTB) ORTBUserKeywords() (err error) {
	val, ok := o.values.GetString(ORTBUserKeywords)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	o.ortb.User.Keywords = val
	return
}

// ORTBUserCustomData will read and set ortb UserCustomData parameter
func (o *OpenRTB) ORTBUserCustomData() (err error) {
	val, ok := o.values.GetString(ORTBUserCustomData)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	o.ortb.User.CustomData = val
	return
}

/*********************** User.Geo ***********************/

// ORTBUserGeoLat will read and set ortb UserGeo.Lat parameter
func (o *OpenRTB) ORTBUserGeoLat() (err error) {
	val, ok, err := o.values.GetFloat64(ORTBUserGeoLat)
	if !ok || err != nil {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	o.ortb.User.Geo.Lat = ptrutil.ToPtr(val)
	return
}

// ORTBUserGeoLon will read and set ortb UserGeo.Lon parameter
func (o *OpenRTB) ORTBUserGeoLon() (err error) {
	val, ok, err := o.values.GetFloat64(ORTBUserGeoLon)
	if !ok || err != nil {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	o.ortb.User.Geo.Lon = ptrutil.ToPtr(val)
	return
}

// ORTBUserGeoType will read and set ortb UserGeo.Type parameter
func (o *OpenRTB) ORTBUserGeoType() (err error) {
	val, ok, err := o.values.GetInt(ORTBUserGeoType)
	if !ok || err != nil {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	geoType := adcom1.LocationType(val)
	o.ortb.User.Geo.Type = geoType
	return
}

// ORTBUserGeoAccuracy will read and set ortb UserGeo.Accuracy parameter
func (o *OpenRTB) ORTBUserGeoAccuracy() (err error) {
	val, ok, err := o.values.GetInt(ORTBUserGeoAccuracy)
	if !ok || err != nil {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	o.ortb.User.Geo.Accuracy = int64(val)
	return
}

// ORTBUserGeoLastFix will read and set ortb UserGeo.LastFix parameter
func (o *OpenRTB) ORTBUserGeoLastFix() (err error) {
	val, ok, err := o.values.GetInt(ORTBUserGeoLastFix)
	if !ok || err != nil {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	o.ortb.User.Geo.LastFix = int64(val)
	return
}

// ORTBUserGeoIPService will read and set ortb UserGeo.IPService parameter
func (o *OpenRTB) ORTBUserGeoIPService() (err error) {
	val, ok, err := o.values.GetInt(ORTBUserGeoIPService)
	if !ok || err != nil {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	IPService := adcom1.IPLocationService(val)
	o.ortb.User.Geo.IPService = IPService
	return
}

// ORTBUserGeoCountry will read and set ortb UserGeo.Country parameter
func (o *OpenRTB) ORTBUserGeoCountry() (err error) {
	val, ok := o.values.GetString(ORTBUserGeoCountry)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	o.ortb.User.Geo.Country = val
	return
}

// ORTBUserGeoRegion will read and set ortb UserGeo.Region parameter
func (o *OpenRTB) ORTBUserGeoRegion() (err error) {
	val, ok := o.values.GetString(ORTBUserGeoRegion)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	o.ortb.User.Geo.Region = val
	return
}

// ORTBUserGeoRegionFips104 will read and set ortb UserGeo.RegionFips104 parameter
func (o *OpenRTB) ORTBUserGeoRegionFips104() (err error) {
	val, ok := o.values.GetString(ORTBUserGeoRegionFips104)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	o.ortb.User.Geo.RegionFIPS104 = val
	return
}

// ORTBUserGeoMetro will read and set ortb UserGeo.Metro parameter
func (o *OpenRTB) ORTBUserGeoMetro() (err error) {
	val, ok := o.values.GetString(ORTBUserGeoMetro)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	o.ortb.User.Geo.Metro = val
	return
}

// ORTBUserGeoCity will read and set ortb UserGeo.City parameter
func (o *OpenRTB) ORTBUserGeoCity() (err error) {
	val, ok := o.values.GetString(ORTBUserGeoCity)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	o.ortb.User.Geo.City = val
	return
}

// ORTBUserGeoZip will read and set ortb UserGeo.Zip parameter
func (o *OpenRTB) ORTBUserGeoZip() (err error) {
	val, ok := o.values.GetString(ORTBUserGeoZip)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	o.ortb.User.Geo.ZIP = val
	return
}

// ORTBUserGeoUtcOffset will read and set ortb UserGeo.UtcOffset parameter
func (o *OpenRTB) ORTBUserGeoUtcOffset() (err error) {
	val, ok, err := o.values.GetInt(ORTBUserGeoUtcOffset)
	if !ok || err != nil {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}
	o.ortb.User.Geo.UTCOffset = int64(val)
	return
}

/*********************** Request.Ext.Parameters ***********************/

// ORTBProfileID will read and set ortb ProfileId parameter
func (o *OpenRTB) ORTBProfileID() (err error) {
	val, ok, err := o.values.GetInt(ORTBProfileID)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidProfileID, "invalid wrapper profile id")
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtProfileId] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBVersionID will read and set ortb VersionId parameter
func (o *OpenRTB) ORTBVersionID() (err error) {
	val, ok, err := o.values.GetInt(ORTBVersionID)
	if !ok || err != nil {
		return
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtVersionId] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data

	return
}

// ORTBSSAuctionFlag will read and set ortb SSAuctionFlag parameter
func (o *OpenRTB) ORTBSSAuctionFlag() (err error) {
	val, ok, err := o.values.GetInt(ORTBSSAuctionFlag)
	if !ok || err != nil {
		return
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtSSAuctionFlag] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBSumryDisableFlag will read and set ortb SumryDisableFlag parameter
func (o *OpenRTB) ORTBSumryDisableFlag() (err error) {
	val, ok, err := o.values.GetInt(ORTBSumryDisableFlag)
	if !ok || err != nil {
		return
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtSumryDisableFlag] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBClientConfigFlag will read and set ortb ClientConfigFlag parameter
func (o *OpenRTB) ORTBClientConfigFlag() (err error) {
	val, ok, err := o.values.GetInt(ORTBClientConfigFlag)
	if !ok || err != nil {
		return
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtClientConfigFlag] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBSupportDeals will read and set ortb ClientConfigFlag parameter
func (o *OpenRTB) ORTBSupportDeals() (err error) {
	val, ok, err := o.values.GetBoolean(ORTBSupportDeals)
	if !ok || err != nil {
		return
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtSupportDeals] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBIncludeBrandCategory will read and set ortb ORTBIncludeBrandCategory parameter
func (o *OpenRTB) ORTBIncludeBrandCategory() (err error) {
	val, ok, err := o.values.GetInt(ORTBIncludeBrandCategory)
	if !ok || err != nil {
		return
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtIncludeBrandCategory] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return err
}

// ORTBSSAI will read and set ortb ssai parameter
func (o *OpenRTB) ORTBSSAI() (err error) {
	val, ok := o.values.GetString(ORTBSSAI)
	if !ok {
		return
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtSsai] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBKeyValues read and set keyval parameter
func (o *OpenRTB) ORTBKeyValues() (err error) {
	val, err := o.values.GetQueryParams(ORTBKeyValues)
	if val == nil {
		return
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtKV] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}
	o.ortb.Ext = data

	return nil
}

// ORTBKeyValuesMap read and set keyval parameter
func (o *OpenRTB) ORTBKeyValuesMap() (err error) {
	val, err := o.values.GetJSON(ORTBKeyValuesMap)
	if val == nil {
		return
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	wrapperExt, ok := reqExt[ORTBExtWrapper].(map[string]interface{})
	if !ok {
		wrapperExt = map[string]interface{}{}
	}
	wrapperExt[ORTBExtKV] = val

	reqExt[ORTBExtWrapper] = wrapperExt
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}
	o.ortb.Ext = data

	return nil
}

/*********************** User.Ext.Consent ***********************/

// ORTBUserExtConsent will read and set ortb User.Ext.Consent parameter
func (o *OpenRTB) ORTBUserExtConsent() (err error) {
	val, ok := o.values.GetString(ORTBUserExtConsent)
	if !ok {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	userExt := map[string]interface{}{}
	if o.ortb.User.Ext != nil {
		err = json.Unmarshal(o.ortb.User.Ext, &userExt)
		if err != nil {
			return
		}
	}
	userExt[ORTBExtConsent] = val

	data, err := json.Marshal(userExt)
	if err != nil {
		return
	}

	o.ortb.User.Ext = data
	return
}

/*********************** Regs.Ext.Gdpr ***********************/

// ORTBRegsExtGdpr will read and set ortb Regs.Ext.Gdpr parameter
func (o *OpenRTB) ORTBRegsExtGdpr() (err error) {
	val, ok, err := o.values.GetInt(ORTBRegsExtGdpr)
	if !ok || err != nil {
		return
	}
	if o.ortb.Regs == nil {
		o.ortb.Regs = &openrtb2.Regs{}
	}
	val8 := int8(val)
	o.ortb.Regs.GDPR = &val8

	regsExt := map[string]interface{}{}
	if o.ortb.Regs.Ext != nil {
		err = json.Unmarshal(o.ortb.Regs.Ext, &regsExt)
		if err != nil {
			return
		}
	}
	regsExt[ORTBExtGDPR] = val

	data, err := json.Marshal(regsExt)
	if err != nil {
		return
	}
	o.ortb.Regs.Ext = data
	return
}

// ORTBRegsExtUSPrivacy will read and set ortb Regs.Ext.USPrivacy parameter
func (o *OpenRTB) ORTBRegsExtUSPrivacy() (err error) {
	val, ok := o.values.GetString(ORTBRegsExtUSPrivacy)
	if !ok {
		return
	}
	if o.ortb.Regs == nil {
		o.ortb.Regs = &openrtb2.Regs{}
	}
	o.ortb.Regs.USPrivacy = val

	regsExt := map[string]interface{}{}
	if o.ortb.Regs.Ext != nil {
		err = json.Unmarshal(o.ortb.Regs.Ext, &regsExt)
		if err != nil {
			return
		}
	}
	regsExt[ORTBExtUSPrivacy] = val

	data, err := json.Marshal(regsExt)
	if err != nil {
		return
	}

	o.ortb.Regs.Ext = data
	return
}

/*********************** Imp.Video.Ext ***********************/

// ORTBImpVideoExtOffset will read and set ortb Imp.Vid.Ext.Offset parameter
func (o *OpenRTB) ORTBImpVideoExtOffset() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoExtOffset)
	if !ok || err != nil {
		return
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	videoExt := map[string]interface{}{}
	if o.ortb.Imp[0].Video.Ext != nil {
		err = json.Unmarshal(o.ortb.Imp[0].Video.Ext, &videoExt)
		if err != nil {
			return
		}
	}

	videoExt[ORTBExtAdPodOffset] = val
	data, err := json.Marshal(videoExt)
	if err != nil {
		return
	}

	o.ortb.Imp[0].Video.Ext = data
	return
}

// ORTBImpVideoExtAdPodMinAds will read and set ortb Imp.Vid.Ext.AdPod.MinAds parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodMinAds() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoExtAdPodMinAds)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: imp.ext.adpod.minads value is invalid")
	}

	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	videoExt := map[string]interface{}{}
	if o.ortb.Imp[0].Video.Ext != nil {
		err = json.Unmarshal(o.ortb.Imp[0].Video.Ext, &videoExt)
		if err != nil {
			return
		}
	}

	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMinAds] = val

	videoExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(videoExt)
	if err != nil {
		return
	}

	o.ortb.Imp[0].Video.Ext = data
	return
}

// ORTBImpVideoExtAdPodMaxAds will read and set ortb Imp.Vid.Ext.AdPod.MaxAds parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodMaxAds() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoExtAdPodMaxAds)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: imp.ext.adpod.maxads value is invalid")
	}

	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	videoExt := map[string]interface{}{}
	if o.ortb.Imp[0].Video.Ext != nil {
		err = json.Unmarshal(o.ortb.Imp[0].Video.Ext, &videoExt)
		if err != nil {
			return
		}
	}

	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMaxAds] = val

	videoExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(videoExt)
	if err != nil {
		return
	}

	o.ortb.Imp[0].Video.Ext = data
	return
}

// ORTBImpVideoExtAdPodMinDuration will read and set ortb Imp.Vid.Ext.AdPod.MinDuration parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodMinDuration() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoExtAdPodMinDuration)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: imp.ext.adpod.minduration value is invalid")
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	videoExt := map[string]interface{}{}
	if o.ortb.Imp[0].Video.Ext != nil {
		err = json.Unmarshal(o.ortb.Imp[0].Video.Ext, &videoExt)
		if err != nil {
			return
		}
	}

	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMinDuration] = val

	videoExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(videoExt)
	if err != nil {
		return
	}

	o.ortb.Imp[0].Video.Ext = data
	return
}

// ORTBImpVideoExtAdPodMaxDuration will read and set ortb Imp.Vid.Ext.AdPod.MaxDuration parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodMaxDuration() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoExtAdPodMaxDuration)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: imp.ext.adpod.maxduration value is invalid")
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	videoExt := map[string]interface{}{}
	if o.ortb.Imp[0].Video.Ext != nil {
		err = json.Unmarshal(o.ortb.Imp[0].Video.Ext, &videoExt)
		if err != nil {
			return
		}
	}

	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMaxDuration] = val

	videoExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(videoExt)
	if err != nil {
		return
	}

	o.ortb.Imp[0].Video.Ext = data
	return
}

// ORTBImpVideoExtAdPodAdvertiserExclusionPercent will read and set ortb Imp.Vid.Ext.AdPod.AdvertiserExclusionPercent parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodAdvertiserExclusionPercent() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoExtAdPodAdvertiserExclusionPercent)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: imp.ext.adpod.advertiserexclusionpercent value is invalid")
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	videoExt := map[string]interface{}{}
	if o.ortb.Imp[0].Video.Ext != nil {
		err = json.Unmarshal(o.ortb.Imp[0].Video.Ext, &videoExt)
		if err != nil {
			return
		}
	}

	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodAdvertiserExclusionPercent] = val

	videoExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(videoExt)
	if err != nil {
		return
	}

	o.ortb.Imp[0].Video.Ext = data
	return
}

// ORTBImpVideoExtAdPodIABCategoryExclusionPercent will read and set ortb Imp.Vid.Ext.AdPod.IABCategoryExclusionPercent parameter
func (o *OpenRTB) ORTBImpVideoExtAdPodIABCategoryExclusionPercent() (err error) {
	val, ok, err := o.values.GetInt(ORTBImpVideoExtAdPodIABCategoryExclusionPercent)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: imp.ext.adpod.iabcategoryexclusionpercent value is invalid")
	}
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	videoExt := map[string]interface{}{}
	if o.ortb.Imp[0].Video.Ext != nil {
		err = json.Unmarshal(o.ortb.Imp[0].Video.Ext, &videoExt)
		if err != nil {
			return
		}
	}

	adpod, ok := videoExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodIABCategoryExclusionPercent] = val

	videoExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(videoExt)
	if err != nil {
		return
	}

	o.ortb.Imp[0].Video.Ext = data
	return
}

/*********************** Req.Ext ***********************/

// ORTBRequestExtAdPodMinAds will read and set ortb Request.Ext.AdPod.MinAds parameter
func (o *OpenRTB) ORTBRequestExtAdPodMinAds() (err error) {
	val, ok, err := o.values.GetInt(ORTBRequestExtAdPodMinAds)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: req.ext.adpod.minads value is invalid")
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMinAds] = val

	reqExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBRequestExtAdPodMaxAds will read and set ortb Request.Ext.AdPod.MaxAds parameter
func (o *OpenRTB) ORTBRequestExtAdPodMaxAds() (err error) {
	val, ok, err := o.values.GetInt(ORTBRequestExtAdPodMaxAds)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: req.ext.adpod.maxads value is invalid")
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMaxAds] = val

	reqExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBRequestExtAdPodMinDuration will read and set ortb Request.Ext.AdPod.MinDuration parameter
func (o *OpenRTB) ORTBRequestExtAdPodMinDuration() (err error) {
	val, ok, err := o.values.GetInt(ORTBRequestExtAdPodMinDuration)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: req.ext.adpod.minduration value is invalid")
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMinDuration] = val

	reqExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBRequestExtAdPodMaxDuration will read and set ortb Request.Ext.AdPod.MaxDuration parameter
func (o *OpenRTB) ORTBRequestExtAdPodMaxDuration() (err error) {
	val, ok, err := o.values.GetInt(ORTBRequestExtAdPodMaxDuration)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: req.ext.adpod.maxduration value is invalid")
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodMaxDuration] = val

	reqExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBRequestExtAdPodAdvertiserExclusionPercent will read and set ortb Request.Ext.AdPod.AdvertiserExclusionPercent parameter
func (o *OpenRTB) ORTBRequestExtAdPodAdvertiserExclusionPercent() (err error) {
	val, ok, err := o.values.GetInt(ORTBRequestExtAdPodAdvertiserExclusionPercent)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: req.ext.adpod.advertiserexclusionpercent value is invalid")
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodAdvertiserExclusionPercent] = val

	reqExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBRequestExtAdPodIABCategoryExclusionPercent will read and set ortb Request.Ext.AdPod.IABCategoryExclusionPercent parameter
func (o *OpenRTB) ORTBRequestExtAdPodIABCategoryExclusionPercent() (err error) {
	val, ok, err := o.values.GetInt(ORTBRequestExtAdPodIABCategoryExclusionPercent)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: req.ext.adpod.iabcategoryexclusionpercent value is invalid")
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodIABCategoryExclusionPercent] = val

	reqExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent will read and set ortb Request.Ext.AdPod.CrossPodAdvertiserExclusionPercent parameter
func (o *OpenRTB) ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent() (err error) {
	val, ok, err := o.values.GetInt(ORTBRequestExtAdPodCrossPodAdvertiserExclusionPercent)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: req.ext.adpod.crosspodadvertiserexclusionpercent value is invalid")
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodCrossPodAdvertiserExclusionPercent] = val

	reqExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent will read and set ortb Request.Ext.AdPod.CrossPodIABCategoryExclusionPercent parameter
func (o *OpenRTB) ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent() (err error) {
	val, ok, err := o.values.GetInt(ORTBRequestExtAdPodCrossPodIABCategoryExclusionPercent)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: req.ext.adpod.crosspodiabcategoryexclusionpercent value is invalid")
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodCrossPodIABCategoryExclusionPercent] = val

	reqExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBRequestExtAdPodIABCategoryExclusionWindow will read and set ortb Request.Ext.AdPod.IABCategoryExclusionWindow parameter
func (o *OpenRTB) ORTBRequestExtAdPodIABCategoryExclusionWindow() (err error) {
	val, ok, err := o.values.GetInt(ORTBRequestExtAdPodIABCategoryExclusionWindow)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: req.ext.adpod.iabcategoryexclusionwindow value is invalid")
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodIABCategoryExclusionWindow] = val

	reqExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBRequestExtAdPodAdvertiserExclusionWindow will read and set ortb Request.Ext.AdPod.AdvertiserExclusionWindow parameter
func (o *OpenRTB) ORTBRequestExtAdPodAdvertiserExclusionWindow() (err error) {
	val, ok, err := o.values.GetInt(ORTBRequestExtAdPodAdvertiserExclusionWindow)
	if !ok {
		return
	}
	if err != nil {
		return NewParseError(nbr.InvalidAdpodConfig, "Invalid adpod configuration: req.ext.adpod.advertiserexclusionwindow value is invalid")
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
	}

	adpod, ok := reqExt[ORTBExtAdPod].(map[string]interface{})
	if !ok {
		adpod = map[string]interface{}{}
	}
	adpod[ORTBExtAdPodAdvertiserExclusionWindow] = val

	reqExt[ORTBExtAdPod] = adpod
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

/*********************** Ext ***********************/

// ORTBBidRequestExt will read and set ortb BidRequest.Ext parameter
func (o *OpenRTB) ORTBBidRequestExt(key string, value *string) (err error) {
	ext := JSONNode{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBSourceExt will read and set ortb Source.Ext parameter
func (o *OpenRTB) ORTBSourceExt(key string, value *string) (err error) {
	if o.ortb.Source == nil {
		o.ortb.Source = &openrtb2.Source{}
	}
	ext := JSONNode{}
	if o.ortb.Source.Ext != nil {
		err = json.Unmarshal(o.ortb.Source.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Source.Ext = data
	return
}

// ORTBRegsExt will read and set ortb Regs.Ext parameter
func (o *OpenRTB) ORTBRegsExt(key string, value *string) (err error) {
	if o.ortb.Regs == nil {
		o.ortb.Regs = &openrtb2.Regs{}
	}
	ext := JSONNode{}
	if o.ortb.Regs.Ext != nil {
		err = json.Unmarshal(o.ortb.Regs.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Regs.Ext = data
	return
}

// ORTBImpExt will read and set ortb Imp.Ext parameter
func (o *OpenRTB) ORTBImpExt(key string, value *string) (err error) {
	ext := JSONNode{}
	if o.ortb.Imp[0].Ext != nil {
		err = json.Unmarshal(o.ortb.Imp[0].Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Imp[0].Ext = data
	return
}

// ORTBImpVideoExt will read and set ortb Imp.Video.Ext parameter
func (o *OpenRTB) ORTBImpVideoExt(key string, value *string) (err error) {
	if o.ortb.Imp[0].Video == nil {
		o.ortb.Imp[0].Video = &openrtb2.Video{}
	}
	ext := JSONNode{}
	if o.ortb.Imp[0].Video.Ext != nil {
		err = json.Unmarshal(o.ortb.Imp[0].Video.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Imp[0].Video.Ext = data
	return
}

// ORTBSiteExt will read and set ortb Site.Ext parameter
func (o *OpenRTB) ORTBSiteExt(key string, value *string) (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	ext := JSONNode{}
	if o.ortb.Site.Ext != nil {
		err = json.Unmarshal(o.ortb.Site.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Site.Ext = data
	return
}

// ORTBSiteContentNetworkExt will read and set ortb Site.Content.Network.Ext parameter
func (o *OpenRTB) ORTBSiteContentNetworkExt(key string, value *string) (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Network == nil {
		o.ortb.Site.Content.Network = &openrtb2.Network{}
	}
	ext := JSONNode{}
	if o.ortb.Site.Content.Network.Ext != nil {
		err = json.Unmarshal(o.ortb.Site.Content.Network.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Site.Content.Network.Ext = data
	return
}

// ORTBSiteContentChannelExt will read and set ortb Site.Content.Channel.Ext parameter
func (o *OpenRTB) ORTBSiteContentChannelExt(key string, value *string) (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Channel == nil {
		o.ortb.Site.Content.Channel = &openrtb2.Channel{}
	}
	ext := JSONNode{}
	if o.ortb.Site.Content.Channel.Ext != nil {
		err = json.Unmarshal(o.ortb.Site.Content.Channel.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Site.Content.Channel.Ext = data
	return
}

// ORTBAppExt will read and set ortb App.Ext parameter
func (o *OpenRTB) ORTBAppExt(key string, value *string) (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}

	ext := JSONNode{}
	if o.ortb.App.Ext != nil {
		err = json.Unmarshal(o.ortb.App.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.App.Ext = data
	return
}

// ORTBAppContentNetworkExt will read and set ortb App.Content.Network.Ext parameter
func (o *OpenRTB) ORTBAppContentNetworkExt(key string, value *string) (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Network == nil {
		o.ortb.App.Content.Network = &openrtb2.Network{}
	}
	ext := JSONNode{}
	if o.ortb.App.Content.Network.Ext != nil {
		err = json.Unmarshal(o.ortb.App.Content.Network.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.App.Content.Network.Ext = data
	return
}

// ORTBAppContentChannelExt will read and set ortb App.Content.Channel.Ext parameter
func (o *OpenRTB) ORTBAppContentChannelExt(key string, value *string) (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Channel == nil {
		o.ortb.App.Content.Channel = &openrtb2.Channel{}
	}
	ext := JSONNode{}
	if o.ortb.App.Content.Channel.Ext != nil {
		err = json.Unmarshal(o.ortb.App.Content.Channel.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.App.Content.Channel.Ext = data
	return
}

// ORTBSitePublisherExt will read and set ortb Site.Publisher.Ext parameter
func (o *OpenRTB) ORTBSitePublisherExt(key string, value *string) (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Publisher == nil {
		o.ortb.Site.Publisher = &openrtb2.Publisher{}
	}

	ext := JSONNode{}
	if o.ortb.Site.Publisher.Ext != nil {
		err = json.Unmarshal(o.ortb.Site.Publisher.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Site.Publisher.Ext = data
	return
}

// ORTBSiteContentExt will read and set ortb Site.Content.Ext parameter
func (o *OpenRTB) ORTBSiteContentExt(key string, value *string) (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}

	ext := JSONNode{}
	if o.ortb.Site.Content.Ext != nil {
		err = json.Unmarshal(o.ortb.Site.Content.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Site.Content.Ext = data
	return
}

// ORTBSiteContentProducerExt will read and set ortb Site.Content.Producer.Ext parameter
func (o *OpenRTB) ORTBSiteContentProducerExt(key string, value *string) (err error) {
	if o.ortb.Site == nil {
		o.ortb.Site = &openrtb2.Site{}
	}
	if o.ortb.Site.Content == nil {
		o.ortb.Site.Content = &openrtb2.Content{}
	}
	if o.ortb.Site.Content.Producer == nil {
		o.ortb.Site.Content.Producer = &openrtb2.Producer{}
	}

	ext := JSONNode{}
	if o.ortb.Site.Content.Producer.Ext != nil {
		err = json.Unmarshal(o.ortb.Site.Content.Producer.Ext, &ext)
		if err != nil {
			return
		}
	}
	SetValue(ext, key, value)

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Site.Content.Producer.Ext = data
	return
}

// ORTBAppPublisherExt will read and set ortb App.Publisher.Ext parameter
func (o *OpenRTB) ORTBAppPublisherExt(key string, value *string) (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Publisher == nil {
		o.ortb.App.Publisher = &openrtb2.Publisher{}
	}
	pubExt := JSONNode{}
	if o.ortb.App.Publisher.Ext != nil {
		err = json.Unmarshal(o.ortb.App.Publisher.Ext, &pubExt)
		if err != nil {
			return
		}
	}
	SetValue(pubExt, key, value)

	data, err := json.Marshal(pubExt)
	if err != nil {
		return
	}

	o.ortb.App.Publisher.Ext = data
	return
}

// ORTBAppContentExt will read and set ortb App.Content.Ext parameter
func (o *OpenRTB) ORTBAppContentExt(key string, value *string) (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}

	cntExt := JSONNode{}
	if o.ortb.App.Content.Ext != nil {
		err = json.Unmarshal(o.ortb.App.Content.Ext, &cntExt)
		if err != nil {
			return
		}
	}
	SetValue(cntExt, key, value)

	data, err := json.Marshal(cntExt)
	if err != nil {
		return
	}

	o.ortb.App.Content.Ext = data
	return
}

// ORTBAppContentProducerExt will read and set ortb App.Content.Producer.Ext parameter
func (o *OpenRTB) ORTBAppContentProducerExt(key string, value *string) (err error) {
	if o.ortb.App == nil {
		o.ortb.App = &openrtb2.App{}
	}
	if o.ortb.App.Content == nil {
		o.ortb.App.Content = &openrtb2.Content{}
	}
	if o.ortb.App.Content.Producer == nil {
		o.ortb.App.Content.Producer = &openrtb2.Producer{}
	}

	pdcExt := JSONNode{}
	if o.ortb.App.Content.Producer.Ext != nil {
		err = json.Unmarshal(o.ortb.App.Content.Producer.Ext, &pdcExt)
		if err != nil {
			return
		}
	}
	SetValue(pdcExt, key, value)

	data, err := json.Marshal(pdcExt)
	if err != nil {
		return
	}

	o.ortb.App.Content.Producer.Ext = data
	return
}

// ORTBDeviceExt will read and set ortb Device.Ext parameter
func (o *OpenRTB) ORTBDeviceExt(key string, value *string) (err error) {
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}

	deviceExt := JSONNode{}
	if o.ortb.Device.Ext != nil {
		err = json.Unmarshal(o.ortb.Device.Ext, &deviceExt)
		if err != nil {
			return
		}
	}
	SetValue(deviceExt, key, value)

	data, err := json.Marshal(deviceExt)
	if err != nil {
		return
	}

	o.ortb.Device.Ext = data
	return
}

// ORTBDeviceGeoExt will read and set ortb Device.Geo.Ext parameter
func (o *OpenRTB) ORTBDeviceGeoExt(key string, value *string) (err error) {
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}
	if o.ortb.Device.Geo == nil {
		o.ortb.Device.Geo = &openrtb2.Geo{}
	}

	deviceGeoExt := JSONNode{}
	if o.ortb.Device.Geo.Ext != nil {
		err = json.Unmarshal(o.ortb.Device.Geo.Ext, &deviceGeoExt)
		if err != nil {
			return
		}
	}
	SetValue(deviceGeoExt, key, value)

	data, err := json.Marshal(deviceGeoExt)
	if err != nil {
		return
	}

	o.ortb.Device.Geo.Ext = data
	return
}

// ORTBUserExt will read and set ortb User.Ext parameter
func (o *OpenRTB) ORTBUserExt(key string, value *string) (err error) {
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	userExt := JSONNode{}
	if o.ortb.User.Ext != nil {
		err = json.Unmarshal(o.ortb.User.Ext, &userExt)
		if err != nil {
			return
		}
	}
	SetValue(userExt, key, value)

	data, err := json.Marshal(userExt)
	if err != nil {
		return
	}

	o.ortb.User.Ext = data
	return
}

// ORTBUserGeoExt will read and set ortb User.Geo.Ext parameter
func (o *OpenRTB) ORTBUserGeoExt(key string, value *string) (err error) {
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	if o.ortb.User.Geo == nil {
		o.ortb.User.Geo = &openrtb2.Geo{}
	}

	geoExt := JSONNode{}
	if o.ortb.User.Geo.Ext != nil {
		err = json.Unmarshal(o.ortb.User.Geo.Ext, &geoExt)
		if err != nil {
			return
		}
	}
	SetValue(geoExt, key, value)

	data, err := json.Marshal(geoExt)
	if err != nil {
		return
	}

	o.ortb.User.Geo.Ext = data
	return
}

// ORTBUserExtConsent will read and set ortb User.Ext.Consent parameter
func (o *OpenRTB) ORTBDeviceExtIfaType() (err error) {
	val, ok := o.values.GetString(ORTBDeviceExtIfaType)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}

	deviceExt := map[string]interface{}{}
	if o.ortb.Device.Ext != nil {
		err = json.Unmarshal(o.ortb.Device.Ext, &deviceExt)
		if err != nil {
			return
		}
	}
	deviceExt[ORTBExtIfaType] = val

	data, err := json.Marshal(deviceExt)
	if err != nil {
		return
	}

	o.ortb.Device.Ext = data
	return
}

// ORTBDeviceExtSessionID will read and set ortb device.Ext.SessionID parameter
func (o *OpenRTB) ORTBDeviceExtSessionID() (err error) {
	val, ok := o.values.GetString(ORTBDeviceExtSessionID)
	if !ok {
		return
	}
	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}

	deviceExt := map[string]interface{}{}
	if o.ortb.Device.Ext != nil {
		err = json.Unmarshal(o.ortb.Device.Ext, &deviceExt)
		if err != nil {
			return
		}
	}
	deviceExt[ORTBExtSessionID] = val

	data, err := json.Marshal(deviceExt)
	if err != nil {
		return
	}
	o.ortb.Device.Ext = data
	return
}

// ORTBDeviceExtATTS will read and set ortb device.ext.atts parameter
func (o *OpenRTB) ORTBDeviceExtATTS() (err error) {
	value, ok, err := o.values.GetFloat64(ORTBDeviceExtATTS)
	if !ok || err != nil {
		return
	}

	if o.ortb.Device == nil {
		o.ortb.Device = &openrtb2.Device{}
	}

	deviceExt := map[string]interface{}{}
	if o.ortb.Device.Ext != nil {
		err = json.Unmarshal(o.ortb.Device.Ext, &deviceExt)
		if err != nil {
			return
		}
	}
	deviceExt[ORTBExtATTS] = value

	data, err := json.Marshal(deviceExt)
	if err != nil {
		return
	}
	o.ortb.Device.Ext = data
	return
}

// ORTBRequestExtPrebidTransparencyContent will read and set ortb Request.Ext.Prebid.Transparency.Content parameter
func (o *OpenRTB) ORTBRequestExtPrebidTransparencyContent() (err error) {
	contentString, ok := o.values.GetString(ORTBRequestExtPrebidTransparencyContent)
	if !ok {
		return
	}

	content := map[string]interface{}{}
	err = json.Unmarshal([]byte(contentString), &content)
	if err != nil {
		return fmt.Errorf(ErrJSONUnmarshalFailed, ORTBRequestExtPrebidTransparencyContent, err.Error(), contentString)
	}

	if len(content) == 0 {
		return
	}

	ext := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &ext)
		if err != nil {
			return
		}
	}

	prebidExt, ok := ext[ORTBExtPrebid].(map[string]interface{})
	if !ok {
		prebidExt = map[string]interface{}{}
	}

	transparancy, ok := prebidExt[ORTBExtPrebidTransparency].(map[string]interface{})
	if !ok {
		transparancy = map[string]interface{}{}
	}
	transparancy[ORTBExtPrebidTransparencyContent] = content
	prebidExt[ORTBExtPrebidTransparency] = transparancy
	ext[ORTBExtPrebid] = prebidExt

	data, err := json.Marshal(ext)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBUserExtEIDS will read and set ortb user.ext.eids parameter
func (o *OpenRTB) ORTBUserExtEIDS() (err error) {
	eidsValue, ok := o.values.GetString(ORTBUserExtEIDS)
	if !ok {
		return
	}

	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}

	userExt := map[string]interface{}{}
	if o.ortb.User.Ext != nil {
		err = json.Unmarshal(o.ortb.User.Ext, &userExt)
		if err != nil {
			return
		}
	}

	eids := []openrtb2.EID{}
	err = json.Unmarshal([]byte(eidsValue), &eids)
	if err != nil {
		return fmt.Errorf(ErrJSONUnmarshalFailed, ORTBUserExtEIDS, "Failed to unmarshal user.ext.eids", eidsValue)
	}

	userExt[ORTBExtEIDS] = eids

	data, err := json.Marshal(userExt)
	if err != nil {
		return
	}

	o.ortb.User.Ext = data
	return
}

// ORTBUserExtSessionDuration will read and set ortb User.Ext.sessionduration parameter
func (o *OpenRTB) ORTBUserExtSessionDuration() (err error) {
	valStr, ok := o.values.GetString(ORTBUserExtSessionDuration)
	if !ok || valStr == "" {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	userExt := map[string]interface{}{}
	if o.ortb.User.Ext != nil {
		if err = json.Unmarshal(o.ortb.User.Ext, &userExt); err != nil {
			return
		}
	}

	val, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		logger.Warn("Invalid session duration value '%v': %v", valStr, err)
		return nil
	}
	userExt[ORTBExtSessionDuration] = int64(val)

	data, err := json.Marshal(userExt)
	if err != nil {
		return
	}

	o.ortb.User.Ext = data
	return
}

// ORTBUserExtImpDepth will read and set ortb User.Ext.impdepth parameter
func (o *OpenRTB) ORTBUserExtImpDepth() (err error) {
	valStr, ok := o.values.GetString(ORTBUserExtImpDepth)
	if !ok || valStr == "" {
		return
	}
	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}
	userExt := map[string]interface{}{}
	if o.ortb.User.Ext != nil {
		if err = json.Unmarshal(o.ortb.User.Ext, &userExt); err != nil {
			return
		}
	}

	val, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		logger.Warn("Invalid imp depth value '%v': %v", valStr, err)
		return nil
	}
	userExt[ORTBExtImpDepth] = int64(val)

	data, err := json.Marshal(userExt)
	if err != nil {
		return
	}

	o.ortb.User.Ext = data
	return
}

// ORTBUserData will read and set ortb user.data parameter
func (o *OpenRTB) ORTBUserData() (err error) {
	dataValue, ok := o.values.GetString(ORTBUserData)
	if !ok {
		return
	}

	if o.ortb.User == nil {
		o.ortb.User = &openrtb2.User{}
	}

	data := []openrtb2.Data{}
	err = json.Unmarshal([]byte(dataValue), &data)
	if err != nil {
		return
	}

	o.ortb.User.Data = data
	return
}

func (o *OpenRTB) ORTBExtPrebidFloorsEnforceFloorDeals() (err error) {
	enforcementString, ok := o.values.GetString(ORTBExtPrebidFloorsEnforcement)
	if !ok {
		return
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
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

	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data
	return
}

// ORTBExtPrebidReturnAllBidStatus sets returnallbidstatus
func (o *OpenRTB) ORTBExtPrebidReturnAllBidStatus() (err error) {
	returnAllbidStatus, ok := o.values.GetString(ORTBExtPrebidReturnAllBidStatus)
	if !ok {
		return
	}

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
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
	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data

	return nil
}

// ORTBExtPrebidBidderParamsPubmaticCDS sets cds in req.ext.prebid.bidderparams.pubmatic
func (o *OpenRTB) ORTBExtPrebidBidderParamsPubmaticCDS() (err error) {
	cdsData, ok := o.values.GetString(ORTBExtPrebidBidderParamsPubmaticCDS)
	if !ok {
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

	reqExt := map[string]interface{}{}
	if o.ortb.Ext != nil {
		err = json.Unmarshal(o.ortb.Ext, &reqExt)
		if err != nil {
			return
		}
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

	data, err := json.Marshal(reqExt)
	if err != nil {
		return
	}

	o.ortb.Ext = data

	return
}

/*********************** Regs.Gpp And Regs.GppSid***********************/

// ORTBRegsGpp will read and set ortb Regs.gpp parameter
func (o *OpenRTB) ORTBRegsGpp() (err error) {
	val, ok := o.values.GetString(ORTBRegsGpp)
	if !ok {
		return
	}
	if o.ortb.Regs == nil {
		o.ortb.Regs = &openrtb2.Regs{}
	}
	o.ortb.Regs.GPP = val
	return
}

// ORTBRegsGpp will read and set ortb Regs.gpp_sid parameter
func (o *OpenRTB) ORTBRegsGppSid() error {
	var err error
	if o.ortb.Regs == nil {
		o.ortb.Regs = &openrtb2.Regs{}
	}
	if o.ortb.Regs.GPPSID, err = o.values.GetInt8Array(ORTBRegsGppSid, ArraySeparator); err != nil {
		return err
	}
	return nil
}
