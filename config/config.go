package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

var pathList = []string{
	"config.json",
	"$XDG_CONFIG_HOME/anki-builder/config.json",
}

type Config struct {
	GeminiAPIKey   string `json:"geminiApiKey"`
	AnkiDeckName   string `json:"ankiDeckName"`
	AnkiConnectURL string `json:"ankiConnectUrl"`
}

func Load() (*Config, error) {
	var cfgBytes []byte

	for _, path := range pathList {
		var err error
		path = os.ExpandEnv(path)
		cfgBytes, err = os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("found but failed to read config file %s: %w", path, err)
		}

		break
	}

	if cfgBytes == nil {
		return nil, fmt.Errorf("no config file found in paths: \n%s", strings.Join(pathList, "\n"))
	}

	var cfg Config
	err := json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config.json: %w", err)
	}

	if cfg.AnkiDeckName == "" {
		return nil, errors.New("anki deck name config value is required")
	}

	if cfg.AnkiConnectURL == "" {
		return nil, errors.New("anki connect url config value is required")
	}

	if cfg.GeminiAPIKey == "" {
		return nil, errors.New("gemini api key config value is required")
	}

	return &cfg, nil
}
