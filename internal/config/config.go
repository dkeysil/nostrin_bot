package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TelegramBot TelegramBotConfig `yaml:"telegramBot"`
	Nostr       NostrConfig       `yaml:"nostr"`
}

type TelegramBotConfig struct {
	Token string `yaml:"token"`
}

type NostrConfig struct {
	RelayURLs []string `yaml:"relayURLs"`
}

func LoadConfig() (*Config, error) {
	f, err := os.Open("config.yaml")
	if err != nil {
		return nil, err
	}

	defer f.Close()

	cfg := &Config{}

	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
