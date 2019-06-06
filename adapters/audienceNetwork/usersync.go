package audienceNetwork

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewAudienceNetworkSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("audienceNetwork", 0, temp, adapters.SyncTypeRedirect)
}
