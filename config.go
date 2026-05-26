package xgdnpay

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type payConfig struct {
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
	BaseURL   string `yaml:"base_url"`
}

type sdkConfig struct {
	Payment payConfig `yaml:"payment"`
}

func loadSDKConfig() *sdkConfig {
	cfg := &sdkConfig{}

	cfg.Payment.AppID = os.Getenv("XGDN_PAY_APP_ID")
	cfg.Payment.AppSecret = os.Getenv("XGDN_PAY_APP_SECRET")
	cfg.Payment.BaseURL = os.Getenv("XGDN_PAY_BASE_URL")

	if cfg.Payment.AppID == "" || cfg.Payment.AppSecret == "" {
		tryYAMLFile(cfg, "config.yaml")
		tryYAMLFile(cfg, "config.local.yaml")
		tryYAMLFile(cfg, filepath.Join("config", "config.yaml"))
	}

	if cfg.Payment.BaseURL == "" {
		cfg.Payment.BaseURL = DefaultBaseURL
	}

	return cfg
}

func tryYAMLFile(cfg *sdkConfig, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var fileCfg sdkConfig
	if yaml.Unmarshal(data, &fileCfg) != nil {
		return
	}
	if cfg.Payment.AppID == "" {
		cfg.Payment.AppID = fileCfg.Payment.AppID
	}
	if cfg.Payment.AppSecret == "" {
		cfg.Payment.AppSecret = fileCfg.Payment.AppSecret
	}
	if cfg.Payment.BaseURL == "" {
		cfg.Payment.BaseURL = fileCfg.Payment.BaseURL
	}
}