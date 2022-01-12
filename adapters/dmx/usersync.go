package dmx

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewDmxSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("dmx", temp, adapters.SyncTypeRedirect)
}
