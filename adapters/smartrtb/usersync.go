package smartrtb

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewSmartRTBSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("smartrtb", temp, adapters.SyncTypeRedirect)
}
