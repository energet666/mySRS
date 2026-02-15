package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OutputDir string   `yaml:"output_dir"`
	Geosite   []string `yaml:"geosite"`
	Geoip     []string `yaml:"geoip"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if cfg.OutputDir == "" {
		cfg.OutputDir = "./output"
	}

	return &cfg, nil
}
