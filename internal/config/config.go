package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	config.Server.Port = 8080 // Default port

	// Try to load from config file
	data, err := os.ReadFile("config.yaml")
	if err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}
