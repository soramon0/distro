package conf

import (
	"errors"

	"github.com/spf13/viper"
)

type Config struct {
	GIN_MODE       string `mapstructure:"GIN_MODE"`
	MONGO_DATABASE string `mapstructure:"MONGO_DATABASE"`
	MONGO_URI      string `mapstructure:"MONGO_URI"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			return nil, errors.New("rename app.example.env to app.env")
		} else {
			return nil, err
		}
	}
	
	var config *Config
	err := viper.Unmarshal(&config)

	return config, err
}
