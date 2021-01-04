package tagbidder

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/PubMatic-OpenWrap/etree"
	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
)

var durationRegExp = regexp.MustCompile(`^([01]?\d|2[0-3]):([0-5]?\d):([0-5]?\d)(\.(\d{1,3}))?$`)

//IVASTTagResponseHandler to parse VAST Tag
type IVASTTagResponseHandler interface {
	ITagResponseHandler
	ParseExtension(version string, tag *etree.Element, bid *adapters.TypedBid) []error
}

//VASTTagResponseHandler to parse VAST Tag
type VASTTagResponseHandler struct {
	IVASTTagResponseHandler
}

//NewVASTTagResponseHandler returns new object
func NewVASTTagResponseHandler() *VASTTagResponseHandler {
	return &VASTTagResponseHandler{}
}

//Validate will return bids
func (handler *VASTTagResponseHandler) Validate(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) []error {
	if response.StatusCode != http.StatusOK {
		return []error{errors.New(`validation failed`)}
	}
	return nil
}

//MakeBids will return bids
func (handler *VASTTagResponseHandler) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if err := handler.Validate(internalRequest, externalRequest, response); len(err) > 0 {
		return nil, err[:]
	}

	bidResponses, err := handler.vastTagToBidderResponse(internalRequest, externalRequest, response)
	fmt.Printf("\n[V1] errors:[%v] bidresponse:[%v]", err, bidResponses)
	return bidResponses, err
}

//ParseExtension will parse VAST XML extension object
func (handler *VASTTagResponseHandler) ParseExtension(version string, ad *etree.Element, bid *adapters.TypedBid) []error {
	return nil
}

func (handler *VASTTagResponseHandler) vastTagToBidderResponse(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	var errs []error

	doc := etree.NewDocument()

	//Read Document
	if err := doc.ReadFromBytes(response.Body); err != nil {
		errs = append(errs, err)
		return nil, errs[:]
	}

	//Check VAST Tag
	vast := doc.Element.FindElement(`./VAST`)
	if vast == nil {
		errs = append(errs, errors.New("VAST Tag Not Found"))
		return nil, errs[:]
	}

	//Check VAST/Ad Tag
	adElement := getAdElement(vast)
	if nil == adElement {
		errs = append(errs, errors.New("VAST/Ad Tag Not Found"))
		return nil, errs[:]
	}

	typedBid := &adapters.TypedBid{
		Bid:      &openrtb.Bid{},
		BidType:  openrtb_ext.BidTypeVideo,
		BidVideo: &openrtb_ext.ExtBidPrebidVideo{},
	}

	creatives := adElement.FindElements("Creatives/Creative")
	if nil != creatives {
		for _, creative := range creatives {
			// get creative id
			typedBid.Bid.CrID = getCreativeID(creative)
			// get duration. Ignore errors
			dur, _ := getDuration(creative)
			typedBid.BidVideo.Duration = int(dur) // prebid expects int value
		}
	}

	bidResponse := &adapters.BidderResponse{
		Bids:     []*adapters.TypedBid{typedBid},
		Currency: `USD`, //TODO: Need to check how to get currency value
	}

	//GetVersion
	version := vast.SelectAttrValue(`version`, `2.0`)

	if err := handler.ParseExtension(version, adElement, typedBid); len(err) > 0 {
		errs = append(errs, err...)
		return nil, errs[:]
	}

	//if bid.price is not set in ParseExtension
	if typedBid.Bid.Price <= 0 {
		price, currency, ok := getPricingDetails(version, adElement)
		if !ok {
			errs = append(errs, errors.New("Bid Price Not Present"))
			return nil, errs[:]
		}
		typedBid.Bid.Price = price
		if len(currency) > 0 {
			bidResponse.Currency = currency
		}
	}

	//if bid.id is not set in ParseExtension
	if len(typedBid.Bid.ID) == 0 {
		typedBid.Bid.ID = getRandomID()
	}

	//if bid.impid is not set in ParseExtension
	if len(typedBid.Bid.ImpID) == 0 {
		typedBid.Bid.ImpID = internalRequest.Imp[externalRequest.ImpIndex].ID
	}

	//if bid.adm is not set in ParseExtension
	if len(typedBid.Bid.AdM) == 0 {
		typedBid.Bid.AdM = string(response.Body)
	}

	return bidResponse, nil
}

func getAdElement(vast *etree.Element) *etree.Element {
	if ad := vast.FindElement(`./Ad/Wrapper`); nil != ad {
		return ad
	}
	if ad := vast.FindElement(`./Ad/InLine`); nil != ad {
		return ad
	}
	return nil
}

func getPricingDetails(version string, ad *etree.Element) (float64, string, bool) {
	var currency string
	var node *etree.Element

	if `3.0` == version {
		node = ad.FindElement(`./Pricing`)
	} else if `2.0` == version {
		node = ad.FindElement(`./Extensions/Extension/Price`)
	}

	if nil == node {
		return 0.0, currency, false
	}

	priceValue, err := strconv.ParseFloat(node.Text(), 64)
	if nil != err {
		return 0.0, currency, false
	}

	currencyNode := node.SelectAttr(`currency`)
	if nil != currencyNode {
		currency = currencyNode.Value
	}

	return priceValue, currency, true
}

var getRandomID = func() string {
	return strconv.FormatInt(rand.Int63(), intBase)
}

// getDuration extracts the duration of the bid from input creative of Linear type.
// The lookup may vary from vast version provided in the input
// returns duration in seconds or error if failed to obtained the duration.
// If multple Linear tags are present, onlyfirst one will be used
//
// It will lookup for duration only in case of creative type is Linear.
// If creative type other than Linear then this function will return error
// For Linear Creative it will lookup for Duration attribute.Duration value will be in hh:mm:ss.mmm format as per VAST specifications
// If Duration attribute not present this will return error
//
// After extracing the duration it will convert it into seconds
//
// The ad server uses the <Duration> element to denote
// the intended playback duration for the video or audio component of the ad.
// Time value may be in the format HH:MM:SS.mmm where .mmm indicates milliseconds.
// Providing milliseconds is optional.
//
// Reference
// 1.https://iabtechlab.com/wp-content/uploads/2019/06/VAST_4.2_final_june26.pdf
// 2.https://iabtechlab.com/wp-content/uploads/2018/11/VAST4.1-final-Nov-8-2018.pdf
// 3.https://iabtechlab.com/wp-content/uploads/2016/05/VAST4.0_Updated_April_2016.pdf
// 4.https://iabtechlab.com/wp-content/uploads/2016/04/VASTv3_0.pdf
func getDuration(creative *etree.Element) (float64, error) {
	if nil == creative {
		return 0, errors.New("Invalid Creative")
	}
	node := creative.FindElement("./Linear/Duration")
	if nil == node {
		return 0, errors.New("Invalid Duration")
	}
	duration := node.Text()
	// check if milliseconds is provided
	match := durationRegExp.FindStringSubmatch(duration)
	if nil == match {
		return 0, errors.New("Invalid Duration")
	}
	repl := "${1}h${2}m${3}s"
	ms := match[5]
	if "" != ms {
		repl += "${5}ms"
	}
	duration = durationRegExp.ReplaceAllString(duration, repl)
	dur, err := time.ParseDuration(duration)
	if err != nil {
		return 0, err
	}
	return dur.Seconds(), err
}

//getCreativeID looks for ID inside input creative tag
func getCreativeID(creative *etree.Element) string {
	if nil == creative {
		return ""
	}
	return creative.SelectAttrValue("id", getRandomID())
}
