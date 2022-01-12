package sovrn

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewSovrnSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("sovrn", temp, adapters.SyncTypeRedirect)
}
