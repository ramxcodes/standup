package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey    string `json:"api_key"`
	ModelName string `json:"model_name"`
	AiEnabled bool   `json:"ai_enabled"`
}

var defaultModel = "gemini-flash-latest"

// Config file path
func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".standup.json")
}

// load config from disk

func Load() (*Config, error) {
	path := getConfigPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{
			ModelName: defaultModel,
			AiEnabled: true,
		}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save config file

func Save(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath(), data, 0644)
}
