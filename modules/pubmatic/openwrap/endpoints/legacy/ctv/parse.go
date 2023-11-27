package ctv

import (
	"fmt"

	"github.com/prebid/prebid-server/hooks/hookstage"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

func GetRequestExtWrapper(payload hookstage.EntrypointPayload, result *hookstage.HookResult[hookstage.EntrypointPayload]) (models.RequestExtWrapper, error) {
	wrapperLocation := []string{"ext", "wrapper"}

	reqExtWrapper, err := models.GetRequestExtWrapper(payload.Body, wrapperLocation...)
	if err != nil {
		return reqExtWrapper, fmt.Errorf("unable to get request extension wrapper : %v", err.Error())
	}

	return reqExtWrapper, err
}
