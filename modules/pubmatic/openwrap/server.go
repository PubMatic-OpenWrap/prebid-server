package openwrap

import (
	"net/http"

	"git.pubmatic.com/PubMatic/go-common/logger"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/wakanda"
)

func initOpenWrapServer(cfg *config.Config) {
	cfg.Wakanda.HostName = cfg.Server.HostName
	cfg.Wakanda.DCName = cfg.Server.DCName
	cfg.Wakanda.PodName = getPodName()
	wakanda.InitWakanda(cfg.Wakanda)
	hbMux := http.NewServeMux()
	hbMux.HandleFunc("/wakanda", wakanda.Handler(cfg.Wakanda))
	srvInterface := ":" + cfg.Server.EndPoint
	go startServer(srvInterface, hbMux)
}

var startServer = func(srvInterface string, hbMux *http.ServeMux) error {
	if err := http.ListenAndServe(srvInterface, hbMux); err != nil {
		logger.Fatal("main.main:unable to start http server for /wakanda handler due to : %s", err.Error())
	}
	return nil
}
