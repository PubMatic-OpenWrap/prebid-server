package audienceNetwork

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewFacebookSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("audienceNetwork", temp, adapters.SyncTypeRedirect)
}
