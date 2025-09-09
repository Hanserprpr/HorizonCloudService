package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"environment"`
	Port        int    `mapstructure:"port"`
	
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		DBName   string `mapstructure:"dbname"`
	} `mapstructure:"database"`

	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`

	JWT struct {
		AccessSecret  string `mapstructure:"access_secret"`
		RefreshSecret string `mapstructure:"refresh_secret"`
		AccessExpire  int    `mapstructure:"access_expire"`  // 秒
		RefreshExpire int    `mapstructure:"refresh_expire"` // 秒
	} `mapstructure:"jwt"`

	Log struct {
		Level      string `mapstructure:"level"`
		FilePath   string `mapstructure:"file_path"`
		MaxSize    int    `mapstructure:"max_size"`
		MaxBackups int    `mapstructure:"max_backups"`
		MaxAge     int    `mapstructure:"max_age"`
		Compress   bool   `mapstructure:"compress"`
	} `mapstructure:"log"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 设置默认值
	viper.SetDefault("environment", "development")
	viper.SetDefault("port", 8081)
	viper.SetDefault("jwt.access_expire", 3600)   // 1小时
	viper.SetDefault("jwt.refresh_expire", 604800) // 7天
	viper.SetDefault("log.level", "info")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}