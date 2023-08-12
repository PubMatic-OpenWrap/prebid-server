package ctv

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func ConvertRequestAndGetRequestExtWrapper(payload hookstage.EntrypointPayload, result *hookstage.HookResult[hookstage.EntrypointPayload]) (models.RequestExtWrapper, error) {

	var reqExtWrapper models.RequestExtWrapper
	var err error
	wrapperLocation := []string{"ext", "wrapper"}

	var body []byte
	if payload.Request.Method == http.MethodGet {
		bidRequest, err := NewOpenRTB(payload.Request).ParseORTBRequest(GetORTBParserMap())
		if err != nil {
			return reqExtWrapper, errors.New("error while parsing ctv get request, reason: " + err.Error())
		}
		body, err = json.Marshal(bidRequest)
		if err != nil {
			return reqExtWrapper, errors.New("error while marshalling ctv get request, reason: " + err.Error())
		}
		reqExtWrapper, err = models.GetRequestExtWrapper(body, wrapperLocation...)
		if err != nil {
			return reqExtWrapper, fmt.Errorf("unable to get request extension wrapper : %v", err.Error())
		}
	}

	if payload.Request.Method == http.MethodPost {
		reqExtWrapper, err = models.GetRequestExtWrapper(payload.Body, wrapperLocation...)
		if err != nil {
			return reqExtWrapper, fmt.Errorf("unable to get request extension wrapper : %v", err.Error())
		}
	}

	result.ChangeSet.AddMutation(func(ep hookstage.EntrypointPayload) (hookstage.EntrypointPayload, error) {
		if ep.Request.Method == http.MethodGet {
			ep.Body = body
		}
		return ep, nil
	}, hookstage.MutationUpdate, "entrypoint-update-ctv-get-method")

	return reqExtWrapper, err
}
