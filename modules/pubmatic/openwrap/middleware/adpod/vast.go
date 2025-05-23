package middleware

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/exchange"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/nbr"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/utils"
	"github.com/rs/vast"
)

const (
	//VAST Constants
	VASTDefaultVersion    = 2.0
	VASTMaxVersion        = 4.0
	VASTDefaultVersionStr = `2.0`
	VASTDefaultTag        = `<VAST version="` + VASTDefaultVersionStr + `"/>`
	VASTElement           = `VAST`
	VASTAdElement         = `Ad`
	VASTWrapperElement    = `Wrapper`
	VASTAdTagURIElement   = `VASTAdTagURI`
	VASTVersionAttribute  = `version`
	VASTSequenceAttribute = `sequence`
	HTTPPrefix            = `http`
)

var (
	VASTVersionsStr   = []string{"0", "1.0", "2.0", "3.0", "4.0"}
	EmptyVASTResponse = []byte(`<VAST version="2.0"/>`)
	//HeaderOpenWrapStatus Status of OW Request
	HeaderOpenWrapStatus = "X-Ow-Status"
	ERROR_CODE           = "ErrorCode"
	ERROR_STRING         = "Error"
	NBR                  = "nbr"
	ERROR                = "error"
	//ErrorFormat parsing error format
	ErrorFormat        = `{"` + ERROR_CODE + `":%v,"` + ERROR_STRING + `":"%s"}`
	NBRFormatWithError = `{"` + NBR + `":%v,"` + ERROR + `":%s}`
	NBRFormatQuote     = `{"` + NBR + `":%v,"` + ERROR + `":"%v"}`
	NBRFormat          = `{"` + NBR + `":%v}`
)

type vastResponse struct {
	debug              string
	WrapperLoggerDebug string
}

func (vr *vastResponse) addOwStatusHeader(headers map[string]string, nbr openrtb3.NoBidReason) {
	if vr.debug == "1" {
		headers[HeaderOpenWrapStatus] = fmt.Sprintf(NBRFormat, nbr)
	}
}

func (vr *vastResponse) formVastResponse(adpodWriter *utils.HTTPResponseBufferWriter) ([]byte, map[string]string, int) {
	var statusCode = http.StatusOK
	var headers = map[string]string{
		ContentType:    ApplicationXML,
		ContentOptions: NoSniff,
	}

	if adpodWriter.Code > 0 && adpodWriter.Code == http.StatusBadRequest {
		vr.addOwStatusHeader(headers, nbr.InvalidVideoRequest)
		return EmptyVASTResponse, headers, adpodWriter.Code
	}

	response, err := io.ReadAll(adpodWriter.Response)
	if err != nil {
		statusCode = http.StatusInternalServerError
		vr.addOwStatusHeader(headers, nbr.InternalError)
		return EmptyVASTResponse, headers, statusCode
	}

	var bidResponse *openrtb2.BidResponse
	err = json.Unmarshal(response, &bidResponse)
	if err != nil {
		statusCode = http.StatusInternalServerError
		vr.addOwStatusHeader(headers, nbr.InternalError)
		return EmptyVASTResponse, headers, statusCode
	}

	if bidResponse.NBR != nil {
		statusCode = http.StatusBadRequest
		vr.addOwStatusHeader(headers, *bidResponse.NBR)
		return EmptyVASTResponse, headers, statusCode
	}

	vast, nbr, err := vr.getVast(bidResponse)
	if nbr != nil {
		vr.addOwStatusHeader(headers, *nbr)
		return EmptyVASTResponse, headers, statusCode
	}

	return []byte(vast), headers, statusCode
}

func (vr *vastResponse) getVast(bidResponse *openrtb2.BidResponse) (string, *openrtb3.NoBidReason, error) {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return "", nbr.EmptySeatBid.Ptr(), errors.New("empty bid response")
	}

	builder := GetAdPodBuilder()
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			if bid.Price <= 0 {
				continue
			}
			if err := builder.Append(&bid); err != nil {
				nbr := exchange.ResponseRejectedGeneral
				return "", &nbr, err
			}
		}
	}

	creative, err := builder.Build()
	if err != nil {
		nbr := exchange.ResponseRejectedGeneral
		return "", &nbr, errors.New("No Bid")
	}

	if vr.debug == "1" || vr.WrapperLoggerDebug == "1" {
		creative = string(addExtInfo([]byte(creative), bidResponse.Ext))
	}

	return creative, nil, nil
}

func addExtInfo(vastBytes []byte, responseExt json.RawMessage) []byte {
	var v vast.VAST
	if err := xml.Unmarshal(vastBytes, &v); err != nil {
		return vastBytes
	}

	if len(v.Ads) == 0 {
		return vastBytes
	}

	owExtBytes := append([]byte("<Ext>"), append(responseExt, []byte("</Ext>")...)...)

	owExt := vast.Extension{
		Type: "OpenWrap",
		Data: owExtBytes,
	}

	ad := v.Ads[0]
	if ad.InLine != nil {
		if ad.InLine.Extensions == nil {
			ad.InLine.Extensions = &([]vast.Extension{})
		}
		*ad.InLine.Extensions = append(*ad.InLine.Extensions, owExt)
	} else if ad.Wrapper != nil {
		if ad.Wrapper.Extensions == nil {
			ad.Wrapper.Extensions = []vast.Extension{}
		}
		ad.Wrapper.Extensions = append(ad.Wrapper.Extensions, owExt)
	}

	newVASTBytes, err := xml.Marshal(v)
	if err != nil {
		return vastBytes
	}

	return newVASTBytes
}
