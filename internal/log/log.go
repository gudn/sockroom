package log

import (
	dlog "log"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	_ "github.com/gudn/sockroom/internal/config"
)

func init() {
	var logger *zap.Logger
	var err error

	if viper.GetBool("debug") {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		dlog.Fatal(err)
	}

	zap.ReplaceGlobals(logger)
}
