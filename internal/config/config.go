package config

import (
	"errors"
	"log"
	"os"

	"github.com/spf13/viper"
)

func init() {
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
}
