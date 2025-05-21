package shared

import (
	"log/slog"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var (
	config     *viper.Viper
	configOnce sync.Once
)

func NewConfig() *viper.Viper {
	configOnce.Do(func() {
		slog.Debug("Initiate new config")
		newConfig := viper.New()

		configFile := "../.env"
		if _, err := os.Stat(configFile); err == nil {
			slog.Debug(".env file found, using .env file")
			newConfig.SetConfigFile(configFile)

			if err := newConfig.ReadInConfig(); err != nil {
				panic(err) // Consider logging instead of panicking in production
			}
		}

		newConfig.AutomaticEnv()
		config = newConfig
	})

	slog.Debug("Returning cached config")
	return config
}
