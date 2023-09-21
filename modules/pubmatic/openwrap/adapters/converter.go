package adapters

import (
	"encoding/json"
)

// convertExtToFieldMap converts bidder json parameter to object
func convertExtToFieldMap(bidderName string, ext json.RawMessage) JSONObject {
	fieldmap := JSONObject{}
	if err := json.Unmarshal(ext, &fieldmap); err != nil {
	}
	return fieldmap
}

// FixBidderParams will fixes bidder parameter types for prebid auction endpoint(UOE-5744)
func FixBidderParams(reqID, adapterName, bidderCode string, ext json.RawMessage) (json.RawMessage, error) {
	/*
		//check if fixing bidder parameters really required
		if err := router.GetBidderParamValidator().Validate(openrtb_ext.BidderName(bidderCode), ext); err == nil {
			//fixing bidder parameter datatype is not required
			return ext, nil
		}
	*/

	//convert jsonstring to jsonobj
	fieldMap := convertExtToFieldMap(bidderCode, ext)

	//get callback function and execute it
	callback := getBuilder(adapterName)

	//executing callback function
	return callback(BidderParameters{
		ReqID:       reqID,
		AdapterName: adapterName, //actual partner name
		BidderCode:  bidderCode,  //alias bidder name
		FieldMap:    fieldMap,
	})
}
