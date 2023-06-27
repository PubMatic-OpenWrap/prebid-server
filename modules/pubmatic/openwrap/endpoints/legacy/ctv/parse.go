package ctv

import (
	"fmt"
	"net/http"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func ConvertRequestAndGetRequestExtWrapper(payload hookstage.EntrypointPayload, result *hookstage.HookResult[hookstage.EntrypointPayload]) (models.RequestExtWrapper, error) {

	var reqExtWrapper models.RequestExtWrapper
	var err error
	wrapperLocation := []string{"ext", "wrapper"}
	if payload.Request.Method == http.MethodPost {
		reqExtWrapper, err = models.GetRequestExtWrapper(payload.Body, wrapperLocation...)
		if err != nil {
			return reqExtWrapper, fmt.Errorf("unable to get request extension wrapper : %v", err.Error())
		}
	}

	return reqExtWrapper, err
}
