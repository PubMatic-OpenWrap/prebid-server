package valueimpression

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewValueImpressionSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("valueimpression", temp, adapters.SyncTypeRedirect)
}
