package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Label      string                `json:"label"`
	Providers  map[string]Provider  `json:"providers"`
}

type Provider struct {
	Type       string                 `json:"type"`
	Arguments  map[string]interface{} `json:"arguments,omitempty"`
}

var AppConfig *Config

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	AppConfig = &cfg
	return nil
}
