package ctv

import (
	"encoding/json"
	"net/url"
	"regexp"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
)

var uidRegexp = regexp.MustCompile(`^(UID2|ID5|BGID|euid|PAIRID|IDL|connectid|firstid|utiq):`)

func GetPubIdFromQueryParams(params url.Values) string {
	pubId := params.Get(ORTBSitePublisherID)
	if len(pubId) == 0 {
		pubId = params.Get(ORTBAppPublisherID)
	}
	return pubId
}

func ValidateEIDs(eids []openrtb2.EID) []openrtb2.EID {
	validEIDs := make([]openrtb2.EID, 0, len(eids))
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

func UpdateUserExtWithValidValues(user *openrtb2.User) {
	if user == nil {
		return
	}

	if user.Ext != nil {
		var userExt openrtb_ext.ExtUser
		err := json.Unmarshal(user.Ext, &userExt)
		if err != nil {
			return
		}
		if userExt.SessionDuration < 0 {
			userExt.SessionDuration = 0
		}

		if userExt.ImpDepth < 0 {
			userExt.ImpDepth = 0
		}
		eids := ValidateEIDs(userExt.Eids)
		userExt.Eids = nil
		if len(eids) > 0 {
			userExt.Eids = eids
		}

		userExtjson, err := json.Marshal(userExt)
		if err == nil {
			user.Ext = userExtjson
		}
	}

	if len(user.EIDs) > 0 {
		eids := ValidateEIDs(user.EIDs)
		user.EIDs = nil
		if len(eids) > 0 {
			user.EIDs = eids
		}
	}
}
