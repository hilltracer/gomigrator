package config

import (
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
	Type string `mapstructure:"type"` // memory | sql
	PG   PGConf `mapstructure:"pg"`
}

type PGConf struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	DBName      string `mapstructure:"dbname"`
	PasswordEnv string `mapstructure:"password_env"`
	SSLMode     string `mapstructure:"sslmode"`
}

func NewConfig(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
