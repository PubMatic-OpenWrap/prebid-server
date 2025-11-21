package ctvvast

import (
	"encoding/json"
	"encoding/xml"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/exchange"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/rs/vast"
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

func formVastResponse(rCtx *models.RequestCtx, bidResponse *openrtb2.BidResponse) ([]byte, *openrtb3.NoBidReason) {
	if bidResponse == nil || bidResponse.SeatBid == nil {
		return EmptyVASTResponse, nil
	}

	builder := GetVastBuilder()
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			if bid.Price <= 0 {
				continue
			}
			if err := builder.Append(&bid); err != nil {
				nbr := exchange.ResponseRejectedGeneral
				return EmptyVASTResponse, &nbr
			}
		}
	}

	creative, err := builder.Build()
	if err != nil {
		nbr := exchange.ResponseRejectedGeneral
		return EmptyVASTResponse, &nbr
	}

	if rCtx.Debug {
		creative = string(addExtInfo([]byte(creative), bidResponse.Ext))
	}

	return []byte(creative), nil
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
