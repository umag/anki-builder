package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	GeminiAPIKey   string `json:"geminiApiKey"`
	AnkiDeck       string `json:"ankiDeck"`
	AnkiConnectURL string `json:"ankiConnectUrl"`
}

func Load() *Config {
	var cfg Config
	cfgBytes, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Failed to read config.json: %v", err)
	}

	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse config.json: %v", err)
	}

	if cfg.GeminiAPIKey == "" {
		log.Fatal("gemini api key config value is required")
	}

	return &cfg
}
