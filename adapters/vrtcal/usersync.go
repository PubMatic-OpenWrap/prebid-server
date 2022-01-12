package vrtcal

import (
	"text/template"

	"github.com/PubMatic-OpenWrap/prebid-server/adapters"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
)

func NewVrtcalSyncer(temp *template.Template) usersync.Usersyncer {
	return adapters.NewSyncer("vrtcal", temp, adapters.SyncTypeRedirect)
}
