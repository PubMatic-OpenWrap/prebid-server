package consumable

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewConsumableSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer(
		"consumable",
		temp,
		adapters.SyncTypeRedirect)
}
