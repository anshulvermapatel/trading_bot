package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Strategy   string `json:"strategy"`
	Symbol     string `json:"symbol"`
	Timeframe  string `json:"timeframe"`
	DataSource string `json:"data_source"`
	Broker     string `json:"broker"`

	Zerodha ZerodhaConfig `json:"zerodha"`
}

type ZerodhaConfig struct {
	APIKey      string `json:"api_key"`
	AccessToken string `json:"access_token"`
}

func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	if cfg.Strategy == "" {
		return Config{}, fmt.Errorf("strategy is required")
	}
	if cfg.Symbol == "" {
		return Config{}, fmt.Errorf("symbol is required")
	}
	if cfg.Timeframe == "" {
		return Config{}, fmt.Errorf("timeframe is required")
	}
	if cfg.DataSource == "" {
		cfg.DataSource = "mock"
	}
	if cfg.Broker == "" {
		cfg.Broker = "mock"
	}
	return cfg, nil
}
