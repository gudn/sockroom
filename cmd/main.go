package main

import (
	"net"
	"net/http"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/gudn/sockroom"
	"github.com/gudn/sockroom/pkg/local"

	_ "github.com/gudn/sockroom/internal/config"
	_ "github.com/gudn/sockroom/internal/log"
)

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

	chans := local.New()
	defer chans.Quit()
	viper.OnConfigChange(func(in fsnotify.Event) {
		chans.Reload()
	})
	handler := sockroom.New(chans)

	return http.Serve(l, handler)
}
