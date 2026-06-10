package signal

import (
	"encoding/base64"
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/util/jsonutil"
)

const (
	googleAndroidAppID = "com.google.ads.mediation.pubmatic.PubMaticMediationAdapter"
	googleIOSAppID     = "GADMediationAdapterPubMatic"
)

// ParseForEndpoint extracts embedded SDK signal data from the original request body.
func ParseForEndpoint(endpoint string, body []byte) *openrtb2.BidRequest {
	switch endpoint {
	case models.EndpointGoogleSDK:
		return ParseGoogle(body)
	case models.EndpointAPS:
		return ParseAPS(body)
	case models.EndpointAppLovinMax:
		return ParseAppLovinMax(body)
	case models.EndpointUnityLevelPlay:
		return ParseUnityLevelPlay(body)
	default:
		return nil
	}
}

func ParseGoogle(body []byte) *openrtb2.BidRequest {
	data, dataType, _, err := jsonparser.Get(body, "imp", "[0]", "ext", "buyer_generated_request_data")
	if err != nil || dataType != jsonparser.Array {
		return nil
	}

	var signalData *openrtb2.BidRequest
	_, _ = jsonparser.ArrayEach(data, func(sdkData []byte, dataType jsonparser.ValueType, _ int, err error) {
		if err != nil || dataType != jsonparser.Object || signalData != nil {
			return
		}

		id, err := jsonparser.GetString(sdkData, "source_app", "id")
		if err != nil || (id != googleAndroidAppID && id != googleIOSAppID) {
			return
		}

		signal, err := jsonparser.GetString(sdkData, "data")
		if err != nil || len(signal) == 0 {
			return
		}

		decodedSignal, err := base64.StdEncoding.DecodeString(signal)
		if err != nil {
			return
		}

		signalData = &openrtb2.BidRequest{}
		if err := jsonutil.Unmarshal(decodedSignal, signalData); err != nil {
			signalData = nil
		}
	})

	return signalData
}

func ParseAPS(body []byte) *openrtb2.BidRequest {
	signal, err := jsonparser.GetString(body, "user", "buyeruid")
	if err != nil || signal == "" {
		return nil
	}
	signalData := &openrtb2.BidRequest{}
	if err := jsonutil.Unmarshal([]byte(signal), signalData); err != nil {
		return nil
	}
	return signalData
}

func ParseAppLovinMax(body []byte) *openrtb2.BidRequest {
	signal, err := jsonparser.GetString(body, "user", "data", "[0]", "segment", "[0]", "signal")
	if err != nil || signal == "" {
		return nil
	}
	signalData := &openrtb2.BidRequest{}
	if err := json.Unmarshal([]byte(signal), signalData); err != nil {
		return nil
	}
	return signalData
}

func ParseUnityLevelPlay(body []byte) *openrtb2.BidRequest {
	token, err := jsonparser.GetString(body, "app", "ext", "token")
	if err != nil || token == "" {
		return nil
	}
	signalData, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil
	}
	signal := &openrtb2.BidRequest{}
	if err := jsonutil.Unmarshal(signalData, signal); err != nil {
		return nil
	}
	return signal
}
