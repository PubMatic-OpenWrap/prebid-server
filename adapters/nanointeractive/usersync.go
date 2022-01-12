package nanointeractive

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewNanoInteractiveSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("nanointeractive", temp, adapters.SyncTypeRedirect)
}
