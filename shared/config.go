package shared

import "github.com/spf13/viper"

func NewConfig() *viper.Viper {
	// Initiate new viper instance
	config := viper.New()

	// Configuration stuff
	config.SetConfigFile(".env")
	config.AddConfigPath("../")
	config.AutomaticEnv()

	// Read the config
	err := config.ReadInConfig()
	if err != nil {
		panic(err)
	}

	// Return the config
	return config
}
