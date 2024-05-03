package openwrap

import (
	"net/http"

	"git.pubmatic.com/PubMatic/go-common/logger"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/wakanda"
)

func initOpenWrapServer(cfg config.Config) {
	hbMux := http.NewServeMux()
	hbMux.HandleFunc("/wakanda", wakanda.Handler(cfg.Wakanda))
	srvInterface := ":" + cfg.Server.Port
	go startServer(srvInterface, hbMux)
	cfg.Wakanda.HostName = cfg.Server.HostName
	cfg.Wakanda.DCName = cfg.Server.DCName
	wakanda.InitWakanda(cfg.Wakanda)
}

var startServer = func(srvInterface string, hbMux *http.ServeMux) error {
	if err := http.ListenAndServe(srvInterface, hbMux); err != nil {
		logger.Fatal("main.main:unable to start http server: %s", err.Error())
		return err
	}
	return nil
}
