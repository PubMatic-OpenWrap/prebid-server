package beintoo

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewBeintooSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("Beintoo", temp, adapters.SyncTypeIframe)
}
