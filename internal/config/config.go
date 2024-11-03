package config

import (
	"flag"
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
		Port   int        `yaml:"port"`
		Host   string     `yaml:"host"`
		Daemon bool       `yaml:"daemon"`
		Auto   AutoConfig `yaml:"auto"`
	} `yaml:"server"`
	Profiles map[string]ProfileConfig `yaml:"profiles"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	config.Server.Port = 8080        // Default port
	config.Server.Host = "localhost" // Default host

	// Define command line flags
	var (
		configPath string
		daemon     bool
		host       string
		auto       bool
	)

	// Define flags with both short and long forms
	flag.StringVar(&configPath, "c", "", "Path to config file")
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.BoolVar(&daemon, "d", false, "Run in daemon mode")
	flag.BoolVar(&daemon, "daemon", false, "Run in daemon mode")
	flag.StringVar(&host, "h", "", "Server host address")
	flag.StringVar(&host, "host", "", "Server host address")
	flag.BoolVar(&auto, "a", false, "Automatically create and start default profile")
	flag.BoolVar(&auto, "auto", false, "Automatically create and start default profile")

	// Parse flags if they haven't been parsed yet
	if !flag.Parsed() {
		flag.Parse()
	}

	// Determine config file path with following precedence:
	// 1. Command line flags (-c or --config)
	// 2. Environment variable (COLIMA_MANAGER_CONFIG)
	// 3. Default path (config.yaml)
	var configFile string
	if configPath != "" {
		configFile = configPath
	} else if envPath := os.Getenv("COLIMA_MANAGER_CONFIG"); envPath != "" {
		configFile = envPath
	} else {
		configFile = "config.yaml"
	}

	// Try to load from config file
	data, err := os.ReadFile(configFile)
	if err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with command line flags if provided
	if daemon {
		config.Server.Daemon = true
	}

	if host != "" {
		config.Server.Host = host
	}

	if auto {
		config.Server.Auto.Enabled = true
		// If no default profile is set, create one with sensible defaults
		if config.Server.Auto.Default == "" {
			config.Server.Auto.Default = "default"
			if config.Profiles == nil {
				config.Profiles = make(map[string]ProfileConfig)
			}
			if _, exists := config.Profiles[config.Server.Auto.Default]; !exists {
				config.Profiles[config.Server.Auto.Default] = ProfileConfig{
					CPUs:           4,
					Memory:         8,
					DiskSize:       60,
					VMType:         "vz",
					Runtime:        "containerd",
					NetworkAddress: true,
					Kubernetes:     true,
				}
			}
		}
	}

	return config, nil
}
