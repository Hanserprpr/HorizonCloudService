package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"environment"`
	Port        int    `mapstructure:"port"`
	
	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`

	JWT struct {
		Secret    string `mapstructure:"secret"`
		ExpiresIn int    `mapstructure:"expires_in"`
	} `mapstructure:"jwt"`

	Services map[string]ServiceConfig `mapstructure:"services"`

	RateLimit struct {
		RequestsPerSecond int `mapstructure:"requests_per_second"`
		BurstSize         int `mapstructure:"burst_size"`
	} `mapstructure:"rate_limit"`
}

type ServiceConfig struct {
	Host   string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	Prefix string `mapstructure:"prefix"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 设置默认值
	viper.SetDefault("environment", "development")
	viper.SetDefault("port", 8080)
	viper.SetDefault("rate_limit.requests_per_second", 100)
	viper.SetDefault("rate_limit.burst_size", 200)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}