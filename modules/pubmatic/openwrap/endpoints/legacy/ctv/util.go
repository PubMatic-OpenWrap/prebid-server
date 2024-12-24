package ctv

import (
	"net/url"
)

func GetPubIdFromQueryParams(params url.Values) string {
	pubId := params.Get(ORTBSitePublisherID)
	if len(pubId) == 0 {
		pubId = params.Get(ORTBAppPublisherID)
	}
	return pubId
}
