package models

import (
	"net/http"
	"net/url"
	"strconv"
)

func GetQueryParamRequestExtWrapper(request *http.Request) (RequestExtWrapper, error) {
	extWrapper := RequestExtWrapper{
		SSAuctionFlag: -1,
	}

	values := request.URL.Query()
	extWrapper.PubId, _ = strconv.Atoi(values.Get(PUBID_KEY))
	extWrapper.ProfileId, _ = strconv.Atoi(values.Get(PROFILEID_KEY))

	purl := request.URL.Query().Get(PAGEURL_KEY)
	if purl == "" {
		return extWrapper, nil
	}

	parsedurl, err := url.Parse(purl)
	if err != nil {
		return extWrapper, nil
	}

	purlValues := parsedurl.Query()

	extWrapper.VersionId, _ = strconv.Atoi(purlValues.Get(VERSION_KEY))

	if purlValues.Get(DEBUG_KEY) == "1" ||
		purlValues.Get(WrapperLoggerDebug) == "1" ||
		purlValues.Get(QADebug) == "1" {
		extWrapper.Debug = true
	}

	if purlValues.Get(ReturnAllBidStatus) == "1" {
		extWrapper.SSAuctionFlag = 0
	}

	return extWrapper, nil
}
