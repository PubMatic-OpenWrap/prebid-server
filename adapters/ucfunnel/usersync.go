package ucfunnel

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewUcfunnelSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("ucfunnel", temp, adapters.SyncTypeRedirect)
}
