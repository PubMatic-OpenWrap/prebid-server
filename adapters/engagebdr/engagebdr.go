package engagebdr

import (
	"encoding/json"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"net/http"

	"fmt"

	"github.com/PubMatic-OpenWrap/openrtb"
	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/errortypes"
)

type EngageBDRAdapter struct {
	http *adapters.HTTPAdapter
	URI  string
}

func (adapter *EngageBDRAdapter) MakeRequests(request *openrtb.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {

	errors := make([]error, 0, len(request.Imp))

	if request.Imp == nil || len(request.Imp) == 0 {
		errors = append(errors, &errortypes.BadInput{
			Message: fmt.Sprintf("Invalid BidRequest. No valid imp."),
		})
		return nil, errors
	}

	// EngageBDR uses different sspid parameters for banner and video.
	sspidImps := make(map[string][]openrtb.Imp)
	for _, imp := range request.Imp {

		if imp.Audio != nil {
			errors = append(errors, &errortypes.BadInput{
				Message: fmt.Sprintf("Ignoring imp id=%s, invalid MediaType. EngageBDR only supports Banner, Video and Native.", imp.ID),
			})
			continue
		}

		var bidderExt adapters.ExtImpBidder
		if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
			errors = append(errors, &errortypes.BadInput{
				Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding extImpBidder, err: %s.", imp.ID, err),
			})
			continue
		}
		impExt := openrtb_ext.ExtImpEngageBDR{}
		err := json.Unmarshal(bidderExt.Bidder, &impExt)
		if err != nil {
			errors = append(errors, &errortypes.BadInput{
				Message: fmt.Sprintf("Ignoring imp id=%s, error while decoding impExt, err: %s.", imp.ID, err),
			})
			continue
		}
		if impExt.Sspid == "" {
			errors = append(errors, &errortypes.BadInput{
				Message: fmt.Sprintf("Ignoring imp id=%s, no sspid present.", imp.ID),
			})
			continue
		}
		sspidImps[impExt.Sspid] = append(sspidImps[impExt.Sspid], imp)
	}

	var adapterRequests []*adapters.RequestData

	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")

	for sspid, imps := range sspidImps {
		if len(imps) > 0 {
			// Make a copy as we don't want to change the original request
			reqCopy := *request
			reqCopy.Imp = imps
			reqJSON, err := json.Marshal(reqCopy)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}
			adapterReq := adapters.RequestData{
				Method:  "POST",
				Uri:     adapter.URI + "?zoneid=" + sspid,
				Body:    reqJSON,
				Headers: headers,
			}
			adapterRequests = append(adapterRequests, &adapterReq)
		}
	}

	return adapterRequests, errors
}

func (adapter *EngageBDRAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	var bidResp openrtb.BidResponse
	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{err}
	}

	bidResponse := adapters.NewBidderResponseWithBidsCapacity(5)

	for _, sb := range bidResp.SeatBid {
		for i := range sb.Bid {
			bidResponse.Bids = append(bidResponse.Bids, &adapters.TypedBid{
				Bid:     &sb.Bid[i],
				BidType: getMediaTypeForImp(sb.Bid[i].ImpID, internalRequest.Imp),
			})
		}
	}
	return bidResponse, nil
}

func getMediaTypeForImp(impId string, imps []openrtb.Imp) openrtb_ext.BidType {
	mediaType := openrtb_ext.BidTypeBanner
	for _, imp := range imps {
		if imp.ID == impId {
			if imp.Video != nil {
				mediaType = openrtb_ext.BidTypeVideo
			} else if imp.Native != nil {
				mediaType = openrtb_ext.BidTypeNative
			}
			return mediaType
		}
	}
	return mediaType
}

func NewEngageBDRBidder(client *http.Client, endpoint string) *EngageBDRAdapter {
	adapter := &adapters.HTTPAdapter{Client: client}
	return &EngageBDRAdapter{
		http: adapter,
		URI:  endpoint,
	}
}
