package vastbidder

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/adapters"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

// responseHandler to parse VAST Tag
type responseHandler struct {
	internalRequest *openrtb2.BidRequest
	externalRequest *adapters.RequestData
	response        *adapters.ResponseData
	impBidderExt    *openrtb_ext.ExtImpVASTBidder
	vastTag         *openrtb_ext.ExtImpVASTBidderTag
	parser          xmlParser
}

// newResponseHandler returns new object
func newResponseHandler(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData, parser xmlParser) *responseHandler {
	return &responseHandler{
		internalRequest: internalRequest,
		externalRequest: externalRequest,
		response:        response,
		parser:          parser,
	}
}

// Validate will return bids
func (handler *responseHandler) Validate() []error {
	if handler.response.StatusCode != http.StatusOK {
		return []error{errNon2xxResponseStatus}
	}

	if len(handler.internalRequest.Imp) < handler.externalRequest.Params.ImpIndex {
		return []error{errInvalidImpressionIndex}
	}

	impExt, err := readImpExt(handler.internalRequest.Imp[handler.externalRequest.Params.ImpIndex].Ext)
	if nil != err {
		return []error{err}
	}

	if len(impExt.Tags) < handler.externalRequest.Params.VASTTagIndex {
		return []error{errInvalidVASTIndex}
	}

	//Initialise Extensions
	handler.impBidderExt = impExt
	handler.vastTag = impExt.Tags[handler.externalRequest.Params.VASTTagIndex]

	handler.parser.SetVASTTag(handler.vastTag)
	if err := handler.parser.Parse(handler.response.Body); err != nil {
		openrtb_ext.XMLLogf(openrtb_ext.XMLLogFormat, handler.parser.Name(), "vastbidder", base64.StdEncoding.EncodeToString(handler.response.Body))
		return []error{err}
	}
	return nil
}

// MakeBids will return bids
func (handler *responseHandler) MakeBids() (*adapters.BidderResponse, []error) {
	// get price and currency details, assumption currency is always returned
	price, currency := handler.parser.GetPricingDetails()
	if price <= 0 {
		price, currency = handler.vastTag.Price, "USD"
		if price <= 0 {
			return nil, []error{errMissingBidPrice}
		}
	}

	// duration prebid expects int value
	dur, err := handler.parser.GetDuration()
	if nil != err {
		//get duration from input bidder vast tag
		dur = handler.vastTag.Duration
	}

	// creating openrtb formatted bid object
	bid := &openrtb2.Bid{
		ID:      generateRandomID(),
		ImpID:   handler.internalRequest.Imp[handler.externalRequest.Params.ImpIndex].ID,
		AdM:     string(handler.response.Body),
		Price:   price,
		CrID:    handler.parser.GetCreativeID(),
		ADomain: handler.parser.GetAdvertiser(),
	}

	// bid.ext settting vasttagid and bid type
	bidExt := openrtb_ext.ExtBid{
		Prebid: &openrtb_ext.ExtBidPrebid{
			Video: &openrtb_ext.ExtBidPrebidVideo{
				VASTTagID: handler.vastTag.TagID,
				Duration:  dur,
			},
			Type: openrtb_ext.BidTypeVideo,
		},
	}
	bid.Ext, _ = json.Marshal(bidExt)

	// bidderresponse generation
	bidResponses := &adapters.BidderResponse{
		Bids: []*adapters.TypedBid{
			{
				Bid:      bid,
				BidType:  bidExt.Prebid.Type,
				BidVideo: bidExt.Prebid.Video,
			},
		},
		Currency: currency,
	}

	return bidResponses, nil
}
