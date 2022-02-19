package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/spf13/viper"
	"github.com/fsnotify/fsnotify"

	"github.com/gudn/sockroom"
	"github.com/gudn/sockroom/pkg/local"
)

func main() {
	viper.SetDefault("local.nworkers", 1)
	viper.SetDefault("local.bufsize", 16)

	viper.SetConfigFile("config.yaml")
	viper.SetConfigFile("config.toml")
	viper.AddConfigPath("/etc/sockroom/")
	viper.AddConfigPath(".")

	viper.WatchConfig()

	viper.SetEnvPrefix("sr")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && !errors.Is(err, os.ErrNotExist) {
			log.Fatal(err)
		}
	}

	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	l, err := net.Listen("tcp", viper.GetString("bind"))
	if err != nil {
		return err
	}
	log.Printf("listening on http://%v", l.Addr())

	chans := local.New()
	defer chans.Quit()
	viper.OnConfigChange(func(in fsnotify.Event) {
		chans.Reload()
	})
	handler := sockroom.New(chans)

	return http.Serve(l, handler)
}
