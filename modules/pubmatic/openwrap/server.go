package openwrap

import (
	"net/http"
	"strings"

	"git.pubmatic.com/PubMatic/go-common/logger"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/config"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/wakanda"
)

func initOpenWrapServer(cfg *config.Config) *http.Server {
	cfg.Wakanda.HostName = cfg.Server.HostName
	cfg.Wakanda.DCName = cfg.Server.DCName
	cfg.Wakanda.PodName = getPodName()
	wakanda.InitWakanda(cfg.Wakanda)
	hbMux := http.NewServeMux()
	hbMux.HandleFunc("/wakanda", wakanda.Handler(cfg.Wakanda))
	srvInterface := strings.TrimPrefix(cfg.Server.EndPoint, "http://")
	server := &http.Server{
		Handler: hbMux,
		Addr:    srvInterface,
	}
	go startServer(server)
	return server
}

func startServer(server *http.Server) {
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("main.main:unable to start http server for /wakanda handler due to : %s", err.Error())
	}
}
