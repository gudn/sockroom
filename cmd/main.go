package main

import (
	"net"
	"net/http"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/gudn/sockroom"
	"github.com/gudn/sockroom/pkg/local"
	"github.com/gudn/sockroom/pkg/nats"

	_ "github.com/gudn/sockroom/internal/config"
	_ "github.com/gudn/sockroom/internal/log"
)

func mode() string {
	if viper.IsSet("mode") {
		return viper.GetString("mode")
	}
	return "local"
}

func main() {
	err := run()
	if err != nil {
		zap.L().Fatal("running error", zap.Error(err))
	}
}

func run() error {
	l, err := net.Listen("tcp", viper.GetString("bind"))
	if err != nil {
		return err
	}
	zap.L().Info("start listening", zap.String("bind", l.Addr().String()))

	var chans sockroom.Channels

	m := mode()
	zap.L().Info("setup mode", zap.String("mode", m))

	if m == "local" {
		chans = local.New()
	} else {
		chans, err = nats.New()
		if err != nil {
			return err
		}
	}

	defer chans.Quit()
	viper.OnConfigChange(func(in fsnotify.Event) {
		chans.Reload()
	})
	handler := sockroom.New(chans)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", handler)

	return http.Serve(l, mux)
}
