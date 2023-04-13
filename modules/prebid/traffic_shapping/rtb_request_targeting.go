package trafficshapping

import (
	"fmt"

	"github.com/prebid/openrtb/v17/openrtb2"
)

type RTBRequestExpression interface {
	IsPresent() bool
}

type RTBRequest struct {
	request *openrtb2.BidRequest
}

func (t RTBRequest) IsPresent(key string) Expression {
	return rtbExpression{
		key,
		t.request,
	}
}

type rtbExpression struct {
	key     string
	request *openrtb2.BidRequest
}

func (r rtbExpression) Evaluate(p map[string]string) bool {
	empty := false
	if r.request == nil {
		empty = true
	}
	request := r.request
	switch r.key {
	case "device.IFA":
		empty = request.Device == nil || request.Device.IFA == ""
	}
	return !empty
}

func (r rtbExpression) GetName() string {
	return fmt.Sprintf("(req.%v = present)", r.key)
}
