package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Payment  PaymentConfig  `mapstructure:"payment"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type PaymentConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
	BaseURL   string `mapstructure:"base_url"`
	PublicURL string `mapstructure:"public_url"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

func Load() *Config {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic("读取配置文件失败: " + err.Error())
		}
	}

	bindEnvVars(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic("解析配置失败: " + err.Error())
	}

	return &cfg
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", "8080")
	v.SetDefault("payment.base_url", "https://pay.xgdn.net")
	v.SetDefault("payment.public_url", "http://localhost:8080")
	v.SetDefault("jwt.secret", "xgdn-shop-demo-secret-key")
	v.SetDefault("database.path", "./shop-demo.db")
}

func bindEnvVars(v *viper.Viper) {
	envBindings := map[string]string{
		"server.port":        "PORT",
		"payment.app_id":     "XGDN_APP_ID",
		"payment.app_secret": "XGDN_APP_SECRET",
		"payment.base_url":   "XGDN_BASE_URL",
		"payment.public_url": "PUBLIC_URL",
		"jwt.secret":         "JWT_SECRET",
		"database.path":      "DB_PATH",
	}

	for key, env := range envBindings {
		if val := v.GetString(env); val != "" {
			v.Set(key, val)
		}
	}
}

func (c *Config) NotifyURL() string {
	return strings.TrimSuffix(c.Payment.PublicURL, "/") + "/api/callback/pay"
}
