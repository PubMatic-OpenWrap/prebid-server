package gamma

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewGammaSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("gamma", temp, adapters.SyncTypeIframe)
}
