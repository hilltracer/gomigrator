package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Logger  LoggerConf  `mapstructure:"logger"`
	Storage StorageConf `mapstructure:"storage"`
}

type LoggerConf struct {
	Level string `mapstructure:"level"`
}

type StorageConf struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	DBName      string `mapstructure:"dbname"`
	PasswordEnv string `mapstructure:"password_env"`
	SSLMode     string `mapstructure:"sslmode"`
}

func New(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
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
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}

	return cfg, nil
}
