package mediafuse

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewMediafuseSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("mediafuse", temp, adapters.SyncTypeIframe)
}
