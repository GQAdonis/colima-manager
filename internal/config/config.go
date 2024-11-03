package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type ProfileConfig struct {
	CPUs           int    `yaml:"cpus"`
	Memory         int    `yaml:"memory"`
	DiskSize       int    `yaml:"disk_size"`
	VMType         string `yaml:"vm_type"`
	Runtime        string `yaml:"runtime"`
	NetworkAddress bool   `yaml:"network_address"`
	Kubernetes     bool   `yaml:"kubernetes"`
}

type AutoConfig struct {
	Enabled bool   `yaml:"enabled"`
	Default string `yaml:"default"`
}

type Config struct {
	Server struct {
		Port int        `yaml:"port"`
		Auto AutoConfig `yaml:"auto"`
	} `yaml:"server"`
	Profiles map[string]ProfileConfig `yaml:"profiles"`
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
