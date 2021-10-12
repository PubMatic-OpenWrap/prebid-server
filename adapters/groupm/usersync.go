package groupm

import (
	"text/template"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/usersync"
)

func NewGroupmSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("groupm", temp, adapters.SyncTypeIframe)
}
