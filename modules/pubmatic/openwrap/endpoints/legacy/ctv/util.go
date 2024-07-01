package ctv

import (
	"net/url"

	"github.com/prebid/openrtb/v20/openrtb2"
)

func GetPubIdFromQueryParams(params url.Values) string {
	pubId := params.Get(ORTBSitePublisherID)
	if len(pubId) == 0 {
		pubId = params.Get(ORTBAppPublisherID)
	}
	return pubId
}

func ValidateEIDs(eids []openrtb2.EID) []openrtb2.EID {
	validEIDs := []openrtb2.EID{}
	for _, eid := range eids {
		validUIDs := make([]openrtb2.UID, 0, len(eid.UIDs))

		for _, uid := range eid.UIDs {
			uid.ID = uidRegexp.ReplaceAllString(uid.ID, "")
			if uid.ID != "" {
				validUIDs = append(validUIDs, uid)
			}
		}

		if len(validUIDs) > 0 {
			eid.UIDs = validUIDs
			validEIDs = append(validEIDs, eid)
		}
	}
	return validEIDs
}
