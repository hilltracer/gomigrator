package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Logger struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logger"`

	Storage struct {
		DSN string `mapstructure:"dsn"` // string «host=… port=…»
	} `mapstructure:"storage"`
}

func New(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	// expand data variables
	data := os.ExpandEnv(string(raw))

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(data)); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
