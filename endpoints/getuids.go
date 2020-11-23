package endpoints

import (
	"net/http"

	"github.com/PubMatic-OpenWrap/prebid-server/config"
	"github.com/PubMatic-OpenWrap/prebid-server/usersync"
	"github.com/julienschmidt/httprouter"

	"encoding/json"
)

type userSyncs struct {
	BuyerUIDs map[string]string `json:"buyeruids,omitempty"`
}

// NewGetUIDsEndpoint implements the /getuid endpoint which
// returns all the existing syncs for the user
func NewGetUIDsEndpoint(cfg config.HostCookie) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		pc := usersync.ParsePBSCookieFromRequest(r, &cfg)
		userSyncs := new(userSyncs)
		userSyncs.BuyerUIDs = pc.GetUIDs()
		json.NewEncoder(w).Encode(userSyncs)
	})
}
