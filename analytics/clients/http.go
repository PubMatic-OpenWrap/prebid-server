package clients

import (
	"net/http"
)

var defaultHttpInstance = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 1024,
	},
	Timeout: 15000,
}

func GetDefaultHttpInstance() *http.Client {
	// TODO 2020-06-22 @see https://github.com/prebid/prebid-server/pull/1331#discussion_r436110097
	return defaultHttpInstance
}
