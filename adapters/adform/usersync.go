package adform

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewAdformSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("adform", temp, adapters.SyncTypeRedirect)
}
