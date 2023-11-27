package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/endpoints/legacy/ctv"
)

func enrichRequestBody(r *http.Request) error {
	bidRequest, err := ctv.NewOpenRTB(r).ParseORTBRequest(ctv.GetORTBParserMap())
	if err != nil {
		return err
	}

	body, err := json.Marshal(bidRequest)
	if err != nil {
		return err
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return nil
}
