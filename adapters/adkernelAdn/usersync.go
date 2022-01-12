package adkernelAdn

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewAdkernelAdnSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("adkernelAdn", temp, adapters.SyncTypeRedirect)
}
