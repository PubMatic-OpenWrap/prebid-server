package adkernel

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewAdkernelSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("adkernel", temp, adapters.SyncTypeRedirect)
}
