package usersync

import (
	"github.com/PubMatic-OpenWrap/prebid-server/pbs"
)

func NewRubiconSyncer(usersyncURL string) Usersyncer {
	return &syncer{
		familyName: "rubicon",
		syncInfo: &pbs.UsersyncInfo{
			URL:         usersyncURL,
			Type:        "redirect",
			SupportCORS: false,
		},
	}
}
