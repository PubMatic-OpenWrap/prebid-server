package pulsepoint

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewPulsepointSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("pulsepoint", temp, adapters.SyncTypeRedirect)
}
