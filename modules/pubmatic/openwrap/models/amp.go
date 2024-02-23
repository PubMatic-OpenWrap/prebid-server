package models

import (
	"net/http"
	"strconv"
)

func GetQueryParamRequestExtWrapper(request *http.Request) (RequestExtWrapper, error) {
	extWrapper := RequestExtWrapper{
		SSAuctionFlag: -1,
	}

	values := request.URL.Query()
	extWrapper.PubId, _ = strconv.Atoi(values.Get(PUBID_KEY))
	extWrapper.ProfileId, _ = strconv.Atoi(values.Get(PROFILEID_KEY))
	extWrapper.VersionId, _ = strconv.Atoi(values.Get(VERSION_KEY))

	return extWrapper, nil
}
