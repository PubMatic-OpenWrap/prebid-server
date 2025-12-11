package ctvvast

import (
	"encoding/json"
	"strings"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/openrtb/v20/openrtb3"
	"github.com/prebid/prebid-server/v3/exchange"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/vastbuilder"
)

const (
	inLineEnd       = "</InLine>"
	wrapperEnd      = "</Wrapper>"
	extensionsStart = "<Extensions>"
	extensionsEnd   = "</Extensions>"
)

var (
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
		return EmptyVASTResponse, openrtb3.NoBidUnknownError.Ptr()
	}

	isNobid := true
	builder := vastbuilder.GetVastBuilder()
	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {
			if bid.Price <= 0 {
				continue
			}
			isNobid = false
			if err := builder.Append(&bid); err != nil {
				nbr := exchange.ResponseRejectedGeneral
				return EmptyVASTResponse, &nbr
			}
		}
	}

	if isNobid {
		return EmptyVASTResponse, openrtb3.NoBidUnknownError.Ptr()
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
	adm := string(vastBytes)
	owExt := "<Extension type=" + `"OpenWrap"` + "><Ext><![CDATA[" + string(responseExt) + "]]></Ext></Extension>"

	// Check if Extensions Exists
	ci := strings.Index(adm, extensionsEnd)
	if ci != -1 {
		adm = strings.Replace(adm, extensionsEnd, owExt+extensionsEnd, 1)
		return []byte(adm)
	}

	// Check if Wrapper Exists
	wi := strings.Index(adm, wrapperEnd)
	if wi != -1 {
		adm = strings.Replace(adm, wrapperEnd, extensionsStart+owExt+extensionsEnd+wrapperEnd, 1)
		return []byte(adm)

	}

	// Check if Inline Exists
	wi = strings.Index(adm, inLineEnd)
	if wi != -1 {
		adm = strings.Replace(adm, inLineEnd, extensionsStart+owExt+extensionsEnd+inLineEnd, 1)
		return []byte(adm)
	}

	return vastBytes
}
