package openx

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewOpenxSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("openx", temp, adapters.SyncTypeRedirect)
}
