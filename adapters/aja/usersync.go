package aja

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewAJASyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("aja", temp, adapters.SyncTypeRedirect)
}
